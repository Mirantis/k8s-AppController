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

package mocks

import (
	"fmt"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/watch"
)

type configMapClient struct {
}

func MakeConfigMap(name string) *v1.ConfigMap {

	configMap := &v1.ConfigMap{}
	configMap.Name = name

	return configMap
}

func (p *configMapClient) List(opts v1.ListOptions) (*v1.ConfigMapList, error) {
	var configMaps []v1.ConfigMap
	for i := 0; i < 3; i++ {
		configMaps = append(configMaps, *MakeConfigMap(fmt.Sprintf("cfgmap-%d", i)))
	}

	return &v1.ConfigMapList{Items: configMaps}, nil
}

func (p *configMapClient) Get(name string) (*v1.ConfigMap, error) {
	if name != "fail" {
		return MakeConfigMap(name), nil
	}
	return MakeConfigMap(name), fmt.Errorf("Mock ConfigMap not created")
}

func (p *configMapClient) Delete(name string, opts *v1.DeleteOptions) error {
	panic("not implemented")
}

func (p *configMapClient) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	panic("not implemented")
}

func (p *configMapClient) Create(configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	return MakeConfigMap(configMap.Name), nil
}

func (p *configMapClient) Update(configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	panic("not implemented")
}

func (p *configMapClient) Watch(opts v1.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *configMapClient) Bind(binding *api.Binding) error {
	panic("not implemented")
}

func (p *configMapClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.ConfigMap, err error) {
	panic("not implemented")
}

func NewConfigMapClient() corev1.ConfigMapInterface {
	return &configMapClient{}
}
