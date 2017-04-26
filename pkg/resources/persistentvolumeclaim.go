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

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var persistentVolumeClaimParamFields = []string{
	"Spec",
}

type newPersistentVolumeClaim struct {
	Base
	PersistentVolumeClaim *v1.PersistentVolumeClaim
	Client                corev1.PersistentVolumeClaimInterface
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
	def.PersistentVolumeClaim = parametrizeResource(def.PersistentVolumeClaim, gc, persistentVolumeClaimParamFields).(*v1.PersistentVolumeClaim)
	return createNewPersistentVolumeClaim(def, c.PersistentVolumeClaims())
}

func createNewPersistentVolumeClaim(def client.ResourceDefinition, c corev1.PersistentVolumeClaimInterface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: newPersistentVolumeClaim{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			PersistentVolumeClaim: def.PersistentVolumeClaim,
			Client:                c,
		},
	}

}

// NewExisting returns PVC controller for existing resource by its name
func (persistentVolumeClaimTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingPersistentVolumeClaim{Name: name, Client: c.PersistentVolumeClaims()}}
}

func persistentVolumeClaimKey(name string) string {
	return "persistentvolumeclaim/" + name
}

// Key returns the PersistentVolumeClaim object identifier
func (p newPersistentVolumeClaim) Key() string {
	return persistentVolumeClaimKey(p.PersistentVolumeClaim.Name)
}

func persistentVolumeClaimStatus(persistentVolumeClaim *v1.PersistentVolumeClaim) (interfaces.ResourceStatus, error) {
	if persistentVolumeClaim.Status.Phase == v1.ClaimBound {
		return interfaces.ResourceReady, nil
	}

	return interfaces.ResourceNotReady, nil
}

// Create looks for the PersistentVolumeClaim in k8s and creates it if not present
func (p newPersistentVolumeClaim) Create() error {
	if err := checkExistence(p); err != nil {
		log.Println("Creating", p.Key())
		p.PersistentVolumeClaim, err = p.Client.Create(p.PersistentVolumeClaim)
		return err
	}
	return nil
}

// Delete deletes persistentVolumeClaim from the cluster
func (p newPersistentVolumeClaim) Delete() error {
	return p.Client.Delete(p.PersistentVolumeClaim.Name, &v1.DeleteOptions{})
}

type existingPersistentVolumeClaim struct {
	Base
	Name   string
	Client corev1.PersistentVolumeClaimInterface
}

// Status returns PVC status.
func (p newPersistentVolumeClaim) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	pvc, err := p.Client.Get(p.PersistentVolumeClaim.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !p.EqualToDefinition(pvc) {
		return interfaces.ResourceWaitingForUpgrade, fmt.Errorf(string(interfaces.ResourceWaitingForUpgrade))
	}

	return persistentVolumeClaimStatus(pvc)
}

// EqualToDefinition returns whether the resource has the same values as provided object
func (p newPersistentVolumeClaim) EqualToDefinition(claim interface{}) bool {
	pvc := claim.(*v1.PersistentVolumeClaim)

	return reflect.DeepEqual(pvc.ObjectMeta, p.PersistentVolumeClaim.ObjectMeta) && reflect.DeepEqual(pvc.Spec, p.PersistentVolumeClaim.Spec)
}

// Key returns the PersistentVolumeClaim object identifier
func (p existingPersistentVolumeClaim) Key() string {
	return persistentVolumeClaimKey(p.Name)
}

// Create looks for existing PVC and returns error if there is no such PVC
func (p existingPersistentVolumeClaim) Create() error {
	return createExistingResource(p)
}

// Status returns PVC status.
func (p existingPersistentVolumeClaim) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	pvc, err := p.Client.Get(p.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return persistentVolumeClaimStatus(pvc)
}

// Delete deletes persistentVolumeClaim from the cluster
func (p existingPersistentVolumeClaim) Delete() error {
	return p.Client.Delete(p.Name, nil)
}
