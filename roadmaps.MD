# Alexiares Roadmap & Working Tree

Derived from [Alexiares Technical Specification (specs.md).md](<Alexiares Technical Specification (specs.md).md>)

**Mission:** Open-source, defensive, local-first Web3 infrastructure intelligence CLI.
**Motto:** Guard the gate before the wallet connects.

---

## Phase 0 — Project Foundation

- [x] **Repo scaffolding**
  - [x] Initialize Go module (`alexiares`)
  - [x] Create directory layout: `cmd/`, `internal/` (collector, dns, tls, html, javascript, favicon, wallet, telegram, redirect, fingerprint, correlation, graph, evidence, output), `signatures/`, `observations/`, `configs/`, `docs/`, `tests/`
  - [x] Decide license (MIT vs Apache 2.0) — MIT
- [x] **CLI skeleton (Cobra)**
  - [x] Root command + version
  - [x] Stub subcommands: `scan`, `graph`, `dns`, `tls`, `html`, `js`, `wallet`, `report`, `update`
- [x] **Configuration**
  - [x] Load `~/.alexiares/config.yaml` (network timeout, user-agent, signatures path, output format)
  - [x] YAML config with sane defaults; JSON handled at the output layer (Phase 7), not the config file

**Definition of Done:** binary builds cross-platform, all subcommands respond with help text, config loads or falls back to defaults. ✅

---

## Phase 1 — Collector (Safe Acquisition)

- [x] **Input parsing**
  - [x] Accept URL, domain, IP (`internal/collector.Classify`)
  - [x] Accept wallet addresses: Ethereum, Bitcoin, Solana, Cardano, Tron, TON (`internal/wallet.Classify`)
  - [x] Batch file input + stdin (`internal/collector.ReadTargets`; CLI wiring lands with `scan` orchestration in Phase 7)
