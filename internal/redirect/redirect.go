// Package redirect detects non-HTTP redirect mechanisms: HTML
// meta-refresh tags and JavaScript location reassignments. HTTP-level
// redirects are already captured by the collector during acquisition;
// this package fills in the two other redirect vectors the spec
// requires, operating purely on already-collected HTML and script
// text with no network I/O.
package redirect

import (
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/alexiares/alexiares/internal/artifact"
)

var (
	metaRefreshURLRe = regexp.MustCompile(`(?i)url\s*=\s*(\S+)`)
	jsRedirectRe     = regexp.MustCompile(`(?:location\.href|window\.location(?:\.href)?|location)\s*(?:=|\.replace\(|\.assign\()\s*['"]([^'"]+)['"]`)
)

// Extract finds meta-refresh and JavaScript redirects in rawHTML and
// scripts. Target URLs are resolved against baseURL when relative.
func Extract(rawHTML string, scripts []string, baseURL string) []artifact.Redirect {
	base, _ := url.Parse(baseURL)

	var out []artifact.Redirect
	out = append(out, metaRefreshRedirects(rawHTML, baseURL, base)...)
	out = append(out, jsRedirects(rawHTML, scripts, baseURL, base)...)
	return out
}

func metaRefreshRedirects(rawHTML, baseURL string, base *url.URL) []artifact.Redirect {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil
	}

	var out []artifact.Redirect
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Meta {
			if strings.EqualFold(attr(n, "http-equiv"), "refresh") {
				if m := metaRefreshURLRe.FindStringSubmatch(attr(n, "content")); m != nil {
					target := strings.Trim(m[1], `'"`)
					out = append(out, artifact.Redirect{
						From:   baseURL,
						To:     resolve(base, target),
						Method: "meta_refresh",
					})
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return out
}

func jsRedirects(rawHTML string, externalScripts []string, baseURL string, base *url.URL) []artifact.Redirect {
	all := append(inlineScripts(rawHTML), externalScripts...)

	seen := make(map[string]bool)
	var out []artifact.Redirect
	for _, script := range all {
		for _, m := range jsRedirectRe.FindAllStringSubmatch(script, -1) {
			target := m[1]
			if seen[target] {
				continue
			}
			seen[target] = true
			out = append(out, artifact.Redirect{
				From:   baseURL,
				To:     resolve(base, target),
				Method: "javascript",
			})
		}
	}
	return out
}

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

func resolve(base *url.URL, ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if base == nil {
		return u.String()
	}
	return base.ResolveReference(u).String()
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
