package startflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestNewClientBaseURL pins the catalog host: the default, and the
// GOANIME_STARTFLIX_URL override (trimmed, without a trailing slash) for when
// the site moves. Not parallel: it sets the environment.
func TestNewClientBaseURL(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_URL", "")
	if got := NewClient().BaseURL(); got != DefaultBase {
		t.Errorf("default base = %q, want %q", got, DefaultBase)
	}
	t.Setenv("GOANIME_STARTFLIX_URL", "  https://startflix.example/ ")
	if got := NewClient().BaseURL(); got != "https://startflix.example" {
		t.Errorf("override base = %q", got)
	}
}

// TestNewClientWithHTTPKeepsTheGuardedStreamClient: a caller-supplied client
// serves the catalog pages, but the proxy's upstream stays the SSRF-guarded
// stream client, so mocking the catalog never unguards playback.
func TestNewClientWithHTTPKeepsTheGuardedStreamClient(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_URL", "")
	hc := &http.Client{}
	c := NewClientWithHTTP(hc)
	if c.http != hc {
		t.Error("the caller's client does not serve the catalog")
	}
	if c.proxyClient() == hc || c.proxyClient() == nil {
		t.Error("the proxy upstream must be the guarded stream client, not the caller's")
	}
}

// TestStreamClientRefusesInternalAddresses is the guard the Abyss proxy relies
// on: media URLs come from a remote page, so the upstream client must never
// reach loopback or private addresses, whatever the URL says.
func TestStreamClientRefusesInternalAddresses(t *testing.T) {
	t.Parallel()
	var reached atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached.Store(true) }))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, http.NoBody)
	resp, err := newStreamClient().Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("the stream client reached a loopback server")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("err = %v, want the SSRF guard's refusal", err)
	}
	if reached.Load() {
		t.Error("the loopback server saw a request")
	}
}

func TestSharedIsOneClient(t *testing.T) {
	t.Parallel()
	first, second := Shared(), Shared()
	if first != second {
		t.Error("Shared returned two clients; the panel and season caches would split")
	}
}

// TestCopyAbyssChunkRangeRejectsBadRanges pins the range check every chunked
// read goes through, before any part is fetched.
func TestCopyAbyssChunkRangeRejectsBadRanges(t *testing.T) {
	t.Parallel()
	chunked := &abyssFile{size: 1000, chunk: &abyssChunkSource{}}
	for _, r := range [][2]int64{{-1, 10}, {10, 5}, {0, 1000}, {999, 2000}} {
		if err := copyAbyssChunkRange(context.Background(), nil, chunked, r[0], r[1], nil); err == nil {
			t.Errorf("range %v accepted", r)
		}
	}
	single := &abyssFile{size: 1000}
	if err := copyAbyssChunkRange(context.Background(), nil, single, 0, 10, nil); err == nil || !strings.Contains(err.Error(), "not a chunked") {
		t.Errorf("single-file file: err = %v", err)
	}
}

// TestPanelsByID: a panel built from a bare id has the URL a StartFlix title
// page would iframe, so the season/player caches and the referer logic treat it
// the same. Not parallel: it sets the environment.
func TestPanelsByID(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "")
	if DefaultPanelBase != "https://www.painel-aso.sbs" {
		t.Errorf("DefaultPanelBase = %q; update this pin together with the host (2026-10-09)", DefaultPanelBase)
	}
	series := SeriesPanel(1396)
	if want, ok := panelFromURL(DefaultPanelBase + "/embed/1396"); !ok || series != want {
		t.Errorf("SeriesPanel = %+v, want %+v", series, want)
	}
	movie := MoviePanel("tt34385135")
	if want, ok := panelFromURL(DefaultPanelBase + "/filme/tt34385135"); !ok || movie != want {
		t.Errorf("MoviePanel = %+v, want %+v", movie, want)
	}
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", " https://painel.example/ ")
	if got := SeriesPanel(1).URL; got != "https://painel.example/embed/1" {
		t.Errorf("override panel URL = %q", got)
	}
}

// TestPanelBase: the panel host defaults to the pinned one and follows a
// trimmed override. Not parallel: it sets the environment.
func TestPanelBase(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "")
	if got := panelBase(); got != DefaultPanelBase {
		t.Errorf("panelBase = %q, want %q", got, DefaultPanelBase)
	}
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "   ")
	if got := panelBase(); got != DefaultPanelBase {
		t.Errorf("blank override: panelBase = %q, want the default", got)
	}
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", " https://painel.example// ")
	if got := panelBase(); got != "https://painel.example" {
		t.Errorf("override: panelBase = %q", got)
	}
}

// TestSeriesPanel: keyed by TMDB id, with the episode endpoint on its origin.
// Not parallel: it sets the environment.
func TestSeriesPanel(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://painel.test")
	p := SeriesPanel(1396)
	if p.URL != "https://painel.test/embed/1396" || p.Kind != KindSeries || p.TMDBID != 1396 || p.IMDBID != "" {
		t.Errorf("SeriesPanel = %+v", p)
	}
	if got := p.EpisodeURL("18145"); got != "https://painel.test/episodio/18145" {
		t.Errorf("EpisodeURL = %q", got)
	}
}

// TestMoviePanel: keyed by IMDb id; the URL is the players list itself.
// Not parallel: it sets the environment.
func TestMoviePanel(t *testing.T) {
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://painel.test")
	p := MoviePanel("tt34385135")
	if p.URL != "https://painel.test/filme/tt34385135" || p.Kind != KindMovie || p.IMDBID != "tt34385135" || p.TMDBID != 0 {
		t.Errorf("MoviePanel = %+v", p)
	}
}
