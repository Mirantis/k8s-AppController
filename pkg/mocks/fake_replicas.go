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

// fakeReplicas implements ReplicaInterface
type fakeReplicas struct {
	fake *testing.Fake
	ns   string
}

var replicaResource = unversioned.GroupVersionResource{
	Group:    "appcontroller.k8s",
	Version:  "v1alpha1",
	Resource: "replica",
}

// Create stores new replica object in the fake k8s
func (fr *fakeReplicas) Create(replica *client.Replica) (result *client.Replica, err error) {
	if replica.GenerateName != "" {
		uuid, _ := newUUID()
		replica.Name = replica.GenerateName + uuid
	}
	replica.SetCreationTimestamp(unversioned.Time{time.Now()})
	uuid, _ := newUUID()
	replica.SetUID(types.UID(uuid))

	obj, err := fr.fake.
		Invokes(testing.NewCreateAction(replicaResource, fr.ns, replica), &client.Replica{})

	if obj == nil {
		return nil, err
	}
	res := obj.(*client.Replica)
	return res, err
}

// Update updates Replica object stored in the fake k8s
func (fr *fakeReplicas) Update(replica *client.Replica) (result *client.Replica, err error) {
	obj, err := fr.fake.
		Invokes(testing.NewUpdateAction(replicaResource, fr.ns, replica), &client.Replica{})

	if obj == nil {
		return nil, err
	}
	return obj.(*client.Replica), err
}

// Delete deletes a replica object by its name
func (fr *fakeReplicas) Delete(name string) error {
	_, err := fr.fake.
		Invokes(testing.NewDeleteAction(replicaResource, fr.ns, name), &client.Replica{})

	return err
}

// List returns a list of flow replicas stored in fake k8s
func (fr *fakeReplicas) List(opts api.ListOptions) (result *client.ReplicaList, err error) {
	obj, err := fr.fake.
		Invokes(testing.NewListAction(replicaResource, fr.ns, opts), &client.ReplicaList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &client.ReplicaList{}
	for _, item := range obj.(*client.ReplicaList).Items {
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
