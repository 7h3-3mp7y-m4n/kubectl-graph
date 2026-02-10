package format

type Formatter interface {
	Format(g *Graph) error
}

func NewFormatter(formatType string) Formatter {
	switch formatType {
	case "json":
		return NewJSONFormatter()
	default:
		return NewTableFormatter()
	}
}
