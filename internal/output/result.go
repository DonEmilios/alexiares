// Package output renders a completed scan as terminal text, JSON,
// GraphML, DOT, CSV, or Markdown. It performs no detection or
// evaluation of its own — internal/evidence has already decided what
// the conclusion is; this package only presents it.
package output

import (
	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/graph"
)

// Artifacts bundles every extractor's output for one target, so a
// saved JSON result carries everything needed to regenerate any other
// format later via `alexiares report`.
type Artifacts struct {
	DNS          artifact.DNSArtifacts        `json:"dns"`
	TLS          *artifact.TLSData            `json:"tls,omitempty"`
	HTML         artifact.HTMLArtifacts       `json:"html"`
	JavaScript   artifact.JavaScriptArtifacts `json:"javascript"`
	Favicon      artifact.FaviconArtifacts    `json:"favicon"`
	Wallets      artifact.WalletArtifacts     `json:"wallets"`
	Telegram     artifact.TelegramArtifacts   `json:"telegram"`
	Redirects    []artifact.Redirect          `json:"redirects,omitempty"`
	Fingerprints artifact.Fingerprints        `json:"fingerprints"`
}

// ScanResult is the complete result of one scan: the spec's JSON
// Output Schema (target, confidence, evidence, artifacts, graph,
// recommendation) plus the related-infrastructure lists the terminal
// and Markdown renderers present alongside them.
type ScanResult struct {
	Target         string          `json:"target"`
	Confidence     string          `json:"confidence"`
	Evidence       []evidence.Item `json:"evidence"`
	Artifacts      Artifacts       `json:"artifacts"`
	Graph          graph.Graph     `json:"graph"`
	Recommendation string          `json:"recommendation"`
	RelatedDomains []string        `json:"related_domains,omitempty"`
	RelatedWallets []string        `json:"related_wallets,omitempty"`
}

// FromReport builds a ScanResult from an evidence.Report plus the
// artifacts and graph gathered alongside it.
func FromReport(report evidence.Report, artifacts Artifacts, g graph.Graph) ScanResult {
	return ScanResult{
		Target:         report.Target,
		Confidence:     report.Confidence,
		Evidence:       report.Evidence,
		Artifacts:      artifacts,
		Graph:          g,
		Recommendation: report.Recommendation,
		RelatedDomains: report.RelatedDomains,
		RelatedWallets: report.RelatedWallets,
	}
}
