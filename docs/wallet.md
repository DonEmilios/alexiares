# `internal/wallet`

**Source:** [`internal/wallet/wallet.go`](../internal/wallet/wallet.go)
**Tests:** `wallet_test.go` — 88.0% coverage
**Position in pipeline:** used both by `internal/collector.Classify` (CLI input routing) and directly inside the scan pipeline (detecting addresses embedded in a page's HTML/JS)

## Purpose

Detects and classifies cryptocurrency wallet addresses across 6 chains — Ethereum, Bitcoin, Solana, Cardano, Tron, TON — plus `.eth` ENS names. It has exactly two entry points and no others:

```go
func Detect(text string) artifact.WalletArtifacts   // find addresses embedded in arbitrary text
func Classify(address string) (artifact.WalletChain, bool)  // is this string, in full, a valid address?
```

## Design notes

**This is format detection, not checksum validation**, and the package doc says so explicitly. A string matching Ethereum's `0x` + 40 hex chars shape is reported as a candidate Ethereum address whether or not its EIP-55 checksum is actually valid. Real drainer kits don't reliably checksum their own addresses either, so validating checksums would filter out some of the exact addresses this tool exists to catch. Full checksum validation (which would need Keccak256 for Ethereum, an extra dependency) is out of scope for v1.

**Pattern order breaks a real ambiguity.** Several chains' address formats are literally indistinguishable from each other by shape alone — Solana's format is a generic 32–44 character base58 string with no prefix, and both Bitcoin legacy addresses and Tron addresses (`T` + 33 base58 chars) are valid base58 strings that also fit Solana's length range. `Classify` returns the *first* matching chain in a deliberately ordered list: prefixed formats (Ethereum's `0x`, Bitcoin's `1`/`3`/`bc1`, Cardano's `addr1`, Tron's `T`, TON's `EQ`/`UQ`) are checked before Solana's un-prefixed catch-all pattern. Get the order wrong and a Tron address silently misclassifies as Solana — this was caught by a test during the build (`TestClassify` failed on exactly this before the reorder) and is the reason the ordering has an explicit comment in the source rather than being implicit.

**Patterns are precompiled once, not per-call.** Each chain has both a substring-search regex (word-boundary delimited, for `Detect`) and a full-string-anchored regex (for `Classify`), both built once at package init via `newPattern`, not recompiled on every call — `Classify` used to do this the slow way (compile a fresh anchored regex every invocation) before being cleaned up.

**Bitcoin token boundaries needed a fix mid-build.** A bot-token-style regex was originally anchored with `\b` before a digit run, but Telegram bot tokens appear directly after a `bot` prefix in real API URLs (`.../bot123456789:TOKEN/...`) — and a letter-to-digit transition (`t` → `1`) is *not* a word boundary in regex terms, since both are "word" characters. (This bit `internal/telegram`, not `wallet`, but it's the same class of regex-boundary mistake — see [`telegram.md`](telegram.md).)

## What it deliberately does not do

- **No correlation against known drainer wallets.** `alexiares wallet <address>` and the wallet branch of `alexiares scan` both currently only classify — they report the chain and stop. Matching a detected address against signature-listed malicious wallets happens in `internal/correlation`, not here; this package's job ends at "is this a well-formed address, and which chain."
