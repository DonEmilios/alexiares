# `internal/update`

**Source:** [`internal/update/`](../internal/update/) (`update.go`, `apply.go`)
**Tests:** `update_test.go`, `apply_test.go` — 80.6% coverage
**Position in pipeline:** not part of the scan pipeline at all — a standalone flow triggered by `alexiares update`, which replaces `signatures.path`'s contents

## Purpose

Fetches a signature repository archive and its detached signature, verifies the signature against a trusted Ed25519 public key, and — only on success — installs it. The package doc states the policy directly: *"There is no fallback to unsigned intelligence: a failed update is a hard error."*

```go
func Run(ctx context.Context, opts Options, destDir string) error
func Verify(data, sig []byte, pub ed25519.PublicKey) error
func Apply(archiveData []byte, destDir string) error
```

`Options{SourceURL, PublicKey, Timeout, UserAgent, MaxBytes}` — `SourceURL` is a plain base URL expected to serve two files, `signatures.tar.gz` and `signatures.tar.gz.sig`. See [`project_alexiares` memory / roadmaps.MD's Known Gaps] for why: this is meant to eventually point at a canonical, community-maintained GitHub repository, not this repo's own bundled `signatures/`.

## `Run`'s sequence

1. Refuses immediately, before any network call, if `PublicKey` is empty (`ErrNoPublicKey`) or `SourceURL` is empty.
2. Fetches both files with a bounded `http.Client` (size-limited via `io.LimitReader`, same pattern as `internal/collector`).
3. `Verify`s the signature. On failure, returns `"update rejected: ..."` and stops — `Apply` is never called.
4. Only on success, calls `Apply` to install the archive.

## `Verify`: the security-critical five lines

```go
func Verify(data, sig []byte, pub ed25519.PublicKey) error {
    if len(pub) != ed25519.PublicKeySize { return error }
    if len(sig) != ed25519.SignatureSize { return error }
    if !ed25519.Verify(pub, data, sig) { return error }
    return nil
}
```

Straight `crypto/ed25519`, no custom cryptography. Verified by test against four cases: a genuinely valid signature is accepted; the same signature over tampered data is rejected; a valid signature made with a *different* key than the one being checked against is rejected; and malformed (wrong-length) signatures or keys are rejected before ever reaching `ed25519.Verify`.

## `Apply`: treating the archive as hostile input

This is the part of the codebase that handles content from a remote source in a way that can write to the local filesystem — the same threat model as `internal/collector`, but with a much larger blast radius if it goes wrong (arbitrary file write, vs. collector's read-only fetch). Three separate defenses:

**Path traversal ("zip-slip") is rejected per-entry, not just discouraged.** `safeJoin` resolves each archive entry's name against the destination directory and confirms the *resolved* path is still inside it — rejecting both absolute paths (`/etc/evil.yaml`) and relative traversal (`../../etc/evil.yaml`) before any file is opened for writing. Verified by a test that crafts exactly such a malicious entry and confirms the file never lands outside the extraction directory, not just that `Apply` returns an error.

**Only regular files and directories are accepted.** Any other tar entry type (symlink, hardlink, device, FIFO) is rejected outright — `Apply` never follows or creates a symlink from archive content, closing off symlink-based extraction attacks entirely rather than trying to sanitize them.

**Every extracted file is size-bounded** (`maxExtractedFileBytes`, 5MB) via `io.LimitReader` reading one byte past the limit and checking the actual bytes written — guarding against a small compressed archive expanding into a huge file (a decompression bomb) at extraction time.

## `Apply`'s atomicity

Extraction happens into a fresh temp directory (`os.MkdirTemp`) first, entirely separate from `destDir`. Only after every entry extracts successfully does `Apply` swap it into place: rename the old `destDir` to `.bak`, rename the new directory into `destDir`, then remove the backup. If the final rename itself fails, the backup is renamed back — best-effort — so a failed update doesn't leave the signature repository *missing* on top of failing to update. A failure at any earlier step (bad archive, path traversal caught, oversized file) leaves the original `destDir` completely untouched, since the temp directory it was extracting into is simply discarded.

## Verified, not just unit-tested

Beyond the package's own offline `httptest`-based tests, the whole flow was run once against real signed content over real localhost HTTP during the build: a throwaway Ed25519 keypair, a real tar.gz signature file, served by `python3 -m http.server`, fetched and installed by the actual compiled `alexiares update` CLI command — confirming the mechanism works end-to-end through the real binary, not just through Go's test harness calling the package directly.
