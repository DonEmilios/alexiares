package main

import (
	"strings"
	"testing"
)

func TestBareInvocationShowsHelpNotError(t *testing.T) {
	stdout, _, err := execute(t)
	if err != nil {
		t.Fatalf("execute() with no args error = %v, want nil", err)
	}
	if !strings.Contains(stdout, "Available Commands:") {
		t.Errorf("stdout = %q, want help text", stdout)
	}
}

// TestBareInvocationNeverPrintsBannerInTests guards the piped/scripted
// case structurally: execute()'s captured output goes through
// SetOut/SetErr, but isInteractiveTerminal() checks the real process
// stdout, which go test never runs as a TTY — so the banner must be
// structurally absent here, the same way it's absent for any piped
// invocation of the real binary.
func TestBareInvocationNeverPrintsBannerInTests(t *testing.T) {
	stdout, _, err := execute(t)
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	for _, bannerLine := range strings.Split(strings.TrimSpace(banner), "\n") {
		if bannerLine == "" {
			continue
		}
		if strings.Contains(stdout, bannerLine) {
			t.Errorf("stdout unexpectedly contains a banner line %q; banner must not appear in non-interactive output", bannerLine)
		}
	}
}

func TestUnknownCommandErrorsWithHint(t *testing.T) {
	_, _, err := execute(t, "bogus-command")
	if err == nil {
		t.Fatal("execute(bogus-command) error = nil, want an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, `unknown command "bogus-command"`) {
		t.Errorf("error = %q, want it to name the unknown command", msg)
	}
	if !strings.Contains(msg, "Run 'alexiares --help' for usage.") {
		t.Errorf("error = %q, want the --help hint preserved", msg)
	}
}

func TestKnownSubcommandsStillWork(t *testing.T) {
	stdout, _, err := execute(t, "version")
	if err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}
	if !strings.Contains(stdout, "alexiares") {
		t.Errorf("stdout = %q, want version output", stdout)
	}
}
