# `internal/fingerprint`

**Source:** [`internal/fingerprint/`](../internal/fingerprint/) (`fingerprint.go`, `html.go`, `simhash.go`)
**Tests:** `fingerprint_test.go`, `html_test.go`, `simhash_test.go` — 96.8% coverage
**Position in pipeline:** third stage — aggregates extractor output into one comparable identifier set; feeds `internal/correlation`

## Purpose

Normalizes a scanned target's artifacts into `artifact.Fingerprints` — the single struct `internal/correlation` matches against every signature.

```go
func Compute(raw artifact.RawResponse, fav artifact.FaviconArtifacts, js artifact.JavaScriptArtifacts) artifact.Fingerprints
```

Every function in this package is pure — no I/O, and the package doc states the determinism guarantee explicitly: identical input always produces identical output, which the correlation engine depends on for byte-for-byte signature matching.

## What it actually computes (two of the four fields are new work here, two are pass-through)

- `Favicon`, `FaviconHash`, `JavaScript`, `Certificate` — copied straight from `fav`, `js`, and `raw.TLS`. The hashing already happened in `internal/favicon`, `internal/javascript`, and `internal/tls`; this package just gathers them into one struct.
- `HTML` and `HTMLSimilarity` — genuinely computed here, and this is the interesting part of the package.

### `HTML`: exact structural hash

`structuralHash` (in `html.go`) walks the parsed DOM and builds a bracket-nested serialization of element tag names *only* — `(html(head(title))(body(div(form(input)(input)))))` — ignoring all text, attributes, and comments, then SHA256s that string. Two pages built from the same template hash identically even if their wallet address, copy, or images differ completely; change the template's shape at all (add a `<span>`, reorder elements) and the hash changes completely too.

### `HTMLSimilarity`: fuzzy structural fingerprint (SimHash)

This is the one piece of real algorithm design in the whole codebase, and it exists to solve a specific problem the exact hash above can't: a phishing kit's clones are rarely byte-identical in structure — one clone adds a tracking `<div>`, another reorders a nav. `structuralHash` would report those as completely unrelated. `simHash` (in `simhash.go`) doesn't:

1. `tagSequence` flattens the DOM into a preorder list of tag names (no bracketing this time — just the linear sequence).
2. That sequence is broken into overlapping 3-tag shingles (`shingleSize = 3` — small enough to stay sensitive, large enough not to overweight generic runs like `div div div`).
3. Each shingle is hashed with FNV-1a 64-bit.
4. For each of the 64 bit positions, every shingle "votes" +1 if its hash has that bit set, −1 if not; the final bit is 1 if the net vote is positive.

The result is a **locality-sensitive** fingerprint: two DOMs that share most of their shingles produce hashes with a small Hamming distance, even if they're not byte-identical. `HammingDistance64(a, b uint64) int` is exported specifically so `internal/correlation` (or a future fuzzy-matching signature) can compare two SimHash values by "how many bits differ" rather than "are they equal." A sequence shorter than one shingle (fewer than 3 tags) hashes to `0` — there are no shingles to vote, so there's nothing meaningful to fingerprint.

This was verified with three kinds of tests, not just "does it run": identical input produces identical output; a near-duplicate (same skeleton, one inserted `<span>`) produces a *small* Hamming distance; and an unrelated DOM produces a *larger* one than the near-duplicate does — proving the metric actually orders "how similar" correctly, not just "different or same."

## Design notes

**A golden test locks the whole `Compute` output**, not just the individual pieces — `tests/golden/fingerprint_test.go` runs a fixed sample page through the entire function and compares the JSON byte-for-byte against a checked-in fixture, catching any change to any of the four fields (including unintended ones from an unrelated refactor) in one place.

**When TLS is absent, `Certificate` is left empty, not a placeholder string.** `Compute` checks `raw.TLS != nil` before reading `raw.TLS.SHA256` — a plain-HTTP target's fingerprint has a genuinely empty certificate field, not `"none"` or a zero-hash sentinel that could accidentally collide with a real (if unlikely) signature entry.
