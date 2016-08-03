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
		log.Println("Found pod, status:", status)
		log.Println("Skipping pod creation")
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
	if p.Pod.Status.Phase != "Succeeded" {
		return "not ready", nil
	}
	return "ready", nil
}
