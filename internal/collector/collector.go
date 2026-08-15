// Package collector safely acquires infrastructure data about a
// target: DNS resolution (implicit, via the HTTP client), HTTP(S)
// requests with redirect tracking, the TLS handshake, HTML, favicon,
// and script bytes, and response headers.
//
// Collector never executes JavaScript, never interacts with wallets or
// smart contracts, never submits forms, and never performs browser
// automation — it is a plain HTTP client that only reads bytes. Every
// operation is bounded by a caller-supplied timeout and every
// downloaded body is size-limited, so a hostile target cannot exhaust
// memory or hang a scan.
package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/alexiares/alexiares/internal/artifact"
	tlsextract "github.com/alexiares/alexiares/internal/tls"
)

// Options configures a Collector's network behavior. All fields have
// safe, spec-aligned defaults via DefaultOptions.
type Options struct {
	// Timeout bounds each individual HTTP request (the primary page,
	// and every subsequently fetched script or favicon).
	Timeout time.Duration
	// UserAgent is sent on every request.
	UserAgent string
	// MaxRedirects caps how many HTTP redirect hops are followed
	// before collection stops and reports the chain gathered so far.
	MaxRedirects int
	// MaxBodyBytes bounds how much of any single response body is
	// read, protecting against unbounded or malicious responses.
	MaxBodyBytes int64
	// MaxScripts caps how many external scripts are downloaded per
	// target.
	MaxScripts int
}

// DefaultOptions returns Alexiares' built-in collection defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:      10 * time.Second,
		UserAgent:    "Alexiares/0.1",
		MaxRedirects: 10,
		MaxBodyBytes: 5 << 20, // 5 MB
		MaxScripts:   25,
	}
}

// withDefaults fills any zero-value field in opts from DefaultOptions.
func (opts Options) withDefaults() Options {
	def := DefaultOptions()
	if opts.Timeout <= 0 {
		opts.Timeout = def.Timeout
	}
	if opts.UserAgent == "" {
		opts.UserAgent = def.UserAgent
	}
	if opts.MaxRedirects <= 0 {
		opts.MaxRedirects = def.MaxRedirects
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = def.MaxBodyBytes
	}
	if opts.MaxScripts <= 0 {
		opts.MaxScripts = def.MaxScripts
	}
	return opts
}

// Collector acquires RawResponse data for a target over HTTP(S).
//
// A Collector is safe for concurrent use: Collect builds a fresh
// per-call redirect log rather than mutating shared state.
type Collector struct {
	opts   Options
	assets *http.Client // used for script/favicon fetches; no redirect logging needed
}

// New builds a Collector from opts. Zero-value fields in opts fall
// back to DefaultOptions.
func New(opts Options) *Collector {
	opts = opts.withDefaults()
	return &Collector{
		opts: opts,
		assets: &http.Client{
			Timeout: opts.Timeout,
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= opts.MaxRedirects {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// Collect fetches target and returns everything Alexiares can safely
// observe about it. target must already be a fetchable URL (see
// Classify for turning raw CLI input into one).
func (c *Collector) Collect(ctx context.Context, target string) (artifact.RawResponse, error) {
	now := time.Now().UTC()

	var redirects []artifact.Redirect
	client := &http.Client{
		Timeout: c.opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				redirects = append(redirects, artifact.Redirect{
					From:   via[len(via)-1].URL.String(),
					To:     req.URL.String(),
					Method: "http",
				})
			}
			req.Header.Set("User-Agent", c.opts.UserAgent)
			if len(via) >= c.opts.MaxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return artifact.RawResponse{}, fmt.Errorf("building request for %s: %w", target, err)
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return artifact.RawResponse{}, fmt.Errorf("fetching %s: %w", target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.opts.MaxBodyBytes))
	if err != nil {
		return artifact.RawResponse{}, fmt.Errorf("reading response body from %s: %w", target, err)
	}
	htmlBody := string(body)

	raw := artifact.RawResponse{
		URL:        target,
		FinalURL:   resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		HTML:       htmlBody,
		Redirects:  redirects,
		TLS:        tlsextract.FromConnectionState(resp.TLS),
		Timeline:   artifact.Timeline{FirstSeen: now, LastSeen: now},
	}

	scriptURLs, faviconURL := discoverAssets(htmlBody, raw.FinalURL)
	raw.ScriptURLs = scriptURLs
	raw.Scripts = c.fetchScripts(ctx, scriptURLs)
	if faviconURL != "" {
		if b := c.fetchBytes(ctx, faviconURL); b != nil {
			raw.Favicon = b
		}
	}

	return raw, nil
}

// fetchScripts downloads up to MaxScripts external script bodies. A
// script that fails to download is skipped rather than aborting the
// scan — partial collection is a result, not an error.
func (c *Collector) fetchScripts(ctx context.Context, urls []string) []string {
	var out []string
	for i, u := range urls {
		if i >= c.opts.MaxScripts {
			break
		}
		if b := c.fetchBytes(ctx, u); b != nil {
			out = append(out, string(b))
		}
	}
	return out
}

// fetchBytes performs a bounded GET, returning nil on any failure so
// callers can treat a missing asset as absent rather than fatal.
func (c *Collector) fetchBytes(ctx context.Context, u string) []byte {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)

	resp, err := c.assets.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.opts.MaxBodyBytes))
	if err != nil {
		return nil
	}
	return body
}

// discoverAssets walks the parsed HTML for <script src> URLs and the
// favicon <link rel="icon"> href, resolving relative URLs against
// baseURL. Inline <script> content is read separately, directly from
// raw.HTML, by the JavaScript extractor.
func discoverAssets(htmlBody, baseURL string) (scriptURLs []string, faviconURL string) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, ""
	}

	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, ""
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Script:
				if src := attr(n, "src"); src != "" {
					if abs := resolve(base, src); abs != "" {
						scriptURLs = append(scriptURLs, abs)
					}
				}
			case atom.Link:
				rel := strings.ToLower(attr(n, "rel"))
				if strings.Contains(rel, "icon") {
					if href := attr(n, "href"); href != "" {
						if abs := resolve(base, href); abs != "" {
							faviconURL = abs
						}
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	if faviconURL == "" {
		if abs := resolve(base, "/favicon.ico"); abs != "" {
			faviconURL = abs
		}
	}
	return scriptURLs, faviconURL
}

// resolve returns ref resolved against base as an absolute URL
// string, or "" if either fails to parse.
func resolve(base *url.URL, ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	return base.ResolveReference(u).String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}
