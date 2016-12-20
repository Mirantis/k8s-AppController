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

	"k8s.io/client-go/1.5/kubernetes/typed/apps/v1alpha1"
	"k8s.io/client-go/1.5/pkg/api"
	appsalpha1 "k8s.io/client-go/1.5/pkg/apis/apps/v1alpha1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type petSetClient struct {
}

func pointer(i int32) *int32 {
	return &i
}

// MakePetSet returns a new K8s PetSet object for the client to return. If it's name is "fail" it will have labels that will cause it's underlying mock Pods to fail.
func MakePetSet(name string) *appsalpha1.PetSet {
	petSet := &appsalpha1.PetSet{}
	petSet.Name = name
	petSet.Spec.Replicas = pointer(int32(3))
	petSet.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	if name == "fail" {
		petSet.Spec.Template.ObjectMeta.Labels["failedpod"] = "yes"
		petSet.Status.Replicas = int32(2)
	} else {
		petSet.Status.Replicas = int32(3)
	}

	return petSet
}

func (r *petSetClient) List(opts api.ListOptions) (*appsalpha1.PetSetList, error) {
	var petSets []appsalpha1.PetSet
	for i := 0; i < 3; i++ {
		petSets = append(petSets, *MakePetSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any failed petSets
	if strings.Index(opts.LabelSelector.String(), "failedpetSet=yes") >= 0 {
		petSets = append(petSets, *MakePetSet("fail"))
	}

	return &appsalpha1.PetSetList{Items: petSets}, nil
}

func (r *petSetClient) Get(name string) (*appsalpha1.PetSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakePetSet(name), nil
}

func (r *petSetClient) Create(rs *appsalpha1.PetSet) (*appsalpha1.PetSet, error) {
	return MakePetSet(rs.Name), nil
}

func (r *petSetClient) Update(rs *appsalpha1.PetSet) (*appsalpha1.PetSet, error) {
	panic("not implemented")
}

func (r *petSetClient) UpdateStatus(rs *appsalpha1.PetSet) (*appsalpha1.PetSet, error) {
	panic("not implemented")
}

func (r *petSetClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *petSetClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *petSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (r *petSetClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (r *petSetClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *appsalpha1.PetSet, err error) {
	panic("not implemented")
}

// NewPetSetClient is a client constructor
func NewPetSetClient() v1alpha1.PetSetInterface {
	return &petSetClient{}
}
