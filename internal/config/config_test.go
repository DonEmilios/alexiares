package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexiares/alexiares/internal/config"
)

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	got, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := config.Default()
	if got != want {
		t.Errorf("Load() = %+v, want default %+v", got, want)
	}
}

func TestLoadMergesOverDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "network:\n  timeout: 30\noutput:\n  format: json\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if got.Network.TimeoutSeconds != 30 {
		t.Errorf("Network.TimeoutSeconds = %d, want 30", got.Network.TimeoutSeconds)
	}
	if got.Output.Format != "json" {
		t.Errorf("Output.Format = %q, want json", got.Output.Format)
	}
	// Unset fields must keep their default value.
	if got.Network.UserAgent != config.Default().Network.UserAgent {
		t.Errorf("Network.UserAgent = %q, want default preserved", got.Network.UserAgent)
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("network: [this is not valid: yaml"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := config.Load(path); err == nil {
		t.Error("Load() error = nil, want error for malformed YAML")
	}
}

func TestNetworkTimeout(t *testing.T) {
	n := config.Network{TimeoutSeconds: 5}
	if got, want := n.Timeout().Seconds(), 5.0; got != want {
		t.Errorf("Timeout() = %.0fs, want %.0fs", got, want)
	}
}
