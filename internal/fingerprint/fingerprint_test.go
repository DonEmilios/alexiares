package fingerprint

import (
	"reflect"
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
)

func TestComputeIsDeterministic(t *testing.T) {
	raw := artifact.RawResponse{
		HTML: `<html><head><title>Claim</title></head><body><form><input name="w"></form></body></html>`,
		TLS:  &artifact.TLSData{SHA256: "deadbeef"},
	}
	fav := artifact.FaviconArtifacts{SHA256: "abc123", Murmur3: -12345}
	js := artifact.JavaScriptArtifacts{SHA256: []string{"hash1", "hash2"}}

	first := Compute(raw, fav, js)
	second := Compute(raw, fav, js)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Compute() not deterministic:\n  first:  %+v\n  second: %+v", first, second)
	}
}

func TestComputeAggregatesAllSources(t *testing.T) {
	raw := artifact.RawResponse{
		HTML: `<html><body></body></html>`,
		TLS:  &artifact.TLSData{SHA256: "cert-sha"},
	}
	fav := artifact.FaviconArtifacts{SHA256: "fav-sha", Murmur3: 42}
	js := artifact.JavaScriptArtifacts{SHA256: []string{"js-sha"}}

	got := Compute(raw, fav, js)

	if got.Favicon != "fav-sha" {
		t.Errorf("Favicon = %q, want fav-sha", got.Favicon)
	}
	if got.FaviconHash != 42 {
		t.Errorf("FaviconHash = %d, want 42", got.FaviconHash)
	}
	if len(got.JavaScript) != 1 || got.JavaScript[0] != "js-sha" {
		t.Errorf("JavaScript = %v, want [js-sha]", got.JavaScript)
	}
	if got.Certificate != "cert-sha" {
		t.Errorf("Certificate = %q, want cert-sha", got.Certificate)
	}
	if got.HTML == "" {
		t.Error("HTML structural hash is empty, want populated")
	}
	if got.HTMLSimilarity == "" {
		t.Error("HTMLSimilarity is empty, want populated")
	}
}

func TestComputeWithoutTLSLeavesCertificateEmpty(t *testing.T) {
	raw := artifact.RawResponse{HTML: `<html></html>`}
	got := Compute(raw, artifact.FaviconArtifacts{}, artifact.JavaScriptArtifacts{})
	if got.Certificate != "" {
		t.Errorf("Certificate = %q, want empty when TLS is nil", got.Certificate)
	}
}
