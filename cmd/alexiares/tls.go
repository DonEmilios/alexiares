package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tlsCmd = &cobra.Command{
	Use:   "tls <domain>",
	Short: "Display certificate intelligence for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return fmt.Errorf("tls: not yet implemented for %q", args[0])
	},
}
