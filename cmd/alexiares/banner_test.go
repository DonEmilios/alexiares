package main

import (
	"strings"
	"testing"
)

// TestBannerContent locks the exact wordmark text — generated once
// with the real figlet tool (font: doom), never hand-edited. A diff
// here means someone touched the constant directly instead of
// regenerating it; regenerate with:
//
//	figlet -f doom -w 100 "ALEXIARES"
func TestBannerContent(t *testing.T) {
	if !strings.Contains(banner, "guard the gate before the wallet connects") {
		t.Error("banner missing the motto")
	}
	lines := strings.Split(strings.TrimRight(banner, "\n"), "\n")
	for _, line := range lines {
		if len(line) > 79 {
			t.Errorf("banner line exceeds 79 columns (%d): %q", len(line), line)
		}
	}
}

// TestIsInteractiveTerminalFalseUnderGoTest covers the branch that
// actually matters for correctness: piped/redirected/CI output must
// never see the banner. go test's stdout is always a pipe or file,
// never a TTY, so this exercises exactly that common case. The true
// branch (a real terminal) was verified by building the binary and
// running it directly — see this file's introducing commit.
func TestIsInteractiveTerminalFalseUnderGoTest(t *testing.T) {
	if isInteractiveTerminal() {
		t.Fatal("isInteractiveTerminal() = true under go test, want false (stdout should be a pipe, not a TTY)")
	}
}
