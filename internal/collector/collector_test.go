package collector_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexiares/alexiares/internal/collector"
)

func testCollector(t *testing.T) *collector.Collector {
	t.Helper()
	return collector.New(collector.Options{
		Timeout:              2 * time.Second,
		UserAgent:            "alexiares-test",
		MaxRedirects:         3,
		MaxBodyBytes:         1 << 20,
		MaxScripts:           10,
		AllowPrivateNetworks: true, // tests target httptest servers on 127.0.0.1
	})
}

func TestCollectBasicResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head><title>hi</title></head><body></body></html>"))
	}))
	defer srv.Close()

	c := testCollector(t)
	raw, err := c.Collect(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if raw.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", raw.StatusCode)
	}
	if !strings.Contains(raw.HTML, "<title>hi</title>") {
		t.Errorf("HTML = %q, want to contain title", raw.HTML)
	}
	if raw.Headers.Get("X-Test") != "yes" {
		t.Errorf("Headers[X-Test] = %q, want yes", raw.Headers.Get("X-Test"))
	}
	if raw.Timeline.FirstSeen.IsZero() {
		t.Error("Timeline.FirstSeen is zero, want set")
	}
}

func TestCollectTracksRedirectChain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/end", http.StatusFound)
	})
	mux.HandleFunc("/end", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testCollector(t)
	raw, err := c.Collect(context.Background(), srv.URL+"/start")
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if len(raw.Redirects) != 1 {
		t.Fatalf("Redirects = %+v, want 1 hop", raw.Redirects)
	}
	if raw.Redirects[0].Method != "http" {
		t.Errorf("Redirects[0].Method = %q, want http", raw.Redirects[0].Method)
	}
	if !strings.HasSuffix(raw.Redirects[0].From, "/start") {
		t.Errorf("Redirects[0].From = %q, want suffix /start", raw.Redirects[0].From)
	}
	if !strings.HasSuffix(raw.Redirects[0].To, "/end") {
		t.Errorf("Redirects[0].To = %q, want suffix /end", raw.Redirects[0].To)
	}
	if !strings.HasSuffix(raw.FinalURL, "/end") {
		t.Errorf("FinalURL = %q, want suffix /end", raw.FinalURL)
	}
}

func TestCollectCapsRedirectChain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop?n="+r.URL.Query().Get("n")+"x", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := collector.New(collector.Options{MaxRedirects: 2, Timeout: 2 * time.Second, AllowPrivateNetworks: true})
	raw, err := c.Collect(context.Background(), srv.URL+"/loop")
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(raw.Redirects) != 2 {
		t.Errorf("Redirects = %d hops, want capped at 2", len(raw.Redirects))
	}
}

func TestCollectDiscoversScriptsAndFavicon(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head>
			<script src="/app.js"></script>
			<link rel="icon" href="/favicon.png">
		</head><body></body></html>`))
	})
	mux.HandleFunc("/app.js", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("console.log('drainer')"))
	})
	mux.HandleFunc("/favicon.png", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testCollector(t)
	raw, err := c.Collect(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if len(raw.Scripts) != 1 || raw.Scripts[0] != "console.log('drainer')" {
		t.Errorf("Scripts = %v, want [console.log('drainer')]", raw.Scripts)
	}
	if len(raw.ScriptURLs) != 1 || !strings.HasSuffix(raw.ScriptURLs[0], "/app.js") {
		t.Errorf("ScriptURLs = %v, want suffix /app.js", raw.ScriptURLs)
	}
	if string(raw.Favicon) != "\x89PNG" {
		t.Errorf("Favicon = %v, want PNG magic bytes", raw.Favicon)
	}
}

func TestCollectBoundsBodySize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 1<<20))) // 1 MB
	}))
	defer srv.Close()

	c := collector.New(collector.Options{MaxBodyBytes: 1024, Timeout: 2 * time.Second, AllowPrivateNetworks: true})
	raw, err := c.Collect(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(raw.HTML) != 1024 {
		t.Errorf("len(HTML) = %d, want bounded to 1024", len(raw.HTML))
	}
}

func TestCollectInvalidTargetErrors(t *testing.T) {
	c := testCollector(t)
	if _, err := c.Collect(context.Background(), "http://127.0.0.1:0/unreachable"); err == nil {
		t.Error("Collect() error = nil, want error for unreachable target")
	}
}

// TestCollectRejectsPrivateTargetByDefault is a regression test for
// the collector's SSRF protection: Alexiares scans attacker-supplied
// targets, so a Collector built without opting into
// AllowPrivateNetworks must never connect to a loopback/private
// address. The same DialContext backs every hop of a request
// (including redirects), so blocking it here also proves a redirect
// to a private address would be blocked identically.
func TestCollectRejectsPrivateTargetByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := collector.New(collector.Options{Timeout: 2 * time.Second}) // AllowPrivateNetworks defaults false
	if _, err := c.Collect(context.Background(), srv.URL); err == nil {
		t.Fatal("Collect() against a loopback target with AllowPrivateNetworks unset: error = nil, want rejection")
	} else if !strings.Contains(err.Error(), "non-public") {
		t.Errorf("error = %q, want it to explain the connection was refused as non-public", err)
	}
}
