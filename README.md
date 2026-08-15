# Alexiares

> **Guard the gate before the wallet connects.**

Alexiares is an open-source infrastructure intelligence CLI that protects Web3 communities from phishing sites, wallet drainers, fake governance portals, malicious frontends, and infrastructure-based attacks.

It is **defensive by design**: given a URL, domain, IP, or wallet address, Alexiares determines whether the target is connected to known malicious infrastructure — and shows you the evidence.

- **Local-first** — runs entirely on your machine, no external database
- **Explainable** — every detection cites the artifacts and signatures behind it
- **Deterministic** — same input + same signatures = same result, every time
- **Community-driven** — intelligence maintained as reviewable YAML signatures
- **No AI** — evidence over scores

---

## Why Alexiares?

A phishing kit can clone a frontend in minutes, but it almost always reuses infrastructure: the same favicon, the same drainer script, the same TLS certificate, the same exfiltration bot, the same wallets. Alexiares fingerprints those artifacts and correlates them against community-maintained intelligence, so one reported drainer exposes the entire cluster behind it.

Output always answers four questions:

1. **What was found?**
2. **Why is it suspicious?**
3. **What infrastructure is related?**
4. **How confident is the conclusion?**

---

## Installation

Alexiares ships as a single static binary (< 15 MB) for macOS, Linux, and Windows.

```bash
go install github.com/alexiares/alexiares@latest
```

Or download a release binary:

```bash
curl -sSL https://github.com/alexiares/alexiares/releases/latest/download/alexiares-$(uname -s)-$(uname -m) -o /usr/local/bin/alexiares && chmod +x /usr/local/bin/alexiares
```

Verify:

```bash
alexiares version
```

---

## Quick Start

Scan a suspicious link before anyone in your community clicks it:

```bash
alexiares scan https://claim-airdrop-example.xyz
```

```text
Alexiares v1.0 — infrastructure intelligence

Target        https://claim-airdrop-example.xyz
Resolved      203.0.113.42 (AS64512, BulletProof-Host)
Certificate   sha256:9f2a…c41e (issued 2 days ago)

DETECTION     Wallet drainer infrastructure

Evidence
  [STRONG]  Favicon MurmurHash3 -204998123
            → matches signature wallet_drainer_cluster_001
  [STRONG]  Script drainer.min.js sha256:e3b0…
            → reused across 14 known domains
  [STRONG]  Telegram exfiltration endpoint api.telegram.org/bot…
  [MEDIUM]  Hosted on IP shared with 6 flagged domains
  [WEAK]    Domain registered 3 days ago

Related infrastructure
  claim-rewards-example.xyz, airdrop-example.app, +12 more
  Wallet 0x123… (drainer receiver, ethereum)

Confidence    HIGH

Recommendation
  Avoid wallet interaction. Do not connect, sign, or approve.
```

---

## Commands

| Command | Purpose |
|---|---|
| `alexiares scan <url>` | Full analysis: collect, extract, fingerprint, correlate, report |
| `alexiares graph <domain>` | Build the infrastructure relationship graph |
| `alexiares dns <domain>` | DNS intelligence (A/AAAA, CNAME, MX, NS, TXT, ASN, PTR) |
| `alexiares tls <domain>` | Certificate intelligence (fingerprint, issuer, SAN, validity) |
| `alexiares html <url>` | HTML artifacts (forms, hidden fields, metadata, comments) |
| `alexiares js <url>` | JavaScript artifacts (hashes, endpoints, webhooks, wallet libs) |
| `alexiares wallet <address>` | Wallet intelligence across chains |
| `alexiares report result.json` | Render a formatted report from saved results |
| `alexiares update` | Pull cryptographically signed signature/rule updates |

### Supported inputs

URLs, domains, IP addresses, and wallet addresses on **Ethereum, Bitcoin, Solana, Cardano, Tron, and TON** — plus batch files and stdin for bulk processing:

```bash
cat suspects.txt | alexiares scan --format csv > results.csv
```

### Output formats

`terminal` (default) · `json` · `graphml` · `dot` · `csv` · `markdown`

