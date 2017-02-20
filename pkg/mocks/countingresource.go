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
	"time"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

// CountingResource is a fake resource that becomes ready after given timeout.
// It also increases the counter when started and decreases it when becomes ready
type CountingResource struct {
	key       string
	status    string
	counter   *CounterWithMemo
	timeout   time.Duration
	startTime time.Time
}

// Key returns a key of the CountingResource
func (c CountingResource) Key() string {
	return c.key
}

// Status returns a status of the CountingResource. It also updates the status
// after provided timeout and decrements counter
func (c *CountingResource) Status(meta map[string]string) (string, error) {
	if time.Since(c.startTime) >= c.timeout && c.status != "ready" {
		c.counter.Dec()
		c.status = "ready"
	}

	return c.status, nil
}

// Create increments counter and sets creation time
func (c *CountingResource) Create() error {
	c.counter.Inc()
	c.startTime = time.Now()
	return nil
}

// Delete does nothing
func (c *CountingResource) Delete() error {
	return nil
}

// Meta returns empty string
func (c *CountingResource) Meta(string) interface{} {
	return nil
}

// NameMatches returns true
func (c *CountingResource) NameMatches(_ client.ResourceDefinition, _ string) bool {
	return true
}

// New returns new fake resource
func (c *CountingResource) New(_ client.ResourceDefinition, _ client.Interface) interfaces.BaseResource {
	return report.SimpleReporter{BaseResource: NewResource("fake", "ready")}
}

// NewExisting returns new existing resource
func (c *CountingResource) NewExisting(name string, _ client.Interface) interfaces.BaseResource {
	return report.SimpleReporter{BaseResource: NewResource(name, "ready")}
}

// StatusIsCacheable is true
func (c *CountingResource) StatusIsCacheable(meta map[string]string) bool {
	return true
}

// NewCountingResource creates new instance of CountingResource
func NewCountingResource(key string, counter *CounterWithMemo, timeout time.Duration) *CountingResource {
	return &CountingResource{
		key:     key,
		status:  "not ready",
		counter: counter,
		timeout: timeout,
	}
}
