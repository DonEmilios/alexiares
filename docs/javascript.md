# `internal/javascript`

**Source:** [`internal/javascript/javascript.go`](../internal/javascript/javascript.go)
**Tests:** `javascript_test.go` — 98.1% coverage
**Position in pipeline:** extractor stage — pure text/pattern analysis, never executes anything

## Purpose

Analyzes JavaScript — both inline `<script>` bodies and externally downloaded script content — without ever running it: per-script SHA256 hashes (feeding `internal/fingerprint`), referenced API endpoints, Telegram/Discord/WalletConnect indicators, and known wallet-library usage.

```go
func Extract(rawHTML string, scriptURLs, externalScripts []string) artifact.JavaScriptArtifacts
```

## Design notes

**Inline scripts are recovered by re-parsing `rawHTML`, not passed in.** `internal/collector` never separates inline script bodies out on its own — they're already sitting in `RawResponse.HTML` as ordinary page text. `Extract` calls its own small HTML walker (`inlineScripts`) to pull out every `<script>` element that has no `src` attribute, then combines those with the caller-supplied `externalScripts` (the bodies `internal/collector` *did* download, keyed by URL) into one set for hashing and pattern-matching. This is the same trade-off `internal/collector`'s design notes mention: the collector doesn't duplicate parsing work the extractor needs anyway.

**Regex categories are independent passes, not one big regex.** `apiEndpointRe`, `telegramRe`, `discordWebhookRe`, `walletConnectRe` each run separately over the joined script text and each result set is deduplicated and sorted independently — this is what keeps `Extract`'s output deterministic byte-for-byte regardless of which script happened to contain which indicator first.

**Wallet library detection is a literal substring list, not a regex.** `walletLibraries` (`web3.js`, `ethers.js`, `wagmi`, `viem`, `window.ethereum`, `window.solana`, `@walletconnect`, etc.) is checked with plain `strings.Contains` against the combined script text. A regex would add complexity for zero benefit here — these are fixed strings, not patterns with variable parts.

**API endpoint matching is intentionally broad, not curated.** `apiEndpointRe` matches any `https?://` URL-shaped string in a script — it's not trying to distinguish "legitimate CDN" from "exfiltration endpoint." That judgment is `internal/evidence`'s job, applied against the signature match, not this extractor's.
