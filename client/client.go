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
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/labels"
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

func getUrlWithOptions(baseURL *url.URL, opts api.ListOptions) *url.URL {
	params := url.Values{}
	//TODO: check other selectors than label
	if opts.LabelSelector != nil {
		params.Add("labelSelector", opts.LabelSelector.String())
	}

	finalUrl := *baseURL
	finalUrl.RawQuery = params.Encode()

	return &finalUrl
}

//implements labels.Selector, but supports only equality
type AppControllerLabelSelector struct {
	Key   string
	Value string
}

func (s AppControllerLabelSelector) Empty() bool {
	return len(s.Key) == 0
}

func (s AppControllerLabelSelector) String() string {
	if s.Empty() {
		return ""
	}
	return fmt.Sprintf("%s=%s", s.Key, s.Value)
}

//not supported, returns copy
func (s AppControllerLabelSelector) Add(r ...labels.Requirement) labels.Selector {
	return s
}

//not supported
func (s AppControllerLabelSelector) Matches(labels.Labels) bool {
	return false
}
