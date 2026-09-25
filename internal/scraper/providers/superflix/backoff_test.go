package superflix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression suite for "está queimando IP muito rápido" (2026-09-21).
//
// Honoring Retry-After inside a single request was not enough. The user typed
// three searches in nineteen seconds:
//
//	00:43:02  tehran  → 429
//	00:43:07  teran   → 429
//	00:43:18  theran  → 429
//
// Each one sent a request into a block the server had just asked us to wait
// out, and the abuse guard is sticky — traffic during a block refreshes it — so
// the retyping itself kept the IP locked out. Restarting `goanime` made it
// worse, because nothing remembered the back-off across processes.
//
// rateLimitGate is what stops that: one 429 silences the host for the advertised
// delay, the state is on disk, and consecutive strikes back off further.

// ── The reported scenario ───────────────────────────────────────────────────

// Three searches in a row against a blocked origin must cost ONE request. The
// second and third are the ones that were burning the IP.
func TestRateLimitGate_RetypingASearchDoesNotKeepHittingABlockedHost(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(10*time.Second))
	c := newOriginClient(o)

	for _, q := range []string{"tehran", "teran", "theran"} {
		_, err := c.SearchMediaWithContext(context.Background(), q)
		require.Errorf(t, err, "the origin is blocked; %q cannot succeed", q)
	}

	assert.Equal(t, 1, o.hitCount(sfSearchPath),
		"only the first search may reach a blocked host; the rest are what refresh the block")
}

// And the user has to be told it was a back-off, not a mystery failure: the
// second search never touched the network, so reporting it as an upstream error
// would be a lie.
func TestRateLimitGate_SuppressedRequestSaysSoAndSendsNothing(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(10*time.Second))
	c := newOriginClient(o)

	_, first := c.SearchMediaWithContext(context.Background(), "tehran")
	require.Error(t, first)
	assert.NotErrorIs(t, first, ErrRateLimited,
		"the first search did reach the origin; it must report the real 429")

	_, second := c.SearchMediaWithContext(context.Background(), "teran")
	require.Error(t, second)
	assert.ErrorIs(t, second, ErrRateLimited)
	assert.Contains(t, second.Error(), "no request sent")
	assert.Equal(t, 1, o.hitCount(sfSearchPath))
}

// Once the back-off lapses the source must come back on its own, without the
// user restarting anything.
func TestRateLimitGate_ReleasesTheHostWhenTheBackOffLapses(t *testing.T) {
	t.Parallel()
	gate := newTestGate("") // bounded to 100ms
	o := newSFOrigin(t, withLimit(1, 50*time.Millisecond), withRetryAfter(10*time.Second))
	c := newOriginClientWithGate(o, gate)

	_, err := c.SearchMediaWithContext(context.Background(), "tehran")
	require.NoError(t, err, "the first request is inside the allowance")
	_, err = c.SearchMediaWithContext(context.Background(), "teran")
	require.Error(t, err, "the second trips the limiter")

	time.Sleep(150 * time.Millisecond) // > the gate's max bound

	_, err = c.SearchMediaWithContext(context.Background(), "theran")
	assert.NoError(t, err, "the gate must reopen on its own once the back-off has been served")
	assert.Equal(t, 3, o.hitCount(sfSearchPath))
}

// ── Surviving a restart ─────────────────────────────────────────────────────

// The user runs `go run ./cmd/goanime` again after a failure. An in-memory
// back-off — which is all the netx circuit breaker offers — is gone by then, so
// the fresh process walks straight back into the block.
func TestRateLimitGate_SurvivesAProcessRestart(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), backoffStateFileName)

	first := newRateLimitGate(path)
	wait := first.arm("superflixapi.quest", sfSearchPath, 10*time.Second)
	require.Positive(t, wait)

	// A new process: same state file, no shared memory.
	restarted := newRateLimitGate(path)
	assert.Positive(t, restarted.retryIn("superflixapi.quest", sfSearchPath),
		"a restart must not hand the user a fresh allowance to burn")
}

func TestRateLimitGate_PersistedStateThatLapsedIsIgnored(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), backoffStateFileName)
	writeBackoffState(t, path, map[string]hostBackoff{
		"superflixapi.quest" + sfSearchPath: {Until: time.Now().Add(-time.Minute), Strikes: 3},
	})

	g := newRateLimitGate(path)
	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath), "an expired back-off must not gate anything")
}

