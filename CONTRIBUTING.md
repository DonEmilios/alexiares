# Contributing to Alexiares

Alexiares' intelligence is only as trustworthy as the process that produces it. Every signature shipped by `alexiares update` runs, unattended, against real user traffic and decides whether to tell someone "this is safe to connect your wallet to." A bad signature is not a cosmetic bug — it is a false sense of security or a false accusation. This document is the strict, load-bearing process that keeps that from happening.

It applies to this repository's bundled `signatures/`/`observations/` today, and is written to carry over directly to the canonical `alexiares-signatures` repository once it exists (see [roadmaps.MD](roadmaps.MD)'s V1.5).

---

## 1. The model this process protects

Two layers, kept structurally separate in code (`internal/intel`'s package doc is explicit about this — there is no function anywhere that turns an Observation into a Signature):

| | **Observation** | **Signature** |
|---|---|---|
| Who can submit | Anyone, no review gate | Only via the pipeline below |
| Trust level | None — raw, unverified | Maintainer-reviewed, used for live detection |
| Effect | Historical record, triage input | Directly drives `alexiares scan` verdicts |
| Promotion | Never automatic | A human rewrites it by hand, from scratch |

Everything below exists to enforce that an Observation can *inform* a Signature but can never *become* one without a human re-deriving the evidence independently.

---

## 2. Roles

Nobody needs a title to submit an Observation. Everything past that requires an explicit role, because each role is a larger blast radius if it goes wrong.

### Contributor (no gate)
Anyone reporting suspicious infrastructure. No access required — a PR against `observations/` is enough.

**Needs to know:** how to fill out the Observation YAML (§4) truthfully and how to attach real evidence (§4.2). Nothing else.

### Observation Triager
Reads incoming Observations and decides which ones are worth escalating into a Signature proposal. Does **not** have merge rights on `signatures/`.

**Requirements:**
- Can independently reproduce a reported finding (visit an archived/sandboxed copy of the site, pull the favicon, pull the script) without connecting a wallet or executing untrusted JavaScript.
- Understands the difference between "this is suspicious" and "this is a reusable, specific fingerprint" — most observations never clear this bar.
- Comfortable rejecting or closing low-quality/spam reports without a maintainer's involvement.

### Signature Maintainer
Writes and reviews Signature YAML. Has merge rights on `signatures/`. This is the role with the most day-to-day trust.

**Requirements:**
- Can read Go well enough to know exactly what `internal/intel.Signature.Validate()` (and `hasCriteria()`) does and does not check — the schema validator is the floor, not the review.
- Can independently recompute every hash type in the schema from a live or archived artifact: SHA256 (`sha256sum`), the project's MurmurHash3-x86-32 favicon convention (base64-encoded icon bytes, matching Shodan's `http.favicon.hash`), and knows how to pull a certificate's SHA256 fingerprint (`openssl x509 -fingerprint -sha256`).
- Understands evidence strength and false-positive blast radius: a shared hosting IP or a stock/default favicon is weak and dangerous to signature; a bespoke drainer script hash is strong. (See §5.3.)
- Has enough Web3/phishing-infrastructure background to recognize wallet-drainer patterns, fake governance/airdrop kits, and common evasion tricks (URL cloaking, redirect chains, cloned DOM with a swapped contract address) — not just "can this YAML parse."
- Never merges their own signature proposal without a second Signature Maintainer's sign-off (two-person rule — see §6).

### Release Signer
Holds (or controls access to) the Ed25519 private key that signs `signatures.tar.gz` before publication. Smallest group, highest trust — a compromised key lets an attacker push malicious "verified" signatures to every installation running `alexiares update`. This is a supply-chain role, not a content-review role.

**Requirements:**
- Understands that this key is the single point of failure for every Alexiares user's trust in `alexiares update`, and treats it accordingly: generated and stored offline or in an HSM/managed KMS, never committed, never pasted into CI logs or chat, never reused for anything else.
- Has a documented, rehearsed key-rotation plan (new keypair, new `update.public_key` default, communicated revocation of the old key) in case of suspected compromise.
- Does **not** also need deep signature-content expertise — their job is verifying that what's about to be signed is exactly what Signature Maintainers merged (diff the tag against `main`, confirm CI's own validation passed), not re-reviewing the content itself.
- Is never the same person as the sole reviewer of the signature they're about to sign, for the same reason financial controls separate "who approves the payment" from "who sends it."

