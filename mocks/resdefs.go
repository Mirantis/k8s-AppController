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

	"github.com/Mirantis/k8s-AppController/client"
	"k8s.io/kubernetes/pkg/api"
)

type resDefClient struct {
	Names []string
}

func (r *resDefClient) List(opts api.ListOptions) (*client.ResourceDefinitionList, error) {
	list := &client.ResourceDefinitionList{}

	for _, name := range r.Names {
		rd := client.ResourceDefinition{}

		splitted := strings.Split(name, "/")
		typ := splitted[0]
		n := strings.Join(splitted[1:], "/")

		if typ == "pod" {
			rd.Pod = MakePod(n)
		} else if typ == "job" {
			rd.Job = MakeJob(n)
		} else if typ == "service" {
			rd.Service = MakeService(n)
		} else if typ == "replicaset" {
			rd.ReplicaSet = MakeReplicaSet(n)
		} else if typ == "petset" {
			rd.PetSet = MakePetSet(n)
		} else {
			log.Fatal("Unrecognized resource type for name ", typ)
		}

		list.Items = append(list.Items, rd)
	}

	return list, nil
}

func NewResourceDefinitionClient(names ...string) client.ResourceDefinitionsInterface {
	return &resDefClient{names}
}
