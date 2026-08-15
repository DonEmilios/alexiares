package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/dns"
)

var dnsCmd = &cobra.Command{
	Use:   "dns <domain>",
	Short: "Display DNS intelligence for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		domain := hostnameOf(args[0])
		ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Network.Timeout())
		defer cancel()

		out := dns.New().Lookup(ctx, domain)
		return printDNSArtifacts(cmd.OutOrStdout(), domain, out)
	},
}

// hostnameOf strips a scheme and path from raw if present, so both
// "example.xyz" and "https://example.xyz/path" resolve the same host.
func hostnameOf(raw string) string {
	if strings.Contains(raw, "://") {
		if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
			return u.Hostname()
		}
	}
	return raw
}

// printDNSArtifacts renders DNS records as plain terminal text. This
// is a placeholder rendering: the full terminal/JSON/GraphML/DOT/CSV/
// Markdown output engine lands in Phase 7.
func printDNSArtifacts(w io.Writer, domain string, d artifact.DNSArtifacts) error {
	ew := &errWriter{w: w}
	ew.printf("DNS intelligence for %s\n\n", domain)
	ew.printList("A", d.IPs)
	ew.printList("AAAA", d.AAAA)
	if d.CNAME != "" {
		ew.printf("CNAME: %s\n", d.CNAME)
	}
	ew.printList("MX", d.MX)
	ew.printList("NS", d.Nameservers)
	ew.printList("TXT", d.TXT)
	ew.printList("PTR", d.PTR)
	ew.printf("\n")
	return ew.err
}

// errWriter accumulates the first write error so a sequence of prints
// can skip individual error checks without silently discarding one.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, args ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, args...)
}

func (e *errWriter) printList(label string, values []string) {
	if len(values) == 0 {
		return
	}
	e.printf("%s:\n", label)
	for _, v := range values {
		e.printf("  %s\n", v)
	}
}
