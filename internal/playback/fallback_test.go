package playback

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AnimeFire is unstable at times (issue #229): when an episode will not load
// on the source the user picked, the same episode of the same title is looked
// for on the others. These tests run that cascade against stand-ins for the
// sources. Not parallel: the lookups are package-level seams.

// TestMain keeps every test off the network: the fallback finds nothing unless
// a test says what the other sources answer.
func TestMain(m *testing.M) {
	searchOnSource = func(context.Context, apisource.SourceKind, string) ([]*models.Anime, error) {
		return nil, errors.New("no network in tests")
	}
	os.Exit(m.Run())
}

// otherSources is what the stand-in sources answer.
type otherSources struct {
	titles   map[apisource.SourceKind][]*models.Anime
	episodes map[string][]models.Episode // by the candidate's URL
	videos   map[string]string           // by the candidate's URL
	searched []apisource.SourceKind
	resolved []string // URLs of the candidates videoURLFor was asked about
}

func stubOtherSources(t *testing.T, o *otherSources) {
	t.Helper()
	prevSearch, prevEpisodes, prevVideo := searchOnSource, episodesOnSource, videoURLFor
	searchOnSource = func(_ context.Context, kind apisource.SourceKind, _ string) ([]*models.Anime, error) {
		o.searched = append(o.searched, kind)
		return o.titles[kind], nil
	}
	episodesOnSource = func(_ context.Context, a *models.Anime) ([]models.Episode, error) {
		return o.episodes[a.URL], nil
	}
	videoURLFor = func(_ context.Context, _ *models.Episode, a *models.Anime) (string, error) {
		o.resolved = append(o.resolved, a.URL)
		if v, ok := o.videos[a.URL]; ok {
			return v, nil
		}
		return "", errors.New("the stream host timed out")
	}
	t.Cleanup(func() { searchOnSource, episodesOnSource, videoURLFor = prevSearch, prevEpisodes, prevVideo })
}

func windBreaker() *models.Anime {
	return &models.Anime{Name: "Wind Breaker", Source: "Animefire.io", URL: "https://animefire.one/anime/wind-breaker", MediaType: models.MediaTypeAnime}
}

func numbered(labels ...string) []models.Episode {
	eps := make([]models.Episode, len(labels))
	for i, l := range labels {
		eps[i] = models.Episode{Number: l, Num: i + 1, URL: "https://other/ep/" + l}
	}
	return eps
}

func TestFallbackKindsLeaveOutTheSourceThatFailedAndTheOnesTurnedOff(t *testing.T) {
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	t.Setenv("GOANIME_DISABLED_SOURCES", "animesdigital")

	kinds := fallbackKinds("Animefire.io")

	assert.NotContains(t, kinds, apisource.AnimeFire, "the source that just failed")
	assert.NotContains(t, kinds, apisource.AnimesDigital, "turned off")
	assert.Contains(t, kinds, apisource.Goyabu, "opted back in")
	assert.NotContains(t, kinds, apisource.StartFlix, "a movie and series catalogue")
}

func TestAnEpisodeThatWillNotLoadPlaysFromAnotherSource(t *testing.T) {
	o := &otherSources{
		titles: map[apisource.SourceKind][]*models.Anime{
			apisource.AnimesDigital: {{Name: "Wind Breaker [Dublado]", URL: "https://animesdigital/wb", Source: "AnimesDigital"}},
		},
		episodes: map[string][]models.Episode{"https://animesdigital/wb": numbered("Episódio 01", "Episódio 02")},
		videos:   map[string]string{"https://animesdigital/wb": "https://cdn/wb-ep2.m3u8"},
	}
	stubOtherSources(t, o)

	url, from, tried := videoFromAnotherSource(context.Background(), windBreaker(), "Episódio 02", 2)

	assert.Equal(t, "https://cdn/wb-ep2.m3u8", url)
	assert.Equal(t, "AnimesDigital", from)
	assert.Equal(t, []string{"AnimesDigital"}, tried)
	assert.Equal(t, []string{"https://animesdigital/wb"}, o.resolved, "the video is asked of the other source's entry, not of the one that failed")
}

func TestAnotherSourceIsTakenOnlyWhenItIsTheSameTitle(t *testing.T) {
	o := &otherSources{
		titles: map[apisource.SourceKind][]*models.Anime{
			apisource.AnimesDigital: {
				{Name: "Wind Breaker 2", URL: "https://animesdigital/wb2"},
				{Name: "Wind Breaker Movie", URL: "https://animesdigital/wbm"},
			},
		},
		episodes: map[string][]models.Episode{"https://animesdigital/wb2": numbered("1", "2"), "https://animesdigital/wbm": numbered("1")},
		videos:   map[string]string{"https://animesdigital/wb2": "https://cdn/wrong", "https://animesdigital/wbm": "https://cdn/wrong"},
	}
	stubOtherSources(t, o)

	url, _, tried := videoFromAnotherSource(context.Background(), windBreaker(), "1", 1)

	assert.Empty(t, url, "another season or a film is not this show")
	assert.NotEmpty(t, tried)
	assert.Empty(t, o.resolved, "nothing was resolved on a title that is not the one")
}

func TestASourceWithoutTheEpisodeIsSkipped(t *testing.T) {
	o := &otherSources{
		titles: map[apisource.SourceKind][]*models.Anime{
			apisource.AnimesDigital: {{Name: "Wind Breaker", URL: "https://animesdigital/wb"}},
			apisource.Anroll:        {{Name: "Wind Breaker", URL: "https://anroll/wb"}},
		},
		episodes: map[string][]models.Episode{
			"https://animesdigital/wb": numbered("1", "2", "3"), // stops short of episode 9
			"https://anroll/wb":        numbered("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		},
		videos: map[string]string{"https://animesdigital/wb": "https://cdn/wrong", "https://anroll/wb": "https://cdn/ep9"},
	}
	stubOtherSources(t, o)

	url, from, _ := videoFromAnotherSource(context.Background(), windBreaker(), "9", 9)

	assert.Equal(t, "https://cdn/ep9", url)
	assert.Equal(t, "AnimesROLL", from)
	assert.Equal(t, []string{"https://anroll/wb"}, o.resolved, "no video was asked for on the source that lacks the episode")
}

func TestASourceWhoseVideoFailsToLoadSendsTheSearchOn(t *testing.T) {
	o := &otherSources{
		titles: map[apisource.SourceKind][]*models.Anime{
			apisource.AnimesDigital: {{Name: "Wind Breaker", URL: "https://animesdigital/wb"}},
			apisource.Anroll:        {{Name: "Wind Breaker", URL: "https://anroll/wb"}},
		},
		episodes: map[string][]models.Episode{"https://animesdigital/wb": numbered("1"), "https://anroll/wb": numbered("1")},
		videos:   map[string]string{"https://anroll/wb": "https://cdn/ok"}, // AnimesDigital's host times out
	}
	stubOtherSources(t, o)

	url, from, tried := videoFromAnotherSource(context.Background(), windBreaker(), "1", 1)

	assert.Equal(t, "https://cdn/ok", url)
	assert.Equal(t, "AnimesROLL", from)
	assert.Equal(t, []string{"AnimesDigital", "AnimesROLL"}, tried, "in the order tried")
}

func TestTheSameAudioIsPreferred(t *testing.T) {
	dub := &models.Anime{Name: "Wind Breaker", Audio: "dub", URL: "https://animesdigital/dub"}
	sub := &models.Anime{Name: "Wind Breaker", Audio: "sub", URL: "https://animesdigital/sub"}

	assert.Same(t, sub, pickSameTitle([]*models.Anime{dub, sub}, "wind breaker", "sub"))
	assert.Same(t, dub, pickSameTitle([]*models.Anime{dub, sub}, "wind breaker", ""), "no preference: the first")
	assert.Same(t, dub, pickSameTitle([]*models.Anime{dub}, "wind breaker", "sub"), "the other audio beats nothing")
	assert.Nil(t, pickSameTitle([]*models.Anime{nil, {Name: "Another Show"}}, "wind breaker", "sub"))
}

func TestNothingIsTriedForAMovieOrWithoutATitle(t *testing.T) {
	o := &otherSources{}
	stubOtherSources(t, o)

	for name, a := range map[string]*models.Anime{
		"no anime":    nil,
		"a movie":     {Name: "Suzume", MediaType: models.MediaTypeMovie, Source: "Animefire.io"},
		"a series":    {Name: "Dark", MediaType: models.MediaTypeTV, Source: "Animefire.io"},
		"a blank one": {Name: " [PT-BR] ", Source: "Animefire.io"},
	} {
		url, _, tried := videoFromAnotherSource(context.Background(), a, "1", 1)
		assert.Empty(t, url, name)
		assert.Empty(t, tried, name)
	}
	assert.Empty(t, o.searched)
}

func TestTheUserLeavingIsNotASourceFailing(t *testing.T) {
	assert.True(t, worthTryingElsewhere(context.Background(), errors.New("the stream host timed out")))
	assert.True(t, worthTryingElsewhere(context.Background(), fmt.Errorf("%w: media offline", player.ErrSourceVideoRemoved)), "a deleted video is the case for another source")

	assert.False(t, worthTryingElsewhere(context.Background(), nil))
	for _, leaving := range []error{player.ErrBackToEpisodeSelection, player.ErrBackRequested, player.ErrBackToAnimeSelection, player.ErrUserQuit, context.Canceled} {
		assert.False(t, worthTryingElsewhere(context.Background(), fmt.Errorf("picking quality: %w", leaving)), leaving.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, worthTryingElsewhere(ctx, errors.New("x")), "the program is stopping")
}

func TestTitleKey(t *testing.T) {
	for in, want := range map[string]string{
		"Wind Breaker":                          "wind breaker",
		"Wind Breaker [PT-BR]":                  "wind breaker",
		"Wind Breaker (Dublado)":                "wind breaker",
		"Wind Breaker Dublado":                  "wind breaker",
		"Wind Breaker - Legendado":              "wind breaker",
		"Wind  Breaker: 2":                      "wind breaker 2",
		"Karakai Jouzu no Takagi-san 2":         "karakai jouzu no takagi san 2",
		"Karakai Jouzu no Takagi-san [HiAnime]": "karakai jouzu no takagi san",
		"":                                      "",
	} {
		assert.Equal(t, want, titleKey(in), in)
	}
	assert.NotEqual(t, titleKey("Wind Breaker"), titleKey("Wind Breaker 2"), "a season marker keeps it another entry")
	assert.False(t, sameTitle("anything", ""), "no wanted title matches nothing")
}

func TestFindEpisodeNumbered(t *testing.T) {
	eps := numbered("Episódio 01", "Episódio 02", "Episódio 03")
	eps = append(eps, models.Episode{Number: "Especial", Num: 4}, models.Episode{Number: "Ep. 10", Num: 5})

	find := func(label string, n int) string {
		if e := findEpisodeNumbered(eps, label, n); e != nil {
			return e.Number
		}
		return ""
	}
	assert.Equal(t, "Episódio 02", find("2", 2))
	assert.Equal(t, "Episódio 03", find("Episode 3", 3))
	assert.Equal(t, "Ep. 10", find("10", 10), "10 is not 1")
	assert.Equal(t, "Episódio 01", find("Episódio 01", 0), "the label stands in for a missing number")
	assert.Empty(t, find("99", 99))
	assert.Empty(t, find("Especial", 0), "a label with no number matches nothing")
	assert.Empty(t, find("", 0))
	assert.Equal(t, "Especial", find("4", 4), "an episode with no number in its label falls back to Num")
}

// ---- PlayEpisode asks the other sources before giving up -----------------------

func playWindBreaker(t *testing.T, primaryErr error) (said []string, err error) {
	t.Helper()
	notices := captureNotices(t)
	prevData := episodeDataFor
	episodeDataFor = func(int, int, *models.Anime) error { return nil }
	t.Cleanup(func() { episodeDataFor = prevData })
	prevVideo := videoURLFor
	videoURLFor = func(context.Context, *models.Episode, *models.Anime) (string, error) { return "", primaryErr }
	t.Cleanup(func() { videoURLFor = prevVideo })

	var paused bool
	var mu sync.Mutex
	eps := []models.Episode{{Number: "1", Num: 1, URL: "https://animefire.one/episode/1"}}
	err = PlayEpisode(context.Background(), windBreaker(), eps, 1, eps[0].URL, "1", false, &paused, &mu)
	return *notices, err
}

func countSearches(t *testing.T) *int {
	t.Helper()
	var n int
	prev := searchOnSource
	searchOnSource = func(context.Context, apisource.SourceKind, string) ([]*models.Anime, error) {
		n++
		return nil, nil
	}
	t.Cleanup(func() { searchOnSource = prev })
	return &n
}

func TestPlayEpisodeLooksOnTheOtherSourcesWhenTheEpisodeWillNotLoad(t *testing.T) {
	searches := countSearches(t)

	said, err := playWindBreaker(t, errors.New("the stream host timed out"))

	assert.Positive(t, *searches, "the other sources were asked")
	require.ErrorIs(t, err, player.ErrBackToEpisodeSelection, "and when none has it the user is told why, as before")
	require.Len(t, said, 1)
	assert.Contains(t, said[0], "the stream host timed out")
}

func TestPlayEpisodeDoesNotLookElsewhereWhenTheUserWentBack(t *testing.T) {
	searches := countSearches(t)

	_, err := playWindBreaker(t, player.ErrBackToEpisodeSelection)

	require.ErrorIs(t, err, player.ErrBackToEpisodeSelection)
	assert.Zero(t, *searches, "going back is not a source failing")
}
