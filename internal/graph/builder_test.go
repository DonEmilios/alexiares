package graph_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/graph"
)

func TestBuilderDeduplicatesNodesAndEdges(t *testing.T) {
	b := graph.NewBuilder()
	b.AddNode("domain:a.example", graph.NodeDomain, "a.example")
	b.AddNode("domain:a.example", graph.NodeDomain, "a.example (duplicate call)")
	b.AddEdge("domain:a.example", "ip:1.2.3.4", graph.EdgeResolvesTo)
	b.AddEdge("domain:a.example", "ip:1.2.3.4", graph.EdgeResolvesTo)

	g := b.Build()
	if len(g.Nodes) != 1 {
		t.Errorf("Nodes = %v, want 1 deduplicated node", g.Nodes)
	}
	if g.Nodes[0].Label != "a.example" {
		t.Errorf("Nodes[0].Label = %q, want first-call label preserved", g.Nodes[0].Label)
	}
	if len(g.Edges) != 1 {
		t.Errorf("Edges = %v, want 1 deduplicated edge", g.Edges)
	}
}

func TestBuilderHasNode(t *testing.T) {
	b := graph.NewBuilder()
	if b.HasNode("domain:a.example") {
		t.Error("HasNode() = true before AddNode, want false")
	}
	b.AddNode("domain:a.example", graph.NodeDomain, "a.example")
	if !b.HasNode("domain:a.example") {
		t.Error("HasNode() = false after AddNode, want true")
	}
}

func TestBuilderDistinguishesEdgeTypesBetweenSameNodes(t *testing.T) {
	b := graph.NewBuilder()
	b.AddEdge("a", "b", graph.EdgeResolvesTo)
	b.AddEdge("a", "b", graph.EdgeRedirectsTo)

	g := b.Build()
	if len(g.Edges) != 2 {
		t.Errorf("Edges = %v, want 2 (different types between same nodes are distinct)", g.Edges)
	}
}
