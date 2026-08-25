package anidb

import (
	"context"
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
// results, or fixtures that no longer resemble what anidb.app sends, both leave
// every assertion passing. Two guards close that gap.
//
//   - The mutation table below breaks one thing in a fixture at a time and
//     requires the client to notice. If a mutation still passes, the
//     corresponding test in client_test.go is vacuous.
//   - TestFixturesMatchLiveShape (live-gated) compares the fixtures against the
//     real endpoints, so a silent upstream change is caught instead of being
//     masked by stale fixtures — the exact failure mode that let the AllAnime
//     source rot undetected.

// mutationServer serves one full chain built from the (possibly mutated)
// fixtures it is given.
type mutationServer struct {
	episodes  string
	languages string // %s twice: base URL
	embed     string // %s once: master playlist URL
	searchDoc string
	master    string
	srv       *httptest.Server
}

func newMutationServer(t *testing.T, m mutationServer) *mutationServer {
	t.Helper()
	s := &m
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/browse":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(s.searchDoc))
		case strings.HasSuffix(r.URL.Path, "/episodes"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(s.episodes))
		case strings.HasSuffix(r.URL.Path, "/languages"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, s.languages, s.srv.URL, s.srv.URL)
		case strings.HasPrefix(r.URL.Path, "/embed/"):
			lang := strings.TrimPrefix(r.URL.Path, "/embed/")
			fmt.Fprintf(w, s.embed, fmt.Sprintf("%s/stream/%s/master.m3u8", s.srv.URL, lang))
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
		episodes:  episodesFixture,
		languages: languagesFixture,
		embed:     embedFixture,
		searchDoc: searchPage,
		master:    masterPlaylist,
	}
}

// exercise runs the whole chain and returns the first error, or nil.
func exercise(t *testing.T, s *mutationServer) error {
	t.Helper()
	c := NewClientForTest(s.srv.URL)

	results, err := c.SearchAnime(context.Background(), "jojo")
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("search: no results")
	}
	eps, err := c.GetAnimeEpisodes(context.Background(), results[0].URL)
	if err != nil {
		return fmt.Errorf("episodes: %w", err)
	}
	if len(eps) == 0 {
		return fmt.Errorf("episodes: empty list returned without an error")
	}
	if _, _, err := c.GetEpisodeStreamURL(context.Background(), eps[0].URL, "720p"); err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	return nil
}

// TestMetaControl_PristineFixturesSucceed is the control: without it, a
// mutation test could "pass" because the chain is broken for some unrelated
// reason.
func TestMetaControl_PristineFixturesSucceed(t *testing.T) {
	t.Parallel()
	s := newMutationServer(t, pristine())
	require.NoError(t, exercise(t, s),
		"the unmutated fixtures must drive the chain end to end, or every mutation below is meaningless")
}

// TestMetaMutations_AreDetected breaks one field at a time. Each case documents
// a real way anidb.app could change under us.
func TestMetaMutations_AreDetected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*mutationServer)
		because string
	}{
		{
			name:    "episodes array renamed",
			mutate:  func(m *mutationServer) { m.episodes = strings.ReplaceAll(m.episodes, `"episodes"`, `"items"`) },
			because: "the episode list key is load-bearing; a rename must not decode to an empty list silently",
		},
		{
			name:    "episode id field renamed",
			mutate:  func(m *mutationServer) { m.episodes = strings.ReplaceAll(m.episodes, `"id"`, `"identifier"`) },
			because: "without ids no episode can be streamed",
		},
		{
			name: "embed_url renamed with a different delimiter",
			mutate: func(m *mutationServer) {
				m.languages = strings.ReplaceAll(m.languages, `"embed_url"`, `"embedUrl"`)
			},
			because: "jsonx matches names case-insensitively but respects delimiters, so embedUrl must NOT bind to embed_url",
		},
		{
			name: "languages array renamed",
			mutate: func(m *mutationServer) {
				m.languages = strings.ReplaceAll(m.languages, `"languages"`, `"tracks"`)
			},
			because: "with no languages decoded there is no embed URL and no stream",
		},
		{
			name: "embed player key renamed",
			mutate: func(m *mutationServer) {
				m.embed = strings.ReplaceAll(m.embed, "file:", "source:")
			},
			because: "the m3u8 is scraped out of the player config by that key",
		},
		{
			name: "anime permalink path changed",
			mutate: func(m *mutationServer) {
				m.searchDoc = strings.ReplaceAll(m.searchDoc, "/anime/", "/series/")
			},
			because: "search results are recognised by their /anime/<slug>-<id> shape",
		},
		{
			name: "permalinks lose their numeric id",
			mutate: func(m *mutationServer) {
				for _, id := range []string{"-4979", "-2534", "-77"} {
					m.searchDoc = strings.ReplaceAll(m.searchDoc, id, "")
				}
			},
			because: "the numeric id is what addresses the episodes API; a slug alone is unusable",
		},
		{
			name: "embed URLs emptied",
			mutate: func(m *mutationServer) {
				m.languages = strings.ReplaceAll(m.languages, `%s/embed/`, ``)
			},
			because: "a language entry without an embed URL is not playable and must not be selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := pristine()
			tt.mutate(&m)
			s := newMutationServer(t, m)
			err := exercise(t, s)
			require.Error(t, err, "mutation went undetected — %s", tt.because)
			t.Logf("detected: %v", err)
		})
	}
}

// TestMetaMutation_MasterPlaylistWithoutVariants proves the quality path is
// really exercised: strip the variant lines and 720p can no longer be selected,
// so the client must fall back to the master rather than fabricate a URL.
func TestMetaMutation_MasterPlaylistWithoutVariants(t *testing.T) {
	t.Parallel()
	m := pristine()
	m.master = "#EXTM3U\n"
	s := newMutationServer(t, m)

	c := NewClientForTest(s.srv.URL)
	got, _, err := c.GetEpisodeStreamURL(context.Background(), s.srv.URL+"/episode/20049", "720p")
	require.NoError(t, err, "a variant-less playlist is still playable via the master")
	assert.True(t, strings.HasSuffix(got, "master.m3u8"),
		"with no variants the master must be returned, got %s", got)
}

// TestFixturesMatchLiveShape compares the shape the fixtures encode against the
// live service. Live-gated: it is a canary for upstream drift, not a CI gate.
func TestFixturesMatchLiveShape(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to compare fixtures against the live service")
	}
	c := NewAniDBClient()

	// Search page: the permalink shape the fixture encodes must still parse.
	results, err := c.SearchAnime(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	require.NotEmpty(t, results, "live browse page no longer yields cards — searchPage fixture is stale")
	require.Regexp(t, `/anime/[a-z0-9-]+-\d+$`, results[0].URL,
		"live permalink shape drifted from the fixture")

	// Episodes: the fixture assumes id/number/filler.
	eps, err := c.GetAnimeEpisodes(context.Background(), results[0].URL)
	require.NoError(t, err)
	require.NotEmpty(t, eps, "live episode list is empty — episodesFixture may be stale")
	require.NotEmpty(t, eps[0].DataID, "live payload no longer carries an episode id")
	require.NotEmpty(t, eps[0].Number, "live payload no longer carries an episode number")

	// Languages + embed: the fixture assumes embed_url and a file: key.
	streamURL, meta, err := c.GetEpisodeStreamURL(context.Background(), eps[0].URL, "best")
	require.NoError(t, err, "live stream chain broke — languagesFixture/embedFixture may be stale")
	require.Contains(t, streamURL, ".m3u8")
	require.NotEmpty(t, meta["audio_lang"], "live payload no longer carries a language code")
}
