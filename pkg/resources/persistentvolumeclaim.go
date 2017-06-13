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

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var persistentVolumeClaimParamFields = []string{
	"Spec",
}

type newPersistentVolumeClaim struct {
	persistentVolumeClaim *v1.PersistentVolumeClaim
	client                corev1.PersistentVolumeClaimInterface
}

type existingPersistentVolumeClaim struct {
	name   string
	client corev1.PersistentVolumeClaimInterface
}

type persistentVolumeClaimTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a persistentvolumeclaim
func (persistentVolumeClaimTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.PersistentVolumeClaim == nil {
		return ""
	}
	return definition.PersistentVolumeClaim.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (persistentVolumeClaimTemplateFactory) Kind() string {
	return "persistentvolumeclaim"
}

// New returns PVC controller for new resource based on resource definition
func (persistentVolumeClaimTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	pvc := parametrizeResource(def.PersistentVolumeClaim, gc, persistentVolumeClaimParamFields).(*v1.PersistentVolumeClaim)
	return newPersistentVolumeClaim{
		persistentVolumeClaim: pvc,
		client:                c.PersistentVolumeClaims(),
	}

}

// NewExisting returns PVC controller for existing resource by its name
func (persistentVolumeClaimTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingPersistentVolumeClaim{name: name, client: c.PersistentVolumeClaims()}
}

func persistentVolumeClaimKey(name string) string {
	return "persistentvolumeclaim/" + name
}

// Key returns the PersistentVolumeClaim object identifier
func (p newPersistentVolumeClaim) Key() string {
	return persistentVolumeClaimKey(p.persistentVolumeClaim.Name)
}

func persistentVolumeClaimProgress(p corev1.PersistentVolumeClaimInterface, name string) (float32, error) {
	persistentVolumeClaim, err := p.Get(name)
	if err != nil {
		return 0, err
	}

	if persistentVolumeClaim.Status.Phase == v1.ClaimBound {
		return 1, nil
	}

	return 0, nil
}

// Create looks for the PersistentVolumeClaim in k8s and creates it if not present
func (p newPersistentVolumeClaim) Create() error {
	if checkExistence(p) {
		return nil
	}
	log.Println("Creating", p.Key())
	obj, err := p.client.Create(p.persistentVolumeClaim)
	p.persistentVolumeClaim = obj
	return err
}

// Delete deletes persistentVolumeClaim from the cluster
func (p newPersistentVolumeClaim) Delete() error {
	return p.client.Delete(p.persistentVolumeClaim.Name, &v1.DeleteOptions{})
}

// GetProgress returns PVC deployment progress.
func (p newPersistentVolumeClaim) GetProgress() (float32, error) {
	return persistentVolumeClaimProgress(p.client, p.persistentVolumeClaim.Name)
}

// Key returns the PersistentVolumeClaim object identifier
func (p existingPersistentVolumeClaim) Key() string {
	return persistentVolumeClaimKey(p.name)
}

// Create looks for existing PVC and returns error if there is no such PVC
func (p existingPersistentVolumeClaim) Create() error {
	return createExistingResource(p)
}

// GetProgress returns PVC deployment progress
func (p existingPersistentVolumeClaim) GetProgress() (float32, error) {
	return persistentVolumeClaimProgress(p.client, p.name)
}

// Delete deletes persistentVolumeClaim from the cluster
func (p existingPersistentVolumeClaim) Delete() error {
	return p.client.Delete(p.name, nil)
}
