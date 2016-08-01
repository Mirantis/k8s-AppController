package main

import (
	"testing"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/pkg/api"
)

func TestAppControllerClient(t *testing.T) {
	c, err := GetAppControllerClient()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	res, err := c.Dependencies().List(api.ListOptions{})
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	pretty.Print(res)

	res2, err := c.Dependencies().Get("dependency-1")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	pretty.Print(res2)
}
