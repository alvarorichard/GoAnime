package providers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The original defect lived exactly here: superFlixProvider.Search called
// adapter.SearchAnime(query), the context-free method, so the per-source
// deadline never reached the HTTP request and an abandoned search kept hammering
// a rate-limited endpoint ("dexter", 2026-09-20).
//
// scraper.ContextualScraper existing is not enough — the provider has to USE it.
// A revert of either half reintroduces the bug, so both are pinned.

// recordingAdapter implements UnifiedScraper and ContextualScraper, recording
// which of the two entry points the caller chose and what context it passed.
type recordingAdapter struct {
	mu           sync.Mutex
	contextCalls atomic.Int32
	contextFree  atomic.Int32
	// gotMarker records whether the context that reached the adapter is the one
	// the caller passed in, identified by a value only that context carries.
	gotMarker bool
}

// callerMarkerKey tags the caller's context. A provider that substitutes
// context.Background() — the original defect — loses it, while still calling
// the right method, so this is what distinguishes the two.
type callerMarkerKey struct{}

func (a *recordingAdapter) SearchAnime(string, ...any) ([]*models.Anime, error) {
	a.contextFree.Add(1)
	return []*models.Anime{{Name: "Dexter"}}, nil
}

func (a *recordingAdapter) SearchAnimeContext(ctx context.Context, _ string, _ ...any) ([]*models.Anime, error) {
	a.contextCalls.Add(1)
	a.mu.Lock()
	a.gotMarker, _ = ctx.Value(callerMarkerKey{}).(bool)
	a.mu.Unlock()
	return []*models.Anime{{Name: "Dexter"}}, nil
}

func (a *recordingAdapter) GetAnimeEpisodes(string) ([]models.Episode, error) { return nil, nil }
func (a *recordingAdapter) GetAnimeEpisodesContext(context.Context, string) ([]models.Episode, error) {
	return nil, nil
}

func (a *recordingAdapter) GetStreamURL(string, ...any) (streamURL string, metadata map[string]string, err error) {
	return "", nil, nil
}

func (a *recordingAdapter) GetStreamURLContext(context.Context, string, ...any) (streamURL string, metadata map[string]string, err error) {
	return "", nil, nil
}
func (a *recordingAdapter) GetType() scraper.ScraperType { return scraper.SuperFlixType }

// providerWithAdapter builds a superFlixProvider whose lazy adapter slot is
// already filled, so the provider's own dispatch can be observed without
// touching the network.
func providerWithAdapter(a scraper.UnifiedScraper) *superFlixProvider {
	p := &superFlixProvider{}
	p.once.Do(func() {}) // consume the Once so scraper() returns the slot below
	p.adapter.UnifiedScraper = a
	return p
}

func TestSuperFlixProvider_SearchUsesTheContextualEntryPoint(t *testing.T) {
	t.Parallel()
	rec := &recordingAdapter{}
	p := providerWithAdapter(rec)

	got, err := p.Search(context.Background(), "dexter")
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, int32(1), rec.contextCalls.Load(),
		"the provider must call SearchAnimeContext, or the per-source deadline never reaches the request")
	assert.Zero(t, rec.contextFree.Load(),
		"the context-free SearchAnime is the shape of the original bug and must not be used here")
}

// Calling the contextual method is not enough on its own: handing it a fresh
// context.Background() would satisfy the assertion above while reintroducing
// the defect verbatim, because the per-source deadline still would not reach
// the request. The context that arrives at the adapter must be the caller's.
//
// It is identified by a value only the caller's context carries, so the check
// holds for a live context — a cancelled one never reaches the adapter, since
// Search refuses it at the top.
func TestSuperFlixProvider_PassesTheCallersContextDown(t *testing.T) {
	t.Parallel()
	rec := &recordingAdapter{}
	p := providerWithAdapter(rec)

	ctx := context.WithValue(context.Background(), callerMarkerKey{}, true)
	_, err := p.Search(ctx, "dexter")
	require.NoError(t, err)

	require.Equal(t, int32(1), rec.contextCalls.Load())
	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.True(t, rec.gotMarker,
		"the adapter got a context the caller never created — a substituted context carries no deadline "+
			"and cannot abort the request, which is exactly the original bug")
}

// A source whose adapter cannot be cancelled leaks work on every abandoned
// search. This is a standing check, not a SuperFlix one: adding a source that
// silently drops the context is how the same outage comes back on another host.
//
// AnimeFire and Goyabu are knowingly non-contextual today — their clients build
// requests without one. They are listed explicitly so the list is a decision,
// not an oversight, and so promoting one of them makes this test go green
// rather than red.
func TestSearchableSources_CancellationCapabilityIsDeclaredOnPurpose(t *testing.T) {
	t.Parallel()
	knownNonCancelable := map[source.SourceKind]string{
		source.AnimeFire: "animefire client builds requests without a context",
		source.Goyabu:    "goyabu client builds requests without a context",
	}

	for _, kind := range []source.SourceKind{source.AnimeFire, source.Goyabu, source.SuperFlix, source.HiAnime} {
		st, ok := source.ScraperTypeFor(kind)
		require.Truef(t, ok, "%s has no scraper type", kind)

		adapter, err := scraper.NewAdapter(st)
		require.NoErrorf(t, err, "%s has no adapter", kind)

		_, contextual := adapter.(scraper.ContextualScraper)
		if why, expected := knownNonCancelable[kind]; expected {
			assert.Falsef(t, contextual,
				"%s now implements ContextualScraper (%s no longer applies) — remove it from the list "+
					"and make its provider use the contextual path", kind, why)
			continue
		}
		assert.Truef(t, contextual,
			"%s must implement ContextualScraper: without it a per-source deadline cannot abort its "+
				"in-flight request, and an abandoned search keeps hitting the upstream host", kind)
	}
}

