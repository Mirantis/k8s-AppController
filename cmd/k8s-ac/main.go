package main

import (
	"log"
	"os"
	"time"

	"github.com/gluke77/k8s-ac/client"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	kClient "k8s.io/kubernetes/pkg/client/unversioned"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: k8s-ac KUBERNETES_CLUSTER_URL")
	}

	url := os.Args[1]

	c, err := client.GetAppControllerClient(url)
	if err != nil {
		log.Fatal(err)
	}

	depList, err := c.Dependencies().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, dep := range depList.Items {
		log.Println(dep.Parent, "->", dep.Child)
	}

	resList, err := c.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	resources := resList.Items

	log.Println(resources)

	config := &restclient.Config{
		Host: url,
	}

	kc, err := kClient.New(config)
	if err != nil {
		log.Fatal(err)
	}

	pods := kc.Pods(api.NamespaceDefault)

	for _, r := range resources {
		if r.Pod != nil {
			name := r.Pod.Name

			log.Println("Found pod", r.Pod.Name)

			pod, err := pods.Create(r.Pod)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Created pod", pod)
			log.Println("Name", pod.Name)
			log.Println("Status", pod.Status.Phase)

			for i := 0; i < 5; i++ {
				pod, err = pods.Get(name)
				if err != nil {
					log.Fatal(err)
				}

				log.Println("Name", pod.Name)
				log.Println("Status", pod.Status.Phase)

				time.Sleep(time.Second * 5)

				if pod.Status.Phase == "Succeeded" {
					break
				}
			}

		} else if r.Job != nil {
			log.Println("Unsupported resource definition", r)
		} else {
			log.Println("Unsupported resource definition", r)
		}
	}
}
