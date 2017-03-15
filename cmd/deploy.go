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

package cmd

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"
	"github.com/spf13/cobra"

	"k8s.io/client-go/pkg/labels"
)

func deploy(cmd *cobra.Command, args []string) {
	var err error

	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		log.Fatal(err)
	}

	labelSelector, err := getLabelSelector(cmd)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Using concurrency:", concurrency)

	var url string
	flowName := ""

	for _, arg := range args {
		if strings.HasPrefix(strings.ToLower(arg), "http://") ||
			strings.HasPrefix(strings.ToLower(arg), "https://") {
			if url != "" {
				log.Fatal("Cluster URL specified more than once")
			}
			url = arg
		} else {
			if flowName != "" {
				log.Fatal("Flow name specified more than once")
			}
			flowName = arg
		}
	}
	if url == "" {
		url = os.Getenv("KUBERNETES_CLUSTER_URL")
	}
	if flowName == "" {
		flowName = interfaces.DefaultFlowName
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}

	sel, err := labels.Parse(labelSelector)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Using label selector:", labelSelector)

	sched := scheduler.New(c, sel, concurrency)
	options := interfaces.DependencyGraphOptions{
		FlowName:     flowName,
		ExportedOnly: true,
	}
	log.Println("Going to deploy flow:", flowName)
	depGraph, err := sched.BuildDependencyGraph(options)
	if err != nil {
		log.Fatal(err)
	}

	stopChan := make(chan struct{})
	depGraph.Deploy(stopChan)

	log.Println("Done")
}

func getLabelSelector(cmd *cobra.Command) (string, error) {
	labelSelector, err := cmd.Flags().GetString("label")
	if labelSelector == "" {
		labelSelector = os.Getenv("KUBERNETES_AC_LABEL_SELECTOR")
	}
	return labelSelector, err
}

// InitRunCommand returns cobra command for performing AppController graph deployment
func InitRunCommand() (*cobra.Command, error) {
	run := &cobra.Command{
		Use:   "run",
		Short: "Start deployment of AppController graph",
		Long:  "Start deployment of AppController graph",
		Run:   deploy,
	}

	var labelSelector string
	run.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector. Overrides KUBERNETES_AC_LABEL_SELECTOR env variable in AppController pod.")

	concurrencyString := os.Getenv("KUBERNETES_AC_CONCURRENCY")

	var err error
	var concurrencyDefault int
	if len(concurrencyString) > 0 {
		concurrencyDefault, err = strconv.Atoi(concurrencyString)
		if err != nil {
			log.Printf("KUBERNETES_AC_CONCURRENCY is set to '%s' but it does not look like an integer: %v",
				concurrencyString, err)
			concurrencyDefault = 0
		}
	}
	var concurrency int
	run.Flags().IntVarP(&concurrency, "concurrency", "c", concurrencyDefault, "concurrency")
	return run, err
}
