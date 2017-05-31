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

	api_errors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
)

// CheckInterval is an interval between rechecking the tree for updates
const (
	CheckInterval = time.Millisecond * 1000
	WaitTimeout   = time.Second * 600
)

// scheduledResource is a wrapper for Resource with attached relationship data
type scheduledResource struct {
	requires         []string
	requiredBy       []string
	started          bool
	ignored          bool
	skipped          bool
	error            error
	existing         bool
	context          *graphContext
	usedInReplicas   []string
	finished         bool
	distance         int
	suffix           string
	resourceMeta     map[string]interface{}
	dependenciesMeta map[string]map[string]string
	interfaces.Resource
	// parentKey -> dependencyMetadata
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

// requestCreation does not create a scheduled resource immediately, but updates state
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
	progress, err := sr.Resource.GetProgress()
	if err == nil {
		return progress == 1, nil
	}
	return false, nil
}

// wait periodically checks resource deployment progress and returns if the resource processing is finished,
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

// GetProgress either returns cached copy of resource's progress or retrieves it via Resource.GetProgress.
// Only 100% completion is cached to avoid inconsistency for resources that may go to failure state over time
// so that if resource becomes ready it stays in this status for the whole deployment duration.
// Errors returned by the resource are never cached, however if AC sees permanent problem with resource it may set the
// error field
func (sr *scheduledResource) GetProgress() (float32, error) {
	sr.Lock()
	defer sr.Unlock()
	if sr.error != nil {
		return 0, sr.error
	}
	if sr.finished {
		return 1, nil
	}
	progress, err := sr.Resource.GetProgress()
	if err == nil && progress == 1 {
		sr.finished = true
	}
	return progress, err
}

func (sr *scheduledResource) isBlocked() bool {
	return len(sr.getBlockedBy()) > 0
}

func (sr *scheduledResource) getBlockedBy() []string {
	blockedBy := make([]string, 0, len(sr.requires))
	var permanentlyBlockedKeys []string
	skipped := true

	for _, reqKey := range sr.requires {
		meta := sr.dependenciesMeta[reqKey]
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
		progress, err := req.GetProgress()
		if onErrorSet {
			if permanentError == nil {
				blockedBy = append(blockedBy, reqKey)
				if err == nil && progress == 1 {
					permanentlyBlockedKeys = append(permanentlyBlockedKeys, reqKey)
				}
			}
		} else {
			skipped = false
			threshold := getSuccessFactor(meta)
			if permanentError != nil || err != nil || progress < threshold {
				blockedBy = append(blockedBy, reqKey)
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

	return blockedBy
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

			attempts := getIntMeta(r, "retry", 1)
			timeoutInSeconds := getIntMeta(r, "timeout", -1)
			onError := getStringMeta(r, "on-error", "")

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
							if req.requestCreation(toCreate) {
								return
							}
							ticker := time.NewTicker(CheckInterval)
							log.Printf("Requesting creation of dependency %v", req.Key())
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

//// getNodeReport acts as a more verbose version of isBlocked. It performs the
//// same check as isBlocked, but returns the DeploymentReport
//func (sr *scheduledResource) getNodeReport(name string) report.NodeReport {
//	var ready bool
//	isBlocked := false
//	dependencies := make([]interfaces.DependencyReport, 0, len(sr.requires))
//	status, err := sr.GetProgress(nil)
//	if err != nil {
//		ready = false
//	} else {
//		ready = status == "ready"
//	}
//	for _, rKey := range sr.requires {
//		r := sr.context.graph.graph[rKey]
//		r.RLock()
//		meta := r.meta[sr.Key()]
//		depReport := r.GetDependencyReport(meta)
//		r.RUnlock()
//		if depReport.Blocks {
//			isBlocked = true
//		}
//		dependencies = append(dependencies, depReport)
//	}
//	return report.NodeReport{
//		Dependent:    name,
//		Dependencies: dependencies,
//		Blocked:      isBlocked,
//		Ready:        ready,
//	}
//}

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
			statusError, ok := err.(*api_errors.StatusError)
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
