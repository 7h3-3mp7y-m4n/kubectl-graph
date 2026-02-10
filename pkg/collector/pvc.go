package collector

import (
	"7h3-3mp7y-m4n/kubectl-graph/pkg/format"

	corev1 "k8s.io/api/core/v1"
)

func convertPVCToResource(pvc *corev1.PersistentVolumeClaim) *format.Resource {
	capacity := ""
	if cap, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
		capacity = cap.String()
	}

	requested := ""
	if req, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
		requested = req.String()
	}

	return &format.Resource{
		Type:      format.ResourceTypePVC,
		Name:      pvc.Name,
		Namespace: pvc.Namespace,
		Status:    string(pvc.Status.Phase),
		Age:       getAge(pvc.CreationTimestamp.Time),
		Details: map[string]interface{}{
			"volumeName": pvc.Spec.VolumeName,
			"capacity":   capacity,
			"requested":  requested,
		},
		Labels: pvc.Labels,
	}
}
