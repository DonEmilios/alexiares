package main

import (
	"errors"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update signatures, rules, and schemas",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return errors.New("update: not yet implemented")
	},
}
