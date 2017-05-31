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

	appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

var statefulSetParamFields = []string{
	"Spec.Template.Spec.Containers.Name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.Name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

// newStatefulSet is a wrapper for K8s StatefulSet object
type newStatefulSet struct {
	statefulSet *appsbeta1.StatefulSet
	client      client.Interface
}

// existingStatefulSet is a wrapper for K8s StatefulSet object which is meant to already be in a cluster before AppController execution
type existingStatefulSet struct {
	name   string
	client client.Interface
}

type statefulSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a StatefulSet
func (statefulSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.StatefulSet == nil {
		return ""
	}
	return definition.StatefulSet.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (statefulSetTemplateFactory) Kind() string {
	return "statefulset"
}

// New returns StatefulSet controller for new resource based on resource definition
func (statefulSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	statefulSet := parametrizeResource(def.StatefulSet, gc, statefulSetParamFields).(*appsbeta1.StatefulSet)
	return createNewStatefulSet(statefulSet, c)
}

// NewExisting returns StatefulSet controller for existing resource by its name
func (statefulSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingStatefulSet{name: name, client: c}
}

func statefulsetProgress(client client.Interface, name string) (float32, error) {
	// Use label from StatefulSet spec to get needed pods
	ps, err := client.StatefulSets().Get(name)
	if err != nil {
		return 0, err
	}
	return podsStateFromLabels(client, ps.Spec.Template.ObjectMeta.Labels)
}

func statefulsetKey(name string) string {
	return "statefulset/" + name
}

// Key returns StatefulSet name
func (p newStatefulSet) Key() string {
	return statefulsetKey(p.statefulSet.Name)
}

// Create looks for the StatefulSet in Kubernetes cluster and creates it if it's not there
func (p newStatefulSet) Create() error {
	if checkExistence(p) {
		return nil
	}
	log.Println("Creating", p.Key())
	obj, err := p.client.StatefulSets().Create(p.statefulSet)
	p.statefulSet = obj
	return err
}

// Delete deletes StatefulSet from the cluster
func (p newStatefulSet) Delete() error {
	return p.client.StatefulSets().Delete(p.statefulSet.Name, nil)
}

// GetProgress returns StatefulSet deployment progress
func (p newStatefulSet) GetProgress() (float32, error) {
	return statefulsetProgress(p.client, p.statefulSet.Name)
}

func createNewStatefulSet(statefulset *appsbeta1.StatefulSet, client client.Interface) interfaces.Resource {
	return newStatefulSet{statefulSet: statefulset, client: client}
}

// Key returns StatefulSet name
func (p existingStatefulSet) Key() string {
	return statefulsetKey(p.name)
}

// Create looks for existing StatefulSet and returns error if there is no such StatefulSet
func (p existingStatefulSet) Create() error {
	return createExistingResource(p)
}

// GetProgress returns StatefulSet deployment progress
func (p existingStatefulSet) GetProgress() (float32, error) {
	return statefulsetProgress(p.client, p.name)
}

// Delete deletes StatefulSet from the cluster
func (p existingStatefulSet) Delete() error {
	return p.client.StatefulSets().Delete(p.name, nil)
}
