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
	"fmt"
	"testing"
	"time"

	"github.com/Mirantis/k8s-AppController/mocks"
	"github.com/Mirantis/k8s-AppController/report"
)

func TestBuildDependencyGraph(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/ready-1", "pod/ready-2")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "pod/ready-2"})

	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Error(err)
	}

	if len(depGraph) != 2 {
		t.Errorf("Wrong length of dependency graph, expected %d, actual %d",
			2, len(depGraph))
	}

	sr, ok := depGraph["pod/ready-1"]

	if !ok {
		t.Errorf("Dependency for '%s' not found in dependency graph", "pod/ready-1")
	}

	if sr.Key() != "pod/ready-1" {
		t.Errorf("Wrong scheduled resource key, expected '%s', actual '%s'",
			"pod/ready-1", sr.Key())
	}

	if len(sr.Requires) != 0 {
		t.Errorf("Wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.Requires))
	}

	if len(sr.RequiredBy) != 1 {
		t.Errorf("Wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.Requires))
	}

	if sr.RequiredBy[0].Key() != "pod/ready-2" {
		t.Errorf("Wrong value of 'RequiredBy' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-2", sr.RequiredBy[0].Key())
	}

	sr, ok = (depGraph)["pod/ready-2"]

	if !ok {
		t.Errorf("Dependency for '%s' not found in dependency graph", "pod/ready-2")
	}

	if sr.Key() != "pod/ready-2" {
		t.Errorf("Wrong scheduled resource key, expected '%s', actual '%s'",
			"pod/ready-2", sr.Key())
	}

	if len(sr.Requires) != 1 {
		t.Errorf("Wrong length of 'Requires' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 1, len(sr.Requires))
	}

	if sr.Requires[0].Key() != "pod/ready-1" {
		t.Errorf("Wrong value of 'Requires' for scheduled resource '%s', expected '%s', actual '%s'",
			sr.Key(), "pod/ready-1", sr.Requires[0].Key())
	}

	if len(sr.RequiredBy) != 0 {
		t.Errorf("Wrong length of 'RequiredBy' for scheduled resource '%s', expected %d, actual %d",
			sr.Key(), 0, len(sr.Requires))
	}
}

func TestIsBlocked(t *testing.T) {
	one := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake1", "not ready")},
		Status:   Init,
		Meta:     map[string]map[string]string{},
	}

	if one.IsBlocked() {
		t.Errorf("Scheduled resource is blocked but it must not")
	}

	two := &ScheduledResource{
		Resource: report.SimpleReporter{BaseResource: mocks.NewResource("fake2", "ready")},
		Status:   Ready,
		Meta:     map[string]map[string]string{},
	}

	one.Requires = []*ScheduledResource{two}

	if one.IsBlocked() {
		t.Errorf("Scheduled resource is blocked but it must not")
	}

	three := &ScheduledResource{
		Resource: report.SimpleReporter{mocks.NewResource("fake3", "not ready")},
		Status:   Ready,
		Meta:     map[string]map[string]string{},
	}

	one.Requires = append(one.Requires, three)

	if !one.IsBlocked() {
		t.Errorf("Scheduled resource is not blocked but it must be")
	}
}

func TestDetectCyclesAcyclic(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/ready-1", "pod/ready-2")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "pod/ready-2"})

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 0 {
		t.Errorf("Cycles detected in an acyclic graph: %v", cycles)
	}
}

func TestDetectCyclesSimpleCycle(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/ready-1", "pod/ready-2")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "pod/ready-2"},
		mocks.Dependency{Parent: "pod/ready-2", Child: "pod/ready-1"})

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 1 {
		t.Errorf("Expected %d cycles, got %d", 1, len(cycles))
		return
	}
}

func TestDetectCyclesSelf(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/ready-1")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "pod/ready-1"})

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 1 {
		t.Errorf("Expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 2 {
		t.Errorf("Expected cycle length to be %d, got %d", 2, len(cycles[0]))
	}

	if cycles[0][0].Key() != "pod/ready-1" {
		t.Errorf("Expected cycle node key to be %s, got %s", "pod/ready-1", cycles[0][0].Key())
	}
	if cycles[0][1].Key() != "pod/ready-1" {
		t.Errorf("Expected cycle node key to be %s, got %s", "pod/ready-1", cycles[0][0].Key())
	}
}

func TestDetectCyclesLongCycle(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/1", "pod/2", "pod/3", "pod/4", "pod/5")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/1", Child: "pod/2"},
		mocks.Dependency{Parent: "pod/2", Child: "pod/3"},
		mocks.Dependency{Parent: "pod/3", Child: "pod/4"},
		mocks.Dependency{Parent: "pod/4", Child: "pod/1"},
	)

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 1 {
		t.Errorf("Expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 4 {
		t.Errorf("Expected cycle length to be %d, got %d", 4, len(cycles[0]))
	}
}

func TestDetectCyclesComplex(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/1", "pod/2", "pod/3", "pod/4", "pod/5")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/1", Child: "pod/2"},
		mocks.Dependency{Parent: "pod/2", Child: "pod/3"},
		mocks.Dependency{Parent: "pod/3", Child: "pod/1"},
		mocks.Dependency{Parent: "pod/4", Child: "pod/1"},
		mocks.Dependency{Parent: "pod/1", Child: "pod/5"},
		mocks.Dependency{Parent: "pod/5", Child: "pod/4"},
	)

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 1 {
		t.Errorf("Expected %d cycles, got %d", 1, len(cycles))
		return
	}

	if len(cycles[0]) != 5 {
		t.Errorf("Expected cycle length to be %d, got %d", 5, len(cycles[0]))
	}
}

