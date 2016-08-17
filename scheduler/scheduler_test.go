package scheduler

import (
	"testing"

	"github.com/Mirantis/k8s-AppController/mocks"
)

func TestBuildDependencyGraph(t *testing.T) {
	c := mocks.NewClient()
	c.ResourceDefinitionsInterface = mocks.NewResourceDefinitionClient("pod/ready-1", "pod/ready-2")
	c.DependenciesInterface = mocks.NewDependencyClient(
		mocks.Dependency{Parent: "pod/ready-1", Child: "pod/ready-2"})

	depGraphPtr, err := BuildDependencyGraph(c, nil)
	if err != nil {
		t.Error(err)
	}

	depGraph := *depGraphPtr

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
	one := &ScheduledResource{Status: Init}

	if one.IsBlocked() {
		t.Errorf("Scheduled resource is blocked but it must not")
	}

	two := &ScheduledResource{Status: Ready}
	three := &ScheduledResource{Status: Ready}

	one.Requires = []*ScheduledResource{two, three}

	if one.IsBlocked() {
		t.Errorf("Scheduled resource is blocked but it must not")
	}

	one.Requires[0].Status = Creating

	if !one.IsBlocked() {
		t.Errorf("Scheduled resource is not blocked but it must be")
	}
}