```bash
alexiares graph example.xyz --format dot | dot -Tpng -o cluster.png
```

---

## How It Works

```text
CLI Input → Collector → Extractors → Fingerprint Engine → Correlation Engine
                                                          (signatures + observations)
                                                                   ↓
                                                            Evidence Engine
                                                              ↓        ↓
                                                        Graph Engine  Output
```

1. **Collector** safely acquires the target: DNS, HTTP(S) with redirect tracking, TLS handshake, HTML, favicon, and scripts. JavaScript is **never executed**.
2. **Extractors** pull artifacts: DNS records, certificate details, DOM structure, script hashes, favicon hashes, wallet addresses, Telegram indicators, redirect chains.
3. **Fingerprint Engine** normalizes artifacts into comparable identifiers (MurmurHash3, SHA256, structural hashes).
4. **Correlation Engine** matches fingerprints against signatures and clusters related infrastructure.
5. **Evidence Engine** weighs the matches (strong / medium / weak), assigns qualitative confidence (Low / Medium / High / Critical), and explains the conclusion.
6. **Graph Engine** exports the relationship graph (domains, IPs, certs, scripts, wallets…) as DOT, GraphML, or JSON.

---

## Intelligence Model

Two strictly separated layers:

- **Signatures** — maintainer-reviewed, trusted detection artifacts. Deterministic matching against favicon hashes, script hashes, certificates, wallets, and Telegram patterns.
- **Observations** — raw community reports. Never trusted by default, never auto-promoted; they provide historical context and feed the maintainer review pipeline.

```yaml
id: wallet_drainer_cluster_001
description: Fake wallet connection infrastructure
favicon:
  murmur3: [-204998123]
javascript:
  sha256: [e3b0...]
wallets:
  ethereum: [0x123...]
telegram:
  patterns: [api.telegram.org/bot]
confidence: high
```

Every artifact carries a timeline (`FirstSeen`, `ReportedAt`, `VerifiedAt`, `LastSeen`) so observation time is never confused with activity time. Signature updates are cryptographically signed.

The signature shown above lives in this repo as a bundled example. In the intended end state, the canonical, continuously updated set lives in a separate GitHub repository maintained by vetted contributors — `alexiares update` pulls from it, the same way `brew` pulls formulae from homebrew-core rather than bundling them.

---

## Configuration

`~/.alexiares/config.yaml`:

```yaml
network:
  timeout: 10
  user_agent: Alexiares/1.0
signatures:
  path: ~/.alexiares/signatures
output:
  format: terminal
```

---

## Safety Guarantees

Alexiares is analysis-only. It will **never**:

- connect wallets
- execute JavaScript from targets
- submit forms
- interact with smart contracts
- modify blockchain state

## Performance

- Single URL scan in **< 2 seconds**
- **< 100 MB** memory
- **< 15 MB** single binary, no external database

---

## What Alexiares Is Not

Not a SIEM, not a SOC platform, not a blockchain explorer, not a smart contract auditor, not an AI threat-intel product. It is a focused infrastructure intelligence tool for Web3 operational security.

---

## Contributing

- **Report malicious infrastructure** by submitting an observation (domain, evidence, screenshots, tx hashes) to `observations/`.
- **Propose signatures** — the canonical signature database is a separate, vetted-contributor GitHub repository, not this one. This repo's `signatures/` directory is a bundled starter set, not the source of truth; `alexiares update` is how an installation pulls the current, maintainer-reviewed set from the canonical repository. (The canonical repository isn't published yet — signature contribution guidance will point here once it is.) All signatures pass YAML schema validation, hash validation, duplicate detection, and maintainer review regardless of which repository they land in.
- **Code contributions** welcome: extractors, output formats, and test samples (unit, integration, regression, and golden tests).

See [roadmaps.MD](roadmaps.MD) for the build plan, [the technical specification](<Alexiares Technical Specification (specs.md).md>) for the system-level architecture, and [docs/](docs/) for a per-module reference — what each package does and why, one file per `internal/` package.

## License

[MIT](LICENSE)
