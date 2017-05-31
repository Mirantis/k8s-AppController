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

	"github.com/Mirantis/k8s-AppController/pkg/mocks"

	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

// TestStatefulSetSuccessCheck checks status of ready StatefulSet
func TestStatefulSetSuccessCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeStatefulSet("notfail"))
	progress, err := statefulsetProgress(c, "notfail")

	if err != nil {
		t.Error(err)
	}

	if progress != 1 {
		t.Errorf("progress must be 1 but got %v", progress)
	}
}

// TestStatefulSetFailCheck checks status of not ready statefulset
func TestStatefulSetFailCheck(t *testing.T) {
	ss := mocks.MakeStatefulSet("fail")
	pod := mocks.MakePod("fail")
	pod.Labels = ss.Spec.Template.ObjectMeta.Labels
	c := mocks.NewClient(ss, pod)
	progress, err := statefulsetProgress(c, "fail")

	if err != nil {
		t.Error(err)
	}

	if progress != 0 {
		t.Errorf("progress must be 0 but got %v", progress)
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
