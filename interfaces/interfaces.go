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

package interfaces

import (
	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/report"
)

//Resource is an interface for AppController supported resources
type Resource interface {
	Key() string
	// Ensure that Status() supports nil as meta
	Status(meta map[string]string) (string, error)
	Create() error
	NameMatches(client.ResourceDefinition, string) bool
	New(client.ResourceDefinition, client.Interface) Resource
	NewExisting(string, client.Interface) Resource
	GetDependencyReport(meta map[string]string) report.DependencyReport
}
