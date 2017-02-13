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
)

// TestServiceAccountSuccessCheck checks status of ready ServiceAccount
func TestServiceAccountSuccessCheck(t *testing.T) {
	c := mocks.NewClient(mocks.ServiceAccounts("notfail"))
	status, err := serviceAccountStatus(c.ServiceAccounts(), "notfail")

	if err != nil {
		t.Error(err)
	}

	if status != "ready" {
		t.Errorf("Status should be `ready`, is `%s` instead.", status)
	}
}

// TestServiceAccountFailCheck checks status of not existing ServiceAccount
func TestServiceAccountFailCheck(t *testing.T) {
	c := mocks.NewClient(mocks.ServiceAccounts())
	status, err := serviceAccountStatus(c.ServiceAccounts(), "fail")

	if err == nil {
		t.Error("Error not found, expected error")
	}

	if status != "error" {
		t.Errorf("Status should be `error`, is `%s` instead.", status)
	}
}
