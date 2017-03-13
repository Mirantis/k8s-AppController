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

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/resources"
	"k8s.io/client-go/pkg/api"
)

// DependencyGraph is a full deployment depGraph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph struct {
	graph     map[string]*ScheduledResource
	scheduler *Scheduler
}

type GraphContext struct {
	args      map[string]string
	graph     *DependencyGraph
	scheduler *Scheduler
}

// Returns the Scheduler that was used to create the dependency graph
func (gc GraphContext) Scheduler() interfaces.Scheduler {
	return gc.scheduler
}

// Returns the currently running dependency graph
func (gc GraphContext) Graph() interfaces.DependencyGraph {
	return gc.graph
}

// newScheduledResourceFor returns new scheduled resource for given resource in init state
func newScheduledResourceFor(r interfaces.Resource) *ScheduledResource {
	return &ScheduledResource{
		Started:  false,
		Ignored:  false,
		Error:    nil,
		Resource: r,
		Meta:     map[string]map[string]string{},
	}
}

func groupDependencies(dependencies []client.Dependency,
	resDefs map[string]client.ResourceDefinition) map[string][]client.Dependency {

	result := map[string][]client.Dependency{}
	isDependant := map[string]bool{}

	for _, dependency := range dependencies {
		group := result[dependency.Parent]
		if group == nil {
			group = []client.Dependency{dependency}
		} else {
			group = append(group, dependency)
		}
		result[dependency.Parent] = group
		isDependant[dependency.Child] = true
	}

	rootDependencies := []client.Dependency{}
	addResource := func(name string) {
		if !isDependant[name] {
			rootDependencies = append(rootDependencies, client.Dependency{Parent: "", Child: name})
			isDependant[name] = true
		}
	}

	for parent := range result {
		addResource(parent)
	}
	for resDef := range resDefs {
		addResource(resDef)
	}
	result[""] = rootDependencies
	return result
}

func getResourceName(resourceDefinition client.ResourceDefinition) (string, string) {
	for _, factory := range resources.KindToResourceTemplate {
		if result := factory.ShortName(resourceDefinition); result != "" {
			return result, factory.Kind()
		}
	}
	return "", ""
}

func (sched *Scheduler) getResourceDefinitions() (map[string]client.ResourceDefinition, error) {
	resDefList, err := sched.client.ResourceDefinitions().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}
	result := map[string]client.ResourceDefinition{}
	for _, resDef := range resDefList.Items {
		name, kind := getResourceName(resDef)
		if name == "" {
			return nil, fmt.Errorf("Invalid resource definition %s", resDef.Name)
		}
		result[kind+"/"+name] = resDef
	}
	return result, nil
}

// newScheduledResource is a constructor for ScheduledResource
func (sched Scheduler) newScheduledResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc GraphContext) (*ScheduledResource, error) {

	var r interfaces.Resource

	resourceTemplate, ok := resources.KindToResourceTemplate[kind]
	if !ok {
		return nil, fmt.Errorf("Not a proper resource kind: %s. Expected '%s'",
			kind, strings.Join(resources.Kinds, "', '"))
	}
	r, err := sched.newResource(kind, name, resDefs, gc, resourceTemplate)
	if err != nil {
		return nil, err
	}

	return newScheduledResourceFor(r), nil
}

func (sched Scheduler) newResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc GraphContext, resourceTemplate interfaces.ResourceTemplate) (interfaces.Resource, error) {
	rd, ok := resDefs[kind+"/"+name]
	if ok {
		log.Printf("Found resource definition for %s/%s", kind, name)
		return resourceTemplate.New(rd, sched.client, gc), nil
	}

	log.Printf("Resource definition for '%s/%s' not found, so it is expected to exist already", kind, name)
	r := resourceTemplate.NewExisting(name, sched.client, gc)
	if r == nil {
		return nil, fmt.Errorf("Existing resource %s/%s cannot be reffered", kind, name)
	}
	return r, nil
}

func keyParts(key string) (kind, name string, err error) {
	parts := strings.Split(key, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("Not a proper resource key: %s. Expected KIND/NAME", key)
	}

	return parts[0], parts[1], nil
}

func NewDependencyGraph(sched *Scheduler) *DependencyGraph {
	return &DependencyGraph{
		graph:     make(map[string]*ScheduledResource),
		scheduler: sched,
	}
}

// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func (sched *Scheduler) BuildDependencyGraph() (interfaces.DependencyGraph, error) {

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

	dependencies := groupDependencies(depList.Items, resDefs)

	depGraph := NewDependencyGraph(sched)

	queue := list.New()
	queue.PushFront("")
	context := GraphContext{scheduler: sched, graph: depGraph}

	for e := queue.Front(); e != nil; e = e.Next() {
		el := e.Value.(string)

		deps := dependencies[el]
		parent := depGraph.graph[el]

		for _, dep := range deps {
			sr := depGraph.graph[dep.Child]
			if sr == nil {
				kind, name, err := keyParts(dep.Child)
				log.Printf("Adding resource %s to the graph", dep.Child)

				sr, err = sched.newScheduledResource(kind, name, resDefs, context)
				if err != nil {
					return nil, err
				}
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
