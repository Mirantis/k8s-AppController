package main

import (
	//"fmt"
	//	"reflect"
	//"unsafe"
	"encoding/json"
	"net/http"
	"net/url"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"
	//client "k8s.io/kubernetes/pkg/client/unversioned"
	//"github.com/kr/pretty"
)

type AppControllerClient struct {
	*http.Client
	Root                   *url.URL
	dependenciesURL        *url.URL
	resourceDefinitionsURL *url.URL
}

func (c *AppControllerClient) Dependencies() DependenciesInterface {
	return newDependencies(c)
}

func (c *AppControllerClient) ResourceDefinitions() ResourceDefinitionsInterface {
	return newResourceDefinitions(c)
}

//Create new client for appcontroller resources
func New(c *restclient.Config) (*AppControllerClient, error) {
	client := &http.Client{}

	root, err := url.Parse(c.Host + c.APIPath)
	if err != nil {
		return nil, err
	}

	//these in front of the path are ugly hacks caused by https://github.com/kubernetes/kubernetes/issues/23831
	deps, err := url.Parse(root.String() + "1/" + c.ContentConfig.GroupVersion.Version + "/namespaces/default/dependencies")
	if err != nil {
		return nil, err
	}

	resources, err := url.Parse(root.String() + "2/" + c.ContentConfig.GroupVersion.Version + "/namespaces/default/definitions")
	if err != nil {
		return nil, err
	}
	return &AppControllerClient{Client: client, Root: root, dependenciesURL: deps, resourceDefinitionsURL: resources}, nil
}

//////////////

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

	//TODO: parse opts
	resp, err := c.r.Get(c.r.dependenciesURL.String())
	if err != nil {
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
	err = json.NewDecoder(resp.Body).Decode(result)

	return
}

type Dependency struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Parents  []string `json:"parents"`
	Children []string `json:"children"`
}

type DependencyList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Dependency `json:"items" protobuf:"bytes,2,rep,name=items"`
}

/////////////////
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

	resp, err := c.r.Get(c.r.resourceDefinitionsURL.String())
	if err != nil {
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

	err = json.NewDecoder(resp.Body).Decode(result)

	return
}

type ResourceDefinition struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	//TODO: add other object types
	Pod api.Pod `json:"pod,omitempty"`
}

type ResourceDefinitionList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ResourceDefinition `json:"items"`
}
