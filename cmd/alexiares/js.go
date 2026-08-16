package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/collector"
	"github.com/alexiares/alexiares/internal/javascript"
)

var jsCmd = &cobra.Command{
	Use:   "js <url>",
	Short: "Extract JavaScript artifacts from a target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		target := collector.Classify(args[0])
		if target.Kind != collector.KindURL && target.Kind != collector.KindIP {
			return fmt.Errorf("js: %q is not a fetchable domain, URL, or IP", args[0])
		}

		c := collector.New(collector.Options{
			Timeout:              cfg.Network.Timeout(),
			UserAgent:            cfg.Network.UserAgent,
			AllowPrivateNetworks: cfg.Network.AllowPrivateNetworks,
		})
		raw, err := c.Collect(cmd.Context(), target.URL)
		if err != nil {
			return err
		}
		artifacts := javascript.Extract(raw.HTML, raw.ScriptURLs, raw.Scripts)

		ew := &errWriter{w: cmd.OutOrStdout()}
		ew.printf("JavaScript artifacts for %s\n\n", target.URL)
		ew.printf("Scripts analyzed: %d (%d inline, %d external)\n", len(artifacts.SHA256), len(artifacts.InlineScripts), len(artifacts.ScriptURLs))
		ew.printList("API endpoints", artifacts.APIEndpoints)
		ew.printList("Telegram references", artifacts.TelegramRefs)
		ew.printList("Discord webhooks", artifacts.DiscordWebhooks)
		ew.printList("Wallet libraries", artifacts.WalletLibraries)
		ew.printList("WalletConnect references", artifacts.WalletConnectRefs)
		ew.printList("SHA256 hashes", artifacts.SHA256)
		return ew.err
	},
}
