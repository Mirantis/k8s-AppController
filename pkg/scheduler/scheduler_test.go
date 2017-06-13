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

package scheduler

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
	"github.com/Mirantis/k8s-AppController/pkg/resources"
)

func TestBuildDependencyGraph(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakePod("ready-1"),
		mocks.MakePod("ready-2"),
		mocks.MakeResourceDefinition("pod/ready-1"),
		mocks.MakeResourceDefinition("pod/ready-2"),
		mocks.MakeDependency("pod/ready-1", "pod/ready-2"),
	)

	sched := New(c, nil, 0)

	dg, err := sched.BuildDependencyGraph(interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Error(err)
		return
	}

	depGraph := dg.(*dependencyGraph).graph

	if len(depGraph) != 2 {
		t.Errorf("wrong length of dependency graph, expected %d, actual %d",
			2, len(depGraph))
	}

	sr, ok := depGraph["pod/ready-1"]

	if !ok {
		t.Errorf("dependency for '%s' not found in dependency graph", "pod/ready-1")
	}

	if sr.Key() != "pod/ready-1" {
		t.Errorf("wrong scheduled resource key, expected '%s', actual '%s'",
			"pod/ready-1", sr.Key())
	}

	if len(sr.requires) != 0 {
		t.Errorf("wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.requires))
	}

	if len(sr.requiredBy) != 1 {
		t.Errorf("wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.requires))
	}

	if sr.requiredBy[0] != "pod/ready-2" {
		t.Errorf("wrong value of 'RequiredBy' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-2", sr.requiredBy[0])
		return
	}

	sr, ok = (depGraph)["pod/ready-2"]

	if !ok {
		t.Errorf("dependency for '%s' not found in dependency graph", "pod/ready-2")
		return
	}

	if sr.Key() != "pod/ready-2" {
		t.Errorf("wrong scheduled resource key, expected '%s', actual '%s'",
			"pod/ready-2", sr.Key())
	}

	if len(sr.requires) != 1 {
		t.Errorf("wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.requires))
	}

	if sr.requires[0] != "pod/ready-1" {
		t.Errorf("wrong value of 'Requires' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-1", sr.requires[0])
	}

	if len(sr.requiredBy) != 0 {
		t.Errorf("wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.requires))
	}
}

func TestIsBlocked(t *testing.T) {
	depGraph := newDependencyGraph(nil, interfaces.DependencyGraphOptions{})
	context := &graphContext{graph: depGraph}

	one := &scheduledResource{
		Resource:         mocks.NewResource("fake1", 0),
		dependenciesMeta: map[string]map[string]string{},
		context:          context,
	}

	depGraph.graph["fake1"] = one

	if one.isBlocked() {
		t.Error("scheduled resource is blocked but it must not")
	}

	two := &scheduledResource{
		Resource:         mocks.NewResource("fake2", 1),
		dependenciesMeta: map[string]map[string]string{},
		context:          context,
	}

	depGraph.graph["fake2"] = two

	one.requires = []string{"fake2"}

	if one.isBlocked() {
		t.Error("scheduled resource is blocked but it must not")
	}

	two.error = errors.New("non-nil error")
	if !one.isBlocked() {
		t.Error("scheduled resource is not blocked but it must be")
	}

	depGraph.graph["fake3"] = &scheduledResource{
		Resource: mocks.NewResource("fake3", 0),
		context:  context,
	}

	two.error = nil
	one.requires = append(one.requires, "fake3")

	if !one.isBlocked() {
		t.Error("scheduled resource is not blocked but it must be")
	}
}

func detectCyclesHelper(dependencies []client.Dependency) [][]string {
	defFlow := newDefaultFlowObject()
	flows := map[string]*client.Flow{
		"flow" + interfaces.DefaultFlowName: defFlow,
	}
	return detectCycles(dependencies, flows, defFlow, false)
}

func TestIsBlockedWithOnErrorDependency(t *testing.T) {
	depGraph := newDependencyGraph(nil, interfaces.DependencyGraphOptions{})
	context := &graphContext{graph: depGraph}

	one := &scheduledResource{
		Resource:         mocks.NewResource("fake1", 0),
		dependenciesMeta: map[string]map[string]string{"fake2": {"on-error": "true"}},
		context:          context,
	}
	depGraph.graph["fake1"] = one

	if one.isBlocked() {
		t.Error("scheduled resource is blocked but it must be not")
	}

	two := &scheduledResource{
		Resource: mocks.NewResource("fake2", 0),
		context:  context,
	}
	depGraph.graph["fake2"] = two

	one.requires = []string{"fake2"}

	if !one.isBlocked() {
		t.Error("scheduled resource is not blocked but it must be")
	}

	two.error = errors.New("non-nil error")
	if one.isBlocked() {
		t.Error("scheduled resource is blocked but it must be not")
	}
}

func TestDetectCyclesAcyclic(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-2"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 0 {
		t.Errorf("cycles detected in an acyclic graph: %v", cycles)
	}
}

func TestDetectCyclesSimpleCycle(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-2"),
		*mocks.MakeDependency("pod/ready-2", "pod/ready-1"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 1 {
		t.Errorf("expected %d cycles, got %d", 1, len(cycles))
		return
	}
}

func TestDetectCyclesSelf(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-1"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 1 {
		t.Errorf("expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 2 {
		t.Errorf("expected cycle length to be %d, got %d", 2, len(cycles[0]))
	}

	if cycles[0][0] != "pod/ready-1" {
		t.Errorf("expected cycle node key to be %s, got %s", "pod/ready-1", cycles[0][0])
	}
	if cycles[0][1] != "pod/ready-1" {
		t.Errorf("expected cycle node key to be %s, got %s", "pod/ready-1", cycles[0][0])
	}
}

func TestDetectCyclesLongCycle(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/1", "pod/2"),
		*mocks.MakeDependency("pod/2", "pod/3"),
		*mocks.MakeDependency("pod/3", "pod/4"),
		*mocks.MakeDependency("pod/4", "pod/1"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 1 {
		t.Errorf("expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 4 {
		t.Errorf("expected cycle length to be %d, got %d", 4, len(cycles[0]))
	}
}

func TestDetectCyclesComplex(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/1", "pod/2"),
		*mocks.MakeDependency("pod/2", "pod/3"),
		*mocks.MakeDependency("pod/3", "pod/1"),
		*mocks.MakeDependency("pod/4", "pod/1"),
		*mocks.MakeDependency("pod/1", "pod/5"),
		*mocks.MakeDependency("pod/5", "pod/4"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 1 {
		t.Errorf("expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 5 {
		t.Errorf("expected cycle length to be %d, got %d", 5, len(cycles[0]))
	}
}

func TestDetectCyclesMultiple(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/1", "pod/2"),
		*mocks.MakeDependency("pod/2", "pod/3"),
		*mocks.MakeDependency("pod/3", "pod/4"),
		*mocks.MakeDependency("pod/4", "pod/2"),
		*mocks.MakeDependency("pod/1", "pod/5"),
		*mocks.MakeDependency("pod/5", "pod/6"),
		*mocks.MakeDependency("pod/6", "pod/7"),
		*mocks.MakeDependency("pod/7", "pod/5"),
	}

	cycles := detectCyclesHelper(dependencies)

	if len(cycles) != 2 {
		t.Errorf("expected %d cycles, got %d", 2, len(cycles))
		return
	}

	if len(cycles[0]) != 3 {
		t.Errorf("expected cycle length to be %d, got %d", 3, len(cycles[0]))
	}

	if len(cycles[1]) != 3 {
		t.Errorf("expected cycle length to be %d, got %d", 3, len(cycles[1]))
	}
}

func TestLimitConcurrency(t *testing.T) {
	for concurrency := range [...]int{0, 3, 5, 10} {
		counter := mocks.NewCounterWithMemo()

		sched := &scheduler{concurrency: concurrency}
		depGraph := newDependencyGraph(sched, interfaces.DependencyGraphOptions{})

		for i := 0; i < 15; i++ {
			key := fmt.Sprintf("resource%d", i)
			r := mocks.NewCountingResource(key, counter, time.Second/4)
			context := &graphContext{graph: depGraph}
			sr := newScheduledResourceFor(r, "", context, false, nil)
			depGraph.graph[sr.Key()] = sr
		}
		stopChan := make(chan struct{})
		depGraph.Deploy(stopChan)

		// Concurrency = 0, means 'disabled' i.e. equal to depGraph size
		if concurrency == 0 {
			concurrency = len(depGraph.graph)
		}
		if counter.Max() != concurrency {
			t.Errorf("expected max concurrency counter %d but got %d", concurrency, counter.Max())
		}
	}
}

func TestStopBeforeDeploymentStarted(t *testing.T) {
	depGraph := newDependencyGraph(&scheduler{}, interfaces.DependencyGraphOptions{})
	sr := &scheduledResource{
		Resource: mocks.NewResource("fake1", 0),
	}
	depGraph.graph[sr.Key()] = sr
	stopChan := make(chan struct{})
	close(stopChan)
	depGraph.Deploy(stopChan)
	progress, _ := sr.Resource.GetProgress()
	if progress != 0 {
		t.Errorf("expected that resource %v wont be in ready progress but got %v", sr.Key(), progress)
	}
}

// TestGraphAllResourceTypes aims to test if all resource types supported by AppController are able to be part of deployment graph
func TestGraphAllResourceTypes(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakePod("ready-1"),
		mocks.MakeJob("ready-2"),
		mocks.MakeReplicaSet("ready-3"),
		mocks.MakeService("ready-4"),
		mocks.MakeStatefulSet("ready-5"),
		mocks.MakeDaemonSet("ready-6"),
		mocks.MakeConfigMap("cfg-1"),
		mocks.MakeDeployment("ready-7"),
		mocks.MakePersistentVolumeClaim("pvc-1"),

		mocks.MakeResourceDefinition("pod/ready-1"),
		mocks.MakeResourceDefinition("job/ready-2"),
		mocks.MakeResourceDefinition("replicaset/ready-3"),
		mocks.MakeResourceDefinition("service/ready-4"),
		mocks.MakeResourceDefinition("statefulset/ready-5"),
		mocks.MakeResourceDefinition("daemonset/ready-6"),
		mocks.MakeResourceDefinition("configmap/cfg-1"),
		mocks.MakeResourceDefinition("secret/secret-1"),
		mocks.MakeResourceDefinition("deployment/ready-7"),
		mocks.MakeResourceDefinition("persistentvolumeclaim/pvc-1"),
		mocks.MakeResourceDefinition("serviceaccount/sa-1"),

		mocks.MakeDependency("pod/ready-1", "job/ready-2"),
		mocks.MakeDependency("job/ready-2", "replicaset/ready-3"),
		mocks.MakeDependency("replicaset/ready-3", "service/ready-4"),
		mocks.MakeDependency("service/ready-4", "statefulset/ready-5"),
		mocks.MakeDependency("statefulset/ready-5", "daemonset/ready-6"),
		mocks.MakeDependency("job/ready-2", "configmap/cfg-1"),
		mocks.MakeDependency("job/ready-2", "secret/secret-1"),
		mocks.MakeDependency("statefulset/ready-5", "deployment/ready-7"),
		mocks.MakeDependency("deployment/ready-7", "persistentvolumeclaim/pvc-1"),
		mocks.MakeDependency("pod/ready-1", "serviceaccount/sa-1"),
	)

	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}

	graph := depGraph.(*dependencyGraph).graph

	expectedLenght := 11
	if len(graph) != expectedLenght {
		ks := make([]string, 0, len(graph))
		for key := range graph {
			ks = append(ks, key)
		}
		keys := strings.Join(ks, ", ")
		t.Errorf("wrong length of dependency graph, expected %d, actual %d. Keys are: %s",
			expectedLenght, len(graph), keys)
	}
}

func TestEmptyStatus(t *testing.T) {
	c := mocks.NewClient()
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	status := depGraph.GetDeploymentStatus()
	if status.Total != 0 {
		t.Errorf("expected total number to be zero but got %d", status.Total)
	}
}

func TestPreparedStatus(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeJob("1"),
		mocks.MakeJob("2"),

		mocks.MakeResourceDefinition("job/1"),
		mocks.MakeResourceDefinition("job/2"),

		mocks.MakeDependency("job/1", "job/2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	status := depGraph.GetDeploymentStatus()
	if status.Total != 2 || status.Failed > 0 || status.Skipped > 0 || status.Progress > 0 || status.Replicas != 1 ||
		status.Finished > 0 || status.TotalGroups != 2 {

		t.Errorf("got unexpected status %v", status)
	}
}

func TestRunningStatus(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeJob("ready-1"),
		mocks.MakeJob("2"),

		mocks.MakeResourceDefinition("job/ready-1"),
		mocks.MakeResourceDefinition("job/2"),

		mocks.MakeDependency("job/ready-1", "job/2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	status := depGraph.GetDeploymentStatus()
	if status.Total != 2 || status.Failed > 0 || status.Skipped > 0 || status.Progress != 0.5 || status.Replicas != 1 ||
		status.Finished != 1 || status.TotalGroups != 2 {

		t.Errorf("got unexpected status %v", status)
	}
}

func TestFinishedStatus(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeJob("ready-1"),
		mocks.MakeJob("ready-2"),

		mocks.MakeResourceDefinition("job/ready-1"),
		mocks.MakeResourceDefinition("job/ready-2"),

		mocks.MakeDependency("job/ready-1", "job/ready-2"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	status := depGraph.GetDeploymentStatus()
	if status.Total != 2 || status.Failed > 0 || status.Skipped > 0 || status.Progress != 1 || status.Replicas != 1 ||
		status.Finished != 2 || status.TotalGroups != 2 {

		t.Errorf("got unexpected status %v", status)
	}
}

// TestGraphNodeStatuses tests a node status report
func TestGraphNodeStatuses(t *testing.T) {
	c := mocks.NewClient(
		mocks.MakeJob("1"),
		mocks.MakeJob("ready-2"),
		mocks.MakeJob("3"),

		mocks.MakeResourceDefinition("job/1"),
		mocks.MakeResourceDefinition("job/ready-2"),
		mocks.MakeResourceDefinition("job/3"),

		mocks.MakeDependency("job/ready-2", "job/1"),
		mocks.MakeDependency("job/3", "job/1"),
	)
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(
		interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	nodeStatuses := depGraph.GetNodeStatuses()

	if len(nodeStatuses) != 3 {
		t.Fatalf("wrong length of a graph 3 != %d", len(nodeStatuses))
	}
	expected := []interfaces.NodeStatus{
		{Name: "job/3", Status: "Starting / in progress", Progress: 0},
		{Name: "job/ready-2", Status: "Finished", Progress: 100},
		{Name: "job/1", Status: "Waiting for job/3", Progress: 0},
	}
	for i, e := range expected {
		if e != nodeStatuses[i] {
			t.Errorf("unexpected report entry %d", i)
		}
	}
}

func makeTaskFunc(i int32, res bool, acc *int32) func(<-chan struct{}) bool {
	return func(<-chan struct{}) bool {
		time.Sleep(time.Second / 10)
		atomic.AddInt32(acc, i)
		return res
	}
}

func runTaskFuncs(t *testing.T, funcs []func(<-chan struct{}) bool, concurrency int, expectedSum int32, expectedResult bool, threshold float64, acc *int32) {
	*acc = 0
	before := time.Now()
	res := runConcurrently(funcs, concurrency, nil)
	after := time.Now()
	if *acc != expectedSum {
		t.Errorf("runConcurrently failed: %d != %d", expectedSum, acc)
	}
	if res != expectedResult {
		t.Error("runConcurrently returned unexpected result")
	}
	if after.Sub(before).Seconds() < threshold-0.01 {
		t.Error("runConcurrently finished too fast")
	}
	if after.Sub(before).Seconds() > threshold+0.05 {
		t.Error("runConcurrently was running too long")
	}
}

// TestRunConcurrently tests runConcurrently function which runs list of concurrent tasks
func TestRunConcurrently(t *testing.T) {
	var acc int32
	var funcs []func(<-chan struct{}) bool
	for i := int32(1); i <= 50; i++ {
		funcs = append(funcs, makeTaskFunc(i, true, &acc))
	}
	var expected int32 = (1 + 50) * 50 / 2
	runTaskFuncs(t, funcs, 0, expected, true, 0.1, &acc)
	runTaskFuncs(t, funcs, 10, expected, true, 0.5, &acc)
	runTaskFuncs(t, funcs, 25, expected, true, 0.2, &acc)
	runTaskFuncs(t, funcs, 50, expected, true, 0.1, &acc)
	runTaskFuncs(t, funcs, 100, expected, true, 0.1, &acc)

	funcs = append(funcs, makeTaskFunc(100, false, &acc))
	runTaskFuncs(t, funcs, 100, expected+100, false, 0.1, &acc)
}

// TestWaitWithZeroTimeout tests resource status wait method with zero timeout value
func TestWaitWithZeroTimeout(t *testing.T) {
	pod := mocks.MakePod("fail")
	resdef := client.ResourceDefinition{Pod: pod}
	options := interfaces.DependencyGraphOptions{FlowName: "test"}
	graph := &dependencyGraph{graphOptions: options}
	gc := &graphContext{graph: graph}
	resource := resources.KindToResourceTemplate["pod"].New(resdef, mocks.NewClient(pod), gc)
	sr := newScheduledResourceFor(resource, "", gc, false, nil)
	stopChan := make(chan struct{})
	defer close(stopChan)

	now := time.Now()
	res, err := sr.wait(CheckInterval, 0, stopChan)
	if res {
		t.Error("wait() succeded")
	}
	if err == nil {
		t.Error("No error was returned")
	} else {
		expectedMessage := "test flow: timeout waiting for resource pod/fail"
		if err.Error() != expectedMessage {
			t.Error("Got unexpected error:", err)
		}
	}
	if time.Now().Sub(now) >= time.Second {
		t.Error("wait() was running for too long")
	}
}
