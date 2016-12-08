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

package scheduler

import (
	"container/list"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
	"github.com/Mirantis/k8s-AppController/report"
	"github.com/Mirantis/k8s-AppController/resources"
)

// ScheduledResourceStatus describes possible status of a single resource
type ScheduledResourceStatus int

// Possible values for ScheduledResourceStatus
const (
	Init ScheduledResourceStatus = iota
	Creating
	Ready
)

// DeploymentStatus describes possible status of whole deployment process
type DeploymentStatus int

// Possible values for DeploymentStatus
const (
	Empty DeploymentStatus = iota
	Prepared
	Running
	Finished
	TimedOut
)

func (s DeploymentStatus) String() string {
	switch s {
	case Empty:
		return "No dependencies loaded"
	case Prepared:
		return "Deployment not started"
	case Running:
		return "Deployment is running"
	case Finished:
		return "Deployment finished"
	case TimedOut:
		return "Deployment timed out"
	}
	panic("Unreachable")
}

// CheckInterval is an interval between rechecking the tree for updates
const (
	CheckInterval = time.Millisecond * 1000
)

// ScheduledResource is a wrapper for Resource with attached relationship data
type ScheduledResource struct {
	Requires   []*ScheduledResource
	RequiredBy []*ScheduledResource
	Status     ScheduledResourceStatus
	interfaces.Resource
	// parentKey -> dependencyMetadata
	Meta map[string]map[string]string
	sync.RWMutex
}

// RequestCreation does not create a scheduled resource immediately, but updates status
// and puts the scheduled resource to corresponding channel. Returns true if
// scheduled resource creation was actually requested, false otherwise.
func (sr *ScheduledResource) RequestCreation(toCreate chan *ScheduledResource) bool {
	sr.RLock()
	// somebody already requested resource creation
	if sr.Status != Init {
		sr.RUnlock()
		return true
	}

	sr.RUnlock()
	sr.Lock()
	defer sr.Unlock()

	if sr.Status == Init && !sr.IsBlocked() {
		sr.Status = Creating
		toCreate <- sr
		return true
	}

	return false
}

// IsFinished returns true if a scheduled resource processing is finished, regardless successfull or not,
// false otherwise. The actual result of processing could be obtained from returned error.
func (sr *ScheduledResource) IsFinished() (bool, error) {
	status, err := sr.Resource.Status(nil)
	if err != nil {
		return true, err
	}

	if status == "ready" {
		sr.Lock()
		sr.Status = Ready
		sr.Unlock()
		return true, nil
	}
	return false, nil
}

// IsBlocked checks whether a scheduled resource can be created. It checks status of resources
// it depends on, via API
func (sr *ScheduledResource) IsBlocked() bool {
	isBlocked := false

	for _, req := range sr.Requires {
		if isBlocked {
			break
		}

		req.RLock()
		meta := req.Meta[sr.Key()]
		if strings.HasPrefix(req.Resource.Key(), "service") {
			status, err := req.Resource.Status(meta)
			if err != nil || status != "ready" {
				isBlocked = true
			}
		} else {
			if meta == nil && req.Status != Ready {
				isBlocked = true
			} else {
				status, err := req.Resource.Status(meta)
				if err != nil || status != "ready" {
					isBlocked = true
				}
			}
		}
		req.RUnlock()
	}

	return isBlocked
}

// DependencyGraph is a full deployment graph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph map[string]*ScheduledResource

func newResource(name string, resDefs []client.ResourceDefinition, c client.Interface, resourceTemplate interfaces.ResourceTemplate) interfaces.Resource {
	for _, rd := range resDefs {
		if resourceTemplate.NameMatches(rd, name) {
			log.Println("Found resource definition for ", name)
			return resourceTemplate.New(rd, c)
		}
	}

	log.Printf("Resource definition for '%s' not found, so it is expected to exist already", name)
	return resourceTemplate.NewExisting(name, c)

}

// NewScheduledResource is a constructor for ScheduledResource
func NewScheduledResource(kind string, name string,
	resDefs []client.ResourceDefinition, c client.Interface) (*ScheduledResource, error) {

	var r interfaces.Resource

	resourceTemplate, ok := resources.KindToResourceTemplate[kind]
	if !ok {
		return nil, fmt.Errorf("Not a proper resource kind: %s. Expected '%s'", kind, strings.Join(resources.Kinds, "', '"))
	}
	r = newResource(name, resDefs, c, resourceTemplate)

	return NewScheduledResourceFor(r), nil
}

// NewScheduledResourceFor returns new scheduled resource for given resource in init state
func NewScheduledResourceFor(r interfaces.Resource) *ScheduledResource {
	return &ScheduledResource{
		Status:   Init,
		Resource: r,
		Meta:     map[string]map[string]string{},
	}
}

