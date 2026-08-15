package graph

import (
	"fmt"

	"github.com/alexiares/alexiares/internal/artifact"
)

// ScanData bundles everything about a scanned target FromScan needs:
// its directly observed infrastructure plus, optionally, the
// correlation engine's matches against the signature repository.
// Correlation may be the zero value for a scan run without signatures
// loaded — FromScan then builds the target's own star graph only.
type ScanData struct {
	Domain       string
	IPs          []string
	Nameservers  []string
	Fingerprints artifact.Fingerprints
	Wallets      artifact.WalletArtifacts
	Redirects    []artifact.Redirect
	Correlation  artifact.Correlation
}

// FromScan builds a Graph from a single scan. It always includes the
// target's own directly observed infrastructure (IPs, nameservers,
// certificate, favicon, scripts, wallets, redirect chain); when
// Correlation carries clusters, it also connects each cluster's
// RelatedDomains and RelatedWallets to the same shared node the
// target matched through, so the graph shows the whole cluster, not
// just the one target.
//
// Telegram indicators are not graphed: the spec defines a Telegram
// node type but no corresponding edge type, and inventing one here
// would be unsupported by the spec's vocabulary.
func FromScan(data ScanData) Graph {
	b := NewBuilder()

	domainID := domainNodeID(data.Domain)
	b.AddNode(domainID, NodeDomain, data.Domain)

	for _, ip := range data.IPs {
		id := ipNodeID(ip)
		b.AddNode(id, NodeIP, ip)
		b.AddEdge(domainID, id, EdgeResolvesTo)
	}

	for _, ns := range data.Nameservers {
		id := nameserverNodeID(ns)
		b.AddNode(id, NodeNameserver, ns)
		b.AddEdge(domainID, id, EdgeSharesNameserver)
	}

	if data.Fingerprints.Certificate != "" {
		id := certificateNodeID(data.Fingerprints.Certificate)
		b.AddNode(id, NodeCertificate, shortHash(data.Fingerprints.Certificate))
		b.AddEdge(domainID, id, EdgeUsesCertificate)
	}

	if data.Fingerprints.Favicon != "" {
		id := faviconNodeID(data.Fingerprints.Favicon)
		b.AddNode(id, NodeFavicon, shortHash(data.Fingerprints.Favicon))
		b.AddEdge(domainID, id, EdgeSharesFavicon)
	}

	for _, hash := range data.Fingerprints.JavaScript {
		id := javascriptNodeID(hash)
		b.AddNode(id, NodeJavaScript, shortHash(hash))
		b.AddEdge(domainID, id, EdgeReusesScript)
	}

	for _, w := range data.Wallets.Addresses {
		id := walletNodeID(w.Address)
		b.AddNode(id, NodeWallet, fmt.Sprintf("%s: %s", w.Chain, w.Address))
		b.AddEdge(domainID, id, EdgeContainsWallet)
	}

	for _, r := range data.Redirects {
		id := urlNodeID(r.To)
		b.AddNode(id, NodeURL, r.To)
		b.AddEdge(domainID, id, EdgeRedirectsTo)
	}

	addClusters(b, domainID, data)

	return b.Build()
}

// addClusters connects each matched cluster's known sibling domains
// and wallets to the shared node the target matched through, so
// viewing the graph shows the whole infrastructure cluster rather
// than an isolated star around the single scanned target.
func addClusters(b *Builder, domainID string, data ScanData) {
	for _, cluster := range data.Correlation.Clusters {
		for _, match := range cluster.Matches {
			hubID, hubType, hubLabel, edgeType, ok := hubFor(match, data.Fingerprints)
			if !ok {
				continue
			}
			b.AddNode(hubID, hubType, hubLabel)
			b.AddEdge(domainID, hubID, edgeType)

			for _, sibling := range cluster.RelatedDomains {
				siblingID := domainNodeID(sibling)
				b.AddNode(siblingID, NodeDomain, sibling)
				b.AddEdge(siblingID, hubID, edgeType)
			}
		}

		for _, addr := range cluster.RelatedWallets {
			id := walletNodeID(addr)
			b.AddNode(id, NodeWallet, addr)
			// A related wallet from the cluster's signature, not
			// necessarily one seen on the target's own page — connect
			// it to the target domain directly so it's still part of
			// the visualized cluster.
			b.AddEdge(domainID, id, EdgeContainsWallet)
		}
	}
}

// hubFor resolves the shared node a Match refers to. For favicon and
// certificate matches it anchors on the target's own fingerprint value
// rather than match.Value, since a favicon match may have fired via
// either its SHA256 or MurmurHash3 sub-check — both refer to the same
// favicon, and the SHA256-keyed node built directly from Fingerprints
// is the one canonical identity for it.
func hubFor(m artifact.Match, fp artifact.Fingerprints) (id string, t NodeType, label string, e EdgeType, ok bool) {
	switch m.Category {
	case artifact.MatchFavicon:
		if fp.Favicon == "" {
			return "", "", "", "", false
		}
		return faviconNodeID(fp.Favicon), NodeFavicon, shortHash(fp.Favicon), EdgeSharesFavicon, true
	case artifact.MatchJavaScript:
		return javascriptNodeID(m.Value), NodeJavaScript, shortHash(m.Value), EdgeReusesScript, true
	case artifact.MatchCertificate:
		return certificateNodeID(m.Value), NodeCertificate, shortHash(m.Value), EdgeUsesCertificate, true
	case artifact.MatchWallet:
		return walletNodeID(m.Value), NodeWallet, m.Value, EdgeContainsWallet, true
	case artifact.MatchIP:
		return ipNodeID(m.Value), NodeIP, m.Value, EdgeResolvesTo, true
	case artifact.MatchNameserver:
		return nameserverNodeID(m.Value), NodeNameserver, m.Value, EdgeSharesNameserver, true
	default:
		// Domain and redirect matches have no separate hub node — the
		// matched domain itself is already the related entity.
		return "", "", "", "", false
	}
}

func domainNodeID(d string) string      { return "domain:" + d }
func ipNodeID(ip string) string         { return "ip:" + ip }
func nameserverNodeID(ns string) string { return "nameserver:" + ns }
func certificateNodeID(h string) string { return "certificate:" + h }
func faviconNodeID(h string) string     { return "favicon:" + h }
func javascriptNodeID(h string) string  { return "javascript:" + h }
func walletNodeID(addr string) string   { return "wallet:" + addr }
func urlNodeID(u string) string         { return "url:" + u }

// shortHash returns a short display label for a long hex hash,
// keeping node labels readable in rendered graphs.
func shortHash(h string) string {
	if len(h) <= 12 {
		return h
	}
	return h[:12] + "…"
}
