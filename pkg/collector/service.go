package collector

import (
	"7h3-3mp7y-m4n/kubectl-graph/pkg/format"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func convertServiceToResource(svc *corev1.Service) *format.Resource {
	ports := ""
	for i, port := range svc.Spec.Ports {
		if i > 0 {
			ports += ", "
		}
		ports += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
	}

	return &format.Resource{
		Type:      format.ResourceTypeService,
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Status:    "Active", // Services don't have a phase
		Age:       getAge(svc.CreationTimestamp.Time),
		Details: map[string]interface{}{
			"type":      string(svc.Spec.Type),
			"clusterIP": svc.Spec.ClusterIP,
			"ports":     ports,
		},
		Labels: svc.Labels,
	}
}
