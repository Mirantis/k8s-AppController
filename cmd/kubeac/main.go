package main

import (
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/kubernetes"
	"github.com/Mirantis/k8s-AppController/resources"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/api"
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

	log.Println("Getting dependencies")
	depList, err := c.Dependencies().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	deps := []scheduler.Dependency{}
	for _, dep := range depList.Items {
		log.Println("Found dependency", dep.Parent, "->", dep.Child)
		deps = append(deps, scheduler.Dependency{Parent: dep.Parent, Child: dep.Child})
	}
	log.Println("Dependencies found so far", deps)

	log.Println("Getting resource definitions")
	resDefList, err := c.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	kc, err := kubernetes.Client(url)
	if err != nil {
		log.Fatal(err)
	}

	pods := kc.Pods()
	jobs := kc.Jobs()

	resList := []scheduler.Resource{}
	for _, r := range resDefList.Items {
		if r.Pod != nil {
			resList = append(resList,
				resources.NewPod(r.Pod, pods))
			log.Println("Found pod definition", r.Pod.Name, r.Pod)
		} else if r.Job != nil {
			resList = append(resList,
				resources.NewJob(r.Job, jobs))
			log.Println("Found job definition", r.Job.Name, r.Job)
		} else {
			log.Println("Found unsupported resource", r)
		}
	}
	log.Println("ResourceDefinitions found so far", resList)

	log.Println("Starting resource creation")

	scheduler.Create(resList, deps)

	log.Println("Done")

}
