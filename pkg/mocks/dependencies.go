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
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"k8s.io/client-go/pkg/api"
)

type Dependency struct {
	Parent string
	Child  string
	Meta   map[string]string
}

type dependencyClient struct {
	dependencies []Dependency
}

func (d *dependencyClient) List(opts api.ListOptions) (*client.DependencyList, error) {
	list := &client.DependencyList{}

	for _, dep := range d.dependencies {
		list.Items = append(
			list.Items,
			client.Dependency{
				Parent: dep.Parent,
				Child:  dep.Child,
				Meta:   make(map[string]string),
			},
		)
	}

	return list, nil
}

func NewDependencyClient(dependencies ...Dependency) client.DependenciesInterface {
	return &dependencyClient{dependencies}
}
