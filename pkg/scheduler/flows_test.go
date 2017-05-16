// Copyright 2017 Mirantis
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

package scheduler

import (
	"errors"
	"strings"
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"

	"k8s.io/client-go/pkg/api"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	batchapiv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

// TestFlowResourcesIdentified tests that the algorithm can identify which resource definitions belong to the flow
// and which are not
func TestFlowResourcesIdentified(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeFlow("test"),

		mocks.MakeResourceDefinition("job/1"),
		mocks.MakeResourceDefinition("job/2"),
		mocks.MakeResourceDefinition("job/3"),
		mocks.MakeResourceDefinition("job/4"),

		mocks.MakeDependency("flow/test", "job/1", "flow=test"),
		mocks.MakeDependency("job/1", "job/2", "flow=test"),
		mocks.MakeDependency("job/2", "job/3"), // job/3 is unreachable from any flow
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			FlowName:     "test",
			ReplicaCount: 1,
		})
	if err != nil {
		t.Fatal(err)
	}

	graph := depGraph.(*DependencyGraph).graph

	expectedLength := 2
	if len(graph) != expectedLength {
		t.Errorf("wrong length of a graph %d != %d", expectedLength, len(graph))
	}

	if _, found := graph["job/1"]; !found {
		t.Error("job/1 is not in the graph")
	}
	if _, found := graph["job/2"]; !found {
		t.Error("job/2 is not in the graph")
	}
}

// TestTopLevelResourcesAssignedToDefaultFlow tests that resource definitions that have no dependencies
// and their dependent resdefs are assigned to default flow. In other words, resdefs that are not reachable
// from any explicit flow belong to default flow
func TestTopLevelResourcesAssignedToDefaultFlow(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeFlow("test"),

		mocks.MakeResourceDefinition("job/1"),
		mocks.MakeResourceDefinition("job/2"),
		mocks.MakeResourceDefinition("job/3"),
		mocks.MakeResourceDefinition("job/4"),
		mocks.MakeResourceDefinition("job/5"),

		// job/4 -> job/5 implicitly belong to default flow
		// job/1 -> job2 explicitly belong to test flow and must not be assigned to default flow
		// job/3 is not reachable from any of the flows because it is neither a top-level resource
		// (because it depends on job/2) not it has a label that would make it part of test flow
		mocks.MakeDependency("flow/test", "job/1", "flow=test"),
		mocks.MakeDependency("job/1", "job/2", "flow=test"),
		mocks.MakeDependency("job/2", "job/3"),
		mocks.MakeDependency("job/4", "job/5"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}

	graph := depGraph.(*DependencyGraph).graph

	expectedLength := 2
	if len(graph) != expectedLength {
		t.Errorf("wrong length of a graph %d != %d", expectedLength, len(graph))
	}

	if _, found := graph["job/4"]; !found {
		t.Error("job/4 is not in the graph")
	}
	if _, found := graph["job/5"]; !found {
		t.Error("job/5 is not in the graph")
	}
}

// TestTriggerFlowIndependently tests that flow is correctly triggered on request
func TestTriggerFlowIndependently(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeFlow("flow1"),
		mocks.MakeFlow("flow2"),

		mocks.MakeResourceDefinition("job/ready-flow1-1"),
		mocks.MakeResourceDefinition("job/ready-flow1-2"),
		mocks.MakeResourceDefinition("job/ready-flow2-1"),
		mocks.MakeResourceDefinition("job/ready-flow2-2"),

		mocks.MakeDependency("flow/flow1", "job/ready-flow1-1", "flow=flow1"),
		mocks.MakeDependency("job/ready-flow1-1", "flow/flow2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow1-2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow2-1", "flow=flow2"),
		mocks.MakeDependency("job/ready-flow2-1", "job/ready-flow2-2", "flow=flow2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "flow2"})
	if err != nil {
		t.Fatal(err)
	}

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs.Items) != 0 {
		t.Fatal("Job list is not empty")
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err = c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	jobNames := map[string]bool{
		"ready-flow2-1": true,
		"ready-flow2-2": true,
	}
	for _, job := range jobs.Items {
		if !jobNames[job.Name] {
			t.Error("found unexpected job", job.Name)
		}
		delete(jobNames, job.Name)
	}

	if len(jobNames) != 0 {
		t.Errorf("not all jobs were created: %d jobs were not found", len(jobNames))
	}
}

// TestUseNotExportedFlow tests that by-default only flows that marked with "exported: true" can be explicitly triggered
func TestUseNotExportedFlow(t *testing.T) {
	flow := mocks.MakeFlow("test")
	flow.Flow.Exported = false

	c := mocks.NewClient(flow)
	sched := New(c, nil, 0)
	_, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "test", ExportedOnly: true})

	if err == nil {
		t.Error("not exported flows cannot be run if ExportedOnly is true")
	} else if err.Error() != "flow test is not exported" {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "test"})

	if err != nil {
		t.Fatal(err)
	}
}

