# `internal/favicon`

**Source:** [`internal/favicon/`](../internal/favicon/) (`favicon.go`, `murmur3.go`)
**Tests:** `favicon_test.go`, `murmur3_test.go` — 100% coverage
**Position in pipeline:** extractor stage — pure hashing, no I/O

## Purpose

Computes two deterministic identifiers from a favicon's raw bytes:

```go
func Compute(data []byte) artifact.FaviconArtifacts  // { Murmur3 int32, SHA256 string, Size int }
```

- **SHA256** — exact-byte matching.
- **MurmurHash3 (32-bit x86)** — computed over a specific base64-wrapped encoding, deliberately matching the convention Shodan's `http.favicon.hash` uses (and which most public malicious-favicon-hash datasets are built on). Matching that exact convention, not inventing a new one, is what makes Alexiares' favicon signatures interoperable with existing community intelligence instead of a second, incompatible hash space.

## Design notes

**MurmurHash3 was implemented from scratch, not imported, and then independently verified.** There's no dependency added for one hash function; `murmur3.go` is a from-scratch, documented implementation of the 32-bit x86 variant. Because a self-written hash function can't self-validate — a bug could produce internally-consistent-but-wrong output forever — it was checked against an independently-generated reference: a temporary Python virtualenv with the `mmh3` package computed reference values for six test strings, which are hardcoded into `murmur3_test.go` with a comment explaining exactly how they were produced. All six matched on the first run.

**The Shodan convention's exact framing matters and is documented as load-bearing.** The favicon bytes are base64-encoded (`encoding/base64`, standard alphabet), then wrapped with a newline every 76 characters *including a trailing newline* — matching Python's `base64.encodebytes` line-wrapping behavior — and *that* wrapped string, not the raw favicon bytes, is what gets hashed. Get the wrapping wrong (skip the trailing newline, use a different line width) and every hash silently stops matching Shodan-sourced signature data while still looking like a valid hash. This was verified end-to-end against a second independent Python-generated reference (256 bytes of `bytes(range(256))`, `base64.encodebytes` + `mmh3.hash`), not just the raw MurmurHash3 algorithm in isolation.

**Empty input returns the zero value, not an error.** A favicon that failed to download (`internal/collector.fetchBytes` returning `nil`) is a normal, expected outcome — `Compute(nil)` returns `artifact.FaviconArtifacts{}` rather than forcing every caller to nil-check first.
