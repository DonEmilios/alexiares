package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update signatures, rules, and schemas",
	Long: "Fetches a signed signature repository archive and installs it, replacing the\n" +
		"local copy at signatures.path. Requires update.source_url and update.public_key\n" +
		"to be set in the config file — Alexiares never installs an update it cannot\n" +
		"cryptographically verify, and ships with no default update source.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		pub, err := cfg.Update.DecodedPublicKey()
		if err != nil {
			return fmt.Errorf("update: %w (set update.public_key in %s)", err, cfgPath)
		}

		err = update.Run(cmd.Context(), update.Options{
			SourceURL: cfg.Update.SourceURL,
			PublicKey: pub,
			Timeout:   cfg.Network.Timeout(),
			UserAgent: cfg.Network.UserAgent,
		}, cfg.Signatures.Path)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Signatures updated at %s\n", cfg.Signatures.Path)
		return err
	},
}
