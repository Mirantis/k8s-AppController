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

package scheduler

import (
	"container/list"
	"fmt"
	"log"
	"strings"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/resources"
)

// DependencyGraph is a full deployment depGraph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph struct {
	graph map[string]*ScheduledResource
	scheduler *Scheduler
}

type GraphContext struct {
	args map[string]string
	graph DependencyGraph
	scheduler *Scheduler
}


func (gc GraphContext) Scheduler() interfaces.Scheduler {
	return gc.scheduler
}

func (gc GraphContext) Args() map[string]string {
	return gc.args
}

func (gc GraphContext) Graph() interfaces.DependencyGraph {
	return gc.graph
}

// NewScheduledResourceFor returns new scheduled resource for given resource in init state
func newScheduledResourceFor(r interfaces.Resource) *ScheduledResource {
	return &ScheduledResource{
		Started:  false,
		Ignored:  false,
		Error:    nil,
		Resource: r,
		Meta:     map[string]map[string]string{},
	}
}


func groupDependencies(dependencies *client.DependencyList) map[string][]client.Dependency {
	result := map[string][]client.Dependency{}
	isDependant := map[string]bool{}

	for _, dependency := range dependencies.Items {
		group := result[dependency.Parent]
		if group == nil {
			group = []client.Dependency{dependency}
		} else {
			group = append(group, dependency)
		}
		result[dependency.Parent] = group
		isDependant[dependency.Child] = true
	}

	defaultFlowName := "flow/" + interfaces.DefaultFlowName
	if defaultFlow := result[defaultFlowName]; defaultFlow == nil {
		defaultFlow = []client.Dependency{}
		for parent := range result {
			if !strings.HasPrefix(parent, "flow/") && !isDependant[parent] {
				defaultFlow = append(defaultFlow, client.Dependency{Parent: defaultFlowName, Child: parent})
			}
		}
		result[defaultFlowName] = defaultFlow
	}
	return result
}

func getResourceName(resourceDefinition client.ResourceDefinition) string {
	for _, factory := range resources.KindToResourceTemplate {
		if result := factory.ShortName(resourceDefinition); result != "" {
			return result
		}
	}
	return ""
}

func (sched *Scheduler) getResourceDefinitions() (map[string]client.ResourceDefinition, error) {
	resDefList, err := sched.client.ResourceDefinitions().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}
	result := map[string]client.ResourceDefinition{}
	for _, resDef := range resDefList.Items {
		name := getResourceName(resDef)
		if name == "" {
			return nil, fmt.Errorf("Invalid resource definition %s", resDef.Name)
		}
		result[name] = resDef
	}
	return result, nil
}

func makeDefaultFlow() *client.Flow {
	return &client.Flow{}
}

func filterDependencies(dependencies map[string][]client.Dependency, parent string, flow *client.Flow) []client.Dependency {
	children := dependencies[parent]
	var result []client.Dependency
	for _, dep := range children {
		matches := true
		if len(flow.Construction) == 0 {
			matches = len(dep.Labels) == 0
		} else {
			for k, v := range flow.Construction {
				if dep.Labels[k] != v {
					matches = false
					break
				}
			}
		}
		if matches {
			result = append(result, dep)
		}
	}
	return result
}

// NewScheduledResource is a constructor for ScheduledResource
func (sched Scheduler) newScheduledResource(kind string, name string,
	resDefs map[string]client.ResourceDefinition, gc GraphContext) (*ScheduledResource, error) {

	var r interfaces.Resource

	resourceTemplate, ok := resources.KindToResourceTemplate[kind]
	if !ok {
		return nil, fmt.Errorf("Not a proper resource kind: %s. Expected '%s'", kind, strings.Join(resources.Kinds, "', '"))
	}
	r, err := sched.newResource(name, resDefs, gc, resourceTemplate)
	if err != nil {
		return nil, err
	}

	return newScheduledResourceFor(r), nil
}

func (sched Scheduler) newResource(name string, resDefs map[string]client.ResourceDefinition,
	gc GraphContext, resourceTemplate interfaces.ResourceTemplate) (interfaces.Resource, error) {
	rd, ok := resDefs[name]
	if ok {
		log.Println("Found resource definition for", name)
		return resourceTemplate.New(rd, sched.client, gc), nil
	}

	log.Printf("Resource definition for '%s' not found, so it is expected to exist already", name)
	r := resourceTemplate.NewExisting(name, sched.client, gc)
	if r == nil {
		return nil, fmt.Errorf("Existing resource %s cannot be reffered", name)
	}
	return r, nil
}

func keyParts(key string) (kind string, name string, err error) {
	parts := strings.Split(key, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("Not a proper resource key: %s. Expected KIND/NAME", key)
	}

	return parts[0], parts[1], nil
}


// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func (sched *Scheduler) BuildDependencyGraph(flowName string, args map[string]string) (interfaces.DependencyGraph, error) {
	log.Println("Looking for the flow", flowName)
	flow, err := sched.client.Flows().Get(flowName)

	if err != nil {
		if errors.IsNotFound(err) {
			if flowName == interfaces.DefaultFlowName {
				flow = makeDefaultFlow()
			} else {
				return nil, fmt.Errorf("Flow %s is not found", flowName)
			}
		} else {
			return nil, err
		}
	} else{
		log.Println("Found ", flowName, flow.Name, flow.Construction)
	}
	fullFlowName := "flow/" + flowName


	log.Println("Getting resource definitions")
	resDefs, err := sched.getResourceDefinitions()
	if err != nil {
		return nil, err
	}

	log.Println("Getting dependencies")
	depList, err := sched.client.Dependencies().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}

	log.Println("Making sure there is no cycles in the depGraph")
	if err = EnsureNoCycles(depList.Items); err != nil {
		return nil, err
	}

	dependencies := groupDependencies(depList)

	depGraph := DependencyGraph{
		graph: make(map[string]*ScheduledResource),
		scheduler: sched,
	}

	if _, ok := dependencies[fullFlowName]; !ok {
		log.Println("Flow depGraph is empty")
		return depGraph, nil
	}

	queue := list.New()
	queue.PushFront(fullFlowName)
	context := GraphContext{scheduler: sched, graph: depGraph}

	for e := queue.Front(); e != nil; e = e.Next() {
		el := e.Value.(string)
		deps := filterDependencies(dependencies, el, flow)
		parent := depGraph.graph[el]

		for _, dep := range deps {
			kind, name, err := keyParts(dep.Child)
			log.Printf("Adding resource %s to the flow depGraph %s", dep.Child, flowName)

			sr, err := sched.newScheduledResource(kind, name, resDefs, context)
			if err != nil {
				return nil, err
			}

			depGraph.graph[dep.Child] = sr
			if parent != nil {
				sr.Requires = append(sr.Requires, parent)
				parent.RequiredBy = append(parent.RequiredBy, sr)
			}
			queue.PushBack(dep.Child)
		}
	}

	return depGraph, nil
}
