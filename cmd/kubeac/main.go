package main

import (
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	kClient "k8s.io/kubernetes/pkg/client/unversioned"
)

type ScheduledPod struct {
	Pod  *api.Pod
	Pods kClient.PodInterface
}

func (p ScheduledPod) Key() string {
	return "pod/" + p.Pod.Name
}

func (p ScheduledPod) Create() error {
	log.Println("Looking for pod", p.Pod.Name)
	status, err := p.Status()

	if err == nil {
		log.Println("Found pod, status:", status)
		log.Println("Skipping pod creation")
		return nil
	}

	log.Println("Creating pod", p.Pod.Name)
	p.Pod, err = p.Pods.Create(p.Pod)
	return err
}

func (p ScheduledPod) Status() (string, error) {
	pod, err := p.Pods.Get(p.Pod.Name)
	if err != nil {
		return "error", err
	}
	p.Pod = pod
	if p.Pod.Status.Phase != "Succeeded" {
		return "not ready", nil
	}
	return "ready", nil
}

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
	resList, err := c.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	config := &restclient.Config{
		Host: url,
	}

	kc, err := kClient.New(config)
	if err != nil {
		log.Fatal(err)
	}

	pods := kc.Pods(api.NamespaceDefault)

	resources := []scheduler.Resource{}
	for _, r := range resList.Items {
		if r.Pod != nil {
			resources = append(resources,
				ScheduledPod{Pod: r.Pod, Pods: pods})
			log.Println("Found pod definition", r.Pod.Name, r.Pod)
		} else {
			log.Println("Found unsupported resource", r)
		}
	}
	log.Println("ResourceDefinitions found so far", resources)

	log.Println("Starting resource creation")

	scheduler.Create(resources, deps)

	log.Println("Done")

}
