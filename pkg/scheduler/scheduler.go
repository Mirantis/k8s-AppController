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
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
	"github.com/Mirantis/k8s-AppController/pkg/resources"

	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
)

// CheckInterval is an interval between rechecking the tree for updates
const (
	CheckInterval = time.Millisecond * 1000
	WaitTimeout   = time.Second * 600
)

// ScheduledResource is a wrapper for Resource with attached relationship data
type ScheduledResource struct {
	Requires       []string
	RequiredBy     []string
	Started        bool
	Ignored        bool
	Error          error
	Existing       bool
	context        *graphContext
	usedInReplicas []string
	status         interfaces.ResourceStatus
	suffix         string
	interfaces.Resource
	// parentKey -> dependencyMetadata
	Meta map[string]map[string]string
	sync.RWMutex
}

// Key returns resource identifier with optional suffix
func (sr *ScheduledResource) Key() string {
	baseKey := sr.Resource.Key()
	if sr.suffix == "" {
		return baseKey
	}
	return baseKey + "/" + sr.suffix
}

// RequestCreation does not create a scheduled resource immediately, but updates status
// and puts the scheduled resource to corresponding channel. Returns true if
// scheduled resource creation was actually requested, false otherwise.
func (sr *ScheduledResource) RequestCreation(toCreate chan *ScheduledResource) bool {
	sr.RLock()
	// somebody already requested resource creation
	if sr.Started {
		sr.RUnlock()
		return true
	}

	sr.RUnlock()
	sr.Lock()
	defer sr.Unlock()

	if !sr.Started && !sr.IsBlocked() {
		sr.Started = true
		toCreate <- sr
		return true
	}
	return false
}

func isResourceFinished(sr *ScheduledResource, ch chan error) bool {
	status, err := sr.Status(nil)
	if err != nil {
		ch <- err
		return true
	}
	if status == interfaces.ResourceReady {
		ch <- nil
		return true
	}
	return false
}

// Wait periodically checks resource status and returns if the resource processing is finished,
// regardless successful or not. The actual result of processing could be obtained from returned error.
func (sr *ScheduledResource) Wait(checkInterval time.Duration, timeout time.Duration, stopChan <-chan struct{}) (bool, error) {
	ch := make(chan error, 1)
	go func(ch chan error) {
		log.Printf("%s flow: waiting for %v to be created", sr.context.graph.graphOptions.FlowName, sr.Key())
		if isResourceFinished(sr, ch) {
			return
		}
		ticker := time.NewTicker(checkInterval)
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				if isResourceFinished(sr, ch) {
					return
				}
			}
		}

	}(ch)

	select {
	case <-stopChan:
		return true, nil
	case err := <-ch:
		return false, err
	case <-time.After(timeout):
		e := fmt.Errorf("%s flow: timeout waiting for resource %s", sr.context.graph.graphOptions.FlowName, sr.Key())
		sr.Lock()
		defer sr.Unlock()
		sr.Error = e
		return false, e
	}
}

// Status either returns cached copy of resource's status or retrieves it via Resource.Status.
// Only ResourceReady is cached to avoid inconsistency for resources that may go to failure state over time
// so that if resource becomes ready it stays in this status for the whole deployment duration.
// Errors returned by the resource are never cached, however if AC sees permanent problem with resource it may set the
// error field
func (sr *ScheduledResource) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	sr.Lock()
	defer sr.Unlock()
	if sr.status != "" || sr.Error != nil {
		return sr.status, sr.Error
	}
	status, err := sr.Resource.Status(meta)
	if err == nil && status == interfaces.ResourceReady {
		sr.status = status
	}
	return status, err
}

// IsBlocked checks whether a scheduled resource can be created. It checks status of resources
// it depends on, via API
func (sr *ScheduledResource) IsBlocked() bool {
	for _, reqKey := range sr.Requires {
		meta := sr.Meta[reqKey]
		_, onErrorSet := meta["on-error"]
		req := sr.context.graph.graph[reqKey]

		status, err := req.Status(meta)

		req.RLock()
		ignored := req.Ignored
		req.RUnlock()

		if err != nil && !onErrorSet && !ignored {
			return true
		} else if status == "ready" && onErrorSet {
			return true
		} else if err == nil && status != "ready" {
			return true
		}
	}
	return false
}

