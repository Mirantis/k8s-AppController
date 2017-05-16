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

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

// MakeConfigMap generates sample ConfigMap object
func MakeConfigMap(name string) *v1.ConfigMap {
	configMap := &v1.ConfigMap{}
	configMap.Name = name
	configMap.Namespace = "testing"
	return configMap
}

// ConfigMaps generates sample ConfigMapList collection object
func ConfigMaps(names ...string) runtime.Object {
	var configMaps []v1.ConfigMap
	for i := 0; i < 3; i++ {
		configMaps = append(configMaps, *MakeConfigMap(fmt.Sprintf("cfgmap-%d", i)))
	}
	for _, name := range names {
		configMaps = append(configMaps, *MakeConfigMap(name))
	}
	return &v1.ConfigMapList{Items: configMaps}
}
