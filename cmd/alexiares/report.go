package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/output"
)

var reportCmd = &cobra.Command{
	Use:   "report <result.json>",
	Short: "Generate a formatted report from saved scan results",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("report: reading %s: %w", args[0], err)
		}

		var result output.ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("report: %s is not a valid scan result: %w", args[0], err)
		}

		format := output.Format(outputFormat(cmd, cfg))
		rendered, err := output.Render(format, result)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), rendered)
		return err
	},
}
