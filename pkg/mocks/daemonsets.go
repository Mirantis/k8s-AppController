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
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

type daemonSetClient struct {
}

// MakeDaemonSet creates a daemonset base in its name
func MakeDaemonSet(name string) *extbeta1.DaemonSet {
	daemonSet := &extbeta1.DaemonSet{}
	daemonSet.Name = name
	daemonSet.Namespace = "testing"
	daemonSet.Status.DesiredNumberScheduled = 3
	if name == "fail" {
		daemonSet.Status.CurrentNumberScheduled = 2
	} else {
		daemonSet.Status.CurrentNumberScheduled = 3
	}

	return daemonSet
}

func DaemonSets(names ...string) runtime.Object {
	var daemonsets []extbeta1.DaemonSet
	for _, name := range names {
		daemonsets = append(daemonsets, *MakeDaemonSet(name))
	}
	return &extbeta1.DaemonSetList{Items: daemonsets}
}