// TestTriggerOneFlowFromAnother tests that one flow can be triggered from another
func TestTriggerOneFlowFromAnother(t *testing.T) {
	flow2 := mocks.MakeFlow("flow2")
	// also test that flows that triggered by other flows don't have to be exported
	flow2.Flow.Exported = false
	c := mocks.NewClient(
		mocks.MakeFlow("flow1"),
		flow2,

		mocks.MakeResourceDefinition("job/ready-flow1-1"),
		mocks.MakeResourceDefinition("job/ready-flow1-2"),
		mocks.MakeResourceDefinition("job/ready-flow2-1"),
		mocks.MakeResourceDefinition("job/ready-flow2-2"),

		mocks.MakeDependency("flow/flow1", "job/ready-flow1-1", "flow=flow1"),
		mocks.MakeDependency("job/ready-flow1-1", "flow/flow2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow1-2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow2-1", "flow=flow2"),
		mocks.MakeDependency("job/ready-flow2-1", "job/ready-flow2-2", "flow=flow2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "flow1"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	jobNames := map[string]bool{
		"ready-flow1-1": true,
		"ready-flow1-2": true,
		"ready-flow2-1": true,
		"ready-flow2-2": true,
	}
	for _, job := range jobs.Items {
		if !jobNames[job.Name] {
			t.Error("found unexpected job", job.Name)
		}
		delete(jobNames, job.Name)
	}

	if len(jobNames) != 0 {
		t.Errorf("not all jobs were created: %d jobs were not found", len(jobNames))
	}
}

// TestParameterPassing tests parameter passing between resource definitions (using Dependency objects)
func TestParameterPassing(t *testing.T) {
	dep := mocks.MakeDependency("job/ready-1", "job/ready-$arg1-$arg2")
	dep.Args = map[string]string{
		"arg1": "x",
		"arg2": "y",
	}

	c := mocks.NewClient(
		mocks.MakeResourceDefinition("job/ready-1"),
		mocks.MakeResourceDefinition("job/ready-$arg1-$arg2"),
		dep,
	)

	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	jobNames := map[string]bool{
		"ready-1":   true,
		"ready-x-y": true,
	}
	for _, job := range jobs.Items {
		if !jobNames[job.Name] {
			t.Error("found unexpected job", job.Name)
		}
		delete(jobNames, job.Name)
	}

	if len(jobNames) != 0 {
		t.Errorf("not all jobs were created: %d jobs were not found", len(jobNames))
	}
}

// TestMultipathParameterPassing tests that single parametrized resource definition that is reachable along two
// paths with different parameter values result in two resources created in case when parameter value is part of resource name
func TestMultipathParameterPassing(t *testing.T) {
	dep1 := mocks.MakeDependency("job/ready-1", "job/ready-$arg")
	dep1.Args = map[string]string{"arg": "x"}
	dep2 := mocks.MakeDependency("job/ready-1", "job/ready-$arg")
	dep2.Args = map[string]string{"arg": "y"}

	c := mocks.NewClient(
		mocks.MakeResourceDefinition("job/ready-1"),
		mocks.MakeResourceDefinition("job/ready-$arg"),
		dep1,
		dep2,
	)

	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	jobNames := map[string]bool{
		"ready-1": true,
		"ready-x": true,
		"ready-y": true,
	}
	for _, job := range jobs.Items {
		if !jobNames[job.Name] {
			t.Error("found unexpected job", job.Name)
		}
		delete(jobNames, job.Name)
	}

	if len(jobNames) != 0 {
		t.Errorf("not all jobs were created: %d jobs were not found", len(jobNames))
	}
}

// TestParametrizedFlow tests ability to declare and pass flow parameters and then use them in dependent resource definitions
func TestParametrizedFlow(t *testing.T) {
	flow := mocks.MakeFlow("flow")
	flow.Flow.Parameters = map[string]client.FlowParameter{
		"arg1": mocks.MakeFlowParameter("p"),
		"arg2": mocks.MakeFlowParameter("q"),
	}

	c := mocks.NewClient(
		flow,
		mocks.MakeResourceDefinition("job/ready-$arg1-$arg2"),
		mocks.MakeDependency("flow/flow", "job/ready-$arg1-$arg2", "flow=flow"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount: 1,
			FlowName:     "flow",
			Args: map[string]string{
				"arg1": "x",
			}})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expectedJobName := "ready-x-q"
	if len(jobs.Items) != 1 {
		t.Errorf("wrong number %d of jobs were created", len(jobs.Items))
	} else if jobs.Items[0].Name != expectedJobName {
		t.Errorf("expected job name: %s, found: %s", expectedJobName, jobs.Items[0].Name)
	}
}

// TestAcNameParameter tests ability to pass $AC_NAME of one flow as an argument for another
func TestAcNameParameter(t *testing.T) {
	flow := mocks.MakeFlow("flow2")
	flow.Flow.Parameters = map[string]client.FlowParameter{
		"name": mocks.MakeFlowParameter(""),
	}

	dep := mocks.MakeDependency("flow/flow1", "flow/flow2", "flow=flow1")
	dep.Args = map[string]string{"name": "$AC_NAME"}

	c := mocks.NewClient(
		mocks.MakeFlow("flow1"),
		flow,
		dep,
		mocks.MakeResourceDefinition("job/ready-$name"),
		mocks.MakeDependency("flow/flow2", "job/ready-$name", "flow=flow2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount: 1,
			FlowName:     "flow1",
		})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	replicas, err := c.Replicas().List(api.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(replicas.Items) != 2 {
		t.Fatalf("there should be 2 replicas - 1 for each flow. Found %d replicas", len(replicas.Items))
	}
	for _, replica := range replicas.Items {
		if replica.FlowName == "flow1" {
			acName := replica.ReplicaName()
			jobName := "ready-" + acName
			_, err := c.Jobs().Get(jobName)
			if err != nil {
				t.Error(err)
			}
			return
		}
	}
	t.Error("no flow1 replica found")
}

// TestUseUndeclaredFlowParameter tests that undeclared flow parameters are not evaluated and thus $something
// remains intact. Note, that with real k8s this would cause deployment error since $ is not a valid character  in
// resource name, which is expected behavior since usage of undeclared parameter is an error that should not be masked
func TestUseUndeclaredFlowParameter(t *testing.T) {
	flow := mocks.MakeFlow("flow")
	flow.Flow.Parameters = map[string]client.FlowParameter{
		"arg1": mocks.MakeFlowParameter("p"),
	}

	c := mocks.NewClient(
		flow,
		mocks.MakeResourceDefinition("job/ready-$arg1-$arg2"),
		mocks.MakeDependency("flow/flow", "job/ready-$arg1-$arg2", "flow=flow"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount: 1,
			FlowName:     "flow",
		})

	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expectedJobName := "ready-p-$arg2"
	if len(jobs.Items) != 1 {
		t.Errorf("wrong number %d of jobs were created", len(jobs.Items))
	} else if jobs.Items[0].Name != expectedJobName {
		t.Errorf("expected job name: %s, found: %s", expectedJobName, jobs.Items[0].Name)
	}
}

// TestPassUndeclaredFlowParameter tests that by-default only declared flow arguments can be passed
func TestPassUndeclaredFlowParameter(t *testing.T) {
	flow := mocks.MakeFlow("flow")
	flow.Flow.Parameters = map[string]client.FlowParameter{
		"arg1": mocks.MakeFlowParameter("p"),
	}

	c := mocks.NewClient(
		mocks.MakeFlow("flow"),
		mocks.MakeResourceDefinition("job/ready-1"),
		mocks.MakeDependency("flow/flow", "job/ready-1", "flow=flow"),
	)
	_, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount: 1,
			FlowName:     "flow",
			Args: map[string]string{
				"arg2": "x",
			}})

	if err == nil {
		t.Error("pass of undeclared argument should not succeed")
	} else if err.Error() != "unexpected argument arg2" {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount:        1,
			FlowName:            "flow",
			AllowUndeclaredArgs: true,
			Args: map[string]string{
				"arg2": "x",
			}})

	if err != nil {
		t.Error("pass of undeclared argument should succeed if it was explicitly permited")
	}
}