// A crash during a long back-off must not gate a host forever.
func TestRateLimitGate_PersistedStateThatIsAbsurdlyFarOutIsIgnored(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), backoffStateFileName)
	writeBackoffState(t, path, map[string]hostBackoff{
		"superflixapi.quest" + sfSearchPath: {Until: time.Now().Add(24 * time.Hour), Strikes: 9},
	})

	g := newRateLimitGate(path)
	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath),
		"state older than the escalation cap cannot be ours; it must not lock the source out")
}

func TestRateLimitGate_CorruptStateIsIgnored(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), backoffStateFileName)
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))

	g := newRateLimitGate(path)
	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath), "unreadable state must fail open, not lock the source out")
}

// ── Back-off arithmetic ─────────────────────────────────────────────────────

// The server's own Retry-After is the wait, not a constant of ours: guessing
// shorter is what provokes the next 429.
func TestRateLimitGate_HonorsTheAdvertisedRetryAfter(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")

	wait := g.arm("superflixapi.quest", sfSearchPath, 42*time.Second)
	assert.Equal(t, 42*time.Second, wait)
}

// A missing or unusable Retry-After still earns a real pause.
func TestRateLimitGate_FallsBackToAFloorWhenNoDelayIsAdvertised(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")

	assert.Equal(t, minBackoff, g.arm("superflixapi.quest", sfSearchPath, 0))
}

// A host that keeps refusing is saying its advertised delay was not enough.
func TestRateLimitGate_EscalatesOnConsecutiveStrikes(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")
	const advertised = 10 * time.Second

	first := g.arm("superflixapi.quest", sfSearchPath, advertised)
	second := g.arm("superflixapi.quest", sfSearchPath, advertised)
	third := g.arm("superflixapi.quest", sfSearchPath, advertised)

	assert.Equal(t, advertised, first)
	assert.Greater(t, second, first, "a second strike must buy more quiet than the first")
	assert.Greater(t, third, second)
	assert.LessOrEqual(t, third, maxBackoff)
}

func TestRateLimitGate_EscalationIsCapped(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")

	var wait time.Duration
	for range 20 {
		wait = g.arm("superflixapi.quest", sfSearchPath, time.Minute)
	}
	assert.Equal(t, maxBackoff, wait,
		"the cap keeps a blocked source recovering on its own instead of being shelved for hours")
}

// Any response that is not a 429 means the host is talking to us again, so the
// strike count must not carry into an unrelated hiccup later.
func TestRateLimitGate_SuccessClearsTheStrikeCount(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")
	const advertised = 10 * time.Second

	g.arm("superflixapi.quest", sfSearchPath, advertised)
	g.arm("superflixapi.quest", sfSearchPath, advertised)
	g.clear("superflixapi.quest", sfSearchPath)

	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath))
	assert.Equal(t, advertised, g.arm("superflixapi.quest", sfSearchPath, advertised),
		"after a success the next 429 starts from the advertised delay again")
}

// ── Blast radius ────────────────────────────────────────────────────────────

// The SuperFlix client also talks to rotating player hosts and CDNs. A rate
// limit on one of those must not silence search, and vice versa.
func TestRateLimitGate_IsPerHost(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")

	g.arm("cdn.example.invalid", "/media.m3u8", 30*time.Second)

	assert.Positive(t, g.retryIn("cdn.example.invalid", "/media.m3u8"))
	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath),
		"one endpoint backing off must not take the others down with it")
}

func TestRateLimitGate_EmptyHostIsANoOp(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")

	assert.Zero(t, g.arm("", sfSearchPath, time.Minute))
	assert.Zero(t, g.retryIn("", sfSearchPath))
	g.clear("", sfSearchPath) // must not panic
}

// A nil gate is the zero value a future caller might reach for; it must be
// inert rather than a nil dereference on the request path.
func TestRateLimitGate_NilIsInert(t *testing.T) {
	t.Parallel()
	var g *rateLimitGate

	assert.Zero(t, g.retryIn("superflixapi.quest", sfSearchPath))
	assert.Zero(t, g.arm("superflixapi.quest", sfSearchPath, time.Minute))
	g.clear("superflixapi.quest", sfSearchPath)
}

