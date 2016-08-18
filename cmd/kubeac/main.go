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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/labels"
)

func main() {
	labelSelector := flag.String("l", "", "label selector")
	flag.Parse()

	var url string
	if len(flag.Args()) > 0 {
		url = flag.Args()[0]
	}
	if url == "" {
		url = os.Getenv("KUBERNETES_CLUSTER_URL")
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}

	sel, err := labels.Parse(*labelSelector)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Using label selector:", *labelSelector)

	depGraph, err := scheduler.BuildDependencyGraph(c, sel)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Checking for circular dependencies.")
	cycles := scheduler.DetectCycles(*depGraph)
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

	scheduler.Create(*depGraph)

	log.Println("Done")

}
