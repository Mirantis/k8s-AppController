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

	//TODO: parse opts
	resp, err := c.r.Get(c.r.resourceDefinitionsURL.String())
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
	Pod api.Pod   `json:"pod,omitempty"`
	Job batch.Job `json:"job,omitempty"`
}

type ResourceDefinitionList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ResourceDefinition `json:"items"`
}
