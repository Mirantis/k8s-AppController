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

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
)

type daemonSetClient struct {
}

// MakeDaemonSet creates a daemonset base in its name
func MakeDaemonSet(name string) *extbeta1.DaemonSet {
	daemonSet := &extbeta1.DaemonSet{}
	daemonSet.Name = name
	daemonSet.Status.DesiredNumberScheduled = 3
	if name == "fail" {
		daemonSet.Status.CurrentNumberScheduled = 2
	} else {
		daemonSet.Status.CurrentNumberScheduled = 3
	}

	return daemonSet
}

func (r *daemonSetClient) List(opts v1.ListOptions) (*extbeta1.DaemonSetList, error) {
	var daemonSets []extbeta1.DaemonSet
	for i := 0; i < 3; i++ {
		daemonSets = append(daemonSets, *MakeDaemonSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any failed daemonSets
	if strings.Index(opts.LabelSelector, "faileddaemonSet=yes") >= 0 {
		daemonSets = append(daemonSets, *MakeDaemonSet("fail"))
	}

	return &extbeta1.DaemonSetList{Items: daemonSets}, nil
}

func (r *daemonSetClient) Get(name string) (*extbeta1.DaemonSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock daemonset %s returned error", name)
	}

	return MakeDaemonSet(name), nil
}

func (r *daemonSetClient) Create(rs *extbeta1.DaemonSet) (*extbeta1.DaemonSet, error) {
	return MakeDaemonSet(rs.Name), nil
}

func (r *daemonSetClient) Update(rs *extbeta1.DaemonSet) (*extbeta1.DaemonSet, error) {
	panic("not implemented")
}

func (r *daemonSetClient) UpdateStatus(rs *extbeta1.DaemonSet) (*extbeta1.DaemonSet, error) {
	panic("not implemented")
}

func (r *daemonSetClient) Delete(name string, opts *v1.DeleteOptions) error {
	panic("not implemented")
}

func (r *daemonSetClient) Watch(opts v1.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *daemonSetClient) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	panic("not implemented")
}

func (r *daemonSetClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *extbeta1.DaemonSet, err error) {
	panic("not implemented")
}

func (r *daemonSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

// NewDaemonSetClient is a daemonset client constructor
func NewDaemonSetClient() v1beta1.DaemonSetInterface {
	return &daemonSetClient{}
}
