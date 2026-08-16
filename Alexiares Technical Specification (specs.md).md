# Alexiares Technical Specification (specs.md)

Version: v0.1 (CLI-first architecture)

Status: Draft

License: MIT / Apache 2.0 (TBD)

---

# Mission

Alexiares is an open-source infrastructure intelligence CLI designed to protect Web3 communities from phishing, wallet drainers, fake governance portals, malicious frontends, and infrastructure-based attacks.

The project is defensive by design.

Alexiares provides infrastructure analysis, artifact correlation, explainable detections, and community-maintained intelligence without relying on AI.

Core principles:

- Local-first execution
- Explainable evidence
- Deterministic detection
- Community-driven intelligence
- Reproducible results
- Modular architecture

---

# Primary Objective

Given a URL, domain, IP, or wallet address, Alexiares should determine whether the target is connected to known malicious infrastructure and provide evidence supporting that assessment.

The tool must prioritize evidence over scores.

Output should answer:

- What was found?
- Why is it suspicious?
- What infrastructure is related?
- How confident is the conclusion?

---

# System Architecture

```text
CLI Input
    |
    |
Collector
    |
    |
Extractors
    |
    |
Fingerprint Engine
    |
    |
Correlation Engine
   / \\
  /   \\
 /     \\
Signatures  Observations
      \\    /
       \\  /
    Evidence Engine
      /      \\
     /        \\
Graph Engine  Output
```

---

# Technology Stack

Language: Go

Reasons:

- single binary distribution
- fast networking
- concurrency
- cross-platform
- easy installation
- minimal runtime dependencies

CLI framework:

- Cobra

Configuration:

- YAML
- JSON

Graph serialization:

- GraphML
- DOT
- JSON

---

# Repository Structure

```text
alexiares/

  cmd/

  internal/

    collector/

    dns/

    tls/

    html/

    javascript/

    favicon/

    wallet/

    telegram/

    redirect/

    fingerprint/

    correlation/

    graph/

    evidence/

    output/

  signatures/

  observations/

  configs/

  docs/

  tests/
```

---

# CLI Commands

## Scan

```bash
alexiares scan https://example.xyz
```

Performs complete analysis.

## Graph

```bash
alexiares graph example.xyz
```

Builds infrastructure relationship graph.

## DNS

```bash
alexiares dns example.xyz
```

Displays DNS intelligence.

## TLS

```bash
alexiares tls example.xyz
```

Displays certificate intelligence.

## HTML

```bash
alexiares html https://example.xyz
```

Extracts HTML artifacts.

## JavaScript

```bash
alexiares js https://example.xyz
```

Extracts JavaScript artifacts.

## Wallet

```bash
alexiares wallet 0xABC...
```

Analyzes wallet intelligence.

## Report

```bash
alexiares report result.json
```

Generates formatted report.

---

# Input Types

Supported inputs:

- URL
- Domain
- IP address
- Ethereum address
- Bitcoin address
- Solana address
- Cardano address
- Tron address
- TON address
- Batch file
- stdin

---

# Collector Module

Purpose:

Safely acquire infrastructure data.

Responsibilities:

- DNS resolution
- HTTP(S) requests
- redirect tracking
- TLS handshake
- HTML download
- favicon download
- JavaScript download
- header collection
- timeout handling
- user-agent management

Safety requirements:

- disable JavaScript execution
- no wallet interaction
- no browser automation in v1
- configurable timeouts
- optional sandboxing

Output:

RawResponse

```go
type RawResponse struct {
    URL string
    FinalURL string
    StatusCode int
    Headers map[string][]string
    HTML string
    Scripts []string
    Favicon []byte
    TLS TLSData
    Redirects []Redirect
}
```

---

# Extractor Modules

## DNS Extractor

Collect:

- A
- AAAA
- CNAME
- MX
- NS
- TXT
- ASN
- PTR

Output:

