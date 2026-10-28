package hianime

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the tests.
//
// A green suite proves nothing on its own: a parser that quietly returns empty
// results, or fixtures that no longer resemble what hianime.at sends, both leave
// every assertion passing. That is exactly how the AllAnime source rotted, and
// how anidb.app reached a site-wide 503 with nobody noticing. Two guards close
// the gap.
//
//   - The mutation table below breaks one thing in a fixture at a time and
//     requires the client to notice. If a mutation still passes, the
//     corresponding test in client_test.go is vacuous.
//   - TestFixturesMatchLiveShape (live-gated) compares the fixtures against the
//     real endpoints, so a silent upstream change is caught instead of being
//     masked by stale fixtures.

// mutationServer serves one full chain built from the (possibly mutated)
// fixtures it is given.
type mutationServer struct {
	searchDoc string
	episodes  string
	servers   string // %s twice: the base64'd sub and dub embed URLs
	embedCfg  string // %s twice: base URL and audio track
	embedKey  string // the XOR key the embed page is written with
	embedVar  string // the JS global the payload is assigned to
	master    string
	srv       *httptest.Server
}

func newMutationServer(t *testing.T, m mutationServer) *mutationServer {
	t.Helper()
	s := &m
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(s.searchDoc))

		case strings.HasPrefix(r.URL.Path, "/api/theme/episode/list/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(s.episodes))

		case r.URL.Path == "/api/theme/episode/servers":
			b64 := func(v string) string { return base64.StdEncoding.EncodeToString([]byte(v)) }
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, s.servers, b64(s.srv.URL+"/embed/sub"), b64(s.srv.URL+"/embed/dub"))

		case strings.HasPrefix(r.URL.Path, "/embed/"):
			audio := strings.TrimPrefix(r.URL.Path, "/embed/")
			config := fmt.Sprintf(s.embedCfg, s.srv.URL, audio)
			raw := []byte(config)
			key := []byte(s.embedKey)
			for i := range raw {
				raw[i] ^= key[i%len(key)]
			}
			fmt.Fprintf(w, `<html><script>%s="%s"</script></html>`,
				s.embedVar, base64.StdEncoding.EncodeToString(raw))

		case strings.HasSuffix(r.URL.Path, "master.m3u8"):
			_, _ = w.Write([]byte(s.master))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

// pristine returns the unmutated fixtures shared with client_test.go.
func pristine() mutationServer {
	return mutationServer{
		searchDoc: searchPage,
		episodes:  episodesFixture,
		servers:   serversFixture,
		embedCfg: `{"src":"%s/stream/%s/master.m3u8","subtitles":` +
			`[{"lang":"en","label":"English","src":"https://cdn.example/en.vtt"}]}`,
		embedKey: obfuscationKey,
		embedVar: "window.__P",
		master:   masterPlaylist,
	}
}

// exercise runs the whole chain and returns the first error, or nil.
func exercise(t *testing.T, s *mutationServer) error {
	t.Helper()
	c := NewClientForTest(s.srv.URL)
	ctx := context.Background()

	results, err := c.SearchAnime(ctx, "cowboy bebop")
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("search: no results")
	}

	eps, err := c.GetAnimeEpisodes(ctx, results[0].URL)
	if err != nil {
		return fmt.Errorf("episodes: %w", err)
	}
	if len(eps) == 0 {
		return fmt.Errorf("episodes: none listed")
	}

	streamURL, _, err := c.GetEpisodeStreamURL(ctx, eps[0].URL, "720p")
	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	if !strings.Contains(streamURL, ".m3u8") {
		return fmt.Errorf("stream: not a playlist: %s", streamURL)
	}
	return nil
}

// TestPristineChainSucceeds is the control: without it, a mutation test that
// "caught" a break might just be failing for an unrelated reason.
func TestPristineChainSucceeds(t *testing.T) {
	t.Parallel()
	require.NoError(t, exercise(t, newMutationServer(t, pristine())))
}