---

## 3. The pipeline — from report to running detection

```
 Contributor          Observation Triager        Signature Maintainer (x2)         Release Signer            Every install
     │                        │                           │                             │                         │
     │  1. submit             │                           │                             │                         │
     ├──────────────────────► │                           │                             │                         │
     │   observations/*.yaml  │  2. triage                │                             │                         │
     │                        ├─────────┐                 │                             │                         │
     │                        │  reproduce independently   │                             │                         │
     │                        │  reject spam/duplicates    │                             │                         │
     │                        │◄────────┘                 │                             │                         │
     │                        │  3. escalate (if warranted)                              │                         │
     │                        ├──────────────────────────►│                             │                         │
     │                        │                           │  4. author signature YAML   │                         │
     │                        │                           │     from scratch, by hand   │                         │
     │                        │                           ├────────┐                    │                         │
     │                        │                           │  re-derive every hash        │                         │
     │                        │                           │  independently (§5.4)        │                         │
     │                        │                           │◄───────┘                    │                         │
     │                        │                           │  5. PR + second Maintainer  │                         │
     │                        │                           │     review (two-person rule)│                         │
     │                        │                           │  6. CI: LoadSignatures       │                         │
     │                        │                           │     schema+hash+dup check    │                         │
     │                        │                           │  7. merge                    │                         │
     │                        │                           ├─────────────────────────────►│                         │
     │                        │                           │                             │  8. build tarball,     │
     │                        │                           │                             │     sign with Ed25519   │
     │                        │                           │                             ├────────────────────────►│
     │                        │                           │                             │  9. publish             │  10. alexiares update
     │                        │                           │                             │     tar.gz + .sig       │      verifies sig,
     │                        │                           │                             │                         │      installs atomically
```

Each numbered step is a hard gate, not a suggestion:

