# `internal/redirect`

**Source:** [`internal/redirect/redirect.go`](../internal/redirect/redirect.go)
**Tests:** `redirect_test.go` — 93.7% coverage
**Position in pipeline:** extractor stage — pure parse of already-collected HTML/scripts, no I/O

## Purpose

The spec names three redirect vectors a phishing page can use: HTTP-level redirects, HTML `<meta http-equiv="refresh">` tags, and JavaScript location reassignment. `internal/collector` already captures the first kind live, during the fetch itself (see [`collector.md`](collector.md)) — this package fills in the other two, which only become visible after the page content has already been collected.

```go
func Extract(rawHTML string, scripts []string, baseURL string) []artifact.Redirect
```

Each result carries a `Method` of `"meta_refresh"` or `"javascript"`, distinguishing it from the collector's own `"http"`-tagged entries in the same `[]artifact.Redirect` slice type.

## Design notes

**Relative redirect targets are resolved against `baseURL`**, the same way `internal/collector` resolves relative script/favicon URLs — a `<meta http-equiv="refresh" content="3; url=/step2">` on `https://phish.example/start` correctly resolves to `https://phish.example/step2`, not a dangling relative path, so downstream domain-based correlation (`internal/correlation`'s redirect-hostname matching) has something meaningful to compare.

**JavaScript redirects are found by pattern matching source text, not by execution.** `jsRedirectRe` matches `location.href = "..."`, `window.location = "..."`, `location.replace(...)`, and `location.assign(...)` as literal source patterns — consistent with the project-wide rule that nothing in this codebase ever runs collected JavaScript. A redirect assigned via a computed expression (`location.href = getTarget()`) is invisible to this approach; that's an accepted limitation of static analysis, not an oversight.

**Inline script extraction is duplicated here, not shared with `internal/javascript`.** Both packages have their own small `inlineScripts`/`textContent`/`attr` helpers that do the same HTML-walk-for-`<script>`-without-`src` work. This is a deliberate, small duplication rather than a shared `htmlutil` package — the project's architecture rule is that extractors never import each other, and a handful of ~10-line private helper functions repeated across three packages is cheaper than the coupling a shared low-level utility package would introduce.

**Duplicate redirect targets within the same page are suppressed**, per-method — if a script sets `window.location` to the same URL twice, `Extract` reports it once.
