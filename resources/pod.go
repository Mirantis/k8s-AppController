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

package resources

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
)

type Pod struct {
	Pod    *api.Pod
	Client unversioned.PodInterface
}

func podKey(name string) string {
	return "pod/" + name
}

func (p Pod) Key() string {
	return podKey(p.Pod.Name)
}

func podStatus(p unversioned.PodInterface, name string) (string, error) {
	pod, err := p.Get(name)
	if err != nil {
		return "error", err
	}

	if pod.Status.Phase == "Succeeded" {
		return "ready", nil
	}

	if pod.Status.Phase == "Running" && isReady(pod) {
		return "ready", nil
	}

	return "not ready", nil
}

func isReady(pod *api.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

func (p Pod) Create() error {
	if err := checkExistence(p); err != nil {
		log.Println("Creating ", p.Key())
		p.Pod, err = p.Client.Create(p.Pod)
		return err
	}
	return nil
}

// Delete deletes pod from the cluster
func (p Pod) Delete() error {
	return p.Client.Delete(p.Pod.Name, nil)
}

func (p Pod) Status(meta map[string]string) (string, error) {
	return podStatus(p.Client, p.Pod.Name)
}

// NameMatches gets resource definition and a name and checks if
// the Pod part of resource definition has matching name.
func (p Pod) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Pod != nil && def.Pod.Name == name
}

// New returns new Pod based on resource definition
func (p Pod) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewPod(def.Pod, c.Pods())
}

// NewExisting returns new ExistingPod based on resource definition
func (p Pod) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingPod(name, c.Pods())
}

func NewPod(pod *api.Pod, client unversioned.PodInterface) Pod {
	return Pod{Pod: pod, Client: client}
}

type ExistingPod struct {
	Name   string
	Client unversioned.PodInterface
}

func (p ExistingPod) Key() string {
	return podKey(p.Name)
}

func (p ExistingPod) Create() error {
	return createExistingResource(p)
}

func (p ExistingPod) Status(meta map[string]string) (string, error) {
	return podStatus(p.Client, p.Name)
}

// Delete deletes pod from the cluster
func (p ExistingPod) Delete() error {
	return p.Client.Delete(p.Name, nil)
}

func NewExistingPod(name string, client unversioned.PodInterface) ExistingPod {
	return ExistingPod{Name: name, Client: client}
}
