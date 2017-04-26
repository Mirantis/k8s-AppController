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

	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

// TestStatefulSetSuccessCheck checks status of ready StatefulSet
func TestStatefulSetSuccessCheck(t *testing.T) {
	name := "notfail"
	ss := mocks.MakeStatefulSet(name)

	client := mocks.NewClient(ss)

	def := MakeDefinition(ss)
	tested := createNewStatefulSet(def, client)

	status, err := tested.Status(nil)
	if err != nil {
		t.Error(err)
	}

	if status != interfaces.ResourceReady {
		t.Errorf("status should be `ready`, is `%s` instead.", status)
	}
}

// TestStatefulSetFailCheck checks status of not ready statefulset
func TestStatefulSetFailCheck(t *testing.T) {
	name := "fail"
	ss := mocks.MakeStatefulSet(name)

	pod := mocks.MakePod(name)
	pod.Labels = ss.Spec.Template.ObjectMeta.Labels

	client := mocks.NewClient(ss, pod)

	def := MakeDefinition(ss)
	tested := createNewStatefulSet(def, client)

	status, err := tested.Status(nil)

	expectedError := "resource pod/fail is not ready"
	if err.Error() != expectedError {
		t.Errorf("expected `%s` as error, got `%s`", expectedError, err.Error())
	}

	if status != interfaces.ResourceNotReady {
		t.Errorf("status should be `not ready`, is `%s` instead.", status)
	}
}

func TestStatefulSetIsEnabled(t *testing.T) {
	c := mocks.NewClient()
	if !c.IsEnabled(v1beta1.SchemeGroupVersion) {
		t.Errorf("%v expected to be enabled", v1beta1.SchemeGroupVersion)
	}
}

func TestStatefulSetDisabledOn14Version(t *testing.T) {
	c := mocks.NewClient1_4()
	if c.IsEnabled(v1beta1.SchemeGroupVersion) {
		t.Errorf("%v expected to be disabled", v1beta1.SchemeGroupVersion)
	}
}

// TestStatefulSetUpgraded tests status behaviour with resource definition differing from object in cluster
func TestStatefulSetUpgraded(t *testing.T) {
	name := "notfail"

	client := mocks.NewClient(mocks.MakeStatefulSet(name))

	def := MakeDefinition(mocks.MakeStatefulSet(name))
	//Make definition differ from client version
	def.StatefulSet.ObjectMeta.Labels = map[string]string{
		"trolo": "lolo",
	}
	tested := createNewStatefulSet(def, client)

	status, err := tested.Status(nil)

	if err != nil {
		t.Error("Error found, expected nil")
	}
	if status != interfaces.ResourceWaitingForUpgrade {
		t.Errorf("Status should be `waiting for upgrade`, is `%s` instead.", status)
	}
}
