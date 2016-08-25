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
	"bytes"
	"encoding/json"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"
)

type Dependency struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Parent string            `json:"parent"`
	Child  string            `json:"child"`
	Meta   map[string]string `json:"meta,omitempty"`
}

type DependencyList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Dependency `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type DependenciesInterface interface {
	List(opts api.ListOptions) (*DependencyList, error)
}

type dependencies struct {
	rc *restclient.RESTClient
}

func newDependencies(c restclient.Config) (*dependencies, error) {
	c.APIPath = "/apis"
	c.ContentConfig = restclient.ContentConfig{
		GroupVersion: &unversioned.GroupVersion{
			Group:   "appcontroller.k8s1",
			Version: "v1alpha1",
		},
		NegotiatedSerializer: api.Codecs,
	}

	rc, err := restclient.RESTClientFor(&c)
	if err != nil {
		return nil, err
	}

	return &dependencies{rc}, nil
}

func (c dependencies) List(opts api.ListOptions) (*DependencyList, error) {
	resp, err := c.rc.Get().
		Namespace("default").
		Resource("dependencies").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &DependencyList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