// The gate is shared across clients on purpose: two SuperFlixClients in one
// process must not each spend a request on a host that is already blocked.
func TestRateLimitGate_IsSharedBetweenClients(t *testing.T) {
	t.Parallel()
	gate := newTestGate("")
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(10*time.Second))

	first := newOriginClientWithGate(o, gate)
	second := newOriginClientWithGate(o, gate)

	_, err := first.SearchMediaWithContext(context.Background(), "tehran")
	require.Error(t, err)

	_, err = second.SearchMediaWithContext(context.Background(), "teran")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
	assert.Equal(t, 1, o.hitCount(sfSearchPath),
		"a second client must inherit the back-off rather than re-discover it the expensive way")
}

// The production gate must never write into the developer's cache from a test
// run, where a stray back-off would then suppress their next real search.
func TestRateLimitGate_TestBinaryDoesNotPersistToTheUserCache(t *testing.T) {
	t.Parallel()
	assert.Empty(t, defaultBackoffStatePath(),
		"a test binary must keep its back-off in memory only")
}

// ── Not gating what must not be gated ───────────────────────────────────────

// The wait inside one request is how a rate limit is actually served on the
// play path. Arming the gate must not cut that short by blocking the retry the
// transport is already waiting for.
func TestRateLimitGate_DoesNotBlockTheRetryInsideAnInFlightRequest(t *testing.T) {
	t.Parallel()
	gate := newTestGate("")
	// First request 429s with a short Retry-After, the retry succeeds.
	o := newSFOrigin(t, withLimit(1, time.Hour), withRetryAfter(time.Second))
	c := newOriginClientWithGate(o, gate)

	// Burn the allowance so the next request is the one that gets 429'd.
	_, err := c.SearchMediaWithContext(context.Background(), "warmup")
	require.NoError(t, err)

	// Play path (no WithoutBrowserSolve): the transport is allowed to wait.
	req, reqErr := newPlayPathRequest(c, "tehran")
	require.NoError(t, reqErr)

	resp, doErr := c.client.Do(req)
	require.NoError(t, doErr, "the in-flight retry must not be suppressed by the entry it just created")
	defer func() { _ = resp.Body.Close() }()

	assert.GreaterOrEqual(t, o.hitCount(sfSearchPath), 2,
		"the transport must have been allowed to retry after serving the advertised wait")
}

func writeBackoffState(t *testing.T, path string, state map[string]hostBackoff) {
	t.Helper()
	blob, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, blob, 0o600))
}

// errRateLimitedIs is a compile-time reminder that callers identify a
// suppressed request through errors.Is, not by matching message text.
var _ = func() bool { return errors.Is(rateLimitedError("h", time.Second), ErrRateLimited) }()

// ── The header pairing ──────────────────────────────────────────────────────

// The request must present ONE coherent browser: the Accept-Language ladder has
// to be the one the browser named in the User-Agent actually sends.
//
// The pair to avoid is not a fixed pair. This test used to forbid Chrome's
// ladder outright, because the client claimed Firefox and sending Chrome's
// ladder under it was the mismatch the host 429'd for a stretch on 2026-09-21.
// The User-Agent has since moved to Chrome — the solver drives Chromium, and
// the Firefox string it used to send is itself throttled now — so Chrome's
// ladder is the CORRECT one and the Firefox ladder became the mismatch.
//
// Hence the rule, not the value: whatever the UA says, the ladder must agree.
// The host's CDN also matches Accept-Language by value (cdn.go), and the mock
// origin rejects an incoherent pair, so every search in this file guards it.
func TestDecorateRequest_SendsTheLadderMatchingTheUserAgent(t *testing.T) {
	t.Parallel()

	c := NewSuperFlixClient()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/", http.NoBody)
	require.NoError(t, err)
	c.decorateRequest(req)

	// Asserted through the same coherence rule the mock origin applies, rather
	// than against one named constant: a browser sends a different ladder per
	// locale, and this client's is pinned to the English one because the host
	// refuses the Portuguese Chrome ladder (client.go).
	assert.True(t, headersDescribeOneBrowser(req.Header),
		"the q-ladder %q does not describe the browser %q claims",
		req.Header.Get("Accept-Language"), req.Header.Get("User-Agent"))
	assert.Equal(t, superFlixAcceptLanguage, req.Header.Get("Accept-Language"))
}

