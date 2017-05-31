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

var configMapParamFields = []string{
	"Data.Keys",
}

type newConfigMap struct {
	configMap *v1.ConfigMap
	client    corev1.ConfigMapInterface
}

type existingConfigMap struct {
	name   string
	client corev1.ConfigMapInterface
}

type configMapTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a configmap
func (configMapTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.ConfigMap == nil {
		return ""
	}
	return definition.ConfigMap.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (configMapTemplateFactory) Kind() string {
	return "configmap"
}

// New returns configMap controller for new resource based on resource definition
func (configMapTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	cm := parametrizeResource(def.ConfigMap, gc, configMapParamFields).(*v1.ConfigMap)
	return newConfigMap{configMap: cm, client: c.ConfigMaps()}
}

// NewExisting returns configMap controller for existing resource by its name
func (configMapTemplateFactory) NewExisting(name string, ci client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingConfigMap{name: name, client: ci.ConfigMaps()}
}

func configMapKey(name string) string {
	return "configmap/" + name
}

// Key returns the configMap object identifier
func (c newConfigMap) Key() string {
	return configMapKey(c.configMap.Name)
}

func configMapProgress(c corev1.ConfigMapInterface, name string) (float32, error) {
	_, err := c.Get(name)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

// GetProgress returns configMap deployment progress
func (c newConfigMap) GetProgress() (float32, error) {
	return configMapProgress(c.client, c.configMap.Name)
}

// Create looks for DaemonSet in k8s and creates it if not present
func (c newConfigMap) Create() error {
	if checkExistence(c) {
		return nil
	}
	log.Println("Creating", c.Key())
	obj, err := c.client.Create(c.configMap)
	c.configMap = obj
	return err
}

// Delete deletes configMap from the cluster
func (c newConfigMap) Delete() error {
	return c.client.Delete(c.configMap.Name, &v1.DeleteOptions{})
}

// Key returns the configMap object identifier
func (c existingConfigMap) Key() string {
	return configMapKey(c.name)
}

// GetProgress returns configMap deployment progress
func (c existingConfigMap) GetProgress() (float32, error) {
	return configMapProgress(c.client, c.name)
}

// Create looks for existing configMap and returns an error if there is no such configMap in a cluster
func (c existingConfigMap) Create() error {
	return createExistingResource(c)
}

// Delete deletes configMap from the cluster
func (c existingConfigMap) Delete() error {
	return c.client.Delete(c.name, nil)
}
