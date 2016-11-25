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
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
)

type Client struct {
	unversioned.PodInterface
	unversioned.JobInterface
	unversioned.ServiceInterface
	unversioned.ReplicaSetInterface
	unversioned.PetSetInterface
	unversioned.DaemonSetInterface
	unversioned.ConfigMapsInterface
	unversioned.SecretsInterface
	client.DependenciesInterface
	client.ResourceDefinitionsInterface
}

func (c *Client) Pods() unversioned.PodInterface {
	return c.PodInterface
}

func (c *Client) Jobs() unversioned.JobInterface {
	return c.JobInterface
}

func (c *Client) Services() unversioned.ServiceInterface {
	return c.ServiceInterface
}

func (c *Client) ReplicaSets() unversioned.ReplicaSetInterface {
	return c.ReplicaSetInterface
}

func (c *Client) PetSets() unversioned.PetSetInterface {
	return c.PetSetInterface
}

//DaemonSets return a DaemonSetInterface of k8s
func (c *Client) DaemonSets() unversioned.DaemonSetInterface {
	return c.DaemonSetInterface
}

func (c *Client) Dependencies() client.DependenciesInterface {
	return c.DependenciesInterface
}

func (c *Client) ResourceDefinitions() client.ResourceDefinitionsInterface {
	return c.ResourceDefinitionsInterface
}

func (c *Client) ConfigMaps() unversioned.ConfigMapsInterface {
	return c.ConfigMapsInterface
}

func (c *Client) Secrets() unversioned.SecretsInterface {
	return c.SecretsInterface
}

func NewClient() *Client {
	return &Client{
		NewPodClient(),
		NewJobClient(),
		NewServiceClient(),
		NewReplicaSetClient(),
		NewPetSetClient(),
		NewDaemonSetClient(),
		NewConfigMapClient(),
		NewSecretClient(),
		NewDependencyClient(),
		NewResourceDefinitionClient(),
	}
}
