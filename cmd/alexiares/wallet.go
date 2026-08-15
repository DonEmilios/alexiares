package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var walletCmd = &cobra.Command{
	Use:   "wallet <address>",
	Short: "Analyze wallet intelligence for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("wallet: not yet implemented for %q", args[0])
	},
}
