package collector_test

import (
	"strings"
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/collector"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    collector.Kind
		wantURL string
		chain   artifact.WalletChain
	}{
		{"full https url", "https://example.xyz/path", collector.KindURL, "https://example.xyz/path", ""},
		{"full http url", "http://example.xyz", collector.KindURL, "http://example.xyz", ""},
		{"bare domain upgraded to https", "example.xyz", collector.KindURL, "https://example.xyz", ""},
		{"ipv4", "203.0.113.42", collector.KindIP, "https://203.0.113.42", ""},
		{"ipv6", "2001:db8::1", collector.KindIP, "https://2001:db8::1", ""},
		{"ethereum wallet", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590", collector.KindWallet, "", artifact.ChainEthereum},
		{"unresolvable", "not a target", collector.KindUnknown, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collector.Classify(tt.raw)
			if got.Kind != tt.want {
				t.Errorf("Classify(%q).Kind = %v, want %v", tt.raw, got.Kind, tt.want)
			}
			if tt.wantURL != "" && got.URL != tt.wantURL {
				t.Errorf("Classify(%q).URL = %q, want %q", tt.raw, got.URL, tt.wantURL)
			}
			if tt.chain != "" && got.Chain != tt.chain {
				t.Errorf("Classify(%q).Chain = %v, want %v", tt.raw, got.Chain, tt.chain)
			}
		})
	}
}

func TestReadTargetsSkipsBlankAndCommentLines(t *testing.T) {
	input := "example.xyz\n\n# a comment\n  another.xyz  \n0x742d35Cc6634C0532925a3b844Bc9e7595f89590\n"
	got, err := collector.ReadTargets(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadTargets() error = %v", err)
	}
	want := []string{"example.xyz", "another.xyz", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590"}
	if len(got) != len(want) {
		t.Fatalf("ReadTargets() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadTargets()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
