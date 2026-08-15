package graph_test

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strings"
	"testing"

	"github.com/alexiares/alexiares/internal/graph"
)

func sampleGraph() graph.Graph {
	b := graph.NewBuilder()
	b.AddNode("domain:phish.example", graph.NodeDomain, "phish.example")
	b.AddNode("favicon:abc123", graph.NodeFavicon, "abc123")
	b.AddEdge("domain:phish.example", "favicon:abc123", graph.EdgeSharesFavicon)
	return b.Build()
}

func TestWriteDOTContainsNodesAndEdges(t *testing.T) {
	out := graph.WriteDOT(sampleGraph())

	if !strings.HasPrefix(out, "digraph alexiares {") {
		t.Errorf("WriteDOT() does not start with digraph header:\n%s", out)
	}
	if !strings.Contains(out, `"domain:phish.example"`) {
		t.Errorf("WriteDOT() missing domain node:\n%s", out)
	}
	if !strings.Contains(out, `"favicon:abc123"`) {
		t.Errorf("WriteDOT() missing favicon node:\n%s", out)
	}
	if !strings.Contains(out, `shares_favicon`) {
		t.Errorf("WriteDOT() missing edge label:\n%s", out)
	}
}

func TestWriteDOTEscapesQuotesInLabels(t *testing.T) {
	b := graph.NewBuilder()
	b.AddNode("n1", graph.NodeDomain, `has "quotes" in it`)
	out := graph.WriteDOT(b.Build())

	if !strings.Contains(out, `\"quotes\"`) {
		t.Errorf("WriteDOT() did not escape embedded quotes:\n%s", out)
	}
}

func TestWriteGraphMLIsValidXML(t *testing.T) {
	out, err := graph.WriteGraphML(sampleGraph())
	if err != nil {
		t.Fatalf("WriteGraphML() error = %v", err)
	}

	var doc struct {
		XMLName xml.Name `xml:"graphml"`
		Graph   struct {
			Nodes []struct {
				ID string `xml:"id,attr"`
			} `xml:"node"`
			Edges []struct {
				Source string `xml:"source,attr"`
				Target string `xml:"target,attr"`
			} `xml:"edge"`
		} `xml:"graph"`
	}
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("WriteGraphML() produced invalid XML: %v\n%s", err, out)
	}
	if len(doc.Graph.Nodes) != 2 {
		t.Errorf("GraphML nodes = %d, want 2", len(doc.Graph.Nodes))
	}
	if len(doc.Graph.Edges) != 1 {
		t.Errorf("GraphML edges = %d, want 1", len(doc.Graph.Edges))
	}
}

func TestWriteGraphMLEscapesSpecialCharacters(t *testing.T) {
	b := graph.NewBuilder()
	b.AddNode("n1", graph.NodeDomain, `<script>alert("x")</script>`)
	out, err := graph.WriteGraphML(b.Build())
	if err != nil {
		t.Fatalf("WriteGraphML() error = %v", err)
	}
	if strings.Contains(out, "<script>") {
		t.Errorf("WriteGraphML() did not escape a label containing markup:\n%s", out)
	}
	// Must still parse cleanly despite the hostile label.
	var doc any
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("WriteGraphML() with hostile label produced invalid XML: %v", err)
	}
}

func TestWriteJSONRoundTripsLosslessly(t *testing.T) {
	g := sampleGraph()

	out, err := graph.WriteJSON(g)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	var got graph.Graph
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshaling WriteJSON() output: %v", err)
	}
	if !reflect.DeepEqual(g, got) {
		t.Errorf("JSON round-trip lost data:\n  original: %+v\n  round-tripped: %+v", g, got)
	}
}
