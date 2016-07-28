package main

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func connect() error {
	config := &restclient.Config{
		Host:     "http://localhost:8800",
		Username: "",
		Password: "",
	}
	client, err := client.New(config)
	if err != nil {
		return err
	}
	pods, err := client.Pods(api.NamespaceDefault).List(api.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Println(pods)
	return nil
}

func main() {

}
