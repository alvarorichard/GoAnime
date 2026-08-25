package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file is the offline cascade: the real anidbProvider, driving the real
// AniDBAdapter, driving the real anidb client, against a scripted server. Only
// the network is fake — every layer the app actually executes is exercised, so
// a break anywhere in provider → adapter → client → parsing fails here.

// anidbStub serves the whole chain and records which paths were hit, so the
// tests can assert the request sequence and not just the final value.
type anidbStub struct {
	srv *httptest.Server

	mu   sync.Mutex
	hits []string

	// Fault injection: when non-zero, the matching stage answers with this code.
	failSearch, failEpisodes, failLanguages, failEmbed int
}

func newAniDBStub(t *testing.T) *anidbStub {
	t.Helper()
	s := &anidbStub{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits = append(s.hits, r.URL.Path)
		s.mu.Unlock()

		switch {
		case r.URL.Path == "/browse":
			if s.failSearch != 0 {
				w.WriteHeader(s.failSearch)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body>
				<a href="%s/anime/cowboy-bebop-42" class="anime-card" title="Cowboy Bebop">
					<img src="https://cdn.example/42.jpg" alt="Cowboy Bebop">
				</a></body></html>`, s.srv.URL)

		case strings.HasSuffix(r.URL.Path, "/episodes"):
			if s.failEpisodes != 0 {
				w.WriteHeader(s.failEpisodes)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"episodes":[
				{"id":9001,"number":1,"filler":false},
				{"id":9002,"number":2,"filler":true}
			]}`)

		case strings.HasSuffix(r.URL.Path, "/languages"):
			if s.failLanguages != 0 {
				w.WriteHeader(s.failLanguages)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"languages":[
				{"code":"eng","name":"English","embed_url":"%s/embed/eng"},
				{"code":"jpn","name":"Japanese","embed_url":"%s/embed/jpn"}
			]}`, s.srv.URL, s.srv.URL)

		case strings.HasPrefix(r.URL.Path, "/embed/"):
			if s.failEmbed != 0 {
				w.WriteHeader(s.failEmbed)
				return
			}
			lang := strings.TrimPrefix(r.URL.Path, "/embed/")
			fmt.Fprintf(w, `<html><script>setup({file: '%s/stream/%s/master.m3u8'})</script></html>`, s.srv.URL, lang)

		case strings.HasSuffix(r.URL.Path, "master.m3u8"):
			fmt.Fprint(w, "#EXTM3U\n"+
				"#EXT-X-STREAM-INF:BANDWIDTH=1218557,RESOLUTION=1920x1080\nindex-f1.m3u8\n"+
				"#EXT-X-STREAM-INF:BANDWIDTH=809659,RESOLUTION=1280x720\nindex-f2.m3u8\n")

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *anidbStub) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

// provider wires the real anidbProvider to an adapter pointed at the stub. The
// sync.Once is consumed up front so lazyGetAdapter keeps the injected adapter
// instead of building a live one.
func (s *anidbStub) provider(t *testing.T) *anidbProvider {
	t.Helper()
	p := &anidbProvider{}
	p.adapter.UnifiedScraper = scraper.NewAniDBAdapterForTest(s.srv.URL)
	p.once.Do(func() {})
	return p
}

// TestAniDBCascade_SearchToStream is the happy path across all four stages.
func TestAniDBCascade_SearchToStream(t *testing.T) {
	t.Parallel()
	stub := newAniDBStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	results, err := p.Search(ctx, "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, results, 1)
	anime := results[0]

	assert.Equal(t, "AniDB", anime.Source, "the provider must stamp the canonical source")
	assert.Contains(t, anime.Name, "[English]", "results must carry the language tag")

	// The tagged result must route back to this same source.
	_, resolved := source.Resolve(anime)
	assert.Equal(t, source.AniDB, resolved.Kind, "reason: %s", resolved.Reason)

	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "1", eps[0].Number)
	assert.True(t, eps[1].IsFiller)

	streamURL, err := p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/stream/jpn/master.m3u8", streamURL)

	assert.Equal(t, []string{
		"/browse",
		"/api/frontend/anime/42/episodes",
		"/api/frontend/episode/9001/languages",
		"/embed/jpn",
	}, stub.paths(), "the cascade must walk exactly these stages, in order")
}

// TestAniDBCascade_QualityReachesTheClient proves the quality argument survives
// provider → adapter → client instead of being dropped in the glue.
func TestAniDBCascade_QualityReachesTheClient(t *testing.T) {
	t.Parallel()
	stub := newAniDBStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	anime := &models.Anime{Source: "AniDB", URL: stub.srv.URL + "/anime/cowboy-bebop-42"}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)

	got, err := p.FetchStreamURL(ctx, &eps[0], anime, "720p")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/stream/jpn/index-f2.m3u8", got,
		"720p must select the 1280x720 variant; the master URL here means quality was dropped")

	assert.Contains(t, stub.paths(), "/stream/jpn/master.m3u8",
		"selecting a variant requires reading the master playlist")
}

// TestAniDBCascade_FailurePropagation checks each stage's failure surfaces as an
// error rather than an empty success.
func TestAniDBCascade_FailurePropagation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		break_ func(*anidbStub)
		run    func(*testing.T, *anidbProvider, *anidbStub) error
	}{
		{
			name:   "search 503",
			break_: func(s *anidbStub) { s.failSearch = http.StatusServiceUnavailable },
			run: func(t *testing.T, p *anidbProvider, _ *anidbStub) error {
				_, err := p.Search(context.Background(), "x")
				return err
			},
		},
		{
			name:   "episodes 500",
			break_: func(s *anidbStub) { s.failEpisodes = http.StatusInternalServerError },
			run: func(t *testing.T, p *anidbProvider, s *anidbStub) error {
				_, err := p.FetchEpisodes(context.Background(),
					&models.Anime{URL: s.srv.URL + "/anime/cowboy-bebop-42"})
				return err
			},
		},
		{
			name:   "languages 404",
			break_: func(s *anidbStub) { s.failLanguages = http.StatusNotFound },
			run: func(t *testing.T, p *anidbProvider, s *anidbStub) error {
				anime := &models.Anime{Source: "AniDB", URL: s.srv.URL + "/anime/cowboy-bebop-42"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.srv.URL + "/episode/9001"}, anime, "best")
				return err
			},
		},
		{
			name:   "embed 500",
			break_: func(s *anidbStub) { s.failEmbed = http.StatusInternalServerError },
			run: func(t *testing.T, p *anidbProvider, s *anidbStub) error {
				anime := &models.Anime{Source: "AniDB", URL: s.srv.URL + "/anime/cowboy-bebop-42"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.srv.URL + "/episode/9001"}, anime, "best")
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stub := newAniDBStub(t)
			tt.break_(stub)
			err := tt.run(t, stub.provider(t), stub)
			require.Error(t, err, "a broken stage must not report success")
		})
	}
}

// TestAniDBCascade_StreamClearsSubtitleState pins the side effects the guide
// calls mandatory: stale subtitles from the previous episode must not leak.
func TestAniDBCascade_StreamClearsSubtitleState(t *testing.T) {
	// Not parallel: mutates process-wide subtitle/source globals.
	stub := newAniDBStub(t)
	p := stub.provider(t)

	anime := &models.Anime{Source: "AniDB", URL: stub.srv.URL + "/anime/cowboy-bebop-42"}
	_, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: stub.srv.URL + "/episode/9001"}, anime, "best")
	require.NoError(t, err)

	assert.Empty(t, util.GetGlobalSubtitles(), "FetchStreamURL must clear subtitles first")
	assert.Equal(t, "AniDB", util.GetGlobalAnimeSource(), "FetchStreamURL must stamp the source")
}
