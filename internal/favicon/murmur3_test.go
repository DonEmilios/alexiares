package favicon

import "testing"

// Reference values generated independently via Python's mmh3 package
// (mmh3.hash(s.encode(), 0)), the reference MurmurHash3 x86_32
// implementation used across the security tooling ecosystem
// (including Shodan's favicon hashing, which Alexiares' favicon
// signatures are designed to interoperate with).
func TestMurmur3_32ReferenceVectors(t *testing.T) {
	tests := []struct {
		input string
		want  int32
	}{
		{"", 0},
		{"test", -1167338989},
		{"a", 1009084850},
		{"abc", -1277324294},
		{"Hello, world!", -1070186941},
		{"The quick brown fox jumps over the lazy dog", 776992547},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := murmur3_32([]byte(tt.input), 0); got != tt.want {
				t.Errorf("murmur3_32(%q, 0) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
