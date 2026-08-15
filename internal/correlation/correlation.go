// Package correlation matches a scanned target's fingerprints and
// extracted artifacts against Alexiares' signature repository,
// grouping every hit per signature into a cluster and surfacing the
// infrastructure that cluster is already known to be connected to.
//
// Matching is deterministic: the same target artifacts against the
// same signature set always produce the same Correlation. It performs
// no I/O and never itself decides confidence — the signature carries
// its own maintainer-assigned confidence, which the evidence engine
// (Phase 7) turns into a recommendation.
package correlation

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/alexiares/alexiares/internal/artifact"
	"github.com/alexiares/alexiares/internal/intel"
)

// Target bundles everything about a scanned target that correlation
// can match against a signature. Fields left at their zero value
// (e.g. no TLS handshake, no wallets found) simply produce no matches
// for that category — Correlate never errors on a partial Target.
type Target struct {
	Domain          string
	Fingerprints    artifact.Fingerprints
	Wallets         artifact.WalletArtifacts
	Telegram        artifact.TelegramArtifacts
	IPs             []string // from the DNS extractor
	Nameservers     []string // from the DNS extractor
	RedirectDomains []string // hostnames the target's redirect chain led to
}

// Correlate matches target against every signature and returns every
// hit, grouped into per-signature clusters, along with the domains and
// wallets those matched signatures already know about.
func Correlate(target Target, signatures []intel.Signature) artifact.Correlation {
	var corr artifact.Correlation

	relatedDomains := make(map[string]bool)
	relatedWallets := make(map[string]bool)

	for _, sig := range signatures {
		matches := matchSignature(target, sig)
		if len(matches) == 0 {
			continue
		}

		corr.Matches = append(corr.Matches, matches...)
		corr.Clusters = append(corr.Clusters, artifact.Cluster{
			SignatureID: sig.ID,
			Description: sig.Description,
			Confidence:  string(sig.Confidence),
			Matches:     matches,
		})

		for _, d := range sig.Domains {
			if d != target.Domain {
				relatedDomains[d] = true
			}
		}
		for _, addrs := range sig.Wallets {
			for _, a := range addrs {
				relatedWallets[a] = true
			}
		}
	}

	corr.RelatedDomains = sortedKeys(relatedDomains)
	corr.RelatedWallets = sortedKeys(relatedWallets)

	return corr
}

// matchSignature checks target against every criterion sig defines,
// returning one Match per hit.
func matchSignature(target Target, sig intel.Signature) []artifact.Match {
	var matches []artifact.Match

	if target.Fingerprints.Favicon != "" {
		for _, h := range sig.Favicon.SHA256 {
			if h == target.Fingerprints.Favicon {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchFavicon, Value: h})
			}
		}
	}
	if target.Fingerprints.FaviconHash != 0 {
		for _, m := range sig.Favicon.Murmur3 {
			if m == target.Fingerprints.FaviconHash {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchFavicon, Value: strconv.FormatInt(int64(m), 10)})
			}
		}
	}

	for _, sigHash := range sig.JavaScript.SHA256 {
		for _, targetHash := range target.Fingerprints.JavaScript {
			if sigHash == targetHash {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchJavaScript, Value: sigHash})
			}
		}
	}

	if target.Fingerprints.Certificate != "" {
		for _, c := range sig.Certificates {
			if c == target.Fingerprints.Certificate {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchCertificate, Value: c})
			}
		}
	}

	for chain, addrs := range sig.Wallets {
		for _, sigAddr := range addrs {
			for _, targetAddr := range target.Wallets.Addresses {
				if string(targetAddr.Chain) == chain && targetAddr.Address == sigAddr {
					matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchWallet, Value: sigAddr})
				}
			}
		}
	}

	telegramIndicators := slices.Concat(target.Telegram.BotTokens, target.Telegram.APIRefs, target.Telegram.Links)
	for _, pattern := range sig.Telegram.Patterns {
		for _, indicator := range telegramIndicators {
			if pattern != "" && strings.Contains(indicator, pattern) {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchTelegram, Value: pattern})
			}
		}
	}

	for _, d := range sig.Domains {
		if d == target.Domain {
			matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchDomain, Value: d})
		}
		for _, redirectHost := range target.RedirectDomains {
			if d == redirectHost {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchRedirect, Value: d})
			}
		}
	}

	for _, sigIP := range sig.IPs {
		for _, targetIP := range target.IPs {
			if sigIP == targetIP {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchIP, Value: sigIP})
			}
		}
	}

	for _, sigNS := range sig.Nameservers {
		for _, targetNS := range target.Nameservers {
			if sigNS == targetNS {
				matches = append(matches, artifact.Match{SignatureID: sig.ID, Category: artifact.MatchNameserver, Value: sigNS})
			}
		}
	}

	return matches
}

func sortedKeys(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
