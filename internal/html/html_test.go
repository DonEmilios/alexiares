package html_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/html"
)

const samplePage = `
<html>
<head>
	<title>Claim your airdrop</title>
	<meta name="description" content="Connect your wallet to claim">
	<meta property="og:title" content="Airdrop">
	<link rel="stylesheet" href="/styles.css">
	<script src="/drainer.js"></script>
	<!-- built with phishkit v2 -->
</head>
<body>
	<img src="/logo.png">
	<iframe src="https://evil.example/widget"></iframe>
	<form action="/connect" method="post">
		<input type="text" name="wallet">
		<input type="hidden" name="csrf" value="abc">
		<input type="hidden" name="referrer" value="xyz">
	</form>
</body>
</html>`

func TestExtractParsesAllArtifactTypes(t *testing.T) {
	got := html.Extract(samplePage)

	if got.Title != "Claim your airdrop" {
		t.Errorf("Title = %q, want %q", got.Title, "Claim your airdrop")
	}
	if got.Metadata["description"] != "Connect your wallet to claim" {
		t.Errorf("Metadata[description] = %q", got.Metadata["description"])
	}
	if got.Metadata["og:title"] != "Airdrop" {
		t.Errorf("Metadata[og:title] = %q", got.Metadata["og:title"])
	}
	if len(got.Comments) != 1 || got.Comments[0] != "built with phishkit v2" {
		t.Errorf("Comments = %v, want [built with phishkit v2]", got.Comments)
	}

	wantResources := map[string]bool{"/styles.css": true, "/drainer.js": true, "/logo.png": true, "https://evil.example/widget": true}
	if len(got.ExternalResources) != len(wantResources) {
		t.Fatalf("ExternalResources = %v, want %d entries", got.ExternalResources, len(wantResources))
	}
	for _, r := range got.ExternalResources {
		if !wantResources[r] {
			t.Errorf("unexpected ExternalResources entry %q", r)
		}
	}

	if len(got.Forms) != 1 {
		t.Fatalf("Forms = %v, want 1 form", got.Forms)
	}
	form := got.Forms[0]
	if form.Action != "/connect" || form.Method != "POST" {
		t.Errorf("form action/method = %q/%q, want /connect POST", form.Action, form.Method)
	}
	if len(form.Fields) != 1 || form.Fields[0] != "wallet" {
		t.Errorf("form.Fields = %v, want [wallet]", form.Fields)
	}
	if len(form.HiddenFields) != 2 {
		t.Errorf("form.HiddenFields = %v, want 2 hidden fields", form.HiddenFields)
	}
}

func TestExtractMalformedHTMLDoesNotError(t *testing.T) {
	got := html.Extract("<html><body><form><input name=broken></html>")
	if len(got.Forms) != 1 {
		t.Errorf("Forms = %v, want tolerant parse of malformed HTML", got.Forms)
	}
}

func TestExtractEmptyInput(t *testing.T) {
	got := html.Extract("")
	if len(got.Forms) != 0 || len(got.Comments) != 0 || len(got.ExternalResources) != 0 {
		t.Errorf("Extract(\"\") = %+v, want empty artifacts", got)
	}
}

func TestExtractDefaultsMethodToGET(t *testing.T) {
	got := html.Extract(`<form action="/x"><input name="q"></form>`)
	if len(got.Forms) != 1 || got.Forms[0].Method != "GET" {
		t.Errorf("Forms[0].Method = %q, want GET default", got.Forms[0].Method)
	}
}
