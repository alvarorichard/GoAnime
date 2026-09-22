package superflix

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression suite for the failure reported on 2026-09-20 ("dexter"): every
// SuperFlix search answered 429 and the source never recovered on its own.
//
// The mechanism, from the user's debug log:
//
//	23:01:30  try=1   ← transport sleeps 10s on Retry-After
//	23:01:40  try=2   ← sleeps again
//	23:01:41  SuperFlix search timed out after 12s   ← dispatcher gives up
//	23:01:50  try=3   ← STILL REQUESTING, 9s after the search was abandoned
//	23:02:01  try=4
//	23:01:58  (next search starts while the previous one is still retrying)
//
// Two independent defects combined into a loop that fed itself:
//
//   - SuperFlixAdapter.SearchAnime called SearchMedia(query), which uses
//     context.Background(). The per-source deadline therefore never reached the
//     HTTP request, so abandoning a search stopped nothing.
//   - The transport honored Retry-After on requests that could not possibly
//     benefit from it, spending the search budget to arrive at the same 429 and
//     adding traffic to an endpoint that was rate limiting precisely because of
//     that traffic.
//
// Each test below pins one invariant that has to hold for the loop to stay
// broken. They use sfOrigin (ratelimit_origin_test.go), which reproduces the
// real abuse guard, and drive the real client and transport.

// ── Invariant 1: a search costs the origin exactly one request ───────────────

func TestSearch_CostsTheOriginExactlyOneRequest(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t)
	c := newOriginClient(o)

	_, err := c.SearchMediaWithContext(context.Background(), "dexter")
	require.NoError(t, err)

	assert.Equal(t, 1, o.hitCount(sfSearchPath),
		"a search must not be amplified into several requests; that is what exhausts the allowance")
}

// ── Invariant 2: a rate-limited search fails fast and does not amplify ───────

func TestSearch_RateLimitedOriginFailsFastWithoutAmplifying(t *testing.T) {
	t.Parallel()
	// limit 0 ⇒ every request is rejected, like an already-blocked endpoint.
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(10*time.Second))
	c := newOriginClient(o)

	ctx, cancel := context.WithTimeout(context.Background(), perSourceBudgetForTest)
	defer cancel()

	start := time.Now()
	_, err := c.SearchMediaWithContext(ctx, "dexter")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "429",
		"the caller must learn it was rate limited, not that something timed out")
	assert.Equal(t, 1, o.hitCount(sfSearchPath),
		"answering a 429 with more requests is what kept the block alive")
	assert.Less(t, elapsed, time.Second,
		"a 10s Retry-After inside a 12s budget can never pay off; waiting it out only burns the search")
}

// ── Invariant 3: an abandoned search stops talking to the origin ─────────────

// This is the defect itself. The dispatcher abandons a source when its deadline
// fires; if the request does not carry that deadline, the transport's retry
// loop keeps going unseen.
func TestSearch_AbandonedSearchStopsHittingTheOrigin(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(time.Second))
	c := newOriginClient(o)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := c.SearchMediaWithContext(ctx, "dexter")
	require.Error(t, err)
	abandoned := time.Now()

	// Any retry the old code would have made lands ~1s later (Retry-After: 1).
	// Fail the moment one arrives rather than counting after a grace period.
	o.failOnRequestAfter(t, sfSearchPath, "the search was already abandoned; its retries must stop with it")
	time.Sleep(2 * time.Second)

	assert.Zero(t, o.hitsAfter(sfSearchPath, abandoned),
		"an abandoned search must not keep requesting: that traffic is what refreshed the block")
}

// The other half of the same invariant, and the one that catches the adapter
// defect rather than the transport one: an IN-FLIGHT search must be torn down
// when the dispatcher walks away from it.
//
// The dispatcher does not wait for a slow source — it abandons the goroutine
// and moves on (searchOneWithTimeout). So this test abandons the call the same
// way production does, then checks the origin actually saw the client hang up.
// With the caller's context dropped on the floor (SearchMedia's
// context.Background(), the shape of the original bug) the request instead
// stays open until the client's own 30s ceiling, holding a connection on a host
// that is already refusing traffic.
func TestSearch_AbandonedSearchIsTornDownAtTheOrigin(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withHold())
	c := newOriginClient(o)

	const budget = 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()

	// Mirror searchOneWithTimeout: run the search, then stop waiting for it.
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = c.SearchMediaWithContext(ctx, "dexter")
	}()
	select {
	case <-done:
	case <-time.After(budget + 200*time.Millisecond):
		t.Fatal("the search outlived its deadline: the context never reached the request")
	}

	assert.True(t, o.awaitClientGone(2*time.Second),
		"the origin never saw the client disconnect — the abandoned search is still holding the connection")
}

