// Package telegram detects Telegram-based exfiltration indicators in
// already-collected text: bot tokens, chat IDs, api.telegram.org
// references, and t.me links. Wallet drainers commonly exfiltrate
// signed transactions or seed phrases to a Telegram bot, making these
// indicators strong evidence of malicious infrastructure.
package telegram

import (
	"regexp"
	"sort"

	"github.com/alexiares/alexiares/internal/artifact"
)

var (
	// No leading \b: real bot tokens appear directly after a "bot"
	// prefix in API URLs (".../bot123456789:TOKEN/..."), where the
	// letter-to-digit transition is not a word boundary.
	botTokenRe = regexp.MustCompile(`\d{8,10}:[A-Za-z0-9_-]{35}\b`)
	apiRefRe   = regexp.MustCompile(`api\.telegram\.org(?:/bot[A-Za-z0-9_:-]+)?[^\s'"` + "`" + `<>)]*`)
	linkRe     = regexp.MustCompile(`\bt\.me/[A-Za-z0-9_]+`)
	chatIDRe   = regexp.MustCompile(`(?i)chat_id["'\s:=]+(-?\d{6,15})`)
)

// Extract scans text for Telegram exfiltration indicators. Each
// result field is deduplicated and sorted for deterministic output.
func Extract(text string) (out artifact.TelegramArtifacts) {
	out.BotTokens = dedupSorted(botTokenRe.FindAllString(text, -1))
	out.APIRefs = dedupSorted(apiRefRe.FindAllString(text, -1))
	out.Links = dedupSorted(linkRe.FindAllString(text, -1))

	for _, m := range chatIDRe.FindAllStringSubmatch(text, -1) {
		out.ChatIDs = append(out.ChatIDs, m[1])
	}
	out.ChatIDs = dedupSorted(out.ChatIDs)

	return out
}

func dedupSorted(matches []string) []string {
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(matches))
	var out []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	sort.Strings(out)
	return out
}
