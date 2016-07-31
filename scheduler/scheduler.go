package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type Resource interface {
	Key() string
	Status() string
	Create()
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

type Dependency struct {
	// The one who depends on
	Depender string
	//  The one who is being depend on
	Dependent string
}

func buildDependencyGraph(resources []Resource, deps []Dependency) DependencyGraph {
	depGraph := DependencyGraph{}

	for _, r := range resources {
		depGraph[r.Key()] = &ScheduledResource{
			Status:   Init,
			Resource: r,
		}
	}

	for _, d := range deps {
		depGraph[d.Depender].Requires = append(
			depGraph[d.Depender].Requires, depGraph[d.Dependent])
		depGraph[d.Dependent].RequiredBy = append(
			depGraph[d.Dependent].RequiredBy, depGraph[d.Depender])
	}

	return depGraph

	//TODO Handle external deps
	//TODO Check cycles in graph
	//TODO Handle deps not in kos
}

func createResources(toCreate chan *ScheduledResource, created chan string) {

	for r := range toCreate {
		go func(r *ScheduledResource) {
			r.Create()

			for r.Resource.Status() != "ready" {
				time.Sleep(time.Millisecond * 100)
			}

			r.Lock()
			r.Status = Ready
			r.Unlock()

			fmt.Println("Ready", r.Key())

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

func Create(resources []Resource, deps []Dependency) {
	depGraph := buildDependencyGraph(resources, deps)
	depCount := len(depGraph)
	toCreate := make(chan *ScheduledResource, depCount)
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

	fmt.Printf("Wait for %d deps to create\n", depCount)
	for i := 0; i < depCount; i++ {
		key := <-created
		fmt.Println(i, "- Created ", key)
	}
	close(toCreate)
	close(created)

	//TODO Make sure every KO gets created eventually
}
