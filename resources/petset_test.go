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

// TestPetSetSuccessCheck checks status of ready PetSet
func TestPetSetSuccessCheck(t *testing.T) {
	c := mocks.NewClient()
	status, err := petSetStatus(c.PetSets(), "notfail", c)

	if err != nil {
		t.Error(err)
	}

	if status != "ready" {
		t.Errorf("Status should be `ready`, is `%s` instead.", status)
	}
}

// TestPetSetFailCheck checks status of not ready petset
func TestPetSetFailCheck(t *testing.T) {
	c := mocks.NewClient()
	status, err := petSetStatus(c.PetSets(), "fail", c)

	expectedError := "Resource pod/pending-lolo0 is not ready"
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != "not ready" {
		t.Errorf("Status should be `not ready`, is `%s` instead.", status)
	}
}