func TestDetectCyclesMultiple(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"pod/1", "pod/2", "pod/3", "pod/4", "pod/5", "pod/6", "pod/7")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/1", Child: "pod/2"},
		mocks.Dependency{Parent: "pod/2", Child: "pod/3"},
		mocks.Dependency{Parent: "pod/3", Child: "pod/4"},
		mocks.Dependency{Parent: "pod/4", Child: "pod/2"},
		mocks.Dependency{Parent: "pod/1", Child: "pod/5"},
		mocks.Dependency{Parent: "pod/5", Child: "pod/6"},
		mocks.Dependency{Parent: "pod/6", Child: "pod/7"},
		mocks.Dependency{Parent: "pod/7", Child: "pod/5"},
	)

	depGraph, _ := BuildDependencyGraph(c, nil)

	cycles := DetectCycles(depGraph)

	if len(cycles) != 2 {
		t.Errorf("Expected %d cycles, got %d", 2, len(cycles))
		return
	}

	if len(cycles[0]) != 3 {
		t.Errorf("Expected cycle length to be %d, got %d", 3, len(cycles[0]))
	}

	if len(cycles[1]) != 3 {
		t.Errorf("Expected cycle length to be %d, got %d", 3, len(cycles[1]))
	}
}

func TestLimitConcurrency(t *testing.T) {
	for concurrency := range [...]int{0, 3, 5, 10} {
		counter := mocks.NewCounterWithMemo()

		depGraph := DependencyGraph{}

		for i := 0; i < 15; i++ {
			key := fmt.Sprintf("resource%d", i)
			r := report.SimpleReporter{BaseResource: mocks.NewCountingResource(key, counter, time.Second*2)}
			sr := NewScheduledResourceFor(r)
			depGraph[sr.Key()] = sr
		}

		Create(depGraph, concurrency)

		// Concurrency = 0, means 'disabled' i.e. equal to depGraph size
		if concurrency == 0 {
			concurrency = len(depGraph)
		}
		if counter.Max() != concurrency {
			t.Errorf("Expected max concurrency counter %d, but got %d", concurrency, counter.Max())
		}
	}
}

// TestGraphAllResourceTypes aims to test if all resource types supported by AppController are able to be part of deployment graph
func TestGraphAllResourceTypes(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"pod/ready-1",
		"job/ready-2",
		"replicaset/ready-3",
		"service/ready-4",
		"petset/ready-5",
		"daemonset/ready-6",
		"configmap/cfg-1",
		"secret/secret-1",
		"deployment/ready-7",
		"persistentvolumeclaim/pvc-1")

	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "job/ready-2"},
		mocks.Dependency{Parent: "job/ready-2", Child: "replicaset/ready-3"},
		mocks.Dependency{Parent: "replicaset/ready-3", Child: "service/ready-4"},
		mocks.Dependency{Parent: "service/ready-4", Child: "petset/ready-5"},
		mocks.Dependency{Parent: "petset/ready-5", Child: "daemonset/ready-6"},
		mocks.Dependency{Parent: "job/ready-2", Child: "configmap/cfg-1"},
		mocks.Dependency{Parent: "job/ready-2", Child: "secret/secret-1"},
		mocks.Dependency{Parent: "petset/ready-5", Child: "deployment/ready-7"},
		mocks.Dependency{Parent: "deployment/ready-7", Child: "persistentvolumeclaim/pvc-1"})

	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedLenght := 10
	if len(depGraph) != expectedLenght {
		keys := ""
		for key, _ := range depGraph {
			keys = keys + ", " + key
		}
		t.Errorf("Wrong length of dependency graph, expected %d, actual %d. Keys are: %s",
			expectedLenght, len(depGraph), keys)
	}
}

func TestEmptyStatus(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient()
	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != Empty {
		t.Errorf("Expected status to be Empty, but got %s", status)
	}
}

func TestPreparedStatus(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"job/1",
		"job/2",
	)
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "job/1", Child: "job/2"},
	)
	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != Prepared {
		t.Errorf("Expected status to be Prepared, but got %s", status)
	}
}

func TestRunningStatus(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"job/ready-1",
		"job/2",
	)
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "job/ready-1", Child: "job/2"},
	)
	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != Running {
		t.Errorf("Expected status to be Running, but got %s", status)
	}
}

func TestFinishedStatus(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"job/ready-1",
		"job/ready-2",
	)
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "job/ready-1", Child: "job/ready-2"},
	)
	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, _ := depGraph.GetStatus()
	if status != Finished {
		t.Errorf("Expected status to be Finished, but got %s", status)
	}
}

// TestGraph tests a simple DependencyGraph report
func TestGraph(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient(
		"job/1",
		"job/ready-2",
		"job/3",
	)
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "job/ready-2", Child: "job/1"},
		mocks.Dependency{Parent: "job/3", Child: "job/1"},
	)
	depGraph, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, report := depGraph.GetStatus()
	if status != Running {
		t.Errorf("Expected status to be Running, but got %s", status)
	}
	if len(report) != 3 {
		t.Errorf("Wrong length of a graph 3 != %d", len(report))
	}
	for _, nodeReport := range report {
		if nodeReport.Dependent == "job/1" {
			if len(nodeReport.Dependencies) != 2 {
				t.Errorf("Wrong length of dependencies 2 != %d", len(nodeReport.Dependencies))
				for _, dependency := range nodeReport.Dependencies {
					if dependency.Dependency == "job/ready-2" {
						if dependency.Blocks {
							t.Errorf("Job 2 should not block")
						}
					} else if dependency.Dependency == "job/3" {
						if !dependency.Blocks {
							t.Errorf("Job 3 should block")
						}
					} else {
						t.Errorf("Unexpected dependency %s", dependency.Dependency)
					}
				}

			}
		}
	}
}