// Every registered source must refuse an already-dead context before doing any
// work. It is the cheapest half of cancellation support and the one a new
// source is most likely to forget.
func TestSearchableSources_RefuseAnAlreadyCancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, kind := range []source.SourceKind{source.AnimeFire, source.Goyabu, source.SuperFlix, source.HiAnime} {
		s, ok := source.Registered(kind)
		require.Truef(t, ok, "%s is not registered", kind)
		sr, ok := s.(source.Searchable)
		require.Truef(t, ok, "%s is not searchable", kind)

		_, err := sr.Search(ctx, "dexter")
		assert.ErrorIsf(t, err, context.Canceled,
			"%s must return the context error instead of starting a search nobody is waiting for", kind)
	}
}

// rateLimitedStub models a source whose upstream is refusing traffic: it fails
// immediately with the real 429 error instead of sitting on the budget.
type rateLimitedStub struct {
	epStubSource
	calls atomic.Int32
}

func (s *rateLimitedStub) Search(ctx context.Context, _ string) ([]*models.Anime, error) {
	s.calls.Add(1)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, errRateLimited
}

var errRateLimited = errors.New("server returned: 429 Too Many Requests")

// The "dexter" scenario end to end at the dispatcher: HiAnime down, the two anime
// sources legitimately empty (Dexter is a US TV series), SuperFlix refusing
// traffic. The user waited 12 seconds per attempt to be told it "timed out".
//
// A source that fails fast must make the whole search fail fast, and the reason
// it gives must survive into the aggregate error — "timed out" and "rate
// limited" call for different reactions from the user.
func TestSearchAll_RateLimitedSourceFailsFastAndKeepsItsReason(t *testing.T) {
	// Swaps the global registry — not parallel.
	sf := &rateLimitedStub{epStubSource: epStubSource{
		desc: source.Descriptor{Kind: source.SuperFlix, Priority: 30},
	}}
	hianime := newSearchStub(source.HiAnime, nil, errors.New("HiAnime search: upstream unavailable with HTTP 503"))
	empty := newSearchStub(source.Goyabu, nil, nil)
	restore := source.SwapRegistryForTesting(sf, hianime, empty)
	t.Cleanup(restore)

	start := time.Now()
	_, err := SearchAll(context.Background(), "dexter")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, errRateLimited,
		"the raw cause must survive so callers can classify the failure")
	assert.Contains(t, err.Error(), "refusing this network",
		"the aggregate must say why SuperFlix refused, not just that everything failed")
	assert.Less(t, elapsed, perSourceSearchTimeout/2,
		"no source was slow; the search must not spend its per-source budget anyway")
	assert.Equal(t, int32(1), sf.calls.Load(),
		"one search is one call per source: retrying inside the fan-out is what fed the rate limiter")
}

// A user reading "all sources failed: SuperFlix: server returned: 429" cannot
// tell whether GoAnime is broken, the title does not exist, or the host is
// refusing them — and only the last one has an action (wait). The per-source
// diagnostics that already feed the circuit breaker are therefore carried into
// the message too, while the raw errors stay in the chain for errors.Is.
func TestSearchAll_AllFailedErrorCarriesTheDiagnostics(t *testing.T) {
	// Swaps the global registry — not parallel.
	sf := &rateLimitedStub{epStubSource: epStubSource{
		desc: source.Descriptor{Kind: source.SuperFlix, Priority: 30},
	}}
	restore := source.SwapRegistryForTesting(sf)
	t.Cleanup(restore)

	_, err := SearchAll(context.Background(), "matrix")
	require.Error(t, err)

	var failure *SearchFailure
	require.ErrorAs(t, err, &failure, "callers need the failure as data, not as a string to parse")

	require.Len(t, failure.Sources, 1)
	assert.Equal(t, source.SuperFlix, failure.Sources[0].Kind)
	assert.True(t, failure.RateLimited(), "a 429 is the one failure with an action attached: wait")

	// Short for the terminal...
	assert.Contains(t, err.Error(), "SuperFlix", "the message must name which source refused")
	assert.NotContains(t, err.Error(), "429",
		"HTTP codes are noise on a terminal; the short line says what to do instead")

	// ...and complete for the log.
	assert.Contains(t, failure.Detail(), "429", "the raw cause must stay reachable for debugging")
	assert.ErrorIs(t, err, errRateLimited,
		"wrapping must survive, or callers can no longer classify the failure")
}
