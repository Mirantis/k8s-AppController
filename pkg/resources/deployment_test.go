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

// TestDeploymentSuccessCheck checks status of ready Deployment
func TestDeploymentSuccessCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeDeployment("notfail"))
	_, err := deploymentProgress(c.Deployments(), "notfail")

	if err != nil {
		t.Error(err)
	}
}

// TestDeploymentFailUpdatedCheck checks status of not ready deployment
func TestDeploymentFailUpdatedCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeDeployment("fail"))
	_, err := deploymentProgress(c.Deployments(), "fail")

	if err != nil {
		t.Error(err)
	}
}

// TestDeploymentFailAvailableCheck checks status of not ready deployment
func TestDeploymentFailAvailableCheck(t *testing.T) {
	c := mocks.NewClient(mocks.MakeDeployment("failav"))
	_, err := deploymentProgress(c.Deployments(), "failav")

	if err != nil {
		t.Error(err)
	}
}
