// Package favicon computes deterministic identifiers for a favicon's
// raw bytes: a MurmurHash3 hash compatible with the convention used
// across the security tooling ecosystem (Shodan's http.favicon.hash
// and the malicious-favicon-hash datasets built on it), and a SHA256
// hash of the raw bytes for exact-byte matching.
package favicon

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/alexiares/alexiares/internal/artifact"
)

// lineWidth is the line length Python's base64.encodestring (and the
// email MIME encoders it wraps) inserts newlines at. Shodan's favicon
// hash convention base64-encodes the icon with this wrapping before
// hashing, so matching it exactly is required for signature
// interoperability with existing community favicon-hash datasets.
const lineWidth = 76

// Compute derives FaviconArtifacts from raw favicon bytes. An empty
// input returns the zero value.
func Compute(data []byte) artifact.FaviconArtifacts {
	if len(data) == 0 {
		return artifact.FaviconArtifacts{}
	}
	return artifact.FaviconArtifacts{
		Murmur3: murmur3_32([]byte(wrappedBase64(data)), 0),
		SHA256:  fmt.Sprintf("%x", sha256.Sum256(data)),
		Size:    len(data),
	}
}

// wrappedBase64 standard-base64-encodes data and inserts a newline
// every lineWidth characters, including a trailing newline — the
// exact framing MurmurHash3 is computed over in the Shodan-style
// favicon hash convention.
func wrappedBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)

	var b strings.Builder
	b.Grow(len(encoded) + len(encoded)/lineWidth + 1)
	for i := 0; i < len(encoded); i += lineWidth {
		end := i + lineWidth
		if end > len(encoded) {
			end = len(encoded)
		}
		b.WriteString(encoded[i:end])
		b.WriteByte('\n')
	}
	return b.String()
}