func keyParts(key string) (kind string, name string, err error) {
	parts := strings.Split(key, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("Not a proper resource key: %s. Expected KIND/NAME", key)
	}

	return parts[0], parts[1], nil
}

// BuildDependencyGraph loads dependencies data and creates the DependencyGraph
func BuildDependencyGraph(c client.Interface, sel labels.Selector) (DependencyGraph, error) {

	log.Println("Getting resource definitions")
	resDefList, err := c.ResourceDefinitions().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	resDefs := resDefList.Items

	log.Println("Getting dependencies")
	depList, err := c.Dependencies().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	depGraph := DependencyGraph{}

	for _, d := range depList.Items {
		parent := d.Parent
		child := d.Child

		// NOTE(ikhudoshyn): We should normalize parent's and child's key
		// to the form KIND/NAME, so that no extra metainformation
		// (e.g. success factor for replica set) will not be a part of the key

		log.Println("Found dependency", parent, "->", child)

		for _, key := range []string{parent, child} {
			if _, ok := depGraph[key]; !ok {
				log.Printf("Resource %s not found in dependecy graph yet, adding.", key)

				kind, name, err := keyParts(key)
				if err != nil {
					return nil, err
				}

				sr, err := NewScheduledResource(kind, name, resDefs, c)
				if err != nil {
					return nil, err
				}

				depGraph[key] = sr
			}
		}

		depGraph[child].Requires = append(
			depGraph[child].Requires, depGraph[parent])

		depGraph[child].Meta[parent] = d.Meta

		depGraph[parent].RequiredBy = append(
			depGraph[parent].RequiredBy, depGraph[child])
	}

	log.Println("Looking for resource definitions not in dependency list")
	for _, r := range resDefList.Items {
		var resource interfaces.Resource

		if r.Pod != nil {
			resource = resources.NewPod(r.Pod, c.Pods(), r.Meta)
		} else if r.Job != nil {
			resource = resources.NewJob(r.Job, c.Jobs(), r.Meta)
		} else if r.Service != nil {
			resource = resources.NewService(r.Service, c.Services(), c, r.Meta)
		} else if r.ReplicaSet != nil {
			resource = resources.NewReplicaSet(r.ReplicaSet, c.ReplicaSets(), r.Meta)
		} else if r.PetSet != nil {
			resource = resources.NewPetSet(r.PetSet, c.PetSets(), c, r.Meta)
		} else if r.DaemonSet != nil {
			resource = resources.NewDaemonSet(r.DaemonSet, c.DaemonSets(), r.Meta)
		} else if r.ConfigMap != nil {
			resource = resources.NewConfigMap(r.ConfigMap, c.ConfigMaps(), r.Meta)
		} else if r.Secret != nil {
			resource = resources.NewSecret(r.Secret, c.Secrets(), r.Meta)
		} else if r.Deployment != nil {
			resource = resources.NewDeployment(r.Deployment, c.Deployments(), r.Meta)
		} else {
			return nil, fmt.Errorf("Found unsupported resource %v", r)
		}

		if _, ok := depGraph[resource.Key()]; !ok {
			log.Printf("Resource %s not found in dependecy graph yet, adding.", resource.Key())
			depGraph[resource.Key()] = NewScheduledResourceFor(resource)
		}
	}

	return depGraph, nil
}

func createResources(toCreate chan *ScheduledResource, finished chan string, ccLimiter chan struct{}) {

	for r := range toCreate {
		go func(r *ScheduledResource, finished chan string, ccLimiter chan struct{}) {
			// Acquire sepmaphor
			ccLimiter <- struct{}{}
			log.Println("Creating resource", r.Key())
			err := r.Create()
			if err != nil {
				log.Printf("Error creating resource %s: %v", r.Key(), err)
			}

			// NOTE(gluke77): We start goroutines for dependencies
			// before the resource becomes ready, since dependencies
			// could have metadata defining their own readiness condition
			for _, req := range r.RequiredBy {
				go func(req *ScheduledResource, toCreate chan *ScheduledResource) {
					for {
						time.Sleep(CheckInterval)
						if req.RequestCreation(toCreate) {
							break
						}
					}
				}(req, toCreate)
			}

			log.Printf("Checking status for %s", r.Key())

			for {
				time.Sleep(CheckInterval)
				var isFinished bool
				isFinished, err = r.IsFinished()
				if isFinished {
					break
				}
			}

			if err != nil {
				log.Printf("Resource %s was not created: %v", r.Key(), err)
			} else {
				log.Printf("Resource %s created", r.Key())
			}
			finished <- r.Key()
			// Release semaphor
			<-ccLimiter
		}(r, finished, ccLimiter)
	}
}

