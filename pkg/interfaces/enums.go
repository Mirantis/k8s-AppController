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

package interfaces

// DeploymentStatus describes possible status of whole deployment process
type DeploymentStatus int

// Possible values for DeploymentStatus
const (
	Empty DeploymentStatus = iota
	Prepared
	Running
	Finished
	TimedOut
)

func (s DeploymentStatus) String() string {
	switch s {
	case Empty:
		return "No dependencies loaded"
	case Prepared:
		return "Deployment not started"
	case Running:
		return "Deployment is running"
	case Finished:
		return "Deployment finished"
	case TimedOut:
		return "Deployment timed out"
	}
	panic("Unreachable")
}

// ScheduledResourceStatus describes possible status of a single resource
type ScheduledResourceStatus int

// Possible values for ScheduledResourceStatus
const (
	Init ScheduledResourceStatus = iota
	Creating
	Ready
	Error
)
