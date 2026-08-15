package intel

import (
	"testing"
	"time"
)

func validObservation() Observation {
	return Observation{
		Type:       "observation",
		Domain:     "phish.example",
		ReportedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Reporter:   "analyst@example.com",
		Evidence:   ObservationEvidence{TxHash: "0xabc"},
	}
}

func TestObservationValidateAccepts(t *testing.T) {
	if err := validObservation().Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestObservationValidateRejects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Observation)
	}{
		{"missing domain", func(o *Observation) { o.Domain = "" }},
		{"missing reporter", func(o *Observation) { o.Reporter = "" }},
		{"missing reported_at", func(o *Observation) { o.ReportedAt = time.Time{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := validObservation()
			tt.mutate(&obs)
			if err := obs.Validate(); err == nil {
				t.Errorf("Validate() error = nil, want error for %s", tt.name)
			}
		})
	}
}
