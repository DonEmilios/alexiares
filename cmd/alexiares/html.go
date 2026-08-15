package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/alexiares/alexiares/internal/collector"
	"github.com/alexiares/alexiares/internal/html"
)

var htmlCmd = &cobra.Command{
	Use:   "html <url>",
	Short: "Extract HTML artifacts from a target",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		target := collector.Classify(args[0])
		if target.Kind != collector.KindURL && target.Kind != collector.KindIP {
			return fmt.Errorf("html: %q is not a fetchable domain, URL, or IP", args[0])
		}

		c := collector.New(collector.Options{
			Timeout:   cfg.Network.Timeout(),
			UserAgent: cfg.Network.UserAgent,
		})
		raw, err := c.Collect(cmd.Context(), target.URL)
		if err != nil {
			return err
		}
		artifacts := html.Extract(raw.HTML)

		ew := &errWriter{w: cmd.OutOrStdout()}
		ew.printf("HTML artifacts for %s\n\n", target.URL)
		ew.printf("Title: %s\n", artifacts.Title)
		if len(artifacts.Metadata) > 0 {
			ew.printf("Metadata:\n")
			keys := make([]string, 0, len(artifacts.Metadata))
			for k := range artifacts.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				ew.printf("  %s: %s\n", k, artifacts.Metadata[k])
			}
		}
		ew.printList("Comments", artifacts.Comments)
		ew.printList("External resources", artifacts.ExternalResources)
		if len(artifacts.Forms) > 0 {
			ew.printf("Forms:\n")
			for _, f := range artifacts.Forms {
				ew.printf("  %s %s fields=%v hidden=%v\n", f.Method, f.Action, f.Fields, f.HiddenFields)
			}
		}
		return ew.err
	},
}
