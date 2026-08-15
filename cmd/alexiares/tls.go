package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/collector"
)

var tlsCmd = &cobra.Command{
	Use:   "tls <domain>",
	Short: "Display certificate intelligence for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		target := collector.Classify(args[0])
		if target.Kind != collector.KindURL && target.Kind != collector.KindIP {
			return fmt.Errorf("tls: %q is not a fetchable domain, URL, or IP", args[0])
		}

		c := collector.New(collector.Options{
			Timeout:   cfg.Network.Timeout(),
			UserAgent: cfg.Network.UserAgent,
		})
		raw, err := c.Collect(cmd.Context(), target.URL)
		if err != nil {
			return err
		}
		if raw.TLS == nil {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s did not present a TLS certificate\n", target.URL)
			return err
		}

		ew := &errWriter{w: cmd.OutOrStdout()}
		ew.printf("Certificate intelligence for %s\n\n", target.URL)
		ew.printf("SHA256:      %s\n", raw.TLS.SHA256)
		ew.printf("Serial:      %s\n", raw.TLS.SerialHex)
		ew.printf("Issuer:      %s\n", raw.TLS.Issuer)
		ew.printf("Subject:     %s\n", raw.TLS.Subject)
		ew.printf("Key:         %s (%d bits)\n", raw.TLS.KeyType, raw.TLS.KeyBits)
		ew.printf("Valid from:  %s\n", raw.TLS.NotBefore.Format("2006-01-02"))
		ew.printf("Valid until: %s\n", raw.TLS.NotAfter.Format("2006-01-02"))
		ew.printList("SAN", raw.TLS.DNSNames)
		return ew.err
	},
}
