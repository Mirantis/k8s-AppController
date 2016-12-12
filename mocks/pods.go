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
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type podClient struct {
}

func MakePod(name string) *api.Pod {
	status := strings.Split(name, "-")[0]

	pod := &api.Pod{}
	pod.Name = name

	if status == "ready" {
		pod.Status.Phase = "Running"
		pod.Status.Conditions = append(
			pod.Status.Conditions,
			api.PodCondition{Type: "Ready", Status: "True"},
		)
	} else {
		pod.Status.Phase = "Pending"
	}

	return pod
}

func (p *podClient) List(opts api.ListOptions) (*api.PodList, error) {
	var pods []api.Pod
	for i := 0; i < 3; i++ {
		pods = append(pods, *MakePod(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any pending pods
	if strings.Index(opts.LabelSelector.String(), "failedpod=yes") >= 0 {
		for i := 0; i < 2; i++ {
			pods = append(pods, *MakePod(fmt.Sprintf("pending-lolo%d", i)))
		}
	}

	return &api.PodList{Items: pods}, nil
}

func (p *podClient) Get(name string) (*api.Pod, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock pod %s returned error", name)
	}

	return MakePod(name), nil
}

func (p *podClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (p *podClient) Create(pod *api.Pod) (*api.Pod, error) {
	return MakePod(pod.Name), nil
}

func (p *podClient) Update(pod *api.Pod) (*api.Pod, error) {
	panic("not implemented")
}

func (p *podClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *podClient) Bind(binding *api.Binding) error {
	panic("not implemented")
}

func (p *podClient) UpdateStatus(pod *api.Pod) (*api.Pod, error) {
	panic("not implemented")
}

func (p *podClient) GetLogs(name string, opts *api.PodLogOptions) *restclient.Request {
	panic("not implemented")
}

func NewPodClient() unversioned.PodInterface {
	return &podClient{}
}
