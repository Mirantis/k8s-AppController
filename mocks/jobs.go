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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type jobClient struct {
}

func MakeJob(name string) *batch.Job {
	status := strings.Split(name, "-")[0]

	job := &batch.Job{}
	job.Name = name

	if status == "ready" {
		job.Status.Conditions = append(
			job.Status.Conditions,
			batch.JobCondition{Type: "Complete", Status: "True"},
		)
	}

	return job
}

func (j *jobClient) List(opts api.ListOptions) (*batch.JobList, error) {
	panic("not implemented")
}

func (j *jobClient) Get(name string) (*batch.Job, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock job %s returned error", name)
	}

	return MakeJob(name), nil
}

func (j *jobClient) Create(job *batch.Job) (*batch.Job, error) {
	return MakeJob(job.Name), nil
}

func (j *jobClient) Update(job *batch.Job) (*batch.Job, error) {
	panic("not implemented")
}

func (j *jobClient) Delete(name string, options *api.DeleteOptions) error {
	panic("not implemented")
}

func (j *jobClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (j *jobClient) UpdateStatus(job *batch.Job) (*batch.Job, error) {
	panic("not implemented")
}

func NewJobClient() unversioned.JobInterface {
	return &jobClient{}
}
