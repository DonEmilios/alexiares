package graph

import (
	"encoding/xml"
	"fmt"
)

// GraphML's schema, modeled as Go structs so encoding/xml handles
// escaping — hand-built XML strings are a routine source of malformed
// or injectable output when a label contains &, <, or >.
type graphMLDocument struct {
	XMLName xml.Name     `xml:"graphml"`
	Xmlns   string       `xml:"xmlns,attr"`
	Keys    []graphMLKey `xml:"key"`
	Graph   graphMLGraph `xml:"graph"`
}

type graphMLKey struct {
	ID       string `xml:"id,attr"`
	For      string `xml:"for,attr"`
	AttrName string `xml:"attr.name,attr"`
	AttrType string `xml:"attr.type,attr"`
}

type graphMLGraph struct {
	ID          string        `xml:"id,attr"`
	EdgeDefault string        `xml:"edgedefault,attr"`
	Nodes       []graphMLNode `xml:"node"`
	Edges       []graphMLEdge `xml:"edge"`
}

type graphMLNode struct {
	ID   string        `xml:"id,attr"`
	Data []graphMLData `xml:"data"`
}

type graphMLEdge struct {
	Source string        `xml:"source,attr"`
	Target string        `xml:"target,attr"`
	Data   []graphMLData `xml:"data"`
}

type graphMLData struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

// WriteGraphML renders g in GraphML format.
func WriteGraphML(g Graph) (string, error) {
	doc := graphMLDocument{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Keys: []graphMLKey{
			{ID: "n_type", For: "node", AttrName: "type", AttrType: "string"},
			{ID: "n_label", For: "node", AttrName: "label", AttrType: "string"},
			{ID: "e_type", For: "edge", AttrName: "type", AttrType: "string"},
		},
		Graph: graphMLGraph{ID: "alexiares", EdgeDefault: "directed"},
	}

	for _, n := range g.Nodes {
		doc.Graph.Nodes = append(doc.Graph.Nodes, graphMLNode{
			ID: n.ID,
			Data: []graphMLData{
				{Key: "n_type", Value: string(n.Type)},
				{Key: "n_label", Value: n.Label},
			},
		})
	}
	for _, e := range g.Edges {
		doc.Graph.Edges = append(doc.Graph.Edges, graphMLEdge{
			Source: e.From,
			Target: e.To,
			Data:   []graphMLData{{Key: "e_type", Value: string(e.Type)}},
		})
	}

	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling GraphML: %w", err)
	}
	return xml.Header + string(out) + "\n", nil
}
