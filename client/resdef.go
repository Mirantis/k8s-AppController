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
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
)

type ResourceDefinition struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	//TODO: add other object types
	Pod        *api.Pod               `json:"pod,omitempty"`
	Job        *batch.Job             `json:"job,omitempty"`
	Service    *api.Service           `json:"service,omitempty"`
	ReplicaSet *extensions.ReplicaSet `json:"replicaset,omitempty"`
	PetSet     *apps.PetSet           `json:"petset,omitempty"`
	DaemonSet  *extensions.DaemonSet  `json:"daemonset,omitempty"`
	Secret     *api.Secret            `json:"secret,omitempty"`
}

type ResourceDefinitionList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ResourceDefinition `json:"items"`
}

type ResourceDefinitionsInterface interface {
	List(opts api.ListOptions) (*ResourceDefinitionList, error)
}

type resourceDefinitions struct {
	rc *restclient.RESTClient
}

func newResourceDefinitions(c restclient.Config) (*resourceDefinitions, error) {
	rc, err := thirdPartyResourceRESTClient(&c)
	if err != nil {
		return nil, err
	}

	return &resourceDefinitions{rc}, nil
}

func (c *resourceDefinitions) List(opts api.ListOptions) (*ResourceDefinitionList, error) {
	resp, err := c.rc.Get().
		Namespace("default").
		Resource("definitions").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &ResourceDefinitionList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
