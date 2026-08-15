package collector

import (
	"bufio"
	"io"
	"net"
	"strings"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/wallet"
)

// Kind identifies what a CLI input string represents, so callers can
// route it to network collection or direct wallet analysis without
// ever treating a wallet address as a fetchable URL.
type Kind int

// Supported input kinds.
const (
	KindUnknown Kind = iota
	KindURL
	KindIP
	KindWallet
)

// Input is a classified CLI target.
type Input struct {
	Kind Kind
	Raw  string
	// URL is the normalized, fetchable URL for KindURL and KindIP
	// inputs (a bare domain or IP is upgraded to "https://...").
	URL   string
	Chain artifact.WalletChain // set only for KindWallet
}

// Classify determines what kind of target raw is: a wallet address, an
// IP, or something fetchable as a URL (a full URL or a bare domain).
//
// Wallet addresses are checked first, since their formats never
// collide with valid hostnames or IPs.
func Classify(raw string) Input {
	trimmed := strings.TrimSpace(raw)

	if chain, ok := wallet.Classify(trimmed); ok {
		return Input{Kind: KindWallet, Raw: trimmed, Chain: chain}
	}

	if ip := net.ParseIP(trimmed); ip != nil {
		return Input{Kind: KindIP, Raw: trimmed, URL: "https://" + trimmed}
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return Input{Kind: KindURL, Raw: trimmed, URL: trimmed}
	}

	if looksLikeDomain(trimmed) {
		return Input{Kind: KindURL, Raw: trimmed, URL: "https://" + trimmed}
	}

	return Input{Kind: KindUnknown, Raw: trimmed}
}

// looksLikeDomain applies a permissive hostname shape check: at least
// one dot, no whitespace, no path separators.
func looksLikeDomain(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t/\\") {
		return false
	}
	return strings.Contains(s, ".")
}

// ReadTargets reads one target per line from r, trimming whitespace
// and skipping empty lines and "#"-prefixed comments. It backs both
// batch-file and stdin input.
func ReadTargets(r io.Reader) ([]string, error) {
	var targets []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		targets = append(targets, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}
