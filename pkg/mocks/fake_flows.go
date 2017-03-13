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

// FakeFlows implements FlowsInterface
type FakeFlows struct {
	fake *testing.Fake
	ns   string
}

var flowResource = unversioned.GroupVersionResource{
	Group:    "appcontroller.k8s",
	Version:  "v1alpha1",
	Resource: "flows",
}

func (c *FakeFlows) Create(flow *client.Flow) (result *client.Flow, err error) {
	obj, err := c.fake.
		Invokes(testing.NewCreateAction(flowResource, c.ns, flow), &client.Flow{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.Flow), err
}

func (c *FakeFlows) Update(flow *client.Flow) (result *client.Flow, err error) {
	obj, err := c.fake.
		Invokes(testing.NewUpdateAction(flowResource, c.ns, flow), &client.Flow{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.Flow), err
}

func (c *FakeFlows) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.fake.
		Invokes(testing.NewDeleteAction(flowResource, c.ns, name), &client.Flow{})

	return err
}

func (c *FakeFlows) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := testing.NewDeleteCollectionAction(flowResource, c.ns, listOptions)

	_, err := c.fake.Invokes(action, &client.FlowList{})
	return err
}

func (c *FakeFlows) Get(name string) (result *client.Flow, err error) {
	obj, err := c.fake.
		Invokes(testing.NewGetAction(flowResource, c.ns, name), &client.Flow{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.Flow), err
}

func (c *FakeFlows) List(opts api.ListOptions) (result *client.FlowList, err error) {
	obj, err := c.fake.
		Invokes(testing.NewListAction(flowResource, c.ns, opts), &client.FlowList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &client.FlowList{}
	for _, item := range obj.(*client.FlowList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}
