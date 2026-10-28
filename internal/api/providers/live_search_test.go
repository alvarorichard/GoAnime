package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func collectLiveEvents(t *testing.T, ctx context.Context, query string, kinds []source.SourceKind) ([]models.SearchEvent, []*models.Anime, error) {
	t.Helper()
	var events []models.SearchEvent
	animes, err := SearchLive(ctx, query, kinds, func(ev models.SearchEvent) {
		events = append(events, ev)
	})
	return events, animes, err
}

func TestPlanLiveSearch_ClassifiesSources(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	goyabu := newSearchStub(source.Goyabu, nil, nil)
	goyabu.desc.DefaultDisabled = true
	off := newSearchStub(source.TopCine, nil, nil)
	restore := source.SwapRegistryForTesting(fire, goyabu, off)
	t.Cleanup(restore)
	t.Setenv("GOANIME_DISABLED_SOURCES", "topcine")

	searchers, offList, searched := planLiveSearch(nil)

	assert.Len(t, searchers, 1)
	assert.Equal(t, []string{"AnimeFire"}, searched)
	require.Len(t, offList, 2)
	assert.Equal(t, models.SearchOffEntry{Name: "Goyabu", Reason: "off by default"}, offList[0])
	assert.Equal(t, models.SearchOffEntry{Name: "TopCine", Reason: "disabled"}, offList[1])
}

func TestSearchLive_StreamsOutcomes(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "[PT-BR] Naruto"}}, nil)
	hi := newSearchStub(source.HiAnime, []*models.Anime{{Name: "[English] Naruto"}}, nil)
	restore := source.SwapRegistryForTesting(fire, hi)
	t.Cleanup(restore)

	events, animes, err := collectLiveEvents(t, context.Background(), "naruto", nil)
	require.NoError(t, err)
	assert.Len(t, animes, 2)

	require.NotEmpty(t, events)
	assert.True(t, events[0].Plan, "the plan comes first")
	assert.Equal(t, 15*time.Second, events[0].GiveUp)
	assert.Len(t, events[0].Searched, 2)

	var done int
	for _, ev := range events[1:] {
		require.False(t, ev.Plan)
		require.NotEmpty(t, ev.Name)
		if ev.State == models.SearchDone {
			done++
			assert.GreaterOrEqual(t, ev.Count, 1)
			assert.GreaterOrEqual(t, ev.Elapsed, time.Duration(0))
		}
	}
	assert.Equal(t, 2, done)
}

func TestSearchLive_FailureEventCarriesReason(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	bad := newSearchStub(source.TopCine, nil, errors.New("boom"))
	restore := source.SwapRegistryForTesting(fire, bad)
	t.Cleanup(restore)

	events, animes, err := collectLiveEvents(t, context.Background(), "x", nil)
	require.NoError(t, err, "one success carries the search")
	assert.Len(t, animes, 1)

	var failed *models.SearchEvent
	for i, ev := range events {
		if !ev.Plan && ev.State == models.SearchFailed {
			failed = &events[i]
		}
	}
	require.NotNil(t, failed)
	assert.Equal(t, "TopCine", failed.Name)
	assert.NotEmpty(t, failed.Reason)
}

func TestSearchLive_AllFailedErrors(t *testing.T) {
	// Swaps the global registry — not parallel.
	bad := newSearchStub(source.TopCine, nil, errors.New("boom"))
	restore := source.SwapRegistryForTesting(bad)
	t.Cleanup(restore)

	events, animes, err := collectLiveEvents(t, context.Background(), "x", nil)
	require.Error(t, err)
	assert.Empty(t, animes)
	assert.True(t, events[0].Plan)
}

func TestSearchLive_NoSearchersErrors(t *testing.T) {
	// Swaps the global registry — not parallel.
	restore := source.SwapRegistryForTesting()
	t.Cleanup(restore)

	_, _, err := collectLiveEvents(t, context.Background(), "x", nil)
	require.Error(t, err)
}

func TestSearchLive_KindFilter(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	hi := newSearchStub(source.HiAnime, []*models.Anime{{Name: "Y"}}, nil)
	restore := source.SwapRegistryForTesting(fire, hi)
	t.Cleanup(restore)

	events, animes, err := collectLiveEvents(t, context.Background(), "x", []source.SourceKind{source.HiAnime})
	require.NoError(t, err)
	assert.Len(t, animes, 1)
	assert.Equal(t, []string{"HiAnime"}, events[0].Searched)
	assert.Equal(t, int32(0), fire.calls.Load())
}

func TestPlanLiveSearch_BreakerOpenCoolsDown(t *testing.T) {
	// Swaps the global registry and breaker — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	flaky := newSearchStub(source.HiAnime, nil, nil)
	restore := source.SwapRegistryForTesting(fire, flaky)
	t.Cleanup(restore)

	prev := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prev })
	for range 3 {
		searchBreaker.RecordFailure(string(source.HiAnime),
			netx.DiagnoseError("HiAnime", "search", netx.NewHTTPStatusError("HiAnime", "search", 521)))
	}

	searchers, off, searched := planLiveSearch(nil)
	assert.Len(t, searchers, 1)
	assert.Equal(t, []string{"AnimeFire"}, searched)
	require.Len(t, off, 1)
	assert.Equal(t, models.SearchOffEntry{Name: "HiAnime", Reason: "cooling down"}, off[0])
}

