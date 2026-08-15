package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTMLCommandExtractsArtifacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head>
			<title>Claim your airdrop</title>
			<meta name="description" content="Connect your wallet">
			<!-- built with phishkit -->
		</head><body>
			<form action="/connect" method="post"><input name="wallet"></form>
		</body></html>`))
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "html", srv.URL)
	if err != nil {
		t.Fatalf("execute(html) error = %v", err)
	}

	for _, want := range []string{"Claim your airdrop", "description: Connect your wallet", "/connect", "wallet"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestHTMLCommandRejectsWalletTarget(t *testing.T) {
	_, _, err := execute(t, "html", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err == nil {
		t.Fatal("execute(html, wallet address): error = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "not a fetchable domain, URL, or IP") {
		t.Errorf("error = %q, want the pre-network rejection message", err)
	}
}

func TestHTMLCommandUnreachableTarget(t *testing.T) {
	_, _, err := execute(t, "html", "http://127.0.0.1:0/unreachable")
	if err == nil {
		t.Fatal("execute(html, unreachable): error = nil, want a network error")
	}
}

func TestHTMLCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "html"); err == nil {
		t.Error("execute(html) with no args: error = nil, want arg-count error")
	}
}
