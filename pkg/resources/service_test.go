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
	"fmt"
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestCheckServiceStatusReady checks if the service status check is fine for healthy service
func TestCheckServiceStatusReady(t *testing.T) {
	c := mocks.NewClient(mocks.MakeService("success"))
	status, err := serviceStatus(c.Services(), "success", c)

	if err != nil {
		t.Errorf("%s", err)
	}

	if status != interfaces.ResourceReady {
		t.Errorf("service should be `ready`, is `%s` instead", status)
	}
}

// TestCheckServiceStatusPodNotReady tests if service which selects failed pods is not ready
func TestCheckServiceStatusPodNotReady(t *testing.T) {
	svc := mocks.MakeService("failedpod")
	pod := mocks.MakePod("error")
	pod.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, pod)
	status, err := serviceStatus(c.Services(), "failedpod", c)

	if err == nil {
		t.Fatal("error should be returned, got nil")
	}
	expectedError := fmt.Sprintf("resource pod/%v is not ready", pod.Name)
	if err.Error() != expectedError {
		t.Errorf("expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
	}
}

// TestCheckServiceStatusJobNotReady tests if service which selects failed pods is not ready
func TestCheckServiceStatusJobNotReady(t *testing.T) {
	svc := mocks.MakeService("failedjob")
	job := mocks.MakeJob("error")
	job.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, job)
	status, err := serviceStatus(c.Services(), "failedjob", c)

	if err == nil {
		t.Error("error should be returned, got nil")
	}

	expectedError := fmt.Sprintf("resource job/%v is not ready", job.Name)
	if err.Error() != expectedError {
		t.Errorf("expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
	}
}

// TestCheckServiceStatusReplicaSetNotReady tests if service which selects failed replicasets is not ready
func TestCheckServiceStatusReplicaSetNotReady(t *testing.T) {
	svc := mocks.MakeService("failedrc")
	rc := mocks.MakeReplicaSet("fail")
	rc.Labels = svc.Spec.Selector
	c := mocks.NewClient(svc, rc)
	status, err := serviceStatus(c.Services(), "failedrc", c)

	if err == nil {
		t.Error("error should be returned, got nil")
	}

	expectedError := fmt.Sprintf("resource replicaset/%v is not ready", rc.Name)
	if err.Error() != expectedError {
		t.Errorf("expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
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
	status, err := serviceStatus(c.Services(), "svc", c)

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

	status, err = serviceStatus(c.Services(), "svc", c)

	if err == nil {
		t.Fatal("error should be returned, got nil")
	}
	expectedError := fmt.Sprintf("resource replicaset/%s is not ready", rs2.Name)
	if err.Error() != expectedError {
		t.Errorf("expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
	}
}
