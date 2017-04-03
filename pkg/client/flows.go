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

// Flow is a pseudo-resource (i.e. it looks like resource but cannot be created in k8s, but can be wrapped with
// ResourceDefinition), that represents the scope and parameters of the named reusable subgraph.
type Flow struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specifies (partial) label that is used to identify dependencies that belong to
	// the construction path of the Flow (i.e. Flows can have different paths for construction and destruction).
	// For example, if we have flow->job dependency and it matches the Construction label it would mean that
	// creating a job is what the flow does. Otherwise it would mean that the job depends on
	// the the flow (i.e. it won't be created before everything, the flow consists of)
	Construction map[string]string `json:"construction,omitempty"`

	// Specifies (partial) label that is used to identify dependencies that belong to the destruction path of the Flow.
	Destruction map[string]string `json:"destruction,omitempty"`

	// Exported flows can be triggered by the user (through the CLI) whereas those that are not
	// can only be triggered by other flows (including DEFAULT flow which is exported by-default)
	Exported bool `json:"exported,omitempty"`

	// Parameters that the flow can accept (i.e. valid inputs for the flow)
	Parameters map[string]FlowParameter `json:"parameters,omitempty"`
}

// FlowParameter represents declaration of a parameter that flow can accept (its description, default value etc)
type FlowParameter struct {
	// Optional default value for the parameter. If the declared parameter has nil Default then the argument for
	// this parameter becomes mandatory (i.e. it MUST be provided)
	Default *string `json:"default,omitempty"`

	// Description of the parameter (help string)
	Description string `json:"description,omitempty"`
}

// Replica resource represents one instance of the replicated flow. Since the only difference between replicas is their
// $AC_NAME value which is replica name this resource is only needed to generate unique name for each replica and make
// those name persistent so that we could do CRD (CRUD without "U") on the list of replica names
type Replica struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	FlowName string `json:"flowName,omitempty"`
}

// List of replicas
type ReplicaList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Replica `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type ReplicasInterface interface {
	List(opts api.ListOptions) (*ReplicaList, error)
	Create(replica *Replica) (*Replica, error)
	Delete(name string) error
	Get(name string) (result *Replica, err error)
}

type replicas struct {
	rc        *rest.RESTClient
	namespace string
}

func newReplicas(c rest.Config, ns string) (*replicas, error) {
	rc, err := thirdPartyResourceRESTClient(&c)
	if err != nil {
		return nil, err
	}

	return &replicas{rc, ns}, nil
}

func (r *replicas) List(opts api.ListOptions) (*ReplicaList, error) {
	resp, err := r.rc.Get().
		Namespace(r.namespace).
		Resource("replicas").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &ReplicaList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *replicas) Create(replica *Replica) (result *Replica, err error) {
	result = &Replica{}
	data, err := r.rc.Post().
		Resource("replicas").
		Namespace(r.namespace).
		Body(replica).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, result)
	return
}

func (r *replicas) Delete(name string) error {
	return r.rc.Delete().
		Namespace(r.namespace).
		Resource("replicas").
		Name(name).
		Do().
		Error()
}

func (r *replicas) Get(name string) (result *Replica, err error) {
	err = r.rc.Get().
		Namespace(r.namespace).
		Resource("replicas").
		Name(name).
		Do().Into(result)

	return
}

func (r *Replica) UnmarshalJSON(data []byte) error {
	type XReplica Replica
	tmp := XReplica{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*r = Replica(tmp)
	return nil
}

func (rl *ReplicaList) UnmarshalJSON(data []byte) error {
	type XReplicaList ReplicaList
	tmp := XReplicaList{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*rl = ReplicaList(tmp)
	return nil
}

func (f *Flow) UnmarshalJSON(data []byte) error {
	type XFlow Flow
	tmp := XFlow{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*f = Flow(tmp)
	return nil
}

func (r *Replica) ReplicaName() string {
	return r.Name[1+strings.LastIndex(r.Name, "-"):]
}
