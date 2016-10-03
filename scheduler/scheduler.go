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

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/resources"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
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
	resources.Resource
	// parentKey -> dependencyMetadata
	Meta map[string]map[string]string
	sync.RWMutex
}

// IsBlocked checks whether d can be created. It checks status of resources
// it depends on, via API
func (d *ScheduledResource) IsBlocked() bool {
	isBlocked := false

	for _, r := range d.Requires {
		if isBlocked {
			break
		}

		r.RLock()
		meta := r.Meta[d.Key()]
		switch r.Resource.(type) {
		case resources.Service:
			status, err := r.Resource.Status(meta)
			if err != nil || status != "ready" {
				isBlocked = true
			}
		default:
			if meta == nil && r.Status != Ready {
				isBlocked = true
			} else {
				status, err := r.Resource.Status(meta)
				if err != nil || status != "ready" {
					isBlocked = true
				}
			}
		}
		r.RUnlock()
	}

	return isBlocked
}

// DependencyGraph is a full deployment graph as a mapping from job keys to
// ScheduledResource pointers
type DependencyGraph map[string]*ScheduledResource

func newResourceForPod(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.Pod != nil && rd.Pod.Name == name {
			log.Println("Found resource definition for pod", name)
			return resources.NewPod(rd.Pod, c.Pods())
		}
	}

	log.Printf("Resource definition for pod '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingPod(name, c.Pods())
}

func newResourceForJob(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.Job != nil && rd.Job.Name == name {
			log.Println("Found resource definition for job", name)
			return resources.NewJob(rd.Job, c.Jobs())
		}
	}

	log.Printf("Resource definition for job '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingJob(name, c.Jobs())
}

func newResourceForService(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.Service != nil && rd.Service.Name == name {
			log.Println("Found resource definition for service", name)
			return resources.NewService(rd.Service, c.Services(), c)
		}
	}

	log.Printf("Resource definition for service '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingService(name, c.Services())
}

func newResourceForReplicaSet(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.ReplicaSet != nil && rd.ReplicaSet.Name == name {
			log.Println("Found resource definition for replica set", name)
			return resources.NewReplicaSet(rd.ReplicaSet, c.ReplicaSets())
		}
	}

	log.Printf("Resource definition for replica set '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingReplicaSet(name, c.ReplicaSets())
}

func newResourceForPetSet(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.PetSet != nil && rd.PetSet.Name == name {
			log.Println("Found resource definition for pet set", name)
			return resources.NewPetSet(rd.PetSet, c.PetSets(), c)
		}
	}

	log.Printf("Resource definition for pet set '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingPetSet(name, c.PetSets(), c)
}

func newResourceForDaemonSet(name string, resDefs []client.ResourceDefinition, c client.Interface) resources.Resource {
	for _, rd := range resDefs {
		if rd.DaemonSet != nil && rd.DaemonSet.Name == name {
			log.Println("Found resource definition for daemon set", name)
			return resources.NewDaemonSet(rd.DaemonSet, c.DaemonSets())
		}
	}

	log.Printf("Resource definition for daemon set '%s' not found, so it is expected to exist already", name)
	return resources.NewExistingDaemonSet(name, c.DaemonSets())
}

// NewScheduledResource is a constructor for ScheduledResource
func NewScheduledResource(kind string, name string,
	resDefs []client.ResourceDefinition, c client.Interface) (*ScheduledResource, error) {

	var r resources.Resource

	if kind == "pod" {
		r = newResourceForPod(name, resDefs, c)
	} else if kind == "job" {
		r = newResourceForJob(name, resDefs, c)
	} else if kind == "service" {
		r = newResourceForService(name, resDefs, c)
	} else if kind == "replicaset" {
		r = newResourceForReplicaSet(name, resDefs, c)
	} else if kind == "petset" {
		r = newResourceForPetSet(name, resDefs, c)
	} else if kind == "daemonset" {
		r = newResourceForDaemonSet(name, resDefs, c)
	} else {
		return nil, fmt.Errorf("Not a proper resource kind: %s. Expected 'pod', 'job', 'service', 'replicaset' or 'daemonset'", kind)
	}

	return NewScheduledResourceFor(r), nil
}

