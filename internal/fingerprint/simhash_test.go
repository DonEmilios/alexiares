package fingerprint

import "testing"

func TestSimHashIdenticalSequencesMatch(t *testing.T) {
	tags := []string{"html", "head", "title", "body", "div", "form", "input"}
	first := simHash(tags)
	second := simHash(tags)
	if first != second {
		t.Errorf("simHash() not deterministic: %d != %d", first, second)
	}
}

func TestSimHashNearDuplicatesHaveLowHammingDistance(t *testing.T) {
	// A phishing clone: same skeleton, one extra decorative <span>
	// inserted deep in the body — most shingles are unaffected.
	original := []string{"html", "head", "title", "meta", "body", "header", "nav", "div", "form", "input", "input", "button", "footer"}
	clone := []string{"html", "head", "title", "meta", "body", "header", "nav", "div", "span", "form", "input", "input", "button", "footer"}

	dist := HammingDistance64(simHash(original), simHash(clone))
	if dist > 20 {
		t.Errorf("HammingDistance64(near-duplicate) = %d, want a small distance (<=20 of 64 bits)", dist)
	}
}

func TestSimHashUnrelatedSequencesDivergeMoreThanNearDuplicates(t *testing.T) {
	original := []string{"html", "head", "title", "meta", "body", "header", "nav", "div", "form", "input", "input", "button", "footer"}
	clone := []string{"html", "head", "title", "meta", "body", "header", "nav", "div", "span", "form", "input", "input", "button", "footer"}
	unrelated := []string{"table", "tr", "td", "table", "tr", "td", "img", "img", "img", "video", "audio", "canvas", "svg"}

	nearDist := HammingDistance64(simHash(original), simHash(clone))
	farDist := HammingDistance64(simHash(original), simHash(unrelated))

	if farDist <= nearDist {
		t.Errorf("expected unrelated sequences to diverge more: near=%d far=%d", nearDist, farDist)
	}
}

func TestSimHashShortSequenceIsZero(t *testing.T) {
	if got := simHash([]string{"html", "body"}); got != 0 {
		t.Errorf("simHash(too short) = %d, want 0", got)
	}
	if got := simHash(nil); got != 0 {
		t.Errorf("simHash(nil) = %d, want 0", got)
	}
}

func TestHammingDistance64(t *testing.T) {
	tests := []struct {
		name string
		a, b uint64
		want int
	}{
		{"identical", 0xFF, 0xFF, 0},
		{"all bits differ", 0x0, 0xFFFFFFFFFFFFFFFF, 64},
		{"one bit differs", 0b1010, 0b1011, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HammingDistance64(tt.a, tt.b); got != tt.want {
				t.Errorf("HammingDistance64(%b, %b) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
