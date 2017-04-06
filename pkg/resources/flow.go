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
	"fmt"
	"log"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

type Flow struct {
	Base
	flow          *client.Flow
	context       interfaces.GraphContext
	status        interfaces.ResourceStatus
	generatedName string
	originalName  string
}

type flowTemplateFactory struct{}

// Returns wrapped resource name if it was a flow
func (flowTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Flow == nil {
		return ""
	}
	return definition.Flow.Name
}

// k8s resource kind that this fabric supports
func (flowTemplateFactory) Kind() string {
	return "flow"
}

// New returns a new object wrapped as Resource
func (flowTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	newFlow := parametrizeResource(def.Flow, gc, "*").(*client.Flow)

	deps := gc.Dependencies()
	var depName string
	if len(deps) > 0 {
		depName = strings.Replace(deps[0].Name, deps[0].GenerateName, "", 1)
	}

	return report.SimpleReporter{
		BaseResource: &Flow{
			Base:          Base{def.Meta},
			flow:          newFlow,
			context:       gc,
			status:        interfaces.ResourceNotReady,
			generatedName: fmt.Sprintf("%s-%s%s", newFlow.Name, depName, gc.GetArg("AC_NAME")),
			originalName:  def.Flow.Name,
		}}
}

// NewExisting returns a new object based on existing one wrapped as Resource
func (flowTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	log.Fatal("Cannot depend on flow that has no resource definition")
	return nil
}

// Identifier of the object
func (f Flow) Key() string {
	return "flow/" + f.generatedName
}

func (f *Flow) buildDependencyGraph(decreaseReplicaCount bool) (interfaces.DependencyGraph, error) {
	args := map[string]string{}
	for arg := range f.flow.Parameters {
		val := f.context.GetArg(arg)
		if val != "" {
			args[arg] = val
		}
	}
	options := interfaces.DependencyGraphOptions{
		FlowName:         f.originalName,
		Args:             args,
		FlowInstanceName: f.generatedName,
	}

	if decreaseReplicaCount {
		// Delete one replica
		options.ReplicaCount = -1
	} else {
		// Recheck existing replica resources or create one replica if none exist
		options.ReplicaCount = 0
		options.MinReplicaCount = 1
	}

	graph, err := f.context.Scheduler().BuildDependencyGraph(options)
	if err != nil {
		return nil, err
	}
	return graph, nil

}

// Create triggers the flow deployment
// Since flow doesn't have any state of its own call to Create() doesn't produce any new Kubernetes resources
// on its own, but instead triggers deployment of resources that belong to the flow
// Create ensures that at least one flow replica exists. It will create first flow replica upon first call, but
// all subsequent calls will do nothing but check the state of the resources from this replica (and recreate them,
// if they were deleted manually between the calls), which makes Create() be idempotent
func (f *Flow) Create() error {
	graph, err := f.buildDependencyGraph(false)
	if err != nil {
		return err
	}
	go func() {
		stopChan := make(chan struct{})
		graph.Deploy(stopChan)
		f.status = interfaces.ResourceReady
	}()
	return nil
}

// Delete releases resources allocated to the last flow replica (i.e. decreases replica count by 1)
// Note, that unlike Create() method Delete() is not idempotent. However, it doesn't create any issues since
// Delete is called during dlow destruction which can happen only once while Create ensures that at least one flow
// replica exists, and as such can be called any number of times
func (f Flow) Delete() error {
	graph, err := f.buildDependencyGraph(true)
	if err != nil {
		return err
	}
	stopChan := make(chan struct{})
	graph.Deploy(stopChan)
	f.status = interfaces.ResourceReady
	return nil
}

// Current status of the flow deployment
func (f Flow) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return f.status, nil
}
