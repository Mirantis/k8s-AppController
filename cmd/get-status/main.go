package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/labels"
)

func main() {
	var err error
	concurrencyString := os.Getenv("KUBERNETES_AC_CONCURRENCY")
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
	flag.IntVar(&concurrency, "c", concurrencyDefault, "concurrency")

	var labelSelector string
	flag.StringVar(&labelSelector, "l", "", "label selector")

	flag.Parse()
	url := os.Getenv("KUBERNETES_CLUSTER_URL")
	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}
	sel, err := labels.Parse(labelSelector)
	if err != nil {
		log.Fatal(err)
	}
	graph, err := scheduler.BuildDependencyGraph(c, sel)
	if err != nil {
		log.Fatal(err)
	}
	status, report, _ := graph.GetStatus()
	fmt.Printf("STATUS: %s\nREPORT:\n%s", status, report)
}
