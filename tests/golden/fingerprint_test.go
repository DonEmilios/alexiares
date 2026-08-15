// Package golden holds output-stability tests: fixed inputs compared
// byte-for-byte against a checked-in golden file. A diff here means
// either a real regression or an intentional format change — in the
// latter case, rerun with -update to regenerate the fixture and
// review the diff before committing it.
package golden

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/favicon"
	"github.com/alexiares/alexiares/internal/fingerprint"
	"github.com/alexiares/alexiares/internal/javascript"
)

var update = flag.Bool("update", false, "regenerate golden fixtures instead of comparing against them")

const samplePage = `<!DOCTYPE html>
<html>
<head>
	<title>Claim your airdrop</title>
	<meta name="description" content="Connect your wallet to claim">
	<script src="/drainer.js"></script>
</head>
<body>
	<header><nav><a href="/">Home</a></nav></header>
	<form action="/connect" method="post">
		<input type="text" name="wallet">
		<input type="hidden" name="csrf" value="abc123">
	</form>
	<script>
		fetch("https://api.telegram.org/bot123456789:AAExampleTokenValue1234567890123456/sendMessage");
	</script>
	<footer>phishkit v2</footer>
</body>
</html>`

var sampleFaviconBytes = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}

func TestFingerprintComputeGolden(t *testing.T) {
	raw := artifact.RawResponse{
		URL:      "https://phish.example/",
		FinalURL: "https://phish.example/",
		HTML:     samplePage,
		TLS: &artifact.TLSData{
			SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	fav := favicon.Compute(sampleFaviconBytes)
	js := javascript.Extract(raw.HTML, []string{"/drainer.js"}, []string{"console.log('external drainer payload')"})

	got := fingerprint.Compute(raw, fav, js)

	goldenPath := filepath.Join("testdata", "fingerprint_golden.json")

	if *update {
		writeGolden(t, goldenPath, got)
	}

	want := readGolden(t, goldenPath)
	compareJSON(t, got, want)
}

func writeGolden(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshaling golden fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating testdata dir: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("writing golden fixture: %v", err)
	}
}

func readGolden(t *testing.T, path string) artifact.Fingerprints {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden fixture %s (run with -update to create it): %v", path, err)
	}
	var v artifact.Fingerprints
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("parsing golden fixture: %v", err)
	}
	return v
}

func compareJSON(t *testing.T, got, want artifact.Fingerprints) {
	t.Helper()
	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("fingerprint output does not match golden fixture (run with -update to regenerate):\ngot:\n%s\nwant:\n%s", gotJSON, wantJSON)
	}
}
