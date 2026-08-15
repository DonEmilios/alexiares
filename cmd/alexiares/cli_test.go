package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexiares/alexiares/internal/config"
)

// execute runs rootCmd with args and returns captured stdout/stderr
// plus whatever error RunE returned.
//
// rootCmd is a package-level singleton reused by every test in this
// package, so its persistent flags (--config, --format) must be reset
// before each call: pflag does not restore a flag to its default
// between Parse() calls, it only overwrites flags that appear in the
// new args. Without this reset, a --format=json passed in one test
// would silently leak into the next test that doesn't mention
// --format at all.
func execute(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cfgPath = config.DefaultPath()
	if setErr := rootCmd.PersistentFlags().Set("format", ""); setErr != nil {
		t.Fatalf("resetting --format flag: %v", setErr)
	}

	var outBuf, errBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs(args)

	err = rootCmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// writeConfig writes a minimal config file pointing signatures.path at
// dir and returns the config file's path, for tests that need
// signatures loaded from a controlled location rather than the real
// ~/.alexiares/signatures.
func writeConfig(t *testing.T, signaturesDir string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "network:\n  timeout: 5\nsignatures:\n  path: " + signaturesDir + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}
