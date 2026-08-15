package main

import (
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	stdout, _, err := execute(t, "version")
	if err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}
	if !strings.Contains(stdout, "alexiares "+version) {
		t.Errorf("stdout = %q, want it to contain %q", stdout, "alexiares "+version)
	}
}
