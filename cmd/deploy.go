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

package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"
)

func deploy(cmd *cobra.Command, args []string) {
	var err error

	var url string
	if len(args) > 0 {
		url = args[0]
	}
	if url == "" {
		url = os.Getenv("KUBERNETES_CLUSTER_URL")
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}
	controller := scheduler.NewDeploymentController(c)
	stopChan := make(chan struct{})
	controller.Run(stopChan)
}

// InitRunCommand returns cobra command for performing AppController graph deployment
func InitRunCommand() (*cobra.Command, error) {
	return &cobra.Command{
		Use:   "run",
		Short: "Start deployment of AppController graph",
		Long:  "Start deployment of AppController graph",
		Run:   deploy,
	}, nil
}
