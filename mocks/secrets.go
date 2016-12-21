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

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/watch"
)

type secretClient struct {
}

func MakeSecret(name string) *v1.Secret {

	secret := &v1.Secret{}
	secret.Name = name

	return secret
}

func (p *secretClient) List(opts v1.ListOptions) (*v1.SecretList, error) {
	var secrets []v1.Secret
	for i := 0; i < 3; i++ {
		secrets = append(secrets, *MakeSecret(fmt.Sprintf("cfgmap-%d", i)))
	}

	return &v1.SecretList{Items: secrets}, nil
}

func (p *secretClient) Get(name string) (*v1.Secret, error) {
	if name != "fail" {
		return MakeSecret(name), nil
	}
	return MakeSecret(name), fmt.Errorf("Mock Secret not created")
}

func (p *secretClient) Delete(name string, opts *v1.DeleteOptions) error {
	panic("not implemented")
}

func (p *secretClient) Create(secret *v1.Secret) (*v1.Secret, error) {
	return MakeSecret(secret.Name), nil
}

func (p *secretClient) Update(secret *v1.Secret) (*v1.Secret, error) {
	panic("not implemented")
}
func (p *secretClient) Watch(opts v1.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *secretClient) Bind(binding *api.Binding) error {
	panic("not implemented")
}

func (p *secretClient) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	panic("not implemented")
}

func (p *secretClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error) {
	panic("not implemented")
}

func NewSecretClient() corev1.SecretInterface {
	return &secretClient{}
}
