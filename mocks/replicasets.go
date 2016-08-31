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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type replicaSetClient struct {
}

func MakeReplicaSet(name string) *extensions.ReplicaSet {
	replicaSet := &extensions.ReplicaSet{}
	replicaSet.Name = name
	replicaSet.Spec.Replicas = 3
	if name == "fail" {
		replicaSet.Status.Replicas = 2
	} else {
		replicaSet.Status.Replicas = 3
	}

	return replicaSet
}

func (r *replicaSetClient) List(opts api.ListOptions) (*extensions.ReplicaSetList, error) {
	var replicaSets []extensions.ReplicaSet
	for i := 0; i < 3; i++ {
		replicaSets = append(replicaSets, *MakeReplicaSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	//use ListOptions.LabelSelector to check if there should be any failed replicaSets
	if strings.Index(opts.LabelSelector.String(), "failedreplicaSet=yes") >= 0 {
		replicaSets = append(replicaSets, *MakeReplicaSet("fail"))
	}

	return &extensions.ReplicaSetList{Items: replicaSets}, nil
}

func (r *replicaSetClient) Get(name string) (*extensions.ReplicaSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeReplicaSet(name), nil
}

func (r *replicaSetClient) Create(rs *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	return MakeReplicaSet(rs.Name), nil
}

func (r *replicaSetClient) Update(rs *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	panic("not implemented")
}

func (r *replicaSetClient) UpdateStatus(rs *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	panic("not implemented")
}

func (r *replicaSetClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *replicaSetClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *replicaSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

func NewReplicaSetClient() unversioned.ReplicaSetInterface {
	return &replicaSetClient{}
}
