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
	"fmt"
	"log"

	"github.com/Mirantis/k8s-AppController/interfaces"
)

var KindToResource = map[string]interfaces.Resource{
	"daemonset":  DaemonSet{},
	"job":        Job{},
	"petset":     PetSet{},
	"pod":        Pod{},
	"replicaset": ReplicaSet{},
	"service":    Service{},
}

var Kinds = getKeys(KindToResource)

func getKeys(m map[string]interfaces.Resource) (keys []string) {
	for key, _ := range m {
		keys = append(keys, key)
	}

	return keys
}

func resourceListReady(resources []interfaces.Resource) (string, error) {
	for _, r := range resources {
		log.Printf("Checking status for resource %s", r.Key())
		status, err := r.Status(nil)
		if err != nil {
			return "error", err
		}
		if status != "ready" {
			return "not ready", fmt.Errorf("Resource %s is not ready", r.Key())
		}
	}
	return "ready", nil
}
