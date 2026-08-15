package redirect_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/redirect"
)

func TestExtractMetaRefresh(t *testing.T) {
	rawHTML := `<html><head><meta http-equiv="refresh" content="0;url=https://evil.example/final"></head></html>`

	got := redirect.Extract(rawHTML, nil, "https://phish.example/")

	if len(got) != 1 {
		t.Fatalf("Extract() = %v, want 1 redirect", got)
	}
	if got[0].Method != "meta_refresh" {
		t.Errorf("Method = %q, want meta_refresh", got[0].Method)
	}
	if got[0].To != "https://evil.example/final" {
		t.Errorf("To = %q, want https://evil.example/final", got[0].To)
	}
}

func TestExtractMetaRefreshRelativeURL(t *testing.T) {
	rawHTML := `<html><head><meta http-equiv="refresh" content="3; url=/step2"></head></html>`
	got := redirect.Extract(rawHTML, nil, "https://phish.example/start")
	if len(got) != 1 {
		t.Fatalf("Extract() = %v, want 1 redirect", got)
	}
	if got[0].To != "https://phish.example/step2" {
		t.Errorf("To = %q, want resolved absolute URL", got[0].To)
	}
}

func TestExtractJavaScriptRedirectInline(t *testing.T) {
	rawHTML := `<html><body><script>window.location.href = "https://evil.example/drain";</script></body></html>`
	got := redirect.Extract(rawHTML, nil, "https://phish.example/")
	if len(got) != 1 {
		t.Fatalf("Extract() = %v, want 1 redirect", got)
	}
	if got[0].Method != "javascript" {
		t.Errorf("Method = %q, want javascript", got[0].Method)
	}
	if got[0].To != "https://evil.example/drain" {
		t.Errorf("To = %q", got[0].To)
	}
}

func TestExtractJavaScriptRedirectExternalScript(t *testing.T) {
	scripts := []string{`location.replace("https://evil.example/final")`}
	got := redirect.Extract("<html></html>", scripts, "https://phish.example/")
	if len(got) != 1 || got[0].To != "https://evil.example/final" {
		t.Errorf("Extract() = %v, want 1 redirect to https://evil.example/final", got)
	}
}

func TestExtractNoRedirects(t *testing.T) {
	got := redirect.Extract("<html><body>nothing here</body></html>", nil, "https://phish.example/")
	if len(got) != 0 {
		t.Errorf("Extract() = %v, want no redirects", got)
	}
}

func TestExtractDeduplicatesJSRedirects(t *testing.T) {
	scripts := []string{
		`window.location = "https://evil.example/x"`,
		`window.location = "https://evil.example/x"`,
	}
	got := redirect.Extract("<html></html>", scripts, "https://phish.example/")
	if len(got) != 1 {
		t.Errorf("Extract() = %v, want deduplicated to 1", got)
	}
}
