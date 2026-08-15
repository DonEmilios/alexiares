package main

import (
	"strings"
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
)

func TestPrintDNSArtifacts(t *testing.T) {
	var b strings.Builder
	d := artifact.DNSArtifacts{
		IPs:         []string{"203.0.113.1"},
		AAAA:        []string{"2001:db8::1"},
		CNAME:       "cdn.evil.example",
		MX:          []string{"mail.evil.example"},
		Nameservers: []string{"ns1.evil.example"},
		TXT:         []string{"v=spf1 -all"},
		PTR:         []string{"host.example"},
	}
	if err := printDNSArtifacts(&b, "phish.example", d); err != nil {
		t.Fatalf("printDNSArtifacts() error = %v", err)
	}
	out := b.String()

	for _, want := range []string{"phish.example", "203.0.113.1", "2001:db8::1", "cdn.evil.example", "mail.evil.example", "ns1.evil.example", "v=spf1 -all", "host.example"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintDNSArtifactsOmitsEmptySections(t *testing.T) {
	var b strings.Builder
	if err := printDNSArtifacts(&b, "minimal.example", artifact.DNSArtifacts{IPs: []string{"203.0.113.1"}}); err != nil {
		t.Fatalf("printDNSArtifacts() error = %v", err)
	}
	out := b.String()
	for _, absent := range []string{"MX:", "NS:", "TXT:", "PTR:", "CNAME:"} {
		if strings.Contains(out, absent) {
			t.Errorf("output contains %q for an empty section, want it omitted:\n%s", absent, out)
		}
	}
}

func TestHostnameOf(t *testing.T) {
	tests := []struct{ in, want string }{
		{"example.com", "example.com"},
		{"https://example.com/path", "example.com"},
		{"http://example.com:8080/", "example.com"},
	}
	for _, tt := range tests {
		if got := hostnameOf(tt.in); got != tt.want {
			t.Errorf("hostnameOf(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestDNSCommandLocalhost exercises the real dns.New().Lookup() path
// against "localhost", which every standard resolver answers from
// /etc/hosts without any network I/O (confirmed empirically:
// LookupHost/LookupAddr for loopback addresses return in
// microseconds). This is the one command test that reaches an actual
// Lookuper rather than a fake — internal/dns's own tests already
// cover the Lookuper logic exhaustively offline via a fake resolver;
// this only proves the CLI wiring reaches it correctly.
func TestDNSCommandLocalhost(t *testing.T) {
	stdout, _, err := execute(t, "dns", "localhost")
	if err != nil {
		t.Fatalf("execute(dns, localhost) error = %v", err)
	}
	if !strings.Contains(stdout, "DNS intelligence for localhost") {
		t.Errorf("stdout = %q, want the localhost header", stdout)
	}
	if !strings.Contains(stdout, "127.0.0.1") {
		t.Errorf("stdout = %q, want the loopback A record", stdout)
	}
}

func TestDNSCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "dns"); err == nil {
		t.Error("execute(dns) with no args: error = nil, want arg-count error")
	}
}
