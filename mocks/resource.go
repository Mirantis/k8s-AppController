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

//Resource is a fake resource
type Resource struct {
	key    string
	status string
}

//Key returns a key of the Resource
func (c Resource) Key() string {
	return c.key
}

//Status returns a status of the Resource
func (c *Resource) Status(meta map[string]string) (string, error) {
	return c.status, nil
}

//Create increments counter and sets creation time
func (c *Resource) Create() error {
	return nil
}

//Update Meta
func (c *Resource) UpdateMeta(m map[string]string) error {
	return nil
}

//NewResource creates new instance of Resource
func NewResource(key string, status string) *Resource {
	return &Resource{
		key:    key,
		status: status,
	}
}
