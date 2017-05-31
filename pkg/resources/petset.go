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
	appsalpha1 "github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
)

var petSetParamFields = []string{
	"Spec.Template.Spec.Containers.name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

// newPetSet is a wrapper for K8s PetSet object
type newPetSet struct {
	petSet *appsalpha1.PetSet
	client client.Interface
}

// existingPetSet is a wrapper for K8s PetSet object which is meant to already be in a cluster bofer AppController execution
type existingPetSet struct {
	name   string
	client client.Interface
}

type petSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a petset
func (petSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.PetSet == nil {
		return ""
	}
	return definition.PetSet.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (petSetTemplateFactory) Kind() string {
	return "petset"
}

// New returns PetSet controller for new resource based on resource definition
func (petSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	petSet := parametrizeResource(def.PetSet, gc, petSetParamFields).(*appsalpha1.PetSet)
	return createNewPetSet(petSet, c)
}

// NewExisting returns PetSet controller for existing resource by its name
func (petSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingPetSet{name: name, client: c}
}

func petsetProgress(client client.Interface, name string) (float32, error) {
	// Use label from petset spec to get needed pods
	ps, err := client.PetSets().Get(name)
	if err != nil {
		return 0, err
	}
	return podsStateFromLabels(client, ps.Spec.Template.ObjectMeta.Labels)
}

func petsetKey(name string) string {
	return "petset/" + name
}

// Key returns PetSet name
func (p newPetSet) Key() string {
	return petsetKey(p.petSet.Name)
}

// Create looks for the PetSet in Kubernetes cluster and creates it if it's not there
func (p newPetSet) Create() error {
	if checkExistence(p) {
		return nil

	}
	log.Println("Creating", p.Key())
	obj, err := p.client.PetSets().Create(p.petSet)
	p.petSet = obj
	return err
}

// Delete deletes PetSet from the cluster
func (p newPetSet) Delete() error {
	return p.client.PetSets().Delete(p.petSet.Name, nil)
}

// GetProgress returns PetSet deployment progress
func (p newPetSet) GetProgress() (float32, error) {
	return petsetProgress(p.client, p.petSet.Name)
}

func createNewPetSet(petset *appsalpha1.PetSet, client client.Interface) interfaces.Resource {
	return newPetSet{petSet: petset, client: client}
}

// Key returns PetSet name
func (p existingPetSet) Key() string {
	return petsetKey(p.name)
}

// Create looks for existing PetSet and returns error if there is no such PetSet
func (p existingPetSet) Create() error {
	return createExistingResource(p)
}

// GetProgress returns PetSet deployment progress
func (p existingPetSet) GetProgress() (float32, error) {
	return petsetProgress(p.client, p.name)
}

// Delete deletes PetSet from the cluster
func (p existingPetSet) Delete() error {
	return p.client.PetSets().Delete(p.name, nil)
}
