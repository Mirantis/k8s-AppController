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

	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	policy "k8s.io/client-go/1.5/pkg/apis/policy/v1alpha1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type podClient struct {
}

func MakePod(name string) *v1.Pod {
	status := strings.Split(name, "-")[0]

	pod := &v1.Pod{}
	pod.Name = name

	if status == "ready" {
		pod.Status.Phase = "Running"
		pod.Status.Conditions = append(
			pod.Status.Conditions,
			v1.PodCondition{Type: "Ready", Status: "True"},
		)
	} else {
		pod.Status.Phase = "Pending"
	}

	return pod
}

func (p *podClient) List(opts api.ListOptions) (*v1.PodList, error) {
	var pods []v1.Pod
	for i := 0; i < 3; i++ {
		pods = append(pods, *MakePod(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any pending pods
	if strings.Index(opts.LabelSelector.String(), "failedpod=yes") >= 0 {
		for i := 0; i < 2; i++ {
			pods = append(pods, *MakePod(fmt.Sprintf("pending-lolo%d", i)))
		}
	}

	return &v1.PodList{Items: pods}, nil
}

func (p *podClient) Get(name string) (*v1.Pod, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock pod %s returned error", name)
	}

	return MakePod(name), nil
}

func (p *podClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (p *podClient) Create(pod *v1.Pod) (*v1.Pod, error) {
	return MakePod(pod.Name), nil
}

func (p *podClient) Update(pod *v1.Pod) (*v1.Pod, error) {
	panic("not implemented")
}

func (p *podClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *podClient) Bind(binding *v1.Binding) error {
	panic("not implemented")
}

func (p *podClient) UpdateStatus(pod *v1.Pod) (*v1.Pod, error) {
	panic("not implemented")
}

func (p *podClient) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	panic("not implemented")
}

func (p *podClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (p *podClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Pod, err error) {
	panic("not implemented")
}

func (p *podClient) Evict(eviction *policy.Eviction) error {
	panic("not implemented")
}

func NewPodClient() corev1.PodInterface {
	return &podClient{}
}