// TestOneBreakIsAlwaysNoticed mutates a single thing per case and requires the
// chain to fail. A case that passes means the matching test asserts nothing.
func TestOneBreakIsAlwaysNoticed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*mutationServer)
	}{
		{
			name:   "search cards are renamed",
			mutate: func(m *mutationServer) { m.searchDoc = strings.ReplaceAll(m.searchDoc, "flw-item", "card") },
		},
		{
			// Both cards, not one: a mutation that leaves a parseable result
			// behind proves nothing, because the chain just uses the other one.
			name: "search results lose their permalink",
			mutate: func(m *mutationServer) {
				m.searchDoc = strings.ReplaceAll(m.searchDoc, "-26\"", "\"")
				m.searchDoc = strings.ReplaceAll(m.searchDoc, "-17\"", "\"")
			},
		},
		{
			name: "the episode envelope stops saying status",
			mutate: func(m *mutationServer) {
				m.episodes = strings.Replace(m.episodes, `"status":true`, `"status":false`, 1)
			},
		},
		{
			name:   "episode entries are renamed",
			mutate: func(m *mutationServer) { m.episodes = strings.ReplaceAll(m.episodes, "ep-item", "episode") },
		},
		{
			name:   "episode entries lose their id",
			mutate: func(m *mutationServer) { m.episodes = strings.ReplaceAll(m.episodes, `data-id=`, `data-episode=`) },
		},
		{
			name:   "episode entries lose their number",
			mutate: func(m *mutationServer) { m.episodes = strings.ReplaceAll(m.episodes, `data-number=`, `data-n=`) },
		},
		{
			name:   "server entries are renamed",
			mutate: func(m *mutationServer) { m.servers = strings.ReplaceAll(m.servers, "server-item", "srv") },
		},
		{
			name:   "the readable server disappears",
			mutate: func(m *mutationServer) { m.servers = strings.ReplaceAll(m.servers, "ZokoAnime", "Vidstream-2") },
		},
		{
			name:   "the embed payload moves to another global",
			mutate: func(m *mutationServer) { m.embedVar = "window.__CONFIG" },
		},
		{
			name:   "the embed obfuscation is rekeyed",
			mutate: func(m *mutationServer) { m.embedKey = "otaku-embed-v2" },
		},
		{
			name: "the embed config stops carrying a stream",
			mutate: func(m *mutationServer) {
				m.embedCfg = `{"stream":"%s/x/%s/master.m3u8","subtitles":[]}`
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

// A dropped RESOLUTION is deliberately NOT in the table above: the client is
// supposed to fall back to the master playlist rather than fail. This pins that
// it is a fallback and not an accident.
func TestUnreadableVariantsFallBackInsteadOfFailing(t *testing.T) {
	t.Parallel()
	m := pristine()
	m.master = strings.ReplaceAll(m.master, "RESOLUTION=1280x720", "BANDWIDTH=809659")
	s := newMutationServer(t, m)

	c := NewClientForTest(s.srv.URL)
	got, _, err := c.GetEpisodeStreamURL(context.Background(),
		s.srv.URL+"/watch/cowboy-bebop-26?ep=9001", "720p")
	require.NoError(t, err)
	assert.Equal(t, s.srv.URL+"/stream/sub/master.m3u8", got)
}

// TestFixturesMatchLiveShape checks the fixtures against the real site, so a
// silent upstream change surfaces here instead of as "no results" in the app.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestFixturesMatchLiveShape(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	c := NewHiAnimeClient()
	ctx := context.Background()

	body, err := c.getBody(ctx, c.baseURL+"/search?keyword=cowboy+bebop", "search", nil)
	require.NoError(t, err)
	for _, marker := range []string{"flw-item", "film-name", "film-poster-img"} {
		assert.Contains(t, string(body), marker,
			"the search fixture encodes %q; the live page no longer has it", marker)
	}

	results, err := c.SearchAnime(ctx, "cowboy bebop")
	require.NoError(t, err)
	require.NotEmpty(t, results, "the live search parsed to nothing — the fixture is stale")

	animeID, err := AnimeID(results[0].URL)
	require.NoError(t, err)
	list, err := c.getFragment(ctx, c.restURL("episode/list/"+animeID), "episodes", c.baseURL+"/")
	require.NoError(t, err)
	for _, marker := range []string{"ep-item", "data-id=", "data-number="} {
		assert.Contains(t, list, marker,
			"the episode fixture encodes %q; the live fragment no longer has it", marker)
	}

	eps, err := c.GetAnimeEpisodes(ctx, results[0].URL)
	require.NoError(t, err)
	require.NotEmpty(t, eps)

	epID, err := EpisodeID(eps[0].URL)
	require.NoError(t, err)
	servers, err := c.getFragment(ctx, c.restURL("episode/servers?episodeId="+epID), "servers", eps[0].URL)
	require.NoError(t, err)
	for _, marker := range []string{"server-item", "data-hash=", "data-type=", zokoServerName} {
		assert.Contains(t, servers, marker,
			"the server fixture encodes %q; the live fragment no longer has it", marker)
	}

	parsed, err := parseServers(servers)
	require.NoError(t, err)
	chosen, ok := selectServer(parsed, audioSub)
	require.True(t, ok, "no readable server live: %+v", parsed)
	assert.True(t, strings.HasPrefix(chosen.Embed, "http"),
		"data-hash must still be plain base64 of a URL; got %q", chosen.Embed)

	page, err := c.getBody(ctx, chosen.Embed, "embed", nil)
	require.NoError(t, err)
	assert.Contains(t, string(page), "window.__P",
		"the embed fixture encodes window.__P; the live player no longer uses it")
	payload, err := decodePayload(page)
	require.NoError(t, err, "the live player no longer decodes with %q", obfuscationKey)
	assert.Contains(t, payload.Src, ".m3u8")
}