// Create starts the deployment of a DependencyGraph
func Create(depGraph DependencyGraph, concurrency int) {

	depCount := len(depGraph)

	concurrencyLimiterLen := depCount
	if concurrency > 0 && concurrency < concurrencyLimiterLen {
		concurrencyLimiterLen = concurrency
	}

	ccLimiter := make(chan struct{}, concurrencyLimiterLen)
	toCreate := make(chan *ScheduledResource, depCount)
	created := make(chan string, depCount)

	go createResources(toCreate, created, ccLimiter)

	for _, r := range depGraph {
		if len(r.Requires) == 0 {
			r.RequestCreation(toCreate)
		}
	}

	log.Printf("Wait for %d deps to create\n", depCount)
	for i := 0; i < depCount; i++ {
		<-created
	}
	close(toCreate)
	close(created)

	// TODO Make sure every KO gets created eventually
}

// DetectCycles implements Kosaraju's algorithm https://en.wikipedia.org/wiki/Kosaraju%27s_algorithm
// for detecting cycles in graph.
// We are depending on the fact that any strongly connected component of a graph is a cycle
// if it consists of more than one vertex
func DetectCycles(depGraph DependencyGraph) [][]*ScheduledResource {
	// is vertex visited in first phase of the algorithm
	visited := make(map[string]bool)
	// is vertex assigned to strongly connected component
	assigned := make(map[string]bool)

	// each key is root of strongly connected component
	// the slice consists of all vertices belonging to strongly connected component to which root belongs
	components := make(map[string][]*ScheduledResource)

	orderedVertices := list.New()

	for key := range depGraph {
		visited[key] = false
		assigned[key] = false
	}

	for key := range depGraph {
		visitVertex(depGraph[key], visited, orderedVertices)
	}

	for e := orderedVertices.Front(); e != nil; e = e.Next() {
		vertex := e.Value.(*ScheduledResource)
		assignVertex(vertex, vertex, assigned, components)
	}

	// if any strongly connected component consist of more than one vertex - it's a cycle
	var cycles [][]*ScheduledResource
	for _, component := range components {
		if len(component) > 1 {
			cycles = append(cycles, component)
		}
	}

	// detect self cycles - not part of Kosaraju's algorithm
	for key, vertex := range depGraph {
		for _, child := range vertex.RequiredBy {
			if key == child.Key() {
				cycles = append(cycles, []*ScheduledResource{vertex, vertex})
			}
		}
	}
	return cycles
}

func visitVertex(vertex *ScheduledResource, visited map[string]bool, orderedVertices *list.List) {
	if visited[vertex.Key()] == false {
		visited[vertex.Key()] = true
		for _, v := range vertex.RequiredBy {
			visitVertex(v, visited, orderedVertices)
		}
		orderedVertices.PushFront(vertex)
	}
}

func assignVertex(vertex, root *ScheduledResource, assigned map[string]bool, components map[string][]*ScheduledResource) {
	if assigned[vertex.Key()] == false {
		var component []*ScheduledResource
		// if component is not yet initiated, make the slice
		component, ok := components[root.Key()]
		if !ok {
			component = make([]*ScheduledResource, 0, 1)
			components[root.Key()] = component
		}

		components[root.Key()] = append(component, vertex)
		assigned[vertex.Key()] = true

		for _, v := range vertex.Requires {
			assignVertex(v, root, assigned, components)
		}
	}
}

// GetNodeReport acts as a more verbose version of IsBlocked. It performs the
// same check as IsBlocked, but returns the DeploymentReport
func (sr *ScheduledResource) GetNodeReport(name string) report.NodeReport {
	var ready bool
	isBlocked := false
	dependencies := make([]interfaces.DependencyReport, 0, len(sr.Requires))
	status, err := sr.Resource.Status(nil)
	if err != nil {
		ready = false
	} else {
		ready = status == "ready"
	}
	for _, r := range sr.Requires {
		r.RLock()
		meta := r.Meta[sr.Key()]
		depReport := r.GetDependencyReport(meta)
		r.RUnlock()
		if depReport.Blocks {
			isBlocked = true
		}
		dependencies = append(dependencies, depReport)
	}
	return report.NodeReport{
		Dependent:    name,
		Dependencies: dependencies,
		Blocked:      isBlocked,
		Ready:        ready,
	}
}

// GetStatus generates data for getting the status of deployment. Returns
// a DeploymentStatus and a human readable report string
// TODO: Allow for other formats of report (e.g. json for visualisations)
func (graph *DependencyGraph) GetStatus() (DeploymentStatus, report.DeploymentReport) {
	var readyExist, nonReadyExist bool
	var status DeploymentStatus
	report := make(report.DeploymentReport, 0, len(*graph))
	for key, resource := range *graph {
		depReport := resource.GetNodeReport(key)
		report = append(report, depReport)
		if depReport.Ready {
			readyExist = true
		} else {
			nonReadyExist = true
		}
	}
	switch {
	case readyExist && nonReadyExist:
		status = Running
	case readyExist:
		status = Finished
	case nonReadyExist:
		status = Prepared
	default:
		status = Empty
	}
	return status, report
}
