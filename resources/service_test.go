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

	"github.com/Mirantis/k8s-AppController/mocks"
)

//TestCheckServiceStatusReady checks if the service status check is fine for healthy service
func TestCheckServiceStatusReady(t *testing.T) {
	c := mocks.NewClient()
	status, err := serviceStatus(c.Services(), "success", c)

	if err != nil {
		t.Errorf("%s", err)
	}

	if status != "ready" {
		t.Errorf("service should be `ready`, is `%s` instead", status)
	}
}

//TestCheckServiceStatusNotReady tests if service which selects failed pods is not ready
func TestCheckServiceStatusNotReady(t *testing.T) {
	c := mocks.NewClient()
	status, err := serviceStatus(c.Services(), "failed", c)

	if err == nil {
		t.Error("Error should be returned, got nil")
	}

	expectedError := "Pod pending-lolo0 is not ready"
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != "not ready" {
		t.Errorf("service should be `not ready`, is `%s` instead", status)
	}
}
