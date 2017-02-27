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

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

type Secret struct {
	Base
	Secret *v1.Secret
	Client corev1.SecretInterface
}

type ExistingSecret struct {
	Base
	Name   string
	Client corev1.SecretInterface
}

func secretKey(name string) string {
	return "secret/" + name
}

func (s Secret) Key() string {
	return secretKey(s.Secret.Name)
}

func (s ExistingSecret) Key() string {
	return secretKey(s.Name)
}

func secretStatus(s corev1.SecretInterface, name string) (interfaces.ResourceStatus, error) {
	_, err := s.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return interfaces.ResourceReady, nil
}

// Status returns interfaces.ResourceReady if the secret is available in cluster
func (s Secret) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return secretStatus(s.Client, s.Secret.Name)
}

func (s Secret) Create() error {
	if err := checkExistence(s); err != nil {
		log.Println("Creating ", s.Key())
		s.Secret, err = s.Client.Create(s.Secret)
		return err
	}
	return nil
}

func (s Secret) Delete() error {
	return s.Client.Delete(s.Secret.Name, nil)
}

func (s Secret) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Secret != nil && def.Secret.Name == name
}

func NewSecret(s *v1.Secret, client corev1.SecretInterface, meta map[string]interface{}) interfaces.Resource {
	return report.SimpleReporter{BaseResource: Secret{Base: Base{meta}, Secret: s, Client: client}}
}

func NewExistingSecret(name string, client corev1.SecretInterface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingSecret{Name: name, Client: client}}
}

func (s Secret) New(def client.ResourceDefinition, ci client.Interface) interfaces.Resource {
	return NewSecret(def.Secret, ci.Secrets(), def.Meta)
}

func (s Secret) NewExisting(name string, ci client.Interface) interfaces.Resource {
	return NewExistingSecret(name, ci.Secrets())
}

// Status returns interfaces.ResourceReady if the secret is available in cluster
func (s ExistingSecret) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return secretStatus(s.Client, s.Name)
}

func (s ExistingSecret) Create() error {
	return createExistingResource(s)
}

func (s ExistingSecret) Delete() error {
	return s.Client.Delete(s.Name, nil)
}
