package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoSignaturesDir locates this repository's signatures/ directory
// relative to this test file, so tests can scan against the real
// bundled example signature without depending on the working
// directory the test runner happens to use.
func repoSignaturesDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve test file path")
	}
	// This file lives at <repo>/cmd/alexiares/scan_test.go.
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "signatures")
}

func TestScanCommandWalletTarget(t *testing.T) {
	stdout, _, err := execute(t, "scan", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err != nil {
		t.Fatalf("execute(scan, wallet) error = %v", err)
	}
	if !strings.Contains(stdout, "chain: ethereum") {
		t.Errorf("stdout = %q, want the classified chain", stdout)
	}
}

func TestScanCommandUnrecognizedTarget(t *testing.T) {
	_, _, err := execute(t, "scan", "not a url or wallet or anything")
	if err == nil {
		t.Fatal("execute(scan, garbage): error = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "not a recognized URL, domain, IP, or wallet address") {
		t.Errorf("error = %q, want the classification error", err)
	}
}

func TestScanCommandCleanTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html><body>Nothing suspicious here.</body></html>"))
	}))
	defer srv.Close()

	cfgFile := writeConfig(t, t.TempDir()) // empty signatures directory
	stdout, _, err := execute(t, "--config", cfgFile, "scan", srv.URL)
	if err != nil {
		t.Fatalf("execute(scan, clean target) error = %v", err)
	}
	if !strings.Contains(stdout, "No known malicious infrastructure detected") {
		t.Errorf("stdout = %q, want the no-detection message", stdout)
	}
}

func TestScanCommandDetectsBundledSignature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body>
			<script>fetch("https://api.telegram.org/bot123456789:AAExampleTokenValue1234567890123456/sendMessage")</script>
		</body></html>`))
	}))
	defer srv.Close()

	cfgFile := writeConfig(t, repoSignaturesDir(t))
	stdout, _, err := execute(t, "--config", cfgFile, "scan", srv.URL)
	if err != nil {
		t.Fatalf("execute(scan, phishing page) error = %v", err)
	}

	for _, want := range []string{"STRONG", "Telegram exfiltration indicator", "wallet_drainer_cluster_001", "HIGH", "Avoid wallet interaction"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestScanCommandJSONFormatMatchesSchema(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	cfgFile := writeConfig(t, t.TempDir())
	stdout, _, err := execute(t, "--config", cfgFile, "scan", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("execute(scan --format json) error = %v", err)
	}
	for _, key := range []string{`"target"`, `"confidence"`, `"evidence"`, `"artifacts"`, `"graph"`, `"recommendation"`} {
		if !strings.Contains(stdout, key) {
			t.Errorf("JSON output missing key %s:\n%s", key, stdout)
		}
	}
}

func TestScanCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "scan"); err == nil {
		t.Error("execute(scan) with no args: error = nil, want arg-count error")
	}
}
