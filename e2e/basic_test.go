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

package integration

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"

	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/client"
)

var _ = Describe("Basic Suite", func() {
	var clientset *kubernetes.Clientset
	var c client.Interface
	var framework GraphFramework
	var namespace *v1.Namespace

	SetDefaultEventuallyTimeout(120 * time.Second)
	SetDefaultEventuallyPollingInterval(5 * time.Second)
	SetDefaultConsistentlyDuration(5 * time.Second)
	SetDefaultConsistentlyPollingInterval(5 * time.Second)

	BeforeEach(func() {
		By("Creating namespace and initializing test framework")
		var err error
		clientset, err = testutils.KubeClient()
		namespaceObj := &v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "e2e-tests-ac-",
				Namespace:    "",
			},
			Status: v1.NamespaceStatus{},
		}
		namespace, err = clientset.Namespaces().Create(namespaceObj)
		Expect(err).NotTo(HaveOccurred())
		c, err = testutils.GetAcClient(namespace.Name)
		Expect(err).NotTo(HaveOccurred())
		framework = GraphFramework{
			client:    c,
			clientset: clientset,
			ns:        namespace.Name,
		}
		By("Deploying appcontroller image")
		framework.Prepare()
	})

	AfterEach(func() {
		By("Removing namespace")
		testutils.DeleteNS(clientset, namespace)
		By("Removing all resource definitions")
		resDefs, err := c.ResourceDefinitions().List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, resDef := range resDefs.Items {
			err := c.ResourceDefinitions().Delete(resDef.Name, nil)
			Expect(err).NotTo(HaveOccurred())
		}
		By("Removing all dependencies")
		deps, err := c.Dependencies().List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, dep := range deps.Items {
			err := c.Dependencies().Delete(dep.Name, nil)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Dependent Pod should not be created if parent is stuck in init", func() {
		parentPod := &v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name: "sleeper-parent",
				Annotations: map[string]string{
					"pod.alpha.kubernetes.io/init-containers": `[{"name": "sleeper-init", "image": "kubernetes/pause"}]`,
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "sleeper",
						Image: "kubernetes/pause",
					},
				},
			},
		}
		childPod := PodPause("child-pod")
		framework.Connect(framework.WrapAndCreate(parentPod), framework.WrapAndCreate(childPod))
		framework.Run()
		testutils.WaitForPod(clientset, namespace.Name, parentPod.Name, "")
		time.Sleep(time.Second)
		_, err := clientset.Pods(namespace.Name).Get(childPod.Name)
		Expect(err).To(HaveOccurred())
	})

	It("Dependent Pod should be created if parent initialises correctly", func() {
		parentPod := PodPause("parent-pod")
		childPod := PodPause("child-pod")
		framework.Connect(framework.WrapAndCreate(parentPod), framework.WrapAndCreate(childPod))
		framework.Run()
		testutils.WaitForPod(clientset, namespace.Name, parentPod.Name, v1.PodRunning)
		testutils.WaitForPod(clientset, namespace.Name, childPod.Name, v1.PodRunning)
	})

	It("Service resdef works correctly with any k8s version", func() {
		pod1 := PodPause("pod1")
		pod1.Labels = map[string]string{"before": "service"}
		pod2 := PodPause("pod2")
		pod2.Labels = map[string]string{"after": "service"}
		ports := []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 9999, TargetPort: intstr.FromInt(9999)}}
		svc := &v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name: "svc1",
			},
			Spec: v1.ServiceSpec{
				Selector: pod1.Labels,
				Type:     v1.ServiceTypeNodePort,
				Ports:    ports,
			},
		}
		By("Creating two pods and one service")
		svcWrapped := framework.WrapAndCreate(svc)
		By("Service depends on first pod")
		framework.Connect(framework.WrapAndCreate(pod1), svcWrapped)
		By("Second pod depends on service")
		framework.Connect(svcWrapped, framework.WrapAndCreate(pod2))
		framework.Run()
		By("Verifying that second pod will enter running state")
		testutils.WaitForPod(clientset, namespace.Name, pod2.Name, v1.PodRunning)
	})

	Describe("Failure handling - subgraph", func() {
		var parentPod *v1.Pod
		var childPod *v1.Pod

		BeforeEach(func() {
			parentPod = PodPause("parent-pod")
			childPod = PodPause("child-pod")
		})

		Context("If parent succeeds", func() {
			It("on-error dependency must not be followed", func() {
				framework.ConnectWithMeta(framework.WrapAndCreate(parentPod), framework.WrapAndCreate(childPod), map[string]string{"on-error": "true"})
				framework.Run()
				testutils.WaitForPod(clientset, namespace.Name, parentPod.Name, v1.PodRunning)
				testutils.WaitForPodNotToBeCreated(clientset, namespace.Name, childPod.Name)
			})
		})

		Context("If parent fails", func() {
			It("On-error dependency must be followed if parent fails", func() {
				framework.ConnectWithMeta(framework.Wrap(parentPod), framework.WrapAndCreate(childPod), map[string]string{"on-error": "true"})
				framework.Run()
				testutils.WaitForPodNotToBeCreated(clientset, namespace.Name, parentPod.Name)
				testutils.WaitForPod(clientset, namespace.Name, childPod.Name, v1.PodRunning)
			})
		})
	})

	Describe("Failure handling - timeout", func() {
		var parentPod *v1.Pod
		var childPod *v1.Pod

		BeforeEach(func() {
			parentPod = DelayedPod("parent-pod", 15)
			childPod = PodPause("child-pod")
		})

		Context("If parent timed out", func() {
			It("on-error dependency must be followed", func() {
				parentResDef := framework.WrapWithMetaAndCreate(parentPod, map[string]interface{}{"timeout": 5})
				framework.ConnectWithMeta(parentResDef, framework.WrapAndCreate(childPod), map[string]string{"on-error": "true"})
				framework.Run()
				testutils.WaitForPod(clientset, namespace.Name, childPod.Name, v1.PodRunning)
			})
		})
	})
})

