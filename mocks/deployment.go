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

	"k8s.io/client-go/1.5/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/api"
	extbeta1 "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type deploymentClient struct {
}

// MakeDeployment creates mock Deployment
func MakeDeployment(name string) *extbeta1.Deployment {
	deployment := &extbeta1.Deployment{}
	deployment.Name = name
	deployment.Spec.Replicas = pointer(int32(3))
	if name == "fail" {
		deployment.Status.UpdatedReplicas = int32(2)
		deployment.Status.AvailableReplicas = int32(3)
	} else if name == "failav" {
		deployment.Status.UpdatedReplicas = int32(3)
		deployment.Status.AvailableReplicas = int32(2)
	} else {
		deployment.Status.UpdatedReplicas = int32(3)
		deployment.Status.AvailableReplicas = int32(3)
	}

	return deployment
}

func (r *deploymentClient) List(opts api.ListOptions) (*extbeta1.DeploymentList, error) {
	var deployments []extbeta1.Deployment
	for i := 0; i < 3; i++ {
		deployments = append(deployments, *MakeDeployment(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any failed deployments
	if strings.Index(opts.LabelSelector.String(), "faileddeployment=yes") >= 0 {
		deployments = append(deployments, *MakeDeployment("fail"))
	}

	return &extbeta1.DeploymentList{Items: deployments}, nil
}

func (r *deploymentClient) Get(name string) (*extbeta1.Deployment, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeDeployment(name), nil
}

func (r *deploymentClient) Create(rs *extbeta1.Deployment) (*extbeta1.Deployment, error) {
	return MakeDeployment(rs.Name), nil
}

func (r *deploymentClient) Update(rs *extbeta1.Deployment) (*extbeta1.Deployment, error) {
	panic("not implemented")
}

func (r *deploymentClient) UpdateStatus(rs *extbeta1.Deployment) (*extbeta1.Deployment, error) {
	panic("not implemented")
}

func (r *deploymentClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (r *deploymentClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (r *deploymentClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (r *deploymentClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (r *deploymentClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *extbeta1.Deployment, err error) {
	panic("not implemented")
}

func (r *deploymentClient) Rollback(*extbeta1.DeploymentRollback) error {
	panic("not implemented")
}

// NewDeploymentClient returns mock deployment client
func NewDeploymentClient() v1beta1.DeploymentInterface {
	return &deploymentClient{}
}
