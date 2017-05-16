// Copyright 2017 Mirantis
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

import "github.com/Mirantis/k8s-AppController/pkg/client"

// MakeFlow generates sample Flow resource definition
func MakeFlow(name string) *client.ResourceDefinition {
	flow := &client.Flow{
		Exported:     true,
		Construction: map[string]string{"flow": name},
	}
	flow.Name = name
	resDef := client.ResourceDefinition{}
	resDef.Flow = flow
	resDef.Name = "flow-" + name
	resDef.Namespace = "testing"

	return &resDef
}

// MakeFlowParameter creates stample Flow parameter object
func MakeFlowParameter(defaultValue string) client.FlowParameter {
	return client.FlowParameter{Default: &defaultValue}
}
