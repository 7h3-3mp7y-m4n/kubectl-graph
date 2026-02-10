package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type TableFormatter struct{}

func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}
func (tf *TableFormatter) Format(g *Graph) error {
	if g.Root == nil {
		return fmt.Errorf("no root resource found")
	}
	printHeader(g.Root)
	printResourceTable("Runtime", g.Resources[ResourceTypeRuntime])
	printResourceTable("Pods", g.Resources[ResourceTypePod])
	printResourceTable("PersistentVolumeClaims", g.Resources[ResourceTypePVC])
	printResourceTable("Services", g.Resources[ResourceTypeService])

	return nil
}

func printHeader(root *Resource) {
	fmt.Printf("\n")
	color.New(color.FgCyan, color.Bold).Printf("📦 %s: %s\n", root.Type, root.Name)
	fmt.Printf("   Namespace: %s\n", root.Namespace)
	fmt.Printf("   Status: %s\n", colorizeStatus(root.Status))
	fmt.Printf("   Age: %s\n", formatAge(root.Age))

	if len(root.Details) > 0 {
		fmt.Printf("   Details:\n")
		for k, v := range root.Details {
			fmt.Printf("     %s: %v\n", k, v)
		}
	}
	fmt.Printf("\n")
}

func printResourceTable(title string, resources []*Resource) {
	if len(resources) == 0 {
		return
	}

	color.New(color.FgYellow, color.Bold).Printf("🔧 %s (%d)\n", title, len(resources))
	fmt.Println()
	fmt.Printf("  %-40s %-15s %-10s %s\n", "NAME", "STATUS", "AGE", "DETAILS")
	fmt.Printf("  %s\n", strings.Repeat("-", 100))
	for _, res := range resources {
		details := formatDetails(res)
		fmt.Printf("  %-40s %-15s %-10s %s\n",
			truncate(res.Name, 40),
			colorizeStatus(res.Status),
			formatAge(res.Age),
			details,
		)
	}

	fmt.Println()
}

func formatDetails(res *Resource) string {
	switch res.Type {
	case ResourceTypePod:
		if restarts, ok := res.Details["restarts"].(int32); ok && restarts > 0 {
			return fmt.Sprintf("restarts: %d", restarts)
		}
	case ResourceTypePVC:
		if capacity, ok := res.Details["capacity"].(string); ok && capacity != "" {
			return fmt.Sprintf("capacity: %s", capacity)
		}
	case ResourceTypeService:
		if ports, ok := res.Details["ports"].(string); ok && ports != "" {
			return fmt.Sprintf("ports: %s", ports)
		}
	case ResourceTypeRuntime:
		if rType, ok := res.Details["type"].(string); ok {
			return rType
		}
	}
	return ""
}

func colorizeStatus(status string) string {
	switch status {
	case "Running", "Bound", "Active", "Ready":
		return color.GreenString(status)
	case "Pending", "Creating":
		return color.YellowString(status)
	case "Failed", "Error", "CrashLoopBackOff":
		return color.RedString(status)
	default:
		return status
	}
}

func formatAge(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
