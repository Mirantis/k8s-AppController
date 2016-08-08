package kubernetes

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type ClientInterface interface {
	Pods() unversioned.PodInterface
	Jobs() unversioned.JobInterface
	Services() unversioned.ServiceInterface
}

type client struct {
	*unversioned.Client
}

var _ ClientInterface = &client{}

func (c *client) Pods() unversioned.PodInterface {
	return c.Client.Pods(api.NamespaceDefault)
}

func (c *client) Jobs() unversioned.JobInterface {
	return c.Client.Extensions().Jobs(api.NamespaceDefault)
}

func (c *client) Services() unversioned.ServiceInterface {
	return c.Client.Services(api.NamespaceDefault)
}

func Client(url string) (ClientInterface, error) {
	config := &restclient.Config{
		Host: url,
	}

	c, err := unversioned.New(config)
	if err != nil {
		return nil, err
	}

	return &client{Client: c}, nil
}
