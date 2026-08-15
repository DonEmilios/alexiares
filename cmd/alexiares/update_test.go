package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildSignedArchive returns a gzipped tar archive containing one
// signature file plus its Ed25519 signature, and the hex-encoded
// public key that verifies it — everything writeUpdateConfig and the
// test server below need to stand in for a real signed update.
func buildSignedArchive(t *testing.T) (archive, sig []byte, pubHex string) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	content := "id: cli_test_signature\ndescription: CLI update test\ndomains: [cli-update-test.example]\nconfidence: low\n"
	hdr := &tar.Header{Name: "infrastructure/cli_test.yaml", Mode: 0o644, Size: int64(len(content))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader() error = %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("tar Write() error = %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar Close() error = %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close() error = %v", err)
	}

	archive = buf.Bytes()
	sig = ed25519.Sign(priv, archive)
	return archive, sig, hex.EncodeToString(pub)
}

func writeUpdateConfig(t *testing.T, sourceURL, signaturesDest, publicKeyHex string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "network:\n  timeout: 5\nsignatures:\n  path: " + signaturesDest +
		"\nupdate:\n  source_url: " + sourceURL + "\n  public_key: " + publicKeyHex + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("writing update config: %v", err)
	}
	return path
}

func TestUpdateCommandRequiresPublicKey(t *testing.T) {
	cfgFile := writeConfig(t, t.TempDir()) // no update section at all
	_, _, err := execute(t, "--config", cfgFile, "update")
	if err == nil {
		t.Fatal("execute(update) with no public key: error = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "public_key") {
		t.Errorf("error = %q, want mention of the missing public key", err)
	}
}

func TestUpdateCommandInstallsSignedArchive(t *testing.T) {
	archive, sig, pubHex := buildSignedArchive(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/signatures.tar.gz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/signatures.tar.gz.sig", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(sig)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "signatures")
	cfgFile := writeUpdateConfig(t, srv.URL, dest, pubHex)

	stdout, _, err := execute(t, "--config", cfgFile, "update")
	if err != nil {
		t.Fatalf("execute(update) error = %v", err)
	}
	if !strings.Contains(stdout, "Signatures updated") {
		t.Errorf("stdout = %q, want a confirmation message", stdout)
	}

	installed := filepath.Join(dest, "infrastructure", "cli_test.yaml")
	if _, statErr := os.Stat(installed); statErr != nil {
		t.Errorf("expected signature file missing after update: %v", statErr)
	}
}

func TestUpdateCommandRejectsWrongKey(t *testing.T) {
	archive, sig, _ := buildSignedArchive(t)
	_, _, wrongPubHex := buildSignedArchive(t) // a different keypair's public key

	mux := http.NewServeMux()
	mux.HandleFunc("/signatures.tar.gz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/signatures.tar.gz.sig", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(sig)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "signatures")
	cfgFile := writeUpdateConfig(t, srv.URL, dest, wrongPubHex)

	_, _, err := execute(t, "--config", cfgFile, "update")
	if err == nil {
		t.Fatal("execute(update) with mismatched key: error = nil, want rejection")
	}

	if _, statErr := os.Stat(dest); statErr == nil {
		t.Error("rejected update was still installed to disk")
	}
}

func TestUpdateCommandRequiresNoArgs(t *testing.T) {
	if _, _, err := execute(t, "update", "unexpected-arg"); err == nil {
		t.Error("execute(update, extra arg): error = nil, want arg-count error")
	}
}
