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
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/Mirantis/k8s-AppController/client"
)

// PetSet is a wrapper for K8s PetSet object
type PetSet struct {
	PetSet    *apps.PetSet
	Client    unversioned.PetSetInterface
	APIClient client.Interface
}

func petSetStatus(p unversioned.PetSetInterface, name string, apiClient client.Interface) (string, error) {
	//Use label from petset spec to get needed pods

	ps, err := p.Get(name)
	if err != nil {
		return "error", err
	}
	var labelSelectors []string
	for k, v := range ps.Spec.Template.ObjectMeta.Labels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
	}
	stringSelector := strings.Join(labelSelectors, ",")
	selector, err := labels.Parse(stringSelector)
	log.Printf("%s,%v\n", stringSelector, selector)
	if err != nil {
		return "error", err
	}
	options := api.ListOptions{LabelSelector: selector}

	pods, err := apiClient.Pods().List(options)
	if err != nil {
		return "error", err
	}
	resources := make([]Resource, 0, len(pods.Items))
	for _, pod := range pods.Items {
		p := pod
		resources = append(resources, NewPod(&p, apiClient.Pods()))
	}

	status, err := resourceListReady(resources)
	if status != "ready" || err != nil {
		return status, err
	}

	return "ready", nil
}

func petSetKey(name string) string {
	return "petset/" + name
}

// Key returns PetSet name
func (p PetSet) Key() string {
	return petSetKey(p.PetSet.Name)
}

// Create looks for a PetSet in Kubernetes cluster and creates it if it's not there
func (p PetSet) Create() error {
	log.Println("Looking for pet set", p.PetSet.Name)
	status, err := p.Status(nil)

	if err == nil {
		log.Printf("Found pet set %s, status: %s ", p.PetSet.Name, status)
		log.Println("Skipping creation of pet set", p.PetSet.Name)
		return nil
	}

	log.Println("Creating pet set", p.PetSet.Name)
	_, err = p.Client.Create(p.PetSet)
	return err
}

// Status returns PetSet status as a string. "ready" is regarded as sufficient for it's dependencies to be created.
func (p PetSet) Status(meta map[string]string) (string, error) {
	return petSetStatus(p.Client, p.PetSet.Name, p.APIClient)
}

// NewPetSet is a constructor
func NewPetSet(petSet *apps.PetSet, client unversioned.PetSetInterface, apiClient client.Interface) PetSet {
	return PetSet{PetSet: petSet, Client: client, APIClient: apiClient}
}

// ExistingPetSet is a wrapper for K8s PetSet object which is meant to already be in a cluster bofer AppController execution
type ExistingPetSet struct {
	Name      string
	Client    unversioned.PetSetInterface
	APIClient client.Interface
}

// Key returns PetSet name
func (p ExistingPetSet) Key() string {
	return petSetKey(p.Name)
}

// Create looks for existing PetSet and returns an error if there is no such PetSet in a cluster
func (p ExistingPetSet) Create() error {
	log.Println("Looking for pet set", p.Name)
	status, err := p.Status(nil)

	if err == nil {
		log.Printf("Found pet set %s, status: %s ", p.Name, status)
		log.Println("Skipping creation of pet set", p.Name)
		return nil
	}

	log.Fatalf("Pet set %s not found", p.Name)
	return errors.New("Pet set not found")
}

// Status returns PetSet status as a string. "ready" is regarded as sufficient for it's dependencies to be created.
func (p ExistingPetSet) Status(meta map[string]string) (string, error) {
	return petSetStatus(p.Client, p.Name, p.APIClient)
}

// NewExistingPetSet is a constructor
func NewExistingPetSet(name string, client unversioned.PetSetInterface, apiClient client.Interface) ExistingPetSet {
	return ExistingPetSet{Name: name, Client: client, APIClient: apiClient}
}
