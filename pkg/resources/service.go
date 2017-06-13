// Copyright 2016 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"log"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/labels"
)

var serviceParamFields = []string{
	"Spec.Selector",
}

type newService struct {
	service *v1.Service
	client  client.Interface
}

type existingService struct {
	name   string
	client client.Interface
}

type serviceTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a service
func (serviceTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Service == nil {
		return ""
	}
	return definition.Service.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (serviceTemplateFactory) Kind() string {
	return "service"
}

// New returns Service controller for new resource based on resource definition
func (serviceTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	service := parametrizeResource(def.Service, gc, serviceParamFields).(*v1.Service)
	return newService{service: service, client: c}
}

// NewExisting returns Service controller for existing resource by its name
func (serviceTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingService{name: name, client: c}
}

func serviceProgress(client client.Interface, name string) (float32, error) {
	service, err := client.Services().Get(name)

	if err != nil {
		return 0, err
	}

	selector := labels.SelectorFromSet(service.Spec.Selector)

	options := v1.ListOptions{LabelSelector: selector.String()}

	pods, err := client.Pods().List(options)
	if err != nil {
		return 0, err
	}
	jobs, err := client.Jobs().List(options)
	if err != nil {
		return 0, err
	}
	replicasets, err := client.ReplicaSets().List(options)
	if err != nil {
		return 0, err
	}
	resources := make([]interfaces.Resource, 0, len(pods.Items)+len(jobs.Items)+len(replicasets.Items))
	for _, pod := range pods.Items {
		resources = append(resources, createNewPod(&pod, client.Pods()))
	}
	for _, j := range jobs.Items {
		resources = append(resources, createNewJob(&j, client.Jobs()))
	}
	for _, r := range replicasets.Items {
		resources = append(resources, createNewReplicaSet(&r, client.ReplicaSets()))
	}
	if client.IsEnabled(v1beta1.SchemeGroupVersion) {
		statefulsets, err := client.StatefulSets().List(options)
		if err != nil {
			return 0, err
		}
		for _, ps := range statefulsets.Items {
			resources = append(resources, createNewStatefulSet(&ps, client))
		}
	} else {
		petsets, err := client.PetSets().List(api.ListOptions{LabelSelector: selector})
		if err != nil {
			return 0, err
		}
		for _, ps := range petsets.Items {
			resources = append(resources, createNewPetSet(&ps, client))
		}
	}
	return resourceListProgress(resources)
}

func serviceKey(name string) string {
	return "service/" + name
}

// Key returns service name
func (s newService) Key() string {
	return serviceKey(s.service.Name)
}

// Create looks for the Service in k8s and creates it if not present
func (s newService) Create() error {
	if checkExistence(s) {
		return nil
	}
	log.Println("Creating", s.Key())
	obj, err := s.client.Services().Create(s.service)
	s.service = obj
	return err
}

// Delete deletes Service from the cluster
func (s newService) Delete() error {
	return s.client.Services().Delete(s.service.Name, nil)
}

// GetProgress returns Service deployment progress
func (s newService) GetProgress() (float32, error) {
	return serviceProgress(s.client, s.service.Name)
}

// Key returns service name
func (s existingService) Key() string {
	return serviceKey(s.name)
}

// Create looks for existing Service and returns error if there is no such Service
func (s existingService) Create() error {
	return createExistingResource(s)
}

// GetProgress returns Service deployment progress
func (s existingService) GetProgress() (float32, error) {
	return serviceProgress(s.client, s.name)
}

// Delete deletes Service from the cluster
func (s existingService) Delete() error {
	return s.client.Services().Delete(s.name, nil)
}
