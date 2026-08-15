package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runScan is implemented incrementally as the collector, extractor,
// fingerprint, correlation, and evidence engines land. Until then it
// reports clearly rather than faking a result.
func runScan(cmd *cobra.Command, target string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	format := outputFormat(cmd, cfg)
	return fmt.Errorf("scan: not yet implemented for %q (format=%s)", target, format)
}
