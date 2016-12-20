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
	"strings"

	"k8s.io/client-go/1.5/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/api"
	extbeta1 "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type replicaSetClient struct {
}

func MakeReplicaSet(name string) *extbeta1.ReplicaSet {
	replicaSet := &extbeta1.ReplicaSet{}
	replicaSet.Name = name
	replicaSet.Spec.Replicas = pointer(int32(2))
	if name != "fail" {
		replicaSet.Status.Replicas = int32(3)
	}

	return replicaSet
}

func (r *replicaSetClient) List(opts api.ListOptions) (*extbeta1.ReplicaSetList, error) {
	var replicaSets []extbeta1.ReplicaSet
	for i := 0; i < 3; i++ {
		replicaSets = append(replicaSets, *MakeReplicaSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any failed replicaSets
	if strings.Index(opts.LabelSelector.String(), "failedreplicaSet=yes") >= 0 {
		replicaSets = append(replicaSets, *MakeReplicaSet("fail"))
	}

	return &extbeta1.ReplicaSetList{Items: replicaSets}, nil
}

func (r *replicaSetClient) Get(name string) (*extbeta1.ReplicaSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeReplicaSet(name), nil
}

func (r *replicaSetClient) Create(rs *extbeta1.ReplicaSet) (*extbeta1.ReplicaSet, error) {
	return MakeReplicaSet(rs.Name), nil
}

func (r *replicaSetClient) Update(rs *extbeta1.ReplicaSet) (*extbeta1.ReplicaSet, error) {
	panic("not implemented")
}

func (r *replicaSetClient) UpdateStatus(rs *extbeta1.ReplicaSet) (*extbeta1.ReplicaSet, error) {
	panic("not implemented")
}

func (r *replicaSetClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *replicaSetClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *replicaSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (r *replicaSetClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (r *replicaSetClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *extbeta1.ReplicaSet, err error) {
	panic("not implemented")
}

func NewReplicaSetClient() v1beta1.ReplicaSetInterface {
	return &replicaSetClient{}
}
