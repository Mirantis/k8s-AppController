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
	"fmt"
	"log"
	"reflect"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var replicaSetParamFields = []string{
	"Spec.Template.Spec.Containers.Name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.Name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

const successFactorKey = "success_factor"

type newReplicaSet struct {
	Base
	ReplicaSet *extbeta1.ReplicaSet
	Client     v1beta1.ReplicaSetInterface
}

type replicaSetTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a replicaset
func (replicaSetTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.ReplicaSet == nil {
		return ""
	}
	return definition.ReplicaSet.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (replicaSetTemplateFactory) Kind() string {
	return "replicaset"
}

// New returns ReplicaSet controller for new resource based on resource definition
func (replicaSetTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	def.ReplicaSet = parametrizeResource(def.ReplicaSet, gc, replicaSetParamFields).(*extbeta1.ReplicaSet)
	return createNewReplicaSet(def, c.ReplicaSets())
}

// NewExisting returns ReplicaSet controller for existing resource by its name
func (replicaSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingReplicaSet{Name: name, Client: c.ReplicaSets()}
}

func replicaSetStatus(rs *extbeta1.ReplicaSet, meta map[string]string) (interfaces.ResourceStatus, error) {
	successFactor, err := getPercentage(successFactorKey, meta)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if rs.Status.Replicas*100 < *rs.Spec.Replicas*successFactor {
		return interfaces.ResourceNotReady, nil
	}

	return interfaces.ResourceReady, nil
}

func replicaSetReport(r v1beta1.ReplicaSetInterface, name string, meta map[string]string) interfaces.DependencyReport {
	rs, err := r.Get(name)
	if err != nil {
		return report.ErrorReport(name, err)
	}
	successFactor, err := getPercentage(successFactorKey, meta)
	if err != nil {
		return report.ErrorReport(name, err)
	}
	percentage := (*rs.Spec.Replicas * 100 / rs.Status.Replicas)
	message := fmt.Sprintf(
		"%d of %d replicas up (%d %%, needed %d%%)",
		rs.Status.Replicas,
		rs.Spec.Replicas,
		percentage,
		successFactor,
	)
	if percentage >= successFactor {
		return interfaces.DependencyReport{
			Dependency: name,
			Blocks:     false,
			Percentage: int(percentage),
			Needed:     int(successFactor),
			Message:    message,
		}
	}
	return interfaces.DependencyReport{
		Dependency: name,
		Blocks:     false,
		Percentage: int(percentage),
		Needed:     int(successFactor),
		Message:    message,
	}
}

func replicaSetKey(name string) string {
	return "replicaset/" + name
}

// Key returns ReplicaSet name
func (r newReplicaSet) Key() string {
	return replicaSetKey(r.ReplicaSet.Name)
}

// Create looks for the ReplicaSet in k8s and creates it if not present
func (r newReplicaSet) Create() error {
	if err := checkExistence(r); err != nil {
		log.Println("Creating", r.Key())
		r.ReplicaSet, err = r.Client.Create(r.ReplicaSet)
		return err
	}
	return nil
}

// Delete deletes ReplicaSet from the cluster
func (r newReplicaSet) Delete() error {
	return r.Client.Delete(r.ReplicaSet.Name, nil)
}

// GetDependencyReport returns a DependencyReport for this ReplicaSet
func (r newReplicaSet) GetDependencyReport(meta map[string]string) interfaces.DependencyReport {
	return replicaSetReport(r.Client, r.ReplicaSet.Name, meta)
}

// Status returns ReplicaSet status based on provided meta.
func (r newReplicaSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	rs, err := r.Client.Get(r.ReplicaSet.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !r.EqualToDefinition(rs) {
		return interfaces.ResourceWaitingForUpgrade, fmt.Errorf(string(interfaces.ResourceWaitingForUpgrade))
	}
	return replicaSetStatus(rs, meta)
}

// EqualToDefinition checks if definition in object is compatible with provided object
func (r newReplicaSet) EqualToDefinition(replicaSetiface interface{}) bool {
	replicaSet := replicaSetiface.(*extbeta1.ReplicaSet)

	return reflect.DeepEqual(replicaSet.ObjectMeta, r.ReplicaSet.ObjectMeta) && reflect.DeepEqual(replicaSet.Spec, r.ReplicaSet.Spec)
}

// StatusIsCacheable returns false if meta contains SuccessFactorKey
func (r newReplicaSet) StatusIsCacheable(meta map[string]string) bool {
	_, ok := meta[successFactorKey]
	return !ok
}

func createNewReplicaSet(def client.ResourceDefinition, client v1beta1.ReplicaSetInterface) newReplicaSet {
	return newReplicaSet{
		Base: Base{
			Definition: def,
			meta:       def.Meta,
		},
		ReplicaSet: def.ReplicaSet,
		Client:     client,
	}
}

type existingReplicaSet struct {
	Base
	Name   string
	Client v1beta1.ReplicaSetInterface
}

// Key returns ReplicaSet name
func (r existingReplicaSet) Key() string {
	return replicaSetKey(r.Name)
}

// Create looks for existing ReplicaSet and returns error if there is no such ReplicaSet
func (r existingReplicaSet) Create() error {
	return createExistingResource(r)
}

// Status returns ReplicaSet status based on provided meta.
func (r existingReplicaSet) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	rs, err := r.Client.Get(r.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}
	return replicaSetStatus(rs, meta)
}

// Delete deletes ReplicaSet from the cluster
func (r existingReplicaSet) Delete() error {
	return r.Client.Delete(r.Name, nil)
}

// GetDependencyReport returns a DependencyReport for this ReplicaSet
func (r existingReplicaSet) GetDependencyReport(meta map[string]string) interfaces.DependencyReport {
	return replicaSetReport(r.Client, r.Name, meta)
}
