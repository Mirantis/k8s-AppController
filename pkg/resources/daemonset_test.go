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

	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestDaemonSetSuccessCheck check status for ready DaemonSet
func TestDaemonSetSuccessCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeDaemonSet("not-fail"))
	progress, err := daemonSetProgress(c.DaemonSets(), "not-fail")

	if err != nil {
		t.Error(err)
	}
	if progress != 1 {
		t.Errorf("progress must be 1 but got %v", progress)
	}
}

// TestDaemonSetFailCheck status of not ready daemonset
func TestDaemonSetFailCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeDaemonSet("fail"))
	_, err := daemonSetProgress(c.DaemonSets(), "fail")
	if err != nil {
		t.Error(err)
	}
}
