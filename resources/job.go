// Copyright 2016 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
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

// Key returns job name
func (j Job) Key() string {
	return jobKey(j.Job.Name)
}

// Status returns job status
func (j Job) Status(meta map[string]string) (string, error) {
	return jobStatus(j.Client, j.Job.Name)
}

// Create creates k8s job object
func (j Job) Create() error {
	log.Println("Looking for job", j.Job.Name)
	status, err := j.Status(nil)

	if err == nil {
		log.Printf("Found job %s, status:%s", j.Job.Name, status)
		log.Println("Skipping creation of job", j.Job.Name)
		return nil
	}

	log.Println("Creating job", j.Job.Name)
	j.Job, err = j.Client.Create(j.Job)
	return err
}

// Delete deletes Job from the cluster
func (j Job) Delete() error {
	return j.Client.Delete(j.Job.Name, nil)
}

// NameMatches gets resource definition and a name and checks if
// the Job part of resource definition has matching name.
func (j Job) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Job != nil && def.Job.Name == name
}

// New returns new Job on resource definition
func (j Job) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewJob(def.Job, c.Jobs())
}

// NewExisting returns new ExistingJob based on resource definition
func (j Job) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingJob(name, c.Jobs())
}

func NewJob(job *batch.Job, client unversioned.JobInterface) Job {
	return Job{Job: job, Client: client}
}

type ExistingJob struct {
	Name   string
	Client unversioned.JobInterface
	Job
}

func (s ExistingJob) Key() string {
	return jobKey(s.Name)
}

func (s ExistingJob) Status(meta map[string]string) (string, error) {
	return jobStatus(s.Client, s.Name)
}

func (s ExistingJob) Create() error {
	log.Println("Looking for job", s.Name)
	status, err := s.Status(nil)

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
