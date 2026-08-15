package intel

import (
	"fmt"
	"time"
)

// ObservationEvidence is the supporting material behind an
// observation. All fields are optional free text — analysts attach
// whatever they have, and a maintainer decides later what it's worth.
type ObservationEvidence struct {
	Screenshot string `yaml:"screenshot,omitempty"`
	TxHash     string `yaml:"tx_hash,omitempty"`
	Wallet     string `yaml:"wallet,omitempty"`
}

// Observation is a raw, untrusted community report. It is never
// promoted to a Signature automatically — see the package doc.
type Observation struct {
	Type       string              `yaml:"type"`
	Domain     string              `yaml:"domain"`
	ReportedAt time.Time           `yaml:"reported_at"`
	Reporter   string              `yaml:"reporter"`
	Evidence   ObservationEvidence `yaml:"evidence"`
}

// Validate checks that an Observation has the minimum information
// needed to be useful: what was reported, by whom, and when.
func (o Observation) Validate() error {
	if o.Domain == "" {
		return fmt.Errorf("observation: domain is required")
	}
	if o.Reporter == "" {
		return fmt.Errorf("observation %q: reporter is required", o.Domain)
	}
	if o.ReportedAt.IsZero() {
		return fmt.Errorf("observation %q: reported_at is required", o.Domain)
	}
	return nil
}
