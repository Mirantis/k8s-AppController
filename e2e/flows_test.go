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
	"strings"
	"time"

	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api"
	v1 "k8s.io/client-go/pkg/api/v1"
)

var _ = Describe("Flows Suite", func() {
	framework := FlowFramework{examplesFramework{testutils.NewAppControllerManager()}}

	It("Example 'flows' should finish and create one graph replica", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{MinReplicaCount: 1})
		framework.validateResourceCounts([]resourceCount{
			{"jobs", "a-job-", false, 1},
			{"jobs", "b-job-", false, 1},
			{"jobs", "test-job", true, 1},
			{"pods", "a-pod-", false, 1},
			{"pods", "b-pod-", false, 1},
			{"pods", "test-pod", true, 1},
			{"replicas", "test-flow", true, 2},
			{"replicas", "DEFAULT", true, 1},
		})
	})

	It("Example 'flows' with replication should finish and create two replicas", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{ReplicaCount: 2})
		framework.validateResourceCounts([]resourceCount{
			{"jobs", "a-job-", false, 2},
			{"jobs", "b-job-", false, 2},
			{"jobs", "test-job", true, 1},
			{"pods", "a-pod-", false, 2},
			{"pods", "b-pod-", false, 2},
			{"pods", "test-pod", true, 1},
			{"replicas", "test-flow", true, 4},
			{"replicas", "DEFAULT", true, 2},
		})
	})

	It("Example 'flows' should cleanup after itself", func() {
		framework.CreateRunAndVerify("flows", interfaces.DependencyGraphOptions{ReplicaCount: 1})

		deleteOptions := interfaces.DependencyGraphOptions{ReplicaCount: 0, FixedNumberOfReplicas: true}
		By("Running appcontroller scheduler")
		framework.RunWithOptions(deleteOptions)
		By("Verifying status of deployment")
		framework.VerifyStatus(deleteOptions)

		framework.validateResourceCounts([]resourceCount{
			{"replicas", "", false, 0},
			{"jobs", "", false, 0},
			{"pods", "", false, 1},
		})
	})
})

type FlowFramework struct {
	examplesFramework
}

type resourceCount struct {
	kind     string
	prefix   string
	equal    bool
	expected int
}

func (ff FlowFramework) validateResourceCounts(counts []resourceCount) {
	for _, item := range counts {
		var suffix string
		if !item.equal {
			suffix = "*"
		}
		Eventually(func() int {
			switch item.kind {
			case "pods":
				return ff.countPods(item.prefix, item.equal)
			case "jobs":
				return ff.countJobs(item.prefix, item.equal)
			case "replicas":
				return ff.countReplicas(item.prefix, item.equal)
			default:
				return -1
			}
		}, 300*time.Second, 5*time.Second).Should(Equal(item.expected),
			fmt.Sprintf("there should be %d %s %s%s", item.expected, item.kind, item.prefix, suffix))
	}

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
	replicas, err := ff.Client.Replicas().List(api.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	count := 0
	for _, item := range replicas.Items {
		if equal && item.ReplicaSpace == prefix ||
			!equal && strings.HasPrefix(item.ReplicaSpace, prefix) && len(item.ReplicaSpace) > len(prefix) {
			count++
		}
	}
	return count
}
