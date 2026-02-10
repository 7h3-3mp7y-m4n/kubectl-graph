package collector

import (
	"7h3-3mp7y-m4n/kubectl-graph/pkg/client"
	"7h3-3mp7y-m4n/kubectl-graph/pkg/format"
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	fluidDatasetLabel = "fluid.io/dataset"
)

type Collector interface {
	Collect(ctx context.Context, namespace, name string) (*format.Graph, error)
}

var datasetGVR = schema.GroupVersionResource{
	Group:    "data.fluid.io",
	Version:  "v1alpha1",
	Resource: "datasets",
}

// DatasetCollector collects Fluid dataset resources
type DatasetCollector struct {
	client *client.Client
}

func NewDatasetCollector(c *client.Client) *DatasetCollector {
	return &DatasetCollector{client: c}
}

func (dc *DatasetCollector) Collect(ctx context.Context, namespace, name string) (*format.Graph, error) {
	// Get the dataset
	dataset, err := dc.getDataset(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	// Create graph with dataset as root
	g := format.NewGraph(dataset)

	// Collect runtime
	runtime, err := dc.getRuntime(ctx, namespace, name)
	if err == nil && runtime != nil {
		g.AddResource(runtime)
		g.AddEdge(dataset, runtime, "owns")
	}

	// Collect pods
	pods, err := dc.getPods(ctx, namespace, name)
	if err == nil {
		for _, pod := range pods {
			g.AddResource(pod)
			if runtime != nil {
				g.AddEdge(runtime, pod, "manages")
			} else {
				g.AddEdge(dataset, pod, "owns")
			}
		}
	}

	// Collect PVCs
	pvcs, err := dc.getPVCs(ctx, namespace, name)
	if err == nil {
		for _, pvc := range pvcs {
			g.AddResource(pvc)
			g.AddEdge(dataset, pvc, "references")
		}
	}

	// Collect Services
	services, err := dc.getServices(ctx, namespace, name)
	if err == nil {
		for _, svc := range services {
			g.AddResource(svc)
			if runtime != nil {
				g.AddEdge(runtime, svc, "exposes")
			} else {
				g.AddEdge(dataset, svc, "owns")
			}
		}
	}

	return g, nil
}

func (dc *DatasetCollector) getDataset(ctx context.Context, namespace, name string) (*format.Resource, error) {
	obj, err := dc.client.DynamicClient.Resource(datasetGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return convertDatasetToResource(obj), nil
}

func (dc *DatasetCollector) getRuntime(ctx context.Context, namespace, name string) (*format.Resource, error) {
	runtimeTypes := []schema.GroupVersionResource{
		{Group: "data.fluid.io", Version: "v1alpha1", Resource: "alluxioruntimes"},
		{Group: "data.fluid.io", Version: "v1alpha1", Resource: "jindoruntimes"},
		{Group: "data.fluid.io", Version: "v1alpha1", Resource: "juicefsruntimes"},
		{Group: "data.fluid.io", Version: "v1alpha1", Resource: "goosefsruntimes"},
	}

	for _, gvr := range runtimeTypes {
		obj, err := dc.client.DynamicClient.Resource(gvr).
			Namespace(namespace).
			Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return convertRuntimeToResource(obj), nil
		}
	}

	return nil, fmt.Errorf("no runtime found")
}

func (dc *DatasetCollector) getPods(ctx context.Context, namespace, datasetName string) ([]*format.Resource, error) {
	labelSelector := fmt.Sprintf("%s=%s", fluidDatasetLabel, datasetName)
	podList, err := dc.client.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	pods := make([]*format.Resource, 0, len(podList.Items))
	for _, pod := range podList.Items {
		pods = append(pods, convertPodToResource(&pod))
	}

	return pods, nil
}

func (dc *DatasetCollector) getPVCs(ctx context.Context, namespace, datasetName string) ([]*format.Resource, error) {
	labelSelector := fmt.Sprintf("%s=%s", fluidDatasetLabel, datasetName)
	pvcList, err := dc.client.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	pvcs := make([]*format.Resource, 0, len(pvcList.Items))
	for _, pvc := range pvcList.Items {
		pvcs = append(pvcs, convertPVCToResource(&pvc))
	}

	return pvcs, nil
}

func (dc *DatasetCollector) getServices(ctx context.Context, namespace, datasetName string) ([]*format.Resource, error) {
	labelSelector := fmt.Sprintf("%s=%s", fluidDatasetLabel, datasetName)
	svcList, err := dc.client.Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	services := make([]*format.Resource, 0, len(svcList.Items))
	for _, svc := range svcList.Items {
		services = append(services, convertServiceToResource(&svc))
	}

	return services, nil
}

func convertDatasetToResource(obj *unstructured.Unstructured) *format.Resource {
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	phase, _, _ := unstructured.NestedString(status, "phase")
	ufsTotal, _, _ := unstructured.NestedString(status, "ufsTotal")
	cached, _, _ := unstructured.NestedString(status, "cacheStates", "cached")

	return &format.Resource{
		Type:      format.ResourceTypeDataset,
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Status:    phase,
		Age:       getAge(obj.GetCreationTimestamp().Time),
		Details: map[string]interface{}{
			"ufsTotal": ufsTotal,
			"cached":   cached,
		},
		Labels: obj.GetLabels(),
	}
}

func convertRuntimeToResource(obj *unstructured.Unstructured) *format.Resource {
	spec, _, _ := unstructured.NestedMap(obj.Object, "spec")
	replicas, _, _ := unstructured.NestedInt64(spec, "replicas")

	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	phase, _, _ := unstructured.NestedString(status, "phase")

	runtimeType := obj.GetKind()

	return &format.Resource{
		Type:      format.ResourceTypeRuntime,
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Status:    phase,
		Age:       getAge(obj.GetCreationTimestamp().Time),
		Details: map[string]interface{}{
			"type":     runtimeType,
			"replicas": replicas,
		},
		Labels: obj.GetLabels(),
	}
}
