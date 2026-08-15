# `internal/tls`

**Source:** [`internal/tls/tls.go`](../internal/tls/tls.go)
**Tests:** `tls_test.go` — 100% coverage
**Position in pipeline:** extractor stage — pure, no I/O

## Purpose

Turns a completed TLS handshake (`*crypto/tls.ConnectionState`, already captured by `internal/collector`) into `artifact.TLSData`: the certificate's SHA256 fingerprint, serial number, issuer, subject, SANs, validity window, and public key type/size.

```go
func FromConnectionState(state *tls.ConnectionState) *artifact.TLSData
```

Returns `nil` for a plain-HTTP target or an empty connection state — never an error. There's nothing to fail at; a missing certificate is a fact about the target, not an exceptional condition.

## Design notes

**This package used to be inline inside `internal/collector`.** It was extracted during the build specifically so it could be unit-tested with a real, locally-generated self-signed certificate (`crypto/x509.CreateCertificate` against a throwaway key), without needing a live TLS handshake in the test. `internal/collector`'s own tests deliberately don't cover the TLS-populated path for the same reason — `httptest.NewTLSServer`'s self-signed cert isn't trusted by a real `http.Client`, and there's no transport-injection point in `Collector` to work around that. Extracting the parsing logic into its own I/O-free package sidesteps the problem entirely: give it a `ConnectionState` value directly, no handshake required.

**Key size handles three concrete types, not an interface guess.** `publicKeyBits` type-switches on `*rsa.PublicKey`, `*ecdsa.PublicKey`, and `ed25519.PublicKey` explicitly — these are the only three public key types Go's `x509` package produces from a parsed certificate. An earlier draft tried a duck-typed `interface{ Size() int }` switch instead, which only actually works for RSA (the only one of the three with a `Size()` method in the standard library) and would have silently reported 0 bits for every ECDSA and Ed25519 certificate. Caught and rewritten before merge.

**The fingerprint is SHA256 of `cert.Raw`** — the full DER-encoded certificate bytes, not just the public key. This is the standard "certificate fingerprint" convention (what browsers show, what `openssl x509 -fingerprint` computes) and what `internal/correlation` matches signatures against.
