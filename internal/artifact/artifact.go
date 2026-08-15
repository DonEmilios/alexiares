// Package artifact defines the data types shared across Alexiares'
// collector, extractors, fingerprint engine, and correlation engine.
//
// Extractors never import each other; they communicate exclusively
// through these types, keeping each extraction module independently
// testable and replaceable.
package artifact

import (
	"net/http"
	"time"
)

// Redirect records a single hop in a redirect chain, regardless of
// whether it originated from an HTTP status code, an HTML meta-refresh
// tag, or a JavaScript location assignment.
type Redirect struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Method string `json:"method"` // "http", "meta_refresh", "javascript"
}

// TLSData captures the certificate presented during collection.
type TLSData struct {
	SHA256    string    `json:"sha256"`
	SerialHex string    `json:"serial_hex"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
	DNSNames  []string  `json:"dns_names"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	KeyType   string    `json:"key_type"`
	KeyBits   int       `json:"key_bits"`
}

// RawResponse is the collector's output: everything safely observable
// about a target without executing any of its code.
type RawResponse struct {
	URL        string      `json:"url"`
	FinalURL   string      `json:"final_url"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	HTML       string      `json:"html"`
	Scripts    []string    `json:"scripts"` // raw script bodies (external + inline), in document order
	ScriptURLs []string    `json:"script_urls"`
	Favicon    []byte      `json:"favicon,omitempty"`
	TLS        *TLSData    `json:"tls,omitempty"`
	Redirects  []Redirect  `json:"redirects"`
	Timeline   Timeline    `json:"timeline"`
}

// DNSArtifacts is the DNS extractor's output.
type DNSArtifacts struct {
	IPs         []string `json:"ips"`
	AAAA        []string `json:"aaaa"`
	Nameservers []string `json:"nameservers"`
	MX          []string `json:"mx"`
	TXT         []string `json:"txt"`
	CNAME       string   `json:"cname,omitempty"`
	PTR         []string `json:"ptr,omitempty"`
	ASN         string   `json:"asn,omitempty"`
	ASNOrg      string   `json:"asn_org,omitempty"`
}

// HTMLArtifacts is the HTML extractor's output.
type HTMLArtifacts struct {
	Forms             []Form            `json:"forms"`
	Metadata          map[string]string `json:"metadata"`
	Comments          []string          `json:"comments"`
	ExternalResources []string          `json:"external_resources"`
	Title             string            `json:"title"`
}

// Form describes a single <form> element and its fields.
type Form struct {
	Action       string   `json:"action"`
	Method       string   `json:"method"`
	Fields       []string `json:"fields"`
	HiddenFields []string `json:"hidden_fields"`
}

// JavaScriptArtifacts is the JavaScript extractor's output.
type JavaScriptArtifacts struct {
	ScriptURLs        []string `json:"script_urls"`
	InlineScripts     []string `json:"inline_scripts"`
	SHA256            []string `json:"sha256"`
	APIEndpoints      []string `json:"api_endpoints"`
	TelegramRefs      []string `json:"telegram_refs"`
	DiscordWebhooks   []string `json:"discord_webhooks"`
	WalletLibraries   []string `json:"wallet_libraries"`
	WalletConnectRefs []string `json:"walletconnect_refs"`
}

// FaviconArtifacts is the favicon extractor's output.
type FaviconArtifacts struct {
	Murmur3 int32  `json:"murmur3"`
	SHA256  string `json:"sha256"`
	Size    int    `json:"size"`
}

// WalletChain identifies a supported blockchain for address detection.
type WalletChain string

// Supported wallet chains.
const (
	ChainEthereum WalletChain = "ethereum"
	ChainBitcoin  WalletChain = "bitcoin"
	ChainSolana   WalletChain = "solana"
	ChainCardano  WalletChain = "cardano"
	ChainTron     WalletChain = "tron"
	ChainTON      WalletChain = "ton"
)

// WalletAddress is a single detected address with its chain.
type WalletAddress struct {
	Chain   WalletChain `json:"chain"`
	Address string      `json:"address"`
}

// WalletArtifacts is the wallet extractor's output.
type WalletArtifacts struct {
	Addresses []WalletAddress `json:"addresses"`
	ENS       []string        `json:"ens,omitempty"`
}

// TelegramArtifacts is the Telegram extractor's output.
type TelegramArtifacts struct {
	BotTokens []string `json:"bot_tokens"`
	ChatIDs   []string `json:"chat_ids"`
	APIRefs   []string `json:"api_refs"`
	Links     []string `json:"links"`
}

// Fingerprints is the fingerprint engine's output: normalized,
// deterministic identifiers derived from a RawResponse plus its
// extracted artifacts.
type Fingerprints struct {
	Favicon     string   `json:"favicon,omitempty"`
	FaviconHash int32    `json:"favicon_murmur3,omitempty"`
	JavaScript  []string `json:"javascript,omitempty"`
	// HTML is a SHA256 hash of the DOM's tag structure (nesting shape
	// only — text, attributes, and comments are ignored), matching
	// only byte-identical page templates.
	HTML string `json:"html,omitempty"`
	// HTMLSimilarity is a 64-bit SimHash, hex-encoded, over shingled
	// tag sequences. Unlike HTML, it supports fuzzy matching: two
	// pages with a low Hamming distance between their SimHash values
	// share most of their template even if some sections differ.
	HTMLSimilarity string `json:"html_similarity,omitempty"`
	Certificate    string `json:"certificate,omitempty"`
}

// Timeline preserves when an artifact was first observed, reported,
// verified, and last seen, so observation time is never confused with
// activity time.
type Timeline struct {
	FirstSeen  time.Time `json:"first_seen,omitzero"`
	ReportedAt time.Time `json:"reported_at,omitzero"`
	VerifiedAt time.Time `json:"verified_at,omitzero"`
	LastSeen   time.Time `json:"last_seen,omitzero"`
}
