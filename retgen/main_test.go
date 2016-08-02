package client

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
	pretty.Print("deps:")
	pretty.Print(res)

	res2, err := c.Dependencies().Get("dependency-1")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	pretty.Print("dep:")
	pretty.Print(res2)

	res3, err := c.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	pretty.Print("resources:")
	pretty.Print(res3)

	res4, err := c.ResourceDefinitions().Get("pod-definition-1")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	pretty.Print("resource:")
	pretty.Print(res4)
}
