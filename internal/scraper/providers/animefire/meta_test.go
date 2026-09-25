package animefire

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the tests.
//
// AnimeFire is the source that proved a green suite proves nothing. The site
// was rewritten as a single-page app and moved host; the scraper kept parsing a
// 13KB shell, found zero results, and reported (nothing, nil) — which reads
// exactly like "no anime matched". It stayed broken for weeks with every test
// passing, because the tests asserted against fixtures nobody had checked
// against the live site since the day they were written.
//
// Unit tests cannot catch that on their own, so two guards sit on top of them:
//
//   - TestOneBreakIsAlwaysNoticed mutates one field of one fixture at a time
//     and requires the client to fail. A mutation that still passes means the
//     test covering that field asserts nothing.
//   - TestFixturesMatchLiveShape (live-gated) checks the fixtures against the
//     real API, so an upstream change surfaces here instead of as silence in
//     the app.

// ── The canonical payloads ──────────────────────────────────────────────────
//
// Captured from api.animefire.one. Every mutation below is a single edit to
// one of these.

func pristineSearch() map[string]any {
	return map[string]any{"data": []any{map[string]any{
		"id":         "one-piece",
		"titles":     map[string]any{"BR": "One Piece", "US": "One Piece"},
		"audio":      "Legendado",
		"poster_src": "https://cdn.animefire.one/one-piece.jpg",
		"status":     "em lançamento",
	}}}
}

func pristineAnime(streamable bool) map[string]any {
	eps := make([]any, 0, 4)
	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("one-piece-%d", i)
		if !streamable {
			id = ""
		}
		eps = append(eps, map[string]any{
			"id": id, "title": fmt.Sprintf("Episódio %d", i),
			"season": 1, "number": i, "synopsis": "...",
		})
	}
	return map[string]any{"data": map[string]any{"episodes": eps}}
}

func pristineEpisode() map[string]any {
	return map[string]any{"data": map[string]any{
		"id": "one-piece-1", "title": "Episódio 1", "number": 1,
		"streams": []any{map[string]any{
			"audio": "Legendado", "is_offline": false,
			"url":       "https://s1.animefire.one/one-piece/1/h.jpg",
			"qualities": []any{"720p"},
		}},
	}}
}

// mutationServer serves one full chain from the (possibly mutated) payloads.
type mutationServer struct {
	search, anime, episode map[string]any
	srv                    *httptest.Server
}

func newMutationServer(t *testing.T, m mutationServer) *mutationServer {
	t.Helper()
	s := &m
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/animes/pesquisar"):
			_ = json.NewEncoder(w).Encode(s.search)
		case strings.HasPrefix(r.URL.Path, "/anime/"):
			_ = json.NewEncoder(w).Encode(s.anime)
		case strings.HasPrefix(r.URL.Path, "/episode/"):
			_ = json.NewEncoder(w).Encode(s.episode)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func pristine() mutationServer {
	return mutationServer{
		search:  pristineSearch(),
		anime:   pristineAnime(true),
		episode: pristineEpisode(),
	}
}

// exercise walks search → episodes → stream and returns the first failure.
//
// "Nothing found" counts as a failure here on purpose. That is the entire point
// of this file: a source that silently returns an empty list is broken, and a
// guard that accepted it would be reproducing the original bug.
func exercise(t *testing.T, s *mutationServer) error {
	t.Helper()
	c := NewClientForTest(s.srv.URL)

	results, err := c.searchAPI("one piece")
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("search: no results — an empty answer is how this source broke silently")
	}

	eps, err := c.episodesAPI(results[0].URL)
	if err != nil {
		return fmt.Errorf("episodes: %w", err)
	}
	if len(eps) == 0 {
		return fmt.Errorf("episodes: none listed")
	}

	streamURL, err := c.streamAPI(eps[0].URL)
	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	if streamURL == "" {
		return fmt.Errorf("stream: empty URL")
	}
	return nil
}

// TestPristineChainSucceeds is the control. Without it, a mutation that
// "caught" a break might just be failing for an unrelated reason.
func TestPristineChainSucceeds(t *testing.T) {
	t.Parallel()
	require.NoError(t, exercise(t, newMutationServer(t, pristine())))
}

