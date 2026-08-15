package wallet_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/wallet"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    artifact.WalletChain
		wantOK  bool
	}{
		{"ethereum", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590", artifact.ChainEthereum, true},
		{"bitcoin legacy", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", artifact.ChainBitcoin, true},
		{"bitcoin bech32", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", artifact.ChainBitcoin, true},
		{"tron", "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf", artifact.ChainTron, true},
		{"cardano", "addr1qxck9c2ty0dqhpvtkjhqyzn5v2f0rjs0ny8u4dqvz2z2mjvk", artifact.ChainCardano, true},
		{"ton", "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", artifact.ChainTON, true},
		{"too short for any chain", "0x1234", "", false},
		{"plain word", "notanaddress", "", false},
		{"trailing text rejected", "0x742d35Cc6634C0532925a3b844Bc9e7595f89590x", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := wallet.Classify(tt.address)
			if ok != tt.wantOK {
				t.Fatalf("Classify(%q) ok = %v, want %v", tt.address, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("Classify(%q) chain = %v, want %v", tt.address, got, tt.want)
			}
		})
	}
}

func TestDetectFindsEmbeddedAddressesAndENS(t *testing.T) {
	text := `
		Send funds to 0x742d35Cc6634C0532925a3b844Bc9e7595f89590 or
		vitalik.eth. Backup: 0x742d35Cc6634C0532925a3b844Bc9e7595f89590
	`
	got := wallet.Detect(text)

	if len(got.Addresses) != 1 {
		t.Fatalf("Detect() found %d addresses, want 1 deduplicated match: %+v", len(got.Addresses), got.Addresses)
	}
	if got.Addresses[0].Chain != artifact.ChainEthereum {
		t.Errorf("Addresses[0].Chain = %v, want ethereum", got.Addresses[0].Chain)
	}
	if len(got.ENS) != 1 || got.ENS[0] != "vitalik.eth" {
		t.Errorf("ENS = %v, want [vitalik.eth]", got.ENS)
	}
}

func TestDetectEmptyTextReturnsNoMatches(t *testing.T) {
	got := wallet.Detect("nothing suspicious here")
	if len(got.Addresses) != 0 || len(got.ENS) != 0 {
		t.Errorf("Detect() = %+v, want empty", got)
	}
}
