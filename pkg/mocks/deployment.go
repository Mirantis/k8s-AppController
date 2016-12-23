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

package mocks

import extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

type deploymentClient struct {
}

// MakeDeployment creates mock Deployment
func MakeDeployment(name string) *extbeta1.Deployment {
	deployment := &extbeta1.Deployment{}
	deployment.Name = name
	deployment.Namespace = "testing"
	deployment.Spec.Replicas = pointer(int32(3))
	if name == "fail" {
		deployment.Status.UpdatedReplicas = int32(2)
		deployment.Status.AvailableReplicas = int32(3)
	} else if name == "failav" {
		deployment.Status.UpdatedReplicas = int32(3)
		deployment.Status.AvailableReplicas = int32(2)
	} else {
		deployment.Status.UpdatedReplicas = int32(3)
		deployment.Status.AvailableReplicas = int32(3)
	}

	return deployment
}