// ResetStatus resets cached status of scheduled resource
func (sr *ScheduledResource) ResetStatus() {
	sr.Lock()
	defer sr.Unlock()
	sr.Error = nil
	sr.status = ""
}

func createResources(toCreate chan *ScheduledResource, finished chan string, ccLimiter chan struct{}, stopChan <-chan struct{}) {
	for r := range toCreate {
		log.Printf("Requesting creation of %v", r.Key())
		select {
		case <-stopChan:
			log.Println("Terminating creation of resources")
			return
		default:
			log.Println("Deployment is not stopped, keep creating")
		}
		go func(r *ScheduledResource, finished chan string, ccLimiter chan struct{}) {
			// Acquire semaphore
			ccLimiter <- struct{}{}
			defer func() {
				<-ccLimiter
			}()

			attempts := resources.GetIntMeta(r.Resource, "retry", 1)
			timeoutInSeconds := resources.GetIntMeta(r.Resource, "timeout", -1)
			onError := resources.GetStringMeta(r.Resource, "on-error", "")

			waitTimeout := WaitTimeout
			if timeoutInSeconds >= 0 {
				waitTimeout = time.Second * time.Duration(timeoutInSeconds)
			}

			for attemptNo := 1; attemptNo <= attempts; attemptNo++ {

				r.ResetStatus()

				var err error

				// NOTE(gluke77): We start goroutines for dependencies
				// before the resource becomes ready, since dependencies
				// could have metadata defining their own readiness condition
				if attemptNo == 1 {
					for _, reqKey := range r.RequiredBy {
						req := r.context.graph.graph[reqKey]
						go func(req *ScheduledResource, toCreate chan *ScheduledResource) {
							ticker := time.NewTicker(CheckInterval)
							for {
								select {
								case <-stopChan:
									log.Println("Terminating creation of dependencies")
									return
								case <-ticker.C:
									if req.RequestCreation(toCreate) {
										return
									}
								}
							}
						}(req, toCreate)
					}
				}

				if attemptNo > 1 && (!r.Existing || r.context.graph.graphOptions.AllowDeleteExternalResources) {
					log.Printf("Trying to delete resource %s after previous unsuccessful attempt", r.Key())
					err = r.Delete()
					if err != nil {
						log.Printf("Error deleting resource %s: %v", r.Key(), err)
					}
				}

				r.RLock()
				ignored := r.Ignored
				r.RUnlock()

				if ignored {
					log.Printf("Skipping creation of resource %s as being ignored", r.Key())
					break
				}

				log.Printf("Deploying resource %s, attempt %d of %d", r.Key(), attemptNo, attempts)
				err = r.Create()
				if err != nil {
					log.Printf("Error deploying resource %s: %v", r.Key(), err)
					continue
				}

				log.Printf("Checking status for %s", r.Key())

				stoped, err := r.Wait(CheckInterval, waitTimeout, stopChan)

				if stoped {
					log.Printf("Received interrupt while waiting for %v. Exiting", r.Key())
					return
				}

				if err == nil {
					log.Printf("Resource %s created", r.Key())
					break
				}

				log.Printf("Resource %s was not created: %v", r.Key(), err)

				if attemptNo >= attempts {
					if onError == "ignore" {
						r.Lock()
						r.Ignored = true
						log.Printf("Resource %s failure ignored -- prooceeding as normal", r.Key())
						r.Unlock()
					} else if onError == "ignore-all" {
						ignoreAll(r)
					}
				}
			}
			finished <- r.Key()
		}(r, finished, ccLimiter)
	}
}

func ignoreAll(top *ScheduledResource) {
	top.Lock()
	top.Ignored = true
	top.Unlock()

	log.Printf("Marking resource %s as ignored", top.Key())

	for _, childKey := range top.RequiredBy {
		child := top.context.graph.graph[childKey]
		ignoreAll(child)
	}
}

