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
	"container/list"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
)

// detectCycles implements Kosaraju's algorithm https://en.wikipedia.org/wiki/Kosaraju%27s_algorithm
// for detecting cycles in graph.
// We are depending on the fact that any strongly connected component of a graph is a cycle
// if it consists of more than one vertex
func detectCycles(dependencies []client.Dependency) [][]string {
	graph := groupDependentResources(dependencies)

	// is vertex visited in first phase of the algorithm
	visited := make(map[string]bool)
	// is vertex assigned to strongly connected component
	assigned := make(map[string]bool)

	// each key is root of strongly connected component
	// the slice consists of all vertices belonging to strongly connected component to which root belongs
	components := make(map[string][]string)

	orderedVertices := list.New()

	for key := range graph {
		visited[key] = false
		assigned[key] = false
	}

	for key := range graph {
		visitVertex(key, graph, visited, orderedVertices)
	}

	for e := orderedVertices.Front(); e != nil; e = e.Next() {
		vertex := e.Value.(string)
		assignVertex(vertex, vertex, assigned, components, graph)
	}

	// if any strongly connected component consist of more than one vertex - it's a cycle
	var cycles [][]string
	for _, component := range components {
		if len(component) > 1 {
			cycles = append(cycles, component)
		}
	}

	// detect self cycles - not part of Kosaraju's algorithm
	for key, vertex := range graph {
		for _, child := range vertex {
			if key == child {
				cycles = append(cycles, []string{key, key})
			}
		}
	}
	return cycles
}

func groupDependentResources(dependencies []client.Dependency) (result map[string][]string) {
	result = map[string][]string{}
	for _, dependency := range dependencies {
		group := result[dependency.Parent]
		if group == nil {
			group = []string{dependency.Child}
		} else {
			group = insertUnique(group, dependency.Child)
		}
		result[dependency.Parent] = group
		if _, ok := result[dependency.Child]; !ok {
			result[dependency.Child] = []string{}
		}
	}
	return
}

func insertUnique(slice []string, value string) []string {
	index := sort.SearchStrings(slice, value)
	if index >= len(slice) || slice[index] != value {
		result := append(slice, "")
		copy(result[index+1:], slice[index:])
		result[index] = value
		return result
	}
	return slice
}

func visitVertex(vertex string, dependencies map[string][]string, visited map[string]bool, orderedVertices *list.List) {
	if !visited[vertex] {
		visited[vertex] = true
		for _, v := range dependencies[vertex] {
			visitVertex(v, dependencies, visited, orderedVertices)
		}
		orderedVertices.PushBack(vertex)
	}
}

func assignVertex(vertex, root string, assigned map[string]bool, components map[string][]string, graph map[string][]string) {
	if !assigned[vertex] {
		var component []string
		// if component is not yet initiated, make the slice
		component, ok := components[root]
		if !ok {
			component = make([]string, 0, 1)
			components[root] = component
		}

		components[root] = append(component, vertex)
		assigned[vertex] = true

		for _, v := range graph[vertex] {
			assignVertex(v, root, assigned, components, graph)
		}
	}
}

func EnsureNoCycles(dependencies []client.Dependency) error {
	cycles := detectCycles(dependencies)
	if len(cycles) > 0 {
		message := "Invalid resource graph. The following cycles were detected:\n"
		for _, cycle := range cycles {
			keys := make([]string, 0, len(cycle))
			for _, vertex := range cycle {
				keys = append(keys, vertex)
			}
			message = fmt.Sprintf("%sCycle: %s\n", message, strings.Join(keys, ", "))
		}

		log.Fatal(message)
		return errors.New("Graph contains cycles")
	}
	return nil
}
