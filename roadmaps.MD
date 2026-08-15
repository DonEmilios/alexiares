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

- [ ] **Evidence evaluation**
  - [ ] Strong: shared JS, favicon, certificate, wallet, Telegram exfiltration
  - [ ] Medium: shared IP, ASN, registrar, redirect chain
  - [ ] Weak: community observation, domain age, keyword similarity
- [ ] **Confidence model** — qualitative levels: Low / Medium / High / Critical
- [ ] **Explainable detections** — every conclusion cites the artifact + matched signature + recommendation
- [ ] **Output formats**
  - [ ] Terminal (default)
  - [ ] JSON (matches schema: target, confidence, evidence, artifacts, graph, recommendation)
  - [ ] GraphML, DOT, CSV, Markdown
- [ ] `alexiares report result.json` command

**Definition of Done:** output answers the four spec questions (what was found / why suspicious / related infrastructure / confidence); JSON output validates against schema; golden tests lock output stability.

---

## Phase 8 — Update System & Hardening

- [ ] `alexiares update` — pulls signatures, rules, schemas
- [ ] Cryptographic signing + verification of updates
- [ ] Security audit against non-goals: never connects wallets, executes JS, submits forms, touches contracts or chain state
- [ ] **Performance targets**
  - [ ] Single URL scan < 2 seconds
  - [ ] Memory < 100 MB
  - [ ] Binary < 15 MB
  - [ ] No external database required
- [ ] **Test suite complete**
  - [ ] Unit tests (every extractor)
  - [ ] Integration tests (known phishing samples)
  - [ ] Regression tests (historical infrastructure)
  - [ ] Golden tests (output stability)

**Definition of Done (V1 release):** all performance targets measured and met; unsigned updates rejected; full test suite green; `alexiares scan <url>` delivers evidence-backed, explainable results.

---

## Future Phases (post-V1)

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