func TestPlanLiveSearch_UnknownKindSkipped(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	restore := source.SwapRegistryForTesting(fire)
	t.Cleanup(restore)

	searchers, off, searched := planLiveSearch([]source.SourceKind{source.AnimeFire, "Nope"})
	assert.Len(t, searchers, 1)
	assert.Empty(t, off)
	assert.Equal(t, []string{"AnimeFire"}, searched)
}

func TestNoResultsInfo(t *testing.T) {
	t.Parallel()
	err := &NoResultsInfo{Query: "x", AnsweredKinds: []source.SourceKind{source.HiAnime}}
	assert.Equal(t, "no results found for: x", err.Error())
	assert.ErrorIs(t, err, api.ErrNoResults)
}

func TestSearchLive_PlanCarriesOff(t *testing.T) {
	// Swaps the global registry — not parallel.
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	goyabu := newSearchStub(source.Goyabu, nil, nil)
	goyabu.desc.DefaultDisabled = true
	restore := source.SwapRegistryForTesting(fire, goyabu)
	t.Cleanup(restore)

	events, _, err := collectLiveEvents(t, context.Background(), "x", nil)
	require.NoError(t, err)
	require.True(t, events[0].Plan)
	require.Len(t, events[0].Off, 1)
	assert.Equal(t, "Goyabu", events[0].Off[0].Name)
	assert.Equal(t, "off by default", events[0].Off[0].Reason)
}

func TestSearchLive_DeadlineFinishesPendingSourcesAndKeepsResults(t *testing.T) {
	fast := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "Naruto"}}, nil)
	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(fast, stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	start := time.Now()
	events, results, err := collectLiveEvents(t, ctx, "naruto", nil)
	require.NoError(t, err)
	require.Less(t, time.Since(start), time.Second)
	require.Len(t, results, 1)
	assert.Equal(t, "Naruto", results[0].Name)
	require.Len(t, events, 3)
	assert.Equal(t, models.SearchFailed, events[2].State)
	assert.Equal(t, "AnimeDrive", events[2].Name)
	assert.Equal(t, "timed out", events[2].Reason)
}

func TestSearchLive_AllPendingTimeout(t *testing.T) {
	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	events, results, err := collectLiveEvents(t, ctx, "x", nil)
	require.Error(t, err)
	assert.Empty(t, results)
	require.Len(t, events, 2)
	assert.Equal(t, models.SearchFailed, events[1].State)
}

func TestSearchLive_CancelMarksPending(t *testing.T) {
	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	events, _, err := collectLiveEvents(t, ctx, "x", nil)
	require.Error(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "cancelled", events[1].Reason)
}

func TestSearchLive_LateResultDoesNotChangeCompletedSearch(t *testing.T) {
	stuck := &hangingSearchSource{
		epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}},
		release:      make(chan struct{}), results: []*models.Anime{{Name: "Late"}},
	}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	events, results, err := collectLiveEvents(t, ctx, "x", nil)
	require.Error(t, err)
	assert.Empty(t, results)
	close(stuck.release)
	assert.Len(t, events, 2)
	assert.Equal(t, models.SearchFailed, events[1].State)
}

func TestSearchLive_CancellationDoesNotTripCircuitBreaker(t *testing.T) {
	prev := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prev })

	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		events, _, err := collectLiveEvents(t, ctx, "x", nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		require.Len(t, events, 2)
		assert.Equal(t, "cancelled", events[1].Reason)
	}

	_, _, open := searchBreaker.IsOpen(string(source.AnimeDrive))
	assert.False(t, open, "live search cancellations must not trip circuit breaker")
}

func TestSearchLive_InFlightCancellation(t *testing.T) {
	prev := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prev })

	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(20*time.Millisecond, cancel)

	events, results, err := collectLiveEvents(t, ctx, "x", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, results)
	require.Len(t, events, 2)
	assert.Equal(t, "cancelled", events[1].Reason)

	_, _, open := searchBreaker.IsOpen(string(source.AnimeDrive))
	assert.False(t, open, "in-flight cancellation must not trip circuit breaker")
}

func TestSearchLive_ParentDeadlineDoesNotTripCircuitBreaker(t *testing.T) {
	prev := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prev })

	stuck := &hangingSearchSource{epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 1}}, release: make(chan struct{})}
	restore := source.SwapRegistryForTesting(stuck)
	t.Cleanup(restore)
	t.Cleanup(func() { close(stuck.release) })

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		events, _, err := collectLiveEvents(t, ctx, "x", nil)
		cancel()
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		require.Len(t, events, 2)
		assert.Equal(t, "timed out", events[1].Reason)
	}

	_, _, open := searchBreaker.IsOpen(string(source.AnimeDrive))
	assert.False(t, open, "live search parent deadlines must not trip circuit breaker")
}

func TestSearchLive_ProviderOwnTimeoutTripsCircuitBreaker(t *testing.T) {
	prev := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prev })

	failing := newSearchStub(source.AnimeDrive, nil, context.DeadlineExceeded)
	restore := source.SwapRegistryForTesting(failing)
	t.Cleanup(restore)

	for i := 0; i < 3; i++ {
		events, _, err := collectLiveEvents(t, context.Background(), "x", nil)
		require.Error(t, err)
		require.Len(t, events, 2)
	}

	_, _, open := searchBreaker.IsOpen(string(source.AnimeDrive))
	assert.True(t, open, "provider's own timeout in live search must trip the circuit breaker")
}
