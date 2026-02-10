package format

import (
	"encoding/json"
	"fmt"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}
func (jf *JSONFormatter) Format(g *Graph) error {
	output := struct {
		Root      *Resource                    `json:"root"`
		Resources map[ResourceType][]*Resource `json:"resources"`
		Edges     []Edge                       `json:"edges"`
	}{
		Root:      g.Root,
		Resources: g.Resources,
		Edges:     g.Edges,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
