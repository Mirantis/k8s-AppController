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
	"strings"
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

// scheduledResource is a wrapper for Resource with attached relationship data
type scheduledResource struct {
	requires       []string
	requiredBy     []string
	started        bool
	ignored        bool
	skipped        bool
	error          error
	existing       bool
	context        *graphContext
	usedInReplicas []string
	status         interfaces.ResourceStatus
	suffix         string
	interfaces.Resource
	// parentKey -> dependencyMetadata
	meta map[string]map[string]string
	sync.RWMutex
}

// Key returns resource identifier with optional suffix
func (sr *scheduledResource) Key() string {
	baseKey := sr.Resource.Key()
	if sr.suffix == "" {
		return baseKey
	}
	return baseKey + "/" + sr.suffix
}

// requestCreation does not create a scheduled resource immediately, but updates status
// and puts the scheduled resource to corresponding channel. Returns true if
// scheduled resource creation was actually requested, false otherwise.
func (sr *scheduledResource) requestCreation(toCreate chan<- *scheduledResource) bool {
	sr.RLock()
	// somebody already requested resource creation
	if sr.started {
		sr.RUnlock()
		return true
	}

	sr.RUnlock()
	isBlocked := sr.isBlocked() && sr.error == nil
	sr.Lock()
	defer sr.Unlock()

	if !sr.started && !isBlocked {
		sr.started = true
		toCreate <- sr
		return true
	}
	return false
}

func isResourceFinished(sr *scheduledResource) (bool, error) {
	status, err := sr.Status(nil)
	if err != nil {
		return status != interfaces.ResourceNotReady, err
	}
	if status == interfaces.ResourceReady {
		return true, nil
	}
	return false, nil
}

// wait periodically checks resource status and returns if the resource processing is finished,
// regardless successful or not. The actual result of processing could be obtained from returned error.
func (sr *scheduledResource) wait(checkInterval time.Duration, timeout time.Duration, stopChan <-chan struct{}) (bool, error) {
	log.Printf("%s flow: waiting for %v to be created", sr.context.graph.graphOptions.FlowName, sr.Key())
	var err error
	var finished bool
	if finished, err = isResourceFinished(sr); finished {
		return false, err
	}
	ticker := time.NewTicker(checkInterval)
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-stopChan:
			return true, nil
		case <-ticker.C:
			if finished, err = isResourceFinished(sr); finished {
				return false, err
			}
		case <-timeoutChan:
			if err == nil {
				err = fmt.Errorf("%s flow: timeout waiting for resource %s", sr.context.graph.graphOptions.FlowName, sr.Key())
			}
			return false, err
		}
	}
}

// Status either returns cached copy of resource's status or retrieves it via Resource.Status.
// Only ResourceReady is cached to avoid inconsistency for resources that may go to failure state over time
// so that if resource becomes ready it stays in this status for the whole deployment duration.
// Errors returned by the resource are never cached, however if AC sees permanent problem with resource it may set the
// error field
func (sr *scheduledResource) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	sr.Lock()
	defer sr.Unlock()
	if sr.status != "" || sr.error != nil {
		return sr.status, sr.error
	}
	status, err := sr.Resource.Status(meta)
	if err == nil && status == interfaces.ResourceReady {
		sr.status = status
	}
	return status, err
}

// isBlocked checks whether a scheduled resource can be created. It checks status of resources
// it depends on, via API.
func (sr *scheduledResource) isBlocked() bool {
	isBlocked := false
	var permanentlyBlockedKeys []string
	skipped := true

	for _, reqKey := range sr.requires {
		meta := sr.meta[reqKey]
		onErrorValue, onErrorSet := meta["on-error"]
		onErrorSet = onErrorSet && onErrorValue != "false"
		req := sr.context.graph.graph[reqKey]

		req.RLock()
		ignored := req.ignored
		permanentError := req.error
		req.RUnlock()

		if ignored {
			continue
		}
		status, err := req.Status(meta)
		if onErrorSet {
			if permanentError == nil {
				isBlocked = true
				if err == nil && status == interfaces.ResourceReady {
					permanentlyBlockedKeys = append(permanentlyBlockedKeys, reqKey)
				}
			}
		} else {
			skipped = false
			if permanentError != nil || err != nil || status != interfaces.ResourceReady {
				isBlocked = true
				if permanentError != nil {
					permanentlyBlockedKeys = append(permanentlyBlockedKeys, reqKey)
				}
			}
		}
	}
	if len(permanentlyBlockedKeys) > 0 {
		sr.Lock()
		sr.error = fmt.Errorf("permanently blocked because of (%s) dependencies", strings.Join(permanentlyBlockedKeys, ","))
		sr.skipped = skipped
		sr.Unlock()
	}

	return isBlocked
}

