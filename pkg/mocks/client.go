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
	"github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"
	alphafake "github.com/Mirantis/k8s-AppController/pkg/client/petsets/typed/apps/v1alpha1/fake"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

func newClient(objects ...runtime.Object) *client.Client {
	fakeClientset := fake.NewSimpleClientset(objects...)
	apps := &alphafake.FakeApps{&fakeClientset.Fake}
	return &client.Client{
		Clientset: fakeClientset,
		AlphaApps: apps,
		Deps:      NewDependencyClient(),
		ResDefs:   NewResourceDefinitionClient(),
		Namespace: "testing",
	}
}

func makeVersionsList(version unversioned.GroupVersion) *unversioned.APIGroupList {
	return &unversioned.APIGroupList{Groups: []unversioned.APIGroup{
		{
			Name: version.Group,
			Versions: []unversioned.GroupVersionForDiscovery{
				{GroupVersion: version.Version},
			},
		},
	}}
}

func NewClient(objects ...runtime.Object) *client.Client {
	c := newClient(objects...)
	c.ApiVersions = makeVersionsList(v1beta1.SchemeGroupVersion)
	return c
}

func NewClient_1_4(objects ...runtime.Object) *client.Client {
	c := newClient(objects...)
	c.ApiVersions = makeVersionsList(v1alpha1.SchemeGroupVersion)
	return c
}
