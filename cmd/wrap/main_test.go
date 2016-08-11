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

package main

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestInput(t *testing.T) {
	in, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	content := "trololo\n lololo\n lololo\n\n"
	expected := "  trololo\n   lololo\n   lololo\n  \n"

	_, err = io.WriteString(in, content)
	if err != nil {
		t.Fatal(err)
	}

	_, err = in.Seek(0, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	result := getInput(in)

	if result != expected {
		t.Errorf("\"%s\" is not equal to \"%s\"", result, expected)
	}

}

func TestKind(t *testing.T) {
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
	kind, err := getKind(yaml)
	if err != nil {
		t.Fatal(err)
	}

	if kind != "job" {
		t.Errorf("Extracted kind should be \"job\", is %s", kind)
	}
}

func TestWrap(t *testing.T) {
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

	wrapped, err := getWrappedYaml(yaml, "name")
	if err != nil {
		t.Fatal(err)
	}
	expected := `apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: name
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
