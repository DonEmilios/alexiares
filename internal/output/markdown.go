package output

import (
	"fmt"
	"strings"
)

// RenderMarkdown renders r as a Markdown report, suitable for pasting
// into an issue, PR, or incident writeup.
func RenderMarkdown(r ScanResult) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "# Alexiares Report: %s\n\n", r.Target)

	confidence := r.Confidence
	if confidence == "" {
		confidence = evidenceNoneLabel
	}
	_, _ = fmt.Fprintf(&b, "**Confidence:** %s\n\n", strings.ToUpper(confidence))
	_, _ = fmt.Fprintf(&b, "**Recommendation:** %s\n\n", r.Recommendation)

	if len(r.Evidence) == 0 {
		b.WriteString("No known malicious infrastructure detected.\n")
		return b.String()
	}

	b.WriteString("## Evidence\n\n")
	b.WriteString("| Strength | Category | Detection | Signature | Value |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, item := range r.Evidence {
		_, _ = fmt.Fprintf(&b, "| %s | %s | %s | `%s` | `%s` |\n",
			strings.ToUpper(string(item.Strength)), item.Category, item.Description, item.SignatureID, item.Value)
	}
	b.WriteByte('\n')

	if len(r.RelatedDomains) > 0 {
		b.WriteString("## Related Domains\n\n")
		for _, d := range r.RelatedDomains {
			_, _ = fmt.Fprintf(&b, "- %s\n", d)
		}
		b.WriteByte('\n')
	}

	if len(r.RelatedWallets) > 0 {
		b.WriteString("## Related Wallets\n\n")
		for _, w := range r.RelatedWallets {
			_, _ = fmt.Fprintf(&b, "- `%s`\n", w)
		}
		b.WriteByte('\n')
	}

	return b.String()
}
