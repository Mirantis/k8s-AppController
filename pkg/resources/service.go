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
	"reflect"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/labels"
)

var serviceParamFields = []string{
	"Spec.Selector",
}

type newService struct {
	Base
	Service   *v1.Service
	Client    corev1.ServiceInterface
	APIClient client.Interface
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
	def.Service = parametrizeResource(def.Service, gc, serviceParamFields).(*v1.Service)
	return createNewService(def, c)
}

func createNewService(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: newService{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			Service:   def.Service,
			Client:    c.Services(),
			APIClient: c,
		},
	}
}

// NewExisting returns Service controller for existing resource by its name
func (serviceTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingService{Name: name, Client: c.Services()}}
}

func serviceStatus(service *v1.Service, apiClient client.Interface) (interfaces.ResourceStatus, error) {
	log.Printf("Checking service status for selector %v", service.Spec.Selector)
	selector := labels.SelectorFromSet(service.Spec.Selector)

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
		def := MakeDefinition(&pod)
		resources = append(resources, createNewPod(def, apiClient.Pods()))
	}
	for _, j := range jobs.Items {
		def := MakeDefinition(&j)
		resources = append(resources, createNewJob(def, apiClient.Jobs()))
	}
	for _, r := range replicasets.Items {
		def := MakeDefinition(&r)
		resources = append(resources, createNewReplicaSet(def, apiClient.ReplicaSets()))
	}
	if apiClient.IsEnabled(v1beta1.SchemeGroupVersion) {
		statefulsets, err := apiClient.StatefulSets().List(options)
		if err != nil {
			return interfaces.ResourceError, err
		}
		for _, ps := range statefulsets.Items {
			def := MakeDefinition(&ps)
			resources = append(resources, createNewStatefulSet(def, apiClient))
		}
	} else {
		petsets, err := apiClient.PetSets().List(api.ListOptions{LabelSelector: selector})
		if err != nil {
			return interfaces.ResourceError, err
		}
		for _, ps := range petsets.Items {
			def := MakeDefinition(&ps)
			resources = append(resources, createNewStatefulSet(def, apiClient))
		}
	}
	status, err := resourceListStatus(resources)
	if status != interfaces.ResourceReady || err != nil {
		return status, err
	}

	return interfaces.ResourceReady, nil
}

func serviceKey(name string) string {
	return "service/" + name
}

// Key returns service name
func (s newService) Key() string {
	return serviceKey(s.Service.Name)
}

// Create looks for the Service in k8s and creates it if not present
func (s newService) Create() error {
	if err := checkExistence(s); err != nil {
		log.Println("Creating", s.Key())
		s.Service, err = s.Client.Create(s.Service)
		return err
	}
	return nil
}

// Delete deletes Service from the cluster
func (s newService) Delete() error {
	return s.Client.Delete(s.Service.Name, nil)
}

// Status returns Service Status. It is based on the status of all objects which match the service selector. If all of them are ready, the Service is considered ready.
func (s newService) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	service, err := s.Client.Get(s.Service.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !s.equalsToDefinition(service) {
		return interfaces.ResourceWaitingForUpgrade, nil
	}
	return serviceStatus(service, s.APIClient)
}

// equalsToDefinition checks if definition in object is compatible with provided object
func (s newService) equalsToDefinition(serviceiface interface{}) bool {
	service := serviceiface.(*v1.Service)

	return reflect.DeepEqual(service.ObjectMeta, s.Service.ObjectMeta) && reflect.DeepEqual(service.Spec, s.Service.Spec)
}

type existingService struct {
	Base
	Name      string
	Client    corev1.ServiceInterface
	APIClient client.Interface
}

// Key returns service name
func (s existingService) Key() string {
	return serviceKey(s.Name)
}

// Create looks for existing Service and returns error if there is no such Service
func (s existingService) Create() error {
	return createExistingResource(s)
}

// Status returns Service Status. It is based on the status of all objects which match the service selector. If all of them are ready, the Service is considered ready.
func (s existingService) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	service, err := s.Client.Get(s.Name)

	if err != nil {
		return interfaces.ResourceError, err
	}
	return serviceStatus(service, s.APIClient)
}

// Delete deletes Service from the cluster
func (s existingService) Delete() error {
	return s.Client.Delete(s.Name, nil)
}
