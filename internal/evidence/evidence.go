// Package evidence turns a correlation result into a human-readable
// conclusion: what was found, why it's suspicious, how confident
// Alexiares is, and what to do about it. It performs no detection of
// its own — internal/correlation already decided what matched — it
// only evaluates, weighs, and explains.
package evidence

import (
	"sort"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/intel"
)

// Strength is how strong a single piece of evidence is, independent
// of the signature's own overall confidence — a shared favicon is
// inherently stronger proof of shared infrastructure than a shared IP
// address ever is, regardless of which signature it came from.
type Strength string

// Evidence strength tiers, per the spec's Evidence Categories.
const (
	Strong Strength = "strong"
	Medium Strength = "medium"
	Weak   Strength = "weak"
)

// strengthOf classifies a match category into the spec's Evidence
// Categories table. Domain matches (the target's own domain, or a
// redirect hop, appearing in a signature's known-domain list) are
// treated as strong: it is direct identity, not inference.
func strengthOf(c artifact.MatchCategory) Strength {
	switch c {
	case artifact.MatchFavicon, artifact.MatchJavaScript, artifact.MatchCertificate,
		artifact.MatchWallet, artifact.MatchTelegram, artifact.MatchDomain:
		return Strong
	case artifact.MatchIP, artifact.MatchNameserver, artifact.MatchRedirect:
		return Medium
	default:
		return Weak
	}
}

// Item is one piece of evidence: what matched, how strong it is, and
// which signature it came from.
type Item struct {
	Category    artifact.MatchCategory `json:"category"`
	Strength    Strength               `json:"strength"`
	Value       string                 `json:"value"`
	SignatureID string                 `json:"signature_id"`
	Description string                 `json:"description"`
}

// confidenceRank orders qualitative confidence so the strongest
// matched signature can decide the overall verdict.
var confidenceRank = map[intel.Confidence]int{
	intel.ConfidenceLow:      1,
	intel.ConfidenceMedium:   2,
	intel.ConfidenceHigh:     3,
	intel.ConfidenceCritical: 4,
}

// Report is Alexiares' final, explainable conclusion for one target.
type Report struct {
	Target         string   `json:"target"`
	Confidence     string   `json:"confidence"` // "" when nothing matched — there is no verdict to be confident in
	Evidence       []Item   `json:"evidence"`
	Recommendation string   `json:"recommendation"`
	RelatedDomains []string `json:"related_domains,omitempty"`
	RelatedWallets []string `json:"related_wallets,omitempty"`
}

// Evaluate builds a Report for target from its correlation result.
func Evaluate(target string, corr artifact.Correlation) Report {
	items := make([]Item, 0, len(corr.Matches))
	for _, m := range corr.Matches {
		items = append(items, Item{
			Category:    m.Category,
			Strength:    strengthOf(m.Category),
			Value:       m.Value,
			SignatureID: m.SignatureID,
			Description: describe(m.Category),
		})
	}
	sortItems(items)

	confidence := overallConfidence(corr.Clusters)

	return Report{
		Target:         target,
		Confidence:     confidence,
		Evidence:       items,
		Recommendation: recommend(confidence),
		RelatedDomains: corr.RelatedDomains,
		RelatedWallets: corr.RelatedWallets,
	}
}

// overallConfidence returns the highest confidence level among every
// matched cluster, or "" if none matched.
func overallConfidence(clusters []artifact.Cluster) string {
	best := ""
	bestRank := 0
	for _, c := range clusters {
		rank := confidenceRank[intel.Confidence(c.Confidence)]
		if rank > bestRank {
			bestRank = rank
			best = c.Confidence
		}
	}
	return best
}

// recommend maps a confidence level to an actionable recommendation.
func recommend(confidence string) string {
	switch intel.Confidence(confidence) {
	case intel.ConfidenceCritical, intel.ConfidenceHigh:
		return "Avoid wallet interaction. Do not connect, sign, or approve."
	case intel.ConfidenceMedium:
		return "Proceed with caution. Independently verify this infrastructure before interacting."
	case intel.ConfidenceLow:
		return "Weak signal only. No strong evidence of malicious infrastructure, but remain cautious."
	default:
		return "No known malicious infrastructure detected."
	}
}

// describe returns the human-readable label for a match category,
// matching the spec's "Detection" phrasing.
func describe(c artifact.MatchCategory) string {
	switch c {
	case artifact.MatchFavicon:
		return "Shared favicon hash"
	case artifact.MatchJavaScript:
		return "Shared JavaScript"
	case artifact.MatchCertificate:
		return "Shared TLS certificate"
	case artifact.MatchWallet:
		return "Known malicious wallet address"
	case artifact.MatchTelegram:
		return "Telegram exfiltration indicator"
	case artifact.MatchDomain:
		return "Known malicious domain"
	case artifact.MatchRedirect:
		return "Redirects to known malicious infrastructure"
	case artifact.MatchIP:
		return "Shared hosting IP"
	case artifact.MatchNameserver:
		return "Shared nameserver"
	default:
		return string(c)
	}
}

// sortItems orders evidence strong-first, then alphabetically within
// a strength tier, so output is deterministic and leads with the most
// convincing evidence.
func sortItems(items []Item) {
	rank := map[Strength]int{Strong: 0, Medium: 1, Weak: 2}
	sort.Slice(items, func(i, j int) bool {
		if rank[items[i].Strength] != rank[items[j].Strength] {
			return rank[items[i].Strength] < rank[items[j].Strength]
		}
		if items[i].Category != items[j].Category {
			return items[i].Category < items[j].Category
		}
		return items[i].Value < items[j].Value
	})
}
