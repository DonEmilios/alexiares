package output

import (
	"fmt"
	"strings"
)

// RenderTerminal renders r as human-readable terminal text: what was
// found, why it's suspicious, what's related, and how confident
// Alexiares is — the four questions the spec requires every output to
// answer.
func RenderTerminal(r ScanResult) string {
	var b strings.Builder

	// strings.Builder's Write methods never return an error, so the
	// return values below are safe to discard.
	_, _ = fmt.Fprintf(&b, "Alexiares — infrastructure intelligence\n\n")
	_, _ = fmt.Fprintf(&b, "Target  %s\n\n", r.Target)

	if len(r.Evidence) == 0 {
		_, _ = fmt.Fprintf(&b, "No known malicious infrastructure detected.\n")
		return b.String()
	}

	_, _ = fmt.Fprintf(&b, "Evidence\n")
	for _, item := range r.Evidence {
		_, _ = fmt.Fprintf(&b, "  [%s]  %s\n", strings.ToUpper(string(item.Strength)), item.Description)
		_, _ = fmt.Fprintf(&b, "            → matches signature %s (%s: %s)\n", item.SignatureID, item.Category, item.Value)
	}
	b.WriteByte('\n')

	if len(r.RelatedDomains) > 0 || len(r.RelatedWallets) > 0 {
		_, _ = fmt.Fprintf(&b, "Related infrastructure\n")
		if len(r.RelatedDomains) > 0 {
			_, _ = fmt.Fprintf(&b, "  %s\n", strings.Join(r.RelatedDomains, ", "))
		}
		for _, w := range r.RelatedWallets {
			_, _ = fmt.Fprintf(&b, "  Wallet %s\n", w)
		}
		b.WriteByte('\n')
	}

	confidence := r.Confidence
	if confidence == "" {
		confidence = evidenceNoneLabel
	}
	_, _ = fmt.Fprintf(&b, "Confidence  %s\n\n", strings.ToUpper(confidence))
	_, _ = fmt.Fprintf(&b, "Recommendation\n  %s\n", r.Recommendation)

	return b.String()
}

const evidenceNoneLabel = "none"
