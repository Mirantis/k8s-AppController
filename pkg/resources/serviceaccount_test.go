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

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestServiceAccountSuccessCheck checks status of ready ServiceAccount
func TestServiceAccountSuccessCheck(t *testing.T) {
	sa := mocks.MakeServiceAccount("notfail")
	c := mocks.NewClient(sa).ServiceAccounts()

	def := MakeDefinition(sa)

	tested := createNewServiceAccount(def, c)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceReady {
		t.Errorf("status should be `ready`, is `%s` instead.", status)
	}
}

// TestServiceAccountFailCheck checks status of not existing ServiceAccount
func TestServiceAccountFailCheck(t *testing.T) {
	c := mocks.NewClient(mocks.ServiceAccounts()).ServiceAccounts()

	def := MakeDefinition(mocks.MakeServiceAccount(""))

	tested := createNewServiceAccount(def, c)

	status, err := tested.Status(nil)

	if err == nil {
		t.Error("error not found, expected error")
	}

	if status != interfaces.ResourceError {
		t.Errorf("status should be `error`, is `%s` instead.", status)
	}
}

// TestServiceAccountUpgraded tests status behaviour with resource definition differing from object in cluster
func TestServiceAccountUpgraded(t *testing.T) {
	name := "notfail"

	client := mocks.NewClient(mocks.MakeServiceAccount(name)).ServiceAccounts()

	def := MakeDefinition(mocks.MakeServiceAccount(name))
	//Make definition differ from client version
	def.ServiceAccount.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}
	tested := createNewServiceAccount(def, client)

	status, err := tested.Status(nil)

	if err == nil {
		t.Error("Error not found, expected error")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}
