package main

import (
	"fmt"

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
	// Args is explicit (rather than left nil) so a typo'd subcommand
	// still fails clearly instead of silently falling through to RunE
	// below as a positional argument. Cobra's own "unknown command"
	// hint line is a special case tied to command-not-found resolution
	// and doesn't fire once root has a RunE — so the hint is embedded
	// directly in this error's message (as a second line) to keep the
	// same two-line UX a not-yet-runnable root would have produced.
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			// The capitalized, punctuated second line is deliberate:
			// this error is always displayed standalone at the top
			// level, never wrapped by a caller, so it's authored as
			// the two lines of CLI text a user reads, not as a Go
			// error meant to compose into a longer chain.
			//nolint:revive,staticcheck // deliberately capitalized/punctuated CLI-facing text; see comment above
			return fmt.Errorf("unknown command %q for %q\nRun '%s --help' for usage.", args[0], cmd.CommandPath(), cmd.CommandPath())
		}
		return nil
	},
	// RunE only fires for a truly bare `alexiares` with no subcommand
	// and no stray arguments — Args above rejects anything else before
	// this ever runs. The banner is gated on isInteractiveTerminal so
	// it never reaches piped, redirected, or CI-captured output.
	RunE: func(cmd *cobra.Command, _ []string) error {
		if isInteractiveTerminal() {
			if _, err := cmd.OutOrStdout().Write([]byte(banner)); err != nil {
				return err
			}
		}
		return cmd.Help()
	},
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