// The ladder must match the User-Agent we claim, whichever browser that is.
//
// Pinned as a rule rather than a value: the UA moved from Firefox to Chrome when
// the Firefox string turned out to be throttled by this host, and a test that
// named one ladder would have had to be rewritten to keep passing instead of
// catching anything.
func TestSuperFlixAcceptLanguage_MatchesTheClaimedBrowser(t *testing.T) {
	t.Parallel()
	require.True(t,
		strings.Contains(SuperFlixUserAgent, "Chrome/") || strings.Contains(SuperFlixUserAgent, "Firefox/"),
		"the UA names a browser this test does not know a ladder for")
	assert.True(t, headersDescribeOneBrowser(http.Header{
		"User-Agent":      []string{SuperFlixUserAgent},
		"Accept-Language": []string{superFlixAcceptLanguage},
	}), "the q-ladder does not describe the browser the User-Agent claims")
}

// A search through the mock is refused outright when the bad pair comes back,
// which is what makes the whole suite guard this and not just the two tests above.
func TestSearch_IsRefusedWhenTheMismatchedHeaderComesBack(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t) // generous allowance: only the header can fail this
	c := newOriginClient(o)
	c.userAgent = SuperFlixUserAgent

	_, err := c.SearchMediaWithContext(context.Background(), "tehran")
	require.NoError(t, err, "the current header must be accepted")

	// Now prove the mock would catch a regression.
	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet,
		c.base()+sfSearchPath+"?s=tehran", http.NoBody)
	require.NoError(t, reqErr)
	// Whichever ladder does NOT go with the current UA is the regression.
	wrong := netx.ChromeAcceptLanguage
	if strings.Contains(SuperFlixUserAgent, "Chrome/") {
		wrong = netx.AcceptLanguage
	}
	req.Header.Set("User-Agent", SuperFlixUserAgent)
	req.Header.Set("Accept-Language", wrong)
	resp, doErr := http.DefaultClient.Do(req)
	require.NoError(t, doErr)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode,
		"the mock must reject the mismatched pair, or it cannot guard anything")
}

// ── Blast radius, per endpoint ──────────────────────────────────────────────

// SuperFlix's abuse guard counts per endpoint: measured 2026-09-21, `/` kept
// answering 200 while `/pesquisar` was still 429ing from the same IP, neither
// response cached. Keying the back-off by host alone would therefore silence
// playback, the episode listing and the metadata lookup because a SEARCH was
// blocked — taking away everything that still worked.
func TestRateLimitGate_ABlockedSearchDoesNotSilenceTheRestOfTheHost(t *testing.T) {
	t.Parallel()
	g := newRateLimitGate("")
	const host = "superflixapi.quest"

	g.arm(host, "/pesquisar", 10*time.Second)

	assert.Positive(t, g.retryIn(host, "/pesquisar"))
	assert.Zero(t, g.retryIn(host, "/serie/30984"),
		"the episode page was never rate limited; playback must stay reachable")
	assert.Zero(t, g.retryIn(host, "/player/bootstrap"))
	assert.Zero(t, g.retryIn(host, "/"),
		"host discovery probes `/`, which the guard counts separately")
}

// The query string is not part of the key: the limit is on the endpoint, so
// retyping the search under a different term must not buy a fresh allowance.
// That retyping is exactly what the user did.
func TestRateLimitGate_KeyIgnoresTheQueryString(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(10*time.Second))
	c := newOriginClient(o)

	_, err := c.SearchMediaWithContext(context.Background(), "kill bill")
	require.Error(t, err)
	_, err = c.SearchMediaWithContext(context.Background(), "kill the bill")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited,
		"a different search term is the same endpoint; it must not slip past the back-off")
	assert.Equal(t, 1, o.hitCount(sfSearchPath))
}

func TestBackoffKey_NormalizesAndRefusesEmptyHosts(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "superflixapi.quest/pesquisar", backoffKey("superflixapi.quest", "/pesquisar"))
	assert.Equal(t, "superflixapi.quest/", backoffKey("superflixapi.quest", ""),
		"a path-less URL is the root, not a second key for the same thing")
	assert.Empty(t, backoffKey("", "/pesquisar"), "without a host there is nothing to gate")
}
