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

package mocks

import (
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"

	alphafake "github.com/Mirantis/k8s-AppController/pkg/client/petsets/typed/apps/v1alpha1/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/testing"
)

func newClientWithFake(apiVersions *unversioned.APIGroupList, objects ...runtime.Object) (client.Interface, *testing.Fake) {
	ns := "testing"
	fakeClientset := fake.NewSimpleClientset(objects...)
	apps := &alphafake.FakeApps{&fakeClientset.Fake}
	deps := &fakeDeps{fake: &fakeClientset.Fake, ns: ns}
	replicas := &fakeReplicas{fake: &fakeClientset.Fake, ns: ns}
	resDefs := &fakeResDef{fake: &fakeClientset.Fake, ns: ns}
	return client.NewClient(
		fakeClientset,
		apps,
		deps,
		resDefs,
		replicas,
		ns,
		apiVersions), &fakeClientset.Fake
}

func newClient(apiVersions *unversioned.APIGroupList, objects ...runtime.Object) client.Interface {
	c, _ := newClientWithFake(apiVersions, objects...)
	return c
}

func makeVersionsList(version unversioned.GroupVersion) *unversioned.APIGroupList {
	return &unversioned.APIGroupList{Groups: []unversioned.APIGroup{
		{
			Name:     version.Group,
			Versions: []unversioned.GroupVersionForDiscovery{{Version: version.Version}},
		},
	}}
}

// NewClient returns new fake client
func NewClient(objects ...runtime.Object) client.Interface {
	return newClient(makeVersionsList(v1beta1.SchemeGroupVersion), objects...)
}

// NewClientWithFake returns new fake client and *testing.Fake object that can be used to catch client events
func NewClientWithFake(objects ...runtime.Object) (client.Interface, *testing.Fake) {
	return newClientWithFake(makeVersionsList(v1beta1.SchemeGroupVersion), objects...)
}

// NewClient1_4 returns fake client for K8s 1.4.x
func NewClient1_4(objects ...runtime.Object) client.Interface {
	return newClient(makeVersionsList(v1alpha1.SchemeGroupVersion), objects...)
}

func normalizeName(name string) string {
	transformations := map[string]string{
		"/": "-",
		"_": "",
		"$": "",
	}
	for key, value := range transformations {
		name = strings.Replace(name, key, value, -1)
	}
	return name
}
