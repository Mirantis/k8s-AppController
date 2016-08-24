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

type Resource interface {
	Key() string
	Status() (string, error)
	Create() error
}

type ScheduledResourceStatus int

const (
	Init ScheduledResourceStatus = iota
	Creating
	Ready
)

type ScheduledResource struct {
	Requires   []*ScheduledResource
	RequiredBy []*ScheduledResource
	Status     ScheduledResourceStatus
	Resource
	sync.RWMutex
}

func (d *ScheduledResource) IsBlocked() bool {
	isBlocked := false

	for _, r := range d.Requires {
		r.RLock()
		if r.Status != Ready {
			isBlocked = true
		}
		r.RUnlock()
	}

	return isBlocked
}

type DependencyGraph map[string]*ScheduledResource

func newResourceForPod(name string, resDefs []client.ResourceDefinition, c client.Interface) Resource {
	for _, rd := range resDefs {
		if rd.Pod != nil && rd.Pod.Name == name {
			log.Println("Found resource definition for pod", name)
			return resources.NewPod(rd.Pod, c.Pods())
		}
	}

	log.Printf("Resource definition for pod '%s' not found, so it is expected to exist already")
	return resources.NewExistingPod(name, c.Pods())
}

func newResourceForJob(name string, resDefs []client.ResourceDefinition, c client.Interface) Resource {
	for _, rd := range resDefs {
		if rd.Job != nil && rd.Job.Name == name {
			log.Println("Found resource definition for job", name)
			return resources.NewJob(rd.Job, c.Jobs())
		}
	}

	log.Printf("Resource definition for job '%s' not found, so it is expected to exist already")
	return resources.NewExistingJob(name, c.Jobs())
}

func newResourceForService(name string, resDefs []client.ResourceDefinition, c client.Interface) Resource {
	for _, rd := range resDefs {
		if rd.Service != nil && rd.Service.Name == name {
			log.Println("Found resource definition for service", name)
			return resources.NewService(rd.Service, c.Services())
		}
	}

	log.Printf("Resource definition for service '%s' not found, so it is expected to exist already")
	return resources.NewExistingService(name, c.Services())
}

func newScheduledResource(kind string, name string,
	resDefs []client.ResourceDefinition, c client.Interface) (*ScheduledResource, error) {

	var r Resource

	if kind == "pod" {
		r = newResourceForPod(name, resDefs, c)
	} else if kind == "job" {
		r = newResourceForJob(name, resDefs, c)
	} else if kind == "service" {
		r = newResourceForService(name, resDefs, c)
	} else {
		return nil, fmt.Errorf("Not a proper resource kind: %s. Expected 'pod','job' or 'service'", kind)
	}

	return newScheduledResourceFor(r), nil
}

func newScheduledResourceFor(r Resource) *ScheduledResource {
	return &ScheduledResource{
		Status:   Init,
		Resource: r,
	}
}

func keyParts(key string) (kind string, name string, err error) {
	parts := strings.Split(key, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("Not a proper resource key: %s. Expected KIND/NAME", key)
	}

	return parts[0], parts[1], nil
}

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

				sr, err := newScheduledResource(kind, name, resDefs, c)
				if err != nil {
					return nil, err
				}

				depGraph[key] = sr
			}
		}

		depGraph[child].Requires = append(
			depGraph[d.Child].Requires, depGraph[parent])
		depGraph[parent].RequiredBy = append(
			depGraph[parent].RequiredBy, depGraph[child])
	}

	log.Println("Looking for resource definitions not in dependency list")
	for _, r := range resDefList.Items {
		var sr *ScheduledResource

		if r.Pod != nil {
			sr = newScheduledResourceFor(resources.NewPod(r.Pod, c.Pods()))
		} else if r.Job != nil {
			sr = newScheduledResourceFor(resources.NewJob(r.Job, c.Jobs()))
		} else if r.Service != nil {
			sr = newScheduledResourceFor(resources.NewService(r.Service, c.Services()))
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

			for {
				time.Sleep(time.Millisecond * 100)
				status, err := r.Resource.Status()
				if err != nil {
					log.Printf("Error getting status for resource %s: %v", r.Key(), err)
				}
				if status == "ready" {
					break
				}
			}

			r.Lock()
			r.Status = Ready
			r.Unlock()

			log.Printf("Resource %s created", r.Key())

			for _, req := range r.RequiredBy {
				if !req.IsBlocked() {
					req.RLock()
					if req.Status == Init {
						req.RUnlock()
						req.Lock()
						if req.Status == Init {
							req.Status = Creating
							toCreate <- req
						}
						req.Unlock()
					} else {
						req.RUnlock()
					}
				}
			}
			created <- r.Key()
			//Release semaphor
			<-ccLimiter
		}(r, created, ccLimiter)
	}
}

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
	cycles := make([][]*ScheduledResource, 0)
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
