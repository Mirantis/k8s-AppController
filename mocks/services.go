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

package mocks

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type serviceClient struct {
}

func MakeService(name string) *api.Service {
	var service *api.Service

	if name == "failed" {
		service = &api.Service{Spec: api.ServiceSpec{Selector: map[string]string{"failed": "yes"}}}
	} else if name == "success" {
		service = &api.Service{Spec: api.ServiceSpec{Selector: map[string]string{"success": "yes"}}}
	} else {
		service = &api.Service{}
	}

	service.Name = name

	return service
}

func (s *serviceClient) List(opts api.ListOptions) (*api.ServiceList, error) {
	panic("not implemented")
}

func (s *serviceClient) Get(name string) (*api.Service, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeService(name), nil
}

func (s *serviceClient) Create(srv *api.Service) (*api.Service, error) {
	return MakeService(srv.Name), nil
}

func (s *serviceClient) Update(srv *api.Service) (*api.Service, error) {
	panic("not implemented")
}

func (s *serviceClient) UpdateStatus(srv *api.Service) (*api.Service, error) {
	panic("not implemented")
}

func (s *serviceClient) Delete(name string) error {
	panic("not implemented")
}

func (s *serviceClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (s *serviceClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

func NewServiceClient() unversioned.ServiceInterface {
	return &serviceClient{}
}
