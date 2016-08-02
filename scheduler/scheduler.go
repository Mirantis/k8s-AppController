package scheduler

import (
	"log"
	"sync"
	"time"
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

type Dependency struct {
	//  The one who is being depend on
	Parent string
	// The one who depends on
	Child string
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
		depGraph[d.Child].Requires = append(
			depGraph[d.Child].Requires, depGraph[d.Parent])
		depGraph[d.Parent].RequiredBy = append(
			depGraph[d.Parent].RequiredBy, depGraph[d.Child])
	}

	return depGraph

	//TODO Handle external deps
	//TODO Check cycles in graph
	//TODO Handle deps not in kos
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

	log.Printf("Wait for %d deps to create\n", depCount)
	for i := 0; i < depCount; i++ {
		<-created
	}
	close(toCreate)
	close(created)

	//TODO Make sure every KO gets created eventually
}
