package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceLister struct {
	client *kubernetes.Clientset
}

func NewResourceLister(client *kubernetes.Clientset) *ResourceLister {
	return &ResourceLister{
		client: client,
	}
}

func (r *ResourceLister) ListResources(ctx context.Context, namespace, deploymentName string) error {
	deployment, err := r.client.AppsV1().
		Deployments(namespace).
		Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment -- %w", err)
	}

	fmt.Printf("Deployment: %s/%s\n", namespace, deploymentName)
	fmt.Printf("Replicas: %d/%d/%d (desired/current/ready)\n",
		*deployment.Spec.Replicas,
		deployment.Status.Replicas,
		deployment.Status.ReadyReplicas)
	fmt.Printf("Strategy: %s\n\n", deployment.Spec.Strategy.Type)

	replicaSets, err := r.client.AppsV1().
		ReplicaSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list replicasets -- %w", err)
	}

	var currentRS *string

	for _, rs := range replicaSets.Items {
		for _, owner := range rs.OwnerReferences {
			if owner.Kind == "Deployment" && owner.Name == deploymentName {
				if rs.Status.Replicas > 0 {
					fmt.Printf("  ReplicaSet: %s\n", rs.Name)
					fmt.Printf("   Replicas: %d/%d/%d (desired/current/ready)\n",
						*rs.Spec.Replicas,
						rs.Status.Replicas,
						rs.Status.ReadyReplicas)
					fmt.Println()

					currentRS = &rs.Name
				}
			}
		}
	}

	if currentRS != nil {
		fmt.Printf("Current ReplicaSet: %s\n", *currentRS)
	}

	return nil
}
