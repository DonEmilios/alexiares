package main

import (
	"strings"
	"testing"
)

func TestWalletCommandValidAddress(t *testing.T) {
	stdout, _, err := execute(t, "wallet", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590")
	if err != nil {
		t.Fatalf("execute(wallet) error = %v", err)
	}
	if !strings.Contains(stdout, "chain: ethereum") {
		t.Errorf("stdout = %q, want it to contain the classified chain", stdout)
	}
}

func TestWalletCommandInvalidAddress(t *testing.T) {
	_, _, err := execute(t, "wallet", "not-a-wallet-address")
	if err == nil {
		t.Fatal("execute(wallet, invalid) error = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "does not match a supported chain") {
		t.Errorf("error = %q, want mention of unsupported format", err)
	}
}

func TestWalletCommandRequiresExactlyOneArg(t *testing.T) {
	if _, _, err := execute(t, "wallet"); err == nil {
		t.Error("execute(wallet) with no args: error = nil, want arg-count error")
	}
	if _, _, err := execute(t, "wallet", "a", "b"); err == nil {
		t.Error("execute(wallet) with 2 args: error = nil, want arg-count error")
	}
}
