package netx

// Regression suite for the FlixHQ/SFlix/9Anime "misleading timeout diagnostic"
// follow-up.
//
// Discovered:  2026-04-28 — user-supplied debug log
//              ("===== GoAnime Debug Session — 2026-04-28 00:15:47 =====")
//              showed `Search source diagnostic details="FlixHQ temporarily
//              unavailable: search timed out after 12s"` for FlixHQ, SFlix
//              and 9Anime simultaneously. Manual probe with curl returned
//              `HTTP/2 522` with `cfOrigin;dur=0` — i.e. Cloudflare reached
//              its edge but could not reach the origin. The sites are dead,
//              not slow.
// Fixed:       2026-04-28 — same-day fix in this commit.
// Root cause:  `searchWithTimeout` (internal/scraper/unified.go) wrapped the
//              context-deadline expiry as a plain
//              `fmt.Errorf("search timed out after %v", ...)`. Downstream
//              `DiagnoseError` correctly classified that as
//              `DiagnosticSourceUnavailable` but had no StatusCode to attach,
//              so `UserMessage()` could only emit the generic
//              "temporarily unavailable: search timed out after 12s" line.
//              Operators looking at that log could not tell the difference
//              between (a) their network being slow, (b) the site being dead,
//              and (c) the GoAnime client being broken.
// Blast radius:diagnostic-only — search results were already correct (the
//              source was correctly skipped, SuperFlix etc. still returned
//              results). The bug was that the WARN line lied about the
//              cause, and users repeatedly opened issues asking us to
//              "fix flixhq" when the actual fix was on the upstream side.
//
// The fix adds a short post-timeout probe (HEAD request, 3s budget) against
// the source's base URL. If the upstream answers with a 5xx status — in
// particular Cloudflare 521-524 — the diagnostic is upgraded to carry that
// StatusCode and `UserMessage()` produces the much more actionable
// "Cloudflare 522/origin down" instead.
//
// The tests below pin the behavior:
//   1. A reachable upstream returning 522 ⇒ upgraded diagnostic with status.
//   2. A flat-out unreachable upstream ⇒ keep the original timeout message
//      (we never downgrade information).
//   3. A 200 OK upstream (slow but alive) ⇒ keep original (don't lie that
//      the site is "down" when it's just slow).
//   4. Empty base URL ⇒ no-op (sources with opaque APIs are unaffected).

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbeOriginStatus_ReachableUpstreamReturnsStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(522)
	}))
	defer srv.Close()

	got := ProbeOriginStatus(context.Background(), srv.URL, 2*time.Second)
	assert.Equal(t, 522, got, "probe must surface the upstream status code")
}

func TestProbeOriginStatus_DeadUpstreamReturnsZero(t *testing.T) {
	// A test server we shut down before probing — connection will fail.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	got := ProbeOriginStatus(context.Background(), url, 1*time.Second)
	assert.Equal(t, 0, got,
		"a connection-level failure must return 0 — callers rely on this to "+
			"avoid synthesising a fake upstream status")
}

func TestProbeOriginStatus_HonorsBudget(t *testing.T) {
	// Server that hangs longer than the probe budget — probe must give up
	// and return 0 rather than blocking the caller.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	start := time.Now()
	got := ProbeOriginStatus(context.Background(), srv.URL, 200*time.Millisecond)
	elapsed := time.Since(start)

	assert.Equal(t, 0, got, "probe must give up cleanly when budget elapses")
	assert.Less(t, elapsed, 1500*time.Millisecond,
		"probe must not exceed its budget by an unreasonable margin "+
			"(otherwise the post-timeout enrichment piles latency on the slow path)")
}

func TestProbeOriginStatus_RejectsEmptyURL(t *testing.T) {
	got := ProbeOriginStatus(context.Background(), "", 1*time.Second)
	assert.Equal(t, 0, got, "empty URL must short-circuit — no network I/O")
}

func TestEnrichTimeoutWithProbe_UpgradesCloudflare522(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(522)
	}))
	defer srv.Close()

	original := fmt.Errorf("search timed out after 12s")
	got := EnrichTimeoutWithProbe(context.Background(), "SFlix", "search",
		srv.URL, original, 2*time.Second)

	require.NotNil(t, got)
	assert.NotSame(t, original, got, "must return a richer error, not the original")

	var diag *SourceDiagnostic
	require.True(t, errors.As(got, &diag),
		"enriched error must be a *SourceDiagnostic so DiagnoseError can use the StatusCode")
	assert.Equal(t, 522, diag.StatusCode)
	assert.Equal(t, DiagnosticSourceUnavailable, diag.Kind)
	assert.True(t, errors.Is(got, ErrSourceUnavailable),
		"sentinel must remain reachable so circuit breaker / health checks keep working")
	assert.Contains(t, diag.UserMessage(), "Cloudflare 522/origin down",
		"upgraded diagnostic must produce the actionable user message")
}

func TestEnrichTimeoutWithProbe_KeepsOriginalWhenUpstreamUnreachable(t *testing.T) {
	// Upstream cannot be reached at all — probe returns 0. We must not
	// fabricate an HTTP status; keep the user's original timeout message.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	original := fmt.Errorf("search timed out after 12s")
	got := EnrichTimeoutWithProbe(context.Background(), "SFlix", "search",
		url, original, 500*time.Millisecond)

	assert.Same(t, original, got,
		"if the probe cannot reach the origin either, we must NOT invent a status — "+
			"return the original error unchanged so the existing 'temporarily unavailable: search timed out' message stands")
}

func TestEnrichTimeoutWithProbe_KeepsOriginalWhenUpstreamHealthy(t *testing.T) {
	// Site responded 200 — it is just slow, not dead. Don't slander it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	original := fmt.Errorf("search timed out after 12s")
	got := EnrichTimeoutWithProbe(context.Background(), "SFlix", "search",
		srv.URL, original, 1*time.Second)

	assert.Same(t, original, got,
		"a 200 OK upstream means slow, not down — we must NOT upgrade to "+
			"DiagnosticSourceUnavailable with a fake status")
}

func TestEnrichTimeoutWithProbe_KeepsOriginalWhenBaseURLEmpty(t *testing.T) {
	// Sources with opaque APIs (e.g. AllAnime GraphQL) have no probable
	// homepage. Empty base URL must short-circuit without doing any I/O.
	original := fmt.Errorf("search timed out after 12s")
	got := EnrichTimeoutWithProbe(context.Background(), "AllAnime", "search",
		"", original, 1*time.Second)

	assert.Same(t, original, got,
		"empty base URL must skip the probe entirely — sources with opaque "+
			"APIs are not regressions")
}

func TestEnrichTimeoutWithProbe_NilErrorIsNoop(t *testing.T) {
	got := EnrichTimeoutWithProbe(context.Background(), "SFlix", "search",
		"http://example.invalid", nil, 1*time.Second)
	assert.NoError(t, got, "must not synthesise an error from nil")
}
