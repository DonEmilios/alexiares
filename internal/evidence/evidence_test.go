package evidence_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/evidence"
)

func TestEvaluateNoMatchesYieldsNoVerdict(t *testing.T) {
	got := evidence.Evaluate("clean.example", artifact.Correlation{})

	if got.Confidence != "" {
		t.Errorf("Confidence = %q, want empty for no matches", got.Confidence)
	}
	if len(got.Evidence) != 0 {
		t.Errorf("Evidence = %v, want empty", got.Evidence)
	}
	if got.Recommendation != "No known malicious infrastructure detected." {
		t.Errorf("Recommendation = %q, want the no-detection message", got.Recommendation)
	}
}

func TestEvaluateHighConfidenceRecommendsAvoidance(t *testing.T) {
	corr := artifact.Correlation{
		Matches: []artifact.Match{
			{SignatureID: "cluster001", Category: artifact.MatchFavicon, Value: "-204998123"},
		},
		Clusters: []artifact.Cluster{
			{SignatureID: "cluster001", Confidence: "high", Matches: []artifact.Match{
				{SignatureID: "cluster001", Category: artifact.MatchFavicon, Value: "-204998123"},
			}},
		},
		RelatedDomains: []string{"sibling.example"},
	}

	got := evidence.Evaluate("phish.example", corr)

	if got.Confidence != "high" {
		t.Errorf("Confidence = %q, want high", got.Confidence)
	}
	if got.Recommendation != "Avoid wallet interaction. Do not connect, sign, or approve." {
		t.Errorf("Recommendation = %q, want avoidance message", got.Recommendation)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Strength != evidence.Strong {
		t.Errorf("Evidence = %+v, want 1 strong item (favicon)", got.Evidence)
	}
	if got.Evidence[0].Description != "Shared favicon hash" {
		t.Errorf("Evidence[0].Description = %q, want %q", got.Evidence[0].Description, "Shared favicon hash")
	}
	if len(got.RelatedDomains) != 1 || got.RelatedDomains[0] != "sibling.example" {
		t.Errorf("RelatedDomains = %v, want [sibling.example]", got.RelatedDomains)
	}
}

func TestEvaluatePicksHighestConfidenceAcrossClusters(t *testing.T) {
	corr := artifact.Correlation{
		Clusters: []artifact.Cluster{
			{SignatureID: "a", Confidence: "medium"},
			{SignatureID: "b", Confidence: "critical"},
			{SignatureID: "c", Confidence: "low"},
		},
	}
	got := evidence.Evaluate("target.example", corr)
	if got.Confidence != "critical" {
		t.Errorf("Confidence = %q, want critical (the highest among clusters)", got.Confidence)
	}
}

func TestEvaluateSortsEvidenceStrongFirst(t *testing.T) {
	corr := artifact.Correlation{
		Matches: []artifact.Match{
			{SignatureID: "s1", Category: artifact.MatchIP, Value: "203.0.113.1"},
			{SignatureID: "s1", Category: artifact.MatchFavicon, Value: "hash1"},
			{SignatureID: "s1", Category: artifact.MatchNameserver, Value: "ns1.example"},
			{SignatureID: "s1", Category: artifact.MatchWallet, Value: "0xabc"},
		},
	}
	got := evidence.Evaluate("target.example", corr)

	if len(got.Evidence) != 4 {
		t.Fatalf("Evidence = %v, want 4 items", got.Evidence)
	}
	// Strong items (favicon, wallet) must sort before medium items (ip, nameserver).
	for i, item := range got.Evidence {
		if item.Strength == evidence.Medium {
			for _, prior := range got.Evidence[:i] {
				if prior.Strength == evidence.Weak {
					t.Errorf("weak evidence %+v sorted before medium %+v", prior, item)
				}
			}
		}
	}
	if got.Evidence[0].Strength != evidence.Strong || got.Evidence[1].Strength != evidence.Strong {
		t.Errorf("first two items = %+v, want both strong", got.Evidence[:2])
	}
}

func TestEvaluateIsDeterministic(t *testing.T) {
	corr := artifact.Correlation{
		Matches: []artifact.Match{
			{SignatureID: "s1", Category: artifact.MatchFavicon, Value: "hash1"},
			{SignatureID: "s1", Category: artifact.MatchWallet, Value: "0xabc"},
		},
		Clusters: []artifact.Cluster{{SignatureID: "s1", Confidence: "high"}},
	}

	first := evidence.Evaluate("target.example", corr)
	second := evidence.Evaluate("target.example", corr)

	if len(first.Evidence) != len(second.Evidence) {
		t.Fatal("Evaluate() evidence count not deterministic")
	}
	for i := range first.Evidence {
		if first.Evidence[i] != second.Evidence[i] {
			t.Errorf("Evaluate() not deterministic at index %d: %+v != %+v", i, first.Evidence[i], second.Evidence[i])
		}
	}
}
