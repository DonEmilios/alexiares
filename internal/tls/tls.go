// Package tls extracts certificate intelligence from a completed TLS
// handshake: fingerprint, serial, issuer, subject, SAN, validity
// period, and key type. It performs no network I/O of its own — it
// operates on the *tls.ConnectionState the collector already captured.
package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"fmt"

	"github.com/alexiares/alexiares/internal/artifact"
)

// FromConnectionState converts a completed TLS handshake's connection
// state into artifact.TLSData. It returns nil if state is nil or
// carries no peer certificate (a plain HTTP response).
func FromConnectionState(state *tls.ConnectionState) *artifact.TLSData {
	if state == nil || len(state.PeerCertificates) == 0 {
		return nil
	}
	cert := state.PeerCertificates[0]

	return &artifact.TLSData{
		SHA256:    fmt.Sprintf("%x", sha256.Sum256(cert.Raw)),
		SerialHex: cert.SerialNumber.Text(16),
		Issuer:    cert.Issuer.String(),
		Subject:   cert.Subject.String(),
		DNSNames:  cert.DNSNames,
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		KeyType:   cert.PublicKeyAlgorithm.String(),
		KeyBits:   publicKeyBits(cert.PublicKey),
	}
}

// publicKeyBits reports the key size in bits for the certificate
// public key types Go's x509 package can produce. Unrecognized key
// types report 0 rather than guessing.
func publicKeyBits(pub any) int {
	switch k := pub.(type) {
	case *rsa.PublicKey:
		return k.N.BitLen()
	case *ecdsa.PublicKey:
		return k.Curve.Params().BitSize
	case ed25519.PublicKey:
		return len(k) * 8
	default:
		return 0
	}
}
