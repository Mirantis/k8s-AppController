package main

import (
	"flag"
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/api"
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
	if url == "" {
		log.Fatal("Usage: kubeac [-l <label>=<value>[,<label>=<value]...] KUBERNETES_CLUSTER_URL")
	}

	sel, err := labels.Parse(*labelSelector)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Using label selector:", *labelSelector)

	c, err := client.GetAppControllerClient(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting resource definitions")
	resDefList, err := c.ResourceDefinitions().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting dependencies")
	depList, err := c.Dependencies().List(api.ListOptions{LabelSelector: sel})
	if err != nil {
		log.Fatal(err)
	}

	scheduler.Create(url, resDefList, depList)

	log.Println("Done")

}
