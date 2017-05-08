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
	"reflect"
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestAllocateReplicas tests replica allocation in different configurations
func TestAllocateReplicas(t *testing.T) {
	flow := mocks.MakeFlow("flow").Flow
	c := mocks.NewClient()
	sched := New(c, nil, 0).(*Scheduler)
	newReplicas1, deleteReplicas, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{ReplicaCount: 3})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas1) != 3 {
		t.Fatal("unexpected new replica count", len(newReplicas1))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 3)

	newReplicas2, deleteReplicas, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{ReplicaCount: 1})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas2) != 1 {
		t.Fatal("unexpected new replica count", len(newReplicas2))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 4)

	allReplicas, deleteReplicas, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{
		ReplicaCount:          5,
		FixedNumberOfReplicas: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(allReplicas) != 5 {
		t.Fatal("unexpected new replica count", len(allReplicas))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 5)

	if !reflect.DeepEqual(newReplicas2[0], allReplicas[3]) || !reflect.DeepEqual(newReplicas1, allReplicas[:3]) {
		t.Error("replica list is not stable")
	}

	allReplicas2, deleteReplicas, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{ReplicaCount: 0})
	if err != nil {
		t.Fatal(err)
	}

	if len(allReplicas2) != 5 {
		t.Fatal("unexpected new replica count", len(allReplicas))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 5)

	if !reflect.DeepEqual(allReplicas, allReplicas2) {
		t.Error("replica list is not stable")
	}
}

// TestDeallocateReplicas tests replica de-allocation in different configurations
func TestDeallocateReplicas(t *testing.T) {
	flow := mocks.MakeFlow("flow").Flow
	c := mocks.NewClient()
	sched := New(c, nil, 0).(*Scheduler)
	newReplicas1, deleteReplicas1, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{ReplicaCount: 5})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas1) != 5 {
		t.Fatal("unexpected new replica count", len(newReplicas1))
	}
	if len(deleteReplicas1) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas1))
	}
	ensureReplicas(c, t, 0, 5)

	newReplicas2, deleteReplicas2, err := sched.allocateReplicas(flow, interfaces.DependencyGraphOptions{ReplicaCount: -2})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas2) != 0 {
		t.Fatal("unexpected new replica count", len(newReplicas2))
	}
	if len(deleteReplicas2) != 2 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas2))
	}

	if !reflect.DeepEqual(deleteReplicas2, newReplicas1[3:]) {
		t.Error("last created replicas should have been scheduled for deletion")
	}

	// allocateReplicas can only create new Replica objects in k8s, but not delete existing
	// replicas can only be deleted when all the resources belonging to it are deleted which happens after deployment
	ensureReplicas(c, t, 0, 5)
}

// TestAllocateReplicasMinMax tests replica allocation with min/max constraints applied
func TestAllocateReplicasMinMax(t *testing.T) {
	flow := mocks.MakeFlow("flow").Flow
	c := mocks.NewClient()
	sched := New(c, nil, 0).(*Scheduler)
	newReplicas, deleteReplicas, err := sched.allocateReplicas(flow,
		interfaces.DependencyGraphOptions{ReplicaCount: 3, MinReplicaCount: 5, MaxReplicaCount: 10})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas) != 5 {
		t.Fatal("unexpected new replica count", len(newReplicas))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 5)

	newReplicas, deleteReplicas, _ = sched.allocateReplicas(flow,
		interfaces.DependencyGraphOptions{ReplicaCount: 9, MinReplicaCount: 5, MaxReplicaCount: 10})

	if len(newReplicas) != 5 {
		t.Fatal("unexpected new replica count", len(newReplicas))
	}
	if len(deleteReplicas) != 0 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 10)

	newReplicas, deleteReplicas, err = sched.allocateReplicas(flow,
		interfaces.DependencyGraphOptions{ReplicaCount: -6, MinReplicaCount: 5, MaxReplicaCount: 10})
	if err != nil {
		t.Fatal(err)
	}

	if len(newReplicas) != 0 {
		t.Fatal("unexpected new replica count", len(newReplicas))
	}
	if len(deleteReplicas) != 5 {
		t.Fatal("unexpected doomed replica count", len(deleteReplicas))
	}
	ensureReplicas(c, t, 0, 10)
}

// TestDependencyToFlowMatching tests how AC identifies if dependency belongs to a given flow path or not
func TestDependencyToFlowMatching(t *testing.T) {
	dep1 := client.Dependency{}
	dep2 := client.Dependency{}
	dep2.Labels = map[string]string{"a": "b"}
	dep3 := client.Dependency{}
	dep3.Labels = map[string]string{"a": "b", "c": "d"}
	dep4 := client.Dependency{}
	dep4.Labels = map[string]string{"c": "d"}

	flow1 := client.Flow{}
	flow2 := client.Flow{
		Construction: map[string]string{"a": "b"},
	}
	flow3 := client.Flow{
		Construction: map[string]string{"a": "b"},
		Destruction:  map[string]string{"a": "b", "c": "d"},
	}
	flow4 := client.Flow{
		Construction: map[string]string{"a": "b", "c": "d"},
		Destruction:  map[string]string{"a": "b"},
	}

	table := []struct {
		dep    client.Dependency
		flow   client.Flow
		destr  bool
		result bool
	}{
		{dep1, flow1, false, true},
		{dep1, flow1, true, false},
		{dep2, flow1, false, true},
		{dep2, flow1, true, false},
		{dep3, flow1, false, true},
		{dep3, flow1, true, false},
		{dep4, flow1, false, true},
		{dep4, flow1, true, false},

		{dep1, flow2, false, false},
		{dep1, flow2, true, false},
		{dep2, flow2, false, true},
		{dep2, flow2, true, false},
		{dep3, flow2, false, true},
		{dep3, flow2, true, false},
		{dep4, flow2, false, false},
		{dep4, flow2, true, false},

		{dep1, flow3, false, false},
		{dep1, flow3, true, false},
		{dep2, flow3, false, true},
		{dep2, flow3, true, false},
		{dep3, flow3, false, false},
		{dep3, flow3, true, true},
		{dep4, flow3, false, false},
		{dep4, flow3, true, false},

		{dep1, flow4, false, false},
		{dep1, flow4, true, false},
		{dep2, flow4, false, false},
		{dep2, flow4, true, true},
		{dep3, flow4, false, true},
		{dep3, flow4, true, true},
		{dep4, flow4, false, false},
		{dep4, flow4, true, false},
	}

	for i, item := range table {
		if canDependencyBelongToFlow(&item.dep, &item.flow, item.destr) != item.result {
			t.Error("invalid dependency match", i)
		}
	}
}
