# `internal/html`

**Source:** [`internal/html/html.go`](../internal/html/html.go)
**Tests:** `html_test.go` — 93.8% coverage
**Position in pipeline:** extractor stage — pure parse of `RawResponse.HTML`, no I/O

## Purpose

Extracts structural artifacts from already-collected HTML: forms (and their fields, split into visible vs. hidden), page metadata (`<meta>` tags, including Open Graph `property=` tags), comments, external resource references (`<script src>`, `<link href>`, `<img src>`, `<iframe src>`), and the page title.

```go
func Extract(rawHTML string) artifact.HTMLArtifacts
```

## Design notes

**Malformed HTML is tolerated, not rejected.** `Extract` returns the zero-value `HTMLArtifacts{}` only if the underlying parser (`golang.org/x/net/html`) fails outright — which is rare, since that parser implements the WHATWG HTML5 parsing algorithm and is built to handle the same tag-soup a real browser tolerates. A phishing page's HTML is rarely well-formed; treating "not valid XML" as an extraction error would make this extractor useless on exactly the pages it's meant to analyze.

**Hidden fields are separated from visible ones at extraction time, not later.** `extractForm` checks each `<input>`'s `type` attribute case-insensitively and splits into `Form.Fields` vs. `Form.HiddenFields` as it walks — hidden fields on a phishing form (CSRF-looking tokens, tracking IDs, pre-filled destination wallet addresses) are exactly the kind of detail worth surfacing distinctly rather than burying in a flat field list.

**Nested `<form>` elements are not double-walked.** `extractForm`'s inner walker explicitly skips descending into a child `<form>` node — it's picked up separately when the outer `Extract` walk reaches it as its own top-level element, so its fields aren't attributed to the wrong (outer) form.

**Metadata falls back from `name=` to `property=`.** A `<meta>` tag's key is read from `name` first, then `property` if `name` is absent — this is what makes Open Graph tags (`<meta property="og:title" ...>`, which use `property` instead of `name`) show up in `HTMLArtifacts.Metadata` without a separate code path.
