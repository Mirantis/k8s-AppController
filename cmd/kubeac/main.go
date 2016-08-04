package main

import (
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/api"
)

func main() {
	var url string
	fromEnv := os.Getenv("KUBERNETES_CLUSTER_URL")
	if fromEnv != "" {
		url = fromEnv
	} else if len(os.Args) != 2 {
		log.Fatal("Usage: k8s-ac KUBERNETES_CLUSTER_URL")
	} else {
		url = os.Args[1]
	}

	c, err := client.GetAppControllerClient(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting resource definitions")
	resDefList, err := c.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Getting dependencies")
	depList, err := c.Dependencies().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	scheduler.Create(url, resDefList, depList)

	log.Println("Done")

}
