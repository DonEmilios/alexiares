package correlation_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/correlation"
	"github.com/alexiares/alexiares/internal/intel"
)

func drainerSignature() intel.Signature {
	return intel.Signature{
		ID:           "wallet_drainer_cluster_001",
		Description:  "Fake wallet connection infrastructure",
		Favicon:      intel.FaviconSignature{Murmur3: []int32{-204998123}, SHA256: []string{"fav-sha"}},
		JavaScript:   intel.JavaScriptSignature{SHA256: []string{"drainer-js-sha"}},
		Certificates: []string{"cert-sha"},
		Wallets:      map[string][]string{"ethereum": {"0xDRAINER"}},
		Telegram:     intel.TelegramSignature{Patterns: []string{"api.telegram.org/bot"}},
		Domains:      []string{"claim-rewards-example.xyz", "airdrop-example.app"},
		IPs:          []string{"203.0.113.42"},
		Nameservers:  []string{"ns1.evil.example"},
		Confidence:   intel.ConfidenceHigh,
	}
}

func TestCorrelateMatchesEveryCategory(t *testing.T) {
	target := correlation.Target{
		Domain: "phish.example",
		Fingerprints: artifact.Fingerprints{
			Favicon:     "fav-sha",
			FaviconHash: -204998123,
			JavaScript:  []string{"drainer-js-sha", "unrelated-hash"},
			Certificate: "cert-sha",
		},
		Wallets: artifact.WalletArtifacts{
			Addresses: []artifact.WalletAddress{{Chain: artifact.ChainEthereum, Address: "0xDRAINER"}},
		},
		Telegram: artifact.TelegramArtifacts{
			APIRefs: []string{"api.telegram.org/bot123456:token"},
		},
		IPs:             []string{"203.0.113.42"},
		Nameservers:     []string{"ns1.evil.example"},
		RedirectDomains: []string{"airdrop-example.app"},
	}

	got := correlation.Correlate(target, []intel.Signature{drainerSignature()})

	if len(got.Clusters) != 1 {
		t.Fatalf("Clusters = %v, want 1 cluster", got.Clusters)
	}
	cluster := got.Clusters[0]
	if cluster.SignatureID != "wallet_drainer_cluster_001" || cluster.Confidence != "high" {
		t.Errorf("cluster = %+v, want id/confidence set", cluster)
	}

	wantCategories := map[artifact.MatchCategory]bool{
		artifact.MatchFavicon:     true,
		artifact.MatchJavaScript:  true,
		artifact.MatchCertificate: true,
		artifact.MatchWallet:      true,
		artifact.MatchTelegram:    true,
		artifact.MatchRedirect:    true,
		artifact.MatchIP:          true,
		artifact.MatchNameserver:  true,
	}
	gotCategories := make(map[artifact.MatchCategory]bool)
	for _, m := range got.Matches {
		gotCategories[m.Category] = true
	}
	for cat := range wantCategories {
		if !gotCategories[cat] {
			t.Errorf("missing expected match category %q; got matches: %+v", cat, got.Matches)
		}
	}

	// Favicon matches on both SHA256 and Murmur3, so 2 favicon matches expected.
	faviconCount := 0
	for _, m := range got.Matches {
		if m.Category == artifact.MatchFavicon {
			faviconCount++
		}
	}
	if faviconCount != 2 {
		t.Errorf("favicon matches = %d, want 2 (sha256 + murmur3)", faviconCount)
	}
}

func TestCorrelateRelatedDomainsExcludesSelf(t *testing.T) {
	sig := drainerSignature()
	sig.Domains = []string{"phish.example", "airdrop-example.app"}

	target := correlation.Target{
		Domain:       "phish.example",
		Fingerprints: artifact.Fingerprints{Favicon: "fav-sha"},
	}

	got := correlation.Correlate(target, []intel.Signature{sig})

	if len(got.RelatedDomains) != 1 || got.RelatedDomains[0] != "airdrop-example.app" {
		t.Errorf("RelatedDomains = %v, want [airdrop-example.app] (self excluded)", got.RelatedDomains)
	}
}

func TestCorrelateRelatedWallets(t *testing.T) {
	target := correlation.Target{
		Domain:       "phish.example",
		Fingerprints: artifact.Fingerprints{Favicon: "fav-sha"},
	}
	got := correlation.Correlate(target, []intel.Signature{drainerSignature()})

	if len(got.RelatedWallets) != 1 || got.RelatedWallets[0] != "0xDRAINER" {
		t.Errorf("RelatedWallets = %v, want [0xDRAINER]", got.RelatedWallets)
	}
}

func TestCorrelateNoMatchesForCleanTarget(t *testing.T) {
	target := correlation.Target{
		Domain: "legit-site.example",
		Fingerprints: artifact.Fingerprints{
			Favicon: "totally-different-hash",
		},
	}
	got := correlation.Correlate(target, []intel.Signature{drainerSignature()})

	if len(got.Matches) != 0 || len(got.Clusters) != 0 {
		t.Errorf("Correlate(clean target) = %+v, want no matches or clusters", got)
	}
}

func TestCorrelateIsDeterministic(t *testing.T) {
	target := correlation.Target{
		Domain: "phish.example",
		Fingerprints: artifact.Fingerprints{
			Favicon: "fav-sha", FaviconHash: -204998123,
		},
	}
	sigs := []intel.Signature{drainerSignature()}

	first := correlation.Correlate(target, sigs)
	second := correlation.Correlate(target, sigs)

	if len(first.Matches) != len(second.Matches) || len(first.Clusters) != len(second.Clusters) {
		t.Errorf("Correlate() not deterministic: first=%+v second=%+v", first, second)
	}
}

func TestCorrelateEmptySignatureSet(t *testing.T) {
	got := correlation.Correlate(correlation.Target{Domain: "anything.example"}, nil)
	if len(got.Matches) != 0 || len(got.Clusters) != 0 || len(got.RelatedDomains) != 0 || len(got.RelatedWallets) != 0 {
		t.Errorf("Correlate(no signatures) = %+v, want all empty", got)
	}
}
