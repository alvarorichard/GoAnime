package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catalogHandler serves one series (panel 55: season 1 in both audios, season 2
// dubbed only), one movie (panel tt1: a dub and a sub player side by side) and
// a movie whose only player is the dub.
func catalogHandler(w http.ResponseWriter, r *http.Request) {
	player := func(src, label string) string {
		return fmt.Sprintf(`<button data-show-player="true" data-source="%s" data-type="jwplayer" data-id="1">%s</button>`, src, label)
	}
	switch {
	case strings.HasSuffix(r.URL.Path, ".mp4"):
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusPartialContent)
	case r.URL.Path == "/series/x/":
		fmt.Fprint(w, `<iframe src="https://painel.test/embed/55"></iframe>`)
	case r.URL.Path == "/filmes/m/":
		fmt.Fprint(w, `<iframe src="https://painel.test/filme/tt1"></iframe>`)
	case r.URL.Path == "/filmes/dubonly/":
		fmt.Fprint(w, `<iframe src="https://painel.test/filme/tt2"></iframe>`)
	case r.URL.Path == "/embed/55":
		fmt.Fprint(w, `<ul class="header-navigation">
			<li data-season-id="s1" data-season-number="1">1</li>
			<li data-season-id="s2" data-season-number="2">2</li></ul>
			<div class="cards">
			<div class="card"><h2 class="card-title">Dublado</h2><ul>
				<li data-season-id="s1" data-episode-id="10"><a>1 - Piloto</a></li>
				<li data-season-id="s1" data-episode-id="11"><a>2 - Segundo</a></li>
				<li data-season-id="s2" data-episode-id="20"><a>1 - Volta</a></li></ul></div>
			<div class="card"><h2 class="card-title">Legendado</h2><ul>
				<li data-season-id="s1" data-episode-id="30"><a>1 - Piloto</a></li></ul></div></div>`)
	case r.URL.Path == "/filme/tt1":
		fmt.Fprint(w, player("https://files.test/dub.mp4", "Player #1 Dublado")+player("https://files.test/sub.mp4", "Player #2 Legendado"))
	case r.URL.Path == "/filme/tt2":
		fmt.Fprint(w, player("https://files.test/dub.mp4", "Player #1 Dublado"))
	case strings.HasPrefix(r.URL.Path, "/episodio/"):
		fmt.Fprint(w, player("https://files.test/ep"+strings.TrimPrefix(r.URL.Path, "/episodio/")+".mp4", "Player #1"))
	default:
		http.NotFound(w, r)
	}
}

func catalogAdapter(t *testing.T) *StartFlixAdapter {
	t.Helper()
	return newStartFlixTestAdapter(t, catalogHandler)
}

const (
	seriesPage = "https://www.startflix.test/series/x/"
	moviePage  = "https://www.startflix.test/filmes/m/"
)

func TestStartFlixCatalogSeasons(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	got, err := a.CatalogSeasons(context.Background(), seriesPage)
	require.NoError(t, err)
	assert.Equal(t, []CatalogSeason{
		{Number: 1, Audios: []string{"Dublado", "Legendado"}},
		{Number: 2, Audios: []string{"Dublado"}},
	}, got)

	none, err := a.CatalogSeasons(context.Background(), moviePage)
	require.NoError(t, err)
	assert.Empty(t, none, "a movie has no seasons")
}

func TestStartFlixCatalogEpisodes(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	ctx := context.Background()

	movie, err := a.CatalogEpisodes(ctx, moviePage, 0, "")
	require.NoError(t, err)
	require.Len(t, movie, 1)
	assert.Equal(t, "https://painel.test/filme/tt1", movie[0].URL)
	assert.Equal(t, 1, movie[0].Num)

	_, err = a.CatalogEpisodes(ctx, seriesPage, 0, "")
	require.ErrorIs(t, err, ErrSeasonRequired)

	dub, err := a.CatalogEpisodes(ctx, seriesPage, 1, "")
	require.NoError(t, err)
	require.Len(t, dub, 2, "no audio takes the dub")
	assert.Equal(t, "https://painel.test/episodio/10", dub[0].URL)
	assert.Equal(t, "Piloto", dub[0].Title.English)
	assert.Equal(t, "1", dub[0].SeasonID)

	sub, err := a.CatalogEpisodes(ctx, seriesPage, 1, "legendado")
	require.NoError(t, err)
	require.Len(t, sub, 1)
	assert.Equal(t, "https://painel.test/episodio/30", sub[0].URL)

	_, err = a.CatalogEpisodes(ctx, seriesPage, 2, "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable)
	assert.Contains(t, err.Error(), "Dublado", "the error says what the season has")

	_, err = a.CatalogEpisodes(ctx, seriesPage, 9, "")
	require.ErrorIs(t, err, ErrSeasonNotFound)
	assert.Contains(t, err.Error(), "1, 2")

	_, err = a.CatalogEpisodes(ctx, seriesPage, 1, "klingon")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

func TestStartFlixCatalogStreamAudio(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	ctx := context.Background()

	u, meta, err := a.CatalogStream(ctx, "https://painel.test/filme/tt1", "Legendado")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/sub.mp4", u)
	assert.Equal(t, "Legendado", meta["audio"])
	assert.Equal(t, "startflix", meta["source"])

	u, _, err = a.CatalogStream(ctx, "https://painel.test/filme/tt1", "dub")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/dub.mp4", u)

	_, _, err = a.CatalogStream(ctx, "https://painel.test/filme/tt2", "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable, "the dub is never played in the sub's place")
	assert.Contains(t, err.Error(), "Dublado")

	u, _, err = a.CatalogStream(ctx, "https://painel.test/episodio/10", "")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep10.mp4", u)

	_, _, err = a.CatalogStream(ctx, "https://painel.test/filme/tt1", "klingon")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

func TestParseCatalogAudio(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]startflix.Audio{
		"": "", " Dublado ": startflix.AudioDubbed, "DUB": startflix.AudioDubbed, "dubbed": startflix.AudioDubbed,
		"legendado": startflix.AudioSubtitled, "Sub": startflix.AudioSubtitled, "subtitled": startflix.AudioSubtitled,
	} {
		got, err := ParseCatalogAudio(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err := ParseCatalogAudio("dublegendado")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

// The adapters other than StartFlix are not catalogues yet.
func TestOnlyStartFlixIsACatalogScraper(t *testing.T) {
	t.Parallel()
	for _, st := range []ScraperType{AnimefireType, HiAnimeType, TopCineType, PobreflixType, PomfyType} {
		a, err := NewAdapter(st)
		require.NoError(t, err)
		_, ok := a.(CatalogScraper)
		assert.False(t, ok, "%v", st)
	}
	a, err := NewAdapter(StartFlixType)
	require.NoError(t, err)
	_, ok := a.(CatalogScraper)
	assert.True(t, ok)
}
