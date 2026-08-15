package output_test

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/graph"
	"github.com/alexiares/alexiares/internal/output"
)

func sampleResult() output.ScanResult {
	return output.ScanResult{
		Target:     "https://phish.example",
		Confidence: "high",
		Evidence: []evidence.Item{
			{Category: "favicon", Strength: evidence.Strong, Value: "-204998123", SignatureID: "cluster001", Description: "Shared favicon hash"},
		},
		Recommendation: "Avoid wallet interaction. Do not connect, sign, or approve.",
		RelatedDomains: []string{"sibling.example"},
		RelatedWallets: []string{"0xDRAINER"},
		Graph: func() graph.Graph {
			b := graph.NewBuilder()
			b.AddNode("domain:phish.example", graph.NodeDomain, "phish.example")
			return b.Build()
		}(),
	}
}

func TestRenderTerminalAnswersTheFourQuestions(t *testing.T) {
	out := output.RenderTerminal(sampleResult())

	checks := []string{
		"phish.example",            // what was scanned
		"Shared favicon hash",      // what was found / why suspicious
		"cluster001",               // matched signature
		"sibling.example",          // related infrastructure
		"HIGH",                     // confidence
		"Avoid wallet interaction", // recommendation
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("RenderTerminal() missing %q in output:\n%s", want, out)
		}
	}
}

func TestRenderTerminalCleanTarget(t *testing.T) {
	out := output.RenderTerminal(output.ScanResult{Target: "clean.example"})
	if !strings.Contains(out, "No known malicious infrastructure detected") {
		t.Errorf("RenderTerminal(clean) = %q, want no-detection message", out)
	}
}

func TestRenderJSONMatchesSchema(t *testing.T) {
	out, err := output.RenderJSON(sampleResult())
	if err != nil {
		t.Fatalf("RenderJSON() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("RenderJSON() produced invalid JSON: %v", err)
	}

	for _, key := range []string{"target", "confidence", "evidence", "artifacts", "graph", "recommendation"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("RenderJSON() missing spec-required key %q; got keys: %v", key, keysOf(parsed))
		}
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestRenderCSVHasHeaderAndRow(t *testing.T) {
	out, err := output.RenderCSV(sampleResult())
	if err != nil {
		t.Fatalf("RenderCSV() error = %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("parsing RenderCSV() output: %v", err)
	}
	if len(rows) != 2 { // header + 1 evidence item
		t.Fatalf("RenderCSV() = %d rows, want 2 (header + 1 item)", len(rows))
	}
	if rows[0][0] != "target" {
		t.Errorf("CSV header[0] = %q, want target", rows[0][0])
	}
	if rows[1][0] != "https://phish.example" {
		t.Errorf("CSV row[0] = %q, want target value", rows[1][0])
	}
}

func TestRenderCSVCleanTargetStillProducesRow(t *testing.T) {
	out, err := output.RenderCSV(output.ScanResult{Target: "clean.example"})
	if err != nil {
		t.Fatalf("RenderCSV() error = %v", err)
	}
	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("parsing CSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("RenderCSV(clean) = %d rows, want 2 (header + empty-evidence row)", len(rows))
	}
}

func TestRenderMarkdownContainsTable(t *testing.T) {
	out := output.RenderMarkdown(sampleResult())
	if !strings.Contains(out, "| Strength | Category | Detection | Signature | Value |") {
		t.Errorf("RenderMarkdown() missing evidence table header:\n%s", out)
	}
	if !strings.Contains(out, "phish.example") {
		t.Errorf("RenderMarkdown() missing target:\n%s", out)
	}
}

func TestRenderDispatchesAllFormats(t *testing.T) {
	r := sampleResult()
	formats := []output.Format{output.FormatTerminal, output.FormatJSON, output.FormatGraphML, output.FormatDOT, output.FormatCSV, output.FormatMarkdown, ""}
	for _, f := range formats {
		if _, err := output.Render(f, r); err != nil {
			t.Errorf("Render(%q) error = %v", f, err)
		}
	}
}

func TestRenderUnsupportedFormatErrors(t *testing.T) {
	if _, err := output.Render("yaml", sampleResult()); err == nil {
		t.Error("Render(yaml) error = nil, want error for unsupported format")
	}
}
