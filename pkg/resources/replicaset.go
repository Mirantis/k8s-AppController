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

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var replicaSetParamFields = []string{
	"Spec.Template.Spec.Containers.name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

type newReplicaSet struct {
	replicaSet *extbeta1.ReplicaSet
	client     v1beta1.ReplicaSetInterface
}

type existingReplicaSet struct {
	name   string
	client v1beta1.ReplicaSetInterface
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
	replicaSet := parametrizeResource(def.ReplicaSet, gc, replicaSetParamFields).(*extbeta1.ReplicaSet)
	return createNewReplicaSet(replicaSet, c.ReplicaSets())
}

// NewExisting returns ReplicaSet controller for existing resource by its name
func (replicaSetTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingReplicaSet{name: name, client: c.ReplicaSets()}
}

func replicaSetProgress(r v1beta1.ReplicaSetInterface, name string) (float32, error) {
	rs, err := r.Get(name)
	if err != nil {
		return 0, err
	}

	total := float32(1)
	if rs.Spec.Replicas != nil && *rs.Spec.Replicas != 0 {
		total = float32(*rs.Spec.Replicas)
	}
	return float32(rs.Status.Replicas) / total, nil
}

func replicaSetKey(name string) string {
	return "replicaset/" + name
}

// Key returns ReplicaSet name
func (r newReplicaSet) Key() string {
	return replicaSetKey(r.replicaSet.Name)
}

// Create looks for the ReplicaSet in k8s and creates it if not present
func (r newReplicaSet) Create() error {
	if checkExistence(r) {
		return nil
	}
	log.Println("Creating", r.Key())
	obj, err := r.client.Create(r.replicaSet)
	r.replicaSet = obj
	return err
}

// Delete deletes ReplicaSet from the cluster
func (r newReplicaSet) Delete() error {
	return r.client.Delete(r.replicaSet.Name, nil)
}

// GetProgress returns ReplicaSet deployment progress
func (r newReplicaSet) GetProgress() (float32, error) {
	return replicaSetProgress(r.client, r.replicaSet.Name)
}

func createNewReplicaSet(replicaSet *extbeta1.ReplicaSet, client v1beta1.ReplicaSetInterface) newReplicaSet {
	return newReplicaSet{replicaSet: replicaSet, client: client}
}

// Key returns ReplicaSet name
func (r existingReplicaSet) Key() string {
	return replicaSetKey(r.name)
}

// Create looks for existing ReplicaSet and returns error if there is no such ReplicaSet
func (r existingReplicaSet) Create() error {
	return createExistingResource(r)
}

// GetProgress returns ReplicaSet deployment progress
func (r existingReplicaSet) GetProgress() (float32, error) {
	return replicaSetProgress(r.client, r.name)
}

// Delete deletes ReplicaSet from the cluster
func (r existingReplicaSet) Delete() error {
	return r.client.Delete(r.name, nil)
}
