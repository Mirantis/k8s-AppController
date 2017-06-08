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

package resources

import (
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestPetSetUpgraded tests status behaviour with resource definition differing from object in cluster
func TestPetSetUpgraded(t *testing.T) {
	t.Skip("Test skipped due to PetSets not being available in fake client anymore. Test is left for compile-time checks.")
	name := "notfail"

	client := mocks.NewClient(mocks.MakePetSet(name))

	def := MakeDefinition(mocks.MakePetSet(name))
	//Make definition differ from client version
	def.PetSet.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}
	tested := newPetSet(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error("Error found, expected nil")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}

// TestPetSetStatus ensures PetSets have implemented comparison function
func TestPetSetStatus(t *testing.T) {
	tested := PetSet{PetSet: mocks.MakePetSet("")}

	if !tested.equalsToDefinition(tested.PetSet) {
		t.Errorf("Definition not equal to itself")
	}
}
