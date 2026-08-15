package intel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

const sigA = `
id: wallet_drainer_cluster_001
description: Fake wallet connection infrastructure
favicon:
  murmur3: [-204998123]
confidence: high
`

const sigB = `
id: fake_governance_portal_002
description: Cloned DAO governance frontend
domains: [fake-gov.example]
confidence: medium
`

func TestLoadSignaturesValidRepository(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "infrastructure/cluster001.yaml", sigA)
	writeFile(t, dir, "domains/portal002.yml", sigB)

	got, err := LoadSignatures(dir)
	if err != nil {
		t.Fatalf("LoadSignatures() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("LoadSignatures() = %d signatures, want 2", len(got))
	}
}

func TestLoadSignaturesDetectsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.yaml", sigA)
	writeFile(t, dir, "b.yaml", sigA) // same id

	_, err := LoadSignatures(dir)
	if err == nil {
		t.Fatal("LoadSignatures() error = nil, want duplicate id error")
	}
	if !strings.Contains(err.Error(), "duplicate signature id") {
		t.Errorf("error = %v, want mention of duplicate signature id", err)
	}
}

func TestLoadSignaturesRejectsInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.yaml", "id: \"\"\ndescription: missing everything\n")

	_, err := LoadSignatures(dir)
	if err == nil {
		t.Fatal("LoadSignatures() error = nil, want validation error")
	}
}

func TestLoadSignaturesRejectsMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "broken.yaml", "id: [this is not: valid yaml")

	_, err := LoadSignatures(dir)
	if err == nil {
		t.Fatal("LoadSignatures() error = nil, want parse error")
	}
}

func TestLoadSignaturesMissingDirectoryIsEmpty(t *testing.T) {
	got, err := LoadSignatures(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("LoadSignatures() error = %v, want nil for missing dir", err)
	}
	if len(got) != 0 {
		t.Errorf("LoadSignatures() = %v, want empty", got)
	}
}

func TestLoadSignaturesIgnoresNonYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sig.yaml", sigA)
	writeFile(t, dir, "README.md", "# not a signature")

	got, err := LoadSignatures(dir)
	if err != nil {
		t.Fatalf("LoadSignatures() error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("LoadSignatures() = %d, want 1 (README.md ignored)", len(got))
	}
}

const obsA = `
type: observation
domain: phish.example
reported_at: 2026-01-01T00:00:00Z
reporter: analyst@example.com
evidence:
  tx_hash: "0xabc"
`

func TestLoadObservationsValidRepository(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "report1.yaml", obsA)

	got, err := LoadObservations(dir)
	if err != nil {
		t.Fatalf("LoadObservations() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("LoadObservations() = %d, want 1", len(got))
	}
	if got[0].Domain != "phish.example" {
		t.Errorf("Domain = %q, want phish.example", got[0].Domain)
	}
}

func TestLoadObservationsRejectsInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.yaml", "domain: \"\"\n")

	_, err := LoadObservations(dir)
	if err == nil {
		t.Fatal("LoadObservations() error = nil, want validation error")
	}
}
