// Package config loads and validates Alexiares' user configuration.
//
// Configuration is read from ~/.alexiares/config.yaml by default. Missing
// files are not an error: Load returns the built-in defaults so the CLI
// works with zero setup.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Network holds settings for outbound collection requests.
type Network struct {
	// TimeoutSeconds bounds every network operation performed by the
	// collector. Zero falls back to Default's timeout.
	TimeoutSeconds int    `yaml:"timeout"`
	UserAgent      string `yaml:"user_agent"`
}

// Timeout returns the configured network timeout as a time.Duration.
func (n Network) Timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

// Signatures holds settings for the local signature repository.
type Signatures struct {
	Path string `yaml:"path"`
}

// Output holds default rendering settings.
type Output struct {
	Format string `yaml:"format"`
}

// Config is the fully resolved Alexiares configuration.
type Config struct {
	Network    Network    `yaml:"network"`
	Signatures Signatures `yaml:"signatures"`
	Output     Output     `yaml:"output"`
}

// Default returns Alexiares' built-in configuration, used whenever no
// config file is present or a field is left unset.
func Default() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return Config{
		Network: Network{
			TimeoutSeconds: 10,
			UserAgent:      "Alexiares/0.1",
		},
		Signatures: Signatures{
			Path: filepath.Join(home, ".alexiares", "signatures"),
		},
		Output: Output{
			Format: "terminal",
		},
	}
}

// DefaultPath returns the default config file location, ~/.alexiares/config.yaml.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".alexiares", "config.yaml")
}

// Load reads the config file at path and merges it over Default. A
// missing file is not an error: Load returns Default() unchanged.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config %s: %w", path, err)
	}

	// Unmarshal onto the defaults so unset fields in the file keep
	// their default value rather than becoming a zero value.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return cfg, nil
}
