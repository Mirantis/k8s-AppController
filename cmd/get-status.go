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
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"
	"github.com/spf13/cobra"

	"k8s.io/client-go/pkg/labels"
)

// GetStatus is a command that prints the deployment status
func getStatus(cmd *cobra.Command, args []string) {
	var err error

	labelSelector, err := getLabelSelector(cmd)
	if err != nil {
		log.Fatal(err)
	}

	getJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		log.Fatal(err)
	}
	getReport, err := cmd.Flags().GetBool("report")
	if err != nil {
		log.Fatal(err)
	}

	flowName := interfaces.DefaultFlowName

	if len(args) > 1 {
		log.Fatal("Too many command line arguments")
	}
	if len(args) == 1 {
		flowName = args[0]
	}
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		log.Fatal(err)
	}

	c, err := client.New(url)
	if err != nil {
		log.Fatal(err)
	}
	sel, err := labels.Parse(labelSelector)
	if err != nil {
		log.Fatal(err)
	}
	anyArgs, err := cmd.Flags().GetBool("undeclared-args")
	if err != nil {
		log.Fatal(err)
	}
	flowArgs, err := cmd.Flags().GetStringArray("arg")
	if err != nil {
		log.Fatal(err)
	}
	flowArgMap := map[string]string{}
	for _, val := range flowArgs {
		key, value, err := parseArg(val)
		if err != nil {
			log.Fatal(err)
		}
		flowArgMap[key] = value
	}

	log.Println("Using label selector:", labelSelector)

	options := interfaces.DependencyGraphOptions{
		FlowName:            flowName,
		ExportedOnly:        true,
		Args:                flowArgMap,
		AllowUndeclaredArgs: anyArgs,
	}

	status, report, err := scheduler.GetStatus(c, sel, options)
	if err != nil {
		log.Fatal(err)
	}
	if getJSON {
		data, err := json.Marshal(report)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(data))
	} else {
		fmt.Printf("STATUS: %s\n", status)
		if getReport {
			data := report.AsText(0)
			for _, line := range data {
				fmt.Println(line)
			}
		}
	}
}

// InitGetStatusCommand is an initialiser for get-status
func InitGetStatusCommand() (*cobra.Command, error) {
	var err error
	run := &cobra.Command{
		Use:   "get-status [flow-name]",
		Short: "Get status of deployment",
		Long:  "Get status of deployment",
		Run:   getStatus,
	}
	var labelSelector string
	run.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector. Overrides KUBERNETES_AC_LABEL_SELECTOR env variable in AppController pod.")

	var getJSON, report bool
	run.Flags().BoolVarP(&getJSON, "json", "j", false, "Output JSON")
	run.Flags().BoolVarP(&report, "report", "r", false, "Get human-readable full report")
	run.Flags().String("url", os.Getenv("KUBERNETES_CLUSTER_URL"),
		"URL of the Kubernetes cluster. Overrides KUBERNETES_CLUSTER_URL env variable in AppController pod.")
	run.Flags().Bool("undeclared-args", false, "Allow undeclared arguments")
	run.Flags().StringArray("arg", []string{}, "Flow arguments (key=value)")

	return run, err
}