// TestParameterPassingBetweenFlows tests that parameter can be passed between flows and parameter value can itself
// require evaluation
func TestParameterPassingBetweenFlows(t *testing.T) {
	flow1 := mocks.MakeFlow("flow1")
	flow1.Flow.Parameters = map[string]client.FlowParameter{
		"arg1": mocks.MakeFlowParameter("p"),
	}
	flow2 := mocks.MakeFlow("flow2")
	flow2.Flow.Parameters = map[string]client.FlowParameter{
		"arg21": mocks.MakeFlowParameter("q"),
		"arg22": mocks.MakeFlowParameter("r"),
		"arg23": mocks.MakeFlowParameter("s"),
	}

	dep := mocks.MakeDependency("flow/flow1", "flow/flow2", "flow=flow1")
	dep.Args = map[string]string{
		"arg21": "$arg1",
		"arg22": "x",
	}

	c := mocks.NewClient(
		flow1,
		flow2,
		mocks.MakeResourceDefinition("job/ready-$arg21-$arg22-$arg23"),
		dep,
		mocks.MakeDependency("flow/flow2", "job/ready-$arg21-$arg22-$arg23", "flow=flow2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount: 1,
			FlowName:     "flow1",
			Args: map[string]string{
				"arg1": "t",
			}})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs, err := c.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expectedJobName := "ready-t-x-s"
	if len(jobs.Items) != 1 {
		t.Errorf("wrong number %d of jobs were created", len(jobs.Items))
	} else if jobs.Items[0].Name != expectedJobName {
		t.Errorf("expected job name: %s, found: %s", expectedJobName, jobs.Items[0].Name)
	}
}

