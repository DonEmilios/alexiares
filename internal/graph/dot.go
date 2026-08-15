package graph

import (
	"fmt"
	"strings"
)

// WriteDOT renders g in Graphviz DOT format.
func WriteDOT(g Graph) string {
	var b strings.Builder
	b.WriteString("digraph alexiares {\n")
	b.WriteString("  rankdir=LR;\n")

	for _, n := range g.Nodes {
		fmt.Fprintf(&b, "  %s [label=%s, shape=%s];\n", dotID(n.ID), dotQuote(n.Label), dotShape(n.Type))
	}
	for _, e := range g.Edges {
		fmt.Fprintf(&b, "  %s -> %s [label=%s];\n", dotID(e.From), dotID(e.To), dotQuote(string(e.Type)))
	}

	b.WriteString("}\n")
	return b.String()
}

// dotID produces a valid, unique DOT node identifier from an internal
// node ID, which may itself contain characters DOT doesn't allow bare
// (colons, dots). Quoting the ID directly sidesteps needing an
// escaping scheme beyond DOT's own quoted-string rules.
func dotID(id string) string {
	return dotQuote(id)
}

func dotQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// dotShape maps a node type to a Graphviz shape, purely for
// readability when the DOT is rendered.
func dotShape(t NodeType) string {
	switch t {
	case NodeDomain, NodeURL:
		return "box"
	case NodeIP, NodeASN:
		return "ellipse"
	case NodeWallet:
		return "hexagon"
	case NodeCertificate:
		return "note"
	default:
		return "ellipse"
	}
}
