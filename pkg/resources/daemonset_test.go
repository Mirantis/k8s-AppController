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

// TestDaemonSetSuccessCheck check status for ready DaemonSet
func TestDaemonSetSuccessCheck(t *testing.T) {
	name := "not-fail"

	client := mocks.NewClient(mocks.DaemonSets(name)).DaemonSets()

	def := MakeDefinition(mocks.MakeDaemonSet(name))
	tested := createNewDaemonSet(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}
	if status != interfaces.ResourceReady {
		t.Errorf("status should be ready , is %s instead", status)
	}
}

// TestDaemonSetFailCheck status of not ready daemonset
func TestDaemonSetFailCheck(t *testing.T) {
	name := "fail"

	client := mocks.NewClient(mocks.DaemonSets(name)).DaemonSets()

	def := MakeDefinition(mocks.MakeDaemonSet(name))
	tested := createNewDaemonSet(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}
	if status != interfaces.ResourceNotReady {
		t.Errorf("status should be not ready, is %s instead.", status)
	}
}

// TestDaemonSetUpgraded tests status behaviour with resource definition differing from object in cluster
func TestDaemonSetUpgraded(t *testing.T) {
	name := "not-fail"

	client := mocks.NewClient(mocks.DaemonSets(name)).DaemonSets()

	def := MakeDefinition(mocks.MakeDaemonSet(name))

	//Make definition differ from client version
	def.DaemonSet.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}

	tested := createNewDaemonSet(def, client)

	status, err := tested.Status(nil)

	if err == nil {
		t.Error("Error not found, expected error")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}
