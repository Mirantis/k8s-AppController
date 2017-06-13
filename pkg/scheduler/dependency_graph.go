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
	"strconv"
	"strings"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/copier"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/resources"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/labels"
)

type dependencyGraph struct {
	graph        map[string]*ScheduledResource
	scheduler    *scheduler
	graphOptions interfaces.DependencyGraphOptions
	finalizer    func()
}

type graphContext struct {
	args      map[string]string
	graph     *dependencyGraph
	scheduler *scheduler
	flow      *client.Flow
	id        string
	replica   string
}

var _ interfaces.GraphContext = &graphContext{}

// Scheduler returns the Scheduler that was used to create the dependency graph
func (gc graphContext) Scheduler() interfaces.Scheduler {
	return gc.scheduler
}

// GetArg returns argument values available in the current graph context
func (gc graphContext) GetArg(name string) string {
	switch name {
	case "AC_NAME":
		return gc.replica
	case "AC_FLOW_NAME":
		return gc.flow.Name
	case "AC_ID":
		return gc.id
	default:
		val, ok := gc.args[name]
		if ok {
			return val
		}
		val, ok = gc.graph.Options().Args[name]
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

// Graph method returns the currently running dependency graph
func (gc graphContext) Graph() interfaces.DependencyGraph {
	return gc.graph
}

// newScheduledResourceFor returns new scheduled resource for given resource in init state
func newScheduledResourceFor(r interfaces.Resource, suffix string, context *graphContext, existing bool) *ScheduledResource {
	return &ScheduledResource{
		Started:  false,
		Ignored:  false,
		Error:    nil,
		Resource: r,
		Meta:     map[string]map[string]string{},
		context:  context,
		Existing: existing,
		suffix:   copier.EvaluateString(suffix, getArgFunc(context)),
	}
}

type sortableDependencyList client.DependencyList

var _ sort.Interface = &sortableDependencyList{}

// Len is the number of dependencies in the collection.
func (d *sortableDependencyList) Len() int {
	return len(d.Items)
}

// Less reports whether the dependency with index i should sort before the element with index j.
func (d *sortableDependencyList) Less(i, j int) bool {
	if d.Items[i].CreationTimestamp.Equal(d.Items[j].CreationTimestamp) {
		return d.Items[i].UID < d.Items[j].UID
	}
	return d.Items[i].CreationTimestamp.Before(d.Items[j].CreationTimestamp)
}

// Swap swaps the elements with indexes i and j.
func (d *sortableDependencyList) Swap(i, j int) {
	d.Items[i], d.Items[j] = d.Items[j], d.Items[i]
}

func (sched *scheduler) getDependencies() ([]client.Dependency, error) {
	depList, err := sched.client.Dependencies().List(api.ListOptions{LabelSelector: sched.selector})
	if err != nil {
		return nil, err
	}
	sortableDepList := sortableDependencyList(*depList)
	sort.Stable(&sortableDepList)

	return sortableDepList.Items, nil

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
				dep := client.Dependency{Parent: defaultFlowName, Child: name}
				dep.Name = name
				defaultFlow = append(defaultFlow, dep)
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

func (sched *scheduler) getResourceDefinitions() (map[string]client.ResourceDefinition, error) {
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
	flow *client.Flow, useDestructionSelector bool) []client.Dependency {

	children := dependencies[parent]
	var result []client.Dependency
	for _, dep := range children {
		if canDependencyBelongToFlow(&dep, flow, useDestructionSelector) {
			result = append(result, dep)
		}
	}
	return result
}

// canDependencyBelongToFlow returns true if the given dependency connects nodes that belong either to construction
// or destruction paths of the flow (depending on useDestructionSelector parameter)
// if construction path is empty then any dependency is going to match it since empty map is contained in any label map
// however, if destruction path is empty or nil, it means that it's not set so no dependency is going to match it.
func canDependencyBelongToFlow(dep *client.Dependency, flow *client.Flow, useDestructionSelector bool) bool {
	var selector map[string]string
	if useDestructionSelector {
		selector = flow.Destruction
		if len(selector) == 0 {
			return false
		}
	} else {
		// if destruction selector is not empty and and is a superset of construction selector
		// (i.e. all the dependencies that match destruction selector will also match construction one but not vice versa)
		// such dependencies are considered to belong to destruction path even though they match the construction selector
		if len(flow.Destruction) > len(flow.Construction) && isMapContainedIn(flow.Construction, flow.Destruction) &&
			canDependencyBelongToFlow(dep, flow, true) {
			return false
		}
		selector = flow.Construction
	}

	for k, v := range selector {
		if dep.Labels[k] != v {
			return false
		}
	}
	return true
}

func isMapContainedIn(contained, containing map[string]string) bool {
	for k := range contained {
		if _, found := containing[k]; !found {
			return false
		}
	}
	return true
}

// newScheduledResource is a constructor for ScheduledResource
func (sched scheduler) newScheduledResource(kind, name, suffix string, resDefs map[string]client.ResourceDefinition,
	gc *graphContext, silent bool) (*ScheduledResource, error) {
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

	return newScheduledResourceFor(r, suffix, gc, existing), nil
}

// newResource returns creates a resource controller for a given resources name and factory.
// It returns the created controller (implementation of interfaces.Resource), flag saying if the created controller
// will create new resource or check the status of existing resource outside (i.e. that doesn't have resource defintition)
// and error (or nil, if no error happened)
func (sched scheduler) newResource(kind, name string, resDefs map[string]client.ResourceDefinition,
	gc interfaces.GraphContext, resourceTemplate interfaces.ResourceTemplate, silent bool) (interfaces.Resource, bool, error) {
	rd, ok := resDefs[kind+"/"+name]
	if ok {
		if !silent {
			log.Printf("Found resource definition for %s/%s", kind, name)
		}
		return resourceTemplate.New(rd, sched.client, gc), false, nil
	}

	if !silent {
		log.Printf("Resource definition for '%s/%s' not found, so it is expected to exist already", kind, name)
	}
	name = copier.EvaluateString(name, getArgFunc(gc))
	r := resourceTemplate.NewExisting(name, sched.client, gc)
	if r == nil {
		return nil, true, fmt.Errorf("existing resource %s/%s cannot be reffered", kind, name)
	}
	return r, true, nil
}

func keyParts(key string) (kind, name, suffix string, err error) {
	parts := strings.SplitN(key, "/", 31)

	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("not a proper resource key: %s. Expected KIND/NAME", key)
	}
	if len(parts) == 3 {
		suffix = parts[2]
	}

	return parts[0], parts[1], suffix, nil
}

func newDependencyGraph(sched *scheduler, options interfaces.DependencyGraphOptions) *dependencyGraph {
	return &dependencyGraph{
		graph:        make(map[string]*ScheduledResource),
		scheduler:    sched,
		graphOptions: options,
	}
}

func getArgFunc(gc interfaces.GraphContext) func(string) string {
	return func(p string) string {
		value := gc.GetArg(p)
		if value == "" {
			return "$" + p
		}
		return value
	}
}

func (sched *scheduler) prepareContext(parentContext *graphContext, dependency *client.Dependency, replica string) *graphContext {
	context := &graphContext{
		scheduler: sched,
		graph:     parentContext.graph,
		flow:      parentContext.flow,
		replica:   replica,
		id:        getVertexID(dependency, replica),
	}

	context.args = make(map[string]string)
	if dependency != nil {
		for key, value := range dependency.Args {
			context.args[key] = copier.EvaluateString(value, getArgFunc(parentContext))
		}
	}
	return context
}

func getVertexID(dependency *client.Dependency, replica string) string {
	var depName string
	if dependency != nil {
		depName = strings.Replace(dependency.Name, dependency.GenerateName, "", 1)
	}
	depName += replica
	return depName
}

func (sched *scheduler) updateContext(context, parentContext *graphContext, dependency client.Dependency) {
	for key, value := range dependency.Args {
		context.args[key] = copier.EvaluateString(value, parentContext.GetArg)
	}
}

func newDefaultFlowObject() *client.Flow {
	return &client.Flow{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Flow",
			APIVersion: client.GroupName + "/" + client.Version,
		},
		ObjectMeta: api.ObjectMeta{
			Name: interfaces.DefaultFlowName,
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

type sortableReplicaList []client.Replica

var _ sort.Interface = &sortableReplicaList{}

// Len is the number of replicas in the collection.
func (rl sortableReplicaList) Len() int {
	return len(rl)
}

// Less reports whether the replica with index i should sort before the element with index j.
func (rl sortableReplicaList) Less(i, j int) bool {
	if rl[i].CreationTimestamp.Equal(rl[j].CreationTimestamp) {
		return rl[i].UID < rl[j].UID
	}
	return rl[i].CreationTimestamp.Before(rl[j].CreationTimestamp)
}

// Swap swaps the elements with indexes i and j.
func (rl sortableReplicaList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

func getReplicaSpace(flow *client.Flow, gc interfaces.GraphContext) string {
	name := flow.ReplicaSpace
	if name == "" {
		name = flow.Name
	}
	return copier.EvaluateString(name, getArgFunc(gc))
}

func (sched *scheduler) getAllReplicas(flow *client.Flow, gc interfaces.GraphContext) ([]client.Replica, error) {
	replicaSpace := getReplicaSpace(flow, gc)
	label := labels.Set{"replicaspace": replicaSpace}
	options := gc.Graph().Options()
	if options.FlowInstanceName != "" {
		label["context"] = options.FlowInstanceName
	}
	existingReplicas, err := sched.client.Replicas().List(api.ListOptions{
		LabelSelector: labels.SelectorFromSet(label),
	})
	if err != nil {
		return nil, err
	}

	sortableReplicas := sortableReplicaList(existingReplicas.Items)
	sort.Stable(sortableReplicas)
	return sortableReplicas, nil
}

// allocateReplicas allocates Replica objects for either creation or deletion.
// it returns list of replicas to be constructed, list of replicas to be destructed and an error.
// new Replica objects are created in this function as a side effect, but not deleted since this can happen only after
// graph replica destruction
func (sched *scheduler) allocateReplicas(flow *client.Flow, gc interfaces.GraphContext) ([]client.Replica, []client.Replica, error) {
	options := gc.Graph().Options()
	replicaSpace := getReplicaSpace(flow, gc)
	label := labels.Set{"replicaspace": replicaSpace}
	if options.FlowInstanceName != "" {
		label["context"] = options.FlowInstanceName
	}
	existingReplicas, err := sched.client.Replicas().List(api.ListOptions{
		LabelSelector: labels.SelectorFromSet(label),
	})
	if err != nil {
		return nil, nil, err
	}
	initialCount := 0
	for _, replica := range existingReplicas.Items {
		if replica.Deployed {
			initialCount++
		}
	}

	targetCount := options.ReplicaCount // absolute number of replicas that we want to have
	if !options.FixedNumberOfReplicas {
		targetCount += initialCount
	}
	if targetCount < options.MinReplicaCount {
		targetCount = options.MinReplicaCount
	}
	// MaxReplicaCount set to 0 means there is no limitation on replica count
	if targetCount > options.MaxReplicaCount && options.MaxReplicaCount > options.MinReplicaCount && options.MaxReplicaCount > 0 {
		targetCount = options.MaxReplicaCount
	}

	adjustedReplicas, err := sched.createReplicas(existingReplicas.Items, targetCount, flow.Name, replicaSpace, label)
	if err != nil {
		return nil, nil, err
	}

	if targetCount < initialCount {
		return nil, adjustedReplicas[targetCount:], nil
	}
	if !options.FixedNumberOfReplicas && targetCount > initialCount {
		return adjustedReplicas[initialCount:targetCount], nil, nil
	}
	return adjustedReplicas[:targetCount], nil, nil
}

// createReplicas creates missing flow replicas up to desiredCount
func (sched *scheduler) createReplicas(
	existingReplicas []client.Replica, desiredCount int, flowName, replicaSpace string, label labels.Set) ([]client.Replica, error) {

	var maxCurrentTime time.Time
	for _, item := range existingReplicas {
		if item.CreationTimestamp.After(maxCurrentTime) {
			maxCurrentTime = item.CreationTimestamp.Time
		}
	}
	replicaList := make(sortableReplicaList, 0, len(existingReplicas))
	unusedReplicas := make(sortableReplicaList, 0, len(existingReplicas))

	for _, replica := range existingReplicas {
		if replica.Deployed {
			replicaList = append(replicaList, replica)
		} else {
			unusedReplicas = append(unusedReplicas, replica)
		}
	}
	sort.Stable(&unusedReplicas)

	for len(replicaList) < desiredCount {
		var replica *client.Replica
		if len(unusedReplicas) > 0 {
			replica = &unusedReplicas[0]
			unusedReplicas = unusedReplicas[1:]
		} else {
			replica = &client.Replica{
				ObjectMeta: api.ObjectMeta{
					GenerateName: "replica-",
					Labels:       label,
					Namespace:    sched.client.Namespace(),
				},
				FlowName:     flowName,
				ReplicaSpace: replicaSpace,
			}
			var err error
			replica, err = sched.client.Replicas().Create(replica)
			if err != nil {
				return nil, err
			}
			if !replica.CreationTimestamp.After(maxCurrentTime) {
				// ensure that new elements in the list have timestamp that exceeds all the timestamps of existing items
				// this guarantees that after the sort all new elements will still go after old ones in the list
				time.Sleep(time.Second)
				sched.client.Replicas().Delete(replica.Name)
				continue
			}
		}
		replicaList = append(replicaList, *replica)
	}
	for _, replica := range unusedReplicas {
		sched.client.Replicas().Delete(replica.Name)
	}

	sort.Stable(&replicaList)
	return replicaList, nil
}

// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func (sched *scheduler) BuildDependencyGraph(options interfaces.DependencyGraphOptions) (interfaces.DependencyGraph, error) {
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
		flow = newDefaultFlowObject()
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
	if err = EnsureNoCycles(depList, resDefs); err != nil {
		return nil, err
	}

	dependencies := groupDependencies(depList, resDefs)

	depGraph := newDependencyGraph(sched, options)
	rootContext := &graphContext{scheduler: sched, graph: depGraph, flow: flow, args: options.Args}

	var replicas, deleteReplicas []client.Replica
	if options.ReplicaCount < 0 && options.FixedNumberOfReplicas {
		replicas, err = sched.getAllReplicas(flow, rootContext)
	} else {
		replicas, deleteReplicas, err = sched.allocateReplicas(flow, rootContext)
	}
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
		silentOptions := options
		silentOptions.Silent = true

		// since we are using dry-run options, allocateReplicas will only return existing replicas
		allReplicasGraph := newDependencyGraph(sched, silentOptions)
		context := &graphContext{scheduler: sched, graph: allReplicasGraph, flow: flow, args: silentOptions.Args}
		allReplicas, err := sched.getAllReplicas(flow, context)
		if err != nil {
			return nil, err
		}

		// create dependency graph that has all the replicas (both those that we are about to delete and those that
		// remain to see what resources belong exclusively to deleted replicas and what resources are shared with
		// replicas alive and thus must non be deleted
		err = sched.fillDependencyGraph(context, resDefs, dependencies, flow, allReplicas, false)
		if err != nil {
			return nil, err
		}

		// compose finalizer method that will delete all the resources belonging to deleted replicas and those
		// that were created specially for destruction (for example, jobs with cleanup scripts)
		depGraph.finalizer = sched.composeDeletingFinalizer(allReplicasGraph, depGraph, deleteReplicas)
	} else {
		depGraph.finalizer = sched.composeAcknowledgingFinalizer(replicas)
	}

	return depGraph, nil
}

func listDependencies(dependencies map[string][]client.Dependency, parent string, flow *client.Flow,
	useDestructionSelector bool, context *graphContext) []client.Dependency {

	deps := filterDependencies(dependencies, parent, flow, useDestructionSelector)
	var result []client.Dependency
	for _, dep := range deps {
		if len(dep.GenerateFor) == 0 {
			result = append(result, dep)
			continue
		}

		var keys []string
		for k := range dep.GenerateFor {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lists := make([][]string, len(dep.GenerateFor))
		for i, key := range keys {
			lists[i] = expandListExpression(copier.EvaluateString(dep.GenerateFor[key], getArgFunc(context)))
		}
		for n, combination := range permute(lists) {
			newArgs := make(map[string]string, len(dep.Args)+len(keys))
			for k, v := range dep.Args {
				newArgs[k] = v
			}
			for i, key := range keys {
				newArgs[key] = combination[i]
			}
			depCopy := dep
			depCopy.Args = newArgs
			depCopy.Name += strconv.Itoa(n + 1)
			result = append(result, depCopy)
		}
	}
	return result
}

func permute(variants [][]string) [][]string {
	switch len(variants) {
	case 0:
		return variants
	case 1:
		var result [][]string
		for _, v := range variants[0] {
			result = append(result, []string{v})
		}
		return result
	default:
		var result [][]string
		for _, tail := range variants[len(variants)-1] {
			for _, p := range permute(variants[:len(variants)-1]) {
				result = append(result, append(p, tail))
			}
		}
		return result
	}
}

func expandListExpression(expr string) []string {
	var result []string
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		isRange := true
		var from, to int

		rangeParts := strings.SplitN(part, "..", 2)
		if len(rangeParts) != 2 {
			isRange = false
		}

		var err error
		if isRange {
			from, err = strconv.Atoi(rangeParts[0])
			if err != nil {
				isRange = false
			}
		}
		if isRange {
			to, err = strconv.Atoi(rangeParts[1])
			if err != nil {
				isRange = false
			}
		}

		if isRange {
			for i := from; i <= to; i++ {
				result = append(result, strconv.Itoa(i))
			}
		} else {
			result = append(result, part)
		}
	}
	return result
}

type interimGraphVertex struct {
	dependency        client.Dependency
	scheduledResource *ScheduledResource
	parentContext     *graphContext
}

func (sched *scheduler) fillDependencyGraph(rootContext *graphContext,
	resDefs map[string]client.ResourceDefinition,
	dependencies map[string][]client.Dependency,
	flow *client.Flow, replicas []client.Replica, useDestructionSelector bool) error {

	var vertices [][]interimGraphVertex
	silent := rootContext.graph.Options().Silent

	for _, replica := range replicas {
		var replicaVertices []interimGraphVertex
		replicaName := replica.ReplicaName()
		replicaContext := sched.prepareContext(rootContext, nil, replicaName)
		queue := list.New()
		queue.PushFront(interimGraphVertex{dependency: client.Dependency{Child: "flow/" + flow.Name}})

		for e := queue.Front(); e != nil; e = e.Next() {
			parent := e.Value.(interimGraphVertex)

			deps := listDependencies(dependencies, parent.dependency.Child, flow, useDestructionSelector, replicaContext)

			for _, dep := range deps {
				if parent.scheduledResource != nil && strings.HasPrefix(parent.scheduledResource.Key(), "flow/") {
					parentFlow := resDefs[dep.Parent]
					if parentFlow.Flow != nil && (canDependencyBelongToFlow(&dep, parentFlow.Flow, true) ||
						canDependencyBelongToFlow(&dep, parentFlow.Flow, false)) {
						continue
					}
				}
				parentContext := replicaContext
				if parent.scheduledResource != nil {
					parentContext = parent.scheduledResource.context
				}

				kind, name, suffix, err := keyParts(dep.Child)

				context := sched.prepareContext(parentContext, &dep, replicaName)
				sr, err := sched.newScheduledResource(kind, name, suffix, resDefs, context, silent)
				if err != nil {
					return err
				}
				sr.usedInReplicas = []string{replicaName}

				vertex := interimGraphVertex{
					scheduledResource: sr,
					dependency:        dep,
					parentContext:     parentContext,
				}
				replicaVertices = append(replicaVertices, vertex)

				if parent.scheduledResource != nil {
					sr.Requires = append(sr.Requires, parent.scheduledResource.Key())
					parent.scheduledResource.RequiredBy = append(parent.scheduledResource.RequiredBy, sr.Key())
					sr.Meta[parent.dependency.Child] = dep.Meta
				}
				queue.PushBack(vertex)
			}
		}
		vertices = append(vertices, replicaVertices)
	}

	if flow.Sequential {
		sched.concatenateReplicas(vertices, rootContext, rootContext.graph.Options())
	} else {
		sched.mergeReplicas(vertices, rootContext, rootContext.graph.Options())
	}
	return nil
}

func (sched *scheduler) mergeReplicas(vertices [][]interimGraphVertex, gc *graphContext,
	options interfaces.DependencyGraphOptions) {

	for _, replicaVertices := range vertices {
		sched.mergeInterimGraphVertices(replicaVertices, gc.graph.graph, options)
	}
}

func (sched *scheduler) concatenateReplicas(vertices [][]interimGraphVertex, gc *graphContext,
	options interfaces.DependencyGraphOptions) {
	graph := gc.graph.graph
	var previousReplicaGraph map[string]*ScheduledResource
	for i, replicaVertices := range vertices {
		replicaGraph := map[string]*ScheduledResource{}
		sched.mergeInterimGraphVertices(replicaVertices, replicaGraph, options)

		if i > 0 {
			correctDuplicateResources(graph, replicaGraph, i)

			for _, leafName := range getLeafs(previousReplicaGraph) {
				for _, rootName := range getRoots(replicaGraph) {
					root := replicaGraph[rootName]
					leaf := previousReplicaGraph[leafName]
					root.Requires = append(root.Requires, leafName)
					leaf.RequiredBy = append(leaf.RequiredBy, rootName)
				}
			}
		}
		previousReplicaGraph = replicaGraph
		for key, value := range replicaGraph {
			graph[key] = value
		}
	}
}

func correctDuplicateResources(existingGraph, newGraph map[string]*ScheduledResource, index int) {
	toReplace := map[string]*ScheduledResource{}
	for key, sr := range newGraph {
		if existingGraph[key] != nil {
			toReplace[key] = sr
		}
	}
	for key, sr := range toReplace {
		sr.context.id = existingGraph[key].context.id
		j := index + 1
		suffix := sr.suffix
		for {
			sr.suffix = fmt.Sprintf("%s #%d", suffix, j)
			if existingGraph[sr.Key()] == nil {
				break
			}
			j++
		}
		for _, rKey := range sr.RequiredBy {
			requires := newGraph[rKey].Requires
			for i, rKey2 := range requires {
				if rKey2 == key {
					requires[i] = sr.Key()
					break
				}
			}
		}
		for _, rKey := range sr.Requires {
			requiredBy := newGraph[rKey].RequiredBy
			for i, rKey2 := range requiredBy {
				if rKey2 == key {
					requiredBy[i] = sr.Key()
					break
				}
			}
		}
		delete(newGraph, key)
		newGraph[sr.Key()] = sr
	}
}

func getRoots(graph map[string]*ScheduledResource) []string {
	var result []string
	for key, sr := range graph {
		if len(sr.Requires) == 0 {
			result = append(result, key)
		}
	}
	return result
}

func getLeafs(graph map[string]*ScheduledResource) []string {
	var result []string
	for key, sr := range graph {
		if len(sr.RequiredBy) == 0 {
			result = append(result, key)
		}
	}
	return result
}

func (sched *scheduler) mergeInterimGraphVertices(vertices []interimGraphVertex, graph map[string]*ScheduledResource,
	options interfaces.DependencyGraphOptions) {

	for _, entry := range vertices {
		key := entry.scheduledResource.Key()
		existingSr := graph[key]
		if existingSr == nil {
			if !options.Silent {
				log.Printf("Adding resource %s to the dependency graph flow %s", key, options.FlowName)
			}
			graph[key] = entry.scheduledResource
		} else {
			sched.updateContext(existingSr.context, entry.parentContext, entry.dependency)
			existingSr.Requires = append(existingSr.Requires, entry.scheduledResource.Requires...)
			existingSr.RequiredBy = append(existingSr.RequiredBy, entry.scheduledResource.RequiredBy...)
			existingSr.usedInReplicas = append(existingSr.usedInReplicas, entry.scheduledResource.usedInReplicas...)
			for metaKey, metaValue := range entry.scheduledResource.Meta {
				existingSr.Meta[metaKey] = metaValue
			}
		}
	}
}

// getResourceDestructors builds a list of functions, each of them delete one of replica resources
func getResourceDestructors(construction, destruction *dependencyGraph, replicaMap map[string]client.Replica, failed *chan *ScheduledResource) []func() bool {
	var destructors []func() bool

	for _, depGraph := range [2]*dependencyGraph{construction, destruction} {
		for _, resource := range depGraph.graph {
			resourceCanBeDeleted := true
			for _, replicaName := range resource.usedInReplicas {
				if _, found := replicaMap[replicaName]; !found {
					resourceCanBeDeleted = false
					break
				}
			}
			if resourceCanBeDeleted {
				destructors = append(destructors, getDestructorFunc(resource, failed))
			}
		}
	}
	return destructors
}

func getDestructorFunc(resource *ScheduledResource, failed *chan *ScheduledResource) func() bool {
	return func() bool {
		res := deleteResource(resource)
		if res != nil {
			*failed <- resource
			return false
		}
		return true
	}
}

// deleteReplicaResources invokes resources destructors and deletes replicas for which 100% of resources were deleted
func deleteReplicaResources(sched *scheduler, destructors []func() bool, replicaMap map[string]client.Replica, failed *chan *ScheduledResource) {
	*failed = make(chan *ScheduledResource, len(destructors))
	defer close(*failed)
	deleted := runConcurrently(destructors, sched.concurrency)
	failedReplicas := map[string]bool{}
	if !deleted {
		log.Println("Some of resources were not deleted")
	}

readFailed:
	for {
		select {
		case res := <-*failed:
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

	if deleteReplicaFuncs != nil && !runConcurrently(deleteReplicaFuncs, sched.concurrency) {
		log.Println("Some of flow replicas were not deleted")
	}
}

func (sched *scheduler) composeDeletingFinalizer(construction, destruction *dependencyGraph, replicas []client.Replica) func() {
	replicaMap := map[string]client.Replica{}
	for _, replica := range replicas {
		replicaMap[replica.ReplicaName()] = replica
	}

	var failed chan *ScheduledResource
	destructors := getResourceDestructors(construction, destruction, replicaMap, &failed)

	return func() {
		log.Print("Performing resource cleanup")
		deleteReplicaResources(sched, destructors, replicaMap, &failed)
	}
}

func makeAcknowledgeReplicaFunc(replica client.Replica, api client.ReplicasInterface) func() bool {
	return func() bool {
		replica.Deployed = true
		log.Printf("%s flow: Marking replica %s as deployed", replica.FlowName, replica.ReplicaName())
		if err := api.Update(&replica); err != nil {
			log.Printf("failed to update replica %s: %v", replica.Name, err)
			return false
		}
		return true
	}
}

func (sched *scheduler) composeAcknowledgingFinalizer(replicas []client.Replica) func() {
	var funcs []func() bool
	for _, replica := range replicas {
		if !replica.Deployed {
			funcs = append(funcs, makeAcknowledgeReplicaFunc(replica, sched.client.Replicas()))
		}
	}

	return func() {
		if !runConcurrently(funcs, sched.concurrency) {
			log.Println("Some of the replicas were not updated!")
		}
	}
}
