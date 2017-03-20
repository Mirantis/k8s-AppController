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
	"strings"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/rest"
)

type Flow struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specifies (partial) label that is used to identify dependencies that belong to
	// the construction path of the Flow (i.e. Flows can have different paths for construction and destruction).
	// For example, if we have flow->job dependency, if this dependency were to confirm to the Construction label
	// it would mean that creating a job is what the flow does. Otherwise it would mean that the job depends on
	// the the flow (i.e. it won't be created before everything, the flow consists of)
	Construction map[string]string `json:"construction,omitempty"`

	// Exported flows can be triggered by the user (through the CLI) whereas those that are not
	// can only be triggered by other flows (including DEFAULT flow which is exported by-default)
	Exported bool `json:"exported,omitempty"`

	// Parameters that the flow can accept (i.e. valid inputs for the flow)
	Parameters map[string]FlowParameter `json:"parameters,omitempty"`

	// How many times the flow subgraph should be replicated. <= 0 means 1 flow replica without AC_NAME assigned
	ReplicaCount int
}

type FlowParameter struct {
	// Optional default value for the parameter. If the declared parameter has nil Default then the argument for
	// this parameter becomes mandatory (i.e. it MUST be provided)
	Default *string `json:"default,omitempty"`

	// Description of the parameter (help string)
	Description string `json:"description,omitempty"`
}

type FlowDeployment struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	FlowName string `json:"flowName,omitempty"`
}

type FlowDeploymentList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []FlowDeployment `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type FlowDeploymentsInterface interface {
	List(opts api.ListOptions) (*FlowDeploymentList, error)
	Create(flowDeployment *FlowDeployment) (*FlowDeployment, error)
	Delete(name string, opts *api.DeleteOptions) error
	Get(name string) (result *FlowDeployment, err error)
}

type flowdeployments struct {
	rc        *rest.RESTClient
	namespace string
}

func newFlowDeployments(c rest.Config, ns string) (*flowdeployments, error) {
	rc, err := thirdPartyResourceRESTClient(&c)
	if err != nil {
		return nil, err
	}

	return &flowdeployments{rc, ns}, nil
}

func (f *flowdeployments) List(opts api.ListOptions) (*FlowDeploymentList, error) {
	resp, err := f.rc.Get().
		Namespace(f.namespace).
		Resource("deployments").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &FlowDeploymentList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *flowdeployments) Create(flowDeployment *FlowDeployment) (result *FlowDeployment, err error) {
	result = &FlowDeployment{}
	data, err := f.rc.Post().
		Resource("deployments").
		Namespace(f.namespace).
		Body(flowDeployment).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, result)
	return
}

func (f *flowdeployments) Delete(name string, opts *api.DeleteOptions) error {
	return f.rc.Delete().
		Namespace(f.namespace).
		Resource("deployments").
		Name(name).
		Body(opts).
		Do().
		Error()
}

func (f *flowdeployments) Get(name string) (result *FlowDeployment, err error) {
	err = f.rc.Get().
		Namespace(f.namespace).
		Resource("deployments").
		Name(name).
		Do().Into(result)

	return
}

func (f *FlowDeployment) UnmarshalJSON(data []byte) error {
	type FlowDeploymentCopy FlowDeployment
	tmp := FlowDeploymentCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := FlowDeployment(tmp)
	*f = tmp2
	return nil
}

func (fl *FlowDeploymentList) UnmarshalJSON(data []byte) error {
	type FlowDeploymentListCopy FlowDeploymentList
	tmp := FlowDeploymentListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := FlowDeploymentList(tmp)
	*fl = tmp2
	return nil
}

func (f *FlowDeployment) ReplicaName() string {
	return f.Name[1+strings.LastIndex(f.Name, "-"):]
}
