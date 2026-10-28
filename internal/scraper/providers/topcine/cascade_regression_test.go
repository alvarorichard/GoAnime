package topcine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. The TopCine catalog itself is healthy;
// what breaks users is downstream (panel coverage, flattened seasons), so this
// cascade pins search → title → languages end to end against one stub, proving
// the stages stay wired for every page shape the catalog serves.

// dexterLikePage is a multi-season series with per-episode players, like
// /serie/dexter live (96 players across 8 seasons, TMDB 1405). playerBase is
// the stub's own base URL so Languages stays hermetic (same scheme+host).
func dexterLikePage(playerBase string, tmdb int, seasons map[int]int) string {
	var b strings.Builder
	b.WriteString("<html><body>")
	for s := 1; s <= len(seasons); s++ {
		n := seasons[s]
		b.WriteString("<select>")
		for e := 1; e <= n; e++ {
			fmt.Fprintf(&b, `<option data-temporada="%d" data-episodio="%d" data-titulo="T%dE%d" data-player="%s/serie/%d/%d/%d">EP %02d</option>`,
				s, e, s, e, playerBase, tmdb, s, e, e)
		}
		b.WriteString("</select>")
	}
	b.WriteString("</body></html>")
	return b.String()
}

// soloLikePage flattens every episode into season 1, like /serie/solo-leveling
// live (25 players under /serie/127532/1/* while the panel splits S1/S2).
func soloLikePage(playerBase string, tmdb, total int) string {
	var b strings.Builder
	b.WriteString("<html><body><select>")
	for e := 1; e <= total; e++ {
		fmt.Fprintf(&b, `<option data-temporada="1" data-episodio="%d" data-titulo="Flat %d" data-player="%s/serie/%d/1/%d">EP %02d</option>`,
			e, e, playerBase, tmdb, e, e)
	}
	b.WriteString("</select></body></html>")
	return b.String()
}

