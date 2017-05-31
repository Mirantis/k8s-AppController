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

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var podParamFields = []string{
	"Spec.Containers.Env",
	"Spec.Containers.Name",
	"Spec.InitContainers.Env",
	"Spec.InitContainers.Name",
}

type newPod struct {
	pod    *v1.Pod
	client corev1.PodInterface
}

type existingPod struct {
	name   string
	client corev1.PodInterface
}

type podTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a pod
func (podTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Pod == nil {
		return ""
	}
	return definition.Pod.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (podTemplateFactory) Kind() string {
	return "pod"
}

// New returns Pod controller for new resource based on resource definition
func (podTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	pod := parametrizeResource(def.Pod, gc, podParamFields).(*v1.Pod)
	return createNewPod(pod, c.Pods())
}

// NewExisting returns Pod controller for existing resource by its name
func (podTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingPod{name: name, client: c.Pods()}
}

func podKey(name string) string {
	return "pod/" + name
}

// Key returns Pod name
func (p newPod) Key() string {
	return podKey(p.pod.Name)
}

func podProgress(p corev1.PodInterface, name string) (float32, error) {
	pod, err := p.Get(name)
	if err != nil {
		return 0, err
	}

	if pod.Status.Phase == "Succeeded" || pod.Status.Phase == "Running" && isReady(pod) {
		return 1, nil
	}

	return 0, nil
}

func isReady(pod *v1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

// Create looks for the Pod in k8s and creates it if not present
func (p newPod) Create() error {
	if checkExistence(p) {
		return nil
	}
	log.Println("Creating", p.Key())
	obj, err := p.client.Create(p.pod)
	p.pod = obj
	return err
}

// Delete deletes pod from the cluster
func (p newPod) Delete() error {
	return p.client.Delete(p.pod.Name, nil)
}

// GetProgress returns Pod deployment progress
func (p newPod) GetProgress() (float32, error) {
	return podProgress(p.client, p.pod.Name)
}

func createNewPod(pod *v1.Pod, client corev1.PodInterface) interfaces.Resource {
	return newPod{pod: pod, client: client}
}

// Key returns Pod name
func (p existingPod) Key() string {
	return podKey(p.name)
}

// Create looks for existing Pod and returns error if there is no such Pod
func (p existingPod) Create() error {
	return createExistingResource(p)
}

// GetProgress returns Pod deployment progress
func (p existingPod) GetProgress() (float32, error) {
	return podProgress(p.client, p.name)
}

// Delete deletes pod from the cluster
func (p existingPod) Delete() error {
	return p.client.Delete(p.name, nil)
}
