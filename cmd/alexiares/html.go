package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html <url>",
	Short: "Extract HTML artifacts from a target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("html: not yet implemented for %q", args[0])
	},
}
