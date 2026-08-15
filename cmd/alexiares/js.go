package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var jsCmd = &cobra.Command{
	Use:   "js <url>",
	Short: "Extract JavaScript artifacts from a target",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return fmt.Errorf("js: not yet implemented for %q", args[0])
	},
}
