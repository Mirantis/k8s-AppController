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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type secretClient struct {
}

func MakeSecret(name string) *api.Secret {

	secret := &api.Secret{}
	secret.Name = name

	return secret
}

func (p *secretClient) List(opts api.ListOptions) (*api.SecretList, error) {
	var secrets []api.Secret
	for i := 0; i < 3; i++ {
		secrets = append(secrets, *MakeSecret(fmt.Sprintf("cfgmap-%d", i)))
	}

	return &api.SecretList{Items: secrets}, nil
}

func (p *secretClient) Get(name string) (*api.Secret, error) {
	if name != "fail" {
		return MakeSecret(name), nil
	} else {
		return MakeSecret(name), fmt.Errorf("Mock Secret not created")
	}
}

func (p *secretClient) Delete(name string) error {
	panic("not implemented")
}

func (p *secretClient) Create(secret *api.Secret) (*api.Secret, error) {
	return MakeSecret(secret.Name), nil
}

func (p *secretClient) Update(secret *api.Secret) (*api.Secret, error) {
	panic("not implemented")
}

func (p *secretClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (p *secretClient) Bind(binding *api.Binding) error {
	panic("not implemented")
}

func NewSecretClient() unversioned.SecretsInterface {
	return &secretClient{}
}
