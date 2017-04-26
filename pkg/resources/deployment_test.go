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

// TestDeploymentSuccessCheck checks status of ready Deployment
func TestDeploymentSuccessCheck(t *testing.T) {
	name := "notfail"

	client := mocks.NewClient(mocks.MakeDeployment(name)).Deployments()

	def := MakeDefinition(mocks.MakeDeployment(name))
	tested := createNewDeployment(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceReady {
		t.Errorf("status should be `ready`, is `%s` instead.", status)
	}
}

// TestDeploymentFailUpdatedCheck checks status of not ready deployment
func TestDeploymentFailUpdatedCheck(t *testing.T) {
	name := "fail"

	client := mocks.NewClient(mocks.MakeDeployment(name)).Deployments()

	def := MakeDefinition(mocks.MakeDeployment(name))
	tested := createNewDeployment(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("status should be `not ready`, is `%s` instead.", status)
	}
}

// TestDeploymentFailAvailableCheck checks status of not ready deployment
func TestDeploymentFailAvailableCheck(t *testing.T) {
	name := "failav"

	client := mocks.NewClient(mocks.MakeDeployment(name)).Deployments()

	def := MakeDefinition(mocks.MakeDeployment(name))
	tested := createNewDeployment(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("status should be `not ready`, is `%s` instead.", status)
	}
}

// TestDeploymentUpgraded tests status behaviour with resource definition differing from object in cluster
func TestDeploymentUpgraded(t *testing.T) {
	name := "notfail"

	client := mocks.NewClient(mocks.MakeDeployment(name)).Deployments()

	def := MakeDefinition(mocks.MakeDeployment(name))
	//Make definition differ from client version
	def.Deployment.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}
	tested := createNewDeployment(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error("Error found, expected nil")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}
