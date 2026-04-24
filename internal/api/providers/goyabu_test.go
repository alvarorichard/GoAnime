package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

type stubLookup struct {
	scraper   scraper.UnifiedScraper
	requested []scraper.ScraperType
	mu        sync.Mutex
}

func (s *stubLookup) GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error) {
	s.mu.Lock()
	s.requested = append(s.requested, scraperType)
	s.mu.Unlock()
	return s.scraper, nil
}

// errLookup simulates a ScraperManager that fails to return a scraper —
// as happens when the scraper binary is missing or the type is unknown.
type errLookup struct{ err error }

func (e *errLookup) GetScraper(_ scraper.ScraperType) (scraper.UnifiedScraper, error) {
	return nil, e.err
}

// mockEpisodeScraper extends mockUnifiedScraper with configurable episode errors.
type mockEpisodeScraper struct {
	mockUnifiedScraper
	episodeErr error
}

func (m *mockEpisodeScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	m.episodesArg = animeURL
	if m.episodeErr != nil {
		return nil, m.episodeErr
	}
	return m.episodes, nil
}

type mockUnifiedScraper struct {
	episodes    []models.Episode
	streamURL   string
	streamErrs  []error
	streamCalls int
	episodesArg string
	streamArg   string
}

func (m *mockUnifiedScraper) SearchAnime(_ string, _ ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (m *mockUnifiedScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	m.episodesArg = animeURL
	return m.episodes, nil
}

func (m *mockUnifiedScraper) GetStreamURL(episodeURL string, _ ...any) (string, map[string]string, error) {
	m.streamArg = episodeURL
	if m.streamCalls < len(m.streamErrs) {
		err := m.streamErrs[m.streamCalls]
		m.streamCalls++
		return "", map[string]string{}, err
	}
	return m.streamURL, map[string]string{}, nil
}

func (m *mockUnifiedScraper) GetType() scraper.ScraperType {
	return scraper.GoyabuType
}

func TestGoyabuProviderFetchEpisodes(t *testing.T) {
	t.Parallel()

	mock := &mockUnifiedScraper{
		episodes: []models.Episode{{Number: "1", URL: "https://goyabu.to/ep/1"}},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &goyabuProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("FetchEpisodes returned error: %v", err)
	}

	if provider.Kind() != source.Goyabu {
		t.Fatalf("provider kind = %s, want %s", provider.Kind(), source.Goyabu)
	}
	if len(episodes) != 1 {
		t.Fatalf("FetchEpisodes returned %d episodes, want 1", len(episodes))
	}
	if mock.episodesArg != anime.URL {
		t.Fatalf("scraper received anime URL %q, want %q", mock.episodesArg, anime.URL)
	}
	if len(lookup.requested) != 1 || lookup.requested[0] != scraper.GoyabuType {
		t.Fatalf("GetScraper called with %v, want [%v]", lookup.requested, scraper.GoyabuType)
	}
}

func TestGoyabuProviderFetchStreamURL(t *testing.T) {
	t.Parallel()

	mock := &mockUnifiedScraper{
		streamURL: "https://cdn.example.com/video.m3u8",
	}
	lookup := &stubLookup{scraper: mock}
	provider := &goyabuProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episode := &models.Episode{Number: "1", URL: "https://goyabu.to/ep/1"}
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
	if len(lookup.requested) != 1 || lookup.requested[0] != scraper.GoyabuType {
		t.Fatalf("GetScraper called with %v, want [%v]", lookup.requested, scraper.GoyabuType)
	}
}

func TestGoyabuProviderFetchStreamURLRetriesSourceUnavailable(t *testing.T) {
	t.Parallel()

	mock := &mockUnifiedScraper{
		streamURL: "https://cdn.example.com/video.m3u8",
		streamErrs: []error{
			fmt.Errorf("temporary wrapper: %w", scraper.ErrSourceUnavailable),
			scraper.ErrSourceUnavailable,
		},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &goyabuProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episode := &models.Episode{Number: "1", URL: "https://goyabu.to/ep/1"}
	streamURL, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err != nil {
		t.Fatalf("FetchStreamURL returned error: %v", err)
	}

	if streamURL != mock.streamURL {
		t.Fatalf("FetchStreamURL = %q, want %q", streamURL, mock.streamURL)
	}
	if mock.streamCalls != 2 {
		t.Fatalf("FetchStreamURL retries = %d, want 2", mock.streamCalls)
	}
}

// Bug scenario: a non-retryable error (e.g. permission denied, 404) was previously
// causing the provider to keep retrying up to maxAttempts times and waste time.
// Fix: only ErrSourceUnavailable triggers retry; all other errors bail immediately.
func TestGoyabuProviderFetchStreamURLNonRetryableErrorStopsImmediately(t *testing.T) {
	t.Parallel()

	permanentErr := errors.New("404 episode not found")
	mock := &mockUnifiedScraper{
		streamErrs: []error{permanentErr},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &goyabuProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episode := &models.Episode{Number: "1", URL: "https://goyabu.to/ep/1"}
	_, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err == nil {
		t.Fatal("expected error from non-retryable failure, got nil")
	}

	// Must fail on first call, not retry.
	if mock.streamCalls != 1 {
		t.Fatalf("non-retryable error caused %d call(s), want exactly 1", mock.streamCalls)
	}
	if !strings.Contains(err.Error(), "404 episode not found") {
		t.Fatalf("error message should contain the original cause, got: %v", err)
	}
}

// Bug scenario: when ALL retries are exhausted (source stays unavailable through
// all maxAttempts attempts), the provider should return an error rather than
// silently returning an empty URL.
func TestGoyabuProviderFetchStreamURLExhaustsAllRetries(t *testing.T) {
	t.Parallel()

	// 4 consecutive ErrSourceUnavailable — fills all maxAttempts slots.
	mock := &mockUnifiedScraper{
		streamErrs: []error{
			scraper.ErrSourceUnavailable,
			scraper.ErrSourceUnavailable,
			scraper.ErrSourceUnavailable,
			scraper.ErrSourceUnavailable,
		},
	}
	lookup := &stubLookup{scraper: mock}
	provider := &goyabuProvider{sm: lookup}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episode := &models.Episode{Number: "1", URL: "https://goyabu.to/ep/1"}
	_, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err == nil {
		t.Fatal("expected error after exhausting all retries, got nil")
	}

	// Should have attempted maxAttempts (4) times.
	if mock.streamCalls != 4 {
		t.Fatalf("expected 4 attempts, got %d", mock.streamCalls)
	}
}

// Bug scenario: scraper initialization failure (e.g. binary missing) was silently
// swallowed in FetchEpisodes. Fix: error must propagate to caller.
func TestGoyabuProviderFetchEpisodesScraperInitFailurePropagates(t *testing.T) {
	t.Parallel()

	initErr := errors.New("goyabu scraper binary not found")
	provider := &goyabuProvider{sm: &errLookup{err: initErr}}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	_, err := provider.FetchEpisodes(context.Background(), anime)
	if err == nil {
		t.Fatal("expected scraper init error to propagate, got nil")
	}
	if !errors.Is(err, initErr) {
		t.Fatalf("expected initErr in chain, got: %v", err)
	}
}

// Same failure path for FetchStreamURL — scraper init failure must propagate.
func TestGoyabuProviderFetchStreamURLScraperInitFailurePropagates(t *testing.T) {
	t.Parallel()

	initErr := errors.New("goyabu scraper binary not found")
	provider := &goyabuProvider{sm: &errLookup{err: initErr}}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episode := &models.Episode{Number: "1", URL: "https://goyabu.to/ep/1"}
	_, err := provider.FetchStreamURL(context.Background(), episode, anime, "best")
	if err == nil {
		t.Fatal("expected scraper init error to propagate, got nil")
	}
	if !errors.Is(err, initErr) {
		t.Fatalf("expected initErr in chain, got: %v", err)
	}
}

// Regression: scraper returning zero episodes must not be treated as success
// that hides a real fetch problem. The empty slice is passed through as-is
// so the caller can decide whether to error (e.g. "no episodes found").
func TestGoyabuProviderFetchEpisodesEmptyListPassedThrough(t *testing.T) {
	t.Parallel()

	mock := &mockUnifiedScraper{episodes: []models.Episode{}}
	provider := &goyabuProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(episodes) != 0 {
		t.Fatalf("expected empty slice, got %d episodes", len(episodes))
	}
}

// Regression: scraper error in GetAnimeEpisodes must surface, not be swallowed.
func TestGoyabuProviderFetchEpisodesScraperErrorPropagates(t *testing.T) {
	t.Parallel()

	scraperErr := fmt.Errorf("cloudflare challenge: %w", scraper.ErrSourceUnavailable)
	mock := &mockEpisodeScraper{episodeErr: scraperErr}
	provider := &goyabuProvider{sm: &stubLookup{scraper: mock}}

	anime := &models.Anime{Name: "Naruto", URL: "https://goyabu.to/anime/naruto"}
	_, err := provider.FetchEpisodes(context.Background(), anime)
	if err == nil {
		t.Fatal("expected episode scraper error to propagate, got nil")
	}
	if !errors.Is(err, scraper.ErrSourceUnavailable) {
		t.Fatalf("expected ErrSourceUnavailable in chain, got: %v", err)
	}
}

// EpisodeNumber returns the string number when Number is set,
// falls back to Num, and returns empty for a nil pointer.
func TestEpisodeNumber(t *testing.T) {
	t.Parallel()

	if got := EpisodeNumber(nil); got != "" {
		t.Fatalf("nil episode: got %q, want %q", got, "")
	}
	if got := EpisodeNumber(&models.Episode{Number: "12"}); got != "12" {
		t.Fatalf("Number field: got %q, want %q", got, "12")
	}
	if got := EpisodeNumber(&models.Episode{Num: 7}); got != "7" {
		t.Fatalf("Num fallback: got %q, want %q", got, "7")
	}
	if got := EpisodeNumber(&models.Episode{}); got != "" {
		t.Fatalf("empty episode: got %q, want %q", got, "")
	}
}

// HasSeasons must always return false for Goyabu (it has no season concept).
func TestGoyabuProviderHasSeasons(t *testing.T) {
	t.Parallel()
	provider := &goyabuProvider{}
	if provider.HasSeasons() {
		t.Fatal("Goyabu provider should not report HasSeasons=true")
	}
}