//NewScheduledResourceFor returns new scheduled resource for given resource in init state
func NewScheduledResourceFor(r resources.Resource) *ScheduledResource {
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
		depGraph[parent].Meta[parent] = d.Meta

		depGraph[parent].RequiredBy = append(
			depGraph[parent].RequiredBy, depGraph[child])
	}

	log.Println("Looking for resource definitions not in dependency list")
	for _, r := range resDefList.Items {
		var sr *ScheduledResource

		if r.Pod != nil {
			sr = NewScheduledResourceFor(resources.NewPod(r.Pod, c.Pods()))
		} else if r.Job != nil {
			sr = NewScheduledResourceFor(resources.NewJob(r.Job, c.Jobs()))
		} else if r.Service != nil {
			sr = NewScheduledResourceFor(resources.NewService(r.Service, c.Services(), c))
		} else if r.ReplicaSet != nil {
			sr = NewScheduledResourceFor(resources.NewReplicaSet(r.ReplicaSet, c.ReplicaSets()))
		} else if r.PetSet != nil {
			sr = NewScheduledResourceFor(resources.NewPetSet(r.PetSet, c.PetSets(), c))
		} else if r.DaemonSet != nil {
			sr = NewScheduledResourceFor(resources.NewDaemonSet(r.DaemonSet, c.DaemonSets()))
		} else {
			return nil, fmt.Errorf("Found unsupported resource %v", r)
		}

		if _, ok := depGraph[sr.Key()]; !ok {
			log.Printf("Resource %s not found in dependecy graph yet, adding.", sr.Key())
			depGraph[sr.Key()] = sr
		}
	}

	return depGraph, nil
}

func createResources(toCreate chan *ScheduledResource, created chan string, ccLimiter chan struct{}) {

	for r := range toCreate {
		go func(r *ScheduledResource, created chan string, ccLimiter chan struct{}) {
			//Acquire sepmaphor
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
				go func(req *ScheduledResource) {
					for {
						time.Sleep(CheckInterval)
						req.RLock()
						if req.Status == Init {
							req.RUnlock()
							req.Lock()
							if req.Status == Init && !req.IsBlocked() {
								req.Status = Creating
								toCreate <- req
								req.Unlock()
								break
							} else {
								req.Unlock()
							}
						} else {
							req.RUnlock()
							break
						}
					}
				}(req)
			}

			log.Printf("Checking status for %s", r.Key())
			for {
				time.Sleep(CheckInterval)
				status, err := r.Resource.Status(nil)
				if err != nil {
					log.Printf("Error getting status for resource %s: %v", r.Key(), err)
				}
				if status == "ready" {
					log.Printf("%s is ready", r.Key())
					break
				}
			}

			r.Lock()
			r.Status = Ready
			r.Unlock()

			log.Printf("Resource %s created", r.Key())
			created <- r.Key()
			//Release semaphor
			<-ccLimiter
		}(r, created, ccLimiter)
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
			r.Lock()
			r.Status = Creating
			toCreate <- r
			r.Unlock()
		}
	}

	log.Printf("Wait for %d deps to create\n", depCount)
	for i := 0; i < depCount; i++ {
		<-created
	}
	close(toCreate)
	close(created)

	//TODO Make sure every KO gets created eventually
}

//DetectCycles implements Kosaraju's algorithm https://en.wikipedia.org/wiki/Kosaraju%27s_algorithm
//for detecting cycles in graph.
//We are depending on the fact that any strongly connected component of a graph is a cycle
//if it consists of more than one vertex
func DetectCycles(depGraph DependencyGraph) [][]*ScheduledResource {
	//is vertex visited in first phase of the algorithm
	visited := make(map[string]bool)
	//is vertex assigned to strongly connected component
	assigned := make(map[string]bool)

	//each key is root of strongly connected component
	//the slice consists of all vertices belonging to strongly connected component to which root belongs
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

	//if any strongly connected component consist of more than one vertex - it's a cycle
	var cycles [][]*ScheduledResource
	for _, component := range components {
		if len(component) > 1 {
			cycles = append(cycles, component)
		}
	}

	//detect self cycles - not part of Kosaraju's algorithm
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
		//if component is not yet initiated, make the slice
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

// GetStatus generates data for getting the status of deployment. Returns
// a DeploymentStatus and a human readable report string
// TODO: Allow for other formats of report (e.g. json for visualisations)
func (graph DependencyGraph) GetStatus() (DeploymentStatus, string, error) {
	report := make([]string, len(graph), len(graph))
	var readyExist, nonReadyExist bool
	var status DeploymentStatus
	var blockedStr string
	i := 0
	for key, resource := range graph {
		status, err := resource.Resource.Status(nil)
		if err != nil {
			status = err.Error()
			nonReadyExist = true
		}
		if status == "ready" {
			readyExist = true
		} else {
			nonReadyExist = true
		}
		if resource.IsBlocked() {
			blockedStr = ", Blocked"
		}
		report[i] = fmt.Sprintf(
			"%s, STATUS: %s%s", key, status, blockedStr,
		)
		i++
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
	return status, strings.Join(report, "\n"), nil
}
