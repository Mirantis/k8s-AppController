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
	"strings"

	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api"
	v1 "k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("Flows Suite", func() {
	framework := FlowFramework{ExamplesFramework{testutils.NewAppControllerManager()}}

	It("Example 'flows' should finish and create one graph replica", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{MinReplicaCount: 1})
		Eventually(func() int {
			return framework.countJobs("a-job-", false)
		}).Should(Equal(1), "1 a-job* should have been created")
		Eventually(func() int {
			return framework.countJobs("b-job-", false)
		}).Should(Equal(1), "1 a-job* should have been created")
		Eventually(func() int {
			return framework.countJobs("test-job", true)
		}).Should(Equal(1), "1 test-job should have been created")
		Eventually(func() int {
			return framework.countPods("a-pod-", false)
		}).Should(Equal(1), "1 a-pod* should have been created")
		Eventually(func() int {
			return framework.countPods("b-pod-", false)
		}).Should(Equal(1), "1 a-pod* should have been created")
		Eventually(func() int {
			return framework.countPods("test-pod", true)
		}).Should(Equal(1), "1 test-pod should have been created")
		Eventually(func() int {
			return framework.countReplicas("test-flow-", false)
		}).Should(Equal(2), "2 test-flow* replicas should have been created")
		Eventually(func() int {
			return framework.countReplicas("DEFAULT", true)
		}).Should(Equal(1), "1 DEFAULT flow replica should have been created")
	})

	It("Example 'flows' with replication should finish and create two replicas", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{ReplicaCount: 2})
		Eventually(func() int {
			return framework.countJobs("a-job-", false)
		}).Should(Equal(2), "1 a-job* should have been created")
		Eventually(func() int {
			return framework.countJobs("b-job-", false)
		}).Should(Equal(2), "1 a-job* should have been created")
		Eventually(func() int {
			return framework.countJobs("test-job", true)
		}).Should(Equal(1), "1 test-job should have been created")
		Eventually(func() int {
			return framework.countPods("a-pod-", false)
		}).Should(Equal(2), "1 a-pod* should have been created")
		Eventually(func() int {
			return framework.countPods("b-pod-", false)
		}).Should(Equal(2), "1 a-pod* should have been created")
		Eventually(func() int {
			return framework.countPods("test-pod", true)
		}).Should(Equal(1), "1 test-pod should have been created")
		Eventually(func() int {
			return framework.countReplicas("test-flow-", false)
		}).Should(Equal(4), "2 test-flow* replicas should have been created")
		Eventually(func() int {
			return framework.countReplicas("DEFAULT", true)
		}).Should(Equal(2), "1 DEFAULT flow replica should have been created")
	})

	It("Example 'flows' should cleanup after itself", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{ReplicaCount: 1})

		deleteOptions := interfaces.DependencyGraphOptions{ReplicaCount: 0, FixedNumberOfReplicas: true}
		By("Running appcontroller scheduler")
		task := framework.RunWithOptions(deleteOptions)
		By("Verifying status of deployment")
		framework.VerifyStatus(task, deleteOptions)

		Eventually(func() int {
			return framework.countReplicas("", false)
		}).Should(Equal(0), "0 replicas should remain")
		Eventually(func() int {
			return framework.countJobs("", false)
		}).Should(Equal(0), "0 jobs should remain")
		Eventually(func() int {
			return framework.countPods("", false)
		}).Should(Equal(1), "only AC pod should remain")
	})
})

type FlowFramework struct {
	ExamplesFramework
}

func (ff FlowFramework) countPods(prefix string, equal bool) int {
	pods, err := ff.Client.Pods().List(v1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	count := 0
	for _, item := range pods.Items {
		if item.Status.Phase == v1.PodSucceeded {
			continue
		}
		if equal && item.Name == prefix ||
			!equal && strings.HasPrefix(item.Name, prefix) && len(item.Name) > len(prefix) {
			count++
		}
	}
	return count
}

func (ff FlowFramework) countJobs(prefix string, equal bool) int {
	jobs, err := ff.Client.Jobs().List(v1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	count := 0
	for _, item := range jobs.Items {
		if equal && item.Name == prefix ||
			!equal && strings.HasPrefix(item.Name, prefix) && len(item.Name) > len(prefix) {
			count++
		}
	}
	return count
}

func (ff FlowFramework) countReplicas(prefix string, equal bool) int {
	jobs, err := ff.Client.Replicas().List(api.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	count := 0
	for _, item := range jobs.Items {
		if equal && item.ReplicaSpace == prefix ||
			!equal && strings.HasPrefix(item.ReplicaSpace, prefix) && len(item.ReplicaSpace) > len(prefix) {
			count++
		}
	}
	return count
}
