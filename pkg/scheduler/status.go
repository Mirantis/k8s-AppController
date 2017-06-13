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

package scheduler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
)

// GetDeploymentStatus returns stat info of the dependency graph which allows to track its progress
func (depGraph dependencyGraph) GetDeploymentStatus() interfaces.DeploymentStatus {
	result := interfaces.DeploymentStatus{Total: len(depGraph.graph)}
	replicas := map[string]bool{}
	resourceGroups := map[string]bool{}

	for _, resource := range depGraph.graph {
		resourceGroups[resource.Resource.Key()] = true
		progress, err := resource.GetProgress()
		if err != nil {
			continue
		}
		resource.RLock()
		if resource.skipped {
			result.Skipped++
			result.Progress++
		} else if resource.error != nil {
			result.Failed++
			result.Progress++
		} else if resource.ignored {
			result.Progress++
		} else if resource.finished {
			result.Finished++
			result.Progress++
		} else {
			result.Progress += progress
		}
		for _, r := range resource.usedInReplicas {
			replicas[r] = true
		}
		resource.RUnlock()
	}
	if result.Total > 0 {
		result.Progress /= float32(result.Total)
	}
	result.TotalGroups = len(resourceGroups)
	result.Replicas = len(replicas)
	return result
}

type sortableNodeList []*scheduledResource

// Len is the number of elements in the node list.
func (lst sortableNodeList) Len() int {
	return len(lst)
}

// Less compares two items in node list
func (lst sortableNodeList) Less(i, j int) bool {
	if lst[i].distance == lst[j].distance {
		return lst[i].Key() < lst[j].Key()
	}
	return lst[i].distance < lst[j].distance
}

// Swap swaps elements in node list
func (lst sortableNodeList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}

var _ sort.Interface = sortableNodeList{}

func (depGraph dependencyGraph) GetNodeStatuses() []interfaces.NodeStatus {
	nodes := make(sortableNodeList, 0, len(depGraph.graph))
	for _, node := range depGraph.graph {
		nodes = append(nodes, node)
	}
	sort.Sort(nodes)

	result := make([]interfaces.NodeStatus, 0, len(nodes))

	for _, node := range nodes {
		nodeStatus := interfaces.NodeStatus{Name: node.Key()}
		progress, err := node.GetProgress()
		node.RLock()

		switch {
		case node.error != nil:
			nodeStatus.Progress = 100
			if node.skipped {
				nodeStatus.Status = "Skipped"
			} else {
				nodeStatus.Status = fmt.Sprintf("Failed: %v", node.error)
			}
		case node.ignored:
			nodeStatus.Status = "Ignored"
			nodeStatus.Progress = 100
		case node.finished:
			nodeStatus.Status = "Finished"
			nodeStatus.Progress = 100
		case node.started:
			if err != nil {
				nodeStatus.Status = fmt.Sprintf("Error: %v", err)
			} else {
				nodeStatus.Status = "In progress"
				nodeStatus.Progress = int(progress * 100)
			}
		default:
			blockedBy := node.getBlockedBy()
			if len(blockedBy) == 0 {
				if err != nil && progress > 0 {
					nodeStatus.Status = "In progress"
					nodeStatus.Progress = int(progress * 100)
				} else {
					nodeStatus.Status = "Starting / in progress"
				}
			} else if len(blockedBy) == len(node.requires) {
				nodeStatus.Status = "Pending"
			} else {
				nodeStatus.Status = fmt.Sprintf("Waiting for %s", strings.Join(blockedBy, ", "))
			}
		}
		node.RUnlock()
		result = append(result, nodeStatus)
	}
	return result
}