// The same invariant on the transport itself, for a request that is NOT
// best-effort (the play path, which legitimately waits out a rate limit). Even
// there, the loop must die with the caller's context rather than outlive it.
func TestTransportRetryLoop_DiesWithTheCallersContext(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t, withLimit(0, time.Second), withRetryAfter(time.Second))
	c := newOriginClient(o)

	// No WithoutBrowserSolve: this is the play path, so the wait is allowed.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+sfSearchPath+"?s=dexter", http.NoBody)
	require.NoError(t, reqErr)
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err, "the cancel must abort the wait instead of being ignored")
	abandoned := time.Now()

	o.failOnRequestAfter(t, sfSearchPath, "the caller cancelled; the retry loop must not continue")
	time.Sleep(2 * time.Second)

	assert.Zero(t, o.hitsAfter(sfSearchPath, abandoned))
}

// ── Invariant 4: the block is allowed to lapse ──────────────────────────────

// The heart of the outage. Against an origin that extends the block whenever it
// receives traffic while blocked, recovery is only possible if the client goes
// quiet. A client that answers 429 with retries can never get out, which is why
// "dexter" failed identically search after search.
func TestSearch_BlockedOriginRecoversBecauseWeGoQuiet(t *testing.T) {
	t.Parallel()
	const window = 400 * time.Millisecond
	o := newSFOrigin(t, withLimit(1, window), withRetryAfter(time.Second), withSticky())
	c := newOriginClient(o)

	// Trip the limiter: first search passes, second is rejected.
	_, err := c.SearchMediaWithContext(context.Background(), "dexter")
	require.NoError(t, err)
	_, err = c.SearchMediaWithContext(context.Background(), "dexter two")
	require.Error(t, err, "the allowance is 1 per window; the second search must be blocked")

	before := o.hitCount(sfSearchPath)

	// Stay quiet for longer than the window. If anything is still retrying in
	// the background, the sticky origin restarts the block and the search below
	// fails — exactly the state the user was stuck in.
	time.Sleep(window + 200*time.Millisecond)

	assert.Equal(t, before, o.hitCount(sfSearchPath),
		"nothing may be talking to the origin while it is blocked, or the block never lapses")

	_, err = c.SearchMediaWithContext(context.Background(), "dexter three")
	assert.NoError(t, err, "after going quiet for one window the source must be usable again")
}

// ── Invariant 5: repeated searches stay inside the real allowance ────────────

// Twelve searches back to back against the production allowance (15 per 10s).
// At one request each this is comfortably inside it; at the old two-to-four
// requests per search it is not.
func TestSearch_ConsecutiveSearchesStayInsideTheRealAllowance(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t) // 15 per 10s, as measured live
	c := newOriginClient(o)

	for i := range 12 {
		// Distinct queries: SearchMediaWithContext caches per query, and a cache
		// hit would hide the traffic this test is about.
		_, err := c.SearchMediaWithContext(context.Background(), fmt.Sprintf("dexter %d", i))
		require.NoErrorf(t, err, "search %d was rate limited; a search must cost one request", i+1)
	}

	assert.Equal(t, 12, o.hitCount(sfSearchPath),
		"12 searches must cost 12 requests — any amplification eats the 15-per-10s allowance")
}

// ── Invariant 6: host discovery does not compete with search ────────────────

// Discovery probes `/`, the search hits `/pesquisar`, and the real abuse guard
// counts per endpoint. Pinning this stops a future change from routing a probe
// through the search endpoint, where it would spend the user's allowance.
func TestSearch_DoesNotShareItsEndpointWithProbes(t *testing.T) {
	t.Parallel()
	o := newSFOrigin(t)
	c := newOriginClient(o)

	_, err := c.SearchMediaWithContext(context.Background(), "dexter")
	require.NoError(t, err)

	assert.Equal(t, 1, o.hitCount(sfSearchPath))
	assert.Zero(t, o.hitCount("/"), "a search must not also fetch the homepage")
}

// perSourceBudgetForTest mirrors the dispatcher's per-source search budget.
// Kept local so this package does not depend on providers (which imports it).
const perSourceBudgetForTest = 12 * time.Second
