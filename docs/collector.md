# `internal/collector`

**Source:** [`internal/collector/`](../internal/collector/) (`collector.go`, `input.go`)
**Tests:** `collector_test.go`, `input_test.go`, `tls_internal_test.go` — 83.7% coverage
**Position in pipeline:** first stage — everything downstream reads from its output (`artifact.RawResponse`)

## Purpose

Two distinct jobs live in this package:

1. **`Classify`** — turn whatever string the user typed on the CLI (`example.xyz`, `https://example.xyz/path`, `203.0.113.1`, `0x742d...`) into a `Kind` (URL, IP, or wallet address on one of 6 chains) so the caller knows whether to fetch it over HTTP or hand it to `internal/wallet` directly.
2. **`Collector.Collect`** — actually fetch a classified target: the page, its TLS certificate, its external scripts, its favicon, and its redirect chain — safely.

"Safely" is not a soft claim here; it's the whole design brief. The package doc comment states it plainly: *never* executes JavaScript, *never* interacts with wallets or contracts, *never* submits a form, *never* does browser automation. It is a GET-only HTTP client that reads bytes. There is no code path in this package — or reachable from it — that could send a transaction, sign anything, or run arbitrary content.

## Public API

```go
func Classify(raw string) Input
func ReadTargets(r io.Reader) ([]string, error)

type Options struct { Timeout, UserAgent, MaxRedirects, MaxBodyBytes, MaxScripts, AllowPrivateNetworks }
func DefaultOptions() Options
func New(opts Options) *Collector
func (c *Collector) Collect(ctx context.Context, target string) (artifact.RawResponse, error)
```

`Input{Kind, Raw, URL, Chain}` is `Classify`'s result — `Kind` is one of `KindURL`, `KindIP`, `KindWallet`, `KindUnknown`. `ReadTargets` reads one target per line (skipping blanks and `#` comments), backing batch-file and stdin input.

## Design notes

**Wallet addresses are checked first, before anything URL-shaped**, because their formats never collide with a valid hostname or IP — there's no ambiguity to resolve, so it's a cheap, unconditional first check rather than a fallback.

**Redirects are capped and logged, not just followed.** `Collect` builds a fresh `CheckRedirect` closure per call (not shared on the `Collector`, so concurrent calls never cross-contaminate their redirect logs) that appends every hop to `RawResponse.Redirects` and stops after `MaxRedirects` (default 10) via `http.ErrUseLastResponse`. This is what lets `internal/graph` draw a `redirects_to` edge for the *entire* observed chain, not just the final destination.

**Every body read goes through `io.LimitReader`.** The main page, every external script, the favicon — all bounded by `MaxBodyBytes` (default 5MB). A hostile server streaming an infinite response can't exhaust memory; it just gets truncated.

**A failed script or favicon download doesn't fail the scan.** `fetchBytes` returns `nil` on any error (bad status, timeout, malformed URL) and callers treat `nil` as "absent," not "fatal." This is the same "partial collection is a result, not an error" policy the DNS extractor follows — see [`dns.md`](dns.md).

**Two separate `http.Client`s exist inside `Collector`** — one built per-`Collect()`-call for the primary page fetch (so its redirect log is call-scoped), and one long-lived one (`assets`) for script/favicon fetches, which don't need redirect *logging*, just a hop cap. Both share the same `*http.Transport`, and therefore the same `DialContext` — including its SSRF protection (below) — so the fix only has to live in one place.

**Every connection is resolved and validated before it's dialed, by default.** `safeDialContext` is Alexiares' answer to the fact that it exists specifically to fetch attacker-controlled URLs: a malicious target's DNS answer or a redirect `Location` header could otherwise point the scanning host at its own loopback interface, an RFC 1918 private address, or a cloud instance's `169.254.169.254` metadata endpoint. The dial function resolves the hostname exactly once and connects to the specific IP it validated as public — it never does a second, independent lookup at connect time, which is what closes the DNS-rebinding gap a naive "check the hostname, then let the transport resolve it again" approach would leave open. This applies uniformly to the initial request and every redirect hop, since they all go through the same `DialContext`. Set `Options.AllowPrivateNetworks` to disable this — only for an operator who intends to point Alexiares at their own internal infrastructure.

**Asset discovery is a real HTML parse (`golang.org/x/net/html`), not a regex.** `discoverAssets` walks the parsed tree for `<script src>` and `<link rel="icon">`, resolving relative URLs against the final (post-redirect) URL. If no favicon link is found, it falls back to `/favicon.ico` by convention.

## What it deliberately does not do

- **No inline script extraction.** Inline `<script>` bodies are part of `raw.HTML` already (regular text in the page source) — `internal/javascript` pulls them back out by re-parsing `raw.HTML`, rather than the collector duplicating that work.
- **No DNS record collection.** The collector's implicit resolution (via `net/http`'s dialer) only proves a hostname resolves — it doesn't surface A/AAAA/MX/NS/TXT record sets. That's `internal/dns`'s job, run as a separate step.
- **No TLS handshake parsing.** `Collect` hands `resp.TLS` straight to `internal/tls.FromConnectionState` — see [`tls.md`](tls.md).
