package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
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

// This file is the offline cascade for Miruro: the real miruroProvider,
// driving the real MiruroAdapter, driving the real miruro client, against a
// scripted miruro.cx and its stream CDNs. Only the network is fake.

const (
	miruroTestID   = "C0wb0yBebop_session-0123456789ab"
	miruroTestSlug = "cowboy-bebop"
)

// miruroPromise marks a ["Promise", id] for miruroDevalue.
type miruroPromise int

// miruroDevalue flattens a value the way SvelteKit's devalue does. Written
// here from the format rather than imported, so a change to the real decoder
// has to be matched on purpose.
func miruroDevalue(root any) []any {
	var flat []any
	var add func(v any) int
	add = func(v any) int {
		idx := len(flat)
		flat = append(flat, nil)
		switch x := v.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			m := map[string]any{}
			for _, k := range keys {
				m[k] = add(x[k])
			}
			flat[idx] = m
		case []any:
			a := []any{}
			for _, c := range x {
				a = append(a, add(c))
			}
			flat[idx] = a
		case miruroPromise:
			flat[idx] = []any{"Promise", add(int(x))}
		default:
			flat[idx] = x
		}
		return idx
	}
	add(root)
	return flat
}

func miruroLine(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b) + "\n"
}

// miruroStub serves the chain and records the paths hit.
type miruroStub struct {
	srv *httptest.Server

	mu   sync.Mutex
	hits []string

	// Fault injection: when non-zero, the matching stage answers with it.
	failSearch, failWatch, failCDN int
}

func newMiruroStub(t *testing.T) *miruroStub {
	t.Helper()
	s := &miruroStub{}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *miruroStub) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.hits = append(s.hits, r.URL.Path)
	s.mu.Unlock()
	base := s.srv.URL
	watch := "/watch/" + miruroTestID + "/"

	switch {
	case r.URL.Path == "/search/__data.json":
		if s.failSearch != 0 {
			w.WriteHeader(s.failSearch)
			return
		}
		page := map[string]any{"items": []any{map[string]any{
			"id":             miruroTestID,
			"title":          map[string]any{"english": "Cowboy Bebop", "romaji": "Cowboy Bebop"},
			"status":         "FINISHED",
			"season_year":    1998,
			"episode_counts": map[string]any{"sub": 26, "dub": 26},
			"external_ids":   map[string]any{"anilist": []any{"1"}, "mal": []any{"1"}},
			"cover_url":      "https://img.example/bebop.jpg",
		}}}
		_, _ = w.Write([]byte(miruroLine(map[string]any{"type": "data", "nodes": []any{
			map[string]any{"type": "skip"},
			map[string]any{"type": "data", "data": miruroDevalue(page)},
		}})))

	case r.URL.Path == watch+"x/__data.json":
		_, _ = w.Write([]byte(miruroLine(map[string]any{
			"type": "redirect", "location": watch + miruroTestSlug + "?ep=" + r.URL.Query().Get("ep"),
		})))

	case r.URL.Path == watch+miruroTestSlug+"/__data.json":
		if s.failWatch != 0 {
			w.WriteHeader(s.failWatch)
			return
		}
		page := map[string]any{"episodesData": miruroPromise(1), "sourcesData": miruroPromise(2)}
		episodes := map[string]any{"episodes": []any{
			map[string]any{"number": 1, "title": "Asteroid Blues", "filler": false, "airDate": "1998-04-03"},
			map[string]any{"number": 2, "title": "Stray Dog Strut", "filler": true},
		}}
		sources := map[string]any{
			"providerOrder": []any{"sun"},
			"tracks": []any{map[string]any{"track": "ssub", "providers": []any{map[string]any{
				"provider": "sun",
				"subtitles": []any{
					map[string]any{"language": "en", "label": "English", "file": base + "/subs/en.vtt", "default": false},
					map[string]any{"language": "pt-BR", "label": "Portuguese (Brazil)", "file": base + "/subs/pt.vtt", "default": false},
					map[string]any{"language": "ar", "label": "Arabic", "file": base + "/subs/ar.vtt", "default": true},
				},
				"servers": []any{map[string]any{
					"server":  "icarus",
					"headers": map[string]any{"Referer": base + "/ultra/", "Origin": base + "/ultra"},
					"streams": []any{map[string]any{"url": base + "/cdn/master.m3u8", "format": "hls"}},
				}},
			}}}},
		}
		body := miruroLine(map[string]any{"type": "data", "nodes": []any{
			map[string]any{"type": "skip"},
			map[string]any{"type": "data", "data": miruroDevalue(page)},
		}}) +
			miruroLine(map[string]any{"type": "chunk", "id": 2, "data": miruroDevalue(sources)}) +
			miruroLine(map[string]any{"type": "chunk", "id": 1, "data": miruroDevalue(episodes)})
		w.Header().Set("Content-Type", "text/sveltekit-data")
		_, _ = w.Write([]byte(body))

	case strings.HasPrefix(r.URL.Path, "/cdn/"):
		if s.failCDN != 0 || r.Header.Get("Referer") != base+"/ultra/" || r.Header.Get("Origin") != base+"/ultra" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1,RESOLUTION=1280x720\nindex-720.m3u8\n"))

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *miruroStub) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

