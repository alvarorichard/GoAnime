package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

type nineAnimeMockScraper struct {
	episodes    []models.Episode
	streamURL   string
	streamErr   error
	metadata    map[string]string
	episodesArg string
	streamArg   string
}

func (m *nineAnimeMockScraper) SearchAnime(_ string, _ ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (m *nineAnimeMockScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	m.episodesArg = animeURL
	return m.episodes, nil
}

func (m *nineAnimeMockScraper) GetStreamURL(episodeURL string, _ ...any) (string, map[string]string, error) {
	m.streamArg = episodeURL
	if m.streamErr != nil {
		return "", nil, m.streamErr
	}
	return m.streamURL, m.metadata, nil
}

func (m *nineAnimeMockScraper) GetType() scraper.ScraperType {
	return scraper.NineAnimeType
}

func resetNineAnimePlaybackState() {
	util.ResetPlaybackState()
	util.SetSubtitlesDisabled(false)
}

func TestNineAnimeProviderFetchEpisodes(t *testing.T) {
	t.Parallel()

	mock := &nineAnimeMockScraper{
		episodes: []models.Episode{{Number: "1", URL: "ep-1", DataID: "ep-1"}},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &nineAnimeProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "677"}
	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("FetchEpisodes returned error: %v", err)
	}

	if provider.Kind() != source.NineAnime {
		t.Fatalf("provider kind = %s, want %s", provider.Kind(), source.NineAnime)
	}
	if len(episodes) != 1 {
		t.Fatalf("FetchEpisodes returned %d episodes, want 1", len(episodes))
	}
	if mock.episodesArg != anime.URL {
		t.Fatalf("scraper received anime URL %q, want %q", mock.episodesArg, anime.URL)
	}
	if len(lookup.requested) != 1 || lookup.requested[0] != scraper.NineAnimeType {
		t.Fatalf("GetScraper called with %v, want [%v]", lookup.requested, scraper.NineAnimeType)
	}
}

func TestNineAnimeProviderFetchStreamURLStoresPlaybackMetadata(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)

	mock := &nineAnimeMockScraper{
		streamURL: "https://cdn.example.com/master.m3u8",
		metadata: map[string]string{
			"referer":         "https://rapid-cloud.co/",
			"subtitles":       "https://cdn.example.com/en.vtt,https://cdn.example.com/pt.vtt",
			"subtitle_labels": "English,Portuguese - Brazilian Portuguese",
		},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &nineAnimeProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1", DataID: "ep-1"}

	streamURL, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err != nil {
		t.Fatalf("FetchStreamURL returned error: %v", err)
	}

	if streamURL != mock.streamURL {
		t.Fatalf("FetchStreamURL = %q, want %q", streamURL, mock.streamURL)
	}
	if mock.streamArg != episode.URL {
		t.Fatalf("scraper received episode URL %q, want %q", mock.streamArg, episode.URL)
	}
	if len(lookup.requested) != 1 || lookup.requested[0] != scraper.NineAnimeType {
		t.Fatalf("GetScraper called with %v, want [%v]", lookup.requested, scraper.NineAnimeType)
	}

	if got := util.GetGlobalReferer(); got != "https://rapid-cloud.co/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://rapid-cloud.co/")
	}

	subtitles := util.GetGlobalSubtitles()
	if len(subtitles) != 2 {
		t.Fatalf("stored %d subtitle tracks, want 2", len(subtitles))
	}
	if subtitles[0].Language != "eng" {
		t.Fatalf("first subtitle language = %q, want %q", subtitles[0].Language, "eng")
	}
	if subtitles[1].Language != "por" {
		t.Fatalf("second subtitle language = %q, want %q", subtitles[1].Language, "por")
	}
}

// Bug scenario: if the upstream rapid-cloud endpoint returns an error (e.g. 403),
// FetchStreamURL must propagate it wrapped with "9anime stream:" prefix so callers
// can distinguish source-level failures from provider-level bugs.
func TestNineAnimeProviderFetchStreamURLErrorPropagates(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)

	upstreamErr := errors.New("403 forbidden from rapid-cloud")
	mock := &nineAnimeMockScraper{streamErr: upstreamErr}
	provider := &nineAnimeProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}
	_, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err == nil {
		t.Fatal("expected error from upstream failure, got nil")
	}
	if !errors.Is(err, upstreamErr) {
		t.Fatalf("expected upstream error in chain, got: %v", err)
	}
}

// Bug scenario: when the scraper returns empty metadata (no referer, no subtitles),
// applyNineAnimePlaybackMetadata must be a no-op — no panic, no state mutation.
func TestNineAnimeProviderFetchStreamURLEmptyMetadataNoSideEffects(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)

	mock := &nineAnimeMockScraper{
		streamURL: "https://cdn.example.com/master.m3u8",
		metadata:  map[string]string{},
	}
	provider := &nineAnimeProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}
	_, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := util.GetGlobalReferer(); got != "" {
		t.Fatalf("GlobalReferer should be empty after empty metadata, got %q", got)
	}
	if subs := util.GetGlobalSubtitles(); len(subs) != 0 {
		t.Fatalf("GlobalSubtitles should be empty after empty metadata, got %d tracks", len(subs))
	}
}

