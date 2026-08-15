package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns <domain>",
	Short: "Display DNS intelligence for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return fmt.Errorf("dns: not yet implemented for %q", args[0])
	},
}
