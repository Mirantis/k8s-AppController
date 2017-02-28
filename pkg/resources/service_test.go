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

	"fmt"

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
		t.Fatal("Error should be returned, got nil")
	}
	expectedError := fmt.Sprintf("Resource pod/%v is not ready", pod.Name)
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
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
		t.Error("Error should be returned, got nil")
	}

	expectedError := fmt.Sprintf("Resource job/%v is not ready", job.Name)
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
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
		t.Error("Error should be returned, got nil")
	}

	expectedError := fmt.Sprintf("Resource replicaset/%v is not ready", rc.Name)
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
	}
}
