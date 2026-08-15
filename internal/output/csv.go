package output

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// RenderCSV renders r's evidence items as CSV, one row per item, for
// bulk processing across many scans.
func RenderCSV(r ScanResult) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)

	header := []string{"target", "category", "strength", "value", "signature_id", "description", "confidence", "recommendation"}
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("writing CSV header: %w", err)
	}

	if len(r.Evidence) == 0 {
		row := []string{r.Target, "", "", "", "", "", r.Confidence, r.Recommendation}
		if err := w.Write(row); err != nil {
			return "", fmt.Errorf("writing CSV row: %w", err)
		}
	}
	for _, item := range r.Evidence {
		row := []string{
			r.Target,
			string(item.Category),
			string(item.Strength),
			item.Value,
			item.SignatureID,
			item.Description,
			r.Confidence,
			r.Recommendation,
		}
		if err := w.Write(row); err != nil {
			return "", fmt.Errorf("writing CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flushing CSV: %w", err)
	}
	return b.String(), nil
}
