package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file is the offline cascade: the real animeDriveProvider, driving the
// real AnimeDriveAdapter, driving the real animedrive client, against a
// scripted server. Only the network is fake — every layer the app actually
// executes is exercised, so a break anywhere in provider → adapter → client →
// parsing fails here.

// animeDriveStub serves the whole chain and records which paths were hit, so
// the tests can assert the request sequence and not just the final value.
type animeDriveStub struct {
	srv *httptest.Server

	mu   sync.Mutex
	hits []string
	refs []string

	// Fault injection: when non-zero, the matching stage answers with this code.
	failSearch, failDetails, failEpisode int
	// deadServers lists the media paths whose probes must fail.
	deadServers map[string]bool
}

func newAnimeDriveStub(t *testing.T) *animeDriveStub {
	t.Helper()
	s := &animeDriveStub{deadServers: map[string]bool{}}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits = append(s.hits, r.URL.Path)
		s.mu.Unlock()

		switch {
		case strings.HasSuffix(r.URL.Path, ".mp4"):
			s.mu.Lock()
			s.refs = append(s.refs, r.Header.Get("Referer"))
			s.mu.Unlock()
			if s.deadServers[r.URL.Path] {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusPartialContent)

		case r.URL.Path == "/" && r.URL.Query().Has("s"):
			if s.failSearch != 0 {
				w.WriteHeader(s.failSearch)
				return
			}
			fmt.Fprintf(w, `<html><body><article class="item">
				<a href="/anime/test-anime/"><img src="/t.jpg"/></a><h3>Test Anime</h3>
			</article></body></html>`)

		case r.URL.Path == "/anime/test-anime/":
			if s.failDetails != 0 {
				w.WriteHeader(s.failDetails)
				return
			}
			fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Test Anime</h1>
				<ul class="episodios">
					<li><a href="/episodio/test-anime-episodio-01/">Episode 1</a></li>
					<li><a href="/episodio/test-anime-episodio-02/">Episode 2</a></li>
				</ul>
			</body></html>`)

		case strings.HasPrefix(r.URL.Path, "/episodio/"):
			if s.failEpisode != 0 {
				w.WriteHeader(s.failEpisode)
				return
			}
			fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/sd.mp4"></video></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/fhd.mp4"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
			`, s.srv.URL, s.srv.URL)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *animeDriveStub) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

// provider wires the real animeDriveProvider to an adapter pointed at the stub.
// The sync.Once is consumed up front so lazyGetAdapter keeps the injected
// adapter instead of building a live one.
func (s *animeDriveStub) provider(t *testing.T) *animeDriveProvider {
	t.Helper()
	p := &animeDriveProvider{}
	p.adapter.UnifiedScraper = scraper.NewAnimeDriveAdapterForTest(s.srv.URL)
	p.once.Do(func() {})
	return p
}

// TestAnimeDriveCascade_SearchToStream is the happy path across all four stages.
func TestAnimeDriveCascade_SearchToStream(t *testing.T) {
	t.Parallel()
	stub := newAnimeDriveStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	results, err := p.Search(ctx, "test anime")
	require.NoError(t, err)
	require.Len(t, results, 1)
	anime := results[0]

	assert.Equal(t, "AnimeDrive", anime.Source, "the provider must stamp the canonical source")
	assert.Contains(t, anime.Name, "[PT-BR]", "results must carry the language tag")

	// The tagged result must route back to this same source.
	_, resolved := source.Resolve(anime)
	assert.Equal(t, source.AnimeDrive, resolved.Kind, "reason: %s", resolved.Reason)

	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "01", eps[0].Number)
	assert.Equal(t, "02", eps[1].Number)

	streamURL, err := p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/fhd.mp4", streamURL,
		"best must select FHD over SD")

	assert.Contains(t, stub.paths(), "/episodio/test-anime-episodio-01/",
		"the cascade must reach the episode page")
}

