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
	"fmt"
	"log"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/labels"
)

type Service struct {
	Base
	Service   *v1.Service
	Client    corev1.ServiceInterface
	APIClient client.Interface
}

type serviceTemplateFactory struct{}

// Returns wrapped resource name if it was a service
func (serviceTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Service == nil {
		return ""
	}
	return definition.Service.Name
}

// k8s resource kind that this fabric supports
func (serviceTemplateFactory) Kind() string {
	return "service"
}

// New returns new Service based on resource definition
func (serviceTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewService(parametrizeResource(def.Service, gc).(*v1.Service), c.Services(), c, def.Meta)
}

// NewExisting returns new ExistingService based on resource definition
func (serviceTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingService(name, c.Services())
}

func serviceStatus(s corev1.ServiceInterface, name string, apiClient client.Interface) (interfaces.ResourceStatus, error) {
	service, err := s.Get(name)

	if err != nil {
		return interfaces.ResourceError, err
	}

	log.Printf("Checking service status for selector %v", service.Spec.Selector)
	for k, v := range service.Spec.Selector {
		stringSelector := fmt.Sprintf("%s=%s", k, v)
		log.Printf("Checking status for %s", stringSelector)
		selector, err := labels.Parse(stringSelector)
		if err != nil {
			return interfaces.ResourceError, err
		}

		options := v1.ListOptions{LabelSelector: selector.String()}

		pods, err := apiClient.Pods().List(options)
		if err != nil {
			return interfaces.ResourceError, err
		}
		jobs, err := apiClient.Jobs().List(options)
		if err != nil {
			return interfaces.ResourceError, err
		}
		replicasets, err := apiClient.ReplicaSets().List(options)
		if err != nil {
			return interfaces.ResourceError, err
		}
		resources := make([]interfaces.BaseResource, 0, len(pods.Items)+len(jobs.Items)+len(replicasets.Items))
		for _, pod := range pods.Items {
			resources = append(resources, NewPod(&pod, apiClient.Pods(), nil))
		}
		for _, j := range jobs.Items {
			resources = append(resources, NewJob(&j, apiClient.Jobs(), nil))
		}
		for _, r := range replicasets.Items {
			resources = append(resources, NewReplicaSet(&r, apiClient.ReplicaSets(), nil))
		}
		if apiClient.IsEnabled(v1beta1.SchemeGroupVersion) {
			statefulsets, err := apiClient.StatefulSets().List(options)
			if err != nil {
				return interfaces.ResourceError, err
			}
			for _, ps := range statefulsets.Items {
				resources = append(resources, NewStatefulSet(&ps, apiClient.StatefulSets(), apiClient, nil))
			}
		} else {
			petsets, err := apiClient.PetSets().List(api.ListOptions{LabelSelector: selector})
			if err != nil {
				return interfaces.ResourceError, err
			}
			for _, ps := range petsets.Items {
				resources = append(resources, NewPetSet(&ps, apiClient.PetSets(), apiClient, nil))
			}
		}
		status, err := resourceListStatus(resources)
		if status != interfaces.ResourceReady || err != nil {
			return status, err
		}
	}

	return interfaces.ResourceReady, nil
}

func serviceKey(name string) string {
	return "service/" + name
}

func (s Service) Key() string {
	return serviceKey(s.Service.Name)
}

func (s Service) Create() error {
	if err := checkExistence(s); err != nil {
		log.Println("Creating", s.Key())
		s.Service, err = s.Client.Create(s.Service)
		return err
	}
	return nil
}

// Delete deletes Service from the cluster
func (s Service) Delete() error {
	return s.Client.Delete(s.Service.Name, nil)
}

// Status returns Service Status. It is based on the status of all objects which match the service selector. If all of them are ready, the Service is considered ready.
func (s Service) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return serviceStatus(s.Client, s.Service.Name, s.APIClient)
}

// NewService is Service constructor. Needs apiClient for service status checks
func NewService(service *v1.Service, client corev1.ServiceInterface, apiClient client.Interface, meta map[string]interface{}) interfaces.Resource {
	return report.SimpleReporter{BaseResource: Service{Base: Base{meta}, Service: service, Client: client, APIClient: apiClient}}
}

// StatusIsCacheable for service always returns false since the status must be
// checked on each request and not be cached
func (s Service) StatusIsCacheable(meta map[string]string) bool {
	return false
}

type ExistingService struct {
	Base
	Name      string
	Client    corev1.ServiceInterface
	APIClient client.Interface
}

func (s ExistingService) Key() string {
	return serviceKey(s.Name)
}

func (s ExistingService) Create() error {
	return createExistingResource(s)
}

// Status returns Service Status. It is based on the status of all objects which match the service selector. If all of them are ready, the Service is considered ready.
func (s ExistingService) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return serviceStatus(s.Client, s.Name, s.APIClient)
}

// Delete deletes Service from the cluster
func (s ExistingService) Delete() error {
	return s.Client.Delete(s.Name, nil)
}

// StatusIsCacheable for service always returns false since the status must be
// checked on each request and not be cached
func (s ExistingService) StatusIsCacheable(meta map[string]string) bool {
	return false
}

func NewExistingService(name string, client corev1.ServiceInterface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingService{Name: name, Client: client}}
}
