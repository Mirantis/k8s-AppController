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
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

type void struct {
	Base
	name string
}

type voidTemplateFactory struct{}

// Returns wrapped resource name. Since void resource cannot be wrapped it is always empty string
func (voidTemplateFactory) ShortName(_ client.ResourceDefinition) string {
	return ""
}

// k8s resource kind that this fabric supports
func (voidTemplateFactory) Kind() string {
	return "void"
}

// New returns new void based on resource definition. Since void cannot be wrapped it is always nil
func (voidTemplateFactory) New(_ client.ResourceDefinition, _ client.Interface, _ interfaces.GraphContext) interfaces.Resource {
	return nil
}

// NewExisting returns new void with specified name
func (voidTemplateFactory) NewExisting(name string, _ client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	name = parametrizeResource(name, gc, []string{"*"}).(string)
	return report.SimpleReporter{BaseResource: void{name: name}}
}

// Key returns void name
func (v void) Key() string {
	return "void/" + v.name
}

// Status always returns "ready"
func (void) Status(_ map[string]string) (interfaces.ResourceStatus, error) {
	return interfaces.ResourceReady, nil
}

// Delete is a no-op method to create resource
func (void) Create() error {
	return nil
}

// Delete is a no-op method to delete resource
func (void) Delete() error {
	return nil
}
