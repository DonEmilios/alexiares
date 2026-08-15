package graph_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/graph"
)

func TestFromScanBuildsDirectInfrastructure(t *testing.T) {
	data := graph.ScanData{
		Domain:      "phish.example",
		IPs:         []string{"203.0.113.42"},
		Nameservers: []string{"ns1.evil.example"},
		Fingerprints: artifact.Fingerprints{
			Certificate: "cert-sha",
			Favicon:     "fav-sha",
			JavaScript:  []string{"js-sha-1"},
		},
		Wallets: artifact.WalletArtifacts{
			Addresses: []artifact.WalletAddress{{Chain: artifact.ChainEthereum, Address: "0xDRAINER"}},
		},
		Redirects: []artifact.Redirect{{From: "https://phish.example", To: "https://evil.example/drain", Method: "http"}},
	}

	g := graph.FromScan(data)

	wantNodeTypes := map[string]graph.NodeType{
		"domain:phish.example":           graph.NodeDomain,
		"ip:203.0.113.42":                graph.NodeIP,
		"nameserver:ns1.evil.example":    graph.NodeNameserver,
		"certificate:cert-sha":           graph.NodeCertificate,
		"favicon:fav-sha":                graph.NodeFavicon,
		"javascript:js-sha-1":            graph.NodeJavaScript,
		"wallet:0xDRAINER":               graph.NodeWallet,
		"url:https://evil.example/drain": graph.NodeURL,
	}
	gotByID := make(map[string]graph.Node)
	for _, n := range g.Nodes {
		gotByID[n.ID] = n
	}
	for id, wantType := range wantNodeTypes {
		n, ok := gotByID[id]
		if !ok {
			t.Errorf("missing expected node %q", id)
			continue
		}
		if n.Type != wantType {
			t.Errorf("node %q type = %q, want %q", id, n.Type, wantType)
		}
	}

	if len(g.Edges) != len(wantNodeTypes)-1 { // one edge per non-domain node, from the domain
		t.Errorf("Edges = %d, want %d", len(g.Edges), len(wantNodeTypes)-1)
	}
}

func TestFromScanConnectsClusterSiblings(t *testing.T) {
	fp := artifact.Fingerprints{Favicon: "shared-fav-sha"}
	data := graph.ScanData{
		Domain:       "phish.example",
		Fingerprints: fp,
		Correlation: artifact.Correlation{
			Clusters: []artifact.Cluster{
				{
					SignatureID: "cluster001",
					Matches: []artifact.Match{
						{SignatureID: "cluster001", Category: artifact.MatchFavicon, Value: "shared-fav-sha"},
					},
					RelatedDomains: []string{"sibling.example"},
				},
			},
		},
	}

	g := graph.FromScan(data)

	var siblingNode, hubNode bool
	var siblingToHubEdge bool
	for _, n := range g.Nodes {
		if n.ID == "domain:sibling.example" {
			siblingNode = true
		}
		if n.ID == "favicon:shared-fav-sha" {
			hubNode = true
		}
	}
	for _, e := range g.Edges {
		if e.From == "domain:sibling.example" && e.To == "favicon:shared-fav-sha" && e.Type == graph.EdgeSharesFavicon {
			siblingToHubEdge = true
		}
	}

	if !siblingNode {
		t.Error("sibling domain node missing from graph")
	}
	if !hubNode {
		t.Error("shared favicon hub node missing from graph")
	}
	if !siblingToHubEdge {
		t.Errorf("sibling domain not connected to shared favicon hub; edges: %+v", g.Edges)
	}
}

func TestFromScanEmptyDataStillProducesDomainNode(t *testing.T) {
	g := graph.FromScan(graph.ScanData{Domain: "bare.example"})
	if len(g.Nodes) != 1 || g.Nodes[0].ID != "domain:bare.example" {
		t.Errorf("FromScan(empty) nodes = %v, want just the domain node", g.Nodes)
	}
	if len(g.Edges) != 0 {
		t.Errorf("FromScan(empty) edges = %v, want none", g.Edges)
	}
}
