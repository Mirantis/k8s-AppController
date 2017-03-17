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

package client

import (
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
)

type Flow struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specifies (partial) label that is used to identify dependencies that belong to
	// the construction path of the Flow (i.e. Flows can have different paths for construction and destruction).
	// For example, if we have flow->job dependency, if this dependency were to confirm to the Construction label
	// it would mean that creating a job is what the flow does. Otherwise it would mean that the job depends on
	// the the flow (i.e. it won't be created before everything, the flow consists of)
	Construction map[string]string `json:"construction,omitempty"`

	// Exported flows can be triggered by the user (through the CLI) whereas those that are not
	// can only be triggered by other flows (including DEFAULT flow which is exported by-default)
	Exported bool `json:"exported,omitempty"`

	// Parameters that the flow can accept (i.e. valid imports for the flow)
	Parameters map[string]FlowParameter `json:"parameters,omitempty"`
}

type FlowParameter struct {
	// Optional default value for the parameter. If the declared parameter has nil Default then the argument for
	// this parameter becomes mandatory (i.e. it MUST be provided)
	Default *string `json:"default,omitempty"`

	// Description of the parameter (help string)
	Description string `json:"description,omitempty"`
}
