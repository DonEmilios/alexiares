// Package integration exercises Alexiares against fixtures that
// resemble real-world targets rather than synthetic unit-test inputs
// — in this file, the repository's own checked-in signatures/ and
// observations/ trees.
package integration

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alexiares/alexiares/internal/intel"
)

// repoRoot locates the repository root relative to this test file, so
// the test works regardless of the working directory it's run from.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve test file path")
	}
	// This file lives at <repo>/tests/integration/repository_test.go.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func TestRepositorySignaturesAreValid(t *testing.T) {
	sigs, err := intel.LoadSignatures(filepath.Join(repoRoot(t), "signatures"))
	if err != nil {
		t.Fatalf("LoadSignatures(repo signatures/) error = %v", err)
	}
	if len(sigs) == 0 {
		t.Error("LoadSignatures(repo signatures/) = 0 signatures, want at least the example signature")
	}
}

func TestRepositoryObservationsAreValid(t *testing.T) {
	obs, err := intel.LoadObservations(filepath.Join(repoRoot(t), "observations"))
	if err != nil {
		t.Fatalf("LoadObservations(repo observations/) error = %v", err)
	}
	if len(obs) == 0 {
		t.Error("LoadObservations(repo observations/) = 0 observations, want at least the example observation")
	}
}
