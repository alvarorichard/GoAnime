package startflix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. StartFlix's catalog, panel and players are
// all healthy; what breaks users is content state (dead panels, emptied
// episodes, unsupported-only hosts), so this cascade pins search → panel →
// seasons → players → stream end to end against one stub, proving each stage
// surfaces its documented sentinel instead of an empty success.

// cascadeStub serves the whole chain. Media files under /files/ answer the
// Range probe; "dead" ones 404. /embed/999 is not on the panel (live: 404
// "Conteúdo não encontrado"). /episodio/empty is the emptied-episode shape
// (live: HTTP 200 with an empty chooser).
func cascadeStub(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/files/"):
			if strings.Contains(r.URL.Path, "dead") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		case r.URL.Path == "/" && r.URL.Query().Has("s"):
			_, _ = fmt.Fprint(w, `<html><body>
				<div class="result-item"><article><div class="details"><div class="title"><a href="https://www.startflix.test/series/show/">Show</a></div></div><div class="image"><img src="https://img.test/t/p/w92/s.jpg"></div><div class="meta"><span class="year">2024</span></div></article></div>
				<div class="result-item"><article><div class="details"><div class="title"><a href="https://www.startflix.test/filmes/film/">Film</a></div></div><div class="image"><img src="https://img.test/t/p/w92/f.jpg"></div><div class="meta"><span class="year">2024</span></div></article></div>
			</body></html>`)
			return
		case r.URL.Path == "/series/show/":
			_, _ = fmt.Fprintf(w, `<html><body><iframe src="%s/embed/100"></iframe></body></html>`, srv.URL)
			return
		case r.URL.Path == "/series/nopanel/":
			_, _ = fmt.Fprint(w, `<html><body><p>no iframe here</p></body></html>`)
			return
		case r.URL.Path == "/filmes/film/":
			_, _ = fmt.Fprintf(w, `<html><body><iframe src="%s/filme/tt100"></iframe></body></html>`, srv.URL)
			return
		case r.URL.Path == "/embed/100":
			_, _ = fmt.Fprint(w, `<html><body>
				<ul class="header-navigation">
					<li data-season-id="s1" data-season-number="1">1 TEMP</li>
					<li data-season-id="s2" data-season-number="2">2 TEMP</li>
				</ul>
				<div class="cards">
					<div class="card"><div class="card-header"><h2 class="card-title">Dublado</h2></div><div class="card-body"><ul>
						<li data-season-id="s1" data-episode-id="11"><a href="#s1_11">1 - Episódio</a></li>
						<li data-season-id="s1" data-episode-id="12"><a href="#s1_12">2 - Episódio</a></li>
						<li data-season-id="s2" data-episode-id="13"><a href="#s2_13">1 - Episódio</a></li>
					</ul></div></div>
					<div class="card"><div class="card-header"><h2 class="card-title">Legendado</h2></div><div class="card-body"><ul>
						<li data-season-id="s1" data-episode-id="21"><a href="#s1_21">1 - Piloto</a></li>
					</ul></div></div>
				</div>
			</body></html>`)
			return
		case r.URL.Path == "/embed/999":
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, "Conteúdo não encontrado.")
			return
		case r.URL.Path == "/episodio/11":
			_, _ = fmt.Fprintf(w, `
				<button data-show-player="true" data-source="%s/files/dead.mp4" data-type="jwplayer" data-id="1">Player #1</button>
				<button data-show-player="true" data-source="%s/files/ok.mp4" data-type="jwplayer" data-id="2">Player #2</button>`,
				srv.URL, srv.URL)
			return
		case r.URL.Path == "/episodio/empty":
			_, _ = fmt.Fprint(w, `<div class="play-banner"><div id="player"></div><div class="choose-player"><h1>SELECIONE UM PLAYER</h1><div id="players"></div></div></div>`)
			return
		case r.URL.Path == "/episodio/unsupported":
			_, _ = fmt.Fprint(w, `<button data-show-player="true" data-source="https://vidsrcme.su/x" data-type="iframe">P</button>`)
			return
		case r.URL.Path == "/filme/tt100":
			_, _ = fmt.Fprintf(w, `<button data-show-player="true" data-source="%s/files/film.mp4" data-type="jwplayer" data-subtitles="%s/subs/pt.vtt" data-id="9">Player #1</button>`,
				srv.URL, srv.URL)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func cascadeClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return NewClientForTest(srv.Client(), srv.URL)
}

// TestStartFlixCascade_SearchToStream walks the full series chain through the
// real client: search, title page, panel, seasons, players, then the stream,
// skipping the dead direct file.
func TestStartFlixCascade_SearchToStream(t *testing.T) {
	t.Parallel()

	srv := cascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	results, err := c.Search(ctx, "show")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, KindSeries, results[0].Kind)
	assert.Equal(t, KindMovie, results[1].Kind)

	panel, err := c.Panel(ctx, srv.URL+"/series/show/")
	require.NoError(t, err)
	assert.Equal(t, KindSeries, panel.Kind)
	assert.Equal(t, 100, panel.TMDBID)

	seasons, err := c.Seasons(ctx, panel)
	require.NoError(t, err)
	require.Len(t, seasons, 2)
	assert.Len(t, seasons[0].Dubbed, 2)
	assert.Len(t, seasons[0].Subtitled, 1)

	stream, err := c.Stream(ctx, panel.EpisodeURL("11"))
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/files/ok.mp4", stream.URL, "the dead file must be skipped")
	assert.Equal(t, srv.URL+"/", stream.Referer)
}

// TestStartFlixCascade_Movie walks the movie chain: the panel itself lists
// the players, with subtitles attached.
func TestStartFlixCascade_Movie(t *testing.T) {
	t.Parallel()

	srv := cascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	panel, err := c.Panel(ctx, srv.URL+"/filmes/film/")
	require.NoError(t, err)
	assert.Equal(t, KindMovie, panel.Kind)
	assert.Equal(t, "tt100", panel.IMDBID)

	stream, err := c.Stream(ctx, panel.URL)
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/files/film.mp4", stream.URL)
	require.Len(t, stream.Subtitles, 1)
	assert.Equal(t, srv.URL+"/subs/pt.vtt", stream.Subtitles[0].URL)
}

// TestStartFlixCascade_FailurePropagation pins the content-state sentinels:
// a title with no panel, a panel without the title, an emptied episode and
// an unsupported-only episode must each fail distinctly.
func TestStartFlixCascade_FailurePropagation(t *testing.T) {
	t.Parallel()

	srv := cascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	_, err := c.Panel(ctx, srv.URL+"/series/nopanel/")
	require.ErrorIs(t, err, ErrNoPanel)

	_, err = c.Seasons(ctx, Panel{URL: srv.URL + "/embed/999", Kind: KindSeries, TMDBID: 999})
	require.ErrorIs(t, err, ErrNotOnPanel)

	_, err = c.Players(ctx, srv.URL+"/episodio/empty")
	require.ErrorIs(t, err, ErrNoPlayers, "a 200 with an empty chooser is not a success")

	_, err = c.Stream(ctx, srv.URL+"/episodio/empty")
	require.ErrorIs(t, err, ErrNoPlayers)

	err = func() error {
		_, err := c.Stream(ctx, srv.URL+"/episodio/unsupported")
		return err
	}()
	require.ErrorIs(t, err, ErrNoSupportedServer)
	var noStream *NoStreamError
	require.True(t, errors.As(err, &noStream))
	assert.Contains(t, noStream.Unsupported, "vidsrcme.su")
}
