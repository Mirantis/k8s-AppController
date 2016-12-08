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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type configMapClient struct {
}

func MakeConfigMap(name string) *api.ConfigMap {

	configMap := &api.ConfigMap{}
	configMap.Name = name

	return configMap
}

func (p *configMapClient) List(opts api.ListOptions) (*api.ConfigMapList, error) {
	var configMaps []api.ConfigMap
	for i := 0; i < 3; i++ {
		configMaps = append(configMaps, *MakeConfigMap(fmt.Sprintf("cfgmap-%d", i)))
	}

	return &api.ConfigMapList{Items: configMaps}, nil
}

func (p *configMapClient) Get(name string) (*api.ConfigMap, error) {
	if name != "fail" {
		return MakeConfigMap(name), nil
	}
	return MakeConfigMap(name), fmt.Errorf("Mock ConfigMap not created")
}

func (p *configMapClient) Delete(name string) error {
	panic("not implemented")
}

func (p *configMapClient) Create(configMap *api.ConfigMap) (*api.ConfigMap, error) {
	return MakeConfigMap(configMap.Name), nil
}

func (p *configMapClient) Update(configMap *api.ConfigMap) (*api.ConfigMap, error) {
	panic("not implemented")
}

func (p *configMapClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *configMapClient) Bind(binding *api.Binding) error {
	panic("not implemented")
}

func NewConfigMapClient() unversioned.ConfigMapsInterface {
	return &configMapClient{}
}
