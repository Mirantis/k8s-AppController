// Copyright 2017 Mirantis
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
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/testing"
)

type fakeDeps struct {
	fake *testing.Fake
	ns   string
}

var dependencyResource = unversioned.GroupVersionResource{
	Group:    "appcontroller.k8s",
	Version:  "v1alpha1",
	Resource: "dependencies",
}

// Create creates new Dependency object in fake K8s
func (c *fakeDeps) Create(dependency *client.Dependency) (result *client.Dependency, err error) {
	obj, err := c.fake.
		Invokes(testing.NewCreateAction(dependencyResource, c.ns, dependency), &client.Dependency{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.Dependency), err
}

// Delete deletes Dependency from fake K8s
func (c *fakeDeps) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.fake.
		Invokes(testing.NewDeleteAction(dependencyResource, c.ns, name), &client.Dependency{})

	return err
}

// List returns Dependency objects stored in fake K8s
func (c *fakeDeps) List(opts api.ListOptions) (result *client.DependencyList, err error) {
	obj, err := c.fake.
		Invokes(testing.NewListAction(dependencyResource, c.ns, opts), &client.DependencyList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &client.DependencyList{}
	for _, item := range obj.(*client.DependencyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}
