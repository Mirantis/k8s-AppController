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

import appsalpha1 "github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"

// MakePetSet returns a new K8s PetSet object for the client to return. If it's name is "fail" it will have labels that will cause it's underlying mock Pods to fail.
func MakePetSet(name string) *appsalpha1.PetSet {
	petSet := &appsalpha1.PetSet{}
	petSet.Name = name
	petSet.Namespace = "testing"
	petSet.Spec.Replicas = pointer(int32(3))
	petSet.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	if name == "fail" {
		petSet.Spec.Template.ObjectMeta.Labels["failedpod"] = "yes"
		petSet.Status.Replicas = int32(2)
	} else {
		petSet.Status.Replicas = int32(3)
	}

	return petSet
}
