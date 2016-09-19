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

func TestKind(t *testing.T) {
	f := Yaml{}
	yaml := `apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    metadata:
      name: pi
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never`
	kind, err := f.ExtractData(yaml)
	if err != nil {
		t.Fatal(err)
	}

	if kind.Kind != "job" {
		t.Errorf("Extracted kind should be \"job\", is %s", kind)
	}
}

func TestWrap(t *testing.T) {
	f := Yaml{}
	yaml := `  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi
  spec:
    template:
      metadata:
        name: pi
      spec:
        containers:
        - name: pi
          image: perl
          command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
        restartPolicy: Never`

	wrapped, err := f.Wrap(yaml)
	if err != nil {
		t.Fatal(err)
	}
	expected := `apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: job-pi
job:
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi
  spec:
    template:
      metadata:
        name: pi
      spec:
        containers:
        - name: pi
          image: perl
          command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
        restartPolicy: Never`
	if wrapped != expected {
		t.Errorf("Wrapped doesn't match expected output\nExpected:\n%s\nAactual:\n%s", expected, wrapped)
	}
}

// TestMultiDoc checks if multi-document yaml file is wrapped properly
func TestMultiDoc(t *testing.T) {
	f := Yaml{}
	yaml := `  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi
  spec:
    trolo:
      lolo: lo
  ---
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi2
  spec:
    trolo:
      lolo: lo
  ---
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi3
  spec:
    trolo:
      lolo: lo`

	wrapped, err := f.Wrap(yaml)
	if err != nil {
		t.Fatal(err)
	}
	expected := `apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: job-pi
job:
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi
  spec:
    trolo:
      lolo: lo
---
apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: job-pi2
job:
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi2
  spec:
    trolo:
      lolo: lo
---
apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: job-pi3
job:
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: pi3
  spec:
    trolo:
      lolo: lo`

	if wrapped != expected {
		t.Errorf("Wrapped doesn't match expected output\nExpected:\n%s\nactual:\n%s", expected, wrapped)
	}
}
