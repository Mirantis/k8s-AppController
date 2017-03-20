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
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"
	"github.com/spf13/cobra"

	"k8s.io/client-go/pkg/labels"
)

var argRe, _ = regexp.Compile(`\s*(\w+)\W+(.*)`)

func deploy(cmd *cobra.Command, args []string) {
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		log.Fatal(err)
	}

	anyArgs, err := cmd.Flags().GetBool("undeclared-args")
	if err != nil {
		log.Fatal(err)
	}
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		log.Fatal(err)
	}

	labelSelector, err := getLabelSelector(cmd)
	if err != nil {
		log.Fatal(err)
	}

	flowArgs, err := cmd.Flags().GetStringArray("arg")
	if err != nil {
		log.Fatal(err)
	}
	flowArgMap := map[string]string{}
	for _, val := range flowArgs {
		key, value, err := parseArg(val)
		if err != nil {
			log.Fatal(err)
		}
		flowArgMap[key] = value
	}
	flowName := interfaces.DefaultFlowName

	if len(args) > 1 {
		log.Fatal("Too many command line arguments")
	}
	if len(args) == 1 {
		flowName = args[0]
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}

	sel, err := labels.Parse(labelSelector)
	if err != nil {
		log.Fatal(err)
	}

	replicasString, err := cmd.Flags().GetString("replicas")
	if err != nil {
		log.Fatal(err)
	}
	replicas, err := strconv.Atoi(strings.TrimSpace(replicasString))
	if err != nil || replicas < 0 {
		log.Fatal("Invalid replica count format")
	}

	log.Println("Using concurrency:", concurrency)
	log.Println("Using label selector:", labelSelector)

	sched := scheduler.New(c, sel, concurrency)
	options := interfaces.DependencyGraphOptions{
		FlowName:               flowName,
		ExportedOnly:           true,
		Args:                   flowArgMap,
		AllowUndeclaredArgs:    anyArgs,
		ReplicaCount:           replicas,
		ReplicaCountIsRelative: replicas > 0 && strings.ContainsAny(replicasString, "+"),
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

func parseArg(arg string) (string, string, error) {
	res := argRe.FindStringSubmatch(arg)
	if len(res) != 3 {
		return "", "", fmt.Errorf("invalid argument %s", arg)
	}
	return res[1], strings.TrimSpace(res[2]), nil
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

	run.Flags().StringP("label", "l", "",
		"Label selector. Overrides KUBERNETES_AC_LABEL_SELECTOR env variable in AppController pod.")

	run.Flags().Bool("undeclared-args", false, "Allow undeclared arguments")

	run.Flags().StringArray("arg", []string{}, "Flow arguments (key=value)")

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
	run.Flags().IntP("concurrency", "c", concurrencyDefault, "concurrency")

	run.Flags().String("url", os.Getenv("KUBERNETES_CLUSTER_URL"),
		"URL of the Kubernetes cluster. Overrides KUBERNETES_CLUSTER_URL env variable in AppController pod.")

	run.Flags().StringP("replicas", "n", "1", "number of flow replicas: [+]number")

	return run, err
}
