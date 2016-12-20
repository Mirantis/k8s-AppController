package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"github.com/spf13/cobra"
	"k8s.io/client-go/1.5/pkg/labels"
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
	sel, err := labels.Parse(labelSelector)
	if err != nil {
		log.Fatal(err)
	}
	graph, err := scheduler.BuildDependencyGraph(c, sel)
	if err != nil {
		log.Fatal(err)
	}
	status, report := graph.GetStatus()
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
		Use:   "get-status",
		Short: "Get status of deployment",
		Long:  "Get status of deployment",
		Run:   getStatus,
	}
	var labelSelector string
	run.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector. Overrides KUBERNETES_AC_LABEL_SELECTOR env variable in AppController pod.")

	var getJSON, report bool
	run.Flags().BoolVarP(&getJSON, "json", "j", false, "Output JSON")
	run.Flags().BoolVarP(&report, "report", "r", false, "Get human-readable full report")
	return run, err
}
