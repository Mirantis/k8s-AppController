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
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
)

// Resource is a fake resource
type Resource struct {
	key      string
	progress float32
}

// Key returns a key of the Resource
func (c Resource) Key() string {
	return c.key
}

// GetProgress returns progress of the Resource
func (c *Resource) GetProgress() (float32, error) {
	return c.progress, nil
}

// Create does nothing
func (c *Resource) Create() error {
	return nil
}

// Delete does nothing
func (c *Resource) Delete() error {
	return nil
}

// New returns new fake resource
func (c *Resource) New(_ client.ResourceDefinition, _ client.Interface) interfaces.Resource {
	return NewResource("fake", 1)
}

// NewExisting returns new existing resource
func (c *Resource) NewExisting(name string, _ client.Interface) interfaces.Resource {
	return NewResource(name, 1)
}

// NewResource creates new instance of Resource
func NewResource(key string, progress float32) *Resource {
	return &Resource{key: key, progress: progress}
}
