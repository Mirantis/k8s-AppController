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

type daemonSetClient struct {
}

// MakeDaemonSet creates a daemonset base in its name
func MakeDaemonSet(name string) *extensions.DaemonSet {
	daemonSet := &extensions.DaemonSet{}
	daemonSet.Name = name
	daemonSet.Status.DesiredNumberScheduled = 3
	if name == "fail" {
		daemonSet.Status.CurrentNumberScheduled = 2
	} else {
		daemonSet.Status.CurrentNumberScheduled = 3
	}

	return daemonSet
}

func (r *daemonSetClient) List(opts api.ListOptions) (*extensions.DaemonSetList, error) {
	var daemonSets []extensions.DaemonSet
	for i := 0; i < 3; i++ {
		daemonSets = append(daemonSets, *MakeDaemonSet(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any failed daemonSets
	if strings.Index(opts.LabelSelector.String(), "faileddaemonSet=yes") >= 0 {
		daemonSets = append(daemonSets, *MakeDaemonSet("fail"))
	}

	return &extensions.DaemonSetList{Items: daemonSets}, nil
}

func (r *daemonSetClient) Get(name string) (*extensions.DaemonSet, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock daemonset %s returned error", name)
	}

	return MakeDaemonSet(name), nil
}

func (r *daemonSetClient) Create(rs *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	return MakeDaemonSet(rs.Name), nil
}

func (r *daemonSetClient) Update(rs *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	panic("not implemented")
}

func (r *daemonSetClient) UpdateStatus(rs *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	panic("not implemented")
}

func (r *daemonSetClient) Delete(name string) error {
	panic("not implemented")
}

func (r *daemonSetClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *daemonSetClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

// NewDaemonSetClient is a daemonset client constructor
func NewDaemonSetClient() unversioned.DaemonSetInterface {
	return &daemonSetClient{}
}
