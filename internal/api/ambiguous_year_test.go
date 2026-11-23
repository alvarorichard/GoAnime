package api

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PlutoTV lists iCarly with no year while TopCine and StartFlix list iCarly
// (2007) and iCarly (2021). The yearless copy used to land in whichever group
// came first — the 2021 revival, when PlutoTV carries the 2007 original. These
// guard the lookup that gives it the original's year.

// imdbICarly is IMDb's suggestion list for "icarly" (checked 2026-11-23).
var imdbICarly = []movie.IMDbTitle{
	{ID: "tt0972534", Title: "iCarly", Year: 2007, Kind: "tvSeries"},
	{ID: "tt13657062", Title: "iCarly", Year: 2021, Kind: "tvSeries"},
	{ID: "tt1286124", Title: "iCarly: iGo to Japan", Year: 2008, Kind: "tvMovie"},
	{ID: "tt2050443", Title: "iCarly Webisodes", Year: 2007, Kind: "tvMiniSeries"},
	{ID: "tt4156236", Title: "iCarly", Year: 2009, Kind: "videoGame"},
}

// stubIMDb swaps the IMDb lookup for fn and empties the cache around the test.
// The seam is global, so these tests run one after another.
func stubIMDb(t *testing.T, fn func(name string) ([]movie.IMDbTitle, error)) *atomic.Int32 {
	t.Helper()
	var calls atomic.Int32
	prev := imdbTitlesFn
	imdbTitlesFn = func(name string) ([]movie.IMDbTitle, error) {
		calls.Add(1)
		return fn(name)
	}
	imdbTitlesMu.Lock()
	imdbTitlesCache = map[string][]movie.IMDbTitle{}
	imdbTitlesMu.Unlock()
	t.Cleanup(func() {
		imdbTitlesFn = prev
		imdbTitlesMu.Lock()
		imdbTitlesCache = map[string][]movie.IMDbTitle{}
		imdbTitlesMu.Unlock()
	})
	return &calls
}

func yr(name, source, year string, mt models.MediaType) *models.Anime {
	return &models.Anime{Name: name, Source: source, Year: year, MediaType: mt}
}

func icarlyResults() (results []*models.Anime, pluto *models.Anime) {
	pluto = yr("[TV] [PT-BR] iCarly", "PlutoTV", "", models.MediaTypeTV)
	return []*models.Anime{
		yr("[TV] [PT-BR] iCarly", "TopCine", "2007", models.MediaTypeTV),
		yr("[TV] [PT-BR] iCarly", "TopCine", "2021", models.MediaTypeTV),
		yr("[TV] [PT-BR] iCarly", "StartFlix", "2021", models.MediaTypeTV),
		yr("[TV] [PT-BR] iCarly", "StartFlix", "2007", models.MediaTypeTV),
		pluto,
	}, pluto
}

func TestResolveAmbiguousYears_PlutoICarlyIsTheOriginal(t *testing.T) {
	calls := stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return imdbICarly, nil })
	results, pluto := icarlyResults()
	resolveAmbiguousYears(results)
	assert.Equal(t, "2007", pluto.Year, "the yearless copy is the original")
	assert.EqualValues(t, 1, calls.Load(), "one title, one lookup")
	assert.Equal(t, "2007", results[0].Year, "copies that had a year keep it")
	assert.Equal(t, "2021", results[1].Year)
}

func TestResolveAmbiguousYears_AsksIMDbWithTheCleanTitle(t *testing.T) {
	var asked string
	stubIMDb(t, func(name string) ([]movie.IMDbTitle, error) { asked = name; return imdbICarly, nil })
	results, _ := icarlyResults()
	results[4].Name = "[TV] [PT-BR] iCarly (Dublado)"
	resolveAmbiguousYears(results)
	assert.Equal(t, "iCarly", asked, "tags and audio words are not part of the title")
}

