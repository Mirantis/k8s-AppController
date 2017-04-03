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
	"github.com/Mirantis/k8s-AppController/pkg/client"
)

type ResourceStatus string

const (
	ResourceReady    ResourceStatus = "ready"
	ResourceNotReady ResourceStatus = "not ready"
	ResourceError    ResourceStatus = "error"
)

const DefaultFlowName = "DEFAULT"

// BaseResource is an interface for AppController supported resources
type BaseResource interface {
	Key() string
	// Ensure that Status() supports nil as meta
	Status(meta map[string]string) (ResourceStatus, error)
	Create() error
	Delete() error
	Meta(string) interface{}
	StatusIsCacheable(meta map[string]string) bool
}

// DependencyReport is a report of a single dependency of a node in graph
type DependencyReport struct {
	Dependency string
	Blocks     bool
	Percentage int
	Needed     int
	Message    string
}

// Resource is an interface for a base resource that implements getting dependency reports
type Resource interface {
	BaseResource
	GetDependencyReport(map[string]string) DependencyReport
}

// ResourceTemplate is an interface for AppController supported resource templates
type ResourceTemplate interface {
	Kind() string
	ShortName(client.ResourceDefinition) string
	New(client.ResourceDefinition, client.Interface, GraphContext) Resource
	NewExisting(string, client.Interface, GraphContext) Resource
}

type DeploymentReport interface {
	AsText(int) []string
}

type DependencyGraph interface {
	GetStatus() (DeploymentStatus, DeploymentReport)
	Deploy(<-chan struct{})
}

type GraphContext interface {
	Scheduler() Scheduler
	GetArg(string) string
	Graph() DependencyGraph
	Dependency() client.Dependency
}

type DependencyGraphOptions struct {
	FlowName                     string
	Args                         map[string]string
	ExportedOnly                 bool
	AllowUndeclaredArgs          bool
	ReplicaCount                 int
	FixedNumberOfReplicas        bool
	MinReplicaCount              int
	MaxReplicaCount              int
	AllowDeleteExternalResources bool
	FlowInstanceName             string
	Silent                       bool
}

type Scheduler interface {
	BuildDependencyGraph(options DependencyGraphOptions) (DependencyGraph, error)
}
