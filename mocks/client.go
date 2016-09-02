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
	"github.com/Mirantis/k8s-AppController/client"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type Client struct {
	unversioned.PodInterface
	unversioned.JobInterface
	unversioned.ServiceInterface
	unversioned.ReplicaSetInterface
	unversioned.DaemonSetInterface
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

func (c *Client) DaemonSets() unversioned.DaemonSetInterface {
	return c.DaemonSetInterface
}

func (c *Client) Dependencies() client.DependenciesInterface {
	return c.DependenciesInterface
}

func (c *Client) ResourceDefinitions() client.ResourceDefinitionsInterface {
	return c.ResourceDefinitionsInterface
}

func NewClient() *Client {
	return &Client{
		NewPodClient(),
		NewJobClient(),
		NewServiceClient(),
		NewReplicaSetClient(),
		NewDaemonSetClient(),
		NewDependencyClient(),
		NewResourceDefinitionClient(),
	}
}
