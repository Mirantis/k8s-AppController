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
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/pkg/types"
	"k8s.io/client-go/testing"
)

// FakeFlowDeployments implements FlowDeploymentsInterface
type FakeFlowDeployments struct {
	fake *testing.Fake
	ns   string
}

var flowDeploymentResource = unversioned.GroupVersionResource{
	Group:    "appcontroller.k8s",
	Version:  "v1alpha1",
	Resource: "deployments",
}

func (fd *FakeFlowDeployments) Create(flowDeployment *client.FlowDeployment) (result *client.FlowDeployment, err error) {
	obj, err := fd.fake.
		Invokes(testing.NewCreateAction(flowDeploymentResource, fd.ns, flowDeployment), &client.FlowDeployment{})

	if obj == nil {
		return nil, err
	}
	res := obj.(*client.FlowDeployment)
	res.SetCreationTimestamp(unversioned.Time{time.Now()})
	uuid, _ := newUUID()
	res.SetUID(types.UID(uuid))

	return res, err
}

func (fd *FakeFlowDeployments) Update(flowDeployment *client.FlowDeployment) (result *client.FlowDeployment, err error) {
	obj, err := fd.fake.
		Invokes(testing.NewUpdateAction(flowDeploymentResource, fd.ns, flowDeployment), &client.FlowDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.FlowDeployment), err
}

func (fd *FakeFlowDeployments) Delete(name string, options *api.DeleteOptions) error {
	_, err := fd.fake.
		Invokes(testing.NewDeleteAction(flowDeploymentResource, fd.ns, name), &client.FlowDeployment{})

	return err
}

func (fd *FakeFlowDeployments) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := testing.NewDeleteCollectionAction(flowDeploymentResource, fd.ns, listOptions)

	_, err := fd.fake.Invokes(action, &client.FlowDeployment{})
	return err
}

func (fd *FakeFlowDeployments) Get(name string) (result *client.FlowDeployment, err error) {
	obj, err := fd.fake.
		Invokes(testing.NewGetAction(flowDeploymentResource, fd.ns, name), &client.FlowDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.FlowDeployment), err
}

func (fd *FakeFlowDeployments) List(opts api.ListOptions) (result *client.FlowDeploymentList, err error) {
	obj, err := fd.fake.
		Invokes(testing.NewListAction(flowDeploymentResource, fd.ns, opts), &client.FlowDeploymentList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &client.FlowDeploymentList{}
	for _, item := range obj.(*client.FlowDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
