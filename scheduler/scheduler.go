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

func newScheduledResource(r Resource) *ScheduledResource {
	return &ScheduledResource{
		Status:   Init,
		Resource: r,
	}
}

func BuildDependencyGraph(c client.Interface, sel labels.Selector) (*DependencyGraph, error) {

	log.Println("Getting resource definitions")
	resDefList, err := c.ResourceDefinitions().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	log.Println("Getting dependencies")
	depList, err := c.Dependencies().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	depGraph := DependencyGraph{}

	pods := c.Pods()
	jobs := c.Jobs()
	services := c.Services()

	for _, r := range resDefList.Items {
		if r.Pod != nil {
			sr := newScheduledResource(resources.NewPod(r.Pod, pods))
			depGraph[sr.Key()] = sr
			log.Println("Found pod definition", r.Pod.Name, r.Pod)
		} else if r.Job != nil {
			sr := newScheduledResource(resources.NewJob(r.Job, jobs))
			depGraph[sr.Key()] = sr
			log.Println("Found job definition", r.Job.Name, r.Job)
		} else if r.Service != nil {
			sr := newScheduledResource(resources.NewService(r.Service, services))
			depGraph[sr.Key()] = sr
			log.Println("Found service definition", r.Service.Name, r.Service)
		} else {
			return nil, fmt.Errorf("Found unsupported resource %v", r)
		}
	}

	for _, d := range depList.Items {
		log.Println("Found dependency", d.Parent, "->", d.Child)

		if _, ok := depGraph[d.Child]; !ok {
			log.Printf("Resource %s is not defined. Skipping dependency %s -> %s",
				d.Child, d.Parent, d.Child)
			continue
		}

		if _, ok := depGraph[d.Parent]; !ok {
			log.Printf("Resource %s is not defined, but %s requires it. So %s is expected to exist",
				d.Parent, d.Child, d.Parent)

			keyParts := strings.Split(d.Parent, "/")

			if len(keyParts) < 2 {
				return nil, fmt.Errorf("Not a proper resource key: %s. Expected RESOURCE_TYPE/NAME", d.Parent)
			}

			typ := keyParts[0]
			name := keyParts[1]

			if typ == "pod" {
				depGraph[d.Parent] = newScheduledResource(resources.NewExistingPod(name, pods))
			} else if typ == "job" {
				depGraph[d.Parent] = newScheduledResource(resources.NewExistingJob(name, jobs))
			} else if typ == "service" {
				depGraph[d.Parent] = newScheduledResource(resources.NewExistingService(name, services))
			} else {
				return nil, fmt.Errorf("Not a proper resource type: %s. Expected 'pod','job' or 'service'", typ)
			}
		}

		depGraph[d.Child].Requires = append(
			depGraph[d.Child].Requires, depGraph[d.Parent])
		depGraph[d.Parent].RequiredBy = append(
			depGraph[d.Parent].RequiredBy, depGraph[d.Child])
	}

	return &depGraph, nil
}

func createResources(toCreate chan *ScheduledResource, created chan string) {

	for r := range toCreate {
		go func(r *ScheduledResource) {
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
		}(r)
	}
}

func Create(depGraph DependencyGraph, concurrency int) {

	depCount := len(depGraph)

	channelSize := depCount
	if concurrency > 0 && concurrency < depCount {
		channelSize = concurrency
	}

	toCreate := make(chan *ScheduledResource, channelSize)
	created := make(chan string, depCount)

	go createResources(toCreate, created)

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
//for dedecting cycles in graph.
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