// TestAnimeDriveCascade_QualityReachesTheClient proves the quality argument
// survives provider → adapter → client instead of being dropped in the glue.
func TestAnimeDriveCascade_QualityReachesTheClient(t *testing.T) {
	t.Parallel()
	stub := newAnimeDriveStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	anime := &models.Anime{Source: "AnimeDrive", URL: stub.srv.URL + "/anime/test-anime/"}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)

	got, err := p.FetchStreamURL(ctx, &eps[0], anime, "worst")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/sd.mp4", got,
		"worst must select SD; the FHD URL here means quality was dropped")
}

// TestAnimeDriveCascade_FailurePropagation checks each stage's failure surfaces
// as an error rather than an empty success.
func TestAnimeDriveCascade_FailurePropagation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		break_ func(*animeDriveStub)
		run    func(*animeDriveProvider, *animeDriveStub) error
	}{
		{
			name:   "search 503",
			break_: func(s *animeDriveStub) { s.failSearch = http.StatusServiceUnavailable },
			run: func(p *animeDriveProvider, _ *animeDriveStub) error {
				_, err := p.Search(context.Background(), "x")
				return err
			},
		},
		{
			name:   "details 500",
			break_: func(s *animeDriveStub) { s.failDetails = http.StatusInternalServerError },
			run: func(p *animeDriveProvider, s *animeDriveStub) error {
				_, err := p.FetchEpisodes(context.Background(),
					&models.Anime{URL: s.srv.URL + "/anime/test-anime/"})
				return err
			},
		},
		{
			name:   "episode 500",
			break_: func(s *animeDriveStub) { s.failEpisode = http.StatusInternalServerError },
			run: func(p *animeDriveProvider, s *animeDriveStub) error {
				anime := &models.Anime{Source: "AnimeDrive", URL: s.srv.URL + "/anime/test-anime/"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.srv.URL + "/episodio/test-anime-episodio-01/"}, anime, "best")
				return err
			},
		},
		{
			name: "all servers dead",
			break_: func(s *animeDriveStub) {
				s.deadServers["/sd.mp4"] = true
				s.deadServers["/fhd.mp4"] = true
			},
			run: func(p *animeDriveProvider, s *animeDriveStub) error {
				anime := &models.Anime{Source: "AnimeDrive", URL: s.srv.URL + "/anime/test-anime/"}
				_, err := p.FetchStreamURL(context.Background(),
					&models.Episode{URL: s.srv.URL + "/episodio/test-anime-episodio-01/"}, anime, "best")
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stub := newAnimeDriveStub(t)
			tt.break_(stub)
			err := tt.run(stub.provider(t), stub)
			require.Error(t, err, "a broken stage must not report success")
		})
	}
}

// TestAnimeDriveCascade_ErrorHasSinglePrefix pins the duplicated-prefix fix:
// leaf diagnostics already read "AnimeDrive stream: ...", so the provider must
// not wrap them a second time.
func TestAnimeDriveCascade_ErrorHasSinglePrefix(t *testing.T) {
	t.Parallel()
	stub := newAnimeDriveStub(t)
	stub.deadServers["/sd.mp4"] = true
	stub.deadServers["/fhd.mp4"] = true
	p := stub.provider(t)

	anime := &models.Anime{Source: "AnimeDrive", URL: stub.srv.URL + "/anime/test-anime/"}
	_, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: stub.srv.URL + "/episodio/test-anime-episodio-01/"}, anime, "best")
	require.Error(t, err)
	assert.Equal(t, 1, strings.Count(err.Error(), "AnimeDrive"),
		"duplicated source prefix in %q", err.Error())
	assert.Contains(t, err.Error(), "none of the 2 servers")
}

