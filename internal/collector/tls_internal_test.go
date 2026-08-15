package collector

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

func selfSignedCert(t *testing.T, pub, priv any) *x509.Certificate {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(42),
		Subject:      pkix.Name{CommonName: "test.example"},
		Issuer:       pkix.Name{CommonName: "test.example"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"test.example"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate() error = %v", err)
	}
	return cert
}

func TestTLSDataNilConnectionState(t *testing.T) {
	if got := tlsData(nil); got != nil {
		t.Errorf("tlsData(nil) = %+v, want nil", got)
	}
	if got := tlsData(&tls.ConnectionState{}); got != nil {
		t.Errorf("tlsData(empty) = %+v, want nil", got)
	}
}

func TestTLSDataExtractsFields(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	cert := selfSignedCert(t, &priv.PublicKey, priv)

	got := tlsData(&tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}})
	if got == nil {
		t.Fatal("tlsData() = nil, want populated TLSData")
	}
	if len(got.SHA256) != 64 { // hex-encoded SHA-256
		t.Errorf("SHA256 = %q, want 64 hex chars", got.SHA256)
	}
	if got.SerialHex != "2a" {
		t.Errorf("SerialHex = %q, want 2a", got.SerialHex)
	}
	if got.KeyBits != 2048 {
		t.Errorf("KeyBits = %d, want 2048", got.KeyBits)
	}
	if len(got.DNSNames) != 1 || got.DNSNames[0] != "test.example" {
		t.Errorf("DNSNames = %v, want [test.example]", got.DNSNames)
	}
}

func TestPublicKeyBits(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey() error = %v", err)
	}
	edPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() error = %v", err)
	}

	tests := []struct {
		name string
		pub  any
		want int
	}{
		{"rsa 2048", &rsaKey.PublicKey, 2048},
		{"ecdsa p256", &ecKey.PublicKey, 256},
		{"ed25519", edPub, 256},
		{"unknown type", "not a key", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publicKeyBits(tt.pub); got != tt.want {
				t.Errorf("publicKeyBits(%s) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}