```go
type DNSArtifacts struct {
    IPs []string
    Nameservers []string
    MX []string
    ASN string
}
```

---

## TLS Extractor

Collect:

- SHA256 fingerprint
- serial number
- issuer
- subject
- SAN
- validity period
- key type

---

## HTML Extractor

Collect:

- DOM structure
- forms
- input fields
- hidden fields
- metadata
- comments
- external resources

---

## JavaScript Extractor

Collect:

- script URLs
- inline scripts
- SHA256 hashes
- API endpoints
- Telegram references
- Discord webhooks
- wallet libraries
- WalletConnect references

---

## Favicon Extractor

Compute:

- MurmurHash3
- SHA256

---

## Wallet Extractor

Detect:

Ethereum

Bitcoin

Solana

Cardano

Tron

TON

Extract:

- addresses
- ENS
- ADA handles (future)

---

## Telegram Extractor

Detect:

- bot tokens
- chat IDs
- api.telegram.org
- t.me links

---

## Redirect Extractor

Capture:

- HTTP redirects
- meta refresh
- JavaScript redirects

---

# Fingerprint Engine

Purpose:

Normalize artifacts into comparable identifiers.

Fingerprint types:

## Favicon

- MurmurHash3
- SHA256

## JavaScript

- SHA256
- normalized AST hash (future)

## HTML

- structural hash
- DOM similarity fingerprint

## TLS

- certificate fingerprint

Output:

```go
type Fingerprints struct {
    Favicon string
    JavaScript []string
    HTML string
    Certificate string
}
```

---

# Intelligence Model

Alexiares separates intelligence into two layers.

## Observations

Raw community reports.

Not trusted by default.

Purpose:

- historical context
- analyst collaboration
- evidence collection

Example:

```yaml
type: observation

domain: example.xyz

reported_at: ...

reporter: ...

evidence:

  screenshot

  tx_hash

  wallet
```

---

## Signatures

Maintainer-reviewed detection artifacts.

Trusted intelligence.

Purpose:

- infrastructure detection
- cluster identification
- deterministic matching

Example:

```yaml
id: wallet_drainer_cluster_001

description: Fake wallet connection infrastructure

favicon:

  murmur3:

    - -204998123

javascript:

  sha256:

    - e3b0...

wallets:

  ethereum:

    - 0x123...

telegram:

  patterns:

    - api.telegram.org/bot

confidence: high
```

Observations may contribute to future signatures.

Observations should never automatically become signatures.

---

# Temporal Intelligence

Every artifact must preserve timeline information.

```go
type Timeline struct {
    FirstSeen time.Time
    ReportedAt time.Time
    VerifiedAt time.Time
    LastSeen time.Time
}
```

This prevents confusion between observation time and activity time.

---

# Correlation Engine

Purpose:

Identify relationships between target artifacts and known intelligence.

Correlation targets:

- favicon
- JavaScript
- certificate
- IP
- ASN
- registrar
- nameserver
- wallet
- Telegram indicators
- redirect infrastructure

Output:

```go
type Correlation struct {
    Matches []Match
    Clusters []Cluster
    RelatedDomains []string
    RelatedWallets []string
}
```

---

# Graph Engine

Purpose:

Represent infrastructure relationships.

Node Types

- Domain
- URL
- IP
- ASN
- Certificate
- Favicon
- JavaScript
- Wallet
- Telegram
- Registrar
- Nameserver

Edge Types

- resolves_to
- hosted_by
- uses_certificate
- shares_favicon
- reuses_script
- contains_wallet
- redirects_to
- registered_with
- shares_nameserver

Graph should support:

- DOT
- GraphML
- JSON

---

# Evidence Engine

The Evidence Engine produces human-readable conclusions.

It does not perform intelligence collection.

Responsibilities:

- evaluate evidence
- assign confidence
- generate recommendations
- explain detections

Evidence Categories

Strong

