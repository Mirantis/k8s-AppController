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
	"github.com/Mirantis/k8s-AppController/pkg/copier"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/resources"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
)

// DependencyGraph is a full deployment depGraph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph struct {
	graph        map[string]*ScheduledResource
	scheduler    *Scheduler
	graphOptions interfaces.DependencyGraphOptions
}

type GraphContext struct {
	args      map[string]string
	graph     *DependencyGraph
	scheduler *Scheduler
	flow      *client.Flow
}

// Returns the Scheduler that was used to create the dependency graph
func (gc GraphContext) Scheduler() interfaces.Scheduler {
	return gc.scheduler
}

// Returns argument values available in the current graph context
func (gc GraphContext) GetArg(name string) string {
	val, ok := gc.args[name]
	if ok {
		return val
	}
	val, ok = gc.graph.graphOptions.Args[name]
	if ok {
		return val
	}
	fp, ok := gc.flow.Parameters[name]
	if ok && fp.Default != nil {
		return *fp.Default
	}
	return ""
}

// Returns the currently running dependency graph
func (gc GraphContext) Graph() interfaces.DependencyGraph {
	return gc.graph
}

// newScheduledResourceFor returns new scheduled resource for given resource in init state
func newScheduledResourceFor(r interfaces.Resource, context *GraphContext) *ScheduledResource {
	return &ScheduledResource{
		Started:  false,
		Ignored:  false,
		Error:    nil,
		Resource: r,
		Meta:     map[string]map[string]string{},
		Context:  context,
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

	defaultFlowName := "flow/" + interfaces.DefaultFlowName
	if defaultFlow := result[defaultFlowName]; defaultFlow == nil {
		defaultFlow = []client.Dependency{}
		addResource := func(name string) {
			if !strings.HasPrefix(name, "flow/") && !isDependant[name] {
				defaultFlow = append(defaultFlow, client.Dependency{Parent: defaultFlowName, Child: name})
				isDependant[name] = true
			}
		}

		for parent := range result {
			addResource(parent)
		}
		for resDef := range resDefs {
			addResource(resDef)
		}
		result[defaultFlowName] = defaultFlow
	}
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
			return nil, fmt.Errorf("invalid resource definition %s", resDef.Name)
		}
		result[kind+"/"+name] = resDef
	}
	return result, nil
}

func filterDependencies(dependencies map[string][]client.Dependency, parent string,
	flow *client.Flow) []client.Dependency {

	children := dependencies[parent]
	var result []client.Dependency
	for _, dep := range children {
		if canDependencyBelongToFlow(&dep, flow) {
			result = append(result, dep)
		}
	}
	return result
}

func canDependencyBelongToFlow(dep *client.Dependency, flow *client.Flow) bool {
	for k, v := range flow.Construction {
		if dep.Labels[k] != v {
			return false
		}
	}
	return true
}

// newScheduledResource is a constructor for ScheduledResource
func (sched Scheduler) newScheduledResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc *GraphContext) (*ScheduledResource, error) {

	var r interfaces.Resource

	resourceTemplate, ok := resources.KindToResourceTemplate[kind]
	if !ok {
		return nil, fmt.Errorf("not a proper resource kind: %s. Expected '%s'",
			kind, strings.Join(resources.Kinds, "', '"))
	}
	r, err := sched.newResource(kind, name, resDefs, gc, resourceTemplate)
	if err != nil {
		return nil, err
	}

	return newScheduledResourceFor(r, gc), nil
}

func (sched Scheduler) newResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc *GraphContext, resourceTemplate interfaces.ResourceTemplate) (interfaces.Resource, error) {
	rd, ok := resDefs[kind+"/"+name]
	if ok {
		log.Printf("Found resource definition for %s/%s", kind, name)
		return resourceTemplate.New(rd, sched.client, *gc), nil
	}

	log.Printf("Resource definition for '%s/%s' not found, so it is expected to exist already", kind, name)
	r := resourceTemplate.NewExisting(name, sched.client, *gc)
	if r == nil {
		return nil, fmt.Errorf("existing resource %s/%s cannot be reffered", kind, name)
	}
	return r, nil
}

func keyParts(key string) (kind, name string, err error) {
	parts := strings.Split(key, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("not a proper resource key: %s. Expected KIND/NAME", key)
	}

	return parts[0], parts[1], nil
}

// Constructor of DependencyGraph
func NewDependencyGraph(sched *Scheduler, options interfaces.DependencyGraphOptions) *DependencyGraph {
	return &DependencyGraph{
		graph:        make(map[string]*ScheduledResource),
		scheduler:    sched,
		graphOptions: options,
	}
}

func (sched *Scheduler) prepareContext(parentContext *GraphContext, dependency client.Dependency) *GraphContext {
	context := &GraphContext{scheduler: sched, graph: parentContext.graph, flow: parentContext.flow}
	context.args = make(map[string]string)
	for key, value := range dependency.Args {
		context.args[key] = copier.EvaluateString(value, parentContext.GetArg)
	}
	return context
}

func (sched *Scheduler) updateContext(context, parentContext *GraphContext, dependency client.Dependency) {
	for key, value := range dependency.Args {
		context.args[key] = copier.EvaluateString(value, parentContext.GetArg)
	}
}

