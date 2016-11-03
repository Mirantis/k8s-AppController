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

package format

import (
	"testing"
)

// TestKindJSON tests data retrieval from k8s objects
func TestKindJSON(t *testing.T) {
	f := JSON{}
	text := `{"kind": "Job"}`
	kind, err := f.ExtractData(text)
	if err != nil {
		t.Fatal(err)
	}

	if kind.Kind != "job" {
		t.Errorf("Extracted kind should be \"job\", is %s", kind)
	}
}

// TestWrapJSON tests if K8s objects are properly wrapped
func TestWrapJSON(t *testing.T) {
	f := JSON{}
	text := `{"kind": "Job", "metadata": {"name": "name"}}` + "\n"

	wrapped, err := f.Wrap(text)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{
    "apiVersion": "appcontroller.k8s/v1alpha1",
    "kind": "Definition",
    "metadata": {
        "name": "job-name"
    },
    "job": {"kind": "Job", "metadata": {"name": "name"}}
}` + "\n"
	if wrapped != expected {
		t.Errorf("Wrapped doesn't match expected output\nExpected:\n%s\nAactual:\n%s", expected, wrapped)
	}
}
