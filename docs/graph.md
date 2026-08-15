# `internal/graph`

**Source:** [`internal/graph/`](../internal/graph/) (`graph.go`, `builder.go`, `scan.go`, `dot.go`, `graphml.go`, `json.go`)
**Tests:** all five files have matching `_test.go` — 85.8% coverage
**Position in pipeline:** fifth stage — consumes `internal/correlation`'s output alongside the raw scan data; feeds `internal/output`

## Purpose

Represents infrastructure relationships as a typed node/edge graph and serializes it three ways (DOT, GraphML, JSON).

```go
type Graph struct { Nodes []Node; Edges []Edge }
func NewBuilder() *Builder                          // dedupes nodes/edges by (id) / (from,to,type)
func FromScan(data ScanData) Graph                   // the actual graph-construction entry point
func WriteDOT(g Graph) string
func WriteGraphML(g Graph) (string, error)
func WriteJSON(g Graph) (string, error)
```

Node types and edge types are both closed enums taken directly from the spec's vocabulary (11 node types: Domain, URL, IP, ASN, Certificate, Favicon, JavaScript, Wallet, Telegram, Registrar, Nameserver; 9 edge types: resolves_to, hosted_by, uses_certificate, shares_favicon, reuses_script, contains_wallet, redirects_to, registered_with, shares_nameserver).

## How `FromScan` builds the graph

Two passes, not one:

1. **The target's own star.** One domain node, with an edge out to a node for each of its IPs (`resolves_to`), nameservers (`shares_nameserver`), certificate (`uses_certificate`), favicon (`shares_favicon`), each JS hash (`reuses_script`), each detected wallet (`contains_wallet`), and each redirect destination (`redirects_to`). This part needs nothing from correlation — it's built from directly observed data alone, so a scan with zero loaded signatures still produces a meaningful (if star-shaped) graph.
2. **Cluster expansion**, only if `ScanData.Correlation` carries matched clusters. For each match in each cluster, `hubFor` resolves which node in the graph that match actually refers to, then every one of that cluster's `RelatedDomains` gets its own node with an edge *into the same hub node* — not into the target. That's what turns the star into an actual cluster picture: viewers see "sibling.example also shares this favicon," not just "the target has a favicon."

## Design notes: the favicon/certificate identity problem

**A favicon match can happen via two different sub-checks (SHA256 or MurmurHash3), and `hubFor` deliberately doesn't use `match.Value` to build the hub node ID for it.** A MurmurHash3 match's `Value` is a decimal integer string; a SHA256 match's `Value` is a hex hash. If the hub node ID were built directly from `match.Value`, the two sub-checks would produce two *different* node IDs for what is, physically, the same favicon — fragmenting one entity into two disconnected graph nodes depending on which check happened to fire. `hubFor` sidesteps this by ignoring `match.Value` entirely for the favicon case and anchoring on `fp.Favicon` instead (the target's own SHA256, already the canonical node built in pass 1) — since a favicon match, by construction, only ever fires because it equals the target's own fingerprint value anyway. Certificate matches don't have this ambiguity (only one hash type), but use the same anchoring pattern for consistency.

**Wallet node IDs deliberately drop the chain prefix** (`"wallet:" + address`, not `"wallet:" + chain + ":" + address`) — a simplification, not an oversight. `artifact.Match` doesn't carry a `Chain` field (only `SignatureID`, `Category`, `Value`), so building a chain-qualified ID from a `Match` alone isn't possible without extending that type. Real address collisions across chains are rare enough in practice that this was an accepted trade-off rather than a reason to widen `Match`.

**Node labels are truncated for long hashes** (`shortHash` — 12 chars + `…` past that) purely for rendered readability; the full value lives in the node's `ID`, so nothing is lost, just not shown at full length in a Graphviz/Gephi label.

## The three serializers

**DOT** (`dot.go`) — hand-built string, but every node/edge label goes through `dotQuote` (escapes `\` and `"`) rather than raw interpolation. Verified against a *real* Graphviz install during the build (`dot -Tsvg`, not just "does the string look right"), confirmed to render.

**GraphML** (`graphml.go`) — modeled as Go structs with `encoding/xml` tags and marshaled, specifically so escaping is handled by the standard library rather than hand-built XML string concatenation, which the source comment calls out directly as "a routine source of malformed or injectable output when a label contains `&`, `<`, or `>`." A test confirms a label containing literal `<script>` markup still parses as valid, well-escaped XML.

**JSON** (`json.go`) — a thin wrapper around `json.MarshalIndent`; exists purely so callers pick DOT/GraphML/JSON through one consistent function-per-format shape rather than JSON being special-cased. Round-trip losslessness (marshal → unmarshal → deep-equal) is covered by a test.

## Known limitations

- **ASN, Registrar, and Telegram node types are defined but never emitted** by `FromScan` — no extractor produces ASN or registrar data (see [`dns.md`](dns.md)), and the spec defines a Telegram node type with no corresponding edge verb, so wiring it in would mean inventing vocabulary the spec doesn't have.
- **`hosted_by` and `registered_with` edge types are defined but unused**, for the same reason (no ASN/registrar data to hang them on).
