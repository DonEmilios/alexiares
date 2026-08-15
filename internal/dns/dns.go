// Package dns collects DNS intelligence for a domain: A/AAAA records,
// nameservers, MX records, TXT records, CNAME, and reverse PTR
// records. Unlike Alexiares' other extractors, this one performs its
// own network I/O — the collector's implicit resolution only proves a
// hostname resolves, it does not surface the full record set.
//
// ASN and ASN organization attribution are intentionally left empty:
// they require an external IP-to-ASN database, which v1 does not
// bundle, in keeping with the "no external database required"
// performance goal.
package dns

import (
	"context"
	"net"
	"strings"

	"github.com/alexiares/alexiares/internal/artifact"
)

// Lookuper is the subset of *net.Resolver's methods the extractor
// needs. It exists so tests can inject a fake resolver and run
// entirely offline, per Alexiares' testing policy.
type Lookuper interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}

// Extractor collects DNS artifacts using a Lookuper.
type Extractor struct {
	resolver Lookuper
}

// New builds an Extractor backed by Go's default system resolver.
func New() *Extractor {
	return &Extractor{resolver: net.DefaultResolver}
}

// NewWithResolver builds an Extractor backed by a caller-supplied
// Lookuper, primarily for tests.
func NewWithResolver(r Lookuper) *Extractor {
	return &Extractor{resolver: r}
}

// Lookup resolves domain's DNS record set. Each record type is
// resolved independently: a domain lacking MX or TXT records is
// normal, not an error, so Lookup never fails outright — it returns
// whatever subset of records it could collect. This matches Alexiares'
// policy that partial collection is a result, not a fatal error.
func (e *Extractor) Lookup(ctx context.Context, domain string) artifact.DNSArtifacts {
	domain = strings.TrimSuffix(domain, ".")
	var out artifact.DNSArtifacts

	if addrs, err := e.resolver.LookupHost(ctx, domain); err == nil {
		for _, a := range addrs {
			ip := net.ParseIP(a)
			switch {
			case ip == nil:
				continue
			case ip.To4() != nil:
				out.IPs = append(out.IPs, a)
			default:
				out.AAAA = append(out.AAAA, a)
			}
		}
	}

	if cname, err := e.resolver.LookupCNAME(ctx, domain); err == nil {
		trimmed := strings.TrimSuffix(cname, ".")
		if !strings.EqualFold(trimmed, domain) {
			out.CNAME = trimmed
		}
	}

	if mxs, err := e.resolver.LookupMX(ctx, domain); err == nil {
		for _, mx := range mxs {
			out.MX = append(out.MX, strings.TrimSuffix(mx.Host, "."))
		}
	}

	if nss, err := e.resolver.LookupNS(ctx, domain); err == nil {
		for _, ns := range nss {
			out.Nameservers = append(out.Nameservers, strings.TrimSuffix(ns.Host, "."))
		}
	}

	if txts, err := e.resolver.LookupTXT(ctx, domain); err == nil {
		out.TXT = txts
	}

	for _, ip := range out.IPs {
		names, err := e.resolver.LookupAddr(ctx, ip)
		if err != nil {
			continue
		}
		for _, n := range names {
			out.PTR = append(out.PTR, strings.TrimSuffix(n, "."))
		}
	}

	return out
}