func getKind(resdef *client.ResourceDefinition) string {
	if resdef.Pod != nil {
		return "pod"
	} else if resdef.Service != nil {
		return "service"
	}
	return ""
}

type GraphFramework struct {
	client    client.Interface
	clientset *kubernetes.Clientset
	ns        string
}

func (g GraphFramework) Wrap(obj runtime.Object) *client.ResourceDefinition {
	resdef := &client.ResourceDefinition{}
	switch v := obj.(type) {
	case *v1.Pod:
		resdef.Pod = v
		resdef.Name = v.Name
	case *v1.Service:
		resdef.Service = v
		resdef.Name = v.Name
	default:
		panic("Unknown type provided")
	}
	return resdef
}

func (g GraphFramework) WrapWithMeta(obj runtime.Object, meta map[string]interface{}) *client.ResourceDefinition {
	resdef := g.Wrap(obj)
	resdef.Meta = meta
	return resdef
}

func (g GraphFramework) WrapAndCreate(obj runtime.Object) *client.ResourceDefinition {
	resdef := g.Wrap(obj)
	_, err := g.client.ResourceDefinitions().Create(resdef)
	Expect(err).NotTo(HaveOccurred())
	return resdef
}

func (g GraphFramework) WrapWithMetaAndCreate(obj runtime.Object, meta map[string]interface{}) *client.ResourceDefinition {
	resdef := g.WrapWithMeta(obj, meta)
	_, err := g.client.ResourceDefinitions().Create(resdef)
	Expect(err).NotTo(HaveOccurred())
	return resdef
}

func (g GraphFramework) ConnectWithMeta(first, second *client.ResourceDefinition, meta map[string]string) {
	dep := &client.Dependency{
		ObjectMeta: api.ObjectMeta{
			GenerateName: "dep-",
			Labels: map[string]string{
				"ns": g.ns,
			},
		},
		Parent: getKind(first) + "/" + first.Name,
		Child:  getKind(second) + "/" + second.Name,
		Meta:   meta,
	}
	_, err := g.client.Dependencies().Create(dep)
	Expect(err).NotTo(HaveOccurred())
}

func (g GraphFramework) Connect(first, second *client.ResourceDefinition) {
	g.ConnectWithMeta(first, second, map[string]string{})
}

func (g GraphFramework) Run() {
	cmd := exec.Command(
		"kubectl",
		"--namespace",
		g.ns,
		"exec",
		"k8s-appcontroller",
		"--",
		"ac-run",
		"-l",
		"ns:"+g.ns,
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

func (g GraphFramework) Prepare() {
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
							Value: g.ns,
						},
					},
				},
			},
		},
	}
	ac, err := g.clientset.Pods(g.ns).Create(appControllerObj)
	Expect(err).NotTo(HaveOccurred())
	testutils.WaitForPod(g.clientset, g.ns, ac.Name, v1.PodRunning)
	Eventually(func() bool {
		_, depsErr := g.client.ResourceDefinitions().List(api.ListOptions{})
		_, defsErr := g.client.Dependencies().List(api.ListOptions{})
		return defsErr == nil && depsErr == nil
	}).Should(BeTrue())
}

func PodPause(name string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "sleeper",
					Image: "kubernetes/pause",
				},
			},
		},
	}
}

func DelayedPod(name string, delay int) *v1.Pod {
	cmdArg := fmt.Sprintf("sleep %d; echo ok > /tmp/health; sleep 60", delay)
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "sleeper",
					Image:   "gcr.io/google_containers/busybox",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", cmdArg},
					ReadinessProbe: &v1.Probe{
						Handler: v1.Handler{
							Exec: &v1.ExecAction{
								Command: []string{"/bin/cat", "/tmp/health"},
							},
						},
					},
				},
			},
		},
	}
}
