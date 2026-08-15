# `internal/intel`

**Source:** [`internal/intel/`](../internal/intel/) (`signature.go`, `observation.go`, `repository.go`)
**Tests:** `signature_test.go`, `observation_test.go`, `repository_test.go` — 85.4% coverage
**Position in pipeline:** loaded once per scan (`intel.LoadSignatures`), feeds `internal/correlation`

## Purpose

Defines Alexiares' two-layer intelligence model and the loader that reads it off disk:

- **`Signature`** — a maintainer-reviewed, trusted detection artifact: favicon hashes, script hashes, certificate fingerprints, wallet addresses (by chain), Telegram patterns, domains, IPs, nameservers, and a qualitative `Confidence`.
- **`Observation`** — a raw, untrusted community report: a domain, who reported it, when, and free-text evidence (screenshot reference, tx hash, wallet).

```go
func LoadSignatures(dir string) ([]Signature, error)
func LoadObservations(dir string) ([]Observation, error)
func (s Signature) Validate() error
func (o Observation) Validate() error
```

## The separation is structural, not a convention

This is the one thing worth reading the actual package doc comment for, so it's quoted rather than paraphrased:

> The separation is structural, not a policy note: this package contains no function that converts an Observation into a Signature. Promoting an observation to a signature is a human, maintainer-review decision — writing a new signature file by hand — never an automated one. Anyone extending this package must preserve that: do not add an Observation-to-Signature conversion here.

In other words: it's not that observations *shouldn't* become signatures automatically, it's that there is no code path capable of it. Enforcing this by absence rather than by a runtime check means there's nothing to accidentally bypass.

## Design notes

**A signature with no detection criteria is a schema violation, not a valid-but-useless entry.** `hasCriteria` checks every matchable field (favicon, JS, certs, Telegram patterns, domains, IPs, nameservers, wallets) and `Validate` rejects the signature outright if all of them are empty. A signature that can never match anything is treated as a maintainer error to catch at load time, not something that silently sits in the repository doing nothing.

**Signature IDs are validated against a slug pattern** (`^[a-z0-9][a-z0-9_-]*[a-z0-9]$`) — lowercase, alphanumeric, `-`/`_` separators only. This isn't just style enforcement: IDs are used as map keys, YAML anchors elsewhere, and eventually as filenames in the canonical signature database's contribution workflow, all of which have their own reasons to reject arbitrary Unicode or whitespace.

**`LoadSignatures`/`LoadObservations` aggregate every error found, not just the first.** A directory of 50 signature files with 3 bad ones reports all 3 problems in one pass (via `errors.Join`), not "first bad file, fix it, rerun, find the second." This matters specifically for a PR-review workflow against a growing signature database — a reviewer wants the whole list of what's wrong, not one-at-a-time whack-a-mole.

**Duplicate signature IDs across *different files* are a load-time error**, not silently last-write-wins. `LoadSignatures` tracks which file first defined each ID and reports the specific collision (`"b.yaml: duplicate signature id \"x\" (first defined in a.yaml)"`) rather than one signature silently overwriting another.

**A missing directory is empty, not an error.** Both loaders treat a nonexistent `signatures/` or `observations/` directory as "nothing here yet," not a failure — a fresh Alexiares install with no signatures configured (see [`update.md`](update.md)) still runs; it just has nothing to correlate against.

## Known limitation

Observations are loaded and validated but the scan pipeline never actually *consults* them — the "weak evidence" tier (community observation, domain age, keyword similarity) described in the spec has no implementation wiring it into `internal/evidence` yet. Tracked honestly in `roadmaps.MD`'s Known Gaps, not silently dropped.
