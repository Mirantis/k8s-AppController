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

var secretParamFields = []string{
	"Data.Keys",
	"StringData.Keys",
}

type newSecret struct {
	secret *v1.Secret
	client corev1.SecretInterface
}

type existingSecret struct {
	name   string
	client corev1.SecretInterface
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
	secret := parametrizeResource(def.Secret, gc, secretParamFields).(*v1.Secret)
	return newSecret{secret: secret, client: c.Secrets()}
}

// NewExisting returns Secret controller for existing resource by its name
func (secretTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingSecret{name: name, client: c.Secrets()}
}

func secretKey(name string) string {
	return "secret/" + name
}

// Key returns the Secret object identifier
func (s newSecret) Key() string {
	return secretKey(s.secret.Name)
}

// Key returns the Secret object identifier
func (s existingSecret) Key() string {
	return secretKey(s.name)
}

func secretProgress(s corev1.SecretInterface, name string) (float32, error) {
	_, err := s.Get(name)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// GetProgress returns Secret deployment progress
func (s newSecret) GetProgress() (float32, error) {
	return secretProgress(s.client, s.secret.Name)
}

// Create looks for the Secret in k8s and creates it if not present
func (s newSecret) Create() error {
	if checkExistence(s) {
		return nil
	}
	log.Println("Creating", s.Key())
	obj, err := s.client.Create(s.secret)
	s.secret = obj
	return err
}

// Delete deletes Secret from the cluster
func (s newSecret) Delete() error {
	return s.client.Delete(s.secret.Name, nil)
}

// GetProgress returns Secret deployment progress
func (s existingSecret) GetProgress() (float32, error) {
	return secretProgress(s.client, s.name)
}

// Create looks for existing Secret and returns error if there is no such Secret
func (s existingSecret) Create() error {
	return createExistingResource(s)
}

// Delete deletes Secret from the cluster
func (s existingSecret) Delete() error {
	return s.client.Delete(s.name, nil)
}
