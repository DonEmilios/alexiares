package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/wallet"
)

var walletCmd = &cobra.Command{
	Use:   "wallet <address>",
	Short: "Analyze wallet intelligence for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		chain, ok := wallet.Classify(address)
		if !ok {
			return fmt.Errorf("wallet: %q does not match a supported chain's address format", address)
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\nchain: %s\n\nCorrelation against known infrastructure is not yet implemented.\n", address, chain)
		return err
	},
}
