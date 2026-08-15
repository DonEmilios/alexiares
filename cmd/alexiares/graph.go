package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/collector"
	"github.com/alexiares/alexiares/internal/dns"
	"github.com/alexiares/alexiares/internal/favicon"
	"github.com/alexiares/alexiares/internal/fingerprint"
	"github.com/alexiares/alexiares/internal/graph"
	"github.com/alexiares/alexiares/internal/javascript"
)

var graphCmd = &cobra.Command{
	Use:   "graph <domain>",
	Short: "Build the infrastructure relationship graph for a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		target := collector.Classify(args[0])
		if target.Kind != collector.KindURL && target.Kind != collector.KindIP {
			return fmt.Errorf("graph: %q is not a fetchable domain, URL, or IP", args[0])
		}
		domain := hostnameOf(args[0])

		c := collector.New(collector.Options{
			Timeout:   cfg.Network.Timeout(),
			UserAgent: cfg.Network.UserAgent,
		})
		raw, err := c.Collect(cmd.Context(), target.URL)
		if err != nil {
			return err
		}

		dnsArtifacts := dns.New().Lookup(cmd.Context(), domain)
		fav := favicon.Compute(raw.Favicon)
		js := javascript.Extract(raw.HTML, raw.ScriptURLs, raw.Scripts)
		fp := fingerprint.Compute(raw, fav, js)

		g := graph.FromScan(graph.ScanData{
			Domain:       domain,
			IPs:          dnsArtifacts.IPs,
			Nameservers:  dnsArtifacts.Nameservers,
			Fingerprints: fp,
			Redirects:    raw.Redirects,
		})

		format := outputFormat(cmd, cfg)
		return writeGraph(cmd, g, format)
	},
}

func writeGraph(cmd *cobra.Command, g graph.Graph, format string) error {
	switch format {
	case "dot":
		_, err := fmt.Fprint(cmd.OutOrStdout(), graph.WriteDOT(g))
		return err
	case "graphml":
		out, err := graph.WriteGraphML(g)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), out)
		return err
	case "json":
		out, err := graph.WriteJSON(g)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), out)
		return err
	case "", "terminal":
		_, err := fmt.Fprint(cmd.OutOrStdout(), graph.WriteDOT(g))
		return err
	default:
		return fmt.Errorf("graph: unsupported format %q (want dot, graphml, or json)", format)
	}
}
