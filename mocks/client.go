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
	"k8s.io/client-go/1.5/kubernetes/typed/apps/v1alpha1"
	batchv1 "k8s.io/client-go/1.5/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/kubernetes/typed/extensions/v1beta1"

	"github.com/Mirantis/k8s-AppController/client"
)

// Interface is as an interface for k8s clients. It expands native k8s client interface.
type Client struct {
	corev1.ConfigMapInterface
	corev1.SecretInterface
	corev1.PodInterface
	batchv1.JobInterface
	corev1.ServiceInterface
	v1beta1.ReplicaSetInterface
	v1alpha1.PetSetInterface
	v1beta1.DaemonSetInterface
	v1beta1.DeploymentInterface
	corev1.PersistentVolumeClaimInterface

	client.DependenciesInterface
	client.ResourceDefinitionsInterface
}

func (c *Client) Pods() corev1.PodInterface {
	return c.PodInterface
}

func (c *Client) Jobs() batchv1.JobInterface {
	return c.JobInterface
}

func (c *Client) Services() corev1.ServiceInterface {
	return c.ServiceInterface
}

func (c *Client) ReplicaSets() v1beta1.ReplicaSetInterface {
	return c.ReplicaSetInterface
}

func (c *Client) PetSets() v1alpha1.PetSetInterface {
	return c.PetSetInterface
}

// DaemonSets return a DaemonSetInterface of k8s
func (c *Client) DaemonSets() v1beta1.DaemonSetInterface {
	return c.DaemonSetInterface
}

// Deployments returns mock deployment client
func (c *Client) Deployments() v1beta1.DeploymentInterface {
	return c.DeploymentInterface
}

func (c *Client) Dependencies() client.DependenciesInterface {
	return c.DependenciesInterface
}

func (c *Client) ResourceDefinitions() client.ResourceDefinitionsInterface {
	return c.ResourceDefinitionsInterface
}

func (c *Client) ConfigMaps() corev1.ConfigMapInterface {
	return c.ConfigMapInterface
}

func (c *Client) Secrets() corev1.SecretInterface {
	return c.SecretInterface
}

func (c *Client) PersistentVolumeClaims() corev1.PersistentVolumeClaimInterface {
	return c.PersistentVolumeClaimInterface
}

func NewClient() *Client {
	return &Client{
		NewConfigMapClient(),
		NewSecretClient(),
		NewPodClient(),
		NewJobClient(),
		NewServiceClient(),
		NewReplicaSetClient(),
		NewPetSetClient(),
		NewDaemonSetClient(),
		NewDeploymentClient(),
		NewPersistentVolumeClaimClient(),
		NewDependencyClient(),
		NewResourceDefinitionClient(),
	}
}
