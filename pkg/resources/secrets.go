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
	"github.com/Mirantis/k8s-AppController/pkg/report"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
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

type secretTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a secret
func (secretTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Secret == nil {
		return ""
	}
	return getObjectName(definition.Secret)
}

// k8s resource kind that this fabric supports
func (secretTemplateFactory) Kind() string {
	return "secret"
}

func (secretTemplateFactory) New(def client.ResourceDefinition, ci client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewSecret(parametrizeResource(def.Secret, gc).(*v1.Secret), ci.Secrets(), def.Meta)
}

func (secretTemplateFactory) NewExisting(name string, ci client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return NewExistingSecret(name, ci.Secrets())
}

func secretKey(name string) string {
	return "secret/" + name
}

func (s Secret) Key() string {
	return secretKey(getObjectName(s.Secret))
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

func (s *Secret) Create() error {
	if err := checkExistence(s); err != nil {
		log.Println("Creating", s.Key())
		s.Secret, err = s.Client.Create(s.Secret)
		return err
	}
	return nil
}

func (s Secret) Delete() error {
	return s.Client.Delete(s.Secret.Name, nil)
}

func NewSecret(s *v1.Secret, client corev1.SecretInterface, meta map[string]interface{}) interfaces.Resource {
	return report.SimpleReporter{BaseResource: &Secret{Base: Base{meta}, Secret: s, Client: client}}
}

func NewExistingSecret(name string, client corev1.SecretInterface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingSecret{Name: name, Client: client}}
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
