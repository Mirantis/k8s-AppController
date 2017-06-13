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
	"time"

	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"
)

var _ = Describe("Basic Suite", func() {
	SetDefaultEventuallyTimeout(120 * time.Second)
	SetDefaultEventuallyPollingInterval(5 * time.Second)
	SetDefaultConsistentlyDuration(5 * time.Second)
	SetDefaultConsistentlyPollingInterval(5 * time.Second)

	framework := GraphFramework{testutils.NewAppControllerManager()}

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
		framework.Connect(
			framework.WrapWithMetaAndCreate(parentPod, map[string]interface{}{"timeout": 30}),
			framework.WrapAndCreate(childPod))
		framework.Run()
		testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, parentPod.Name, "")
		time.Sleep(time.Second)
		_, err := framework.Clientset.Pods(framework.Namespace.Name).Get(childPod.Name)
		Expect(err).To(HaveOccurred())
	})

	It("Dependent Pod should be created if parent initialises correctly", func() {
		parentPod := PodPause("parent-pod")
		childPod := PodPause("child-pod")
		framework.Connect(framework.WrapAndCreate(parentPod), framework.WrapAndCreate(childPod))
		framework.Run()
		testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, parentPod.Name, v1.PodRunning)
		testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, childPod.Name, v1.PodRunning)
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
		testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, pod2.Name, v1.PodRunning)
	})

	It("Deployment should finish even if appcontroller pod was terminated", func() {
		By("Creating resource definition with single pod")
		pod1 := PodPause("pod1")
		framework.WrapAndCreate(pod1)
		framework.RunAsynchronously()
		framework.DeleteAppControllerPod()
		By("Verify that pod is consistently not found")
		Consistently(func() bool {
			_, err := framework.Client.Pods().Get(pod1.Name)
			return errors.IsNotFound(err)
		}, 5*time.Second, 1*time.Second).Should(BeTrue(), "Pod was unexpectadly created")
		By("Recreate appcontroller pod and verify that pod was successfully created")
		framework.Prepare()
		testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, pod1.Name, v1.PodRunning)
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
				testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, parentPod.Name, v1.PodRunning)
				testutils.WaitForPodNotToBeCreated(framework.Clientset, framework.Namespace.Name, childPod.Name)
			})
		})

		Context("If parent fails", func() {
			It("On-error dependency must be followed if parent fails", func() {
				framework.ConnectWithMeta(framework.Wrap(parentPod), framework.WrapAndCreate(childPod), map[string]string{"on-error": "true"})
				framework.Run()
				testutils.WaitForPodNotToBeCreated(framework.Clientset, framework.Namespace.Name, parentPod.Name)
				testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, childPod.Name, v1.PodRunning)
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
				testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, childPod.Name, v1.PodRunning)
			})
		})
	})

	Describe("Failure handling - ignore", func() {
		var parentPod *v1.Pod
		var childPod *v1.Pod

		BeforeEach(func() {
			parentPod = DelayedPod("parent-pod", 15)
			childPod = PodPause("child-pod")
		})

		Context("If failed parent is marked on-error:ignore", func() {
			It("dependency must be followed", func() {
				parentResDef := framework.WrapWithMetaAndCreate(parentPod, map[string]interface{}{"timeout": 5, "on-error": "ignore"})
				framework.Connect(parentResDef, framework.WrapAndCreate(childPod))
				framework.Run()
				testutils.WaitForPod(framework.Clientset, framework.Namespace.Name, childPod.Name, v1.PodRunning)
			})
		})
	})

	Describe("Failure handling - ignore-all", func() {
		var parentPod *v1.Pod
		var childPod *v1.Pod
		var grandChildPod *v1.Pod

		BeforeEach(func() {
			parentPod = DelayedPod("parent-pod", 15)
			childPod = PodPause("child-pod")
			grandChildPod = PodPause("grand-child-pod")
		})

		Context("If failed parent is marked on-error:ignore-all", func() {
			It("all children including transitive must not be created", func() {
				parentResDef := framework.WrapWithMetaAndCreate(parentPod, map[string]interface{}{"timeout": 5, "on-error": "ignore-all"})
				childResDef := framework.WrapAndCreate(childPod)
				grandChildResDef := framework.WrapAndCreate(grandChildPod)
				framework.Connect(parentResDef, childResDef)
				framework.Connect(childResDef, grandChildResDef)
				framework.Run()
				testutils.WaitForPodNotToBeCreated(framework.Clientset, framework.Namespace.Name, childPod.Name)
				testutils.WaitForPodNotToBeCreated(framework.Clientset, framework.Namespace.Name, grandChildPod.Name)
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
	*testutils.AppControllerManager
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
	_, err := g.Client.ResourceDefinitions().Create(resdef)
	Expect(err).NotTo(HaveOccurred())
	return resdef
}

func (g GraphFramework) WrapWithMetaAndCreate(obj runtime.Object, meta map[string]interface{}) *client.ResourceDefinition {
	resdef := g.WrapWithMeta(obj, meta)
	_, err := g.Client.ResourceDefinitions().Create(resdef)
	Expect(err).NotTo(HaveOccurred())
	return resdef
}

func (g GraphFramework) ConnectWithMeta(first, second *client.ResourceDefinition, meta map[string]string) {
	dep := &client.Dependency{
		ObjectMeta: api.ObjectMeta{
			GenerateName: "dep-",
			Labels: map[string]string{
				"ns": g.Namespace.Name,
			},
		},
		Parent: getKind(first) + "/" + first.Name,
		Child:  getKind(second) + "/" + second.Name,
		Meta:   meta,
	}
	_, err := g.Client.Dependencies().Create(dep)
	Expect(err).NotTo(HaveOccurred())
}

func (g GraphFramework) Connect(first, second *client.ResourceDefinition) {
	g.ConnectWithMeta(first, second, map[string]string{})
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