// TestAnimeDriveCascade_StreamCarriesPlaybackState pins the side effects the
// guide calls mandatory, plus the one this source cannot play without.
//
// The referer is the load-bearing one: three of the four mp4 hosts 403 every
// request without "https://animesdrive.cloud/" (trailing slash included), so
// an episode that resolves and sets no referer plays nowhere.
func TestAnimeDriveCascade_StreamCarriesPlaybackState(t *testing.T) {
	// Not parallel: mutates process-wide playback globals.
	stub := newAnimeDriveStub(t)
	p := stub.provider(t)

	util.SetGlobalReferer("https://stale.example/")
	t.Cleanup(func() {
		util.ClearGlobalSubtitles()
		util.SetGlobalReferer("")
	})

	anime := &models.Anime{Source: "AnimeDrive", URL: stub.srv.URL + "/anime/test-anime/"}
	stream, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: stub.srv.URL + "/episodio/test-anime-episodio-01/"}, anime, "")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/fhd.mp4", stream,
		"empty quality must default to best")

	assert.Equal(t, "AnimeDrive", util.GetGlobalAnimeSource(), "FetchStreamURL must stamp the source")
	assert.Equal(t, stub.srv.URL+"/", util.GetGlobalReferer(),
		"without the site referer the CDNs 403 and mpv stalls")

	stub.mu.Lock()
	defer stub.mu.Unlock()
	require.NotEmpty(t, stub.refs, "the availability probes must reach the server")
	for _, ref := range stub.refs {
		assert.Equal(t, stub.srv.URL+"/", ref, "every probe must carry the slashed referer")
	}
}

// animeDriveFakeAdapter records what the provider passes down, for branches no
// stub page can reach (empty stream URL, default quality).
type animeDriveFakeAdapter struct {
	streamURL  string
	metadata   map[string]string
	err        error
	gotOptions []any
}

func (f *animeDriveFakeAdapter) SearchAnime(string, ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (f *animeDriveFakeAdapter) GetAnimeEpisodes(string) ([]models.Episode, error) {
	return nil, nil
}

func (f *animeDriveFakeAdapter) GetStreamURL(_ string, options ...any) (streamURL string, metadata map[string]string, err error) {
	f.gotOptions = options
	return f.streamURL, f.metadata, f.err
}

func (f *animeDriveFakeAdapter) GetType() scraper.ScraperType {
	return scraper.AnimeDriveType
}

func animeDriveProviderWithFake(t *testing.T, fake *animeDriveFakeAdapter) *animeDriveProvider {
	t.Helper()
	p := &animeDriveProvider{}
	p.adapter.UnifiedScraper = fake
	p.once.Do(func() {})
	return p
}

func TestAnimeDriveProvider_SearchSupportsLegacyAdapter(t *testing.T) {
	p := animeDriveProviderWithFake(t, &animeDriveFakeAdapter{})
	results, err := p.Search(context.Background(), "naruto")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAnimeDriveProvider_EmptyStreamURLIsAnError(t *testing.T) {
	// Not parallel: FetchStreamURL stamps process-wide playback globals.
	p := animeDriveProviderWithFake(t, &animeDriveFakeAdapter{metadata: map[string]string{}})

	_, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: "https://animesdrive.cloud/episodio/x"},
		&models.Anime{Source: "AnimeDrive"}, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty stream URL")
}

func TestAnimeDriveProvider_EmptyQualityDefaultsToBest(t *testing.T) {
	// Not parallel: asserts process-wide playback globals.
	fake := &animeDriveFakeAdapter{
		streamURL: "https://cdn.example/v.mp4",
		metadata:  map[string]string{"referer": "https://animesdrive.cloud/"},
	}
	p := animeDriveProviderWithFake(t, fake)

	// Not parallel with the playback-state test: asserts a global.
	util.SetGlobalReferer("")
	t.Cleanup(func() { util.SetGlobalReferer("") })

	got, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: "https://animesdrive.cloud/episodio/x"},
		&models.Anime{Source: "AnimeDrive"}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/v.mp4", got)
	require.Len(t, fake.gotOptions, 1)
	assert.Equal(t, "best", fake.gotOptions[0])
	assert.Equal(t, "https://animesdrive.cloud/", util.GetGlobalReferer())
}