func TestResolveAmbiguousYears_NoLookupUnlessYearsDisagree(t *testing.T) {
	calls := stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return imdbICarly, nil })
	cases := map[string][]*models.Anime{
		"one year only": {
			yr("[TV] [PT-BR] iCarly", "TopCine", "2021", models.MediaTypeTV),
			yr("[TV] [PT-BR] iCarly", "PlutoTV", "", models.MediaTypeTV),
		},
		"every copy has a year": {
			yr("[TV] [PT-BR] iCarly", "TopCine", "2007", models.MediaTypeTV),
			yr("[TV] [PT-BR] iCarly", "TopCine", "2021", models.MediaTypeTV),
		},
		"no year at all": {
			yr("[TV] [PT-BR] iCarly", "PlutoTV", "", models.MediaTypeTV),
			yr("[TV] [PT-BR] iCarly", "Pomfy", "", models.MediaTypeTV),
		},
		"the years belong to different titles": {
			yr("[TV] [PT-BR] iCarly", "PlutoTV", "", models.MediaTypeTV),
			yr("[TV] [PT-BR] Dexter", "TopCine", "2006", models.MediaTypeTV),
			yr("[TV] [PT-BR] Dexter", "TopCine", "2021", models.MediaTypeTV),
		},
		"anime is never looked up": {
			yr("[PT-BR] Frieren", "AnimeFire", "2023", models.MediaTypeAnime),
			yr("[English] Frieren", "HiAnime", "2024", models.MediaTypeAnime),
			yr("[PT-BR] Frieren", "Goyabu", "", models.MediaTypeAnime),
		},
		"a movie and a series are different kinds": {
			yr("[TV] [PT-BR] Dexter", "TopCine", "2006", models.MediaTypeTV),
			yr("[Movie] [PT-BR] Dexter", "TopCine", "2020", models.MediaTypeMovie),
			yr("[TV] [PT-BR] Dexter", "PlutoTV", "", models.MediaTypeTV),
		},
	}
	for name, results := range cases {
		resolveAmbiguousYears(results)
		assert.Zero(t, calls.Load(), "%s: IMDb must not be asked", name)
	}
	resolveAmbiguousYears([]*models.Anime{nil, nil})
	assert.Zero(t, calls.Load())
}

func TestResolveAmbiguousYears_FailureLeavesTheCopyAlone(t *testing.T) {
	stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return nil, errors.New("offline") })
	results, pluto := icarlyResults()
	resolveAmbiguousYears(results)
	assert.Empty(t, pluto.Year)
	assert.Equal(t, "2007", results[0].Year)
}

func TestResolveAmbiguousYears_NoExactTitleLeavesTheCopyAlone(t *testing.T) {
	stubIMDb(t, func(string) ([]movie.IMDbTitle, error) {
		return []movie.IMDbTitle{
			{ID: "tt1286124", Title: "iCarly: iGo to Japan", Year: 2008, Kind: "tvMovie"},
			{ID: "tt4156236", Title: "iCarly", Year: 2009, Kind: "videoGame"},
			{ID: "tt2050443", Title: "iCarly Webisodes", Year: 2007, Kind: "tvMiniSeries"},
		}, nil
	})
	results, pluto := icarlyResults()
	resolveAmbiguousYears(results)
	assert.Empty(t, pluto.Year, "a game, a TV movie and a spin-off are not the series")
}

func TestResolveAmbiguousYears_KindDecidesWhichTitleCounts(t *testing.T) {
	stubIMDb(t, func(string) ([]movie.IMDbTitle, error) {
		return []movie.IMDbTitle{
			{ID: "tt0773262", Title: "Dexter", Year: 2006, Kind: "tvSeries"},
			{ID: "tt0000001", Title: "Dexter", Year: 1999, Kind: "movie"},
			{ID: "tt0000002", Title: "Dexter", Year: 2020, Kind: "movie"},
		}, nil
	})
	films := yr("[Movie] [PT-BR] Dexter", "PlutoTV", "", models.MediaTypeMovie)
	resolveAmbiguousYears([]*models.Anime{
		yr("[Movie] [PT-BR] Dexter", "TopCine", "1999", models.MediaTypeMovie),
		yr("[Movie] [PT-BR] Dexter", "TopCine", "2020", models.MediaTypeMovie),
		films,
	})
	assert.Equal(t, "1999", films.Year, "a movie takes the earliest movie, not the 2006 series")
}

