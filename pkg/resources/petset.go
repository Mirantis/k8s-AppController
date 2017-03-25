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
	"github.com/Mirantis/k8s-AppController/pkg/client/petsets/typed/apps/v1alpha1"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

// PetSet is a wrapper for K8s PetSet object
type PetSet struct {
	Base
	PetSet    *appsalpha1.PetSet
	Client    v1alpha1.PetSetInterface
	APIClient client.Interface
}

type petSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a petset
func (petSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.PetSet == nil {
		return ""
	}
	return getObjectName(definition.PetSet)
}

// k8s resource kind that this fabric supports
func (petSetTemplateFactory) Kind() string {
	return "petset"
}

// New returns new PetSet based on resource definition
func (petSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	newPetSet := parametrizeResource(def.PetSet, gc,
		"Spec.Template.Spec.Containers.Env",
		"Spec.Template.Spec.InitContainers.Env").(*appsalpha1.PetSet)
	return NewPetSet(newPetSet, c.PetSets(), c, def.Meta)
}

// NewExisting returns new ExistingPetSet based on resource definition
func (petSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingPetSet(name, c.PetSets(), c)
}

func petsetStatus(p v1alpha1.PetSetInterface, name string, apiClient client.Interface) (interfaces.ResourceStatus, error) {
	// Use label from petset spec to get needed pods
	ps, err := p.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}
	return podsStateFromLabels(apiClient, ps.Spec.Template.ObjectMeta.Labels)
}

func petsetKey(name string) string {
	return "petset/" + name
}

// Key returns PetSet name
func (p PetSet) Key() string {
	return petsetKey(getObjectName(p.PetSet))
}

// Create looks for a PetSet in Kubernetes cluster and creates it if it's not there
func (p *PetSet) Create() error {
	if err := checkExistence(p); err != nil {
		log.Println("Creating", p.Key())
		p.PetSet, err = p.Client.Create(p.PetSet)
		return err
	}
	return nil
}

// Delete deletes PetSet from the cluster
func (p PetSet) Delete() error {
	return p.Client.Delete(p.PetSet.Name, nil)
}

// Status returns PetSet status. interfaces.ResourceReady is regarded as sufficient for it's dependencies to be created.
func (p PetSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return petsetStatus(p.Client, p.PetSet.Name, p.APIClient)
}

// NewPetSet is a constructor
func NewPetSet(petset *appsalpha1.PetSet, client v1alpha1.PetSetInterface, apiClient client.Interface, meta map[string]interface{}) interfaces.Resource {
	return report.SimpleReporter{BaseResource: &PetSet{Base: Base{meta}, PetSet: petset, Client: client, APIClient: apiClient}}
}

// ExistingPetSet is a wrapper for K8s PetSet object which is meant to already be in a cluster bofer AppController execution
type ExistingPetSet struct {
	Base
	Name      string
	Client    v1alpha1.PetSetInterface
	APIClient client.Interface
}

// Key returns PetSet name
func (p ExistingPetSet) Key() string {
	return petsetKey(p.Name)
}

// Create looks for existing PetSet and returns an error if there is no such PetSet in a cluster
func (p ExistingPetSet) Create() error {
	return createExistingResource(p)
}

// Status returns PetSet status. interfaces.ResourceReady is regarded as sufficient for it's dependencies to be created.
func (p ExistingPetSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return petsetStatus(p.Client, p.Name, p.APIClient)
}

// Delete deletes PetSet from the cluster
func (p ExistingPetSet) Delete() error {
	return p.Client.Delete(p.Name, nil)
}

// NewExistingPetSet is a constructor
func NewExistingPetSet(name string, client v1alpha1.PetSetInterface, apiClient client.Interface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingPetSet{Name: name, Client: client, APIClient: apiClient}}
}
