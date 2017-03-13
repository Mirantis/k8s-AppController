// Copyright 2017 Mirantis
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

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/rest"
)

type Flow struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Construction map[string]string `json:"construction,omitempty"`
}

type FlowList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Flow `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type FlowsInterface interface {
	List(opts api.ListOptions) (*FlowList, error)
	Create(*Flow) (*Flow, error)
	Delete(name string, opts *api.DeleteOptions) error
	Get(name string) (result *Flow, err error)
}

type flows struct {
	rc        *rest.RESTClient
	namespace string
}

func newFlows(c rest.Config, ns string) (*flows, error) {
	rc, err := thirdPartyResourceRESTClient(&c)
	if err != nil {
		return nil, err
	}

	return &flows{rc, ns}, nil
}

func (f *flows) List(opts api.ListOptions) (*FlowList, error) {
	resp, err := f.rc.Get().
		Namespace(f.namespace).
		Resource("flows").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &FlowList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *flows) Create(flow *Flow) (result *Flow, err error) {
	result = &Flow{}
	err = f.rc.Post().
		Resource("flows").
		Namespace(f.namespace).
		Body(flow).
		Do().
		Into(result)
	return
}

func (f *flows) Delete(name string, opts *api.DeleteOptions) error {
	return f.rc.Delete().
		Namespace(f.namespace).
		Resource("flows").
		Name(name).
		Body(opts).
		Do().
		Error()
}

func (f *flows) Get(name string) (*Flow, error) {
	data, err := f.rc.Get().
		Namespace(f.namespace).
		Resource("flows").
		Name(name).
		Do().Raw()
	if err != nil {
		return nil, err
	}

	result := &Flow{}

	// temporary solution until serialization is fixed
	json.Unmarshal(data, result)
	return result, nil
}
