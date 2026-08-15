# `internal/config`

**Source:** [`internal/config/config.go`](../internal/config/config.go)
**Tests:** `config_test.go` — 76.9% coverage (just under the project's 80% floor — see Known limitation below)
**Position in pipeline:** loaded once at the start of every command via `loadConfig()` in `cmd/alexiares/root.go`

## Purpose

Loads and validates `~/.alexiares/config.yaml` (or whatever path `--config` points at), merged over built-in defaults.

```go
func Default() Config
func DefaultPath() string
func Load(path string) (Config, error)

type Config struct { Network, Signatures, Output, Update }
func (n Network) Timeout() time.Duration
func (u Update) DecodedPublicKey() (ed25519.PublicKey, error)
```

## Design notes

**A missing config file is not an error — it's the normal case.** The package doc states this directly: *"Missing files are not an error: Load returns the built-in defaults so the CLI works with zero setup."* This is what lets `alexiares scan <url>` work immediately on a fresh install with no configuration step required first.

**`Load` unmarshals onto an already-populated `Default()`, not a zero-value struct.** This is a small but easy-to-get-wrong detail: `yaml.Unmarshal(data, &cfg)` only overwrites fields the YAML file actually mentions — a config file that sets `output.format: json` and nothing else still gets the default network timeout and user agent, rather than those fields silently becoming Go's zero values (timeout `0`, meaning "no timeout" to `net/http`, which would be a real behavioral bug, not just a missing default).

**`Update.PublicKey` is a hex string in YAML, decoded to `ed25519.PublicKey` only when actually needed** (`DecodedPublicKey`, called from `cmd/alexiares/update.go`), not eagerly at load time. Hex is the natural text encoding for a 32-byte key in a YAML file a human might paste into; keeping `Config` itself as plain strings (rather than decoding into `[]byte` at load time) means a config file with a malformed key doesn't break `Load` for commands that never touch `update` at all — the decode error only surfaces when `alexiares update` actually runs.

**`Update`'s doc comment states the "no fallback" policy at the config layer, not just in `internal/update`:** *"Both fields must be set for an update to run at all: Alexiares never fetches signature updates without a trusted key to verify them against."* This is why `configs/default.yaml` ships with `update.source_url: ""` and `update.public_key: ""` rather than guessing at a plausible-looking default — an empty value fails loudly and immediately (`cmd/alexiares/update.go` checks this before any network call), rather than a wrong-but-present-looking default failing confusingly later at signature verification.

## Known limitation

**Coverage sits at 76.9%, just under the 80% floor.** The untested paths are the `os.UserHomeDir()` failure branches in `Default()` and `DefaultPath()` (falling back to `"."` when the OS can't determine a home directory) — a condition that's difficult to trigger portably in a test without mocking the OS layer, and rare enough in practice (a process with no resolvable home directory) that it wasn't judged worth the added complexity of an injectable home-directory abstraction just to cover it.
