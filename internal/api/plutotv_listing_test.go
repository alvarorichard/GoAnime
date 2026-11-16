package api

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
)

// One answer of the catalog holds every season of a series, and a download of
// a whole series asks for it once for the season numbers and once per season.
// The listing is therefore kept for a short while and reused.

const ptSeasonsPath = "/v4/vod/series/" + ptSeriesID + "/seasons"

// plutoClock replaces the listing cache's clock for one test.
func plutoClock(t *testing.T) *time.Time {
	t.Helper()
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	prev := plutoShows.now
	plutoShows.now = func() time.Time { return now }
	t.Cleanup(func() { plutoShows.now = prev })
	return &now
}

func TestAWholeSeriesIsListedFromOneAnswer(t *testing.T) {
	f := newPlutoAPIFake(t)
	series := plutoSeries()

	numbers, err := GetPlutoTVSeasonNumbers(series)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2}, numbers)
	for _, n := range numbers {
		eps, err := GetPlutoTVSeasonEpisodes(series, n)
		require.NoError(t, err)
		require.NotEmpty(t, eps)
	}
	assert.Equal(t, 1, f.hit(ptSeasonsPath), "the numbers and both seasons came from the one answer")
}

func TestAListingIsReadAgainOnceItIsOld(t *testing.T) {
	f := newPlutoAPIFake(t)
	now := plutoClock(t)
	series := plutoSeries()

	_, err := GetPlutoTVSeasonNumbers(series)
	require.NoError(t, err)
	*now = now.Add(plutoShowTTL - time.Second)
	_, err = GetPlutoTVSeasonNumbers(series)
	require.NoError(t, err)
	assert.Equal(t, 1, f.hit(ptSeasonsPath), "still fresh")

	*now = now.Add(2 * time.Second)
	_, err = GetPlutoTVSeasonNumbers(series)
	require.NoError(t, err)
	assert.Equal(t, 2, f.hit(ptSeasonsPath), "past its time: the catalog is asked again")
}

func TestAFailedListingIsNotKept(t *testing.T) {
	f := newPlutoAPIFake(t)
	good := f.seasons
	f.seasons = `{"_id": "` + ptSeriesID + `", "name": "Death Note", "seasons": []}`
	_, err := GetPlutoTVSeasonNumbers(plutoSeries())
	require.Error(t, err, "a series that lists nothing")

	f.seasons = good
	numbers, err := GetPlutoTVSeasonNumbers(plutoSeries())
	require.NoError(t, err, "the failure was not remembered")
	assert.Equal(t, []int{1, 2}, numbers)
	assert.Equal(t, 2, f.hit(ptSeasonsPath))
}

func TestOldListingsAreDroppedWhenANewOneIsKept(t *testing.T) {
	newPlutoAPIFake(t)
	now := plutoClock(t)
	plutoShows.mu.Lock()
	plutoShows.byURL["https://pluto.tv/on-demand/series/old"] = plutoShowEntry{at: now.Add(-plutoShowTTL)}
	plutoShows.byURL["https://pluto.tv/on-demand/series/recent"] = plutoShowEntry{at: now.Add(-time.Second)}
	plutoShows.mu.Unlock()

	_, err := GetPlutoTVSeasonNumbers(plutoSeries())
	require.NoError(t, err)

	plutoShows.mu.Lock()
	defer plutoShows.mu.Unlock()
	assert.NotContains(t, plutoShows.byURL, "https://pluto.tv/on-demand/series/old", "nothing stale is kept around")
	assert.Contains(t, plutoShows.byURL, "https://pluto.tv/on-demand/series/recent")
	assert.Contains(t, plutoShows.byURL, plutotv.SeriesURL(ptSeriesID))
}

func TestForgetPlutoTVListingsReadsTheCatalogAgain(t *testing.T) {
	f := newPlutoAPIFake(t)
	_, err := GetPlutoTVSeasonNumbers(plutoSeries())
	require.NoError(t, err)
	ForgetPlutoTVListings()
	_, err = GetPlutoTVSeasonNumbers(plutoSeries())
	require.NoError(t, err)
	assert.Equal(t, 2, f.hit(ptSeasonsPath))
}

func TestListingsAreSafeToReadFromManyDownloadsAtOnce(t *testing.T) {
	newPlutoAPIFake(t)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			numbers, err := GetPlutoTVSeasonNumbers(plutoSeries())
			assert.NoError(t, err)
			assert.Equal(t, []int{1, 2}, numbers)
		})
	}
	wg.Wait()
}

// ---- the caller's context ---------------------------------------------------

func TestACancelledListingStopsAtOnce(t *testing.T) {
	f := newPlutoAPIFake(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	started := time.Now()
	_, err := GetPlutoTVEpisodes(ctx, plutoSeries())
	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(started), 5*time.Second, "not left to the listing budget")
	assert.Zero(t, f.hit(ptSeasonsPath), "nothing was asked of the catalog")
}

func TestACancelledStreamStopsAtOnce(t *testing.T) {
	f := newPlutoAPIFake(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := ResolvePlutoTVStream(ctx, plutoSeries(), &models.Episode{URL: plutotv.EpisodeURL(ptSeriesID, ptEpisode1)}, "best")
	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, f.hit("/v2/stitch/hls/episode/"+ptEpisode1+"/master.m3u8"))
}

func TestAMovieNeedsNoCatalogEvenWhenTheCallerHasGone(t *testing.T) {
	f := newPlutoAPIFake(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	eps, err := GetPlutoTVEpisodes(ctx, plutoMovie())
	require.NoError(t, err, "a movie's one entry is its own address: nothing is requested")
	require.Len(t, eps, 1)
	assert.Zero(t, f.boots)
}
