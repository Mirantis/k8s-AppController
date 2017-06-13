// Copyright 2017 Mirantis
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
	"log"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var deploymentParamFields = []string{
	"Spec.Template.Spec.Containers.Name",
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.InitContainers.Name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.ObjectMeta",
}

// deployment is wrapper for K8s Deployment object
type newDeployment struct {
	deployment *extbeta1.Deployment
	client     v1beta1.DeploymentInterface
}

// existingDeployment is a wrapper for K8s Deployment object which is deployed on a cluster before AppController
type existingDeployment struct {
	name   string
	client v1beta1.DeploymentInterface
}

type deploymentTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a deployment
func (deploymentTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Deployment == nil {
		return ""
	}
	return definition.Deployment.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (deploymentTemplateFactory) Kind() string {
	return "deployment"
}

// New returns Deployment controller for new resource based on resource definition
func (deploymentTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	deployment := parametrizeResource(def.Deployment, gc, deploymentParamFields).(*extbeta1.Deployment)
	return newDeployment{deployment: deployment, client: c.Deployments()}
}

// NewExisting returns Deployment controller for existing resource by its name
func (deploymentTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingDeployment{name: name, client: c.Deployments()}
}

func deploymentKey(name string) string {
	return "deployment/" + name
}

func deploymentProgress(d v1beta1.DeploymentInterface, name string) (float32, error) {
	deployment, err := d.Get(name)
	if err != nil {
		return 0, err
	}

	var totalReplicas float32 = 1
	if deployment.Spec.Replicas != nil {
		totalReplicas = float32(*deployment.Spec.Replicas)
	}
	return float32(deployment.Status.AvailableReplicas) / totalReplicas, nil
}

// Key return Deployment name
func (d newDeployment) Key() string {
	return deploymentKey(d.deployment.Name)
}

// GetProgress returns Deployment progress
func (d newDeployment) GetProgress() (float32, error) {
	return deploymentProgress(d.client, d.deployment.Name)
}

// Create looks for the Deployment in K8s and creates it if not present
func (d newDeployment) Create() error {
	if checkExistence(d) {
		return nil
	}
	log.Println("Creating", d.Key())
	obj, err := d.client.Create(d.deployment)
	d.deployment = obj
	return err
}

// Delete deletes Deployment from the cluster
func (d newDeployment) Delete() error {
	return d.client.Delete(d.deployment.Name, nil)
}

// Key returns Deployment name
func (d existingDeployment) Key() string {
	return deploymentKey(d.name)
}

// GetProgress returns Deployment deployment progress
func (d existingDeployment) GetProgress() (float32, error) {
	return deploymentProgress(d.client, d.name)
}

// Create looks for existing Deployment and returns error if there is no such Deployment
func (d existingDeployment) Create() error {
	return createExistingResource(d)
}

// Delete deletes Deployment from the cluster
func (d existingDeployment) Delete() error {
	return d.client.Delete(d.name, nil)
}
