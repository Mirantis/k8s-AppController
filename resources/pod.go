package resources

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	kClient "k8s.io/kubernetes/pkg/client/unversioned"
)

type Pod struct {
	Pod    *api.Pod
	Client kClient.PodInterface
}

func (p Pod) Key() string {
	return "pod/" + p.Pod.Name
}

func (p Pod) Create() error {
	log.Println("Looking for pod", p.Pod.Name)
	status, err := p.Status()

	if err == nil {
		log.Printf("Found pod %s, status: %s ", p.Pod.Name, status)
		log.Println("Skipping creation of pod", p.Pod.Name)
		return nil
	}

	log.Println("Creating pod", p.Pod.Name)
	p.Pod, err = p.Client.Create(p.Pod)
	return err
}

func (p Pod) Status() (string, error) {
	pod, err := p.Client.Get(p.Pod.Name)
	if err != nil {
		return "error", err
	}
	p.Pod = pod

	if p.Pod.Status.Phase == "Succeeded" {
		return "ready", nil
	}

	if p.Pod.Status.Phase == "Running" && p.isReady() {
		return "ready", nil
	}

	return "not ready", nil
}

func (p Pod) isReady() bool {
	for _, cond := range p.Pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}
