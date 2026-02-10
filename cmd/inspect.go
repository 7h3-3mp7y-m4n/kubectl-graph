package cmd

import (
	"7h3-3mp7y-m4n/kubectl-graph/pkg/client"
	"7h3-3mp7y-m4n/kubectl-graph/pkg/collector"
	"7h3-3mp7y-m4n/kubectl-graph/pkg/format"
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	namespace  string
	output     string
	kubeconfig string
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [resource-type] [resource-name]",
	Short: "Inspect a Kubernetes resource and its dependencies",
	Args:  cobra.ExactArgs(2),
	Run:   runInspect,
}

func init() {
	inspectCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the resource")
	inspectCmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table|tree|json")
	inspectCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (defaults to ~/.kube/config)")
}

func runInspect(cmd *cobra.Command, args []string) {
	resourceType := args[0]
	resourceName := args[1]
	if output != "table" && output != "tree" && output != "json" {
		exitWithError("invalid output format", fmt.Errorf("must be one of: table, tree, json"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	k8sClient, err := client.NewClient(kubeconfig)
	if err != nil {
		exitWithError("failed to create kubernetes client", err)
	}
	var c collector.Collector
	switch resourceType {
	case "dataset":
		c = collector.NewDatasetCollector(k8sClient)
	default:
		exitWithError("unsupported resource type", fmt.Errorf("type '%s' not supported yet", resourceType))
	}
	resourceGraph, err := c.Collect(ctx, namespace, resourceName)
	if err != nil {
		exitWithError("failed to collect resources", err)
	}
	formatter := format.NewFormatter(output)
	if err := formatter.Format(resourceGraph); err != nil {
		exitWithError("failed to format results", err)
	}
}
