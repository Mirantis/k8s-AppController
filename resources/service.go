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
	"errors"
	"fmt"
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
	"github.com/Mirantis/k8s-AppController/report"
)

type Service struct {
	Service   *api.Service
	Client    unversioned.ServiceInterface
	APIClient client.Interface
}

func serviceStatus(s unversioned.ServiceInterface, name string, apiClient client.Interface) (string, error) {
	service, err := s.Get(name)

	if err != nil {
		return "error", err
	}

	log.Printf("Checking service status for selector %v", service.Spec.Selector)
	for k, v := range service.Spec.Selector {
		stringSelector := fmt.Sprintf("%s=%s", k, v)
		log.Printf("Checking status for %s", stringSelector)
		selector, err := labels.Parse(stringSelector)
		if err != nil {
			return "error", err
		}

		options := api.ListOptions{LabelSelector: selector}

		pods, err := apiClient.Pods().List(options)
		if err != nil {
			return "error", err
		}
		jobs, err := apiClient.Jobs().List(options)
		if err != nil {
			return "error", err
		}
		replicasets, err := apiClient.ReplicaSets().List(options)
		if err != nil {
			return "error", err
		}
		petsets, err := apiClient.PetSets().List(options)
		if err != nil {
			return "error", err
		}
		resources := make([]interfaces.Resource, 0, len(pods.Items)+len(jobs.Items)+len(replicasets.Items))
		for _, pod := range pods.Items {
			p := pod
			resources = append(resources, NewPod(&p, apiClient.Pods()))
		}
		for _, job := range jobs.Items {
			j := job
			resources = append(resources, NewJob(&j, apiClient.Jobs()))
		}
		for _, rs := range replicasets.Items {
			r := rs
			resources = append(resources, NewReplicaSet(&r, apiClient.ReplicaSets()))
		}
		for _, ps := range petsets.Items {
			p := ps
			resources = append(resources, NewPetSet(&p, apiClient.PetSets(), apiClient))
		}
		status, err := resourceListReady(resources)
		if status != "ready" || err != nil {
			return status, err
		}
	}

	return "ready", nil
}

func serviceKey(name string) string {
	return "service/" + name
}

func (s Service) Key() string {
	return serviceKey(s.Service.Name)
}

func (s Service) Create() error {
	log.Println("Looking for service", s.Service.Name)
	status, err := s.Status(nil)

	if err == nil {
		log.Printf("Found service %s, status: %s ", s.Service.Name, status)
		log.Println("Skipping creation of service", s.Service.Name)
		return nil
	}

	log.Println("Creating service", s.Service.Name)
	s.Service, err = s.Client.Create(s.Service)
	return err
}

// Delete deletes Service from the cluster
func (s Service) Delete() error {
	return s.Client.Delete(s.Service.Name)
}

func (s Service) Status(meta map[string]string) (string, error) {
	return serviceStatus(s.Client, s.Service.Name, s.APIClient)
}

// NameMatches gets resource definition and a name and checks if
// the Service part of resource definition has matching name.
func (s Service) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Service != nil && def.Service.Name == name
}

// New returns new Service based on resource definition
func (s Service) New(def client.ResourceDefinition, c client.Interface) interfaces.Reporter {
	return report.SimpleReporter{Resource: NewService(def.Service, c.Services(), c)}
}

// NewExisting returns new ExistingService based on resource definition
func (s Service) NewExisting(name string, c client.Interface) interfaces.Reporter {
	return report.SimpleReporter{Resource: NewExistingService(name, c.Services())}
}

//NewService is Service constructor. Needs apiClient for service status checks
func NewService(service *api.Service, client unversioned.ServiceInterface, apiClient client.Interface) Service {
	return Service{Service: service, Client: client, APIClient: apiClient}
}

type ExistingService struct {
	Name      string
	Client    unversioned.ServiceInterface
	APIClient client.Interface
	Service
}

func (s ExistingService) Key() string {
	return serviceKey(s.Name)
}

func (s ExistingService) Create() error {
	log.Println("Looking for service", s.Name)
	status, err := s.Status(nil)

	if err == nil {
		log.Printf("Found service %s, status: %s ", s.Name, status)
		log.Println("Skipping creation of service", s.Name)
		return nil
	}

	log.Fatalf("Service %s not found", s.Name)
	return errors.New("Service not found")
}

func (s ExistingService) Status(meta map[string]string) (string, error) {
	return serviceStatus(s.Client, s.Name, s.APIClient)
}

func NewExistingService(name string, client unversioned.ServiceInterface) ExistingService {
	return ExistingService{Name: name, Client: client}
}
