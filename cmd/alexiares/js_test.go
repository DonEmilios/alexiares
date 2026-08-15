package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSCommandExtractsArtifacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body>
			<script>fetch("https://api.telegram.org/bot123456789:AAExampleTokenValue1234567890123456/sendMessage")</script>
		</body></html>`))
	}))
	defer srv.Close()

	stdout, _, err := execute(t, "js", srv.URL)
	if err != nil {
		t.Fatalf("execute(js) error = %v", err)
	}

	for _, want := range []string{"Scripts analyzed: 1", "api.telegram.org/bot"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestJSCommandRejectsWalletTarget(t *testing.T) {
	_, _, err := execute(t, "js", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err == nil {
		t.Fatal("execute(js, wallet address): error = nil, want rejection")
	}
}

func TestJSCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "js"); err == nil {
		t.Error("execute(js) with no args: error = nil, want arg-count error")
	}
}