// TestOneBreakIsAlwaysNoticed breaks one thing per case and requires the chain
// to fail. A case that passes means the matching assertion elsewhere is
// vacuous.
func TestOneBreakIsAlwaysNoticed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*mutationServer)
	}{
		{
			name: "the search envelope is renamed",
			mutate: func(m *mutationServer) {
				m.search = map[string]any{"results": m.search["data"]}
			},
		},
		{
			name: "search results lose their id",
			mutate: func(m *mutationServer) {
				for _, a := range m.search["data"].([]any) {
					a.(map[string]any)["id"] = ""
				}
			},
		},
		{
			name: "search results lose every title",
			mutate: func(m *mutationServer) {
				for _, a := range m.search["data"].([]any) {
					a.(map[string]any)["titles"] = map[string]any{}
				}
			},
		},
		{
			name: "the episode list is renamed",
			mutate: func(m *mutationServer) {
				d := m.anime["data"].(map[string]any)
				d["capitulos"] = d["episodes"]
				delete(d, "episodes")
			},
		},
		{
			name:   "episodes lose their id",
			mutate: func(m *mutationServer) { m.anime = pristineAnime(false) },
		},
		{
			name: "the stream list is renamed",
			mutate: func(m *mutationServer) {
				d := m.episode["data"].(map[string]any)
				d["fontes"] = d["streams"]
				delete(d, "streams")
			},
		},
		{
			name: "the stream loses its url",
			mutate: func(m *mutationServer) {
				for _, s := range m.episode["data"].(map[string]any)["streams"].([]any) {
					s.(map[string]any)["url"] = ""
				}
			},
		},
		{
			// The shape the original bug wore: a well-formed answer carrying
			// nothing. The app read that as "no anime matched" and dropped the
			// source from the fan-out without a word, which is why exercise()
			// treats an empty result as a failure rather than a valid outcome.
			name: "the API answers with an empty catalogue",
			mutate: func(m *mutationServer) {
				m.search = map[string]any{"data": []any{}}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := pristine()
			tt.mutate(&m)
			require.Error(t, exercise(t, newMutationServer(t, m)),
				"this break went unnoticed, so the test covering it asserts nothing")
		})
	}
}

// An offline episode is content state, not a scraping failure, and has to stay
// distinguishable from one — otherwise the player blames itself and the user is
// told to try again on a title that will never work.
func TestOfflineStaysDistinguishableFromBroken(t *testing.T) {
	t.Parallel()
	m := pristine()
	for _, s := range m.episode["data"].(map[string]any)["streams"].([]any) {
		s.(map[string]any)["is_offline"] = true
	}
	srv := newMutationServer(t, m)

	_, err := NewClientForTest(srv.srv.URL).streamAPI(srv.srv.URL + "/episode/one-piece-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, netx.ErrMediaOffline,
		"an offline episode must classify as missing content, not as a parser failure")
}

// TestFixturesMatchLiveShape checks the payload shapes against the real API, so
// the next rewrite surfaces here instead of as silence in the app.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestFixturesMatchLiveShape(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	c := NewAnimefireClient()

	var search apiSearchResponse
	require.NoError(t, c.getJSON("/animes/pesquisar?q="+"one+piece", &search),
		"the live search endpoint no longer answers the shape we decode")
	require.NotEmpty(t, search.Data, "the live search parsed to nothing — the fixture is stale")
	first := search.Data[0]
	assert.NotEmpty(t, first.ID, "results no longer carry an id")
	assert.NotEmpty(t, first.title(), "results no longer carry a title in any language we read")

	results, err := c.searchAPI("one piece")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	eps, err := c.episodesAPI(results[0].URL)
	require.NoError(t, err, "the live episode listing no longer parses")
	require.NotEmpty(t, eps, "the live listing parsed to zero episodes")
	assert.NotEmpty(t, eps[0].URL)

	// Not require: a title can legitimately be offline. What must hold is that
	// it fails as CONTENT, not as a parser error.
	streamURL, err := c.streamAPI(eps[0].URL)
	if err != nil {
		assert.ErrorIs(t, err, netx.ErrMediaOffline,
			"a live stream lookup failed in a way that is not 'offline': %v", err)
		return
	}
	assert.Contains(t, streamURL, "http")
	_ = context.Background()
}
