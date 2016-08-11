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

package client

import (
	"encoding/json"
	"errors"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/batch"
)

type ResourceDefinitionsInterface interface {
	List(opts api.ListOptions) (*ResourceDefinitionList, error)
	Get(name string) (*ResourceDefinition, error)
}

type resourceDefinition struct {
	r *AppControllerClient
}

func newResourceDefinitions(c *AppControllerClient) *resourceDefinition {
	return &resourceDefinition{c}
}

func (c *resourceDefinition) List(opts api.ListOptions) (result *ResourceDefinitionList, err error) {
	result = &ResourceDefinitionList{}

	url := getUrlWithOptions(c.r.resourceDefinitionsURL, opts)
	resp, err := c.r.Get(url.String())
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New(resp.Status)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(result)

	return
}

func (c *resourceDefinition) Get(name string) (result *ResourceDefinition, err error) {
	result = &ResourceDefinition{}
	resp, err := c.r.Get(c.r.resourceDefinitionsURL.String() + "/" + name)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New(resp.Status)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(result)

	return
}

type ResourceDefinition struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	//TODO: add other object types
	Pod     *api.Pod     `json:"pod,omitempty"`
	Job     *batch.Job   `json:"job,omitempty"`
	Service *api.Service `json:"service,omitempty"`
}

type ResourceDefinitionList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ResourceDefinition `json:"items"`
}
