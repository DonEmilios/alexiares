// Package wallet detects cryptocurrency wallet addresses in text and
// classifies standalone address strings by chain.
//
// Detection is format-based (regular expressions matching each chain's
// address shape). It is not checksum validation: a string that matches
// a chain's pattern is reported as a candidate address for that chain,
// even if its checksum would fail. This mirrors how phishing kits
// themselves rarely validate checksums either, and keeps detection
// dependency-free.
package wallet

import (
	"regexp"
	"sort"

	"github.com/alexiares/alexiares/internal/artifact"
)

type pattern struct {
	chain    artifact.WalletChain
	body     string // pattern body, without \b anchors
	find     *regexp.Regexp
	classify *regexp.Regexp
}

func newPattern(chain artifact.WalletChain, body string) pattern {
	return pattern{
		chain:    chain,
		body:     body,
		find:     regexp.MustCompile(`\b(?:` + body + `)\b`),
		classify: regexp.MustCompile(`^(?:` + body + `)$`),
	}
}

// Order matters: chains with a distinguishing prefix are matched
// before Solana's pattern, which is a generic base58 blob with no
// prefix and would otherwise shadow Tron and Bitcoin addresses that
// happen to also be valid base58.
var patterns = []pattern{
	newPattern(artifact.ChainEthereum, `0x[0-9a-fA-F]{40}`),
	newPattern(artifact.ChainBitcoin, `bc1[a-z0-9]{25,90}|[13][a-km-zA-HJ-NP-Z1-9]{25,34}`),
	newPattern(artifact.ChainCardano, `addr1[a-z0-9]{20,103}`),
	newPattern(artifact.ChainTron, `T[1-9A-HJ-NP-Za-km-z]{33}`),
	newPattern(artifact.ChainTON, `(?:EQ|UQ)[A-Za-z0-9_-]{46}`),
	newPattern(artifact.ChainSolana, `[1-9A-HJ-NP-Za-km-z]{32,44}`),
}

// ensAddress matches ".eth" ENS names.
var ensAddress = regexp.MustCompile(`\b[a-zA-Z0-9-]{1,63}\.eth\b`)

// Detect scans arbitrary text for wallet addresses across all
// supported chains and returns the deduplicated matches in a stable
// order (by chain, then by address).
func Detect(text string) artifact.WalletArtifacts {
	seen := make(map[artifact.WalletAddress]bool)
	var out []artifact.WalletAddress

	for _, p := range patterns {
		for _, m := range p.find.FindAllString(text, -1) {
			addr := artifact.WalletAddress{Chain: p.chain, Address: m}
			if !seen[addr] {
				seen[addr] = true
				out = append(out, addr)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Chain != out[j].Chain {
			return out[i].Chain < out[j].Chain
		}
		return out[i].Address < out[j].Address
	})

	ensSeen := make(map[string]bool)
	var ens []string
	for _, m := range ensAddress.FindAllString(text, -1) {
		if !ensSeen[m] {
			ensSeen[m] = true
			ens = append(ens, m)
		}
	}
	sort.Strings(ens)

	return artifact.WalletArtifacts{Addresses: out, ENS: ens}
}

// Classify reports whether address matches exactly one supported
// chain's format. It anchors the match to the full string, unlike
// Detect which finds addresses embedded in larger text.
func Classify(address string) (artifact.WalletChain, bool) {
	for _, p := range patterns {
		if p.classify.MatchString(address) {
			return p.chain, true
		}
	}
	return "", false
}
