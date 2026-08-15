package dns_test

import (
	"context"
	"errors"
	"net"
	"sort"
	"testing"

	"github.com/alexiares/alexiares/internal/dns"
)

// fakeResolver implements dns.Lookuper entirely in memory, so tests
// never touch the network.
type fakeResolver struct {
	hosts map[string][]string
	cname map[string]string
	mx    map[string][]*net.MX
	ns    map[string][]*net.NS
	txt   map[string][]string
	ptr   map[string][]string
}

func (f *fakeResolver) LookupHost(_ context.Context, host string) ([]string, error) {
	if v, ok := f.hosts[host]; ok {
		return v, nil
	}
	return nil, errors.New("no such host")
}

func (f *fakeResolver) LookupCNAME(_ context.Context, host string) (string, error) {
	if v, ok := f.cname[host]; ok {
		return v, nil
	}
	return host + ".", nil
}

func (f *fakeResolver) LookupMX(_ context.Context, name string) ([]*net.MX, error) {
	if v, ok := f.mx[name]; ok {
		return v, nil
	}
	return nil, errors.New("no MX records")
}

func (f *fakeResolver) LookupNS(_ context.Context, name string) ([]*net.NS, error) {
	if v, ok := f.ns[name]; ok {
		return v, nil
	}
	return nil, errors.New("no NS records")
}

func (f *fakeResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if v, ok := f.txt[name]; ok {
		return v, nil
	}
	return nil, errors.New("no TXT records")
}

func (f *fakeResolver) LookupAddr(_ context.Context, addr string) ([]string, error) {
	if v, ok := f.ptr[addr]; ok {
		return v, nil
	}
	return nil, errors.New("no PTR records")
}

func TestLookupPopulatesAllRecordTypes(t *testing.T) {
	fake := &fakeResolver{
		hosts: map[string][]string{"phish.example": {"203.0.113.42", "2001:db8::1"}},
		cname: map[string]string{"phish.example": "cdn.evil.example."},
		mx:    map[string][]*net.MX{"phish.example": {{Host: "mail.phish.example.", Pref: 10}}},
		ns:    map[string][]*net.NS{"phish.example": {{Host: "ns1.evil.example."}, {Host: "ns2.evil.example."}}},
		txt:   map[string][]string{"phish.example": {"v=spf1 include:_spf.evil.example ~all"}},
		ptr:   map[string][]string{"203.0.113.42": {"host.bulletproof.example."}},
	}

	got := dns.NewWithResolver(fake).Lookup(context.Background(), "phish.example")

	if len(got.IPs) != 1 || got.IPs[0] != "203.0.113.42" {
		t.Errorf("IPs = %v, want [203.0.113.42]", got.IPs)
	}
	if len(got.AAAA) != 1 || got.AAAA[0] != "2001:db8::1" {
		t.Errorf("AAAA = %v, want [2001:db8::1]", got.AAAA)
	}
	if got.CNAME != "cdn.evil.example" {
		t.Errorf("CNAME = %q, want cdn.evil.example", got.CNAME)
	}
	if len(got.MX) != 1 || got.MX[0] != "mail.phish.example" {
		t.Errorf("MX = %v, want [mail.phish.example]", got.MX)
	}
	sort.Strings(got.Nameservers)
	if len(got.Nameservers) != 2 || got.Nameservers[0] != "ns1.evil.example" {
		t.Errorf("Nameservers = %v, want 2 trimmed entries", got.Nameservers)
	}
	if len(got.TXT) != 1 {
		t.Errorf("TXT = %v, want 1 record", got.TXT)
	}
	if len(got.PTR) != 1 || got.PTR[0] != "host.bulletproof.example" {
		t.Errorf("PTR = %v, want [host.bulletproof.example]", got.PTR)
	}
}

func TestLookupMissingRecordsAreOmittedNotErrors(t *testing.T) {
	fake := &fakeResolver{
		hosts: map[string][]string{"minimal.example": {"203.0.113.1"}},
	}

	got := dns.NewWithResolver(fake).Lookup(context.Background(), "minimal.example")

	if len(got.IPs) != 1 {
		t.Errorf("IPs = %v, want 1 entry", got.IPs)
	}
	if got.CNAME != "" {
		t.Errorf("CNAME = %q, want empty (self-referencing CNAME suppressed)", got.CNAME)
	}
	if len(got.MX) != 0 || len(got.Nameservers) != 0 || len(got.TXT) != 0 || len(got.PTR) != 0 {
		t.Errorf("expected all optional record types empty, got %+v", got)
	}
}

func TestLookupUnresolvableDomainReturnsEmptyNotError(t *testing.T) {
	fake := &fakeResolver{}
	got := dns.NewWithResolver(fake).Lookup(context.Background(), "does-not-exist.invalid")
	if len(got.IPs) != 0 {
		t.Errorf("IPs = %v, want empty for unresolvable domain", got.IPs)
	}
}
