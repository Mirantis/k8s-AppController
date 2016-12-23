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

import appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"

type statefulSetClient struct {
}

func pointer(i int32) *int32 {
	return &i
}

// MakeStatefulSet returns a new K8s StatefulSet object for the client to return. If it's name is "fail" it will have labels that will cause it's underlying mock Pods to fail.
func MakeStatefulSet(name string) *appsbeta1.StatefulSet {
	statefulSet := &appsbeta1.StatefulSet{}
	statefulSet.Name = name
	statefulSet.Namespace = "testing"
	statefulSet.Spec.Replicas = pointer(int32(3))
	statefulSet.Spec.Template.ObjectMeta.Labels = make(map[string]string)
	if name == "fail" {
		statefulSet.Spec.Template.ObjectMeta.Labels["failedpod"] = "yes"
		statefulSet.Status.Replicas = int32(2)
	} else {
		statefulSet.Status.Replicas = int32(3)
	}

	return statefulSet
}
