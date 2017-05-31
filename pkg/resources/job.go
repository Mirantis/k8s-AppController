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
	"log"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/pkg/apis/batch/v1"
)

var jobParamFields = []string{
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.Containers.name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.Spec.InitContainers.name",
	"Spec.Template.ObjectMeta",
}

type newJob struct {
	job    *v1.Job
	client batchv1.JobInterface
}

type existingJob struct {
	name   string
	client batchv1.JobInterface
}

func jobKey(name string) string {
	return "job/" + name
}

type jobTemplateFactory struct{}

// ShortName returns wrapped resource name if it was a job
func (jobTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Job == nil {
		return ""
	}
	return definition.Job.Name
}

// Kind returns a k8s resource kind that this fabric supports
func (jobTemplateFactory) Kind() string {
	return "job"
}

// New returns Job controller for new resource based on resource definition
func (jobTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	job := parametrizeResource(def.Job, gc, jobParamFields).(*v1.Job)
	return createNewJob(job, c.Jobs())
}

// NewExisting returns Job controller for existing resource by its name
func (jobTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return existingJob{name: name, client: c.Jobs()}
}

func jobProgress(j batchv1.JobInterface, name string) (float32, error) {
	job, err := j.Get(name)
	if err != nil {
		return 0, err
	}

	for _, cond := range job.Status.Conditions {
		if cond.Type == "Complete" && cond.Status == "True" {
			return 1, nil
		}
	}

	return 0, nil
}

// Key returns job name
func (j newJob) Key() string {
	return jobKey(j.job.Name)
}

// GetProgress returns job deployment progress
func (j newJob) GetProgress() (float32, error) {
	return jobProgress(j.client, j.job.Name)
}

// Create looks for the Job in k8s and creates it if not present
func (j newJob) Create() error {
	if checkExistence(j) {
		return nil
	}
	log.Println("Creating", j.Key())
	obj, err := j.client.Create(j.job)
	j.job = obj
	return err
}

// Delete deletes Job from the cluster
func (j newJob) Delete() error {
	return j.client.Delete(j.job.Name, nil)
}

func createNewJob(job *v1.Job, client batchv1.JobInterface) interfaces.Resource {
	return newJob{job: job, client: client}
}

// Key return Job name
func (j existingJob) Key() string {
	return jobKey(j.name)
}

// GetProgress returns job deployment progress
func (j existingJob) GetProgress() (float32, error) {
	return jobProgress(j.client, j.name)
}

// Create looks for existing Job and returns error if there is no such Job
func (j existingJob) Create() error {
	return createExistingResource(j)
}

// Delete deletes Job from the cluster
func (j existingJob) Delete() error {
	return j.client.Delete(j.name, nil)
}
