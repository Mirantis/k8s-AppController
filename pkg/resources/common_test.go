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

package resources

import (
	"testing"

	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

// TestGetStringMeta checks metadata retrieval from a resource
func TestGetStringMeta(t *testing.T) {
	r := mocks.NewResource("fake", "not ready")

	if GetStringMeta(r, "non-existing key", "default") != "default" {
		t.Error("GetStringMeta for non-existing key returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": "value"})

	if GetStringMeta(r, "non-existing key", "default") != "default" {
		t.Error("GetStringMeta for non-existing key returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": 1})

	if GetStringMeta(r, "key", "default") != "default" {
		t.Error("GetStringMeta for non-string value returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": "value"})

	if GetStringMeta(r, "key", "default") != "value" {
		t.Error("GetStringMeta returned not an actual value")
	}
}

// TestGetIntMeta checks metadata retrieval from a resource
func TestGetIntMeta(t *testing.T) {
	r := mocks.NewResource("fake", "not ready")

	if GetIntMeta(r, "non-existing key", -1) != -1 {
		t.Error("GetIntMeta for non-existing key returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": "value"})

	if GetIntMeta(r, "non-existing key", -1) != -1 {
		t.Error("GetIntMeta for non-existing key returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": "value"})

	if GetIntMeta(r, "key", -1) != -1 {
		t.Error("GetIntMeta for non-int value returned not a default value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": 42})

	if GetIntMeta(r, "key", -1) != 42 {
		t.Error("GetIntMeta returned not an actual value")
	}

	r = mocks.NewResourceWithMeta("fake", "not ready", map[string]interface{}{"key": 42.})

	if GetIntMeta(r, "key", -1) != 42 {
		t.Error("GetIntMeta returned not an actual value")
	}
}
