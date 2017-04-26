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

// TestConfigMapSuccessCheck checks status of ready ConfigMap
func TestConfigMapSuccessCheck(t *testing.T) {
	name := "notfail"
	client := mocks.NewClient(mocks.ConfigMaps(name))

	def := MakeDefinition(mocks.MakeConfigMap(name))
	tested := createNewConfigMap(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceReady {
		t.Errorf("status should be `ready`, is `%s` instead.", status)
	}
}

// TestConfigMapFailCheck checks status of not existing ConfigMap
func TestConfigMapFailCheck(t *testing.T) {
	name := "fail"
	client := mocks.NewClient()

	def := MakeDefinition(mocks.MakeConfigMap(name))
	tested := createNewConfigMap(def, client)

	status, err := tested.Status(nil)

	if err == nil {
		t.Error("error not found, expected error")
	}

	if status != interfaces.ResourceError {
		t.Errorf("status should be `error`, is `%s` instead.", status)
	}
}

// TestConfigMapUpgraded tests status behaviour with resource definition differing from object in cluster
func TestConfigMapUpgraded(t *testing.T) {
	name := "notfail"
	client := mocks.NewClient(mocks.ConfigMaps(name))

	def := MakeDefinition(mocks.MakeConfigMap(name))

	//Make definition differ from client version
	def.ConfigMap.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}

	tested := createNewConfigMap(def, client)

	status, err := tested.Status(nil)

	if err == nil {
		t.Error("Error not found, expected error")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}
