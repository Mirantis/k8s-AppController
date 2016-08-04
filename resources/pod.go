package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type Pod struct {
	Pod    *api.Pod
	Client unversioned.PodInterface
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

	if p.Pod.Status.Phase == "Running" && isReady(p.Pod) {
		return "ready", nil
	}

	return "not ready", nil
}

func isReady(pod *api.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

func NewPod(pod *api.Pod, client unversioned.PodInterface) Pod {
	return Pod{Pod: pod, Client: client}
}

type ExistingPod struct {
	Name   string
	Client unversioned.PodInterface
}

func (p ExistingPod) Key() string {
	return "pod/" + p.Name
}

func (p ExistingPod) Create() error {
	log.Println("Looking for pod", p.Name)
	status, err := p.Status()

	if err == nil {
		log.Printf("Found pod %s, status: %s ", p.Name, status)
		log.Println("Skipping creation of pod", p.Name)
		return nil
	}

	log.Fatalf("Pod %s not found", p.Name)
	return errors.New("Pod not found")
}

func (p ExistingPod) Status() (string, error) {
	pod, err := p.Client.Get(p.Name)
	if err != nil {
		return "error", err
	}

	if pod.Status.Phase == "Succeeded" {
		return "ready", nil
	}

	if pod.Status.Phase == "Running" && isReady(pod) {
		return "ready", nil
	}

	return "not ready", nil
}

func NewExistingPod(name string, client unversioned.PodInterface) ExistingPod {
	return ExistingPod{Name: name, Client: client}
}