// Bug scenario: 9anime sometimes returns subtitle URLs with blank entries
// ("url1,,url3") from a CDN concat bug. Empty URLs must be filtered, not stored
// as subtitle tracks (would cause mpv to fail on launch).
func TestNineAnimeProviderFetchStreamURLFiltersEmptySubtitleURLs(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)

	mock := &nineAnimeMockScraper{
		streamURL: "https://cdn.example.com/master.m3u8",
		metadata: map[string]string{
			"subtitles":       "https://cdn.example.com/en.vtt,,https://cdn.example.com/pt.vtt",
			"subtitle_labels": "English,,Portuguese - Brazilian Portuguese",
		},
	}
	provider := &nineAnimeProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}
	if _, err := provider.FetchStreamURL(context.Background(), episode, anime, "best"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subs := util.GetGlobalSubtitles()
	if len(subs) != 2 {
		t.Fatalf("expected 2 subtitle tracks after filtering blanks, got %d: %+v", len(subs), subs)
	}
}

// Bug scenario: when there are fewer labels than subtitle URLs, the provider used to
// panic with an index-out-of-range. Fix: missing labels must fall back to "Unknown".
func TestNineAnimeProviderFetchStreamURLMissingLabelsDefaultToUnknown(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)

	mock := &nineAnimeMockScraper{
		streamURL: "https://cdn.example.com/master.m3u8",
		metadata: map[string]string{
			// Three subtitle URLs, only one label supplied.
			"subtitles":       "https://cdn.example.com/en.vtt,https://cdn.example.com/pt.vtt,https://cdn.example.com/es.vtt",
			"subtitle_labels": "English",
		},
	}
	provider := &nineAnimeProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}
	if _, err := provider.FetchStreamURL(context.Background(), episode, anime, "best"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subs := util.GetGlobalSubtitles()
	if len(subs) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(subs))
	}
	// Tracks without a label should report "Unknown" language, not panic.
	if subs[1].Language != "unknown" {
		t.Fatalf("track 2 language = %q, want %q", subs[1].Language, "unknown")
	}
	if subs[2].Language != "unknown" {
		t.Fatalf("track 3 language = %q, want %q", subs[2].Language, "unknown")
	}
}

// When SubtitlesDisabled is set (user passed --no-subs), the provider must NOT
// store subtitle tracks — even if the scraper returns them.
func TestNineAnimeProviderFetchStreamURLSubtitlesDisabledSkipsStorage(t *testing.T) {
	t.Cleanup(resetNineAnimePlaybackState)
	util.SetSubtitlesDisabled(true)

	mock := &nineAnimeMockScraper{
		streamURL: "https://cdn.example.com/master.m3u8",
		metadata: map[string]string{
			"referer":         "https://rapid-cloud.co/",
			"subtitles":       "https://cdn.example.com/en.vtt",
			"subtitle_labels": "English",
		},
	}
	provider := &nineAnimeProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "677", Source: "9Anime"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}
	if _, err := provider.FetchStreamURL(context.Background(), episode, anime, "best"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Referer should still be stored (it's for HTTP requests, not mpv).
	if got := util.GetGlobalReferer(); got != "https://rapid-cloud.co/" {
		t.Fatalf("GlobalReferer = %q, want %q", got, "https://rapid-cloud.co/")
	}
	// But subtitles must be blocked.
	if subs := util.GetGlobalSubtitles(); len(subs) != 0 {
		t.Fatalf("expected no subtitle tracks when disabled, got %d", len(subs))
	}
}

// Scraper init failure (GetScraper returns error) must propagate for both
// FetchEpisodes and FetchStreamURL.
func TestNineAnimeProviderScraperInitFailurePropagates(t *testing.T) {
	t.Parallel()

	initErr := errors.New("9anime scraper unavailable")
	provider := &nineAnimeProvider{sm: &errLookup{err: initErr}}

	anime := &models.Anime{Name: "Naruto", URL: "677"}
	episode := &models.Episode{Number: "1", URL: "ep-1"}

	if _, err := provider.FetchEpisodes(context.Background(), anime); !errors.Is(err, initErr) {
		t.Fatalf("FetchEpisodes: expected initErr in chain, got: %v", err)
	}
	if _, err := provider.FetchStreamURL(context.Background(), episode, anime, "best"); !errors.Is(err, initErr) {
		t.Fatalf("FetchStreamURL: expected initErr in chain, got: %v", err)
	}
}

// nineAnimeSubtitleLanguage mapping covers the languages declared in source_providers.go.
func TestNineAnimeSubtitleLanguageMapping(t *testing.T) {
	t.Parallel()

	cases := []struct{ label, want string }{
		{"English", "eng"},
		{"Portuguese - Brazilian Portuguese", "por"},
		{"Spanish", "spa"},
		{"Japanese", "jpn"},
		{"French", "fre"},
		{"German", "ger"},
		{"Italian", "ita"},
		{"Arabic", "ara"},
		{"Klingon", "unknown"},
		{"", "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			if got := nineAnimeSubtitleLanguage(tc.label); got != tc.want {
				t.Fatalf("nineAnimeSubtitleLanguage(%q) = %q, want %q", tc.label, got, tc.want)
			}
		})
	}
}

func TestNineAnimeProviderKindAndSeasons(t *testing.T) {
	t.Parallel()
	p := &nineAnimeProvider{}
	if p.Kind() != source.NineAnime {
		t.Fatalf("Kind = %s, want %s", p.Kind(), source.NineAnime)
	}
	if p.HasSeasons() {
		t.Fatal("9Anime provider should not report HasSeasons=true")
	}
}