func topcineCascadeStub(t *testing.T) *httptest.Server {
	t.Helper()
	pages := map[string]string{
		"/filme/film":         `<html><body><iframe src="https://azullog.test/filme/555"></iframe></body></html>`,
		"/serie/noplayer":     `<html><body><p>Em breve</p></body></html>`,
		"/serie/player-embed": `<html><body><iframe src="https://azullog.test/filme/0"></iframe></body></html>`,
	}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/buscar":
			_, _ = fmt.Fprint(w, `<html><body>
				<article class="pagina-busca-card"><a href="/serie/dexter"><img src="https://image.tmdb.org/t/p/w342/d.jpg"><h2>Dexter</h2><small>2006 • Série</small></a></article>
				<article class="pagina-busca-card"><a href="/serie/flat"><img src="https://image.tmdb.org/t/p/w342/f.jpg"><h2>Flat Show</h2><small>2024 • Série</small></a></article>
				<article class="pagina-busca-card"><a href="/filme/film"><img src="https://image.tmdb.org/t/p/w342/m.jpg"><h2>Film</h2><small>2024 • Filme</small></a></article>
			</body></html>`)
		case strings.HasSuffix(r.URL.Path, "carregar_idiomas") || r.URL.Query().Get("acao") == "carregar_idiomas":
			if strings.Contains(r.URL.Path, "/refused/") {
				_, _ = fmt.Fprint(w, `{"sucesso":false}`)
				return
			}
			_, _ = fmt.Fprintf(w, `{"sucesso":true,"tem_dublado":true,"tem_legendado":true,"url_dublado":"%s/e/dub","url_legendado":"%s/e/leg"}`,
				srv.URL, srv.URL)
		default:
			switch r.URL.Path {
			case "/serie/dexter":
				_, _ = fmt.Fprint(w, dexterLikePage(srv.URL, 1405, map[int]int{1: 12, 2: 12}))
				return
			case "/serie/flat":
				_, _ = fmt.Fprint(w, soloLikePage(srv.URL, 127532, 5))
				return
			}
			if page, ok := pages[r.URL.Path]; ok {
				_, _ = fmt.Fprint(w, page)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func stubHost(raw string) string {
	raw = strings.TrimPrefix(raw, "https://")
	return strings.TrimPrefix(raw, "http://")
}

func cascadeClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return NewClientForTest(srv.Client(), strings.TrimRight(srv.URL, "/"))
}

// TestTopCineCascade_SearchToLanguages is the happy path across all three
// client stages for a multi-season series.
func TestTopCineCascade_SearchToLanguages(t *testing.T) {
	t.Parallel()

	srv := topcineCascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	results, err := c.Search(ctx, "dexter")
	require.NoError(t, err)
	require.Len(t, results, 3)

	var dexter *Media
	for i := range results {
		if results[i].Title == "Dexter" {
			dexter = &results[i]
		}
	}
	require.NotNil(t, dexter, "Dexter must be in the results")
	assert.Equal(t, KindSeries, dexter.Kind)

	title, err := c.Title(ctx, srv.URL+"/serie/dexter")
	require.NoError(t, err)
	assert.Equal(t, KindSeries, title.Kind)
	assert.Equal(t, 1405, title.TMDBID)
	require.Len(t, title.Seasons, 2)
	assert.Len(t, title.Seasons[0].Episodes, 12)
	assert.Equal(t, "T1E1", title.Seasons[0].Episodes[0].Title)

	langs, err := c.Languages(ctx, title.Seasons[0].Episodes[0].PlayerURL)
	require.NoError(t, err)
	assert.True(t, langs.Dubbed)
	assert.True(t, langs.Subtitled)
	require.Len(t, langs.Hosts, 1)
	assert.Equal(t, stubHost(srv.URL), langs.Hosts[0])
}

// TestTopCineCascade_FlattenedSeason documents the Solo Leveling live shape:
// every player hangs under season 1. The client reports the page faithfully
// (one season); the api layer maps it onto the panel's split seasons.
func TestTopCineCascade_FlattenedSeason(t *testing.T) {
	t.Parallel()

	srv := topcineCascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	title, err := c.Title(ctx, srv.URL+"/serie/flat")
	require.NoError(t, err)
	require.Len(t, title.Seasons, 1, "the flattened page names a single season")
	assert.Len(t, title.Seasons[0].Episodes, 5)
	assert.Equal(t, "Flat 5", title.EpisodeTitle(1, 5))
	assert.Empty(t, title.EpisodeTitle(2, 1), "there is no season 2 on the page itself")
}

// TestTopCineCascade_MovieTitle reads the TMDB id off the iframed player.
func TestTopCineCascade_MovieTitle(t *testing.T) {
	t.Parallel()

	srv := topcineCascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	title, err := c.Title(ctx, srv.URL+"/filme/film")
	require.NoError(t, err)
	assert.Equal(t, KindMovie, title.Kind)
	assert.Equal(t, 555, title.TMDBID)
	assert.NotEmpty(t, title.PlayerURL)
}

// TestTopCineCascade_FailurePropagation checks each stage's failure surfaces
// as the documented sentinel, not an empty success.
func TestTopCineCascade_FailurePropagation(t *testing.T) {
	t.Parallel()

	srv := topcineCascadeStub(t)
	c := cascadeClient(t, srv)
	ctx := context.Background()

	_, err := c.Title(ctx, srv.URL+"/serie/noplayer")
	require.ErrorIs(t, err, ErrNoPlayer, "a page with no player names no TMDB id")

	_, err = c.Title(ctx, srv.URL+"/serie/player-embed")
	require.ErrorIs(t, err, ErrNoPlayer, "an iframe with TMDB id 0 is not a player")

	_, err = c.Languages(ctx, srv.URL+"/refused/1/1/1")
	require.Error(t, err, "a player refusing the language request must fail")

	_, err = c.Languages(ctx, "/relativo/1")
	require.Error(t, err, "a relative player URL must fail")

	_, err = c.Search(ctx, "  ")
	require.NoError(t, err, "a blank query sends no request")
}