func (sched *Scheduler) newDefaultFlowObject() *client.Flow {
	return &client.Flow{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Flow",
			APIVersion: client.GroupName + "/" + client.Version,
		},
		ObjectMeta: api.ObjectMeta{
			Name:      interfaces.DefaultFlowName,
			Namespace: sched.client.Namespace(),
		},
		Exported: true,
	}
}

func checkArgs(options interfaces.DependencyGraphOptions, flow *client.Flow) error {
	if !options.AllowUndeclaredArgs {
		for key := range options.Args {
			if _, ok := flow.Parameters[key]; !ok {
				return fmt.Errorf("unexpected argument %s", key)
			}
		}
	}
	for key, value := range flow.Parameters {
		if value.Default == nil {
			if _, ok := options.Args[key]; !ok {
				return fmt.Errorf("mandatory argument %s was not provided", key)
			}
		}
	}
	return nil
}

func unique(slice []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, t := range slice {
		if _, found := seen[t]; !found {
			result = append(result, t)
			seen[t] = true
		}
	}
	return result
}

// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func (sched *Scheduler) BuildDependencyGraph(options interfaces.DependencyGraphOptions) (interfaces.DependencyGraph, error) {
	if options.FlowName == "" {
		options.FlowName = interfaces.DefaultFlowName
	}

	log.Println("Getting resource definitions")
	resDefs, err := sched.getResourceDefinitions()
	if err != nil {
		return nil, err
	}

	fullFlowName := "flow/" + options.FlowName
	flowResDef, ok := resDefs[fullFlowName]
	if !ok && options.FlowName != interfaces.DefaultFlowName || ok && flowResDef.Flow == nil {
		return nil, fmt.Errorf("flow %s is not found", options.FlowName)
	}

	flow := flowResDef.Flow
	if flow == nil {
		flow = sched.newDefaultFlowObject()
	}

	if !flow.Exported && options.ExportedOnly {
		return nil, fmt.Errorf("flow %s is not exported", options.FlowName)
	}

	err = checkArgs(options, flow)
	if err != nil {
		return nil, err
	}

	log.Println("Getting dependencies")
	depList, err := sched.client.Dependencies().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}

	log.Println("Making sure there is no cycles in the dependency graph")
	if err = EnsureNoCycles(depList.Items); err != nil {
		return nil, err
	}

	dependencies := groupDependencies(depList.Items, resDefs)

	depGraph := NewDependencyGraph(sched, options)
	rootContext := &GraphContext{scheduler: sched, graph: depGraph, flow: flow, args: options.Args}

	if _, ok := dependencies[fullFlowName]; !ok {
		log.Printf("Flow %s is empty", options.FlowName)
		return depGraph, nil
	}

	type Block struct {
		dependency        client.Dependency
		scheduledResource *ScheduledResource
		parentContext     *GraphContext
	}
	blocks := map[string][]*Block{}

	queue := list.New()
	queue.PushFront(&Block{dependency: client.Dependency{Child: fullFlowName}})

	for e := queue.Front(); e != nil; e = e.Next() {
		parent := e.Value.(*Block)

		deps := filterDependencies(dependencies, parent.dependency.Child, flow)

		for _, dep := range deps {
			if parent.scheduledResource != nil && strings.HasPrefix(parent.scheduledResource.Key(), "flow/") {
				parentFlow := resDefs[dep.Parent]
				if parentFlow.Flow != nil && canDependencyBelongToFlow(&dep, parentFlow.Flow) {
					continue
				}
			}
			parentContext := rootContext
			if parent.scheduledResource != nil {
				parentContext = parent.scheduledResource.Context
			}

			kind, name, err := keyParts(dep.Child)

			context := sched.prepareContext(parentContext, dep)
			sr, err := sched.newScheduledResource(kind, name, resDefs, context)
			if err != nil {
				return nil, err
			}

			block := &Block{
				scheduledResource: sr,
				dependency:        dep,
				parentContext:     parentContext,
			}

			blocks[dep.Child] = append(blocks[dep.Child], block)

			if parent.scheduledResource != nil {
				sr.Requires = append(sr.Requires, parent.scheduledResource.Key())
				parent.scheduledResource.RequiredBy = append(parent.scheduledResource.RequiredBy, sr.Key())
				sr.Meta[parent.dependency.Child] = dep.Meta
			}
			queue.PushBack(block)
		}
	}
	for _, block := range blocks {
		for _, entry := range block {
			key := entry.scheduledResource.Key()
			existingSr := depGraph.graph[key]
			if existingSr == nil {
				log.Printf("Adding resource %s to the dependency graph flow %s", key, options.FlowName)
				depGraph.graph[key] = entry.scheduledResource
			} else {
				sched.updateContext(existingSr.Context, entry.parentContext, entry.dependency)
				existingSr.Requires = append(existingSr.Requires, entry.scheduledResource.Requires...)
				existingSr.RequiredBy = append(existingSr.RequiredBy, entry.scheduledResource.RequiredBy...)
				for metaKey, metaValue := range entry.scheduledResource.Meta {
					existingSr.Meta[metaKey] = metaValue
				}
			}
		}
	}

	for _, value := range depGraph.graph {
		value.RequiredBy = unique(value.RequiredBy)
		value.Requires = unique(value.Requires)
	}

	return depGraph, nil
}
