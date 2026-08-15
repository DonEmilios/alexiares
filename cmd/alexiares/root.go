package main

import (
	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/config"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "alexiares",
	Short: "Infrastructure intelligence for Web3 operational security",
	Long: "Alexiares detects phishing infrastructure, wallet drainers, and malicious\n" +
		"frontends by fingerprinting and correlating collected artifacts against\n" +
		"community-maintained signatures. Guard the gate before the wallet connects.",
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultPath(), "path to config file")
	rootCmd.PersistentFlags().StringP("format", "f", "", "output format (terminal, json, graphml, dot, csv, markdown)")

	rootCmd.AddCommand(
		versionCmd,
		scanCmd,
		graphCmd,
		dnsCmd,
		tlsCmd,
		htmlCmd,
		jsCmd,
		walletCmd,
		reportCmd,
		updateCmd,
	)
}

// loadConfig reads the config file selected via --config, falling back
// to Alexiares' defaults when the file does not exist.
func loadConfig() (config.Config, error) {
	return config.Load(cfgPath)
}

// outputFormat resolves the effective output format for a command:
// the --format flag, falling back to the config file's setting.
func outputFormat(cmd *cobra.Command, cfg config.Config) string {
	f, _ := cmd.Flags().GetString("format")
	if f != "" {
		return f
	}
	if cfg.Output.Format != "" {
		return cfg.Output.Format
	}
	return "terminal"
}
