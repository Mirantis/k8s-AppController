package scheduler

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/kubernetes"
	"github.com/Mirantis/k8s-AppController/resources"
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

func buildDependencyGraph(url string, resList *client.ResourceDefinitionList,
	depList *client.DependencyList) DependencyGraph {

	depGraph := DependencyGraph{}

	kc, err := kubernetes.Client(url)
	if err != nil {
		log.Fatal(err)
	}

	pods := kc.Pods()
	jobs := kc.Jobs()

	for _, r := range resList.Items {
		if r.Pod != nil {
			sr := newScheduledResource(resources.NewPod(r.Pod, pods))
			depGraph[sr.Key()] = sr
			log.Println("Found pod definition", r.Pod.Name, r.Pod)
		} else if r.Job != nil {
			sr := newScheduledResource(resources.NewJob(r.Job, jobs))
			depGraph[sr.Key()] = sr
			log.Println("Found job definition", r.Job.Name, r.Job)
		} else {
			log.Fatalf("Found unsupported resource", r)
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
				log.Fatalf("Not a proper resource key: %s. Expected RESOURCE_TYPE/NAME", d.Parent)
			}

			typ := keyParts[0]
			name := keyParts[1]

			if typ == "pod" {
				depGraph[d.Parent] = newScheduledResource(resources.NewExistingPod(name, pods))
			} else if typ == "job" {
				depGraph[d.Parent] = newScheduledResource(resources.NewExistingJob(name, jobs))
			} else {
				log.Fatalf("Not a proper resource type: %s. Expected 'pod' or 'job'", typ)
			}
		}

		depGraph[d.Child].Requires = append(
			depGraph[d.Child].Requires, depGraph[d.Parent])
		depGraph[d.Parent].RequiredBy = append(
			depGraph[d.Parent].RequiredBy, depGraph[d.Child])
	}

	return depGraph

	//TODO Handle external depList
	//TODO Check cycles in graph
	//TODO Handle depList not in kos
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

func Create(url string, resList *client.ResourceDefinitionList,
	depList *client.DependencyList) {

	depGraph := buildDependencyGraph(url, resList, depList)
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

	log.Printf("Wait for %d depList to create\n", depCount)
	for i := 0; i < depCount; i++ {
		<-created
	}
	close(toCreate)
	close(created)

	//TODO Make sure every KO gets created eventually
}
