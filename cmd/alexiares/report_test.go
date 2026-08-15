package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/output"
)

func writeSampleResult(t *testing.T) string {
	t.Helper()
	result := output.ScanResult{
		Target:     "https://phish.example",
		Confidence: "high",
		Evidence: []evidence.Item{
			{Category: "favicon", Strength: evidence.Strong, Value: "-1", SignatureID: "cluster001", Description: "Shared favicon hash"},
		},
		Recommendation: "Avoid wallet interaction. Do not connect, sign, or approve.",
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshaling sample result: %v", err)
	}
	path := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("writing sample result: %v", err)
	}
	return path
}

func TestReportCommandDefaultFormat(t *testing.T) {
	path := writeSampleResult(t)
	stdout, _, err := execute(t, "report", path)
	if err != nil {
		t.Fatalf("execute(report) error = %v", err)
	}
	if !strings.Contains(stdout, "phish.example") || !strings.Contains(stdout, "Shared favicon hash") {
		t.Errorf("stdout = %q, missing expected content", stdout)
	}
}

func TestReportCommandJSONFormat(t *testing.T) {
	path := writeSampleResult(t)
	stdout, _, err := execute(t, "report", path, "--format", "json")
	if err != nil {
		t.Fatalf("execute(report --format json) error = %v", err)
	}
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(stdout), &parsed); jsonErr != nil {
		t.Fatalf("report --format json did not produce valid JSON: %v\noutput: %s", jsonErr, stdout)
	}
}

func TestReportCommandMarkdownFormat(t *testing.T) {
	path := writeSampleResult(t)
	stdout, _, err := execute(t, "report", path, "--format", "markdown")
	if err != nil {
		t.Fatalf("execute(report --format markdown) error = %v", err)
	}
	if !strings.HasPrefix(stdout, "# Alexiares Report") {
		t.Errorf("stdout = %q, want a Markdown report heading", stdout)
	}
}

func TestReportCommandMissingFile(t *testing.T) {
	_, _, err := execute(t, "report", filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("execute(report, missing file): error = nil, want error")
	}
	if !strings.Contains(err.Error(), "reading") {
		t.Errorf("error = %q, want mention of the read failure", err)
	}
}

func TestReportCommandInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json at all"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, _, err := execute(t, "report", path)
	if err == nil {
		t.Fatal("execute(report, invalid JSON): error = nil, want error")
	}
	if !strings.Contains(err.Error(), "not a valid scan result") {
		t.Errorf("error = %q, want mention of invalid scan result", err)
	}
}

func TestReportCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "report"); err == nil {
		t.Error("execute(report) with no args: error = nil, want arg-count error")
	}
}
