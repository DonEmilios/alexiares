package update_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexiares/alexiares/internal/update"
)

func TestVerifyAcceptsValidSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	data := []byte("signature repository archive bytes")
	sig := ed25519.Sign(priv, data)

	if err := update.Verify(data, sig, pub); err != nil {
		t.Errorf("Verify() error = %v, want nil for a valid signature", err)
	}
}

func TestVerifyRejectsTamperedData(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	sig := ed25519.Sign(priv, []byte("original data"))

	if err := update.Verify([]byte("tampered data"), sig, pub); err == nil {
		t.Error("Verify() error = nil, want rejection of tampered data")
	}
}

func TestVerifyRejectsWrongKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
	data := []byte("archive bytes")
	sig := ed25519.Sign(priv, data)

	if err := update.Verify(data, sig, otherPub); err == nil {
		t.Error("Verify() error = nil, want rejection when signed with a different key")
	}
}

func TestVerifyRejectsMalformedSignatureOrKey(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	if err := update.Verify([]byte("data"), []byte("too short"), pub); err == nil {
		t.Error("Verify() error = nil, want rejection of malformed signature")
	}
	if err := update.Verify([]byte("data"), make([]byte, ed25519.SignatureSize), []byte("short key")); err == nil {
		t.Error("Verify() error = nil, want rejection of malformed public key")
	}
}

func TestRunAppliesOnlyASignedUpdate(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	archive := buildTarGz(t, map[string]string{"new.yaml": "fresh signature"})
	sig := ed25519.Sign(priv, archive)

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
	err = update.Run(context.Background(), update.Options{
		SourceURL: srv.URL,
		PublicKey: pub,
		Timeout:   2 * time.Second,
	}, dest)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "new.yaml")); err != nil {
		t.Errorf("expected file missing after Run(): %v", err)
	}
}

func TestRunRejectsUnsignedUpdate(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	_, wrongPriv, _ := ed25519.GenerateKey(rand.Reader)

	archive := buildTarGz(t, map[string]string{"new.yaml": "should not be installed"})
	wrongSig := ed25519.Sign(wrongPriv, archive) // signed with a key the server doesn't claim

	mux := http.NewServeMux()
	mux.HandleFunc("/signatures.tar.gz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/signatures.tar.gz.sig", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(wrongSig)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "signatures")
	err := update.Run(context.Background(), update.Options{
		SourceURL: srv.URL,
		PublicKey: pub,
		Timeout:   2 * time.Second,
	}, dest)
	if err == nil {
		t.Fatal("Run() error = nil, want rejection of an update signed with the wrong key")
	}

	if _, statErr := os.Stat(filepath.Join(dest, "new.yaml")); statErr == nil {
		t.Error("rejected update was still applied to disk")
	}
}

func TestRunRequiresPublicKey(t *testing.T) {
	err := update.Run(context.Background(), update.Options{SourceURL: "https://example.invalid"}, t.TempDir())
	if err != update.ErrNoPublicKey {
		t.Errorf("Run() error = %v, want ErrNoPublicKey", err)
	}
}

func TestRunRequiresSourceURL(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	err := update.Run(context.Background(), update.Options{PublicKey: pub}, t.TempDir())
	if err == nil {
		t.Error("Run() error = nil, want error for missing source URL")
	}
}
