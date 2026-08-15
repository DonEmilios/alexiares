package output

import (
	"encoding/json"
	"fmt"
)

// RenderJSON renders r as indented JSON matching the spec's JSON
// Output Schema.
func RenderJSON(r ScanResult) (string, error) {
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling scan result JSON: %w", err)
	}
	return string(out) + "\n", nil
}
