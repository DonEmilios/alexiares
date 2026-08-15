package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph <domain>",
	Short: "Build the infrastructure relationship graph for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("graph: not yet implemented for %q", args[0])
	},
}
