package main

import (
	"log"
	"os"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: check-pod-status POD-NAME")
	}

	podName := os.Args[1]

	config := &restclient.Config{
		Host: "http://127.0.0.1:8080",
	}

	client, err := client.New(config)
	if err != nil {
		log.Fatal(err)
	}

	pod, err := client.Pods(api.NamespaceDefault).Get(podName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Pod Status: %v\n", pod.Status.Phase)
}
