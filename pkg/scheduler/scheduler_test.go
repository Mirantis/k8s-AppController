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
	"testing"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
	"github.com/Mirantis/k8s-AppController/pkg/report"
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

	dg, err := sched.BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	depGraph := dg.(*DependencyGraph).graph

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

	if len(sr.Requires) != 0 {
		t.Errorf("wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.Requires))
	}

	if len(sr.RequiredBy) != 1 {
		t.Errorf("wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.Requires))
	}

	if sr.RequiredBy[0].Key() != "pod/ready-2" {
		t.Errorf("wrong value of 'RequiredBy' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-2", sr.RequiredBy[0].Key())
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

	if len(sr.Requires) != 1 {
		t.Errorf("wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.Requires))
	}

	if sr.Requires[0].Key() != "pod/ready-1" {
		t.Errorf("wrong value of 'Requires' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-1", sr.Requires[0].Key())
	}

	if len(sr.RequiredBy) != 0 {
		t.Errorf("wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.Requires))
	}
}

func TestIsBlocked(t *testing.T) {
	one := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake1", "not ready")},
		Meta:     map[string]map[string]string{},
	}

	if one.IsBlocked() {
		t.Error("scheduled resource is blocked but it must not")
	}

	two := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake2", "ready")},
		Meta:     map[string]map[string]string{},
	}

	one.Requires = []*ScheduledResource{two}

	if one.IsBlocked() {
		t.Error("scheduled resource is blocked but it must not")
	}

	two.Error = errors.New("non-nil error")
	if !one.IsBlocked() {
		t.Error("scheduled resource is not blocked but it must be")
	}

	three := &ScheduledResource{
		Resource: report.SimpleReporter{mocks.NewResource("fake3", "not ready")},
		Meta:     map[string]map[string]string{},
	}

	two.Error = nil
	one.Requires = append(one.Requires, three)

	if !one.IsBlocked() {
		t.Errorf("scheduled resource is not blocked but it must be")
	}
}

func TestIsBlockedWithOnErrorDependency(t *testing.T) {
	one := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake1", "not ready")},
		Meta:     map[string]map[string]string{},
	}

	if one.IsBlocked() {
		t.Error("scheduled resource is blocked but it must be not")
	}

	two := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake2", "not ready")},
		Meta:     map[string]map[string]string{},
	}

	one.Requires = []*ScheduledResource{two}
	one.Meta["fake2"] = map[string]string{"on-error": "true"}

	if !one.IsBlocked() {
		t.Error("scheduled resource is not blocked but it must be")
	}

	two.Error = errors.New("non-nil error")
	if one.IsBlocked() {
		t.Error("scheduled resource is blocked but it must be not")
	}
}

func TestDetectCyclesAcyclic(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-2"),
	}

	cycles := detectCycles(dependencies)

	if len(cycles) != 0 {
		t.Errorf("cycles detected in an acyclic graph: %v", cycles)
	}
}

func TestDetectCyclesSimpleCycle(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-2"),
		*mocks.MakeDependency("pod/ready-2", "pod/ready-1"),
	}

	cycles := detectCycles(dependencies)

	if len(cycles) != 1 {
		t.Errorf("expected %d cycles, got %d", 1, len(cycles))
		return
	}
}

func TestDetectCyclesSelf(t *testing.T) {
	dependencies := []client.Dependency{
		*mocks.MakeDependency("pod/ready-1", "pod/ready-1"),
	}

	cycles := detectCycles(dependencies)

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

	cycles := detectCycles(dependencies)

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

	cycles := detectCycles(dependencies)

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

	cycles := detectCycles(dependencies)

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

		sched := &Scheduler{concurrency: concurrency}
		depGraph := NewDependencyGraph(sched, interfaces.DependencyGraphOptions{})

		for i := 0; i < 15; i++ {
			key := fmt.Sprintf("resource%d", i)
			r := report.SimpleReporter{BaseResource: mocks.NewCountingResource(key, counter, time.Second/4)}
			sr := newScheduledResourceFor(r)
			depGraph.graph[sr.Key()] = sr
		}
		stopChan := make(chan struct{})
		depGraph.Deploy(stopChan)

		// Concurrency = 0, means 'disabled' i.e. equal to depGraph size
		if concurrency == 0 {
			concurrency = len(depGraph.graph)
		}
		if counter.Max() != concurrency {
			t.Errorf("expected max concurrency counter %d, but got %d", concurrency, counter.Max())
		}
	}
}

func TestStopBeforeDeploymentStarted(t *testing.T) {
	depGraph := NewDependencyGraph(&Scheduler{}, interfaces.DependencyGraphOptions{})
	sr := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake1", "not ready")},
	}
	depGraph.graph[sr.Key()] = sr
	stopChan := make(chan struct{})
	close(stopChan)
	depGraph.Deploy(stopChan)
	status, _ := sr.Resource.Status(nil)
	if status == interfaces.ResourceReady {
		t.Errorf("expected that resource %v wont be in ready status, but got %v", sr.Key(), status)
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

	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}

	graph := depGraph.(*DependencyGraph).graph

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
	status, _ := depGraph.GetStatus()
	if status != interfaces.Empty {
		t.Errorf("expected status to be Empty, but got %s", status)
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
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != interfaces.Prepared {
		t.Errorf("expected status to be Prepared, but got %s", status)
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
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != interfaces.Running {
		t.Errorf("expected status to be Running, but got %s", status)
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
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != interfaces.Finished {
		t.Errorf("expected status to be Finished, but got %s", status)
	}
}

// TestGraph tests a simple DependencyGraph report
func TestGraph(t *testing.T) {
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
	depGraph, err := New(c, nil, 0).BuildDependencyGraph(interfaces.DependencyGraphOptions{})
	if err != nil {
		t.Fatal(err)
	}
	status, rep := depGraph.GetStatus()
	if status != interfaces.Running {
		t.Errorf("expected status to be Running, but got %s", status)
	}

	deploymentReport := rep.(report.DeploymentReport)

	if len(deploymentReport) != 3 {
		t.Errorf("wrong length of a graph 3 != %d", len(deploymentReport))
	}
	for _, nodeReport := range deploymentReport {
		if nodeReport.Dependent == "job/1" {
			if len(nodeReport.Dependencies) != 2 {
				t.Errorf("wrong length of dependencies 2 != %d", len(nodeReport.Dependencies))
				for _, dependency := range nodeReport.Dependencies {
					if dependency.Dependency == "job/ready-2" {
						if dependency.Blocks {
							t.Error("job 2 should not block")
						}
					} else if dependency.Dependency == "job/3" {
						if !dependency.Blocks {
							t.Error("job 3 should block")
						}
					} else {
						t.Errorf("unexpected dependency %s", dependency.Dependency)
					}
				}

			}
		}
	}
}
