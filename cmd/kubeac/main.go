package main

import (
	"log"
	"os"
	"sort"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/restclient"
	kClient "k8s.io/kubernetes/pkg/client/unversioned"
)

type ScheduledPod struct {
	Pod    *api.Pod
	Client kClient.PodInterface
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
	p.Pod, err = p.Client.Create(p.Pod)
	return err
}

func (p ScheduledPod) Status() (string, error) {
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

type byLastProbeTime []batch.JobCondition

func (b byLastProbeTime) Len() int {
	return len(b)
}

func (b byLastProbeTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byLastProbeTime) Less(i, j int) bool {
	return b[i].LastProbeTime.After(b[j].LastProbeTime.Time)
}

type ScheduledJob struct {
	Job    *batch.Job
	Client kClient.JobInterface
}

func (s ScheduledJob) Key() string {
	return "job/" + s.Job.Name
}

func (s ScheduledJob) Status() (string, error) {
	job, err := s.Client.Get(s.Job.Name)
	if err != nil {
		return "error", err
	}
	s.Job = job
	conds := job.Status.Conditions
	sort.Sort(byLastProbeTime(conds))

	if len(conds) == 0 {
		return "not ready", nil
	}

	if conds[0].Type != "Complete" || conds[0].Status != "True" {
		return "not ready", nil
	}

	return "ready", nil
}

func (s ScheduledJob) Create() error {
	log.Println("Looking for job", s.Job.Name)
	status, err := s.Status()

	if err == nil {
		log.Println("Found job, status:", status)
		log.Println("Skipping job creation")
		return nil
	}

	log.Println("Creating job", s.Job.Name)
	s.Job, err = s.Client.Create(s.Job)
	return err
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
	jobs := kc.Extensions().Jobs(api.NamespaceDefault)

	resources := []scheduler.Resource{}
	for _, r := range resList.Items {
		if r.Pod != nil {
			resources = append(resources,
				ScheduledPod{Pod: r.Pod, Client: pods})
			log.Println("Found pod definition", r.Pod.Name, r.Pod)
		} else if r.Job != nil {
			resources = append(resources,
				ScheduledJob{Job: r.Job, Client: jobs})
			log.Println("Found job definition", r.Job.Name, r.Job)
		} else {
			log.Println("Found unsupported resource", r)
		}
	}
	log.Println("ResourceDefinitions found so far", resources)

	log.Println("Starting resource creation")

	scheduler.Create(resources, deps)

	log.Println("Done")

}
