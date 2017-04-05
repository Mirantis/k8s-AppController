// Copyright 2017 Mirantis
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

package cmd

import (
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"

	"github.com/spf13/cobra"
)

var Deploy *cobra.Command = initDeployCommand()

func deploy(cmd *cobra.Command, args []string) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		log.Fatal(err)
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Watching for scheduled deployments")
	stopChan := make(chan struct{})
	scheduler.ProcessDeploymentTasks(c, stopChan)
}

func initDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy scheduled AppController flows",
		Long:  "Deploy scheduled AppController flows",
		Run:   deploy,
	}

	cmd.Flags().String("url", os.Getenv("KUBERNETES_CLUSTER_URL"),
		"URL of the Kubernetes cluster. Overrides KUBERNETES_CLUSTER_URL env variable in AppController pod.")

	return cmd
}
