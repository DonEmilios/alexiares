package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/artifact"
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
			Timeout:              cfg.Network.Timeout(),
			UserAgent:            cfg.Network.UserAgent,
			AllowPrivateNetworks: cfg.Network.AllowPrivateNetworks,
		})
		raw, err := c.Collect(cmd.Context(), target.URL)
		if err != nil {
			return err
		}
		return printTLSCertificate(cmd.OutOrStdout(), target.URL, raw.TLS)
	},
}

// printTLSCertificate renders certificate intelligence as plain
// terminal text, or a "no certificate" message when data is nil (a
// plain-HTTP target). Extracted from RunE so it's testable without a
// live TLS handshake.
func printTLSCertificate(w io.Writer, target string, data *artifact.TLSData) error {
	if data == nil {
		_, err := fmt.Fprintf(w, "%s did not present a TLS certificate\n", target)
		return err
	}

	ew := &errWriter{w: w}
	ew.printf("Certificate intelligence for %s\n\n", target)
	ew.printf("SHA256:      %s\n", data.SHA256)
	ew.printf("Serial:      %s\n", data.SerialHex)
	ew.printf("Issuer:      %s\n", data.Issuer)
	ew.printf("Subject:     %s\n", data.Subject)
	ew.printf("Key:         %s (%d bits)\n", data.KeyType, data.KeyBits)
	ew.printf("Valid from:  %s\n", data.NotBefore.Format("2006-01-02"))
	ew.printf("Valid until: %s\n", data.NotAfter.Format("2006-01-02"))
	ew.printList("SAN", data.DNSNames)
	return ew.err
}
