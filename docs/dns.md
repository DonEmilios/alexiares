# `internal/dns`

**Source:** [`internal/dns/dns.go`](../internal/dns/dns.go)
**Tests:** `dns_test.go` — 93.3% coverage
**Position in pipeline:** extractor stage — the only extractor that performs its own network I/O

## Purpose

Resolves a domain's full DNS record set: A, AAAA, CNAME, MX, NS, TXT, and reverse PTR. The collector's implicit resolution (dialing the target) only proves a hostname resolves *to something*; it never surfaces the actual record set, which is what shared-nameserver and shared-IP correlation needs.

```go
type Lookuper interface { LookupHost, LookupCNAME, LookupMX, LookupNS, LookupTXT, LookupAddr }
func New() *Extractor                        // backed by net.DefaultResolver
func NewWithResolver(r Lookuper) *Extractor   // backed by a fake, for tests
func (e *Extractor) Lookup(ctx, domain) artifact.DNSArtifacts
```

## Design notes

**`Lookuper` is an interface matching `*net.Resolver`'s method set exactly, so it's satisfied implicitly** — production code never wraps or adapts anything, it just passes `net.DefaultResolver` straight in. Tests inject a fake implementing the same six methods, entirely in memory, so `dns_test.go` never touches the network (per the project's testing policy) while still exercising the real branching logic.

**`Lookup` never returns an error.** Each record type is resolved independently and a missing one (no MX record is completely normal for most domains) is not a failure — the function returns whatever subset it could collect. This is deliberate: a hard error here would mean one missing TXT record aborts an entire scan, which is the wrong failure mode for "partial collection is a result, not an error," the same policy `internal/collector` follows for failed script downloads.

**RFC 7505 null MX records are suppressed, not reported as blank.** `net.LookupMX` returns a single record with `Host: "."` for a domain that explicitly declares "I accept no mail" (RFC 7505). Trimming that record's trailing dot naively produces an empty string — which, unguarded, gets appended to `DNSArtifacts.MX` anyway and renders as a blank line under the MX section. This was a real bug, found live scanning `example.com` during testing, fixed by checking `host != ""` after trimming on MX, NS, and PTR alike (the same latent pattern existed in all three, even though only MX was observed failing).

**A self-referencing CNAME is suppressed.** `net.LookupCNAME` returns the canonical name even when there's no real CNAME record — in that case it returns the domain itself, trailing-dotted. `Lookup` compares the trimmed result against `domain` and leaves `CNAME` empty rather than reporting a domain as its own alias.

## Known limitation

**ASN and ASN-org attribution are never populated**, and this is explained in the package doc comment itself, not hidden: attributing an IP to its autonomous system requires an external IP-to-ASN database, and Alexiares' performance goal is explicitly "no external database required." `artifact.DNSArtifacts.ASN`/`ASNOrg` exist as fields (for a future bundled or fetched database to fill in) but nothing populates them today. Correlation on ASN and registrar data is consequently unimplemented too — see [`correlation.md`](correlation.md).
