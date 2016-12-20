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

package mocks

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/client-go/1.5/kubernetes/typed/batch/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/apis/batch/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

type jobClient struct {
}

func MakeJob(name string) *v1.Job {
	status := strings.Split(name, "-")[0]

	job := &v1.Job{}
	job.Name = name

	if status == "ready" {
		job.Status.Conditions = append(
			job.Status.Conditions,
			v1.JobCondition{Type: "Complete", Status: "True"},
		)
	}

	return job
}

func (j *jobClient) List(opts api.ListOptions) (*v1.JobList, error) {
	var jobs []v1.Job
	for i := 0; i < 3; i++ {
		jobs = append(jobs, *MakeJob(fmt.Sprintf("ready-trolo%d", i)))
	}

	// use ListOptions.LabelSelector to check if there should be any pending jobs
	if strings.Index(opts.LabelSelector.String(), "failedjob=yes") >= 0 {
		for i := 0; i < 2; i++ {
			jobs = append(jobs, *MakeJob(fmt.Sprintf("pending-lolo%d", i)))
		}
	}

	return &v1.JobList{Items: jobs}, nil
}

func (j *jobClient) Get(name string) (*v1.Job, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock job %s returned error", name)
	}

	return MakeJob(name), nil
}

func (j *jobClient) Create(job *v1.Job) (*v1.Job, error) {
	return MakeJob(job.Name), nil
}

func (j *jobClient) Update(job *v1.Job) (*v1.Job, error) {
	panic("not implemented")
}

func (j *jobClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (j *jobClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (j *jobClient) UpdateStatus(job *v1.Job) (*v1.Job, error) {
	panic("not implemented")
}

func (j *jobClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (j *jobClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Job, err error) {
	panic("not implemented")
}

func NewJobClient() batchv1.JobInterface {
	return &jobClient{}
}
