package main

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <target>",
	Short: "Perform a complete infrastructure analysis of a target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(cmd, args[0])
	},
}
