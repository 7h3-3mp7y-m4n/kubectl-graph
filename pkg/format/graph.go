package format

import "time"

type ResourceType string

const (
	ResourceTypeDataset     ResourceType = "Dataset"
	ResourceTypeRuntime     ResourceType = "Runtime"
	ResourceTypePod         ResourceType = "Pod"
	ResourceTypePVC         ResourceType = "PersistentVolumeClaim"
	ResourceTypePV          ResourceType = "PersistentVolume"
	ResourceTypeService     ResourceType = "Service"
	ResourceTypeStatefulSet ResourceType = "StatefulSet"
	ResourceTypeDaemonSet   ResourceType = "DaemonSet"
)

type Resource struct {
	Type       ResourceType
	Name       string
	Namespace  string
	Status     string
	Age        time.Duration
	Details    map[string]interface{}
	Labels     map[string]string
	Conditions []Condition
}

type Condition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

type Graph struct {
	Root      *Resource
	Resources map[ResourceType][]*Resource
	Edges     []Edge
}

type Edge struct {
	From *Resource
	To   *Resource
	Type string
}

func NewGraph(root *Resource) *Graph {
	return &Graph{
		Root:      root,
		Resources: make(map[ResourceType][]*Resource),
		Edges:     make([]Edge, 0),
	}
}

func (g *Graph) AddResource(resource *Resource) {
	if g.Resources[resource.Type] == nil {
		g.Resources[resource.Type] = make([]*Resource, 0)
	}
	g.Resources[resource.Type] = append(g.Resources[resource.Type], resource)
}

func (g *Graph) AddEdge(from, to *Resource, edgeType string) {
	g.Edges = append(g.Edges, Edge{
		From: from,
		To:   to,
		Type: edgeType,
	})
}

func (g *Graph) GetChildren(parent *Resource) []*Resource {
	children := make([]*Resource, 0)
	for _, edge := range g.Edges {
		if edge.From == parent {
			children = append(children, edge.To)
		}
	}
	return children
}
