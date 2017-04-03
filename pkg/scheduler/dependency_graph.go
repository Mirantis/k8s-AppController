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
	"sort"
	"strings"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/copier"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/resources"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/labels"
)

// DependencyGraph is a full deployment depGraph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph struct {
	graph        map[string]*ScheduledResource
	scheduler    *Scheduler
	graphOptions interfaces.DependencyGraphOptions
	finalizer    func()
}

type GraphContext struct {
	args       map[string]string
	graph      *DependencyGraph
	scheduler  *Scheduler
	flow       *client.Flow
	replica    string
	dependency client.Dependency
}

// Returns the Scheduler that was used to create the dependency graph
func (gc GraphContext) Scheduler() interfaces.Scheduler {
	return gc.scheduler
}

// Returns argument values available in the current graph context
func (gc GraphContext) GetArg(name string) string {
	switch name {
	case "AC_NAME":
		return gc.replica
	default:
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
}

// Returns the currently running dependency graph
func (gc GraphContext) Graph() interfaces.DependencyGraph {
	return gc.graph
}

// Returns the currently running dependency graph
func (gc GraphContext) Dependency() client.Dependency {
	return gc.dependency
}

// newScheduledResourceFor returns new scheduled resource for given resource in init state
func newScheduledResourceFor(r interfaces.Resource, context *GraphContext, existing bool) *ScheduledResource {
	return &ScheduledResource{
		Started:  false,
		Ignored:  false,
		Error:    nil,
		Resource: r,
		Meta:     map[string]map[string]string{},
		Context:  context,
		Existing: existing,
	}
}

type DependencyList client.DependencyList
var _ sort.Interface = &DependencyList{}

func (d *DependencyList) Len() int {
	return len(d.Items)
}

func (d *DependencyList) Less(i, j int) bool {
	if d.Items[i].CreationTimestamp.Equal(d.Items[j].CreationTimestamp) {
		return d.Items[i].UID < d.Items[j].UID
	}
	return d.Items[i].CreationTimestamp.Before(d.Items[j].CreationTimestamp)
}

func (d *DependencyList) Swap(i, j int) {
	d.Items[i], d.Items[j] = d.Items[j], d.Items[i]
}

func (sched *Scheduler) getDependencies() ([]client.Dependency, error) {
	depList, err := sched.client.Dependencies().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}
	tmp := DependencyList(*depList)
	sort.Stable(&tmp)

	return tmp.Items, nil

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
	flow *client.Flow, destruction bool) []client.Dependency {

	children := dependencies[parent]
	var result []client.Dependency
	for _, dep := range children {
		if canDependencyBelongToFlow(&dep, flow, destruction) {
			result = append(result, dep)
		}
	}
	return result
}

func canDependencyBelongToFlow(dep *client.Dependency, flow *client.Flow, destruction bool) bool {
	var selector map[string]string
	if destruction {
		selector = flow.Destruction
		if selector == nil || len(selector) == 0 {
			return false
		}
	} else {
		selector = flow.Construction
	}

	for k, v := range selector {
		if dep.Labels[k] != v {
			return false
		}
	}
	return true
}

// newScheduledResource is a constructor for ScheduledResource
func (sched Scheduler) newScheduledResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc *GraphContext, silent bool) (*ScheduledResource, error) {

	var r interfaces.Resource

	resourceTemplate, ok := resources.KindToResourceTemplate[kind]
	if !ok {
		return nil, fmt.Errorf("not a proper resource kind: %s. Expected '%s'",
			kind, strings.Join(resources.Kinds, "', '"))
	}
	r, existing, err := sched.newResource(kind, name, resDefs, gc, resourceTemplate, silent)
	if err != nil {
		return nil, err
	}

	return newScheduledResourceFor(r, gc, existing), nil
}

