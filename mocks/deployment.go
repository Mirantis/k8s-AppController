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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type deploymentClient struct {
}

// MakeDeployment creates mock Deployment
func MakeDeployment(name string) *extensions.Deployment {
	deployment := &extensions.Deployment{}
	deployment.Name = name
	deployment.Spec.Replicas = 3
	if name == "fail" {
		deployment.Status.UpdatedReplicas = 2
		deployment.Status.AvailableReplicas = 3
	} else if name == "failav" {
		deployment.Status.UpdatedReplicas = 3
		deployment.Status.AvailableReplicas = 2
	} else {
		deployment.Status.UpdatedReplicas = 3
		deployment.Status.AvailableReplicas = 3
	}

	return deployment
}

func (r *deploymentClient) List(opts api.ListOptions) (*extensions.DeploymentList, error) {
	var deployments []extensions.Deployment
	for i := 0; i < 3; i++ {
		deployments = append(deployments, *MakeDeployment(fmt.Sprintf("ready-trolo%d", i)))
	}

	//use ListOptions.LabelSelector to check if there should be any failed deployments
	if strings.Index(opts.LabelSelector.String(), "faileddeployment=yes") >= 0 {
		deployments = append(deployments, *MakeDeployment("fail"))
	}

	return &extensions.DeploymentList{Items: deployments}, nil
}

func (r *deploymentClient) Get(name string) (*extensions.Deployment, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeDeployment(name), nil
}

func (r *deploymentClient) Create(rs *extensions.Deployment) (*extensions.Deployment, error) {
	return MakeDeployment(rs.Name), nil
}

func (r *deploymentClient) Update(rs *extensions.Deployment) (*extensions.Deployment, error) {
	panic("not implemented")
}

func (r *deploymentClient) UpdateStatus(rs *extensions.Deployment) (*extensions.Deployment, error) {
	panic("not implemented")
}

func (r *deploymentClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *deploymentClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *deploymentClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

func (r *deploymentClient) Rollback(*extensions.DeploymentRollback) error {
	panic("not implemented")
}

// NewDeploymentClient returns mock deployment client
func NewDeploymentClient() unversioned.DeploymentInterface {
	return &deploymentClient{}
}
