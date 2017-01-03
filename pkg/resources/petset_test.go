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

	"github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestPetSetSuccessCheck checks status of ready PetSet
func TestPetSetSuccessCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakePetSet("notfail"))
	status, err := petsetStatus(c.PetSets(), "notfail", c)

	if err != nil {
		t.Error(err)
	}

	if status != "ready" {
		t.Errorf("Status should be `ready`, is `%s` instead.", status)
	}
}

// TestPetSetFailCheck checks status of not ready statefulset
func TestPetSetFailCheck(t *testing.T) {
	ss := mocks.MakePetSet("fail")
	pod := mocks.MakePod("fail")
	pod.Labels = ss.Spec.Template.ObjectMeta.Labels
	c := mocks.NewClient(ss, pod)
	status, err := petsetStatus(c.PetSets(), "fail", c)

	expectedError := "Resource pod/fail is not ready"
	if err.Error() != expectedError {
		t.Errorf("Expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != "not ready" {
		t.Errorf("Status should be `not ready`, is `%s` instead.", status)
	}
}

func TestPetSetIsEnabled(t *testing.T) {
	c := mocks.NewClient1_4()
	if !c.IsEnabled(v1alpha1.SchemeGroupVersion) {
		t.Errorf("%v expected to be enabled", v1alpha1.SchemeGroupVersion)
	}
}
