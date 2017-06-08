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
	"reflect"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var secretParamFields = []string{
	"Data.Keys",
	"StringData.Keys",
}

type newSecret struct {
	Base
	Secret *v1.Secret
	Client corev1.SecretInterface
}

type existingSecret struct {
	Base
	Name   string
	Client corev1.SecretInterface
}

type secretTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a secret
func (secretTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Secret == nil {
		return ""
	}
	return definition.Secret.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (secretTemplateFactory) Kind() string {
	return "secret"
}

// New returns Secret controller for new resource based on resource definition
func (secretTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	def.Secret = parametrizeResource(def.Secret, gc, secretParamFields).(*v1.Secret)
	return createNewSecret(def, c.Secrets())
}

func createNewSecret(def client.ResourceDefinition, c corev1.SecretInterface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: newSecret{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			Secret: def.Secret,
			Client: c,
		},
	}
}

// NewExisting returns Secret controller for existing resource by its name
func (secretTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingSecret{Name: name, Client: c.Secrets()}}
}

func secretKey(name string) string {
	return "secret/" + name
}

// Key returns the Secret object identifier
func (s newSecret) Key() string {
	return secretKey(s.Secret.Name)
}

// Key returns the Secret object identifier
func (s existingSecret) Key() string {
	return secretKey(s.Name)
}

// Status returns interfaces.ResourceReady if the secret is available in cluster
func (s newSecret) Status(_ map[string]string) (interfaces.ResourceStatus, error) {
	secret, err := s.Client.Get(s.Secret.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !s.equalsToDefinition(secret) {
		return interfaces.ResourceWaitingForUpgrade, nil
	}

	return interfaces.ResourceReady, nil
}

// equalsToDefinition checks if definition in object is compatible with provided object
func (s newSecret) equalsToDefinition(secretiface interface{}) bool {
	secret := secretiface.(*v1.Secret)

	return reflect.DeepEqual(secret.ObjectMeta, s.Secret.ObjectMeta) && reflect.DeepEqual(secret.Data, s.Secret.Data) && reflect.DeepEqual(secret.StringData, s.Secret.StringData) && secret.Type == s.Secret.Type
}

// Create looks for the Secret in k8s and creates it if not present
func (s newSecret) Create() error {
	if err := checkExistence(s); err != nil {
		log.Println("Creating", s.Key())
		s.Secret, err = s.Client.Create(s.Secret)
		return err
	}
	return nil
}

// UpdateFromDefinition updates k8s object with the definition contents
func (s newSecret) UpdateFromDefinition() (err error) {
	s.Secret, err = s.Client.Update(s.Definition.Secret)
	return err
}

// Delete deletes Secret from the cluster
func (s newSecret) Delete() error {
	return s.Client.Delete(s.Secret.Name, nil)
}

// Status returns interfaces.ResourceReady if the secret is available in cluster
func (s existingSecret) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	_, err := s.Client.Get(s.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return interfaces.ResourceReady, nil
}

// Create looks for existing Secret and returns error if there is no such Secret
func (s existingSecret) Create() error {
	return createExistingResource(s)
}

// Delete deletes Secret from the cluster
func (s existingSecret) Delete() error {
	return s.Client.Delete(s.Name, nil)
}
