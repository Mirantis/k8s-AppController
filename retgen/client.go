package client

import (
	"net/http"
	"net/url"

	"k8s.io/kubernetes/pkg/client/restclient"
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
