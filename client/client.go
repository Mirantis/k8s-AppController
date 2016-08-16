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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type Interface interface {
	Pods() unversioned.PodInterface
	Jobs() unversioned.JobInterface
	Services() unversioned.ServiceInterface
	Dependencies() DependenciesInterface
	ResourceDefinitions() ResourceDefinitionsInterface
}

type client struct {
	*unversioned.Client
	DependenciesInterface
	ResourceDefinitionsInterface
}

var _ Interface = &client{}

func (c client) Dependencies() DependenciesInterface {
	return c.DependenciesInterface
}

func (c client) ResourceDefinitions() ResourceDefinitionsInterface {
	return c.ResourceDefinitionsInterface
}

func (c client) Pods() unversioned.PodInterface {
	return c.Client.Pods(api.NamespaceDefault)
}

func (c client) Jobs() unversioned.JobInterface {
	return c.Client.Extensions().Jobs(api.NamespaceDefault)
}

func (c client) Services() unversioned.ServiceInterface {
	return c.Client.Services(api.NamespaceDefault)
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
	}, nil
}

func New(url string) (Interface, error) {
	var rc *restclient.Config
	var err error

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

	return newForConfig(*rc)
}