func TestResolveAmbiguousYears_AskedOncePerTitleAcrossSearches(t *testing.T) {
	calls := stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return imdbICarly, nil })
	for range 3 {
		results, pluto := icarlyResults()
		resolveAmbiguousYears(results)
		assert.Equal(t, "2007", pluto.Year)
	}
	assert.EqualValues(t, 1, calls.Load(), "the search retry loop must not ask IMDb again")
}

func TestResolveAmbiguousYears_CapsTheLookups(t *testing.T) {
	calls := stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return nil, nil })
	var results []*models.Anime
	for _, title := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"} {
		results = append(results,
			yr("[TV] [PT-BR] "+title, "TopCine", "2001", models.MediaTypeTV),
			yr("[TV] [PT-BR] "+title, "TopCine", "2002", models.MediaTypeTV),
			yr("[TV] [PT-BR] "+title, "PlutoTV", "", models.MediaTypeTV))
	}
	resolveAmbiguousYears(results)
	assert.EqualValues(t, maxAmbiguousLookups, calls.Load())
}

func TestOriginalYearPicksTheEarliestExactMatch(t *testing.T) {
	g := &yearGroup{mediaType: models.MediaTypeTV, name: "iCarly"}
	assert.Equal(t, "2007", originalYear(imdbICarly, g))
	assert.Equal(t, "2007", originalYear([]movie.IMDbTitle{imdbICarly[1], imdbICarly[0]}, g), "order does not matter")
	assert.Empty(t, originalYear(nil, g))
	assert.Empty(t, originalYear([]movie.IMDbTitle{{Title: "iCarly", Year: 0, Kind: "tvSeries"}}, g), "a title with no year names none")
	assert.Equal(t, "2007", originalYear([]movie.IMDbTitle{
		{Title: "iCarly", Year: 2007, Kind: "tvSeries"},
		{Title: "iCarly", Year: 0, Kind: "tvSeries"},
	}, g), "an unreleased title never beats a dated one")
	assert.Equal(t, "2021", originalYear([]movie.IMDbTitle{{Title: "ICARLY", Year: 2021, Kind: "tvMiniSeries"}}, g), "case and mini-series count")
	require.Empty(t, originalYear([]movie.IMDbTitle{{Title: "iCarly", Year: 2007, Kind: "movie"}}, g), "a movie is not a series")
}

func TestYearTitleKey(t *testing.T) {
	for in, want := range map[string]string{
		"[TV] [PT-BR] iCarly":                 "icarly",
		"[Movie] [PT-BR] Dexter (2020)":       "dexter",
		"[TV] [PT-BR] iCarly (Dublado)":       "icarly",
		"[PT-BR] O Gato de Botas (Legendado)": "ogatodebotas",
		"iCarly":                              "icarly",
		"[TV] [PT-BR]":                        "",
	} {
		assert.Equal(t, want, yearTitleKey(in), in)
	}
}

// TestSearchPipelineResolvesYearlessCopies drives the search the way the app
// does: the screen that picks a result must be handed the PlutoTV copy with
// the original's year, or it groups with the 2021 revival again.
func TestSearchPipelineResolvesYearlessCopies(t *testing.T) {
	stubIMDb(t, func(string) ([]movie.IMDbTitle, error) { return imdbICarly, nil })
	results, pluto := icarlyResults()
	for _, r := range results {
		r.URL = "https://example.test/" + r.Source + "/" + r.Year
	}
	var offered []*models.Anime
	_, err := searchAnimeEnhanced("icarly", "", func(context.Context, string, []apisource.SourceKind) ([]*models.Anime, error) {
		return results, nil
	}, func(animes []*models.Anime) (*models.Anime, error) {
		offered = animes
		return animes[0], nil
	}, func(*models.Anime) error { return nil })
	require.NoError(t, err)
	var got *models.Anime
	for _, a := range offered {
		if a == pluto {
			got = a
		}
	}
	require.NotNil(t, got, "the PlutoTV copy reaches the picker")
	assert.Equal(t, "2007", got.Year)
}
