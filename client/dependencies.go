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
