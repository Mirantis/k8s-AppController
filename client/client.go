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
	"log"
	"os"

	"k8s.io/kubernetes/pkg/api"
	apiUnversioned "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

// Interface is as an interface for k8s clients. It expands native k8s client interface.
type Interface interface {
	Secrets() unversioned.SecretsInterface
	Pods() unversioned.PodInterface
	Jobs() unversioned.JobInterface
	Services() unversioned.ServiceInterface
	ReplicaSets() unversioned.ReplicaSetInterface
	PetSets() unversioned.PetSetInterface
	DaemonSets() unversioned.DaemonSetInterface
	Dependencies() DependenciesInterface
	ResourceDefinitions() ResourceDefinitionsInterface
}

type client struct {
	*unversioned.Client
	DependenciesInterface
	ResourceDefinitionsInterface
	namespace string
}

var _ Interface = &client{}

// Dependencies returns dependency client for ThirdPartyResource created by AppController
func (c client) Dependencies() DependenciesInterface {
	return c.DependenciesInterface
}

// ResourceDefinitions returns resource definition client for ThirdPartyResource created by AppController
func (c client) ResourceDefinitions() ResourceDefinitionsInterface {
	return c.ResourceDefinitionsInterface
}

// Secrets returns K8s Secrets client for ac namespace
func (c client) Secrets() unversioned.SecretsInterface {
	return c.Client.Secrets(c.namespace)
}

// Pods returns K8s Pod client for ac namespace
func (c client) Pods() unversioned.PodInterface {
	return c.Client.Pods(c.namespace)
}

// Jobs returns K8s Job client for ac namespace
func (c client) Jobs() unversioned.JobInterface {
	return c.Client.Extensions().Jobs(c.namespace)
}

// Services returns K8s Service client for ac namespace
func (c client) Services() unversioned.ServiceInterface {
	return c.Client.Services(c.namespace)
}

// ReplicaSets returns K8s ReplicaSet client for ac namespace
func (c client) ReplicaSets() unversioned.ReplicaSetInterface {
	return c.Client.Extensions().ReplicaSets(c.namespace)
}

// PetSets returns K8s PetSet client for ac namespace
func (c client) PetSets() unversioned.PetSetInterface {
	return c.Client.Apps().PetSets(c.namespace)
}

//DaemonSets return K8s DaemonSet client for ac namespace
func (c client) DaemonSets() unversioned.DaemonSetInterface {
	return c.Client.Extensions().DaemonSets(c.namespace)
}

func newForConfig(c restclient.Config) (Interface, error) {
	deps, err := newDependencies(c)
	if err != nil {
		return nil, err
	}
	resdefs, err := newResourceDefinitions(c)
	if err != nil {
		return nil, err
	}
	cl, err := unversioned.New(&c)
	if err != nil {
		return nil, err
	}

	return &client{
		Client:                       cl,
		DependenciesInterface:        deps,
		ResourceDefinitionsInterface: resdefs,
		namespace:                    getNamespace(),
	}, nil
}

func thirdPartyResourceRESTClient(c *restclient.Config) (*restclient.RESTClient, error) {
	c.APIPath = "/apis"
	c.ContentConfig = restclient.ContentConfig{
		GroupVersion: &apiUnversioned.GroupVersion{
			Group:   "appcontroller.k8s",
			Version: "v1alpha1",
		},
		NegotiatedSerializer: api.Codecs,
	}
	rc, err := restclient.RESTClientFor(c)
	return rc, err
}

// GetConfig returns restclient.Config for given URL.
// If url is empty, assume in-cluster config. Otherwise, return config for remote cluster.
func GetConfig(url string) (rc *restclient.Config, err error) {
	if url == "" {
		log.Println("No Kubernetes cluster URL provided. Assume in-cluster.")
		rc, err = restclient.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		rc = &restclient.Config{
			Host: url,
		}
	}
	return
}

// New returns client k8s api server under given url
func New(url string) (Interface, error) {
	rc, err := GetConfig(url)
	if err != nil {
		return nil, err
	}
	return newForConfig(*rc)
}

func getNamespace() string {
	ns := os.Getenv("KUBERNETES_AC_POD_NAMESPACE")
	if ns == "" {
		ns = api.NamespaceDefault
	}
	return ns
}
