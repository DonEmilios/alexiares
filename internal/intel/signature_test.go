package intel

import "testing"

func validSignature() Signature {
	return Signature{
		ID:          "wallet_drainer_cluster_001",
		Description: "Fake wallet connection infrastructure",
		Favicon:     FaviconSignature{Murmur3: []int32{-204998123}},
		Confidence:  ConfidenceHigh,
	}
}

func TestSignatureValidateAccepts(t *testing.T) {
	if err := validSignature().Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestSignatureValidateRejects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Signature)
	}{
		{"missing id", func(s *Signature) { s.ID = "" }},
		{"id with uppercase", func(s *Signature) { s.ID = "Bad_ID" }},
		{"id with spaces", func(s *Signature) { s.ID = "bad id" }},
		{"missing description", func(s *Signature) { s.Description = "" }},
		{"invalid confidence", func(s *Signature) { s.Confidence = "extreme" }},
		{"bad sha256 length", func(s *Signature) { s.JavaScript.SHA256 = []string{"deadbeef"} }},
		{"bad sha256 hex", func(s *Signature) {
			s.JavaScript.SHA256 = []string{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"}
		}},
		{"no detection criteria", func(s *Signature) { s.Favicon = FaviconSignature{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := validSignature()
			tt.mutate(&sig)
			if err := sig.Validate(); err == nil {
				t.Errorf("Validate() error = nil, want error for %s", tt.name)
			}
		})
	}
}

func TestSignatureValidateAcceptsWalletOnlyCriteria(t *testing.T) {
	sig := Signature{
		ID:          "wallet-only-signature",
		Description: "Known drainer receiving wallet",
		Wallets:     map[string][]string{"ethereum": {"0x123"}},
		Confidence:  ConfidenceMedium,
	}
	if err := sig.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil for wallet-only signature", err)
	}
}

func TestSignatureValidateRejectsEmptyWalletCriteria(t *testing.T) {
	sig := validSignature()
	sig.Favicon = FaviconSignature{}
	sig.Wallets = map[string][]string{"ethereum": {}} // key present, no addresses
	if err := sig.Validate(); err == nil {
		t.Error("Validate() error = nil, want error for empty wallet list")
	}
}
