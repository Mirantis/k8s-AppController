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
	"log"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var serviceAccountParamFields = []string{
	"Secrets",
	"ImagePullSecrets",
}

type newServiceAccount struct {
	serviceAccount *v1.ServiceAccount
	client         corev1.ServiceAccountInterface
}

type existingServiceAccount struct {
	name   string
	client corev1.ServiceAccountInterface
}

type serviceAccountTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a serviceaccount
func (serviceAccountTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.ServiceAccount == nil {
		return ""
	}
	return definition.ServiceAccount.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (serviceAccountTemplateFactory) Kind() string {
	return "serviceaccount"
}

// New returns serviceAccount controller for new resource based on resource definition
func (serviceAccountTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	serviceAccount := parametrizeResource(def.ServiceAccount, gc, serviceAccountParamFields).(*v1.ServiceAccount)
	return newServiceAccount{serviceAccount: serviceAccount, client: c.ServiceAccounts()}
}

// NewExisting returns serviceAccount controller for existing resource by its name
func (serviceAccountTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingServiceAccount{name: name, client: c.ServiceAccounts()}
}

func serviceAccountKey(name string) string {
	return "serviceaccount/" + name
}

// Key returns serviceAccount name
func (c newServiceAccount) Key() string {
	return serviceAccountKey(c.serviceAccount.Name)
}

func serviceAccountProgress(c corev1.ServiceAccountInterface, name string) (float32, error) {
	_, err := c.Get(name)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

// GetProgress returns ServiceAccount deployment progress
func (c newServiceAccount) GetProgress() (float32, error) {
	return serviceAccountProgress(c.client, c.serviceAccount.Name)
}

// Create looks for the serviceAccount in k8s and creates it if not present
func (c newServiceAccount) Create() error {
	if checkExistence(c) {
		return nil
	}
	log.Println("Creating", c.Key())
	obj, err := c.client.Create(c.serviceAccount)
	c.serviceAccount = obj
	return err
}

// Delete deletes serviceAccount from the cluster
func (c newServiceAccount) Delete() error {
	return c.client.Delete(c.serviceAccount.Name, &v1.DeleteOptions{})
}

// Key returns serviceAccount name
func (c existingServiceAccount) Key() string {
	return serviceAccountKey(c.name)
}

// GetProgress returns serviceAccount deployment progress
func (c existingServiceAccount) GetProgress() (float32, error) {
	return serviceAccountProgress(c.client, c.name)
}

// Create looks for existing serviceAccount and returns error if there is no such serviceAccount
func (c existingServiceAccount) Create() error {
	return createExistingResource(c)
}

// Delete deletes serviceAccount from the cluster
func (c existingServiceAccount) Delete() error {
	return c.client.Delete(c.name, nil)
}
