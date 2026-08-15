package favicon_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/favicon"
)

func testBytes256() []byte {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

// Reference value generated independently via Python:
//
//	base64.encodebytes(bytes(range(256)))  # wraps at 76 chars + trailing newlines
//	mmh3.hash(encoded, 0)
//
// This is the Shodan http.favicon.hash convention that Alexiares'
// favicon signatures are designed to interoperate with.
func TestComputeMatchesShodanConvention(t *testing.T) {
	got := favicon.Compute(testBytes256())

	if got.Murmur3 != -757223386 {
		t.Errorf("Murmur3 = %d, want -757223386", got.Murmur3)
	}
	if got.SHA256 != "40aff2e9d2d8922e47afd4648e6967497158785fbd1da870e7110266bf944880" {
		t.Errorf("SHA256 = %q, want matching digest", got.SHA256)
	}
	if got.Size != 256 {
		t.Errorf("Size = %d, want 256", got.Size)
	}
}

func TestComputeEmptyInput(t *testing.T) {
	got := favicon.Compute(nil)
	var zero struct {
		Murmur3 int32
		SHA256  string
		Size    int
	}
	if got.Murmur3 != zero.Murmur3 || got.SHA256 != zero.SHA256 || got.Size != zero.Size {
		t.Errorf("Compute(nil) = %+v, want zero value", got)
	}
}

func TestComputeIsDeterministic(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	first := favicon.Compute(data)
	second := favicon.Compute(data)
	if first != second {
		t.Errorf("Compute() not deterministic: %+v != %+v", first, second)
	}
}
