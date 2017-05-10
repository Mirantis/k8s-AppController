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
	"github.com/Mirantis/k8s-AppController/pkg/report"

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var daemonSetParamFields = []string{
	"Spec.Template.Spec.Containers.Name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.Name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

// DaemonSet is wrapper for K8s DaemonSet object
type DaemonSet struct {
	Base
	DaemonSet *extbeta1.DaemonSet
	Client    v1beta1.DaemonSetInterface
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
	newDaemonSet := parametrizeResource(def.DaemonSet, gc, daemonSetParamFields).(*extbeta1.DaemonSet)
	return report.SimpleReporter{BaseResource: DaemonSet{Base: Base{def.Meta}, DaemonSet: newDaemonSet, Client: c.DaemonSets()}}
}

// NewExisting returns DaemonSets controller for existing resource by its name
func (d daemonSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingDaemonSet(name, c.DaemonSets())
}

func daemonSetKey(name string) string {
	return "daemonset/" + name
}

func daemonSetStatus(d v1beta1.DaemonSetInterface, name string) (interfaces.ResourceStatus, error) {
	daemonSet, err := d.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}
	if daemonSet.Status.CurrentNumberScheduled == daemonSet.Status.DesiredNumberScheduled {
		return interfaces.ResourceReady, nil
	}
	return interfaces.ResourceNotReady, nil
}

// Key return DaemonSet name
func (d DaemonSet) Key() string {
	return daemonSetKey(d.DaemonSet.Name)
}

// Status returns DaemonSet status. interfaces.ResourceReady means that its dependencies can be created
func (d DaemonSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return daemonSetStatus(d.Client, d.DaemonSet.Name)
}

// Create looks for DaemonSet in k8s and creates it if not present
func (d DaemonSet) Create() error {
	if err := checkExistence(d); err != nil {
		log.Println("Creating", d.Key())
		d.DaemonSet, err = d.Client.Create(d.DaemonSet)
		return err
	}
	return nil
}

// Delete deletes DaemonSet from the cluster
func (d DaemonSet) Delete() error {
	return d.Client.Delete(d.DaemonSet.Name, &v1.DeleteOptions{})
}

// ExistingDaemonSet is a wrapper for K8s DaemonSet object which is deployed on a cluster before AppController
type ExistingDaemonSet struct {
	Base
	Name   string
	Client v1beta1.DaemonSetInterface
}

// Key returns DaemonSet name
func (d ExistingDaemonSet) Key() string {
	return daemonSetKey(d.Name)
}

// Status returns DaemonSet status. interfaces.ResourceReady means that its dependencies can be created
func (d ExistingDaemonSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return daemonSetStatus(d.Client, d.Name)
}

// Create looks for existing DaemonSet and returns error if there is no such DaemonSet
func (d ExistingDaemonSet) Create() error {
	return createExistingResource(d)
}

// Delete deletes DaemonSet from the cluster
func (d ExistingDaemonSet) Delete() error {
	return d.Client.Delete(d.Name, nil)
}

// NewExistingDaemonSet is a constructor
func NewExistingDaemonSet(name string, client v1beta1.DaemonSetInterface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingDaemonSet{Name: name, Client: client}}
}
