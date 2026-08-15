// Package intel defines Alexiares' two-layer intelligence model.
//
// Signatures are maintainer-reviewed, trusted detection artifacts used
// for deterministic matching. Observations are raw, untrusted
// community reports kept for historical context and analyst review.
//
// The separation is structural, not a policy note: this package
// contains no function that converts an Observation into a Signature.
// Promoting an observation to a signature is a human, maintainer-review
// decision — writing a new signature file by hand — never an automated
// one. Anyone extending this package must preserve that: do not add a
// Observation-to-Signature conversion here.
package intel

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

// Confidence is a qualitative detection confidence level.
type Confidence string

// Supported confidence levels, per the spec's qualitative model.
const (
	ConfidenceLow      Confidence = "low"
	ConfidenceMedium   Confidence = "medium"
	ConfidenceHigh     Confidence = "high"
	ConfidenceCritical Confidence = "critical"
)

func (c Confidence) valid() bool {
	switch c {
	case ConfidenceLow, ConfidenceMedium, ConfidenceHigh, ConfidenceCritical:
		return true
	default:
		return false
	}
}

// FaviconSignature matches on favicon fingerprints.
type FaviconSignature struct {
	Murmur3 []int32  `yaml:"murmur3,omitempty"`
	SHA256  []string `yaml:"sha256,omitempty"`
}

// JavaScriptSignature matches on script content hashes.
type JavaScriptSignature struct {
	SHA256 []string `yaml:"sha256,omitempty"`
}

// TelegramSignature matches on Telegram exfiltration patterns
// (literal substrings or references, not full regular expressions —
// keeping signatures auditable by non-programmers).
type TelegramSignature struct {
	Patterns []string `yaml:"patterns,omitempty"`
}

// Signature is a maintainer-reviewed, trusted detection artifact.
type Signature struct {
	ID          string              `yaml:"id"`
	Description string              `yaml:"description"`
	Favicon     FaviconSignature    `yaml:"favicon,omitempty"`
	JavaScript  JavaScriptSignature `yaml:"javascript,omitempty"`
	// Certificates holds SHA256 fingerprints of malicious TLS certs.
	Certificates []string            `yaml:"certificates,omitempty"`
	Wallets      map[string][]string `yaml:"wallets,omitempty"` // chain -> addresses
	Telegram     TelegramSignature   `yaml:"telegram,omitempty"`
	Domains      []string            `yaml:"domains,omitempty"`
	Confidence   Confidence          `yaml:"confidence"`
}

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$`)

// Validate checks a Signature against Alexiares' schema: required
// fields, a valid confidence level, well-formed hashes, and at least
// one detection criterion — a signature that can never match anything
// is a maintainer error, not a valid entry.
func (s Signature) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("signature: id is required")
	}
	if !idPattern.MatchString(s.ID) {
		return fmt.Errorf("signature %q: id must be lowercase alphanumeric with - or _ separators", s.ID)
	}
	if s.Description == "" {
		return fmt.Errorf("signature %q: description is required", s.ID)
	}
	if !s.Confidence.valid() {
		return fmt.Errorf("signature %q: confidence %q must be one of low, medium, high, critical", s.ID, s.Confidence)
	}

	for _, h := range s.Favicon.SHA256 {
		if err := validateSHA256(h); err != nil {
			return fmt.Errorf("signature %q: favicon.sha256: %w", s.ID, err)
		}
	}
	for _, h := range s.JavaScript.SHA256 {
		if err := validateSHA256(h); err != nil {
			return fmt.Errorf("signature %q: javascript.sha256: %w", s.ID, err)
		}
	}
	for _, h := range s.Certificates {
		if err := validateSHA256(h); err != nil {
			return fmt.Errorf("signature %q: certificates: %w", s.ID, err)
		}
	}

	if !s.hasCriteria() {
		return fmt.Errorf("signature %q: must define at least one detection criterion (favicon, javascript, certificates, wallets, telegram, or domains)", s.ID)
	}

	return nil
}

// hasCriteria reports whether the signature carries any matchable
// indicator at all.
func (s Signature) hasCriteria() bool {
	if len(s.Favicon.Murmur3) > 0 || len(s.Favicon.SHA256) > 0 {
		return true
	}
	if len(s.JavaScript.SHA256) > 0 {
		return true
	}
	if len(s.Certificates) > 0 {
		return true
	}
	if len(s.Telegram.Patterns) > 0 {
		return true
	}
	if len(s.Domains) > 0 {
		return true
	}
	for _, addrs := range s.Wallets {
		if len(addrs) > 0 {
			return true
		}
	}
	return false
}

func validateSHA256(h string) error {
	if len(h) != 64 {
		return fmt.Errorf("%q is not a 64-character SHA256 hex digest", h)
	}
	if _, err := hex.DecodeString(h); err != nil {
		return fmt.Errorf("%q is not valid hex: %w", h, err)
	}
	return nil
}