// Options method returns options that were used to build the dependency graph
func (depGraph dependencyGraph) Options() interfaces.DependencyGraphOptions {
	return depGraph.graphOptions
}

// Deploy starts the deployment of a DependencyGraph
func (depGraph dependencyGraph) Deploy(stopChan <-chan struct{}) {

	depCount := len(depGraph.graph)
	concurrency := depGraph.scheduler.concurrency

	concurrencyLimiterLen := depCount
	if concurrency > 0 && concurrency < concurrencyLimiterLen {
		concurrencyLimiterLen = concurrency
	}

	ccLimiter := make(chan struct{}, concurrencyLimiterLen)
	toCreate := make(chan *ScheduledResource, depCount)
	created := make(chan string, depCount)
	defer func() {
		close(toCreate)
		close(created)
	}()

	go createResources(toCreate, created, ccLimiter, stopChan)

	for _, r := range depGraph.graph {
		if len(r.Requires) == 0 {
			r.RequestCreation(toCreate)
		}
	}

	log.Printf("%s flow: waiting for %d resource deployments", depGraph.graphOptions.FlowName, depCount)
	for i := 0; i < depCount; {
		select {
		case <-stopChan:
			log.Printf("Deployment of %s is stopped", depGraph.graphOptions.FlowName)
			return
		case <-created:
			i++
			log.Printf("%s flow: %v out of %v were created", depGraph.graphOptions.FlowName, i, depCount)
		}
	}
	if depGraph.finalizer != nil {
		depGraph.finalizer()
	}
	// TODO Make sure every KO gets created eventually
}

// GetNodeReport acts as a more verbose version of IsBlocked. It performs the
// same check as IsBlocked, but returns the DeploymentReport
func (sr *ScheduledResource) GetNodeReport(name string) report.NodeReport {
	var ready bool
	isBlocked := false
	dependencies := make([]interfaces.DependencyReport, 0, len(sr.Requires))
	status, err := sr.Status(nil)
	if err != nil {
		ready = false
	} else {
		ready = status == "ready"
	}
	for _, rKey := range sr.Requires {
		r := sr.context.graph.graph[rKey]
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
func (depGraph dependencyGraph) GetStatus() (interfaces.DeploymentStatus, interfaces.DeploymentReport) {
	var readyExist, nonReadyExist bool
	var status interfaces.DeploymentStatus
	deploymentReport := make(report.DeploymentReport, 0, len(depGraph.graph))
	for key, resource := range depGraph.graph {
		depReport := resource.GetNodeReport(key)
		deploymentReport = append(deploymentReport, depReport)
		if depReport.Ready {
			readyExist = true
		} else {
			nonReadyExist = true
		}
	}
	switch {
	case readyExist && nonReadyExist:
		status = interfaces.Running
	case readyExist:
		status = interfaces.Finished
	case nonReadyExist:
		status = interfaces.Prepared
	default:
		status = interfaces.Empty
	}
	return status, deploymentReport
}

func runConcurrently(funcs []func() bool, concurrency int) bool {
	if concurrency < 1 {
		concurrency = len(funcs)
	}

	sem := make(chan bool, concurrency)
	defer close(sem)

	result := true

	for _, f := range funcs {
		sem <- true
		go func(foo func() bool) {
			defer func() { <-sem }()
			if !foo() {
				result = false
			}
		}(f)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return result
}

func deleteResource(resource *ScheduledResource) error {
	if !resource.Existing || resource.context.graph.graphOptions.AllowDeleteExternalResources {
		log.Printf("%s flow: Deleting resource %s", resource.context.flow.Name, resource.Key())
		err := resource.Delete()
		if err != nil {
			statusError, ok := err.(*errors.StatusError)
			if ok && statusError.Status().Reason == unversioned.StatusReasonNotFound {
				return nil
			}
			log.Printf("cannot delete resource %s: %v", resource.Key(), err)
		}
		return err
	}
	log.Printf("%s: Won't delete external resource %s", resource.context.flow.Name, resource.Key())
	return nil
}
