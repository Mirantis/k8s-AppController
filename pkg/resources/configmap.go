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

var configMapParamFields = []string{
	"Data.Keys",
}

type newConfigMap struct {
	Base
	ConfigMap *v1.ConfigMap
	Client    corev1.ConfigMapInterface
}

type existingConfigMap struct {
	Base
	Name   string
	Client corev1.ConfigMapInterface
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
	return report.SimpleReporter{BaseResource: newConfigMap{Base: Base{def.Meta}, ConfigMap: cm, Client: c.ConfigMaps()}}
}

// NewExisting returns configMap controller for existing resource by its name
func (configMapTemplateFactory) NewExisting(name string, ci client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingConfigMap{Name: name, Client: ci.ConfigMaps()}}
}

func configMapKey(name string) string {
	return "configmap/" + name
}

// Key returns the configMap object identifier
func (c newConfigMap) Key() string {
	return configMapKey(c.ConfigMap.Name)
}

func configMapStatus(c corev1.ConfigMapInterface, name string) (interfaces.ResourceStatus, error) {
	_, err := c.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return interfaces.ResourceReady, nil
}

// Status returns configMap status. interfaces.ResourceReady means that its dependencies can be created
func (c newConfigMap) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return configMapStatus(c.Client, c.ConfigMap.Name)
}

// Create looks for DaemonSet in k8s and creates it if not present
func (c newConfigMap) Create() error {
	if err := checkExistence(c); err != nil {
		log.Println("Creating", c.Key())
		c.ConfigMap, err = c.Client.Create(c.ConfigMap)
		return err
	}
	return nil
}

// Delete deletes configMap from the cluster
func (c newConfigMap) Delete() error {
	return c.Client.Delete(c.ConfigMap.Name, &v1.DeleteOptions{})
}

// Key returns the configMap object identifier
func (c existingConfigMap) Key() string {
	return configMapKey(c.Name)
}

// Status returns configMap status. interfaces.ResourceReady means that its dependencies can be created
func (c existingConfigMap) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return configMapStatus(c.Client, c.Name)
}

// Create looks for existing configMap and returns an error if there is no such configMap in a cluster
func (c existingConfigMap) Create() error {
	return createExistingResource(c)
}

// Delete deletes configMap from the cluster
func (c existingConfigMap) Delete() error {
	return c.Client.Delete(c.Name, nil)
}
