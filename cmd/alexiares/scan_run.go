package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/collector"
	"github.com/alexiares/alexiares/internal/config"
	"github.com/alexiares/alexiares/internal/correlation"
	"github.com/alexiares/alexiares/internal/dns"
	"github.com/alexiares/alexiares/internal/evidence"
	"github.com/alexiares/alexiares/internal/favicon"
	"github.com/alexiares/alexiares/internal/fingerprint"
	"github.com/alexiares/alexiares/internal/graph"
	"github.com/alexiares/alexiares/internal/html"
	"github.com/alexiares/alexiares/internal/intel"
	"github.com/alexiares/alexiares/internal/javascript"
	"github.com/alexiares/alexiares/internal/output"
	"github.com/alexiares/alexiares/internal/redirect"
	"github.com/alexiares/alexiares/internal/telegram"
	"github.com/alexiares/alexiares/internal/wallet"
)

// runScan performs a complete infrastructure analysis: collect,
// extract, fingerprint, correlate against the local signature
// repository, evaluate the evidence, and render the result.
//
// A wallet-address target skips network collection entirely — there
// is no infrastructure to fetch — and is classified directly.
func runScan(cmd *cobra.Command, target string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	classified := collector.Classify(target)
	switch classified.Kind {
	case collector.KindWallet:
		return runWalletScan(cmd, classified)
	case collector.KindURL, collector.KindIP:
		return runInfrastructureScan(cmd, cfg, classified)
	default:
		return fmt.Errorf("scan: %q is not a recognized URL, domain, IP, or wallet address", target)
	}
}

func runWalletScan(cmd *cobra.Command, target collector.Input) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\nchain: %s\n\nCorrelation against known infrastructure is not yet implemented for standalone wallet scans.\n", target.Raw, target.Chain)
	return err
}

func runInfrastructureScan(cmd *cobra.Command, cfg config.Config, target collector.Input) error {
	ctx := cmd.Context()
	domain := hostnameOf(target.URL)

	c := collector.New(collector.Options{
		Timeout:   cfg.Network.Timeout(),
		UserAgent: cfg.Network.UserAgent,
	})
	raw, err := c.Collect(ctx, target.URL)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	dnsArtifacts := dns.New().Lookup(ctx, domain)
	htmlArtifacts := html.Extract(raw.HTML)
	jsArtifacts := javascript.Extract(raw.HTML, raw.ScriptURLs, raw.Scripts)
	favArtifacts := favicon.Compute(raw.Favicon)
	fp := fingerprint.Compute(raw, favArtifacts, jsArtifacts)

	// Wallet and Telegram detection scan the page text plus every
	// downloaded external script: raw.HTML alone would still catch
	// inline <script> bodies (they're part of the HTML text), but not
	// indicators that only appear in an externally hosted script.
	combinedText := raw.HTML + "\n" + strings.Join(raw.Scripts, "\n")
	walletArtifacts := wallet.Detect(combinedText)
	telegramArtifacts := telegram.Extract(combinedText)

	redirects := append(append([]artifact.Redirect{}, raw.Redirects...), redirect.Extract(raw.HTML, raw.Scripts, raw.FinalURL)...)

	sigs, err := intel.LoadSignatures(cfg.Signatures.Path)
	if err != nil {
		return fmt.Errorf("scan: loading signatures: %w", err)
	}

	corr := correlation.Correlate(correlation.Target{
		Domain:          domain,
		Fingerprints:    fp,
		Wallets:         walletArtifacts,
		Telegram:        telegramArtifacts,
		IPs:             dnsArtifacts.IPs,
		Nameservers:     dnsArtifacts.Nameservers,
		RedirectDomains: redirectHostnames(redirects),
	}, sigs)

	report := evidence.Evaluate(target.URL, corr)

	g := graph.FromScan(graph.ScanData{
		Domain:       domain,
		IPs:          dnsArtifacts.IPs,
		Nameservers:  dnsArtifacts.Nameservers,
		Fingerprints: fp,
		Wallets:      walletArtifacts,
		Redirects:    redirects,
		Correlation:  corr,
	})

	result := output.FromReport(report, output.Artifacts{
		DNS:          dnsArtifacts,
		TLS:          raw.TLS,
		HTML:         htmlArtifacts,
		JavaScript:   jsArtifacts,
		Favicon:      favArtifacts,
		Wallets:      walletArtifacts,
		Telegram:     telegramArtifacts,
		Redirects:    redirects,
		Fingerprints: fp,
	}, g)

	format := output.Format(outputFormat(cmd, cfg))
	rendered, err := output.Render(format, result)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), rendered)
	return err
}

// redirectHostnames extracts the hostname of each redirect's
// destination, for domain-based correlation matching.
func redirectHostnames(redirects []artifact.Redirect) []string {
	var out []string
	for _, r := range redirects {
		if u, err := url.Parse(r.To); err == nil && u.Hostname() != "" {
			out = append(out, u.Hostname())
		}
	}
	return out
}
