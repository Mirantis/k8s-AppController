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

type flow struct {
	Base
	flow         *client.Flow
	context      interfaces.GraphContext
	originalName string
	instanceName string
	currentGraph interfaces.DependencyGraph
}

type flowTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a flow
func (flowTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Flow == nil {
		return ""
	}
	return definition.Flow.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (flowTemplateFactory) Kind() string {
	return "flow"
}

// New returns Flow controller for new resource based on resource definition
func (flowTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	newFlow := parametrizeResource(def.Flow, gc, []string{"*"}).(*client.Flow)

	dep := gc.Dependency()
	var depName string
	if dep != nil {
		depName = strings.Replace(dep.Name, dep.GenerateName, "", 1)
	}

	return report.SimpleReporter{
		BaseResource: &flow{
			Base:         Base{def.Meta},
			flow:         newFlow,
			context:      gc,
			originalName: def.Flow.Name,
			instanceName: fmt.Sprintf("%s%s", depName, gc.GetArg("AC_NAME")),
		}}
}

// NewExisting returns Flow controller for existing resource by its name. Since flow is not a real k8s resource
// this case is not possible
func (flowTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	log.Fatal("Cannot depend on flow that has no resource definition")
	return nil
}

// Key return Flow identifier
func (f flow) Key() string {
	return "flow/" + f.flow.Name
}

func (f *flow) buildDependencyGraph(replicaCount int, silent bool) (interfaces.DependencyGraph, error) {
	args := map[string]string{}
	for arg := range f.flow.Parameters {
		val := f.context.GetArg(arg)
		if val != "" {
			args[arg] = val
		}
	}
	fixedNumberOfReplicas := false
	if replicaCount > 0 {
		fixedNumberOfReplicas = f.context.Graph().Options().FixedNumberOfReplicas
	} else if replicaCount == 0 {
		fixedNumberOfReplicas = true
		replicaCount = -1
	}
	options := interfaces.DependencyGraphOptions{
		FlowName:              f.originalName,
		Args:                  args,
		FlowInstanceName:      f.instanceName,
		ReplicaCount:          replicaCount,
		Silent:                silent,
		FixedNumberOfReplicas: fixedNumberOfReplicas,
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
func (f *flow) Create() error {
	graph := f.currentGraph

	if graph == nil {
		var err error
		graph, err = f.buildDependencyGraph(1, false)
		if err != nil {
			return err
		}
		f.currentGraph = graph
	}
	go func() {
		stopChan := make(chan struct{})
		graph.Deploy(stopChan)
	}()
	return nil
}

// Delete releases resources allocated to the last flow replica (i.e. decreases replica count by 1)
// Note, that unlike Create() method Delete() is not idempotent. However, it doesn't create any issues since
// Delete is called during dlow destruction which can happen only once while Create ensures that at least one flow
// replica exists, and as such can be called any number of times
func (f flow) Delete() error {
	graph, err := f.buildDependencyGraph(-1, false)
	if err != nil {
		return err
	}
	stopChan := make(chan struct{})
	graph.Deploy(stopChan)
	return nil
}

// Status returns current status of the flow deployment
func (f flow) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	graph := f.currentGraph
	if graph == nil {
		var err error
		graph, err = f.buildDependencyGraph(0, true)
		if err != nil {
			return interfaces.ResourceError, err
		}
	}

	status, _ := graph.GetStatus()

	switch status {
	case interfaces.Empty:
		fallthrough
	case interfaces.Finished:
		return interfaces.ResourceReady, nil
	case interfaces.Prepared:
		fallthrough
	case interfaces.Running:
		return interfaces.ResourceNotReady, nil
	default:
		return interfaces.ResourceError, nil
	}
}
