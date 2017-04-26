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
	"fmt"
	"log"
	"reflect"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"

	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/pkg/apis/batch/v1"
)

var jobParamFields = []string{
	"Spec.Template.Spec.Containers.Env",
	"Spec.Template.Spec.Containers.Name",
	"Spec.Template.Spec.InitContainers.Env",
	"Spec.Template.Spec.InitContainers.Name",
	"Spec.Template.ObjectMeta",
}

type newJob struct {
	Base
	Job    *v1.Job
	Client batchv1.JobInterface
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
	def.Job = parametrizeResource(def.Job, gc, jobParamFields).(*v1.Job)
	return createNewJob(def, c.Jobs())
}

// NewExisting returns Job controller for existing resource by its name
func (jobTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{BaseResource: existingJob{Name: name, Client: c.Jobs()}}
}

func jobStatus(job *v1.Job) (interfaces.ResourceStatus, error) {
	for _, cond := range job.Status.Conditions {
		if cond.Type == "Complete" && cond.Status == "True" {
			return interfaces.ResourceReady, nil
		}
	}

	return interfaces.ResourceNotReady, nil
}

// Key returns job name
func (j newJob) Key() string {
	return jobKey(j.Job.Name)
}

// Status returns job status
func (j newJob) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	job, err := j.Client.Get(j.Job.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	if !j.EqualToDefinition(job) {
		return interfaces.ResourceWaitingForUpgrade, fmt.Errorf(string(interfaces.ResourceWaitingForUpgrade))
	}

	return jobStatus(job)
}

// EqualToDefinition checks if definition in object is compatible with provided object
func (j newJob) EqualToDefinition(jobiface interface{}) bool {
	job := jobiface.(*v1.Job)

	return reflect.DeepEqual(job.ObjectMeta, j.Job.ObjectMeta) && reflect.DeepEqual(job.Spec, j.Job.Spec)
}

// Create looks for the Job in k8s and creates it if not present
func (j newJob) Create() error {
	if err := checkExistence(j); err != nil {
		log.Println("Creating", j.Key())
		j.Job, err = j.Client.Create(j.Job)
		return err
	}
	return nil
}

// Delete deletes Job from the cluster
func (j newJob) Delete() error {
	return j.Client.Delete(j.Job.Name, nil)
}

func createNewJob(def client.ResourceDefinition, client batchv1.JobInterface) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: newJob{
			Base: Base{
				Definition: def,
				meta:       def.Meta,
			},
			Job:    def.Job,
			Client: client,
		},
	}
}

type existingJob struct {
	Base
	Name   string
	Client batchv1.JobInterface
}

// Key return Job name
func (j existingJob) Key() string {
	return jobKey(j.Name)
}

// Status returns job status
func (j existingJob) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	job, err := j.Client.Get(j.Name)
	if err != nil {
		return interfaces.ResourceError, err
	}

	return jobStatus(job)
}

// Create looks for existing Job and returns error if there is no such Job
func (j existingJob) Create() error {
	return createExistingResource(j)
}

// Delete deletes Job from the cluster
func (j existingJob) Delete() error {
	return j.Client.Delete(j.Name, nil)
}
