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
