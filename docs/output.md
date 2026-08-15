# `internal/output`

**Source:** [`internal/output/`](../internal/output/) (`result.go`, `render.go`, `terminal.go`, `json.go`, `csv.go`, `markdown.go`)
**Tests:** `output_test.go` (+ `tests/golden/output_test.go` locking byte-for-byte stability) — 87.7% coverage
**Position in pipeline:** final stage — renders `evidence.Report` + artifacts + graph into whatever format the user asked for

## Purpose

Bundles a completed scan into one struct and renders it six ways. It performs no detection or evaluation of its own — the package doc is explicit that `internal/evidence` has already decided the conclusion; this package only presents it.

```go
type Artifacts struct { DNS, TLS, HTML, JavaScript, Favicon, Wallets, Telegram, Redirects, Fingerprints }
type ScanResult struct { Target, Confidence, Evidence, Artifacts, Graph, Recommendation, RelatedDomains, RelatedWallets }
func FromReport(report evidence.Report, artifacts Artifacts, g graph.Graph) ScanResult

func Render(format Format, r ScanResult) (string, error)  // dispatches to one of the six below
func RenderTerminal(r ScanResult) string
func RenderJSON(r ScanResult) (string, error)
func RenderCSV(r ScanResult) (string, error)
func RenderMarkdown(r ScanResult) string
// DOT and GraphML are dispatched straight through to internal/graph.WriteDOT / WriteGraphML on r.Graph
```

## `ScanResult` is the spec's JSON Output Schema, plus what the other renderers need

The spec's JSON Output Schema names exactly six top-level keys: `target`, `confidence`, `evidence`, `artifacts`, `graph`, `recommendation`. `ScanResult` has all six, in that shape, verified by a test that unmarshals `RenderJSON`'s output and checks every key is present. `RelatedDomains`/`RelatedWallets` are additive beyond the spec's minimal schema — they exist because the terminal and Markdown renderers need them for their own "related infrastructure" section, and duplicating that data by re-deriving it from `Evidence` at render time would be more code than just carrying it through.

**`Artifacts` bundles every extractor's output, not just what evidence/correlation needed**, specifically so a saved JSON file is self-sufficient. `alexiares report result.json` re-renders a previously saved scan into any other format without re-scanning — that only works if the saved JSON actually contains the DNS records, TLS cert, HTML forms, and everything else, not just the subset that happened to matter for detection.

## Design notes

**Terminal, CSV, and Markdown all handle "nothing detected" as a real, distinct case**, not an empty table. `RenderTerminal` short-circuits to `"No known malicious infrastructure detected."` when `Evidence` is empty, rather than printing empty Evidence/Confidence/Recommendation sections. `RenderCSV` still emits exactly one row in that case (target + confidence + recommendation, with the per-evidence columns blank) rather than a header with zero data rows — a CSV consumer processing many scans wants one row per target regardless of outcome, not a variable row count depending on whether anything was found.

**CSV goes through `encoding/csv`, not string-joining with commas.** Evidence descriptions and recommendations are free text that can legitimately contain commas or quotes; hand-built CSV would corrupt on exactly that content. The standard library writer handles quoting correctly.

**GraphML/DOT aren't reimplemented here** — `Render` dispatches straight to `graph.WriteGraphML(r.Graph)` / `graph.WriteDOT(r.Graph)`. This package's own five files handle the five formats that operate on the *whole* `ScanResult`; the two graph-specific formats already have a correct, tested implementation one layer down, so `Render` just forwards to it rather than wrapping or duplicating.

**A golden test locks all five non-DOT-duplicate formats' exact output**, not just their structure — `tests/golden/output_test.go` runs a fixed `ScanResult` through terminal, JSON, CSV, Markdown, and DOT, and diffs against checked-in fixtures byte-for-byte. This is what "output stability," one of the spec's explicit testing tiers, actually means in this codebase: a change that alters formatting anywhere fails CI immediately, with a diff, rather than being caught later by someone noticing the terminal output looks different.
