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
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/Mirantis/k8s-AppController/pkg/client"

	"k8s.io/client-go/pkg/api"
)

var counter int32 = 0

func MakeDependency(parent, child string, label ...string) *client.Dependency {
	labelMap := map[string]string{}
	for _, l := range label {
		parts := strings.Split(l, "=")
		if len(parts) == 2 {
			labelMap[parts[0]] = parts[1]
		}
	}

	c := atomic.AddInt32(&counter, 1)

	return &client.Dependency{
		ObjectMeta: api.ObjectMeta{
			Name:      normalizeName(fmt.Sprintf("%s-%s-%d", parent, child, c)),
			Namespace: "testing",
			Labels:    labelMap,
		},
		Parent: parent,
		Child:  child,
	}
}
