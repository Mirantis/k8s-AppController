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
	"github.com/Mirantis/k8s-AppController/pkg/report"

	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

// StatefulSet is a wrapper for K8s StatefulSet object
type StatefulSet struct {
	Base
	StatefulSet *appsbeta1.StatefulSet
	Client      v1beta1.StatefulSetInterface
	APIClient   client.Interface
}

type statefulSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a statefulset
func (statefulSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.StatefulSet == nil {
		return ""
	}
	return getObjectName(definition.StatefulSet)
}

// k8s resource kind that this fabric supports
func (statefulSetTemplateFactory) Kind() string {
	return "statefulset"
}

// New returns new StatefulSet based on resource definition
func (statefulSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	newStatefulSet := parametrizeResource(def.StatefulSet, gc,
		"Spec.Template.Spec.Containers.Env",
		"Spec.Template.Spec.InitContainers.Env").(*appsbeta1.StatefulSet)
	return NewStatefulSet(newStatefulSet, c.StatefulSets(), c, def.Meta)
}

// NewExisting returns new ExistingStatefulSet based on resource definition
func (statefulSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingStatefulSet(name, c.StatefulSets(), c)
}

func statefulsetStatus(p v1beta1.StatefulSetInterface, name string, apiClient client.Interface) (interfaces.ResourceStatus, error) {
	// Use label from statefulset spec to get needed pods
	ps, err := p.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}
	return podsStateFromLabels(apiClient, ps.Spec.Template.ObjectMeta.Labels)
}

func statefulsetKey(name string) string {
	return "statefulset/" + name
}

// Key returns StatefulSet name
func (p StatefulSet) Key() string {
	return statefulsetKey(getObjectName(p.StatefulSet))
}

// Create looks for a StatefulSet in Kubernetes cluster and creates it if it's not there
func (p *StatefulSet) Create() error {
	if err := checkExistence(p); err != nil {
		log.Println("Creating", p.Key())
		p.StatefulSet, err = p.Client.Create(p.StatefulSet)
		return err
	}
	return nil
}

// Delete deletes StatefulSet from the cluster
func (p StatefulSet) Delete() error {
	return p.Client.Delete(p.StatefulSet.Name, nil)
}

// Status returns StatefulSet status. interfaces.ResourceReady is regarded as sufficient for it's dependencies to be created.
func (p StatefulSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return statefulsetStatus(p.Client, p.StatefulSet.Name, p.APIClient)
}

// NewStatefulSet is a constructor
func NewStatefulSet(statefulset *appsbeta1.StatefulSet, client v1beta1.StatefulSetInterface, apiClient client.Interface, meta map[string]interface{}) interfaces.Resource {
	return report.SimpleReporter{BaseResource: &StatefulSet{Base: Base{meta}, StatefulSet: statefulset, Client: client, APIClient: apiClient}}
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

// Status returns StatefulSet status. interfaces.ResourceReady is regarded as sufficient for it's dependencies to be created.
func (p ExistingStatefulSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
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
