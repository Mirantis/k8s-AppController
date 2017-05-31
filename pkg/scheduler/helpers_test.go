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
	"strings"
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestPermute tests permute function
func TestPermute(t *testing.T) {
	alphabets := [][]string{
		{"1", "2", "3"},
		{"+", "-"},
		{"a", "b"},
		{"="},
	}

	expected := map[string]bool{
		"1+a=": true,
		"1+b=": true,
		"1-a=": true,
		"1-b=": true,
		"2+a=": true,
		"2+b=": true,
		"2-a=": true,
		"2-b=": true,
		"3+a=": true,
		"3+b=": true,
		"3-a=": true,
		"3-b=": true,
	}
	permutations := permute(alphabets)
	for _, combination := range permutations {
		combinationStr := strings.Join(combination, "")
		if !expected[combinationStr] {
			t.Errorf("unexpected combination %s", combinationStr)
		} else {
			delete(expected, combinationStr)
		}
	}
	if len(expected) != 0 {
		t.Error("not all combinations were generated")
	}

	alphabets = append(alphabets, make([]string, 0))
	if len(permute(alphabets)) != 0 {
		t.Error("empty alphabet didin't result in empty permutation list")
	}
}

// TestExpendListExpression tests list expression translation to list of strings
func TestExpendListExpression(t *testing.T) {
	table := map[string][]string{
		"1":              {"1"},
		"1..5":           {"1", "2", "3", "4", "5"},
		"2..-1":          {},
		"a, b":           {"a", "b"},
		"a, b, 2..4":     {"a", "b", "2", "3", "4"},
		"-1..1, 2..4, x": {"-1", "0", "1", "2", "3", "4", "x"},
		"a..b":           {"a..b"},
		"..":             {".."},
		"1...3":          {"1...3"},
		"1..b":           {"1..b"},
		"a..b, 1..3":     {"a..b", "1", "2", "3"},
		"a..b, c..d":     {"a..b", "c..d"},
		"":               {},
	}
	for expr, expected := range table {
		result := expandListExpression(expr)
		if len(result) != len(expected) {
			t.Errorf("unexpected result length for expression %s: %d != %d", expr, len(result), len(expected))
		} else {
			for i := range expected {
				if expected[i] != result[i] {
					t.Errorf("invalid entry %d for expression %s: %s != %s", i, expr, expected[i], result[i])
				}
			}
		}
	}
}

// TestGetStringMeta checks metadata retrieval from a resource
func TestGetStringMeta(t *testing.T) {
	r := &scheduledResource{Resource: mocks.NewResource("fake", 0)}

	if getStringMeta(r, "non-existing key", "default") != "default" {
		t.Error("GetStringMeta for non-existing key returned not a default value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": "value"},
	}

	if getStringMeta(r, "non-existing key", "default") != "default" {
		t.Error("GetStringMeta for non-existing key returned not a default value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": 1},
	}

	if getStringMeta(r, "key", "default") != "default" {
		t.Error("GetStringMeta for non-string value returned not a default value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": "value"},
	}

	if getStringMeta(r, "key", "default") != "value" {
		t.Error("GetStringMeta returned not an actual value")
	}
}

// TestGetIntMeta checks metadata retrieval from a resource
func TestGetIntMeta(t *testing.T) {
	r := &scheduledResource{Resource: mocks.NewResource("fake", 0)}

	if getIntMeta(r, "non-existing key", -1) != -1 {
		t.Error("GetIntMeta for non-existing key returned not a default value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": "value"},
	}

	if getIntMeta(r, "non-existing key", -1) != -1 {
		t.Error("GetIntMeta for non-existing key returned not a default value")
	}

	if getIntMeta(r, "key", -1) != -1 {
		t.Error("GetIntMeta for non-int value returned not a default value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": 42},
	}
	if getIntMeta(r, "key", -1) != 42 {
		t.Error("GetIntMeta returned not an actual value")
	}

	r = &scheduledResource{
		Resource:     mocks.NewResource("fake", 0),
		resourceMeta: map[string]interface{}{"key": 42.},
	}
	if getIntMeta(r, "key", -1) != 42 {
		t.Error("GetIntMeta returned not an actual value")
	}
}
