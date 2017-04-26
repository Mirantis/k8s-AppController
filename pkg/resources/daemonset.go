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
	"fmt"
	"log"
	"reflect"

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
	def.DaemonSet = parametrizeResource(def.DaemonSet, gc, daemonSetParamFields).(*extbeta1.DaemonSet)
	return createNewDaemonSet(def, c.DaemonSets())
}

// NewExisting returns DaemonSets controller for existing resource by its name
func (d daemonSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingDaemonSet(name, c.DaemonSets())
}

func daemonSetKey(name string) string {
	return "daemonset/" + name
}

func daemonSetStatus(d *extbeta1.DaemonSet) (interfaces.ResourceStatus, error) {
	if d.Status.CurrentNumberScheduled == d.Status.DesiredNumberScheduled {
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
	ds, err := d.Client.Get(d.DaemonSet.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !d.EqualToDefinition(ds) {
		return interfaces.ResourceWaitingForUpgrade, fmt.Errorf(string(interfaces.ResourceWaitingForUpgrade))
	}
	return daemonSetStatus(ds)
}

// EqualToDefinition returns whether the resource has the same values as provided object
func (d DaemonSet) EqualToDefinition(daemonset interface{}) bool {
	ds := daemonset.(*extbeta1.DaemonSet)

	return reflect.DeepEqual(ds.ObjectMeta, d.DaemonSet.ObjectMeta) && reflect.DeepEqual(ds.Spec, d.DaemonSet.Spec)
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

func createNewDaemonSet(def client.ResourceDefinition, client v1beta1.DaemonSetInterface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: DaemonSet{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			DaemonSet: def.DaemonSet,
			Client:    client,
		},
	}
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
	ds, err := d.Client.Get(d.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return daemonSetStatus(ds)
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