func ensureReplicas(client client.Interface, t *testing.T, jobCount, replicaCount int, trace ...string) []batchapiv1.Job {
	var suffix string
	if len(trace) > 0 {
		suffix = " [" + strings.Join(trace, " ") + "]"
	}

	replicas, err := client.Replicas().List(api.ListOptions{})
	if err != nil {
		t.Fatal(err, suffix)
	}

	if len(replicas.Items) != replicaCount {
		t.Fatalf("wrong number of replicas were found: expected %d, found %d%s", replicaCount, len(replicas.Items), suffix)
	}

	jobs, err := client.Jobs().List(api_v1.ListOptions{})
	if err != nil {
		t.Fatal(err, suffix)
	}

	if len(jobs.Items) != jobCount {
		t.Fatalf("wrong number of jobs were found: expected %d, found %d%s", jobCount, len(jobs.Items), suffix)
	}
	return jobs.Items
}

// TestReplication tests basic flow replication capability
func TestReplication(t *testing.T) {
	replicaCount := 3
	c := mocks.NewClient(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b$AC_NAME"),
		mocks.MakeDependency("flow/test", "job/ready-a$AC_NAME", "flow=test"),
		mocks.MakeDependency("job/ready-a$AC_NAME", "job/ready-b$AC_NAME", "flow=test"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: replicaCount, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs := ensureReplicas(c, t, 2*replicaCount, replicaCount)

	for _, job := range jobs {
		if !strings.HasPrefix(job.Name, "ready-a") && !strings.HasPrefix(job.Name, "ready-b") {
			t.Errorf("unexpected job name %s", job.Name)
		}
	}
}

// TestReplicationWithSharedResources tests that resouce definitions that don't have evaluated parts in their name
// are shared across all flow replics
func TestReplicationWithSharedResources(t *testing.T) {
	replicaCount := 3
	c := mocks.NewClient(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b"),
		mocks.MakeDependency("flow/test", "job/ready-a$AC_NAME", "flow=test"),
		mocks.MakeDependency("job/ready-a$AC_NAME", "job/ready-b", "flow=test"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: replicaCount, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs := ensureReplicas(c, t, replicaCount+1, replicaCount)

	fondSharedResource := false
	for _, job := range jobs {
		if job.Name == "ready-b" && !fondSharedResource {
			fondSharedResource = true
			continue
		}
		if !strings.HasPrefix(job.Name, "ready-a") || len(job.Name) == 7 {
			t.Errorf("unexpected job name %s", job.Name)
		}
	}
}

// TestReplicationScaleUp tests ability to increase replica count to either some fixed value or by some delta
func TestReplicationScaleUp(t *testing.T) {
	initialReplicaCount := 3
	adjustedReplicaCount := 4
	replicaCountDelta := 2

	c := mocks.NewClient(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-$AC_NAME"),
		mocks.MakeDependency("flow/test", "job/ready-$AC_NAME", "flow=test"),
	)
	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount:          initialReplicaCount,
			FlowName:              "test",
			FixedNumberOfReplicas: true,
		})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, initialReplicaCount, initialReplicaCount)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount:          adjustedReplicaCount,
			FlowName:              "test",
			FixedNumberOfReplicas: true,
		})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, adjustedReplicaCount, adjustedReplicaCount)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{
			ReplicaCount:          replicaCountDelta,
			FlowName:              "test",
			FixedNumberOfReplicas: false,
		})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	jobs := ensureReplicas(c, t, adjustedReplicaCount+replicaCountDelta, adjustedReplicaCount+replicaCountDelta)

	for _, job := range jobs {
		if !strings.HasPrefix(job.Name, "ready-") || len(job.Name) == 6 {
			t.Errorf("unexpected job name %s", job.Name)
		}
	}
}