// provider wires the real miruroProvider to an adapter pointed at the stub.
// The sync.Once is consumed up front so lazyGetAdapter keeps it.
func (s *miruroStub) provider(t *testing.T) *miruroProvider {
	t.Helper()
	p := &miruroProvider{}
	p.adapter.UnifiedScraper = scraper.NewMiruroAdapterForTest(s.srv.URL)
	p.once.Do(func() {})
	return p
}

func TestMiruroCascade_SearchToStream(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	stub := newMiruroStub(t)
	p := stub.provider(t)
	ctx := context.Background()

	results, err := p.Search(ctx, "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, results, 1)
	anime := results[0]
	assert.Equal(t, "Miruro", anime.Source)
	assert.Contains(t, anime.Name, "[English]")
	assert.Equal(t, 1, anime.AnilistID, "the site hands out the AniList id, so enrichment looks it up exactly")

	_, resolved := source.Resolve(anime)
	assert.Equal(t, source.Miruro, resolved.Kind, "reason: %s", resolved.Reason)

	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "Asteroid Blues", eps[0].Title.English)
	assert.True(t, eps[1].IsFiller)

	got, err := p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/cdn/master.m3u8", got)

	watch := "/watch/" + miruroTestID + "/"
	assert.Equal(t, []string{
		"/search/__data.json",
		watch + "x/__data.json",
		watch + miruroTestSlug + "/__data.json",
		watch + miruroTestSlug + "/__data.json",
		"/cdn/master.m3u8",
	}, stub.paths(), "the episode URLs carry the canonical slug, so the stream costs no redirect")
}

func TestMiruroCascade_QualityReachesTheClient(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	stub := newMiruroStub(t)
	p := stub.provider(t)
	anime := &models.Anime{Source: "Miruro", URL: stub.srv.URL + "/info/" + miruroTestID}
	eps, err := p.FetchEpisodes(context.Background(), anime)
	require.NoError(t, err)

	got, err := p.FetchStreamURL(context.Background(), &eps[0], anime, "720p")
	require.NoError(t, err)
	assert.Equal(t, stub.srv.URL+"/cdn/index-720.m3u8", got, "the master URL here means the quality was dropped")
}

func TestMiruroCascade_FailurePropagation(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	anime := func(s *miruroStub) *models.Anime {
		return &models.Anime{Source: "Miruro", URL: s.srv.URL + "/info/" + miruroTestID}
	}
	episode := func(s *miruroStub) *models.Episode {
		return &models.Episode{URL: s.srv.URL + "/watch/" + miruroTestID + "/" + miruroTestSlug + "?ep=1"}
	}
	tests := []struct {
		name   string
		break_ func(*miruroStub)
		run    func(*miruroProvider, *miruroStub) error
	}{
		{"search 503", func(s *miruroStub) { s.failSearch = http.StatusServiceUnavailable }, func(p *miruroProvider, _ *miruroStub) error {
			_, err := p.Search(context.Background(), "x")
			return err
		}},
		{"watch page 500", func(s *miruroStub) { s.failWatch = http.StatusInternalServerError }, func(p *miruroProvider, s *miruroStub) error {
			_, err := p.FetchEpisodes(context.Background(), anime(s))
			return err
		}},
		{"every mirror refuses", func(s *miruroStub) { s.failCDN = http.StatusForbidden }, func(p *miruroProvider, s *miruroStub) error {
			_, err := p.FetchStreamURL(context.Background(), episode(s), anime(s), "best")
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := newMiruroStub(t)
			tt.break_(stub)
			require.Error(t, tt.run(stub.provider(t), stub), "a broken stage must not report success")
		})
	}
}

// TestMiruroCascade_StreamCarriesPlaybackState pins what mpv needs: the
// Origin as well as the Referer (one CDN family refuses without it), and the
// episode's own subtitles replacing the previous episode's, Arabic left out.
func TestMiruroCascade_StreamCarriesPlaybackState(t *testing.T) {
	// Not parallel: mutates process-wide playback globals.
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	stub := newMiruroStub(t)
	p := stub.provider(t)

	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://stale.example/old.vtt", Label: "Stale"}})
	util.SetGlobalReferer("https://stale.example/")
	t.Cleanup(func() {
		util.ClearGlobalSubtitles()
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
	})

	anime := &models.Anime{Source: "Miruro", URL: stub.srv.URL + "/info/" + miruroTestID}
	_, err := p.FetchStreamURL(context.Background(),
		&models.Episode{URL: stub.srv.URL + "/watch/" + miruroTestID + "/" + miruroTestSlug + "?ep=1"}, anime, "best")
	require.NoError(t, err)

	assert.Equal(t, "Miruro", util.GetGlobalAnimeSource())
	assert.Equal(t, stub.srv.URL+"/ultra/", util.GetGlobalReferer())
	assert.Equal(t, stub.srv.URL+"/ultra", util.GetGlobalOrigin(), "without the Origin this CDN answers 403")

	subs := util.GetGlobalSubtitles()
	require.Len(t, subs, 2)
	assert.Equal(t, []string{"English", "Portuguese (Brazil)"}, []string{subs[0].Label, subs[1].Label})
	assert.Equal(t, []string{"English", "Portuguese"}, []string{subs[0].Language, subs[1].Language})
}
