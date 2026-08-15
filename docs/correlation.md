# `internal/correlation`

**Source:** [`internal/correlation/correlation.go`](../internal/correlation/correlation.go)
**Tests:** `correlation_test.go` — 100% coverage
**Position in pipeline:** fourth stage — matches `internal/fingerprint`'s output and the extractors' artifacts against `internal/intel`'s loaded signatures; feeds `internal/evidence` and `internal/graph`

## Purpose

The actual detection step. Everything before this stage collects and normalizes; this is where a scanned target gets checked against every loaded signature.

```go
type Target struct { Domain, Fingerprints, Wallets, Telegram, IPs, Nameservers, RedirectDomains }
func Correlate(target Target, signatures []intel.Signature) artifact.Correlation
```

`Correlate` never returns an error — a `Target` with mostly-zero fields (nothing found, no TLS handshake) simply produces zero matches. There's no "correlation failed" condition, only "correlation found nothing."

## What gets matched against what

Eight categories, each checked independently in `matchSignature`: favicon (both SHA256 and MurmurHash3 sub-checks — a signature can match on either), JavaScript hashes, TLS certificate fingerprint, wallet addresses (chain-scoped — an address only matches if both the address *and* its chain agree), Telegram indicators (substring match against a signature's patterns), domains (the target's own domain, and separately, its redirect chain's hostnames — two different `MatchCategory` values for the same signature field), IPs, and nameservers.

**Domain matches split into two categories on purpose.** `sig.Domains` is checked twice per signature: once against `target.Domain` directly (`MatchDomain` — "this scanned domain *is* a known malicious one") and once against every hostname in `target.RedirectDomains` (`MatchRedirect` — "this scanned domain *redirects to* a known malicious one"). Both are real, distinct evidence and `internal/evidence` weighs them differently (domain identity is strong; a redirect chain landing somewhere known is medium) — collapsing them into one category would lose that distinction.

## Design notes

**Confidence is never decided here.** The package doc states this directly: correlation "never itself decides confidence — the signature carries its own maintainer-assigned confidence." `Correlate` copies `sig.Confidence` verbatim into the resulting `Cluster`; picking the *overall* confidence across multiple matched clusters, and turning that into a recommendation, is `internal/evidence`'s job entirely.

**Per-cluster `RelatedDomains`/`RelatedWallets` exist specifically so `internal/graph` can draw accurate edges.** `Correlate` tracks two levels of "related infrastructure": a top-level union across every matched signature (`Correlation.RelatedDomains`), and a per-signature set scoped to just that cluster (`Cluster.RelatedDomains`). The per-cluster version is what lets the graph engine connect a sibling domain to *the specific shared node* (a favicon, say) the match actually went through, rather than drawing an edge to the target with no way to say why. See [`graph.md`](graph.md) for how that connection gets made.

**The scanned target's own domain is excluded from "related."** When building `clusterDomains`, `Correlate` skips any signature-listed domain equal to `target.Domain` — a signature that happens to list the target itself as one of its known domains shouldn't report the target as "related to itself."

**`slices.Concat` combines Telegram indicator sources for one pass.** `target.Telegram.BotTokens`, `.APIRefs`, and `.Links` are concatenated once per signature check rather than checked as three separate inner loops — a small readability choice (the code originally nested three `append` calls, which was harder to read than it needed to be for what it does).