// TestNoOpReplication tests flow replication with zero replicas
func TestNoOpReplication(t *testing.T) {
	replicaCount := 0
	c := mocks.NewClient(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-$AC_NAME"),
		mocks.MakeDependency("flow/test", "job/ready-$AC_NAME", "flow=test"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: replicaCount, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)
}

// TestCompositionFlowReplication tests replication of flow that triggers another flow
func TestCompositionFlowReplication(t *testing.T) {
	replicaCount := 2

	c := mocks.NewClient(
		mocks.MakeFlow("flow1"),
		mocks.MakeFlow("flow2"),

		mocks.MakeResourceDefinition("job/ready-flow1-1-$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-flow1-2-$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-flow2-1-$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-flow2-2-$AC_NAME"),

		mocks.MakeDependency("flow/flow1", "job/ready-flow1-1-$AC_NAME", "flow=flow1"),
		mocks.MakeDependency("job/ready-flow1-1-$AC_NAME", "flow/flow2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow1-2-$AC_NAME", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-flow2-1-$AC_NAME", "flow=flow2"),
		mocks.MakeDependency("job/ready-flow2-1-$AC_NAME", "job/ready-flow2-2-$AC_NAME", "flow=flow2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: replicaCount, FlowName: "flow1"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	jobs := ensureReplicas(c, t, 4*replicaCount, 2*replicaCount)

	jobCounts := map[string]int{
		"ready-flow1-1": replicaCount,
		"ready-flow1-2": replicaCount,
		"ready-flow2-1": replicaCount,
		"ready-flow2-2": replicaCount,
	}
	prefixLength := len("ready-flow1-1")
	for _, job := range jobs {
		if jobCounts[job.Name[:prefixLength]] == 0 {
			t.Error("found unexpected job", job.Name)
		} else {
			jobCounts[job.Name[:prefixLength]] = jobCounts[job.Name[:prefixLength]] - 1
		}
	}

	for k, v := range jobCounts {
		if v != 0 {
			t.Errorf("unexpected count of jobs %s: %d != %d", k, replicaCount-v, replicaCount)
		}
	}

	replicas, err := c.Replicas().List(api.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(replicas.Items) != 2*replicaCount {
		t.Errorf("unexpected count of replicas: %d != %d", len(replicas.Items), 2*replicaCount)
	}
}

// TestDestruction test ability to decrease replica count by/to a give value and ensures, that resources
// that belong to deleted replicas are deleted. Shared resources must be deleted with the last replica.
func TestDestruction(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b"),
		mocks.MakeDependency("flow/test", "job/ready-a$AC_NAME", "flow=test"),
		mocks.MakeDependency("job/ready-a$AC_NAME", "job/ready-b", "flow=test"),
	)

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 5, FlowName: "test", FixedNumberOfReplicas: true})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 6, 5)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 3, FlowName: "test", FixedNumberOfReplicas: true})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 4, 3)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -2, FlowName: "test", FixedNumberOfReplicas: false})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	jobs := ensureReplicas(c, t, 2, 1)
	if jobs[0].Name != "ready-b" && jobs[1].Name != "ready-b" {
		t.Error("shared resource was deleted ahead of time")
	}

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "test", FixedNumberOfReplicas: false})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)
}

