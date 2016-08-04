package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

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

type Job struct {
	Job    *batch.Job
	Client unversioned.JobInterface
}

func (s Job) Key() string {
	return "job/" + s.Job.Name
}

func (s Job) Status() (string, error) {
	job, err := s.Client.Get(s.Job.Name)
	if err != nil {
		return "error", err
	}
	s.Job = job

	for _, cond := range s.Job.Status.Conditions {
		if cond.Type == "Complete" && cond.Status == "True" {
			return "ready", nil
		}
	}

	return "not ready", nil
}

func (s Job) Create() error {
	log.Println("Looking for job", s.Job.Name)
	status, err := s.Status()

	if err == nil {
		log.Printf("Found job %s, status:%s", s.Job.Name, status)
		log.Println("Skipping creation of job", s.Job.Name)
		return nil
	}

	log.Println("Creating job", s.Job.Name)
	s.Job, err = s.Client.Create(s.Job)
	return err
}

func NewJob(job *batch.Job, client unversioned.JobInterface) Job {
	return Job{Job: job, Client: client}
}

type ExistingJob struct {
	Name   string
	Client unversioned.JobInterface
}

func (s ExistingJob) Key() string {
	return "job/" + s.Name
}

func (s ExistingJob) Status() (string, error) {
	job, err := s.Client.Get(s.Name)
	if err != nil {
		return "error", err
	}

	for _, cond := range job.Status.Conditions {
		if cond.Type == "Complete" && cond.Status == "True" {
			return "ready", nil
		}
	}

	return "not ready", nil
}

func (s ExistingJob) Create() error {
	log.Println("Looking for job", s.Name)
	status, err := s.Status()

	if err == nil {
		log.Printf("Found job %s, status:%s", s.Name, status)
		return nil
	}

	log.Fatalf("Job %s not found", s.Name)
	return errors.New("Job not found")
}

func NewExistingJob(name string, client unversioned.JobInterface) ExistingJob {
	return ExistingJob{Name: name, Client: client}
}
