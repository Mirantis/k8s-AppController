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

	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type serviceClient struct {
}

// MakeService creates a service based on its name
func MakeService(name string) *v1.Service {
	var service *v1.Service

	service = &v1.Service{Spec: v1.ServiceSpec{Selector: map[string]string{name: "yes"}}}

	service.Name = name

	return service
}

func (s *serviceClient) List(opts api.ListOptions) (*v1.ServiceList, error) {
	panic("not implemented")
}

func (s *serviceClient) Get(name string) (*v1.Service, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock service %s returned error", name)
	}

	return MakeService(name), nil
}

func (s *serviceClient) Create(srv *v1.Service) (*v1.Service, error) {
	return MakeService(srv.Name), nil
}

func (s *serviceClient) Update(srv *v1.Service) (*v1.Service, error) {
	panic("not implemented")
}

func (s *serviceClient) UpdateStatus(srv *v1.Service) (*v1.Service, error) {
	panic("not implemented")
}

func (s *serviceClient) Delete(name string, opts *api.DeleteOptions) error {
	panic("not implemented")
}

func (s *serviceClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (s *serviceClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (s *serviceClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (s *serviceClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Service, err error) {
	panic("not implemented")
}

func NewServiceClient() corev1.ServiceInterface {
	return &serviceClient{}
}
