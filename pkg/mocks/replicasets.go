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

import (
	"strings"

	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// MakeReplicaSet generates sample ReplicaSet object
func MakeReplicaSet(name string) *extbeta1.ReplicaSet {
	status := strings.Split(name, "-")[0]

	replicaSet := &extbeta1.ReplicaSet{}
	replicaSet.Name = name
	replicaSet.Namespace = "testing"
	replicaSet.Spec.Replicas = pointer(int32(2))
	if status != "fail" {
		replicaSet.Status.Replicas = int32(3)
	}

	return replicaSet
}
