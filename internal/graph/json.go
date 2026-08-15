package graph

import (
	"encoding/json"
	"fmt"
)

// WriteJSON renders g as indented JSON. Graph's fields already carry
// json tags, so this is a thin wrapper — it exists so callers can
// select DOT, GraphML, or JSON output through one consistent
// function-per-format interface.
func WriteJSON(g Graph) (string, error) {
	out, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling graph JSON: %w", err)
	}
	return string(out) + "\n", nil
}
