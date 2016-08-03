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
