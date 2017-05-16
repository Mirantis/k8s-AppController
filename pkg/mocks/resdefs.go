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
	"log"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"

	"k8s.io/client-go/pkg/api"
)

// MakeResourceDefinition generates sample ResourceDefinition object
func MakeResourceDefinition(name string) *client.ResourceDefinition {
	resName := normalizeName(name)

	rd := &client.ResourceDefinition{
		ObjectMeta: api.ObjectMeta{Name: resName, Namespace: "testing"},
	}

	splitted := strings.Split(name, "/")
	objectType := splitted[0]
	n := strings.Join(splitted[1:], "/")

	switch objectType {
	case "pod":
		rd.Pod = MakePod(n)
	case "job":
		rd.Job = MakeJob(n)
	case "service":
		rd.Service = MakeService(n)
	case "replicaset":
		rd.ReplicaSet = MakeReplicaSet(n)
	case "statefulset":
		rd.StatefulSet = MakeStatefulSet(n)
	case "petset":
		rd.PetSet = MakePetSet(n)
	case "daemonset":
		rd.DaemonSet = MakeDaemonSet(n)
	case "configmap":
		rd.ConfigMap = MakeConfigMap(n)
	case "secret":
		rd.Secret = MakeSecret(n)
	case "deployment":
		rd.Deployment = MakeDeployment(n)
	case "persistentvolumeclaim":
		rd.PersistentVolumeClaim = MakePersistentVolumeClaim(n)
	case "serviceaccount":
		rd.ServiceAccount = MakeServiceAccount(n)
	default:
		log.Fatal("Unrecognized resource type for name ", objectType)
	}
	return rd
}
