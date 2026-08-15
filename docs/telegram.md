# `internal/telegram`

**Source:** [`internal/telegram/telegram.go`](../internal/telegram/telegram.go)
**Tests:** `telegram_test.go` — 100% coverage
**Position in pipeline:** extractor stage — pure pattern matching, no I/O

## Purpose

Detects Telegram-based exfiltration indicators in already-collected text: bot tokens, chat IDs, `api.telegram.org` references, and `t.me` links. Wallet drainers commonly exfiltrate signed transactions or seed phrases to a Telegram bot — this is one of the spec's designated *strong* evidence categories (see [`evidence.md`](evidence.md)) for exactly that reason.

```go
func Extract(text string) artifact.TelegramArtifacts
```

## Design notes

**The bot-token regex has no leading word boundary, and that's a fix, not the original design.** Real Telegram bot API calls look like `.../bot123456789:TOKEN/sendMessage` — the digits of the bot ID appear directly after the literal `bot` prefix with no separator. A regex `\b\d{8,10}:...` looks reasonable but is wrong: `\b` matches at a transition between a "word" character and a non-word character, and both `t` (end of "bot") and `1` (start of the digits) are word characters — there is no boundary there at all, so the pattern silently fails to match the single most common real-world occurrence of the string it's meant to catch. Caught by a failing test during the build (`TestExtractDetectsAllIndicators`), fixed by dropping the leading `\b` and relying on the trailing one (which *is* a real boundary, since a token is followed by `/` or end-of-string).

**Chat IDs are extracted via a capture group, not a standalone pattern**, because a bare "6 to 15 digit number, possibly negative" has no distinguishing shape of its own — `-1001234567890` looks like nothing in particular out of context. `chatIDRe` anchors on the `chat_id` key itself (`chat_id["'\s:=]+(-?\d{6,15})`, case-insensitive) so only numbers that are actually labeled as a chat ID get reported, not every stray large integer in a script.

**Every result field is deduplicated and sorted**, same pattern as `internal/javascript`'s regex categories — required for `Extract`'s output to be byte-identical across repeated calls on the same input, which the correlation and evidence engines downstream depend on for deterministic matching.
