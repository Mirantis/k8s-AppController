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

	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

// StatefulSet is a wrapper for K8s StatefulSet object
type StatefulSet struct {
	Base
	StatefulSet *appsbeta1.StatefulSet
	Client      v1beta1.StatefulSetInterface
	APIClient   client.Interface
}

func statefulsetStatus(p v1beta1.StatefulSetInterface, name string, apiClient client.Interface) (string, error) {
	// Use label from statefulset spec to get needed pods
	ps, err := p.Get(name)
	if err != nil {
		return "error", err
	}
	return podsStateFromLabels(apiClient, ps.Spec.Template.ObjectMeta.Labels)
}

func statefulsetKey(name string) string {
	return "statefulset/" + name
}

// Key returns StatefulSet name
func (p StatefulSet) Key() string {
	return statefulsetKey(p.StatefulSet.Name)
}

// Create looks for a StatefulSet in Kubernetes cluster and creates it if it's not there
func (p StatefulSet) Create() error {
	if err := checkExistence(p); err != nil {
		log.Println("Creating ", p.Key())
		_, err = p.Client.Create(p.StatefulSet)
		return err
	}
	return nil
}

// Delete deletes StatefulSet from the cluster
func (p StatefulSet) Delete() error {
	return p.Client.Delete(p.StatefulSet.Name, nil)
}

// Status returns StatefulSet status as a string. "ready" is regarded as sufficient for it's dependencies to be created.
func (p StatefulSet) Status(meta map[string]string) (string, error) {
	return statefulsetStatus(p.Client, p.StatefulSet.Name, p.APIClient)
}

// NameMatches gets resource definition and a name and checks if
// the StatefulSet part of resource definition has matching name.
func (p StatefulSet) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.StatefulSet != nil && def.StatefulSet.Name == name
}

// New returns new StatefulSet based on resource definition
func (p StatefulSet) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewStatefulSet(def.StatefulSet, c.StatefulSets(), c, def.Meta)
}

// NewExisting returns new ExistingStatefulSet based on resource definition
func (p StatefulSet) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingStatefulSet(name, c.StatefulSets(), c)
}

// NewStatefulSet is a constructor
func NewStatefulSet(statefulset *appsbeta1.StatefulSet, client v1beta1.StatefulSetInterface, apiClient client.Interface, meta map[string]string) interfaces.Resource {
	return report.SimpleReporter{BaseResource: StatefulSet{Base: Base{meta}, StatefulSet: statefulset, Client: client, APIClient: apiClient}}
}

// ExistingStatefulSet is a wrapper for K8s StatefulSet object which is meant to already be in a cluster bofer AppController execution
type ExistingStatefulSet struct {
	Base
	Name      string
	Client    v1beta1.StatefulSetInterface
	APIClient client.Interface
}

// Key returns StatefulSet name
func (p ExistingStatefulSet) Key() string {
	return statefulsetKey(p.Name)
}

// Create looks for existing StatefulSet and returns an error if there is no such StatefulSet in a cluster
func (p ExistingStatefulSet) Create() error {
	return createExistingResource(p)
}

// Status returns StatefulSet status as a string. "ready" is regarded as sufficient for it's dependencies to be created.
func (p ExistingStatefulSet) Status(meta map[string]string) (string, error) {
	return statefulsetStatus(p.Client, p.Name, p.APIClient)
}

// Delete deletes StatefulSet from the cluster
func (p ExistingStatefulSet) Delete() error {
	return p.Client.Delete(p.Name, nil)
}

// NewExistingStatefulSet is a constructor
func NewExistingStatefulSet(name string, client v1beta1.StatefulSetInterface, apiClient client.Interface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingStatefulSet{Name: name, Client: client, APIClient: apiClient}}
}
