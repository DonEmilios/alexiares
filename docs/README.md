# Module Documentation

One file per `internal/` package, in pipeline order. Each covers that module's purpose, public API, and the non-obvious design decisions and trade-offs behind it — the "why," not a restatement of the godoc (`go doc ./internal/<package>` already gives you that).

For the system-level architecture (the pipeline diagram, the CLI command list, the JSON output schema), see [the technical specification](<../Alexiares Technical Specification (specs.md).md>). For what's built vs. not yet, see [roadmaps.MD](../roadmaps.MD).

## Collection & extraction

| Module | What it does |
|---|---|
| [artifact](artifact.md) | Shared data types every other package communicates through — no logic, no dependencies |
| [collector](collector.md) | Classifies CLI input and safely fetches a target: HTML, TLS, scripts, favicon, redirects |
| [wallet](wallet.md) | Detects and classifies wallet addresses across 6 chains |
| [dns](dns.md) | Resolves the full DNS record set (A/AAAA/CNAME/MX/NS/TXT/PTR) |
| [tls](tls.md) | Parses certificate intelligence from a completed handshake |
| [html](html.md) | Extracts forms, metadata, comments, external resources from collected HTML |
| [javascript](javascript.md) | Hashes and pattern-matches inline + external scripts, never executing them |
| [favicon](favicon.md) | Computes SHA256 + Shodan-convention MurmurHash3 favicon fingerprints |
| [telegram](telegram.md) | Detects bot tokens, chat IDs, and Telegram links (exfiltration indicators) |
| [redirect](redirect.md) | Detects meta-refresh and JavaScript redirects (the two the collector can't see live) |

## Intelligence, matching & conclusion

| Module | What it does |
|---|---|
| [fingerprint](fingerprint.md) | Aggregates artifacts into one comparable identifier set — includes the SimHash fuzzy-matching design |
| [intel](intel.md) | The two-layer signature/observation model, schema validation, and the structural signature/observation separation |
| [correlation](correlation.md) | Matches a target's fingerprints against every loaded signature |
| [evidence](evidence.md) | Turns matches into strength-weighted evidence, a confidence verdict, and a recommendation |

## Presentation & distribution

| Module | What it does |
|---|---|
| [graph](graph.md) | Builds and serializes the infrastructure relationship graph (DOT/GraphML/JSON) |
| [output](output.md) | Renders a completed scan as terminal/JSON/CSV/Markdown/DOT/GraphML |
| [update](update.md) | Fetches, Ed25519-verifies, and atomically installs a signed signature archive |
| [config](config.md) | Loads and validates `~/.alexiares/config.yaml` |

`cmd/alexiares` itself isn't documented here — it's intentionally thin CLI wiring (argument parsing, calling into the packages above), not a module with its own design decisions to explain. See its own package-level tests (`cmd/alexiares/*_test.go`) for what each command does.
