// Package javascript extracts artifacts from collected JavaScript:
// per-script SHA256 hashes, referenced API endpoints, Telegram
// exfiltration indicators, Discord webhooks, wallet libraries, and
// WalletConnect references. It performs no network I/O and never
// executes any of the analyzed code — it only pattern-matches over
// text already gathered by the collector.
package javascript

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/alexiares/alexiares/internal/artifact"
)

var (
	apiEndpointRe    = regexp.MustCompile(`https?://[a-zA-Z0-9.-]+(?::[0-9]+)?(?:/[^\s'"` + "`" + `<>)]*)?`)
	telegramRe       = regexp.MustCompile(`(?:api\.telegram\.org/bot[A-Za-z0-9_:-]+|t\.me/[A-Za-z0-9_]+)`)
	discordWebhookRe = regexp.MustCompile(`discord(?:app)?\.com/api/webhooks/\d+/[A-Za-z0-9_-]+`)
	walletConnectRe  = regexp.MustCompile(`(?i)walletconnect|bridge\.walletconnect\.org|wc:[a-f0-9-]+@[12]`)
)

// walletLibraries are known wallet-integration library identifiers
// searched for as literal, case-sensitive substrings.
var walletLibraries = []string{
	"web3.js", "ethers.js", "wagmi", "viem", "web3modal",
	"window.ethereum", "window.solana", "@solana/web3.js",
	"@walletconnect", "web3-react",
}

// Extract combines the collector's downloaded external script bodies
// with inline <script> content parsed from rawHTML, then derives
// JavaScriptArtifacts from the combined set.
func Extract(rawHTML string, scriptURLs, externalScripts []string) artifact.JavaScriptArtifacts {
	inline := inlineScripts(rawHTML)
	all := make([]string, 0, len(inline)+len(externalScripts))
	all = append(all, inline...)
	all = append(all, externalScripts...)

	out := artifact.JavaScriptArtifacts{
		ScriptURLs:    scriptURLs,
		InlineScripts: inline,
	}

	combined := strings.Join(all, "\n")
	out.APIEndpoints = dedupSorted(apiEndpointRe.FindAllString(combined, -1))
	out.TelegramRefs = dedupSorted(telegramRe.FindAllString(combined, -1))
	out.DiscordWebhooks = dedupSorted(discordWebhookRe.FindAllString(combined, -1))
	out.WalletConnectRefs = dedupSorted(walletConnectRe.FindAllString(combined, -1))

	for _, s := range all {
		out.SHA256 = append(out.SHA256, fmt.Sprintf("%x", sha256.Sum256([]byte(s))))
	}

	var libs []string
	for _, lib := range walletLibraries {
		if strings.Contains(combined, lib) {
			libs = append(libs, lib)
		}
	}
	out.WalletLibraries = libs

	return out
}

// inlineScripts returns the text content of every <script> element in
// rawHTML that has no src attribute, in document order.
func inlineScripts(rawHTML string) []string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil
	}

	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Script && attr(n, "src") == "" {
			if body := textContent(n); strings.TrimSpace(body) != "" {
				out = append(out, body)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return out
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return b.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// dedupSorted returns matches deduplicated with a stable sort order,
// so Extract's output is deterministic regardless of match order.
func dedupSorted(matches []string) []string {
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(matches))
	var out []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	sort.Strings(out)
	return out
}
