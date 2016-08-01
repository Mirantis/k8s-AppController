package main

import (
	//"fmt"
	//"net/http"
	//	"reflect"
	//"unsafe"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"
	//client "k8s.io/kubernetes/pkg/client/unversioned"
	//"github.com/kr/pretty"
)

type AppControllerClient struct {
	*restclient.RESTClient
}

func (c *AppControllerClient) Dependencies() DependenciesInterface {
	return newDependencies(c)
}

//func (c *AppControllerClient) ResourceDefinitions() ResourceDefinitionsInterface {
//	return newResourceDefinitions(c)
//}

//Create new client for appcontroller resources
func New(c *restclient.Config) (*AppControllerClient, error) {
	err := restclient.SetKubernetesDefaults(c)
	if err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(c)
	if err != nil {
		return nil, err
	}
	//unversionedClient, err := client.New(c)
	//if err != nil {
	//	return nil, err
	//}

	return &AppControllerClient{RESTClient: client}, nil
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

	//err = c.r.Get().Resource("dependencies").VersionedParams(&opts, codec).Do().Into(result)
	err = c.r.Get().Resource("dependencies").Do().Into(result)

	return
}

func (c *dependency) Get(name string) (result *Dependency, err error) {
	result = &Dependency{}
	err = c.r.Get().Resource("dependencies").Name(name).Do().Into(result)

	return
}

type Dependency struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Parents  []string `json:"parents"`
	Children []string `json:"children"`
}

func (o *Dependency) GetObjectKind() unversioned.ObjectKind {
	return &o.TypeMeta
}

type DependencyList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of ThirdPartyResources.
	Items []Dependency `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (o *DependencyList) GetObjectKind() unversioned.ObjectKind {
	return &o.TypeMeta
}
