package playback

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAllAnimeSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{"source field", &models.Anime{Source: "AllAnime"}, true},
		{"url contains", &models.Anime{URL: "https://allanime.to/x"}, true},
		{"short id", &models.Anime{URL: "hHjXnUTda"}, true},
		{"http url unrelated", &models.Anime{URL: "https://animefire.io/x"}, false},
		{"empty", &models.Anime{}, false},
		// Regression: explicit non-AllAnime source must win even when the URL
		// is a short ID — SuperFlix/SFlix use numeric TMDB IDs as URLs and
		// were misrouted into AllAnime navigation.
		{"superflix numeric id", &models.Anime{Source: "SuperFlix", URL: "1234"}, false},
		{"sflix short id", &models.Anime{Source: "SFlix", URL: "abc123"}, false},
		{"other source with allanime-like id", &models.Anime{Source: "AnimeFire", URL: "hHjXnUTda"}, false},
		{"numeric-only id without source", &models.Anime{URL: "12345"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAllAnimeSource(tt.anime))
		})
	}
}

func TestExtractAllAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"short id", "hHjXnUTda", "hHjXnUTda"},
		// Regression: the old scan returned "https:" (the scheme) for any full
		// URL, poisoning the navigator cache key and every episode fetch.
		{"allanime url extracts id segment", "https://allanime.to/anime/abc123XYZ/title", "abc123XYZ"},
		{"allanime bangumi url extracts id segment", "https://allanime.to/bangumi/xyz789AB/some-title", "xyz789AB"},
		{"non-allanime", "https://example.com/x", "https://example.com/x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractAllAnimeID(tt.url))
		})
	}
}

func TestAllAnimeNavigator_GetNextEpisode(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"1", "2", "3"}}

	next, err := nav.GetNextEpisode("1")
	require.NoError(t, err)
	assert.Equal(t, "2", next)

	_, err = nav.GetNextEpisode("3")
	assert.Error(t, err)

	_, err = nav.GetNextEpisode("not-a-number")
	assert.Error(t, err)
}

func TestAllAnimeNavigator_GetPreviousEpisode(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"1", "2", "3"}}

	prev, err := nav.GetPreviousEpisode("2")
	require.NoError(t, err)
	assert.Equal(t, "1", prev)

	_, err = nav.GetPreviousEpisode("1")
	assert.Error(t, err)

	_, err = nav.GetPreviousEpisode("bad")
	assert.Error(t, err)
}

func TestAllAnimeNavigator_GetTotalEpisodes(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{episodes: []string{"a", "b", "c", "d"}}
	assert.Equal(t, 4, nav.GetTotalEpisodes())
}

func TestAllAnimeNavigator_ListAllEpisodes(t *testing.T) {
	t.Parallel()
	// Must return the real episode numbers from the source, not fabricated
	// 1..N values that lie for lists starting at "0" or containing specials.
	nav := &AllAnimeNavigator{episodes: []string{"0", "1", "5.5"}}
	list := nav.ListAllEpisodes()
	assert.Equal(t, []string{"0", "1", "5.5"}, list)
}

func TestAllAnimeNavigator_GetNextEpisode_ZeroBasedList(t *testing.T) {
	t.Parallel()
	// List starts at "0": 3 entries but last real episode is "2". The old
	// len()-based check accepted phantom episode "3" (3 <= len(3)).
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"0", "1", "2"}}

	next, err := nav.GetNextEpisode("1")
	require.NoError(t, err)
	assert.Equal(t, "2", next)

	_, err = nav.GetNextEpisode("2")
	assert.Error(t, err, "episode 3 does not exist in a 0-based list of 3 entries")
}

func TestAllAnimeNavigator_GetPreviousEpisode_ZeroBasedList(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"0", "1", "2"}}

	// Episode "0" exists in the list, so previous from "1" must reach it.
	prev, err := nav.GetPreviousEpisode("1")
	require.NoError(t, err)
	assert.Equal(t, "0", prev)

	_, err = nav.GetPreviousEpisode("0")
	assert.Error(t, err)
}

func TestNewAllAnimeNavigator_RejectsNonAllAnime(t *testing.T) {
	t.Parallel()
	_, err := NewAllAnimeNavigator(&models.Anime{Source: "AnimeFire", URL: "https://animefire.io/x"})
	assert.Error(t, err)
}

func TestHandleAllAnimeEpisodeNavigation_InvalidDirection(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "shortid12"}
	// Seed the navigator cache so we don't hit the network.
	nav := &AllAnimeNavigator{animeID: "shortid12", episodes: []string{"1", "2"}}
	navigatorCacheMu.Lock()
	navigatorCache["shortid12"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "shortid12")
		navigatorCacheMu.Unlock()
	})

	_, err := HandleAllAnimeEpisodeNavigation(anime, "1", "sideways")
	assert.Error(t, err)
}
