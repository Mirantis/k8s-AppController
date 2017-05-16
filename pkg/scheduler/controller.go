// Copyright 2017 Mirantis
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
	"log"
	"sync"
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// ProcessDeploymentTasks picks deployment tasks (configMap objects) one by one in chronological order,
// restores graph options from it and deploys the graph. This is the main function of long-running deployment process
func ProcessDeploymentTasks(client client.Interface, stopChan <-chan struct{}) {
	added := make(chan *v1.ConfigMap)
	deleted := make(chan *v1.ConfigMap)
	defer close(added)
	defer close(deleted)

	opts := v1.ListOptions{
		LabelSelector: "AppController=FlowDeployment",
	}
	source := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			return client.ConfigMaps().List(opts)
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			return client.ConfigMaps().Watch(opts)
		},
	}
	go enqueueTasks(added, deleted, client)
	_, cfgController := cache.NewInformer(
		source,
		&v1.ConfigMap{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				added <- obj.(*v1.ConfigMap)
			},
			DeleteFunc: func(obj interface{}) {
				deleted <- obj.(*v1.ConfigMap)

			},
		},
	)
	cfgController.Run(stopChan)
}

func enqueueTasks(added, deleted <-chan *v1.ConfigMap, client client.Interface) {
	taskQueue := list.New()
	cond := sync.NewCond(&sync.Mutex{})
	listSync := &sync.Mutex{}
	go deployTasks(taskQueue, listSync, cond, client)
	for {
		select {
		case add := <-added:
			listSync.Lock()
			tasks := taskQueue.Front()
			if tasks == nil {
				taskQueue.PushFront(add)
			} else {
				for tasks != nil {
					v := tasks.Value.(*v1.ConfigMap)
					if isBefore(v, add) {
						taskQueue.InsertBefore(add, tasks)
						break
					}
					tasks = tasks.Next()
				}
				if tasks == nil {
					taskQueue.PushBack(add)
				}
			}
			listSync.Unlock()
			cond.Signal()
		case del := <-deleted:
			signal := false
			listSync.Lock()
			for n := taskQueue.Front(); n != nil && !signal; n = n.Next() {
				if n.Value.(*v1.ConfigMap).Name == del.Name {
					taskQueue.Remove(n)
					signal = true
				}
			}
			listSync.Unlock()
			if signal {
				cond.Signal()
			}
		}
	}
}

func isBefore(a, b *v1.ConfigMap) bool {
	if a.CreationTimestamp == b.CreationTimestamp {
		return a.ResourceVersion < b.ResourceVersion
	}
	return a.CreationTimestamp.Before(b.CreationTimestamp)
}

func deployTasks(taskQueue *list.List, mutex *sync.Mutex, cond *sync.Cond, client client.Interface) {
	var lastBack, processing *list.Element
	var abortChan chan struct{}
	for {
		var back *list.Element
		cond.L.Lock()
		for {
			mutex.Lock()
			back = taskQueue.Back()
			processingNodeWasRemoved := processing != nil &&
				taskQueue.Front() != processing && processing.Prev() == nil &&
				processing.Next() == nil
			mutex.Unlock()

			if processingNodeWasRemoved || back != nil && processing == nil {
				break
			}
			cond.Wait()
		}

		cond.L.Unlock()
		if processing != nil {
			if abortChan != nil {
				abortChan <- struct{}{}
				close(abortChan)
				abortChan = nil
				processing = nil
			}
			continue
		}
		if abortChan != nil {
			close(abortChan)
			abortChan = nil
		}
		if back != lastBack {
			time.Sleep(time.Second)
			lastBack = back
			continue
		}

		listElement := back.Value.(*v1.ConfigMap)

		abortChan = make(chan struct{}, 1)
		processing = back
		lastBack = processing.Prev()
		go func(ch <-chan struct{}, el *v1.ConfigMap) {
			defer func() {
				processing = nil
				client.ConfigMaps().Delete(el.Name, nil)
				cond.Signal()
			}()

			log.Println("Starting deployment task", el.Name)
			sched, opts, err := FromConfig(client, el.Data)
			if err != nil {
				log.Println(err)
				return
			}
			Deploy(sched, opts, true, ch)
		}(abortChan, listElement)
	}
}
