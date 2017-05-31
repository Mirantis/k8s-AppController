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

// DefaultFlowName is the name of default flow (main dependency graph)
const DefaultFlowName = "DEFAULT"

// Resource is an interface for AppController supported resources
type Resource interface {
	Key() string
	GetProgress() (float32, error)
	Create() error
	Delete() error
}

// ResourceTemplate is an interface for AppController supported resource templates
type ResourceTemplate interface {
	Kind() string
	ShortName(client.ResourceDefinition) string
	New(client.ResourceDefinition, client.Interface, GraphContext) Resource
	NewExisting(string, client.Interface, GraphContext) Resource
}

// DeploymentStatus is the structure containing deployment status - different stats and progress info
type DeploymentStatus struct {
	Total       int
	TotalGroups int
	Failed      int
	Skipped     int
	Finished    int
	Replicas    int
	Progress    float32
}

// NodeStatus represents status of each graph node
type NodeStatus struct {
	Name     string
	Status   string
	Progress int
}

// DependencyGraph represents operations on dependency graph
type DependencyGraph interface {
	GetDeploymentStatus() DeploymentStatus
	GetNodeStatuses() []NodeStatus
	Deploy(<-chan struct{}) bool
	Options() DependencyGraphOptions
}

// GraphContext represents context of dependency graph. Resource factories get implementation of this interface
// so that they can get graph, they created in, its options, arguments and call Scheduler API
type GraphContext interface {
	Scheduler() Scheduler
	GetArg(string) string
	Graph() DependencyGraph
}

// DependencyGraphOptions contains all the input required to build a dependency graph
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

// Scheduler interface is an API to build dependency graphs and manage their settings
type Scheduler interface {
	BuildDependencyGraph(options DependencyGraphOptions) (DependencyGraph, error)
	Serialize(options DependencyGraphOptions) map[string]string
	CreateDeployment(options DependencyGraphOptions) (string, error)
}
