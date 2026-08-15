package output

import (
	"fmt"

	"github.com/alexiares/alexiares/internal/graph"
)

// Format identifies a supported output format.
type Format string

// Supported output formats, per the spec's Output Formats section.
const (
	FormatTerminal Format = "terminal"
	FormatJSON     Format = "json"
	FormatGraphML  Format = "graphml"
	FormatDOT      Format = "dot"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
)

// Render dispatches r to the renderer for format. An empty format
// renders as terminal text, the spec's default.
func Render(format Format, r ScanResult) (string, error) {
	switch format {
	case "", FormatTerminal:
		return RenderTerminal(r), nil
	case FormatJSON:
		return RenderJSON(r)
	case FormatGraphML:
		return graph.WriteGraphML(r.Graph)
	case FormatDOT:
		return graph.WriteDOT(r.Graph), nil
	case FormatCSV:
		return RenderCSV(r)
	case FormatMarkdown:
		return RenderMarkdown(r), nil
	default:
		return "", fmt.Errorf("output: unsupported format %q (want terminal, json, graphml, dot, csv, or markdown)", format)
	}
}
