package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/scheduler"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/labels"
)

// GetStatus is a command that prints the deployment status
func GetStatus(cmd *cobra.Command, args []string) {
	var err error

	labelSelector, err := getLabelSelector(cmd)
	if err != nil {
		log.Fatal(err)
	}

	shouldReport, err := cmd.Flags().GetBool("report")
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
	status, report, _ := graph.GetStatus()
	fmt.Printf("STATUS: %s\n", status)
	if shouldReport {
		fmt.Printf("REPORT:\n%s", report)
	}
}

// InitGetStatusCommand is an initialiser for get-status
func InitGetStatusCommand() (*cobra.Command, error) {
	var err error
	run := &cobra.Command{
		Use:   "get-status",
		Short: "Get status of deployment",
		Long:  "Get status of deployment",
		Run:   GetStatus,
	}
	var labelSelector string
	run.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector. Overrides KUBERNETES_AC_LABEL_SELECTOR env variable in AppController pod.")

	var shouldReport bool
	run.Flags().BoolVarP(&shouldReport, "report", "r", false, "Print report")
	return run, err
}
