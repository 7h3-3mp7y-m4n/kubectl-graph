package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	namespace  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kubectl-graph",
		Short: "A kubectl plugin to help with deployment troubleshooting",
		Long:  `kubectl-grah provides utilities to list and diagnose Kubernetes deployment resources.`,
	}
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace to operate in")
	rootCmd.AddCommand(list())
	// rootCmd.AddCommand(diagnose())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}

func list() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list DEPLOYMENT_NAME",
		Short: "List all resources associated with a deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deploymentName := args[0]
			client, err := getKubernetesClient()
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}
			lister := NewResourceLister(client)
			return lister.ListResources(context.Background(), namespace, deploymentName)
		},
	}
	return cmd
}

// func diagnose() *cobra.Command {
// 	var outputFile string
// 	cmd := &cobra.Command{
// 		Use:   "diagnose DEPLOYMENT_NAME",
// 		Short: "Diagnose issues with a deployment",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			deploymentName := args[0]
// 			client, err := getKubernetesClient()
// 			if err != nil {
// 				return fmt.Errorf("failed to create kubernetes client: %w", err)
// 			}
// 			diagnoser := NewDiagnoser(client)
// 			return diagnoser.Diagnose(context.Background(), namespace, deploymentName, outputFile)
// 		},
// 	}

// 	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for diagnostic report (default: stdout)")

// 	return cmd
// }

func getKubernetesClient() (*kubernetes.Clientset, error) {
	kubeconfigPath := kubeconfig
	if kubeconfigPath == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfigPath = env
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			kubeconfigPath = home + "/.kube/config"
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
