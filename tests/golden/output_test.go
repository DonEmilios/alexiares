package golden

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/graph"
	"github.com/alexiares/alexiares/internal/output"
)

func sampleScanResult() output.ScanResult {
	b := graph.NewBuilder()
	b.AddNode("domain:phish.example", graph.NodeDomain, "phish.example")
	b.AddNode("favicon:abc123", graph.NodeFavicon, "abc123")
	b.AddEdge("domain:phish.example", "favicon:abc123", graph.EdgeSharesFavicon)

	return output.ScanResult{
		Target:     "https://phish.example",
		Confidence: "high",
		Evidence: []evidence.Item{
			{Category: "favicon", Strength: evidence.Strong, Value: "-204998123", SignatureID: "wallet_drainer_cluster_001", Description: "Shared favicon hash"},
			{Category: "telegram", Strength: evidence.Strong, Value: "api.telegram.org/bot", SignatureID: "wallet_drainer_cluster_001", Description: "Telegram exfiltration indicator"},
		},
		Recommendation: "Avoid wallet interaction. Do not connect, sign, or approve.",
		RelatedDomains: []string{"claim-rewards-example.xyz"},
		Graph:          b.Build(),
	}
}

// TestOutputRenderGolden locks the byte-for-byte shape of every
// output format against a fixed ScanResult. A diff here is either a
// real regression or an intentional format change (rerun with
// -update and review the diff before committing).
func TestOutputRenderGolden(t *testing.T) {
	result := sampleScanResult()

	formats := []output.Format{
		output.FormatTerminal,
		output.FormatJSON,
		output.FormatCSV,
		output.FormatMarkdown,
		output.FormatDOT,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			got, err := output.Render(format, result)
			if err != nil {
				t.Fatalf("Render(%s) error = %v", format, err)
			}

			path := filepath.Join("testdata", "output_"+string(format)+".golden")
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("creating testdata dir: %v", err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("writing golden fixture: %v", err)
				}
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading golden fixture %s (run with -update to create it): %v", path, err)
			}
			if got != string(want) {
				t.Errorf("Render(%s) does not match golden fixture (run with -update to regenerate):\ngot:\n%s\nwant:\n%s", format, got, want)
			}
		})
	}
}
