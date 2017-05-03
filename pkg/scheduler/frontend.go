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
	"log"
	"strconv"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/labels"
)

type Scheduler struct {
	client      client.Interface
	selector    labels.Selector
	concurrency int
}

// New creates and initializes instance of Scheduler
func New(client client.Interface, selector labels.Selector, concurrency int) interfaces.Scheduler {
	return &Scheduler{
		client:      client,
		selector:    selector,
		concurrency: concurrency,
	}
}

// Serialize converts scheduler settings and DependencyGraphOptions into string map which can be saved in ConfigMap
// object. Together is has all the information required to build and deploy dependency graph
func (sched *Scheduler) Serialize(options interfaces.DependencyGraphOptions) map[string]string {
	result := map[string]string{
		"selector":                     sched.selector.String(),
		"concurrency":                  strconv.Itoa(sched.concurrency),
		"flowName":                     options.FlowName,
		"exportedOnly":                 strconv.FormatBool(options.ExportedOnly),
		"allowUndeclaredArgs":          strconv.FormatBool(options.AllowUndeclaredArgs),
		"replicaCount":                 strconv.Itoa(options.ReplicaCount),
		"fixedNumberOfReplicas":        strconv.FormatBool(options.FixedNumberOfReplicas),
		"minReplicaCount":              strconv.Itoa(options.MinReplicaCount),
		"maxReplicaCount":              strconv.Itoa(options.MaxReplicaCount),
		"allowDeleteExternalResources": strconv.FormatBool(options.AllowDeleteExternalResources),
	}
	for key, value := range options.Args {
		result["arg."+key] = value
	}
	return result
}

// FromConfig deserializes setting map produced by Serialize method into Scheduler and DependencyGraphOptions objects
func FromConfig(client client.Interface, cfg map[string]string) (interfaces.Scheduler, interfaces.DependencyGraphOptions, error) {
	var concurrency int
	var opts interfaces.DependencyGraphOptions
	var selector labels.Selector

	getters := []func() error{
		func() error { return getSelector(cfg, "selector", &selector) },
		func() error { return getInt(cfg, "concurrency", &concurrency) },
		func() error { return getString(cfg, "flowName", &opts.FlowName) },
		func() error { return getBool(cfg, "exportedOnly", &opts.ExportedOnly) },
		func() error { return getBool(cfg, "allowUndeclaredArgs", &opts.AllowDeleteExternalResources) },
		func() error { return getInt(cfg, "replicaCount", &opts.ReplicaCount) },
		func() error { return getBool(cfg, "fixedNumberOfReplicas", &opts.FixedNumberOfReplicas) },
		func() error { return getInt(cfg, "minReplicaCount", &opts.MinReplicaCount) },
		func() error { return getInt(cfg, "maxReplicaCount", &opts.MaxReplicaCount) },
		func() error { return getBool(cfg, "allowDeleteExternalResources", &opts.AllowDeleteExternalResources) },
	}

	for _, getter := range getters {
		if err := getter(); err != nil {
			return nil, opts, err
		}
	}

	opts.Args = map[string]string{}
	for key, value := range cfg {
		if strings.HasPrefix(key, "arg.") {
			opts.Args[key[4:]] = value
		}
	}

	if opts.FlowName == "" {
		opts.FlowName = interfaces.DefaultFlowName
	}

	return New(client, selector, concurrency), opts, nil
}

func getInt(cfg map[string]string, key string, out *int) (err error) {
	val := cfg[key]
	if val == "" {
		*out = 0
		return nil
	}
	*out, err = strconv.Atoi(val)
	return
}

func getBool(cfg map[string]string, key string, out *bool) (err error) {
	val := cfg[key]
	if val == "" {
		*out = false
		return nil
	}
	*out, err = strconv.ParseBool(val)
	return
}

func getString(cfg map[string]string, key string, out *string) error {
	*out = cfg[key]
	return nil
}

func getSelector(cfg map[string]string, key string, out *labels.Selector) (err error) {
	*out, err = labels.Parse(cfg[key])
	return err
}

// CreateDeployment creates deployment task (ConfigMap object) for standalone deployment process and returns its name/error
func (sched *Scheduler) CreateDeployment(options interfaces.DependencyGraphOptions) (string, error) {
	cm := &v1.ConfigMap{}
	cm.Data = sched.Serialize(options)
	cm.GenerateName = "appcontrollerdeployment-"
	cm.Labels = map[string]string{"AppController": "FlowDeployment"}
	task, err := sched.client.ConfigMaps().Create(cm)
	if err != nil {
		return "", err
	}
	return task.Name, nil
}

// Deploy deploys the dependency graph either in-place or by creating deployment task for a standalone process
func Deploy(sched interfaces.Scheduler, options interfaces.DependencyGraphOptions, inplace bool, stopChan <-chan struct{}) (string, error) {
	var task string
	if inplace {
		log.Println("Going to deploy flow:", options.FlowName)
		depGraph, err := sched.BuildDependencyGraph(options)
		if err != nil {
			return "", err
		}
		var ch <-chan struct{}
		if stopChan == nil {
			ch := make(chan struct{})
			defer close(ch)
		} else {
			ch = stopChan
		}
		depGraph.Deploy(ch)
	} else {
		log.Printf("Scheduling deployment of %s flow", options.FlowName)
		var err error
		task, err = sched.CreateDeployment(options)
		if err != nil {
			return "", err
		}
		log.Printf("Scheduled deployment task %s", task)
	}
	log.Println("Done")
	return task, nil
}

// GetStatus returns deployment status
func GetStatus(client client.Interface, selector labels.Selector,
	options interfaces.DependencyGraphOptions) (interfaces.DeploymentStatus, interfaces.DeploymentReport, error) {

	silent := options.Silent
	if options.FlowName == "" {
		options.FlowName = interfaces.DefaultFlowName
	}
	options.ReplicaCount = 0
	options.Silent = true
	options.FixedNumberOfReplicas = false
	options.MinReplicaCount = 0
	options.MaxReplicaCount = 0

	if !silent {
		log.Println("Getting status of flow", options.FlowName)
	}
	sched := New(client, selector, 0)
	graph, err := sched.BuildDependencyGraph(options)
	if err != nil {
		return interfaces.Empty, nil, err
	}
	status, report := graph.GetStatus()
	return status, report, nil
}
