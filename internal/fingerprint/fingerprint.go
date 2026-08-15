// Package fingerprint normalizes collected artifacts into comparable,
// deterministic identifiers: favicon hashes, JavaScript hashes, an
// HTML structural hash plus a fuzzy DOM similarity fingerprint, and
// the TLS certificate fingerprint.
//
// Every function in this package is pure: it performs no I/O and,
// given the same input, always returns the same output. The
// correlation engine depends on that determinism to match targets
// against signatures byte-for-byte.
package fingerprint

import (
	"fmt"

	"github.com/alexiares/alexiares/internal/artifact"
)

// Compute derives Fingerprints from a RawResponse and the favicon and
// JavaScript artifacts already extracted from it. Fields for artifacts
// absent from the input (no favicon downloaded, no TLS handshake) are
// left at their zero value rather than omitted from the struct.
func Compute(raw artifact.RawResponse, fav artifact.FaviconArtifacts, js artifact.JavaScriptArtifacts) artifact.Fingerprints {
	tags := tagSequence(raw.HTML)

	fp := artifact.Fingerprints{
		Favicon:        fav.SHA256,
		FaviconHash:    fav.Murmur3,
		JavaScript:     js.SHA256,
		HTML:           structuralHash(raw.HTML),
		HTMLSimilarity: fmt.Sprintf("%016x", simHash(tags)),
	}
	if raw.TLS != nil {
		fp.Certificate = raw.TLS.SHA256
	}
	return fp
}
