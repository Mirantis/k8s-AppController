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

		options := api.ListOptions{LabelSelector: selector}

		pods, err := apiClient.Pods().List(options)
		if err != nil {
			return "error", err
		}
		for _, pod := range pods.Items {
			log.Printf("Checking status for pod %s", pod.Name)
			p := NewPod(&pod, apiClient.Pods())
			status, err := p.Status()
			if err != nil {
				return "error", err
			}
			if status != "ready" {
				return "not ready", fmt.Errorf("Pod %s is not ready", pod.Name)
			}
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
	status, err := s.Status()

	if err == nil {
		log.Printf("Found service %s, status: %s ", s.Service.Name, status)
		log.Println("Skipping creation of service", s.Service.Name)
		return nil
	}

	log.Println("Creating service", s.Service.Name)
	s.Service, err = s.Client.Create(s.Service)
	return err
}

func (s Service) Status() (string, error) {
	return serviceStatus(s.Client, s.Service.Name, s.APIClient)
}

//NewService is Service constructor. Needs apiClient for service status checks
func NewService(service *api.Service, client unversioned.ServiceInterface, apiClient client.Interface) Service {
	return Service{Service: service, Client: client, APIClient: apiClient}
}

type ExistingService struct {
	Name      string
	Client    unversioned.ServiceInterface
	APIClient client.Interface
}

func (s ExistingService) Key() string {
	return serviceKey(s.Name)
}

func (s ExistingService) Create() error {
	log.Println("Looking for service", s.Name)
	status, err := s.Status()

	if err == nil {
		log.Printf("Found service %s, status: %s ", s.Name, status)
		log.Println("Skipping creation of service", s.Name)
		return nil
	}

	log.Fatalf("Service %s not found", s.Name)
	return errors.New("Service not found")
}

func (s ExistingService) Status() (string, error) {
	return serviceStatus(s.Client, s.Name, s.APIClient)
}

func NewExistingService(name string, client unversioned.ServiceInterface) ExistingService {
	return ExistingService{Name: name, Client: client}
}
