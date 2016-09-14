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
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type petSetClient struct {
}

// MakePetSet returns a new K8s PetSet object for the client to return. If it's name is "fail" it will have labels that will cause it's underlying mock Pods to fail.
func MakePetSet(name string) *apps.PetSet {
	petSet := &apps.PetSet{}
	petSet.Name = name
	petSet.Spec.Replicas = 3
	petSet.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	if name == "fail" {
		petSet.Spec.Template.ObjectMeta.Labels["failedpod"] = "yes"
		petSet.Status.Replicas = 2
	} else {
		petSet.Status.Replicas = 3
	}

	return petSet
}

func (r *petSetClient) List(opts api.ListOptions) (*apps.PetSetList, error) {
	var petSets []apps.PetSet
	for i := 0; i < 3; i++ {
		petSets = append(petSets, *MakePetSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	//use ListOptions.LabelSelector to check if there should be any failed petSets
	if strings.Index(opts.LabelSelector.String(), "failedpetSet=yes") >= 0 {
		petSets = append(petSets, *MakePetSet("fail"))
	}

	return &apps.PetSetList{Items: petSets}, nil
}

func (r *petSetClient) Get(name string) (*apps.PetSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakePetSet(name), nil
}

func (r *petSetClient) Create(rs *apps.PetSet) (*apps.PetSet, error) {
	return MakePetSet(rs.Name), nil
}

func (r *petSetClient) Update(rs *apps.PetSet) (*apps.PetSet, error) {
	panic("not implemented")
}

func (r *petSetClient) UpdateStatus(rs *apps.PetSet) (*apps.PetSet, error) {
	panic("not implemented")
}

func (r *petSetClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *petSetClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *petSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

// NewPetSetClient is a client constructor
func NewPetSetClient() unversioned.PetSetInterface {
	return &petSetClient{}
}
