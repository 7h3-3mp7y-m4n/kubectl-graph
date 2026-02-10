package collector

import (
	"7h3-3mp7y-m4n/kubectl-graph/pkg/format"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func convertPodToResource(pod *corev1.Pod) *format.Resource {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}

	conditions := make([]format.Condition, 0)
	for _, cond := range pod.Status.Conditions {
		conditions = append(conditions, format.Condition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return &format.Resource{
		Type:      format.ResourceTypePod,
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Age:       getAge(pod.CreationTimestamp.Time),
		Details: map[string]interface{}{
			"node":     pod.Spec.NodeName,
			"restarts": restarts,
			"hostIP":   pod.Status.HostIP,
			"podIP":    pod.Status.PodIP,
		},
		Labels:     pod.Labels,
		Conditions: conditions,
	}
}

func getAge(t time.Time) time.Duration {
	return time.Since(t)
}
