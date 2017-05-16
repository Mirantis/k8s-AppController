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

// fakeResDef implements ResourceDefinitionsInterface
type fakeResDef struct {
	fake *testing.Fake
	ns   string
}

var resdefResource = unversioned.GroupVersionResource{
	Group:    "appcontroller.k8s",
	Version:  "v1alpha1",
	Resource: "definitions",
}

// Create creates new Definition object in fake K8s
func (c *fakeResDef) Create(resDef *client.ResourceDefinition) (result *client.ResourceDefinition, err error) {
	obj, err := c.fake.
		Invokes(testing.NewCreateAction(resdefResource, c.ns, resDef), &client.ResourceDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.ResourceDefinition), err
}

// Delete deletes Definition object from fake K8s
func (c *fakeResDef) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.fake.
		Invokes(testing.NewDeleteAction(resdefResource, c.ns, name), &client.ResourceDefinition{})

	return err
}

// List returns Definition objects stored in fake K8s
func (c *fakeResDef) List(opts api.ListOptions) (result *client.ResourceDefinitionList, err error) {
	obj, err := c.fake.
		Invokes(testing.NewListAction(resdefResource, c.ns, opts), &client.ResourceDefinitionList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &client.ResourceDefinitionList{}
	for _, item := range obj.(*client.ResourceDefinitionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}
