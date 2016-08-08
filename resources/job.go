package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type Job struct {
	Job    *batch.Job
	Client unversioned.JobInterface
}

func jobKey(name string) string {
	return "job/" + name
}

func jobStatus(j unversioned.JobInterface, name string) (string, error) {
	job, err := j.Get(name)
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

func (s Job) Key() string {
	return jobKey(s.Job.Name)
}

func (s Job) Status() (string, error) {
	return jobStatus(s.Client, s.Job.Name)
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
	return jobKey(s.Name)
}

func (s ExistingJob) Status() (string, error) {
	return jobStatus(s.Client, s.Name)
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