// TestCompositeFlowDestruction tests destruction of replicas of the flow that triggers another flow
func TestCompositeFlowDestruction(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeFlow("flow1"),
		mocks.MakeFlow("flow2"),
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-c$AC_NAME"),
		mocks.MakeDependency("flow/flow1", "job/ready-a$AC_NAME", "flow=flow1"),
		mocks.MakeDependency("job/ready-a$AC_NAME", "flow/flow2", "flow=flow1"),
		mocks.MakeDependency("flow/flow2", "job/ready-b$AC_NAME", "flow=flow2"),
		mocks.MakeDependency("flow/flow2", "job/ready-c$AC_NAME", "flow=flow2"),
	)

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 5, FlowName: "flow1", FixedNumberOfReplicas: true})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 15, 10)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 3, FlowName: "flow1", FixedNumberOfReplicas: true})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 9, 6)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -2, FlowName: "flow1", FixedNumberOfReplicas: false})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 3, 2)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 0, FlowName: "flow1", FixedNumberOfReplicas: true})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)
}

// TestCleanupResources tests ability to specify resource subgraph that should be created before each replica
// destruction (for example, to perform cleanup action with resource script/container). After destruction both
// cleanup and replica resources must be deleted
func TestCleanupResources(t *testing.T) {
	flow := mocks.MakeFlow("test")
	flow.Flow.Destruction = map[string]string{"flow": "test", "phase": "delete"}

	c, fake := mocks.NewClientWithFake(
		flow,
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-c$AC_NAME"),
		mocks.MakeDependency("flow/test", "job/ready-a$AC_NAME", "flow=test"),
		mocks.MakeDependency("flow/test", "job/ready-b$AC_NAME", "flow=test", "phase=delete"),
		mocks.MakeDependency("job/ready-b$AC_NAME", "job/ready-c$AC_NAME", "flow=test", "phase=delete"),
	)

	var aCreated, bCreated, cCreated, aDeleted, bDeleted, cDeleted bool
	fake.PrependReactor("*", "jobs",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			switch action.GetVerb() {
			case "create":
				ca := action.(k8stesting.CreateAction)
				obj := ca.GetObject().(*batchapiv1.Job)
				if strings.HasPrefix(obj.Name, "ready-a") {
					aCreated = true
				} else if strings.HasPrefix(obj.Name, "ready-b") {
					bCreated = true
				} else if strings.HasPrefix(obj.Name, "ready-c") {
					cCreated = true
				}
			case "delete":
				da := action.(k8stesting.DeleteAction)
				if strings.HasPrefix(da.GetName(), "ready-a") {
					aDeleted = true
				} else if strings.HasPrefix(da.GetName(), "ready-b") {
					bDeleted = true
				} else if strings.HasPrefix(da.GetName(), "ready-c") {
					cDeleted = true
				}
			}

			return false, nil, nil
		})

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 1, 1)
	if !(aCreated && !aDeleted && !bCreated && !bDeleted && !cCreated && !cDeleted) {
		t.Error("invalid state before destrution")
	}

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)
	if !(aCreated && aDeleted && bCreated && bDeleted && cCreated && cDeleted) {
		t.Error("invalid state after destruction")
	}
}

