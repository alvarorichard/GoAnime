package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// Where Pluto TV is plugged into the search: the selector, the sources named
// on a result, the breakdown the debug log prints.

func TestRegistryKindsForPlutoTV(t *testing.T) {
	for _, selector := range []string{"plutotv", "pluto"} {
		assert.Equal(t, []apisource.SourceKind{apisource.PlutoTV}, registryKindsForSource(selector), selector)
	}
	assert.Contains(t, registryKindsForSource("movietv"), apisource.PlutoTV, "-dm and the movie tab search it")
	assert.Contains(t, registryKindsForSource("ptbr"), apisource.PlutoTV, "its catalog has anime, and in Brazil it is Portuguese")
	assert.Contains(t, registryKindsForSource("pt-br"), apisource.PlutoTV)
	assert.Nil(t, registryKindsForSource(""), "no selector searches every registered source, Pluto TV among them")
}

func TestPlutoTVNeverShadowsAnotherSelector(t *testing.T) {
	for _, selector := range []string{"hianime", "animefire", "goyabu", "animesdigital", "anroll", "animesfhd", "startflix", "topcine", "pobreflix", "pomfy", "animedrive"} {
		assert.NotContains(t, registryKindsForSource(selector), apisource.PlutoTV, selector)
	}
}

func TestIdentifyAnimeSourcesNamesPlutoTV(t *testing.T) {
	t.Run("by the selector", func(t *testing.T) {
		for _, selector := range []string{"plutotv", "pluto"} {
			animes := []*models.Anime{{Name: "x", URL: "https://example.com/a"}}
			assert.NoError(t, identifyAnimeSources(animes, selector))
			assert.Equal(t, "PlutoTV", animes[0].Source, selector)
		}
	})
	t.Run("by the URL", func(t *testing.T) {
		for _, raw := range []string{
			"https://pluto.tv/on-demand/movies/62b9ea9ec4eb8f00140cb633",
			"https://pluto.tv/on-demand/series/6206bba315c8610013683974",
		} {
			animes := []*models.Anime{{Name: "x", URL: raw}}
			assert.NoError(t, identifyAnimeSources(animes, ""))
			assert.Equal(t, "PlutoTV", animes[0].Source, raw)
		}
	})
	t.Run("a source already named is kept", func(t *testing.T) {
		animes := []*models.Anime{{Name: "x", URL: "https://pluto.tv/on-demand/movies/x", Source: "Pomfy"}}
		assert.NoError(t, identifyAnimeSources(animes, "plutotv"))
		assert.Equal(t, "Pomfy", animes[0].Source)
	})
	t.Run("another site is not Pluto TV", func(t *testing.T) {
		animes := []*models.Anime{{Name: "x", URL: "https://animefire.io/a"}}
		assert.NoError(t, identifyAnimeSources(animes, ""))
		assert.Equal(t, "Animefire.io", animes[0].Source)
	})
}

func TestCountSourceBreakdownCountsPlutoTV(t *testing.T) {
	got := countSourceBreakdown([]*models.Anime{
		{Source: "PlutoTV"}, {Source: "plutotv"}, {Source: "Pomfy"}, {Source: "AnimesFHD"}, {Source: "Nope"},
	})
	assert.Equal(t, 2, got.PlutoTV, "the spelling is not case-sensitive")
	assert.Equal(t, 1, got.AnimesFHD)
}

func TestPlutoTVResultsOpenTheEpisodeFlowWithoutASpinner(t *testing.T) {
	// Listing a series opens the season picker: a spinner running beside it
	// would eat the terminal. The source is named, so it holds even when the
	// media type was lost on the way from a saved entry.
	assert.True(t, (&models.Media{Source: "PlutoTV"}).HasInteractiveEpisodeFlow())
	assert.True(t, (&models.Media{Source: "PlutoTV", MediaType: models.MediaTypeMovie}).HasInteractiveEpisodeFlow())
}

// Pluto TV's catalog has anime, so it is searched on the anime tabs too, not
// only on MOVIES/TV: the PT-BR group, the ENGLISH tab and ALL.
func TestPlutoTVIsSearchedOnTheAnimeTabs(t *testing.T) {
	for _, tab := range []string{"ptbr", "pt-br", "english", "movietv", "plutotv", "pluto"} {
		assert.Contains(t, registryKindsForSource(tab), apisource.PlutoTV, tab)
	}
	assert.Empty(t, registryKindsForSource(""), "ALL passes no kinds, which means every registered source")
}

// --source hianime still means HiAnime alone: the ENGLISH tab has its own
// selector, so adding Pluto to the tab did not change the flag.
func TestTheHiAnimeFlagStaysHiAnimeOnly(t *testing.T) {
	assert.Equal(t, []apisource.SourceKind{apisource.HiAnime}, registryKindsForSource("hianime"))
	assert.NotContains(t, registryKindsForSource("goyabu"), apisource.PlutoTV)
}

// A search made on a language tab carries that language to every source, so a
// source of mixed languages (Pluto TV) can keep the titles that fit the tab.
func TestASearchCarriesTheLanguageOfItsTab(t *testing.T) {
	for scope, want := range map[string]string{
		"english": "EN", "ptbr": "PT", "pt-br": "PT",
		"": "", "movietv": "", "plutotv": "", "hianime": "", "animefire": "",
	} {
		var got string
		asked := false
		search := func(ctx context.Context, _ string, _ []apisource.SourceKind) ([]*models.Anime, error) {
			got, asked = apisource.LanguageFrom(ctx), true
			return []*models.Anime{{Name: "X", URL: "https://example.test/x", Source: "HiAnime"}}, nil
		}
		pick := func(animes []*models.Anime) (*models.Anime, error) { return animes[0], nil }
		_, err := searchAnimeEnhanced("x", scope, search, pick, func(*models.Anime) error { return nil })
		require.NoError(t, err, scope)
		require.True(t, asked, scope)
		assert.Equal(t, want, got, "scope %q", scope)
	}
}
