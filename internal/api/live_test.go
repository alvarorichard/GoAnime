package api

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanShowLiveSearch(t *testing.T) {
	prev := searchLiveFetchFn
	searchLiveFetchFn = nil
	t.Cleanup(func() { searchLiveFetchFn = prev })

	assert.False(t, canShowLiveSearch(), "no seam, no live screen")
}

func TestRunLiveSearchChosen(t *testing.T) {
	prevFetch := searchLiveFetchFn
	prevRunner := searchLiveRunnerFn
	t.Cleanup(func() { searchLiveFetchFn = prevFetch; searchLiveRunnerFn = prevRunner })

	chosen := &models.Anime{Name: "X", Source: "Animefire.io"}
	searchLiveFetchFn = func(ctx context.Context, query string, kinds []apisource.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error) {
		emit(models.SearchEvent{Plan: true, Searched: []string{"AnimeFire"}, GiveUp: 1})
		return []*models.Anime{chosen}, nil
	}
	searchLiveRunnerFn = func(tui.SearchLiveOptions, <-chan models.SearchEvent) (tui.SearchLiveOutcome, error) {
		return tui.SearchLiveOutcome{Chosen: chosen}, nil
	}

	animes, got, err := runLiveSearch("x", nil)
	require.NoError(t, err)
	assert.Nil(t, animes, "early pick skips the collection")
	assert.Same(t, chosen, got)
}

func TestRunLiveSearchCompleted(t *testing.T) {
	prevFetch := searchLiveFetchFn
	prevRunner := searchLiveRunnerFn
	t.Cleanup(func() { searchLiveFetchFn = prevFetch; searchLiveRunnerFn = prevRunner })

	collected := []*models.Anime{{Name: "X"}}
	searchLiveFetchFn = func(ctx context.Context, query string, kinds []apisource.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error) {
		emit(models.SearchEvent{Plan: true, Searched: []string{"AnimeFire"}, GiveUp: 1})
		return collected, nil
	}
	searchLiveRunnerFn = func(tui.SearchLiveOptions, <-chan models.SearchEvent) (tui.SearchLiveOutcome, error) {
		return tui.SearchLiveOutcome{Completed: true}, nil
	}

	animes, got, err := runLiveSearch("x", nil)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.Same(t, collected[0], animes[0])
}

func TestRunLiveSearchError(t *testing.T) {
	prevFetch := searchLiveFetchFn
	prevRunner := searchLiveRunnerFn
	t.Cleanup(func() { searchLiveFetchFn = prevFetch; searchLiveRunnerFn = prevRunner })

	searchLiveFetchFn = func(ctx context.Context, query string, kinds []apisource.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error) {
		return nil, nil
	}
	want := errors.New("no tty")
	searchLiveRunnerFn = func(tui.SearchLiveOptions, <-chan models.SearchEvent) (tui.SearchLiveOutcome, error) {
		return tui.SearchLiveOutcome{}, want
	}

	_, _, err := runLiveSearch("x", nil)
	assert.ErrorIs(t, err, want)
}

func TestRunLiveSearchForwardsQueryAndEvents(t *testing.T) {
	prevFetch := searchLiveFetchFn
	prevRunner := searchLiveRunnerFn
	t.Cleanup(func() { searchLiveFetchFn = prevFetch; searchLiveRunnerFn = prevRunner })

	plan := models.SearchEvent{Plan: true, Searched: []string{"AnimeFire"}, GiveUp: 1}
	done := models.SearchEvent{Name: "AnimeFire", State: models.SearchDone, Count: 2}
	var gotQuery string
	var gotKinds []apisource.SourceKind
	searchLiveFetchFn = func(ctx context.Context, query string, kinds []apisource.SourceKind, emit func(models.SearchEvent)) ([]*models.Anime, error) {
		gotQuery, gotKinds = query, kinds
		emit(plan)
		emit(done)
		return []*models.Anime{{Name: "X"}}, nil
	}
	var gotOptsQuery string
	var seen []models.SearchEvent
	searchLiveRunnerFn = func(opts tui.SearchLiveOptions, events <-chan models.SearchEvent) (tui.SearchLiveOutcome, error) {
		gotOptsQuery = opts.Query
		for range 200 {
			drained := false
			for {
				select {
				case ev := <-events:
					seen = append(seen, ev)
					drained = true
				default:
					goto drained
				}
			}
		drained:
			if len(seen) >= 2 {
				break
			}
			if !drained {
				time.Sleep(5 * time.Millisecond)
			}
		}
		return tui.SearchLiveOutcome{Completed: true}, nil
	}

	animes, chosen, err := runLiveSearch("namakura gatana", []apisource.SourceKind{apisource.AnimeFire})
	require.NoError(t, err)
	assert.Nil(t, chosen)
	require.Len(t, animes, 1)
	assert.Equal(t, "namakura gatana", gotQuery)
	assert.Equal(t, []apisource.SourceKind{apisource.AnimeFire}, gotKinds)
	assert.Equal(t, "namakura gatana", gotOptsQuery)
	require.Len(t, seen, 2, "the screen sees plan plus outcome")
	assert.True(t, seen[0].Plan)
	assert.Equal(t, 2, seen[1].Count)
}

func TestHumanQuery(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "namakra gatana", humanQuery("namakra-gatana"))
	assert.Equal(t, "naruto", humanQuery("naruto"))
	assert.Equal(t, "", humanQuery(""))
	assert.Equal(t, "re zero", humanQuery("re-zero"))
}

func TestSearchAnimeAutoSkipsLive(t *testing.T) {
	sentinel := SearchLiveFunc(func(context.Context, string, []apisource.SourceKind, func(models.SearchEvent)) ([]*models.Anime, error) {
		return nil, nil
	})
	prevLive, prevSearch := searchLiveFetchFn, searchFetchFn
	t.Cleanup(func() { searchLiveFetchFn = prevLive; searchFetchFn = prevSearch })
	searchLiveFetchFn = sentinel

	var sawNilHook bool
	searchFetchFn = func(context.Context, string, []apisource.SourceKind) ([]*models.Anime, error) {
		sawNilHook = searchLiveFetchFn == nil
		return nil, errors.New("nope")
	}
	_, _ = SearchAnimeAuto("x", "", nil)
	assert.True(t, sawNilHook, "auto reopen must not see the live screen")
	assert.NotNil(t, searchLiveFetchFn, "the hook is restored afterwards")
}

func FuzzHumanQuery(f *testing.F) {
	for _, seed := range []string{"", "-", "a", "namakra-gatana", "--x--", "re-zero"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, name string) {
		got := humanQuery(name)
		if got != name {
			assert.NotContains(t, got, "-")
		}
		if name != "" && strings.Trim(name, "- ") != "" {
			assert.NotEmpty(t, got)
		}
	})
}