- shared JavaScript
- shared favicon
- shared certificate
- shared wallet
- Telegram exfiltration

Medium

- shared IP
- shared ASN
- shared registrar
- redirect chain

Weak

- community observation
- domain age
- keyword similarity

---

# Confidence Model

Use qualitative confidence.

Levels:

- Low
- Medium
- High
- Critical

Example:

```text
Detection

Shared favicon hash

Evidence

MurmurHash3: -204998123

Matched Signature

wallet_drainer_cluster_001

Confidence

High

Recommendation

Avoid wallet interaction
```

---

# Output Formats

Terminal

Default.

JSON

Machine-readable.

GraphML

Graph analysis.

DOT

Visualization.

CSV

Bulk processing.

Markdown

Documentation.

---

# JSON Output Schema

```json
{
  "target": "https://example.xyz",
  "confidence": "high",
  "evidence": [],
  "artifacts": {},
  "graph": {},
  "recommendation": "avoid"
}
```

---

# Configuration

Configuration file:

~/.alexiares/config.yaml

Example:

```yaml
network:

  timeout: 10

  user_agent: Alexiares/0.1

signatures:

  path: ~/.alexiares/signatures

output:

  format: terminal
```

---

# Signature Repository

Structure:

```text
signatures/

  favicon/

  javascript/

  certificates/

  wallets/

  telegram/

  domains/

  infrastructure/

  rules/
```

Validation:

- YAML schema
- hash validation
- duplicate detection
- maintainer review

Distribution:

The canonical signature intelligence does not live inside the Alexiares CLI repository. It lives in a separate, dedicated GitHub repository maintained by vetted contributors, decoupled from the CLI's own release cycle — the same relationship Homebrew has to homebrew-core, or a YARA/Semgrep engine has to its rule registry. Access to that repository is access-controlled: contributors submit signatures via PR, a maintainer reviews and merges, and CI builds, signs, and publishes the resulting archive.

The CLI repository's own `signatures/` directory is a bundled starter/example set only, not the source of truth. `alexiares update` (see Update System, below) is how a local installation stays current against the canonical repository.

---

# Update System

```bash
alexiares update
```

Fetches the current signature archive and its detached signature from the source configured in `update.source_url` (`~/.alexiares/config.yaml`), verifies the signature against `update.public_key`, and — only on success — replaces the local signature directory (`signatures.path`) with the verified contents.

Updates:

- signatures
- rules
- schemas

Updates must be cryptographically signed; an update that fails verification is rejected outright and never touches disk. There is no hardcoded default `update.source_url` — it points at whichever GitHub-hosted signature database repository the operator (or a future Alexiares-maintained default) chooses to trust.

---

# Security Requirements

The tool must never:

- connect wallets
- execute malicious JavaScript
- submit forms
- interact with smart contracts
- modify blockchain state
- connect to a loopback, private, or link-local address by default (SSRF protection — `internal/collector.safeDialContext` validates every connection, including redirect hops, before dialing it)

Analysis only.

---

# Testing Strategy

Unit tests

Each extractor.

Integration tests

Known phishing samples.

Regression tests

Historical infrastructure.

Golden tests

Output stability.

---

# Performance Goals

Single URL scan:

< 2 seconds

Memory:

< 100 MB

Binary size:

< 15 MB

No external database required.

---

# Future Roadmap

V1

CLI

DNS

TLS

HTML

JavaScript

Favicon

Wallet extraction

Graph output

Signatures

V2

Passive DNS integration

Certificate transparency

Browser extension

Wallet integration

V3

Distributed signature network

Protocol monitoring

CI integrations

Community dashboards

---

# Non-Goals

Alexiares is not:

- a SIEM
- a SOC platform
- a blockchain explorer
- a smart contract auditor
- an AI threat intelligence product

It is a focused infrastructure intelligence tool for Web3 operational security.

---

# Project Motto

**Guard the gate before the wallet connects.**