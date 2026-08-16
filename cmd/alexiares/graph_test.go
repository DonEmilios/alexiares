package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGraphCommandDefaultsToDOT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "--config", writeNetworkTestConfig(t), "graph", srv.URL)
	if err != nil {
		t.Fatalf("execute(graph) error = %v", err)
	}
	if !strings.HasPrefix(stdout, "digraph alexiares {") {
		t.Errorf("stdout = %q, want DOT output by default", stdout)
	}
}

func TestGraphCommandJSONFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "--config", writeNetworkTestConfig(t), "graph", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("execute(graph --format json) error = %v", err)
	}
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(stdout), &parsed); jsonErr != nil {
		t.Fatalf("graph --format json did not produce valid JSON: %v\noutput: %s", jsonErr, stdout)
	}
	if _, ok := parsed["nodes"]; !ok {
		t.Errorf("JSON output missing \"nodes\" key: %s", stdout)
	}
}

func TestGraphCommandGraphMLFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "--config", writeNetworkTestConfig(t), "graph", srv.URL, "--format", "graphml")
	if err != nil {
		t.Fatalf("execute(graph --format graphml) error = %v", err)
	}
	var doc any
	if xmlErr := xml.Unmarshal([]byte(stdout), &doc); xmlErr != nil {
		t.Fatalf("graph --format graphml did not produce valid XML: %v\noutput: %s", xmlErr, stdout)
	}
}

func TestGraphCommandUnsupportedFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	_, _, err := execute(t, "--config", writeNetworkTestConfig(t), "graph", srv.URL, "--format", "yaml")
	if err == nil {
		t.Fatal("execute(graph --format yaml): error = nil, want rejection of unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error = %q, want mention of unsupported format", err)
	}
}

func TestGraphCommandRejectsWalletTarget(t *testing.T) {
	_, _, err := execute(t, "graph", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err == nil {
		t.Fatal("execute(graph, wallet address): error = nil, want rejection")
	}
}

func TestGraphCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "graph"); err == nil {
		t.Error("execute(graph) with no args: error = nil, want arg-count error")
	}
}
