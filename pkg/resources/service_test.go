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

package resources

import (
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestCheckServiceStatusReady checks if the service status check is fine for healthy service
func TestCheckServiceStatusReady(t *testing.T) {
	c := mocks.NewClient(mocks.MakeService("success"))
	progress, err := serviceProgress(c, "success")

	if err != nil {
		t.Errorf("%s", err)
	}

	if progress != 1 {
		t.Errorf("progress must be 1 but got %v", progress)
	}
}

// TestCheckServiceStatusPodNotReady tests if service which selects failed pods is not ready
func TestCheckServiceStatusPodNotReady(t *testing.T) {
	svc := mocks.MakeService("failedpod")
	pod := mocks.MakePod("error")
	pod.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, pod)
	progress, err := serviceProgress(c, "failedpod")

	if err != nil {
		t.Fatal(err)
	}
	if progress != 0 {
		t.Errorf("progress must be 0 but got %v", progress)
	}
}

// TestCheckServiceStatusJobNotReady tests if service which selects failed pods is not ready
func TestCheckServiceStatusJobNotReady(t *testing.T) {
	svc := mocks.MakeService("failedjob")
	job := mocks.MakeJob("error")
	job.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, job)
	progress, err := serviceProgress(c, "failedjob")

	if err != nil {
		t.Fatal(err)
	}

	if progress != 0 {
		t.Errorf("progress must be 0 but got %v", progress)
	}
}

// TestCheckServiceStatusReplicaSetNotReady tests if service which selects failed replicasets is not ready
func TestCheckServiceStatusReplicaSetNotReady(t *testing.T) {
	svc := mocks.MakeService("failedrc")
	rc := mocks.MakeReplicaSet("fail")
	rc.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, rc)
	progress, err := serviceProgress(c, "failedrc")

	if err != nil {
		t.Fatal(err)
	}

	if progress != 0 {
		t.Errorf("progress must be 0 but got %v", progress)
	}
}

// TestCheckServiceStatusOnPartialSelector tests that resources that match only part of selector don't affect service status
func TestCheckServiceStatusOnPartialSelector(t *testing.T) {
	svc := mocks.MakeService("svc")
	svc.Spec.Selector = map[string]string{
		"a": "x",
		"b": "y",
	}
	rs := mocks.MakeReplicaSet("fail")
	rs.Labels = map[string]string{
		"a": "x",
	}

	pod := mocks.MakePod("fail")
	pod.Labels = map[string]string{
		"b": "y",
	}
	job := mocks.MakeJob("ready")
	job.Labels = map[string]string{
		"a": "x",
		"b": "y",
	}

	c := mocks.NewClient(svc, rs, pod, job)
	progress, err := serviceProgress(c, "svc")

	if progress != 1 {
		t.Errorf("progress must be 0.5 but got %v", progress)
	}
	if err != nil {
		t.Fatalf("error should be nil, got %v", err)
	}

	rs2 := mocks.MakeReplicaSet("fail-2")
	rs2.Labels = map[string]string{
		"a": "x",
		"b": "y",
		"c": "z",
	}
	c.ReplicaSets().Create(rs2)

	progress, err = serviceProgress(c, "svc")

	if err != nil {
		t.Error(err)
	}

	if progress != 0.5 {
		t.Errorf("progress must be 0.5 but got %v", progress)
	}
}
