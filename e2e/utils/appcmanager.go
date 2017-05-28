// Copyright 2017 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
)

const (
	appcontrollerPod = "k8s-appcontroller"
)

// AppControllerManager is the test controller exposes remote AC operations and test primitives for dependency graphs
type AppControllerManager struct {
	Client    client.Interface
	Clientset *kubernetes.Clientset
	ns        string
	Namespace *v1.Namespace

	acPod *v1.Pod
}

// Run runs dependency graph deployment with default settings
func (a *AppControllerManager) Run() {
	a.RunWithOptions(false, interfaces.DependencyGraphOptions{MinReplicaCount: 1})
}

// Start runs dependency graph deployment with default settings without waiting for deployment to complete
func (a *AppControllerManager) Start() {
	a.RunWithOptions(true, interfaces.DependencyGraphOptions{MinReplicaCount: 1})
}

// RunWithOptions runs dependency graph deployment with given settings
func (a *AppControllerManager) RunWithOptions(runAsync bool, options interfaces.DependencyGraphOptions) {
	sched := scheduler.New(a.Client, nil, 0)

	task, err := scheduler.Deploy(sched, options, false, nil)
	Expect(err).NotTo(HaveOccurred())
	if !runAsync {
		Eventually(
			func() error {
				_, err := a.Client.ConfigMaps().Get(task)
				return err
			},
			300*time.Second, 5*time.Second).Should(HaveOccurred(), "Deployment job wasn't completed")
	}
}

// DeleteAppControllerPod deletes pod, where AppController is running
func (a *AppControllerManager) DeleteAppControllerPod() {
	By("Removing pod  " + appcontrollerPod)
	err := a.Client.Pods().Delete(appcontrollerPod, nil)
	Expect(err).NotTo(HaveOccurred())
	Eventually(func() bool {
		_, err := a.Client.Pods().Get(appcontrollerPod)
		return errors.IsNotFound(err)
	}, 20*time.Second, 1*time.Second).Should(BeTrue(), "Appcontroller pod wasn't removed in time")
}

// Prepare starts AppController pod
func (a *AppControllerManager) Prepare() {
	appControllerObj := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: appcontrollerPod,
			Annotations: map[string]string{
				"pod.alpha.kubernetes.io/init-containers": `[{"name": "kubeac-bootstrap", "image": "mirantis/k8s-appcontroller", "imagePullPolicy": "Never", "command": ["kubeac", "bootstrap", "/opt/kubeac/manifests"]}]`,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Always",
			Containers: []v1.Container{
				{
					Name:            "kubeac",
					Image:           "mirantis/k8s-appcontroller",
					Command:         []string{"kubeac", "deploy"},
					ImagePullPolicy: v1.PullNever,
					Env: []v1.EnvVar{
						{
							Name:  "KUBERNETES_AC_LABEL_SELECTOR",
							Value: "",
						},
						{
							Name:  "KUBERNETES_AC_POD_NAMESPACE",
							Value: a.Namespace.Name,
						},
					},
				},
			},
		},
	}
	var err error
	a.acPod, err = a.Clientset.Pods(a.Namespace.Name).Create(appControllerObj)
	Expect(err).NotTo(HaveOccurred())
	WaitForPod(a.Clientset, a.Namespace.Name, a.acPod.Name, v1.PodRunning)
	Eventually(func() bool {
		_, depsErr := a.Client.ResourceDefinitions().List(api.ListOptions{})
		_, defsErr := a.Client.Dependencies().List(api.ListOptions{})
		return defsErr == nil && depsErr == nil
	}, 120*time.Second, 5*time.Second).Should(BeTrue())
}

// BeforeEach is an action executed before each test
func (a *AppControllerManager) BeforeEach() {
	var err error
	a.Clientset, err = KubeClient()
	Expect(err).NotTo(HaveOccurred())
	AddServiceAccountToAdmins(a.Clientset)
	By("Creating namespace and initializing test framework")
	namespaceObj := &v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "e2e-tests-ac-",
			Namespace:    "",
		},
		Status: v1.NamespaceStatus{},
	}
	a.Namespace, err = a.Clientset.Namespaces().Create(namespaceObj)
	Expect(err).NotTo(HaveOccurred())
	By("Creating AC client with namespace " + a.Namespace.Name)
	a.Client, err = GetAcClient(a.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())
	By("Deploying appcontroller image")
	a.Prepare()
}

// AfterEach is an action executed after each test
func (a *AppControllerManager) AfterEach() {
	By("Dumping appcontroller logs")
	if CurrentGinkgoTestDescription().Failed && a.acPod != nil {
		DumpLogs(a.Clientset, a.acPod)
	}
	By("Removing namespace")
	DeleteNS(a.Clientset, a.Namespace)
	RemoveServiceAccountFromAdmins(a.Clientset)
}

// NewAppControllerManager creates new instance of AppControllerManager
func NewAppControllerManager() *AppControllerManager {
	appc := &AppControllerManager{}
	BeforeEach(appc.BeforeEach)
	AfterEach(appc.AfterEach)
	return appc
}