1. **Submit** — a Contributor opens a PR adding one file to `observations/`, following §4.
2. **Triage** — an Observation Triager reproduces the finding independently (never trusting the submitter's word alone) and either closes it (spam, duplicate, unreproducible, too generic) or escalates it.
3. **Escalate** — the Triager hands the reproduced evidence to a Signature Maintainer. The Observation itself is never copy-pasted into a Signature file; it is a pointer, not a source.
4. **Author** — a Signature Maintainer writes a brand-new Signature YAML by hand (§5), re-deriving every hash from the artifact themselves.
5. **Second review** — a *different* Signature Maintainer independently re-derives at least one hash and checks §5.3's false-positive criteria before approving. One maintainer never merges their own signature.
6. **CI validation** — `LoadSignatures` runs: YAML schema, hash-format validation (§5.2), duplicate-ID detection across the whole tree. A red CI run blocks merge unconditionally.
7. **Merge** — into `signatures/` (this repo today; `alexiares-signatures` once V1.5 stands it up).
8. **Build & sign** — CI (in the canonical repo) packages `signatures.tar.gz` and the Release Signer's process signs it with the trusted Ed25519 key. Nothing is signed off a Signature Maintainer's merge alone — signing is a distinct, separately-triggered action.
9. **Publish** — the archive and detached `.sig` go to a stable URL (GitHub Releases).
10. **Distribute** — `alexiares update` fetches both files, calls `update.Verify()` against the pinned `update.public_key`, and only on success atomically replaces the local signature directory. A failed or missing signature leaves the previous, still-verified signatures untouched — there is no fallback to trusting unsigned content.

---

## 4. Observation YAML — exact structure

One file per report, saved under `observations/`. Filename convention: `<domain-with-dashes>-<yyyy-mm-dd>.yaml` (e.g. `claim-airdrop-example-xyz-2026-08-16.yaml`).

```yaml
# observations/claim-airdrop-example-xyz-2026-08-16.yaml
type: observation                        # required, literal string "observation"
domain: claim-airdrop-example.xyz        # required — the domain you're reporting
reported_at: 2026-08-16T14:00:00Z        # required — RFC3339 UTC timestamp
reporter: your-github-handle             # required — accountability, not anonymous
evidence:                                # optional block; include whatever you have
  screenshot: https://...                # optional — link to an archived screenshot
  tx_hash: "0xabc123..."                 # optional — on-chain evidence of drainer activity
  wallet: "0x123456789abcdef..."         # optional — receiving/drainer wallet address
```

**Field rules** (enforced by `intel.Observation.Validate()`):
- `domain`, `reporter`, and `reported_at` are required — a report with no accountable source or timestamp is rejected outright.
- `evidence` fields are free text on purpose: attach whatever you actually have. A maintainer decides later what it's worth — do not fabricate or round up fields to look more complete.

**Rules for what an Observation must never contain:**
- No working exploit code, no live drainer script contents, no copy of malicious JavaScript payloads.
- No unredacted personal information about a *suspected* individual beyond what's operationally necessary to investigate (a receiving wallet address is fine; doxxing a person is not).
- No unarchived links that would need to be visited live and unsandboxed to verify — prefer archive.org/similar snapshots over a link that could reinfect a reviewer's browser.

---

## 5. Signature YAML — exact structure

One file per cluster, placed in the subdirectory matching its **primary** evidence type. Filename convention: `<id>.yaml`, and the filename's basename must equal the `id` field.

```
signatures/
  favicon/        # signatures whose strongest/only criterion is a favicon hash
  javascript/      # strongest criterion is a script hash
  certificates/     # strongest criterion is a TLS cert fingerprint
  wallets/         # strongest criterion is a known drainer receiving address
  telegram/         # strongest criterion is an exfiltration bot pattern
  domains/          # strongest criterion is a known-bad domain list
  infrastructure/    # multi-vector clusters combining several of the above (see the bundled example)
  rules/            # reserved for future composite/behavioral matching — no schema
                     # implemented yet; do not place plain Signature YAML here
```

```yaml
# signatures/infrastructure/wallet_drainer_cluster_001.yaml
id: wallet_drainer_cluster_001
description: Fake wallet connection infrastructure reusing a shared drainer script and exfiltration bot
favicon:
  murmur3:
    - -204998123                          # int32, Shodan http.favicon.hash convention
  sha256: []                              # optional alternative/addition to murmur3
javascript:
  sha256:
    - e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
certificates: []                          # optional: TLS cert SHA256 fingerprints
wallets:                                  # optional: chain -> receiving addresses
  ethereum:
    - "0x123456789abcdef0123456789abcdef01234567"
telegram:
  patterns:
    - api.telegram.org/bot
domains: []                               # optional: known-bad domains for this cluster
ips: []                                   # optional: shared hosting IPs (medium strength — see §5.3)
nameservers: []                           # optional: shared nameservers (medium strength)
confidence: high                          # required — low | medium | high | critical
```

**Field rules** (enforced by `intel.Signature.Validate()` — CI rejects any violation):
- `id` is required, must match `^[a-z0-9][a-z0-9_-]*[a-z0-9]$` (lowercase, alphanumeric, `-`/`_` separators), and must be unique across the *entire* signature tree — `LoadSignatures` fails the whole batch on any duplicate.
- `description` is required — plain-language, specific enough that a reader who isn't the author understands what this cluster is.
- `confidence` is required and must be one of `low`, `medium`, `high`, `critical`.
- Every SHA256 field (`favicon.sha256`, `javascript.sha256`, `certificates`) must be exactly 64 hex characters — anything else fails validation.
- At least one of `favicon`, `javascript`, `certificates`, `wallets`, `telegram`, or `domains` must be non-empty (`hasCriteria()`). A signature with zero matchable indicators is rejected — it can never fire and is a maintainer error, not a valid entry.
- `ips`/`nameservers` currently have no independent bearing on `Validate()`'s "has criteria" check beyond being counted — treat them as *supporting*, never sole, evidence (§5.3).

### 5.1 Choosing `confidence`
This is a judgment call `Validate()` cannot make for you — it only checks the value is one of the four allowed strings, not whether it's honest:
- **critical / high** — reserved for clusters with at least one specific, non-generic strong indicator (a bespoke script hash, a specific favicon that isn't a stock template, a certificate fingerprint tied to this infrastructure alone).
- **medium** — shared infrastructure signals only (a hosting IP, a nameserver) without a specific artifact match.
- **low** — weak/contextual signals (e.g. would back a domain-age or keyword-similarity match once that evidence tier is wired in — not yet consumed by the scan pipeline as of this writing).

### 5.2 Hash format cheat sheet
| Field | Source command |
|---|---|
| `favicon.murmur3` | MurmurHash3 x86-32 of the base64-encoded favicon bytes (Shodan `http.favicon.hash` convention) — see `internal/fingerprint`'s favicon hasher, do not hand-roll this |
| `favicon.sha256` / `javascript.sha256` / `certificates` | `sha256sum <file>` or `openssl x509 -in cert.pem -noout -fingerprint -sha256` |

### 5.3 False-positive blast radius — the review's real job
CI checks schema. It cannot check whether an indicator is *too common to signature*. A reviewer must reject (not just flag) any proposal where:
- A favicon hash matches a default/stock icon shipped by a popular framework or CMS, not something the phishing kit generated itself.
- An IP or nameserver is shared, widely-used hosting (major cloud providers, big shared-hosting platforms) rather than infrastructure specific to this cluster.
- A JavaScript hash matches an unmodified, widely-used public library rather than attacker-authored code.
- A domain, wallet, or Telegram pattern is too broad (e.g. a bare `t.me` reference with no bot-specific path).

If in doubt, downgrade `confidence` or drop the indicator rather than ship something that will flag legitimate infrastructure.

### 5.4 Independent hash re-derivation (non-negotiable)
A Signature Maintainer must never copy a hash out of a submitted Observation or a third-party report and paste it into a Signature file. Every hash in a merged Signature must have been computed by a maintainer, from an artifact the maintainer personally pulled (or from an archived/sandboxed copy) — not trusted secondhand. This is the single most important rule in this document: it is the only thing standing between "someone claimed this hash matches a drainer" and "this hash actually, verifiably does."

---

## 6. Merge & release rules

- **Two-person rule on every Signature merge.** The author never approves their own PR. A second Signature Maintainer independently re-derives at least one indicator before approving.
- **CI is a floor, not a review.** `LoadSignatures`' schema/hash/duplicate checks must pass, but passing CI never substitutes for §5.3/§5.4's human judgment.
- **Signing is a separate, deliberate action from merging.** A merge to `main` never auto-triggers a signed release. The Release Signer (or their CI process, keyed to a secret only they control) explicitly builds and signs a release archive as its own step.
- **The signing key never touches a Signature Maintainer's machine.** Content review and cryptographic signing are different trust domains by design — see §2's Release Signer rationale.
- **A rejected `alexiares update` never falls back to trusting anything unsigned.** This is enforced in code (`internal/update.Run` returns `ErrNoPublicKey` / a verification error rather than proceeding) — the process must never route around it operationally (e.g. by telling users to manually download and drop in an unverified `signatures/` directory as standard guidance).

---

## 7. Hard rules — never

- Never write code that converts an `Observation` into a `Signature` automatically. This is enforced by `internal/intel`'s package doc as a structural rule, not a style preference — do not add one, even behind a flag.
- Never merge a signature the author also approved.
- Never trust a submitter-provided hash without independently re-deriving it.
- Never signature an indicator that fails the §5.3 blast-radius check.
- Never include working exploit code, live drainer scripts, or executable payloads inside `signatures/` or `observations/` — these are detection fingerprints, not malware samples.
- Never let the Ed25519 private key exist in plaintext outside its designated secure storage (no commits, no CI logs, no chat, no shared drives).
- Never publish a signature archive that CI's own validation didn't pass.
- Never ship personally identifying information about a suspected individual beyond what's operationally necessary to investigate the infrastructure.

---

## 8. Quick checklist (for a PR description)

**Observation PRs:**
- [ ] `domain`, `reporter`, `reported_at` filled in truthfully
- [ ] Evidence attached is real, not embellished
- [ ] No exploit code, no live unsandboxed links, no unnecessary PII

**Signature PRs:**
- [ ] `id` is unique, lowercase, matches the required pattern, and equals the filename
- [ ] Placed in the directory matching its primary evidence type
- [ ] Every hash independently re-derived by the author (§5.4), not copied from a report
- [ ] `confidence` honestly reflects §5.1's tiers
- [ ] Passed the §5.3 false-positive blast-radius check
- [ ] A second Signature Maintainer reviewed and independently verified at least one indicator
- [ ] `LoadSignatures` (CI) is green

See also: [README.md](README.md) for the project overview, [roadmaps.MD](roadmaps.MD) for the canonical-repository timeline (V1.5), [the technical specification](<Alexiares Technical Specification (specs.md).md>) for the Signature Repository spec, and [docs/intel.md](docs/intel.md) for the `internal/intel` package reference.