func createResources(toCreate chan *scheduledResource, finished chan<- *scheduledResource, ccLimiter chan struct{}, stopChan <-chan struct{}) {
	for r := range toCreate {
		select {
		case <-stopChan:
			log.Println("Terminating creation of resources")
			return
		default:
		}
		go func(r *scheduledResource, finished chan<- *scheduledResource, ccLimiter chan struct{}) {
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
				var err error

				// NOTE(gluke77): We start goroutines for dependencies
				// before the resource becomes ready, since dependencies
				// could have metadata defining their own readiness condition
				if attemptNo == 1 {
					for _, reqKey := range r.requiredBy {
						req := r.context.graph.graph[reqKey]
						go func(req *scheduledResource, toCreate chan<- *scheduledResource) {
							ticker := time.NewTicker(CheckInterval)
							for {
								select {
								case <-stopChan:
									log.Println("Terminating creation of dependencies")
									return
								case <-ticker.C:
									if req.requestCreation(toCreate) {
										return
									}
								}
							}
						}(req, toCreate)
					}
				}

				if attemptNo > 1 && (!r.existing || r.context.graph.graphOptions.AllowDeleteExternalResources) {
					log.Printf("Trying to delete resource %s after previous unsuccessful attempt", r.Key())
					err = r.Delete()
					if err != nil {
						log.Printf("Error deleting resource %s: %v", r.Key(), err)
					}
				}

				r.RLock()
				ignored := r.ignored
				srErr := r.error
				r.RUnlock()

				if ignored || srErr != nil {
					log.Printf("Skipping creation of resource %s", r.Key())
					break
				}

				log.Printf("Deploying resource %s, attempt %d of %d", r.Key(), attemptNo, attempts)
				err = r.Create()
				if err != nil {
					log.Printf("Error deploying resource %s: %v", r.Key(), err)
				} else {
					log.Printf("Checking status for %s", r.Key())

					var stopped bool
					stopped, err = r.wait(CheckInterval, waitTimeout, stopChan)

					if stopped {
						log.Printf("Received interrupt while waiting for %v. Exiting", r.Key())
						return
					}

					if err == nil {
						log.Printf("Resource %s created", r.Key())
						break
					}
					log.Printf("Resource %s was not created: %v", r.Key(), err)
				}

				if attemptNo == attempts {
					switch onError {
					case "ignore":
						r.Lock()
						r.ignored = true
						log.Printf("Resource %s failure ignored -- prooceeding as normal", r.Key())
						r.Unlock()
					case "ignore-all":
						ignoreAll(r)
					default:
						r.Lock()
						r.error = err
						r.Unlock()
					}
				}
			}
			finished <- r
		}(r, finished, ccLimiter)
	}
}

func ignoreAll(top *scheduledResource) {
	top.Lock()
	top.ignored = true
	top.Unlock()

	log.Printf("Marking resource %s as ignored", top.Key())

	for _, childKey := range top.requiredBy {
		child := top.context.graph.graph[childKey]
		ignoreAll(child)
	}
}

// Options method returns options that were used to build the dependency graph
func (depGraph dependencyGraph) Options() interfaces.DependencyGraphOptions {
	return depGraph.graphOptions
}

// Deploy starts the deployment of a DependencyGraph
func (depGraph dependencyGraph) Deploy(stopChan <-chan struct{}) bool {
	depCount := len(depGraph.graph)
	concurrency := depGraph.scheduler.concurrency

	concurrencyLimiterLen := depCount
	if concurrency > 0 && concurrency < concurrencyLimiterLen {
		concurrencyLimiterLen = concurrency
	}

	ccLimiter := make(chan struct{}, concurrencyLimiterLen)
	toCreate := make(chan *scheduledResource, depCount)
	created := make(chan *scheduledResource, depCount)
	defer func() {
		close(toCreate)
		close(created)
	}()

	go createResources(toCreate, created, ccLimiter, stopChan)

	for _, r := range depGraph.graph {
		if len(r.requires) == 0 {
			r.requestCreation(toCreate)
		}
	}

	log.Printf("%s flow: waiting for %d resource deployments", depGraph.graphOptions.FlowName, depCount)
	result := true
	for i := 0; i < depCount; {
		select {
		case <-stopChan:
			log.Printf("Deployment of %s is stopped", depGraph.graphOptions.FlowName)
			return false
		case sr := <-created:
			if sr.error != nil && !sr.skipped {
				result = false
			}
			i++
			log.Printf("%s flow: %v out of %v were created", depGraph.graphOptions.FlowName, i, depCount)
		}
	}
	if depGraph.finalizer != nil {
		depGraph.finalizer(stopChan)
	}
	return result
}

// getNodeReport acts as a more verbose version of isBlocked. It performs the
// same check as isBlocked, but returns the DeploymentReport
func (sr *scheduledResource) getNodeReport(name string) report.NodeReport {
	var ready bool
	isBlocked := false
	dependencies := make([]interfaces.DependencyReport, 0, len(sr.requires))
	status, err := sr.Status(nil)
	if err != nil {
		ready = false
	} else {
		ready = status == "ready"
	}
	for _, rKey := range sr.requires {
		r := sr.context.graph.graph[rKey]
		r.RLock()
		meta := r.meta[sr.Key()]
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
		depReport := resource.getNodeReport(key)
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

func runConcurrently(funcs []func(<-chan struct{}) bool, concurrency int, stopChan <-chan struct{}) bool {
	if concurrency < 1 {
		concurrency = len(funcs)
	}

	sem := make(chan bool, concurrency)
	defer close(sem)

	result := true

	for _, f := range funcs {
		sem <- true
		select {
		case <-stopChan:
			result = false
			<-sem
		default:
			go func(foo func(<-chan struct{}) bool) {
				defer func() { <-sem }()
				if !foo(stopChan) {
					result = false
				}
			}(f)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return result
}

func deleteResource(resource *scheduledResource) error {
	if !resource.existing || resource.context.graph.graphOptions.AllowDeleteExternalResources {
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
