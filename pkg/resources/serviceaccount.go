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
	"github.com/Mirantis/k8s-AppController/pkg/report"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type ServiceAccount struct {
	Base
	ServiceAccount *v1.ServiceAccount
	Client         corev1.ServiceAccountInterface
}

type ExistingServiceAccount struct {
	Base
	Name   string
	Client corev1.ServiceAccountInterface
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

// New returns ServiceAccount controller for new resource based on resource definition
func (serviceAccountTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	serviceAccount := parametrizeResource(def.ServiceAccount, gc).(*v1.ServiceAccount)
	return report.SimpleReporter{
		BaseResource: ServiceAccount{Base: Base{def.Meta}, ServiceAccount: serviceAccount, Client: c.ServiceAccounts()},
	}
}

// NewExisting returns ServiceAccount controller for existing resource by its name
func (serviceAccountTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingServiceAccount{Name: name, Client: c.ServiceAccounts()}}
}

func serviceAccountKey(name string) string {
	return "serviceaccount/" + name
}

// Key returns ServiceAccount name
func (c ServiceAccount) Key() string {
	return serviceAccountKey(c.ServiceAccount.Name)
}

func serviceAccountStatus(c corev1.ServiceAccountInterface, name string) (interfaces.ResourceStatus, error) {
	_, err := c.Get(name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return interfaces.ResourceReady, nil
}

// Status returns ServiceAccount status
func (c ServiceAccount) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return serviceAccountStatus(c.Client, c.ServiceAccount.Name)
}

// Create looks for the ServiceAccount in k8s and creates it if not present
func (c ServiceAccount) Create() error {
	if err := checkExistence(c); err != nil {
		log.Println("Creating", c.Key())
		c.ServiceAccount, err = c.Client.Create(c.ServiceAccount)
		return err
	}
	return nil
}

// Delete deletes ServiceAccount from the cluster
func (c ServiceAccount) Delete() error {
	return c.Client.Delete(c.ServiceAccount.Name, &v1.DeleteOptions{})
}

// Key returns ServiceAccount name
func (c ExistingServiceAccount) Key() string {
	return serviceAccountKey(c.Name)
}

// Status returns ServiceAccount status
func (c ExistingServiceAccount) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return serviceAccountStatus(c.Client, c.Name)
}

// Create looks for existing ServiceAccount and returns error if there is no such ServiceAccount
func (c ExistingServiceAccount) Create() error {
	return createExistingResource(c)
}

// Delete deletes ServiceAccount from the cluster
func (c ExistingServiceAccount) Delete() error {
	return c.Client.Delete(c.Name, nil)
}
