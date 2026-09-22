package providers

import (
	"context"
	"encoding/base64"
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

// This file is the offline cascade: the real hianimeProvider, driving the real
// HiAnimeAdapter, driving the real hianime client, against a scripted server. Only
// the network is fake — every layer the app actually executes is exercised, so
// a break anywhere in provider → adapter → client → parsing fails here.

// obfuscate reproduces the embed player's encoding: JSON, XOR'd with the
// player's build tag, base64'd. Written here from the format rather than
// imported, so a change to the real decoder has to be matched deliberately
// instead of cancelling itself out.
func obfuscate(payload string) string {
	key := []byte("otaku-embed-v1")
	raw := []byte(payload)
	for i := range raw {
		raw[i] ^= key[i%len(key)]
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// hianimeStub serves the whole chain and records which paths were hit, so the
// tests can assert the request sequence and not just the final value.
type hianimeStub struct {
	srv *httptest.Server

	mu   sync.Mutex
	hits []string

	// Fault injection: when non-zero, the matching stage answers with this code.
	failSearch, failEpisodes, failServers, failEmbed int
}

func newHiAnimeStub(t *testing.T) *hianimeStub {
	t.Helper()
	s := &hianimeStub{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits = append(s.hits, r.URL.Path)
		s.mu.Unlock()

		switch {
		case r.URL.Path == "/search":
			if s.failSearch != 0 {
				w.WriteHeader(s.failSearch)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><div class="flw-item">
				<div class="film-poster">
					<img src="https://cdn.example/42.jpg" class="film-poster-img" alt="Cowboy Bebop">
				</div>
				<div class="film-detail"><h3 class="film-name">
					<a href="%s/cowboy-bebop-42" title="Cowboy Bebop" class="dynamic-name">Cowboy Bebop</a>
				</h3></div>
			</div></body></html>`, s.srv.URL)

		case strings.HasPrefix(r.URL.Path, "/api/theme/episode/list/"):
			if s.failEpisodes != 0 {
				w.WriteHeader(s.failEpisodes)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			// The site answers with rendered HTML inside a JSON envelope.
			fmt.Fprint(w, `{"status":true,"totalItems":2,"html":"`+
				`<a title=\"Asteroid Blues\" class=\"ssl-item ep-item\" data-number=\"1\" data-id=\"9001\"></a>`+
				`<a title=\"Stray Dog Strut\" class=\"ssl-item ep-item ssl-item-filler\" data-number=\"2\" data-id=\"9002\"></a>`+
				`"}`)

		case r.URL.Path == "/api/theme/episode/servers":
			if s.failServers != 0 {
				w.WriteHeader(s.failServers)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			sub := base64.StdEncoding.EncodeToString([]byte(s.srv.URL + "/embed/sub"))
			dub := base64.StdEncoding.EncodeToString([]byte(s.srv.URL + "/embed/dub"))
			// MegaPlay is listed first on purpose: the client must skip the
			// servers it cannot read rather than taking whatever comes first.
			fmt.Fprint(w, `{"status":true,"html":"`+
				`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"HD-1\" data-hash=\"`+
				base64.StdEncoding.EncodeToString([]byte("https://megaplay.example/stream/s-2/1/sub"))+`\"></div>`+
				`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"ZokoAnime\" data-hash=\"`+sub+`\"></div>`+
				`<div class=\"item server-item\" data-type=\"dub\" data-server-name=\"ZokoAnime\" data-hash=\"`+dub+`\"></div>`+
				`"}`)

		case strings.HasPrefix(r.URL.Path, "/embed/"):
			if s.failEmbed != 0 {
				w.WriteHeader(s.failEmbed)
				return
			}
			audio := strings.TrimPrefix(r.URL.Path, "/embed/")
			config := fmt.Sprintf(
				`{"src":"%s/stream/%s/master.m3u8","subtitles":[`+
					`{"lang":"en","label":"English","src":"%s/subs/en.vtt","default":true},`+
					`{"lang":"en","label":"Portuguese (- Brazilian)","src":"%s/subs/pt.vtt"}]}`,
				s.srv.URL, audio, s.srv.URL, s.srv.URL)
			fmt.Fprintf(w, `<html><script>window.__P="%s"</script></html>`, obfuscate(config))

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

func (s *hianimeStub) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

// episodeURL is the watch URL the client hands back and later reads its own id
// out of.
func (s *hianimeStub) episodeURL(id string) string {
	return s.srv.URL + "/watch/cowboy-bebop-42?ep=" + id
}

// provider wires the real hianimeProvider to an adapter pointed at the stub. The
// sync.Once is consumed up front so lazyGetAdapter keeps the injected adapter
// instead of building a live one.
func (s *hianimeStub) provider(t *testing.T) *hianimeProvider {
	t.Helper()
	p := &hianimeProvider{}
	p.adapter.UnifiedScraper = scraper.NewHiAnimeAdapterForTest(s.srv.URL)
	p.once.Do(func() {})
	return p
}

// TestHiAnimeCascade_SearchToStream is the happy path across all four stages.
func TestHiAnimeCascade_SearchToStream(t *testing.T) {
	t.Parallel()
	stub := newHiAnimeStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	results, err := p.Search(ctx, "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, results, 1)
	anime := results[0]

	assert.Equal(t, "HiAnime", anime.Source, "the provider must stamp the canonical source")
	assert.Contains(t, anime.Name, "[English]", "results must carry the language tag")

	// The tagged result must route back to this same source.
	_, resolved := source.Resolve(anime)
	assert.Equal(t, source.HiAnime, resolved.Kind, "reason: %s", resolved.Reason)

	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "1", eps[0].Number)
	assert.Equal(t, "Asteroid Blues", eps[0].Title.English)
	assert.True(t, eps[1].IsFiller)

	streamURL, err := p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/stream/sub/master.m3u8", streamURL,
		"the subtitled ZokoAnime entry is the one this source can read")

	assert.Equal(t, []string{
		"/search",
		"/api/theme/episode/list/42",
		"/api/theme/episode/servers",
		"/embed/sub",
	}, stub.paths(), "the cascade must walk exactly these stages, in order")
}

// TestHiAnimeCascade_QualityReachesTheClient proves the quality argument survives
// provider → adapter → client instead of being dropped in the glue.
func TestHiAnimeCascade_QualityReachesTheClient(t *testing.T) {
	t.Parallel()
	stub := newHiAnimeStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	anime := &models.Anime{Source: "HiAnime", URL: stub.srv.URL + "/cowboy-bebop-42"}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)

	got, err := p.FetchStreamURL(ctx, &eps[0], anime, "720p")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/stream/sub/index-f2.m3u8", got,
		"720p must select the 1280x720 variant; the master URL here means quality was dropped")

	assert.Contains(t, stub.paths(), "/stream/sub/master.m3u8",
		"selecting a variant requires reading the master playlist")
}

// TestHiAnimeCascade_FailurePropagation checks each stage's failure surfaces as an
// error rather than an empty success.
func TestHiAnimeCascade_FailurePropagation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		break_ func(*hianimeStub)
		run    func(*testing.T, *hianimeProvider, *hianimeStub) error
	}{
		{
			name:   "search 503",
			break_: func(s *hianimeStub) { s.failSearch = http.StatusServiceUnavailable },
			run: func(t *testing.T, p *hianimeProvider, _ *hianimeStub) error {
				_, err := p.Search(context.Background(), "x")
				return err
			},
		},
		{
			name:   "episodes 500",
			break_: func(s *hianimeStub) { s.failEpisodes = http.StatusInternalServerError },
			run: func(t *testing.T, p *hianimeProvider, s *hianimeStub) error {
				_, err := p.FetchEpisodes(context.Background(),
					&models.Anime{URL: s.srv.URL + "/cowboy-bebop-42"})
				return err
			},
		},
		{
			name:   "servers 404",
			break_: func(s *hianimeStub) { s.failServers = http.StatusNotFound },
			run: func(t *testing.T, p *hianimeProvider, s *hianimeStub) error {
				anime := &models.Anime{Source: "HiAnime", URL: s.srv.URL + "/cowboy-bebop-42"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.episodeURL("9001")}, anime, "best")
				return err
			},
		},
		{
			name:   "embed 500",
			break_: func(s *hianimeStub) { s.failEmbed = http.StatusInternalServerError },
			run: func(t *testing.T, p *hianimeProvider, s *hianimeStub) error {
				anime := &models.Anime{Source: "HiAnime", URL: s.srv.URL + "/cowboy-bebop-42"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.episodeURL("9001")}, anime, "best")
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stub := newHiAnimeStub(t)
			tt.break_(stub)
			err := tt.run(t, stub.provider(t), stub)
			require.Error(t, err, "a broken stage must not report success")
		})
	}
}

// TestHiAnimeCascade_StreamCarriesPlaybackState pins the side effects the guide
// calls mandatory, plus the two this source cannot play without.
//
// The referer is the load-bearing one: the stream CDN 403s part of its URLs
// without it, so an episode that resolves and sets no referer plays on some
// titles and stalls on others. Subtitles come from the same payload and must
// REPLACE the previous episode's rather than accumulating or leaking.
func TestHiAnimeCascade_StreamCarriesPlaybackState(t *testing.T) {
	// Not parallel: mutates process-wide playback globals.
	stub := newHiAnimeStub(t)
	p := stub.provider(t)

	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://stale.example/old.vtt", Label: "Stale"}})
	util.SetGlobalReferer("https://stale.example/")
	t.Cleanup(func() {
		util.ClearGlobalSubtitles()
		util.SetGlobalReferer("")
	})

	anime := &models.Anime{Source: "HiAnime", URL: stub.srv.URL + "/cowboy-bebop-42"}
	_, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: stub.episodeURL("9001")}, anime, "best")
	require.NoError(t, err)

	assert.Equal(t, "HiAnime", util.GetGlobalAnimeSource(), "FetchStreamURL must stamp the source")
	assert.Equal(t, stub.srv.URL+"/", util.GetGlobalReferer(),
		"without the embed host's referer the CDN 403s part of its playlists")

	subs := util.GetGlobalSubtitles()
	require.Len(t, subs, 2, "the episode's own subtitle tracks must reach playback")
	labels := []string{subs[0].Label, subs[1].Label}
	assert.Equal(t, []string{"English", "Portuguese (- Brazilian)"}, labels)
	assert.NotContains(t, labels, "Stale", "the previous episode's tracks must not survive")
}
