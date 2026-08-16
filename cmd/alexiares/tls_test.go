package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexiares/alexiares/internal/artifact"
)

func TestPrintTLSCertificateNilData(t *testing.T) {
	var b strings.Builder
	if err := printTLSCertificate(&b, "https://plain.example", nil); err != nil {
		t.Fatalf("printTLSCertificate() error = %v", err)
	}
	if !strings.Contains(b.String(), "did not present a TLS certificate") {
		t.Errorf("output = %q, want the no-certificate message", b.String())
	}
}

func TestPrintTLSCertificatePopulated(t *testing.T) {
	data := &artifact.TLSData{
		SHA256:    "abc123",
		SerialHex: "2a",
		Issuer:    "CN=Test CA",
		Subject:   "CN=phish.example",
		DNSNames:  []string{"phish.example", "*.phish.example"},
		NotBefore: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		KeyType:   "ECDSA",
		KeyBits:   256,
	}
	var b strings.Builder
	if err := printTLSCertificate(&b, "https://phish.example", data); err != nil {
		t.Fatalf("printTLSCertificate() error = %v", err)
	}
	out := b.String()

	for _, want := range []string{"abc123", "2a", "CN=Test CA", "CN=phish.example", "ECDSA (256 bits)", "2026-01-01", "2026-04-01", "phish.example", "*.phish.example"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestTLSCommandRejectsWalletTarget(t *testing.T) {
	_, _, err := execute(t, "tls", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err == nil {
		t.Fatal("execute(tls, wallet address): error = nil, want rejection before any network call")
	}
	if !strings.Contains(err.Error(), "not a fetchable domain, URL, or IP") {
		t.Errorf("error = %q, want the pre-network rejection message", err)
	}
}

func TestTLSCommandNoCertificateOverPlainHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "--config", writeNetworkTestConfig(t), "tls", srv.URL)
	if err != nil {
		t.Fatalf("execute(tls) error = %v", err)
	}
	if !strings.Contains(stdout, "did not present a TLS certificate") {
		t.Errorf("stdout = %q, want the no-certificate message for a plain HTTP target", stdout)
	}
}

func TestTLSCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "tls"); err == nil {
		t.Error("execute(tls) with no args: error = nil, want arg-count error")
	}
}
