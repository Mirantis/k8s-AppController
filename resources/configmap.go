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

type ConfigMap struct {
	ConfigMap *api.ConfigMap
	Client    unversioned.ConfigMapsInterface
}

type ExistingConfigMap struct {
	Name   string
	Client unversioned.ConfigMapsInterface
}

func configMapKey(name string) string {
	return "configMap/" + name
}

func (c ConfigMap) Key() string {
	return configMapKey(c.ConfigMap.Name)
}

func configMapStatus(c unversioned.ConfigMapsInterface, name string) (string, error) {
	_, err := c.Get(name)
	if err != nil {
		return "error", err
	}

	return "ready", nil
}

func (c ConfigMap) Status(meta map[string]string) (string, error) {
	return configMapStatus(c.Client, c.ConfigMap.Name)
}

func (c ConfigMap) Create() error {
	if err := checkExistence(c); err != nil {
		log.Println("Creating ", c.Key())
		c.ConfigMap, err = c.Client.Create(c.ConfigMap)
		return err
	}
	return nil
}

func (c ConfigMap) Delete() error {
	return c.Client.Delete(c.ConfigMap.Name)
}

func (c ConfigMap) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.ConfigMap != nil && def.ConfigMap.Name == name
}

func NewConfigMap(c *api.ConfigMap, client unversioned.ConfigMapsInterface) ConfigMap {
	return ConfigMap{ConfigMap: c, Client: client}
}

func NewExistingConfigMap(name string, client unversioned.ConfigMapsInterface) ExistingConfigMap {
	return ExistingConfigMap{Name: name, Client: client}
}

func (c ConfigMap) New(def client.ResourceDefinition, ci client.Interface) interfaces.Resource {
	return NewConfigMap(def.ConfigMap, ci.ConfigMaps())
}

func (c ConfigMap) NewExisting(name string, ci client.Interface) interfaces.Resource {
	return NewExistingConfigMap(name, ci.ConfigMaps())
}

func (c ExistingConfigMap) Key() string {
	return configMapKey(c.Name)
}

func (c ExistingConfigMap) Status(meta map[string]string) (string, error) {
	return configMapStatus(c.Client, c.Name)
}

func (c ExistingConfigMap) Create() error {
	return createExistingResource(c)
}

func (c ExistingConfigMap) Delete() error {
	return c.Client.Delete(c.Name)
}