// TestSharedReplicaSpace tests how different flows may share their replicas
func TestSharedReplicaSpace(t *testing.T) {
	parallel := mocks.MakeFlow("parallel")
	parallel.Flow.ReplicaSpace = "myFlow"
	sequential := mocks.MakeFlow("sequential")
	sequential.Flow.ReplicaSpace = "myFlow"

	c := mocks.NewClient(
		parallel,
		sequential,
		mocks.MakeResourceDefinition("job/ready-a$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready-b$AC_NAME"),
		mocks.MakeDependency("flow/parallel", "job/ready-a$AC_NAME", "flow=parallel"),
		mocks.MakeDependency("flow/parallel", "job/ready-b$AC_NAME", "flow=parallel"),
		mocks.MakeDependency("flow/sequential", "job/ready-a$AC_NAME", "flow=sequential"),
		mocks.MakeDependency("job/ready-a$AC_NAME", "job/ready-b$AC_NAME", "flow=sequential"),
	)

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "parallel"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 2, 1)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "sequential"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "sequential"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 2, 1)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "parallel"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0)
}

// TestDeleteExistingResources tests that by-default flow destruction deletes only the resources that were previously
// created by it
func TestDeleteExistingResources(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeJob("ready-1"),
		mocks.MakeFlow("test"),
		mocks.MakeDependency("flow/test", "job/ready-1", "flow=test"),
	)

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 1, 1)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 1, 0, "invalid attempt to delete external resources")

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 1, 1)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -1, FlowName: "test", AllowDeleteExternalResources: true})
	if err != nil {
		t.Fatal(err)
	}
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 0, "valid attempt to delete external resources")
}

// TestCleanupResourcesErrorHandling tests that replica can only be deleted when all of it resources are deleted
func TestCleanupResourcesErrorHandling(t *testing.T) {
	c, fake := mocks.NewClientWithFake(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("job/ready-$AC_NAME"),
		mocks.MakeDependency("flow/test", "job/ready-$AC_NAME", "flow=test"),
	)

	prevented := false
	// this makes fail the first attempt to delete a job
	fake.PrependReactor("delete", "jobs",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			if !prevented {
				prevented = true
				return true, nil, errors.New("resouce cannot be deleted")
			}
			return false, nil, nil
		})

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 2, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 2, 2)

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: -2, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 1, 1)
}

// TestDeploymentRecoveryForRelativeReplicaCount tests that if we deploy graph with relative replica count,
// abort deployment in the middle and then restart it again, we don't get double replica count created
func TestDeploymentRecoveryForRelativeReplicaCount(t *testing.T) {
	c, fake := mocks.NewClientWithFake(
		mocks.MakeFlow("test"),
		mocks.MakeResourceDefinition("pod/ready-$AC_NAME"),
		mocks.MakeResourceDefinition("job/ready"),
		mocks.MakeDependency("flow/test", "pod/ready-$AC_NAME", "flow=test"),
		mocks.MakeDependency("pod/ready-$AC_NAME", "job/ready", "flow=test"),
	)

	stopChan := make(chan struct{})
	prevented := false
	// this makes deployment stop on the first job
	fake.PrependReactor("create", "jobs",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			if !prevented {
				prevented = true
				stopChan <- struct{}{}
				return true, nil, errors.New("resouce cannot be created")
			}
			return false, nil, nil
		})

	sched := New(c, nil, 0)
	depGraph, err := sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 2, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)

	ensureReplicas(c, t, 0, 2)
	replicas, _ := c.Replicas().List(api.ListOptions{})
	for _, r := range replicas.Items {
		if r.Deployed {
			t.Error("there must not be deployed replica if deployment was aborted")
		}
	}

	depGraph, err = sched.BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 2, FlowName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	depGraph.Deploy(stopChan)
	ensureReplicas(c, t, 1, 2)
	replicas, _ = c.Replicas().List(api.ListOptions{})
	for _, r := range replicas.Items {
		if !r.Deployed {
			t.Error("all replica must be deployed after successful graph deployment")
		}
	}
}
