# Alexiares Code Style Guide (codestyle.md)

How code is written, formatted, and documented in this repository. These rules apply to every contribution — human or AI-generated. When a model (Claude, Copilot, etc.) writes code for this repo, it must follow this document exactly.

Companion docs: [the technical specification](<Alexiares Technical Specification (specs.md).md>) · [roadmaps.MD](roadmaps.MD)

---

## 1. Ground Rules

1. **Production-ready only.** No placeholder code, no `// TODO: implement later` in merged code, no stubbed returns that fake success. If a feature is partial, it lives on a branch.
2. **Deterministic.** Same input + same signatures = same output, byte for byte. No map-order-dependent output, no wall-clock values in results (timestamps come from artifact timelines, never `time.Now()` at render time), no randomness without an injected seed.
3. **Safe by construction.** The collector must make unsafe operations impossible at the type level — there is no code path that executes JavaScript, submits forms, connects wallets, or touches chain state.
4. **Explainable.** Every detection the code emits must carry its evidence. A function that returns a conclusion without the artifacts behind it is incomplete.
5. **Small, reviewable changes.** One logical change per commit. If a PR needs a novel to explain, split it.

---

## 2. Formatting & Tooling

Non-negotiable. CI rejects anything that fails these:

```bash
gofmt -l .            # must output nothing
goimports -l .        # must output nothing
go vet ./...          # must pass
golangci-lint run     # must pass with the repo's .golangci.yml
go test ./...         # must pass
```

- **Formatting is `gofmt`'s job, not yours.** Never hand-align code or argue style that a formatter decides.
- **Imports** grouped by `goimports`: stdlib, external, internal — separated by blank lines.
- **Line length:** no hard limit, but if a line needs horizontal scrolling to review, restructure it.
- **Enabled linters** (minimum): `errcheck`, `govet`, `staticcheck`, `unused`, `ineffassign`, `misspell`, `gocritic`, `revive`.

---

## 3. Project Layout

Follow the structure from the spec. New code goes where its responsibility lives:

```text
cmd/            Cobra commands only — flag parsing, wiring, output selection.
                No business logic. A command should read like a table of contents.
internal/       All engine logic. One package per module (collector, dns, tls,
                html, javascript, favicon, wallet, telegram, redirect,
                fingerprint, correlation, graph, evidence, output).
signatures/     YAML intelligence. Never Go code.
observations/   YAML community reports. Never Go code.
configs/        Default configs and schemas.
tests/          Integration, regression, and golden test fixtures.
```

Rules:

- `internal/` packages **never import `cmd/`**. Dependencies point inward.
- Extractors **never import each other**. They share types via a common `internal/artifact` package, not cross-imports.
- No `util`, `helpers`, `common`, or `misc` packages. Name packages after what they provide.

---

## 4. Naming

- Packages: short, lowercase, singular — `fingerprint`, not `fingerprints` or `fingerprintUtils`.
- Exported identifiers: `CamelCase`, and the package name is part of the name — `fingerprint.Compute`, not `fingerprint.ComputeFingerprint`.
- Interfaces: named for behavior (`Extractor`, `Serializer`), defined **where they are consumed**, not where they are implemented.
- Errors: sentinel errors as `ErrXxx` (`ErrInvalidSignature`); error types as `XxxError` (`ValidationError`).
- Acronyms keep their case: `DNSArtifacts`, `parseURL`, `TLSData` — never `DnsArtifacts` or `parseUrl`.
- Test helpers: `newTestCollector(t)`, taking `testing.TB` and calling `t.Helper()`.

---

## 5. Documentation

### 5.1 Godoc — required

Every exported package, type, function, method, and constant has a godoc comment. Full sentences, starting with the identifier's name:

```go
// Package fingerprint normalizes collected artifacts into comparable,
// deterministic identifiers (favicon hashes, script hashes, structural
// HTML hashes, and certificate fingerprints).
//
// All functions in this package are pure: identical input always yields
// an identical fingerprint, which the correlation engine and golden
// tests depend on.
package fingerprint

// Compute derives all fingerprint types from the raw response.
//
// It never performs network I/O; every input must already be collected.
// The returned Fingerprints is complete — fields for artifacts absent
// from the response are set to their zero value, never omitted.
func Compute(raw collector.RawResponse) (Fingerprints, error)
```

A doc comment answers: what it does, what it guarantees (determinism, no I/O, thread safety), and what the caller must know about edge cases. It does **not** restate the signature ("Compute computes...") and stop there.

### 5.2 Inline comments — sparse and load-bearing

Comment **why**, never **what**. The code shows what.

```go
// Bad — narrates the code:
// loop over the scripts and hash them
for _, s := range raw.Scripts { ... }

// Good — states a constraint the code can't show:
// Scripts are hashed pre-normalization: drainer kits are matched on
// exact bytes, and normalizing would break existing signature hashes.
for _, s := range raw.Scripts { ... }
```

If you feel the need to narrate a block, extract it into a well-named function instead.

### 5.3 Doc updates travel with code

