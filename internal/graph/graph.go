// Package graph represents infrastructure relationships as a typed
// node/edge graph, built from a scan's collected artifacts and
// correlation results, and serializable to DOT, GraphML, and JSON.
//
// Two node types the spec names — ASN and Registrar — have no
// producer yet: no extractor collects ASN attribution (no bundled
// IP-to-ASN database, per the "no external database" performance
// goal) or registrar data (no WHOIS extractor exists). Their type
// constants are defined for schema completeness and for signatures or
// future extractors to use, but FromScan never emits them today.
package graph

// NodeType identifies what a Node represents.
type NodeType string

// Supported node types, per the spec's infrastructure graph model.
const (
	NodeDomain      NodeType = "domain"
	NodeURL         NodeType = "url"
	NodeIP          NodeType = "ip"
	NodeASN         NodeType = "asn"
	NodeCertificate NodeType = "certificate"
	NodeFavicon     NodeType = "favicon"
	NodeJavaScript  NodeType = "javascript"
	NodeWallet      NodeType = "wallet"
	NodeTelegram    NodeType = "telegram"
	NodeRegistrar   NodeType = "registrar"
	NodeNameserver  NodeType = "nameserver"
)

// EdgeType identifies the relationship an Edge represents.
type EdgeType string

// Supported edge types, per the spec's infrastructure graph model.
const (
	EdgeResolvesTo       EdgeType = "resolves_to"
	EdgeHostedBy         EdgeType = "hosted_by"
	EdgeUsesCertificate  EdgeType = "uses_certificate"
	EdgeSharesFavicon    EdgeType = "shares_favicon"
	EdgeReusesScript     EdgeType = "reuses_script"
	EdgeContainsWallet   EdgeType = "contains_wallet"
	EdgeRedirectsTo      EdgeType = "redirects_to"
	EdgeRegisteredWith   EdgeType = "registered_with"
	EdgeSharesNameserver EdgeType = "shares_nameserver"
)

// Node is a single graph vertex. ID is unique within a Graph and is
// how Edges reference their endpoints.
type Node struct {
	ID    string   `json:"id"`
	Type  NodeType `json:"type"`
	Label string   `json:"label"`
}

// Edge is a directed, typed relationship between two Node IDs.
type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Type EdgeType `json:"type"`
}

// Graph is a full infrastructure relationship graph.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
