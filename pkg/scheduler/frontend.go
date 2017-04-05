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

func New(client client.Interface, selector labels.Selector, concurrency int) interfaces.Scheduler {
	return &Scheduler{
		client:      client,
		selector:    selector,
		concurrency: concurrency,
	}
}

func (sched *Scheduler) Serialize(options interfaces.DependencyGraphOptions) map[string]string {
	result := map[string]string{
		"selector":                     sched.selector.String(),
		"concuurency":                  strconv.Itoa(sched.concurrency),
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

func FromConfig(client client.Interface, cfg map[string]string) (interfaces.Scheduler, interfaces.DependencyGraphOptions, error) {
	var concurrency int
	var opts interfaces.DependencyGraphOptions
	var selector labels.Selector

	parsers := []func() error{
		func() error { return getSelector(cfg, "selector", &selector) },
		func() error { return getInt(cfg, "concuurency", &concurrency) },
		func() error { return getString(cfg, "flowName", &opts.FlowName) },
		func() error { return getBool(cfg, "exportedOnly", &opts.ExportedOnly) },
		func() error { return getBool(cfg, "allowUndeclaredArgs", &opts.AllowDeleteExternalResources) },
		func() error { return getInt(cfg, "replicaCount", &opts.ReplicaCount) },
		func() error { return getBool(cfg, "fixedNumberOfReplicas", &opts.FixedNumberOfReplicas) },
		func() error { return getInt(cfg, "minReplicaCount", &opts.MinReplicaCount) },
		func() error { return getInt(cfg, "maxReplicaCount", &opts.MaxReplicaCount) },
		func() error { return getBool(cfg, "allowDeleteExternalResources", &opts.AllowDeleteExternalResources) },
	}

	for _, parser := range parsers {
		err := parser()
		if err != nil {
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

func (sched *Scheduler) CreateDeployment(options interfaces.DependencyGraphOptions) error {
	cm := &v1.ConfigMap{}
	cm.Data = sched.Serialize(options)
	cm.GenerateName = "appcontrollerdeployment-"
	cm.Labels = map[string]string{"AppController": "FlowDeployment"}
	_, err := sched.client.ConfigMaps().Create(cm)
	return err
}

func Deploy(sched interfaces.Scheduler, options interfaces.DependencyGraphOptions, inplace bool, stopChan <-chan struct{}) error {
	if inplace {
		log.Println("Going to deploy flow:", options.FlowName)
		depGraph, err := sched.BuildDependencyGraph(options)
		if err != nil {
			return err
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
		err := sched.CreateDeployment(options)
		if err != nil {
			return err
		}
	}
	log.Println("Done")
	return nil
}