A PR that changes behavior updates the affected godoc, `docs/`, and README examples **in the same PR**. Stale documentation is treated as a bug.

---

## 6. Error Handling

- **Never discard errors.** No `_ = err`, no empty `if err != nil {}` blocks. `errcheck` enforces this.
- **Wrap with context** at each layer boundary, lowercase, no punctuation, `%w`:

  ```go
  if err := yaml.Unmarshal(data, &sig); err != nil {
      return nil, fmt.Errorf("parsing signature %s: %w", path, err)
  }
  ```

- **Sentinel errors** for conditions callers branch on; check with `errors.Is` / `errors.As`, never string matching.
- **Panics are bugs.** Library code never panics on bad input — it returns an error. `panic` is reserved for provably impossible states, and the message must say so.
- **Partial failure is a result, not an error.** A scan where DNS succeeds but TLS times out returns artifacts *plus* a recorded failure per module — it does not abort. The evidence engine reports what was and wasn't collectable.
- **User-facing errors** (from `cmd/`) are actionable: what failed, why, what to try. Stack traces never reach the terminal by default.

---

## 7. Concurrency

- Collector fan-out uses `errgroup` with a bounded limit; goroutines are never spawned without an owner that waits for them.
- Every network operation takes a `context.Context` as the first parameter and honors cancellation and the configured timeout. No `context.Background()` below `cmd/`.
- Shared mutable state is a last resort. Prefer channels or per-goroutine results merged after `Wait()`.
- Anything concurrent gets a test run under `go test -race` in CI.

---

## 8. Testing

Every module ships with tests. The spec's four tiers, and where they live:

| Tier | Location | Convention |
|---|---|---|
| Unit | next to the code, `*_test.go` | table-driven, one table per behavior |
| Integration | `tests/integration/` | known phishing samples as fixtures, no live network |
| Regression | `tests/regression/` | one test per historical incident, named after it |
| Golden | `tests/golden/` | rendered output compared byte-for-byte, `-update` flag to regenerate |

Rules:

- **Table-driven by default**, with a `name` field per case:

  ```go
  tests := []struct {
      name string
      in   string
      want string
      err  error
  }{
      {name: "ethereum address lowercase", in: "0xabc...", want: "ethereum"},
      {name: "invalid checksum rejected", in: "0xABC...", err: wallet.ErrChecksum},
  }
  ```

- **No live network in tests.** Collectors are tested against `httptest.Server` and recorded fixtures. Tests must pass offline.
- **Test the contract, not the implementation.** Tests import the package as a consumer (`package wallet_test`) unless white-box access is genuinely needed.
- **Determinism tests are mandatory** for fingerprint and output code: run twice, require identical bytes.
- Coverage target: ≥ 80% per `internal/` package, and 100% of exported functions exercised. Coverage is a floor, not a goal — an untested edge case is a bug waiting for a CVE writeup.

---

## 9. Dependencies

- Stdlib first. A dependency must earn its place: actively maintained, permissive license, no transitive bloat that threatens the < 15 MB binary target.
- Pin via `go.mod`; `go.sum` committed; `go mod tidy` clean in CI.
- **Forbidden:** headless browsers, JS engines/interpreters, wallet SDKs that can sign or send, anything that phones home. These violate the safety model regardless of convenience.

---

## 10. Security-Sensitive Code

Extra rules for `collector/` and anything touching remote content:

- Treat **all** collected bytes as hostile. Size-limit every read (`io.LimitReader`), never `io.ReadAll` an unbounded body.
- Parse untrusted HTML/JS/YAML with hardened settings; fuzz tests (`go test -fuzz`) required for every parser entry point.
- Redirect following is capped and logged; each hop is recorded as an artifact.
- Signature update verification failures are **fatal** — never fall back to unsigned intelligence.
- No secrets in code, fixtures, or test data. Real bot tokens found in samples are redacted to the pattern level before committing.

---

## 11. Commits & Pull Requests

- Commit format: `type(scope): summary` — types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `sig` (signature updates).

  ```text
  feat(favicon): add MurmurHash3 fingerprinting
  fix(collector): bound redirect chain to 10 hops
  sig(telegram): add exfil patterns for drainer_cluster_007
  ```

- Subject ≤ 72 chars, imperative mood. Body explains *why* when the diff doesn't.
- A PR is mergeable when: CI green (fmt, vet, lint, race, tests), docs updated, no decrease in coverage, and — for detection logic — a fixture demonstrating the new detection.

---

## 12. Definition of "Production Ready"

Code merges to `main` only when all of these hold:

- [ ] `gofmt`, `goimports`, `go vet`, `golangci-lint` clean
- [ ] All exported identifiers documented (godoc)
- [ ] Errors wrapped with context; no discarded errors
- [ ] Context-aware and race-clean (`go test -race`)
- [ ] Tests at the appropriate tier(s), passing offline
- [ ] Deterministic output verified where applicable
- [ ] No forbidden dependencies or unsafe code paths
- [ ] Docs and examples updated in the same PR
- [ ] Performance budget respected (scan < 2 s, memory < 100 MB, binary < 15 MB)
