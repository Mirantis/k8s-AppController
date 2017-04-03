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
	newFlow := parametrizeResource(def.Flow, gc).(*client.Flow)

	dep := gc.Dependency()
	depName := strings.Replace(dep.Name, dep.GenerateName, "", 1)

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

func (f *Flow) constructGraph(destroy bool) (interfaces.DependencyGraph, error) {
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

	if destroy {
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

// Triggers the flow deployment like it was the resource creation
func (f *Flow) Create() error {
	graph, err := f.constructGraph(false)
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

// Deletes resources allocated to the flow
func (f Flow) Delete() error {
	graph, err := f.constructGraph(true)
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
