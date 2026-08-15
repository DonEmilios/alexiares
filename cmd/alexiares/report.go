package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report <result.json>",
	Short: "Generate a formatted report from saved scan results",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("report: not yet implemented for %q", args[0])
	},
}
