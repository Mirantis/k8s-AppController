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
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

type AppControllerManager struct {
	Client    client.Interface
	Clientset *kubernetes.Clientset
	ns        string
	Namespace *v1.Namespace
}

func (a *AppControllerManager) Run() {
	cmd := exec.Command(
		"kubectl",
		"--namespace",
		a.Namespace.Name,
		"exec",
		"k8s-appcontroller",
		"--",
		"ac-run",
		"-l",
		"ns:"+a.Namespace.Name,
	)
	out, err := cmd.Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			exErr := err.(*exec.ExitError)
			Fail(string(out) + string(exErr.Stderr))
		default:
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func (a *AppControllerManager) Prepare() {
	appControllerObj := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "k8s-appcontroller",
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
					Command:         []string{"/usr/bin/run_runit"},
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
	ac, err := a.Clientset.Pods(a.Namespace.Name).Create(appControllerObj)
	Expect(err).NotTo(HaveOccurred())
	WaitForPod(a.Clientset, a.Namespace.Name, ac.Name, v1.PodRunning)
	Eventually(func() bool {
		_, depsErr := a.Client.ResourceDefinitions().List(api.ListOptions{})
		_, defsErr := a.Client.Dependencies().List(api.ListOptions{})
		return defsErr == nil && depsErr == nil
	}, 120*time.Second, 5*time.Second).Should(BeTrue())
}

func (a *AppControllerManager) BeforeEach() {
	var err error
	a.Clientset, err = KubeClient()
	Expect(err).NotTo(HaveOccurred())
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

func (a *AppControllerManager) AfterEach() {
	By("Removing namespace")
	DeleteNS(a.Clientset, a.Namespace)
	By("Removing all resource definitions")
	resDefs, err := a.Client.ResourceDefinitions().List(api.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	for _, resDef := range resDefs.Items {
		err := a.Client.ResourceDefinitions().Delete(resDef.Name, nil)
		Expect(err).NotTo(HaveOccurred())
	}
	By("Removing all dependencies")
	deps, err := a.Client.Dependencies().List(api.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	for _, dep := range deps.Items {
		err := a.Client.Dependencies().Delete(dep.Name, nil)
		Expect(err).NotTo(HaveOccurred())
	}
}

func NewAppControllerManager() *AppControllerManager {
	appc := &AppControllerManager{}
	BeforeEach(appc.BeforeEach)
	AfterEach(appc.AfterEach)
	return appc
}
