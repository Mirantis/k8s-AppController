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
	"github.com/Mirantis/k8s-AppController/pkg/client"
	alphafake "github.com/Mirantis/k8s-AppController/pkg/client/petsets/typed/apps/v1alpha1/fake"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/runtime"
)

func NewClient(objects ...runtime.Object) *client.Client {
	fakeClientset := fake.NewSimpleClientset(objects...)
	apps := &alphafake.FakeApps{fakeClientset.Fake}
	return &client.Client{
		Clientset: fakeClientset,
		AlphaApps: apps,
		Deps:      NewDependencyClient(),
		ResDefs:   NewResourceDefinitionClient(),
		Namespace: "testing",
	}
}
