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
	"reflect"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

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
	Base
	Pod    *v1.Pod
	Client corev1.PodInterface
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
	def.Pod = parametrizeResource(def.Pod, gc, podParamFields).(*v1.Pod)
	return createNewPod(def, c.Pods())
}

// NewExisting returns Pod controller for existing resource by its name
func (podTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingPod{Name: name, Client: c.Pods()}}
}

func podKey(name string) string {
	return "pod/" + name
}

// Key returns Pod name
func (p newPod) Key() string {
	return podKey(p.Pod.Name)
}

func podStatus(pod *v1.Pod) (interfaces.ResourceStatus, error) {
	if pod.Status.Phase == "Succeeded" {
		return interfaces.ResourceReady, nil
	}

	if pod.Status.Phase == "Running" && isReady(pod) {
		return interfaces.ResourceReady, nil
	}

	return interfaces.ResourceNotReady, nil
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
	if err := checkExistence(p); err != nil {
		log.Println("Creating", p.Key())
		p.Pod, err = p.Client.Create(p.Pod)
		return err
	}
	return nil
}

// UpdateFromDefinition updates k8s object with the definition contents
func (p newPod) UpdateFromDefinition() (err error) {
	p.Pod, err = p.Client.Update(p.Definition.Pod)
	return err
}

// Delete deletes pod from the cluster
func (p newPod) Delete() error {
	return p.Client.Delete(p.Pod.Name, nil)
}

// Status returns pod status. It returns interfaces.ResourceReady if the pod is succeeded or running with succeeding readiness probe.
func (p newPod) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	pod, err := p.Client.Get(p.Pod.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !p.equalsToDefinition(pod) {
		return interfaces.ResourceWaitingForUpgrade, nil
	}
	return podStatus(pod)
}

// equalsToDefinition checks if definition in object is compatible with provided object
func (p newPod) equalsToDefinition(podiface interface{}) bool {
	pod := podiface.(*v1.Pod)

	return reflect.DeepEqual(pod.ObjectMeta, p.Pod.ObjectMeta) && reflect.DeepEqual(pod.Spec, p.Pod.Spec)
}

func createNewPod(def client.ResourceDefinition, c corev1.PodInterface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: newPod{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			Pod:    def.Pod,
			Client: c,
		},
	}
}

type existingPod struct {
	Base
	Name   string
	Client corev1.PodInterface
}

// Key returns Pod name
func (p existingPod) Key() string {
	return podKey(p.Name)
}

// Create looks for existing Pod and returns error if there is no such Pod
func (p existingPod) Create() error {
	return createExistingResource(p)
}

// Status returns pod status. It returns interfaces.ResourceReady if the pod is succeeded or running with succeeding readiness probe.
func (p existingPod) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	pod, err := p.Client.Get(p.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}
	return podStatus(pod)
}

// Delete deletes pod from the cluster
func (p existingPod) Delete() error {
	return p.Client.Delete(p.Name, nil)
}
