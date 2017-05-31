// Copyright 2017 Mirantis
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

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var daemonSetParamFields = []string{
	"Spec.Template.Spec.Containers.name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

// newDaemonSet is wrapper for K8s DaemonSet object
type newDaemonSet struct {
	daemonSet *extbeta1.DaemonSet
	client    v1beta1.DaemonSetInterface
}

// existingDaemonSet is a wrapper for K8s DaemonSet object which is deployed on a cluster before AppController
type existingDaemonSet struct {
	name   string
	client v1beta1.DaemonSetInterface
}

type daemonSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a daemonset
func (daemonSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.DaemonSet == nil {
		return ""
	}
	return definition.DaemonSet.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (daemonSetTemplateFactory) Kind() string {
	return "daemonset"
}

// New returns DaemonSets controller for new resource based on resource definition
func (d daemonSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	daemonSet := parametrizeResource(def.DaemonSet, gc, daemonSetParamFields).(*extbeta1.DaemonSet)
	return newDaemonSet{daemonSet: daemonSet, client: c.DaemonSets()}
}

// NewExisting returns DaemonSets controller for existing resource by its name
func (d daemonSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return newExistingDaemonSet(name, c.DaemonSets())
}

func daemonSetKey(name string) string {
	return "daemonset/" + name
}

func daemonSetProgress(d v1beta1.DaemonSetInterface, name string) (float32, error) {
	daemonSet, err := d.Get(name)
	if err != nil {
		return 0, err
	}
	if daemonSet.Status.DesiredNumberScheduled == 0 {
		return 1, nil
	}
	return float32(daemonSet.Status.CurrentNumberScheduled) / float32(daemonSet.Status.DesiredNumberScheduled), nil
}

// Key return DaemonSet name
func (d newDaemonSet) Key() string {
	return daemonSetKey(d.daemonSet.Name)
}

// GetProgress returns DaemonSet deployment progress
func (d newDaemonSet) GetProgress() (float32, error) {
	return daemonSetProgress(d.client, d.daemonSet.Name)
}

// Create looks for DaemonSet in k8s and creates it if not present
func (d newDaemonSet) Create() error {
	if checkExistence(d) {
		return nil
	}
	log.Println("Creating", d.Key())
	obj, err := d.client.Create(d.daemonSet)
	d.daemonSet = obj
	return err
}

// Delete deletes DaemonSet from the cluster
func (d newDaemonSet) Delete() error {
	return d.client.Delete(d.daemonSet.Name, &v1.DeleteOptions{})
}

// Key returns DaemonSet name
func (d existingDaemonSet) Key() string {
	return daemonSetKey(d.name)
}

// GetProgress returns DaemonSet deployment progress
func (d existingDaemonSet) GetProgress() (float32, error) {
	return daemonSetProgress(d.client, d.name)
}

// Create looks for existing DaemonSet and returns error if there is no such DaemonSet
func (d existingDaemonSet) Create() error {
	return createExistingResource(d)
}

// Delete deletes DaemonSet from the cluster
func (d existingDaemonSet) Delete() error {
	return d.client.Delete(d.name, nil)
}

// newExistingDaemonSet is a constructor
func newExistingDaemonSet(name string, client v1beta1.DaemonSetInterface) interfaces.Resource {
	return existingDaemonSet{name: name, client: client}
}
