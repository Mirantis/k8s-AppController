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

	// Construction specifies (partial) label that is used to identify dependencies that belong to
	// the construction path of the Flow (i.e. Flows can have different paths for construction and destruction).
	// For example, if we have flow->job dependency and it matches the Construction label it would mean that
	// creating a job is what the flow does. Otherwise it would mean that the job depends on
	// the the flow (i.e. it won't be created before everything, the flow consists of)
	Construction map[string]string `json:"construction,omitempty"`

	// Destruction specifies (partial) label that is used to identify dependencies that belong to the destruction path of the Flow.
	Destruction map[string]string `json:"destruction,omitempty"`

	// Exported flows can be triggered by the user (through the CLI) whereas those that are not
	// can only be triggered by other flows (including DEFAULT flow which is exported by-default)
	Exported bool `json:"exported,omitempty"`

	// Flow replicas must be deployed sequentially, one by one
	Sequential bool `json:"sequential,omitempty"`

	// Parameters that the flow can accept (i.e. valid inputs for the flow)
	Parameters map[string]FlowParameter `json:"parameters,omitempty"`

	// ReplicaSpace name allows to share replicas across flows. Replicas of different flows with the same ReplicaSpace
	// name are treated as replicas of each of those flows. Thus replicas created by one flow can be deleted by another
	// or there might be several different ways to create replicas (e.g. two flows that represent two ways to deploy
	// something, but no matter what flow was used, the result would be one more application instance (cluster node)
	// represented by created replica
	// By default, replica space name equal to the flow name
	ReplicaSpace string `json:"replicaSpace,omitempty"`
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
// those name persistent so that we could do CRUD on the list of replica names
type Replica struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// flow name
	FlowName string `json:"flowName,omitempty"`

	// replica-space name
	ReplicaSpace string `json:"replicaSpace,omitempty"`

	// AC sets this field to true after deployment of the dependency graph for this replica
	Deployed bool `json:"deployed,omitempty"`
}

// ReplicaList is a list of replicas
type ReplicaList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Replica `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// ReplicasInterface is an interface to access flow replica objects
type ReplicasInterface interface {
	List(opts api.ListOptions) (*ReplicaList, error)
	Create(replica *Replica) (*Replica, error)
	Delete(name string) error
	Update(replica *Replica) error
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

// List returns a list of flow replicas stored in k8s
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
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// Create stores new replica object in the k8s
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

// Delete deletes a replica object by its name
func (r *replicas) Delete(name string) error {
	return r.rc.Delete().
		Namespace(r.namespace).
		Resource("replicas").
		Name(name).
		Do().
		Error()
}

// Update persists object changes back to k8s
func (r *replicas) Update(replica *Replica) error {
	_, err := r.rc.Put().
		Namespace(r.namespace).
		Resource("replicas").
		Name(replica.Name).
		Body(replica).
		DoRaw()
	return err
}

// ReplicaName returns replica name (which then goes into $AC_NAME var) encoded in Replica k8s object name
// this is done in order to take advantage of GenerateName field which makes k8s generate
// unique name that cannot collide with other replicas
func (r *Replica) ReplicaName() string {
	return r.Name[1+strings.LastIndex(r.Name, "-"):]
}