- [x] **Network collection**
  - [x] DNS resolution (implicit via Go's HTTP client dialer)
  - [x] HTTP(S) requests with redirect tracking
  - [x] TLS handshake capture
  - [x] HTML, favicon, and JavaScript download
  - [x] Header collection
- [x] **Safety guarantees**
  - [x] No JavaScript execution
  - [x] No wallet interaction, no form submission, no smart-contract calls
  - [x] Configurable timeouts + user-agent management
  - [x] `RawResponse` struct populated (URL, FinalURL, StatusCode, Headers, HTML, Scripts, Favicon, TLS, Redirects)

**Definition of Done:** `RawResponse` produced for any valid input; unsafe operations are structurally impossible; timeouts respected. ✅

---

## Phase 2 — Extractors

- [x] **DNS Extractor** — A, AAAA, CNAME, MX, NS, TXT, ASN, PTR → `DNSArtifacts` (ASN/ASNOrg intentionally empty; no bundled IP-to-ASN database, per the "no external database" performance goal)
- [x] **TLS Extractor** — SHA256 fingerprint, serial, issuer, subject, SAN, validity, key type
- [x] **HTML Extractor** — DOM structure, forms, input/hidden fields, metadata, comments, external resources
- [x] **JavaScript Extractor** — script URLs, inline scripts, SHA256 hashes, API endpoints, Telegram refs, Discord webhooks, wallet libraries, WalletConnect refs
- [x] **Favicon Extractor** — MurmurHash3 + SHA256 (verified against an independent Python mmh3 implementation; matches the Shodan favicon-hash convention for signature interoperability)
- [x] **Wallet Extractor** — detect ETH/BTC/SOL/ADA/TRON/TON addresses, ENS (ADA handles later)
- [x] **Telegram Extractor** — bot tokens, chat IDs, `api.telegram.org`, `t.me` links
- [x] **Redirect Extractor** — HTTP redirects (collector), meta refresh, JavaScript redirects

**Definition of Done:** each extractor has unit tests; extractors run independently and produce typed artifact structs. ✅

---

## Phase 3 — Fingerprint Engine

- [x] Favicon fingerprints (MurmurHash3, SHA256)
- [x] JavaScript fingerprints (SHA256; normalized AST hash deferred)
- [x] HTML structural hash (exact tag-shape SHA256) + DOM similarity fingerprint (64-bit SimHash over tag shingles, Hamming-distance comparable)
- [x] TLS certificate fingerprint
- [x] `Fingerprints` struct output (Favicon, JavaScript[], HTML, HTMLSimilarity, Certificate)

**Definition of Done:** identical artifacts always produce identical fingerprints (deterministic); covered by golden tests. ✅

---

## Phase 4 — Intelligence Model

- [x] **Observations layer** (untrusted community reports)
  - [x] YAML schema: type, domain, reported_at, reporter, evidence (screenshot, tx_hash, wallet)
  - [x] Never auto-promoted to signatures (structurally enforced — `internal/intel` has no Observation→Signature conversion path, documented in the package doc)
- [x] **Signatures layer** (maintainer-reviewed, trusted)
  - [x] YAML schema: id, description, favicon hashes, JS hashes, wallets, telegram patterns, confidence
  - [x] Repository structure: `favicon/`, `javascript/`, `certificates/`, `wallets/`, `telegram/`, `domains/`, `infrastructure/`, `rules/`
  - [x] Validation: YAML schema check, hash validation, duplicate detection, maintainer review flow (PR-based, per codestyle.md)
- [x] **Temporal intelligence**
  - [x] `Timeline` struct on every artifact (FirstSeen, ReportedAt, VerifiedAt, LastSeen)

**Definition of Done:** invalid signatures rejected at load; observation vs. signature separation enforced; timelines preserved end-to-end. ✅

---

## Phase 5 — Correlation Engine

- [x] Match target fingerprints against signatures: favicon, JavaScript, certificate, IP, nameserver, wallet, Telegram indicators, redirect infrastructure (ASN and registrar matching deferred — no extractor collects either yet; see Phase 2's ASN note and the missing WHOIS/registrar extractor)
- [x] Cluster identification (one cluster per matched signature, carrying its matches, description, and confidence)
- [x] `Correlation` output: Matches, Clusters, RelatedDomains, RelatedWallets

**Definition of Done:** deterministic matching — same input + same signature set always yields the same matches ✅ (verified by test); regression tests over historical infrastructure — none exist yet, since the project has no documented historical incidents to regress against; the `tests/regression/` tier is ready for the first real case.

---

## Phase 6 — Graph Engine

- [x] Node types: Domain, URL, IP, ASN, Certificate, Favicon, JavaScript, Wallet, Telegram, Registrar, Nameserver (ASN/Registrar/Telegram defined for schema completeness; not yet emitted by `FromScan` — no ASN/registrar data collected, and the spec defines no edge verb for Telegram)
- [x] Edge types: resolves_to, hosted_by, uses_certificate, shares_favicon, reuses_script, contains_wallet, redirects_to, registered_with, shares_nameserver (hosted_by/registered_with defined but unemitted, same reason)
- [x] Serializers: DOT, GraphML, JSON
- [x] `alexiares graph <target>` command wired (`--format dot|graphml|json`)

**Definition of Done:** exported graphs open cleanly in standard tooling — verified: DOT output rendered successfully with a real Graphviz install (`dot -Tsvg`); round-trip JSON serialization is lossless (verified by test). ✅

---

## Phase 7 — Evidence Engine & Output

- [x] **Evidence evaluation**
  - [x] Strong: shared JS, favicon, certificate, wallet, Telegram exfiltration
  - [~] Medium: shared IP ✅, redirect chain ✅ — ASN and registrar remain unimplemented (no data source, per Phase 2/5 notes)
  - [ ] Weak: community observation, domain age, keyword similarity — **not yet wired.** Observations are loadable and validated (Phase 4) but the scan pipeline doesn't consult them; domain age and keyword similarity have no implementation at all. Real gap, not a stub — tracked here for follow-up.
- [x] **Confidence model** — qualitative levels: Low / Medium / High / Critical (highest confidence among matched clusters wins)
- [x] **Explainable detections** — every evidence item cites its category, value, matched signature ID, and description; every report carries a recommendation
- [x] **Output formats**
  - [x] Terminal (default) — answers the four spec questions; verified end-to-end against a live local server matching the bundled example signature
  - [x] JSON (matches schema: target, confidence, evidence, artifacts, graph, recommendation) — verified against real output
  - [x] GraphML, DOT, CSV, Markdown
- [x] `alexiares report result.json` command — round-trips a saved JSON scan into any other format

**Definition of Done:** output answers the four spec questions ✅ (verified live); JSON output validates against schema ✅ (tested); golden tests lock output stability ✅ (`tests/golden/output_test.go`, all 5 non-wallet formats).

---

## Phase 8 — Update System & Hardening

- [x] `alexiares update` — pulls signatures, rules, schemas (verified end-to-end against a real signed archive over local HTTP)
- [x] Cryptographic signing + verification of updates (Ed25519; unsigned/wrong-key/tampered updates rejected and never touch disk, verified by test)
- [x] Security audit against non-goals — verified by grep across the whole tree, not just asserted:
  - No wallet interaction: `internal/wallet` is read-only regex classification; zero `PrivateKey`/`.Sign(` usage anywhere outside `internal/update`'s own signature *verification*
  - No JS execution: no JS-engine dependency in `go.mod` (only Cobra, yaml.v3, x/net); zero references to any JS runtime
  - No form submission: collector and update fetcher use `http.MethodGet` exclusively — grepped, no `POST` anywhere
  - No contract/chain interaction: no blockchain client library in `go.sum`; no `os/exec` anywhere (no shell-out surface at all)
- [x] **Performance targets** — measured against a real target (`example.com`), not estimated:
  - [x] Single URL scan < 2 seconds — measured 0.82–1.46s across 3 runs
  - [x] Memory < 100 MB — measured 18 MB peak RSS (`/usr/bin/time -l`)
  - [x] Binary < 15 MB — measured 10.1 MB unstripped, 7.0 MB with `-ldflags="-s -w"`
  - [x] No external database required — signatures load from flat YAML on disk; no DB driver dependency
- [~] **Test suite**
  - [x] Unit tests (every extractor, every internal package)
  - [x] Integration tests — bundled-signature detection and clean-target non-detection, both exercising the full collect→extract→fingerprint→correlate→evaluate pipeline (`tests/integration/scan_test.go`), plus repository validation
  - [ ] Regression tests (historical infrastructure) — **none yet.** The project has no documented historical incidents to regress against; fabricating one would be dishonest. `tests/regression/` is scaffolded and ready for the first real case.
  - [x] Golden tests (output stability) — fingerprint engine and all 5 non-wallet output formats

**Definition of Done (V1 release):** all performance targets measured and met ✅; unsigned updates rejected ✅ (verified by test); full test suite green ✅ (only the regression tier is empty, honestly, for lack of real incidents); `alexiares scan <url>` delivers evidence-backed, explainable results ✅ (verified live and by integration test).

---

## Known Gaps (honest, tracked — not silently dropped)

- **Weak evidence tier unwired.** Observations load and validate (Phase 4) but the scan pipeline never consults them; domain age and keyword similarity have no implementation. Only strong/medium evidence currently reaches a report.
- **No regression tests.** The project has no documented historical incidents yet, so `tests/regression/` is empty by honesty, not oversight.
- **ASN and registrar data uncollected.** No bundled IP-to-ASN database (by design — see the spec's "no external database" goal) and no WHOIS/RDAP extractor exists yet, so ASN/registrar signature fields and graph nodes are defined but never populated.
- **Telegram has no graph edge type.** The spec defines a Telegram node type but no corresponding edge verb, so Telegram indicators appear in evidence and correlation but not in `alexiares graph` output.
- **No canonical signature database yet.** The intended architecture (per the Signature Repository section of the spec) is that signatures live in a separate, vetted-contributor GitHub repository — this repo's `signatures/` is a bundled starter set, not the source of truth. That external repository doesn't exist yet, so `update.source_url` in `configs/default.yaml` has no real default to point at. `internal/update` is fully built and tested against the *mechanism* (fetch, Ed25519-verify, atomically install a signed tarball) — what's missing is the repository itself, its contribution/review workflow, and the CI job that builds and signs the tarball on merge.

## Future Phases (post-V1)

- **V1.5 — Canonical signature database** (concrete next step, not yet scheduled)
  - Stand up a separate GitHub repository (e.g. `alexiares-signatures`) as the source of truth for signatures, decoupled from this CLI's release cycle
  - Define its contribution workflow: PR → vetted-maintainer review → merge, mirroring the validation this repo already enforces (YAML schema, hash format, duplicate-ID detection)
  - CI on that repository builds `signatures.tar.gz`, signs it with the maintainers' Ed25519 key, and publishes both to a stable URL (e.g. GitHub Releases)
  - Point `configs/default.yaml`'s `update.source_url` and `update.public_key` at that published feed, so a fresh `alexiares` install has a real default instead of an empty one
- **V2**
  - Passive DNS integration
  - Certificate transparency monitoring
  - Browser extension
  - Wallet integration
- **V3**
  - Distributed signature network
  - Protocol monitoring
  - CI integrations
  - Community dashboards

---

## Non-Goals (guardrails)

- Not a SIEM, SOC platform, blockchain explorer, smart contract auditor, or AI threat-intel product
- Detection stays deterministic, local-first, and explainable — evidence over scores
