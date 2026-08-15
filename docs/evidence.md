# `internal/evidence`

**Source:** [`internal/evidence/evidence.go`](../internal/evidence/evidence.go)
**Tests:** `evidence_test.go` — 75.6% coverage (just under the project's 80% floor — see Known limitation below)
**Position in pipeline:** sixth stage — turns `internal/correlation`'s output into a human-readable verdict; feeds `internal/output`

## Purpose

The spec asks every scan to answer four questions: what was found, why is it suspicious, what infrastructure is related, how confident is the conclusion. `internal/correlation` answers the first and third; this package answers the second and fourth, and packages all of it into one `Report`.

```go
func Evaluate(target string, corr artifact.Correlation) Report

type Report struct {
    Target, Confidence, Recommendation string
    Evidence []Item
    RelatedDomains, RelatedWallets []string
}
```

The package doc is explicit that this is evaluation only, not detection: "It performs no detection of its own — `internal/correlation` already decided what matched — it only evaluates, weighs, and explains."

## The strength model

Every `Match` category gets classified into one of three tiers via `strengthOf`, directly from the spec's Evidence Categories table:

- **Strong** — favicon, JavaScript, certificate, wallet, Telegram, **and domain**.
- **Medium** — IP, nameserver, redirect.
- **Weak** — the default case (currently unreachable, since every defined `MatchCategory` is explicitly classified above — see Known limitation).

**Domain matches are strong, and that's a deliberate extension beyond the spec's literal wording**, not an oversight. The spec's table doesn't list "domain" explicitly among its strong examples, but a domain match means the scanned target's hostname *is itself* one of a signature's known malicious domains — direct identity, not an inference the way "this IP happens to also host something else bad" is. The source comment states this reasoning directly.

## Confidence: highest-wins, not averaged

`overallConfidence` ranks `low`/`medium`/`high`/`critical` numerically and returns whichever matched cluster carries the *highest* rank — not an average, not the first match found. One `critical`-confidence signature match outweighs five `low`-confidence ones. If nothing matched at all, confidence is the empty string, not a synthetic "none" level — the `Report.Confidence` doc comment says so directly: `"" when nothing matched — there is no verdict to be confident in`.

## Recommendation is a pure function of confidence

`recommend` is a plain switch: `critical`/`high` → avoid wallet interaction; `medium` → proceed with caution, verify independently; `low` → weak signal, remain cautious; empty → no known malicious infrastructure detected. It never looks at *which* evidence produced the confidence level, only the level itself — a design choice that keeps the recommendation text small and predictable rather than trying to generate a bespoke sentence per evidence combination.

## Evidence is sorted, not just collected

`sortItems` orders strong-first, then alphabetically by category and value within a tier — so the terminal/Markdown/JSON output always leads with the most convincing evidence, and, combined with `internal/correlation`'s own determinism, `Evaluate` produces byte-identical `Report`s for identical input (verified by a dedicated determinism test).

## Known limitation

**Coverage sits at 75.6%, just under the project's 80% floor**, and the reason is structural: `strengthOf`'s `default: return Weak` branch and `describe`'s `default: return string(c)` branch are both dead code paths in the current system — every `artifact.MatchCategory` value that `internal/correlation` can actually produce is explicitly handled in the switch above it. They exist as safety nets for a category added to the enum later without updating this switch, not as reachable behavior today. The "weak" evidence tier itself (community observations, domain age, keyword similarity) is defined in the type system (`Weak` strength constant exists) but nothing produces it — see [`intel.md`](intel.md)'s note on the same gap. This is tracked honestly in `roadmaps.MD`'s Known Gaps, not hidden.
