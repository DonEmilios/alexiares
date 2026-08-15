package fingerprint

import "testing"

func TestStructuralHashIgnoresTextAndAttributes(t *testing.T) {
	a := structuralHash(`<html><body><div class="a" id="x">Hello</div></body></html>`)
	b := structuralHash(`<html><body><div class="different" id="y">Goodbye, world</div></body></html>`)

	if a == "" {
		t.Fatal("structuralHash() = empty, want a hash")
	}
	if a != b {
		t.Errorf("structuralHash differs for same shape, different text/attrs:\n  a: %s\n  b: %s", a, b)
	}
}

func TestStructuralHashDiffersForDifferentShape(t *testing.T) {
	a := structuralHash(`<html><body><div><span></span></div></body></html>`)
	b := structuralHash(`<html><body><div></div><span></span></body></html>`)

	if a == b {
		t.Error("structuralHash() equal for different DOM shapes, want different hashes")
	}
}

func TestStructuralHashDeterministic(t *testing.T) {
	page := `<html><head><title>x</title></head><body><form><input></form></body></html>`
	first := structuralHash(page)
	second := structuralHash(page)
	if first != second {
		t.Errorf("structuralHash() not deterministic: %q != %q", first, second)
	}
}

func TestTagSequenceOrder(t *testing.T) {
	got := tagSequence(`<html><head><title>x</title></head><body><div></div></body></html>`)
	want := []string{"html", "head", "title", "body", "div"}
	if len(got) != len(want) {
		t.Fatalf("tagSequence() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tagSequence()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
