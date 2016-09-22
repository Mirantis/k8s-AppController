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
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
)

func deploy(cmd *cobra.Command, args []string) {
	var err error

	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		log.Fatal(err)
	}

	labelSelector, err := cmd.Flags().GetString("label")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Using concurrency:", concurrency)

	var url string
	if len(args) > 0 {
		url = args[0]
	}
	if url == "" {
		url = os.Getenv("KUBERNETES_CLUSTER_URL")
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

	depGraph, err := scheduler.BuildDependencyGraph(c, sel)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Checking for circular dependencies.")
	cycles := scheduler.DetectCycles(depGraph)
	if len(cycles) > 0 {
		message := "Cycles detected, terminating:\n"
		for _, cycle := range cycles {
			keys := make([]string, 0, len(cycle))
			for _, vertex := range cycle {
				keys = append(keys, vertex.Key())
			}
			message = fmt.Sprintf("%sCycle: %s\n", message, strings.Join(keys, ", "))
		}

		log.Fatal(message)
	} else {
		log.Println("No cycles detected.")
	}

	scheduler.Create(depGraph, concurrency)

	log.Println("Done")

}

var Deploy = &cobra.Command{
	Use:   "deploy",
	Short: "Start deployment of AppController graph",
	Long:  "Start deployment of AppController graph",
	Run:   deploy,
}
