// Package update fetches and applies signed signature repository
// updates.
//
// An update is only ever applied after its Ed25519 signature verifies
// against the configured trusted public key. A missing, malformed, or
// invalid signature causes the update to be rejected outright — the
// existing local signature repository is left untouched. There is no
// fallback to unsigned intelligence: a failed update is a hard error,
// per the spec's requirement that updates be cryptographically signed.
package update

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Options configures how an update is fetched and verified.
type Options struct {
	// SourceURL is the base URL serving two files: "signatures.tar.gz"
	// and "signatures.tar.gz.sig" (a raw Ed25519 signature over the
	// archive bytes).
	SourceURL string
	// PublicKey is the trusted Ed25519 public key updates must be
	// signed with. Required — Run refuses to fetch anything without one.
	PublicKey ed25519.PublicKey
	Timeout   time.Duration
	UserAgent string
	// MaxBytes bounds how much of either downloaded file is read,
	// protecting against an unbounded or hostile response.
	MaxBytes int64
}

func (o Options) withDefaults() Options {
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}
	if o.UserAgent == "" {
		o.UserAgent = "Alexiares/0.1"
	}
	if o.MaxBytes <= 0 {
		o.MaxBytes = 20 << 20 // 20 MB
	}
	return o
}

// ErrNoPublicKey is returned when Options carries no trusted key to
// verify against — Alexiares never falls back to trusting an update
// on the strength of its source URL alone.
var ErrNoPublicKey = errors.New("update: no trusted public key configured")

// Run fetches the archive and signature from opts.SourceURL, verifies
// the signature, and, only on success, extracts the archive into
// destDir. destDir's prior contents are replaced atomically: nothing
// is left partially written if any step fails.
func Run(ctx context.Context, opts Options, destDir string) error {
	opts = opts.withDefaults()
	if len(opts.PublicKey) == 0 {
		return ErrNoPublicKey
	}
	if opts.SourceURL == "" {
		return errors.New("update: no source URL configured")
	}

	client := &http.Client{Timeout: opts.Timeout}

	archive, err := fetch(ctx, client, opts, opts.SourceURL+"/signatures.tar.gz")
	if err != nil {
		return fmt.Errorf("update: fetching archive: %w", err)
	}
	sig, err := fetch(ctx, client, opts, opts.SourceURL+"/signatures.tar.gz.sig")
	if err != nil {
		return fmt.Errorf("update: fetching signature: %w", err)
	}

	if err := Verify(archive, sig, opts.PublicKey); err != nil {
		return fmt.Errorf("update rejected: %w", err)
	}

	if err := Apply(archive, destDir); err != nil {
		return fmt.Errorf("update: applying archive: %w", err)
	}
	return nil
}

// Verify checks that sig is a valid Ed25519 signature over data made
// with the private key matching pub.
func Verify(data, sig []byte, pub ed25519.PublicKey) error {
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("public key is %d bytes, want %d", len(pub), ed25519.PublicKeySize)
	}
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("signature is %d bytes, want %d", len(sig), ed25519.SignatureSize)
	}
	if !ed25519.Verify(pub, data, sig) {
		return errors.New("signature verification failed")
	}
	return nil
}

func fetch(ctx context.Context, client *http.Client, opts Options, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", opts.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(io.LimitReader(resp.Body, opts.MaxBytes))
}
