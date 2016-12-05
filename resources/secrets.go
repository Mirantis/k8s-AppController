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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
)

type Secret struct {
	Secret *api.Secret
	Client unversioned.SecretsInterface
}

type ExistingSecret struct {
	Name   string
	Client unversioned.SecretsInterface
	Secret
}

func secretKey(name string) string {
	return "secret/" + name
}

func (c Secret) Key() string {
	return secretKey(c.Secret.Name)
}

func secretStatus(c unversioned.SecretsInterface, name string) (string, error) {
	_, err := c.Get(name)
	if err != nil {
		return "error", err
	}

	return "ready", nil
}

func (c Secret) Status(meta map[string]string) (string, error) {
	return secretStatus(c.Client, c.Secret.Name)
}

func (c Secret) Create() error {
	if err := checkExistence(c); err != nil {
		log.Println("Creating ", c.Key())
		c.Secret, err = c.Client.Create(c.Secret)
		return err
	}
	return nil
}

func (c Secret) Delete() error {
	return c.Client.Delete(c.Secret.Name)
}

func (c Secret) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Secret != nil && def.Secret.Name == name
}

func NewSecret(c *api.Secret, client unversioned.SecretsInterface) Secret {
	return Secret{Secret: c, Client: client}
}

func NewExistingSecret(name string, client unversioned.SecretsInterface) ExistingSecret {
	return ExistingSecret{Name: name, Client: client}
}

func (c Secret) New(def client.ResourceDefinition, ci client.Interface) interfaces.Resource {
	return NewSecret(def.Secret, ci.Secrets())
}

func (c Secret) NewExisting(name string, ci client.Interface) interfaces.Resource {
	return NewExistingSecret(name, ci.Secrets())
}
