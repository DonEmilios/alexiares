# `internal/artifact`

**Source:** [`internal/artifact/artifact.go`](../internal/artifact/artifact.go)
**Tests:** none directly — it's pure data types, exercised through every package that uses them

## Purpose

Defines every data type shared across the pipeline: what the collector produces, what each extractor produces, what the fingerprint engine produces, and what the correlation engine produces. It has no logic — no functions, only types and constants — and no dependencies beyond the standard library.

This is the load-bearing design decision of the whole codebase: **extractors never import each other.** `internal/html` doesn't know `internal/javascript` exists, `internal/dns` doesn't know `internal/wallet` exists. They all import `artifact` and only `artifact`, and communicate exclusively by producing and consuming its types. That's what lets `go doc ./internal/html` be understood in isolation, and what let every extractor in this codebase get built, tested, and reviewed independently.

## What's in it

Grouped by where in the pipeline each type is produced:

| Produced by | Type |
|---|---|
| `internal/collector` | `RawResponse`, `Redirect`, `TLSData` |
| `internal/dns` | `DNSArtifacts` |
| `internal/html` | `HTMLArtifacts`, `Form` |
| `internal/javascript` | `JavaScriptArtifacts` |
| `internal/favicon` | `FaviconArtifacts` |
| `internal/wallet` | `WalletArtifacts`, `WalletAddress`, `WalletChain` (+ 6 chain constants) |
| `internal/telegram` | `TelegramArtifacts` |
| `internal/fingerprint` | `Fingerprints` |
| `internal/correlation` | `Correlation`, `Cluster`, `Match`, `MatchCategory` (+ 9 category constants) |
| everywhere | `Timeline` |

## Design notes

**`Timeline` is embedded, not bolted on.** `RawResponse.Timeline` is set once, at collection time (`FirstSeen`/`LastSeen`). The spec's four timestamp fields (`FirstSeen`, `ReportedAt`, `VerifiedAt`, `LastSeen`) exist so a signature's "when was this first seen doing this" is never confused with "when did an analyst report it" or "when did a maintainer verify it" — three genuinely different moments that a single `CreatedAt` field would collapse together.

**`Fingerprints.HTML` vs `Fingerprints.HTMLSimilarity` are deliberately two different hashes**, not one. `HTML` is a SHA256 over the DOM's exact tag-nesting shape (see [`fingerprint.md`](fingerprint.md)) — it matches a byte-identical clone and nothing else. `HTMLSimilarity` is a 64-bit SimHash — it matches a *near*-identical clone (same template, different copy/images) via Hamming distance. Collapsing these into one field would have meant picking one matching strategy and losing the other.

**`MatchCategory` is a closed, spec-derived enum** (9 constants: favicon, javascript, certificate, wallet, telegram, domain, redirect, ip, nameserver) — not an arbitrary string. `internal/evidence` switches on it exhaustively to assign strength tiers; a typo'd category string would silently fall through to "weak" instead of failing to compile.

**`Cluster` carries its own `RelatedDomains`/`RelatedWallets`**, separate from `Correlation`'s top-level aggregate of the same name. The top-level fields are the union across every matched signature — useful for "what else is related to this scan, period." The per-cluster fields are what actually let `internal/graph` draw an accurate edge: "domain X shares *this specific* favicon with the target" requires knowing which signature (and which of its indicators) X came from, which the flattened union alone can't answer.

## Known limitation

`DNSArtifacts.ASN` / `ASNOrg` are defined but never populated — no extractor produces them. See [`dns.md`](dns.md).