func (sched Scheduler) newResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc *GraphContext, resourceTemplate interfaces.ResourceTemplate, silent bool) (interfaces.Resource, bool, error) {
	rd, ok := resDefs[kind+"/"+name]
	if ok {
		if !silent {
			log.Printf("Found resource definition for %s/%s", kind, name)
		}
		return resourceTemplate.New(rd, sched.client, *gc), false, nil
	}

	if !silent {
		log.Printf("Resource definition for '%s/%s' not found, so it is expected to exist already", kind, name)
	}
	r := resourceTemplate.NewExisting(name, sched.client, *gc)
	if r == nil {
		return nil, true, fmt.Errorf("existing resource %s/%s cannot be reffered", kind, name)
	}
	return r, true, nil
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

func (sched *Scheduler) prepareContext(parentContext *GraphContext, dependency client.Dependency, replica string) *GraphContext {
	context := &GraphContext{
		scheduler:  sched,
		graph:      parentContext.graph,
		flow:       parentContext.flow,
		replica:    replica,
		dependency: dependency,
	}

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

type ReplicaList client.ReplicaList
var _ sort.Interface = &ReplicaList{}

func (rl *ReplicaList) Len() int {
	return len(rl.Items)
}

func (rl *ReplicaList) Less(i, j int) bool {
	if rl.Items[i].CreationTimestamp.Equal(rl.Items[j].CreationTimestamp) {
		return rl.Items[i].UID < rl.Items[j].UID
	}
	return rl.Items[i].CreationTimestamp.Before(rl.Items[j].CreationTimestamp)
}

func (rl *ReplicaList) Swap(i, j int) {
	rl.Items[i], rl.Items[j] = rl.Items[j], rl.Items[i]
}

func (sched *Scheduler) allocateReplicas(flow *client.Flow, options interfaces.DependencyGraphOptions) ([]client.Replica, []client.Replica, error) {
	name := flow.Name
	if options.FlowInstanceName != "" {
		name = options.FlowInstanceName
	}
	label := labels.Set{"flow": name}
	var tmp ReplicaList
	var maxCurrentTime time.Time

	count := options.ReplicaCount
	if options.FixedNumberOfReplicas || options.ReplicaCount <= 0 {
		existingReplicas, err := sched.client.Replicas().List(api.ListOptions{
			LabelSelector: labels.SelectorFromSet(label),
		})
		if err != nil {
			return nil, nil, err
		}
		for _, item := range existingReplicas.Items {
			if item.CreationTimestamp.After(maxCurrentTime) {
				maxCurrentTime = item.CreationTimestamp.Time
			}
		}
		tmp = ReplicaList(*existingReplicas)
		if !options.FixedNumberOfReplicas && options.ReplicaCount <= 0 {
			count = len(tmp.Items) + options.ReplicaCount
		}
	}
	if count < options.MinReplicaCount && options.MinReplicaCount > 0 {
		count = options.MinReplicaCount
	}
	if count > options.MaxReplicaCount && options.MaxReplicaCount > options.MinReplicaCount && options.MaxReplicaCount > 0 {
		count = options.MaxReplicaCount
	}

	if count < len(tmp.Items) {
		sort.Stable(&tmp)
		return nil, tmp.Items[count:], nil
	}

	for len(tmp.Items) < count {
		replica := &client.Replica{}
		replica.GenerateName = strings.ToLower(flow.Name) + "-"
		replica.FlowName = flow.Name
		replica.Labels = label
		replica.Namespace = sched.client.Namespace()
		replica, err := sched.client.Replicas().Create(replica)
		if err != nil {
			return nil, nil, err
		}
		if !replica.CreationTimestamp.After(maxCurrentTime) {
			time.Sleep(time.Second)
			sched.client.Replicas().Delete(replica.Name)
			continue
		}
		tmp.Items = append(tmp.Items, *replica)
	}

	sort.Stable(&tmp)

	return tmp.Items[:count], nil, nil
}

// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func (sched *Scheduler) BuildDependencyGraph(options interfaces.DependencyGraphOptions) (interfaces.DependencyGraph, error) {
	if options.FlowName == "" {
		options.FlowName = interfaces.DefaultFlowName
	}

	if !options.Silent {
		log.Println("Getting resource definitions")
	}
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

	if !options.Silent {
		log.Println("Getting dependencies")
	}
	depList, err := sched.getDependencies()
	if err != nil {
		return nil, err
	}

	if !options.Silent {
		log.Println("Making sure there is no cycles in the dependency graph")
	}
	if err = EnsureNoCycles(depList); err != nil {
		return nil, err
	}

	dependencies := groupDependencies(depList, resDefs)

	depGraph := NewDependencyGraph(sched, options)
	rootContext := &GraphContext{scheduler: sched, graph: depGraph, flow: flow, args: options.Args}

	replicas, deleteReplicas, err := sched.allocateReplicas(flow, options)
	if err != nil {
		return nil, err
	}

	err = sched.fillDependencyGraph(rootContext, resDefs, dependencies, flow, replicas, false)
	if err != nil {
		return nil, err
	}
	err = sched.fillDependencyGraph(rootContext, resDefs, dependencies, flow, deleteReplicas, true)
	if err != nil {
		return nil, err
	}

	for _, value := range depGraph.graph {
		value.RequiredBy = unique(value.RequiredBy)
		value.Requires = unique(value.Requires)
		value.usedInReplicas = unique(value.usedInReplicas)
	}

	if len(deleteReplicas) > 0 {
		optionsCopy := options
		optionsCopy.ReplicaCount = 0
		optionsCopy.MinReplicaCount = 0
		optionsCopy.FixedNumberOfReplicas = false
		optionsCopy.Silent = true

		allReplicas, _, err := sched.allocateReplicas(flow, optionsCopy)
		allReplicasGraph := NewDependencyGraph(sched, optionsCopy)
		context := &GraphContext{scheduler: sched, graph: allReplicasGraph, flow: flow, args: optionsCopy.Args}
		err = sched.fillDependencyGraph(context, resDefs, dependencies, flow, allReplicas, false)
		if err != nil {
			return nil, err
		}
		depGraph.finalizer = sched.composeFinalizer(allReplicasGraph, depGraph, deleteReplicas)
	}

	return depGraph, nil
}

func (sched *Scheduler) fillDependencyGraph(rootContext *GraphContext,
	resDefs map[string]client.ResourceDefinition,
	dependencies map[string][]client.Dependency,
	flow *client.Flow, replicas []client.Replica, destruction bool) error {

	type Block struct {
		dependency        client.Dependency
		scheduledResource *ScheduledResource
		parentContext     *GraphContext
	}
	blocks := map[string][]*Block{}
	silent := rootContext.graph.graphOptions.Silent

	for _, replica := range replicas {
		replicaName := replica.ReplicaName()
		queue := list.New()
		queue.PushFront(&Block{dependency: client.Dependency{Child: "flow/" + flow.Name}})

		for e := queue.Front(); e != nil; e = e.Next() {
			parent := e.Value.(*Block)

			deps := filterDependencies(dependencies, parent.dependency.Child, flow, destruction)

			for _, dep := range deps {
				if parent.scheduledResource != nil && strings.HasPrefix(parent.scheduledResource.Key(), "flow/") {
					parentFlow := resDefs[dep.Parent]
					if parentFlow.Flow != nil && (canDependencyBelongToFlow(&dep, parentFlow.Flow, true) ||
						canDependencyBelongToFlow(&dep, parentFlow.Flow, false)) {
						continue
					}
				}
				parentContext := rootContext
				if parent.scheduledResource != nil {
					parentContext = parent.scheduledResource.Context
				}

				kind, name, err := keyParts(dep.Child)

				context := sched.prepareContext(parentContext, dep, replicaName)
				sr, err := sched.newScheduledResource(kind, name, resDefs, context, silent)
				if err != nil {
					return err
				}
				sr.usedInReplicas = []string{replicaName}

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
				existingSr := rootContext.graph.graph[key]
				if existingSr == nil {
					if !silent {
						log.Printf("Adding resource %s to the dependency graph flow %s", key, flow.Name)
					}
					rootContext.graph.graph[key] = entry.scheduledResource
				} else {
					sched.updateContext(existingSr.Context, entry.parentContext, entry.dependency)
					existingSr.Requires = append(existingSr.Requires, entry.scheduledResource.Requires...)
					existingSr.RequiredBy = append(existingSr.RequiredBy, entry.scheduledResource.RequiredBy...)
					existingSr.usedInReplicas = append(existingSr.usedInReplicas, entry.scheduledResource.usedInReplicas...)
					for metaKey, metaValue := range entry.scheduledResource.Meta {
						existingSr.Meta[metaKey] = metaValue
					}
				}
			}
		}
	}
	return nil
}

func (sched *Scheduler) composeFinalizer(construction, destruction *DependencyGraph, replicas []client.Replica) func() {
	replicaMap := map[string]client.Replica{}
	for _, replica := range replicas {
		replicaMap[replica.ReplicaName()] = replica
	}

	var destructors []func() bool
	var failed chan *ScheduledResource

	var resourceCount int
	for _, depGraph := range [2]*DependencyGraph{construction, destruction} {
		for _, resource := range depGraph.graph {
			release := true
			for _, replicaName := range resource.usedInReplicas {
				if _, found := replicaMap[replicaName]; !found {
					release = false
					break
				}
			}
			if release {
				resourceCopy := resource
				f := func() bool {
					if !resourceCopy.Existing || resourceCopy.Context.graph.graphOptions.AllowDeleteExternalResources {
						log.Printf("%s flow: Deleting resource %s", resourceCopy.Context.flow.Name, resourceCopy.Key())
						err := resourceCopy.Delete()
						if err != nil {
							log.Println(err)
							statusError, ok := err.(*errors.StatusError)
							if !ok || statusError.Status().Reason != unversioned.StatusReasonNotFound {
								failed <- resourceCopy
							}
						}
						return err == nil
					} else {
						log.Printf("%s: Won't delete external resource %s", resourceCopy.Context.flow.Name, resourceCopy.Key())
						return true
					}
				}
				destructors = append(destructors, f)
				resourceCount++
			}
		}
	}

	return func() {
		failed = make(chan *ScheduledResource, resourceCount)
		defer close(failed)
		runConcurrently(destructors, sched.concurrency)
		failedReplicas := map[string]bool{}

	readFailed:
		for {
			select {
			case res := <-failed:
				for _, replicaName := range res.usedInReplicas {
					failedReplicas[replicaName] = true
				}
			default:
				break readFailed
			}
		}
		var deleteReplicaFuncs []func() bool

		for replicaName, replicaObject := range replicaMap {
			if _, found := failedReplicas[replicaName]; found {
				continue
			}
			replicaNameCopy := replicaName
			replicaObjectCopy := replicaObject
			deleteReplicaFuncs = append(deleteReplicaFuncs, func() bool {
				log.Printf("%s flow: Deleting replica %s", replicaObjectCopy.FlowName, replicaNameCopy)
				err := sched.client.Replicas().Delete(replicaObjectCopy.Name)
				if err != nil {
					log.Println(err)
				}
				return true
			})

		}

		if deleteReplicaFuncs != nil {
			runConcurrently(deleteReplicaFuncs, sched.concurrency)
		}
	}
}
