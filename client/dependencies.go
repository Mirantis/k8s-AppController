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
)

type DependenciesInterface interface {
	List(opts api.ListOptions) (*DependencyList, error)
	Get(name string) (*Dependency, error)
}

type dependency struct {
	r *AppControllerClient
}

func newDependencies(c *AppControllerClient) *dependency {
	return &dependency{c}
}

func (c *dependency) List(opts api.ListOptions) (result *DependencyList, err error) {
	result = &DependencyList{}

	url := getUrlWithOptions(c.r.dependenciesURL, opts)
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

func (c *dependency) Get(name string) (result *Dependency, err error) {
	result = &Dependency{}
	resp, err := c.r.Get(c.r.dependenciesURL.String() + "/" + name)
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

type Dependency struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Parent string `json:"parent"`
	Child  string `json:"child"`
}

type DependencyList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Dependency `json:"items" protobuf:"bytes,2,rep,name=items"`
}
