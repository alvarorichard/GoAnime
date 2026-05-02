package api

import (
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSourceProvider struct {
	kind        SourceKind
	episodes    []models.Episode
	streamURL   string
	episodesHit int
	streamHit   int
	gotAnime    *models.Anime
	gotEpisode  *models.Episode
	gotQuality  string
	episodesErr error
	streamErr   error
}

func (m *mockSourceProvider) Kind() SourceKind {
	return m.kind
}

func (m *mockSourceProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	m.episodesHit++
	m.gotAnime = anime
	return m.episodes, m.episodesErr
}

func (m *mockSourceProvider) FetchStreamURL(anime *models.Anime, episode *models.Episode, quality string) (string, error) {
	m.streamHit++
	m.gotAnime = anime
	m.gotEpisode = episode
	m.gotQuality = quality
	return m.streamURL, m.streamErr
}

type mockUnifiedScraper struct {
	episodes       []models.Episode
	episodesErr    error
	streamURL      string
	streamErr      error
	gotEpisodesURL string
	gotStreamURL   string
	gotOptions     []any
}

func (m *mockUnifiedScraper) SearchAnime(string, ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (m *mockUnifiedScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	m.gotEpisodesURL = animeURL
	return m.episodes, m.episodesErr
}

func (m *mockUnifiedScraper) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	m.gotStreamURL = episodeURL
	m.gotOptions = options
	return m.streamURL, nil, m.streamErr
}

func (m *mockUnifiedScraper) GetType() scraper.ScraperType {
	return scraper.AllAnimeType
}

func TestFetchEpisodesWithResolvedSourceDispatchesProviders(t *testing.T) {
	t.Parallel()

	for _, kind := range []SourceKind{SourceAllAnime, SourceAnimefire, SourceAnimeDrive, SourceGoyabu} {
		kind := kind
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()

			provider := &mockSourceProvider{
				kind:     kind,
				episodes: []models.Episode{{Number: "1", Num: 1}},
			}
			anime := &models.Anime{Name: "Naruto", Source: string(kind)}
			resolved := ResolvedSource{Kind: kind, Name: string(kind)}

			episodes, err := fetchEpisodesWithResolvedSource(anime, resolved, func(requested SourceKind) (SourceProvider, bool) {
				if requested != kind {
					return nil, false
				}
				return provider, true
			})

			require.NoError(t, err)
			require.Len(t, episodes, 1)
			assert.Equal(t, 1, provider.episodesHit)
			assert.Same(t, anime, provider.gotAnime)
		})
	}
}

func TestFetchStreamURLWithResolvedSourceDispatchesProviders(t *testing.T) {
	t.Parallel()

	for _, kind := range []SourceKind{SourceAllAnime, SourceAnimefire, SourceAnimeDrive, SourceGoyabu} {
		kind := kind
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()

			provider := &mockSourceProvider{
				kind:      kind,
				streamURL: "https://example.com/video.m3u8",
			}
			anime := &models.Anime{Name: "Naruto", Source: string(kind)}
			episode := &models.Episode{Number: "1", URL: "https://example.com/episode/1"}
			resolved := ResolvedSource{Kind: kind, Name: string(kind)}

			streamURL, err := fetchStreamURLWithResolvedSource(anime, episode, "1080", resolved, func(requested SourceKind) (SourceProvider, bool) {
				if requested != kind {
					return nil, false
				}
				return provider, true
			})

			require.NoError(t, err)
			assert.Equal(t, "https://example.com/video.m3u8", streamURL)
			assert.Equal(t, 1, provider.streamHit)
			assert.Same(t, anime, provider.gotAnime)
			assert.Same(t, episode, provider.gotEpisode)
			assert.Equal(t, "1080", provider.gotQuality)
		})
	}
}

func TestFetchEpisodesWithResolvedSourceRequiresRegisteredProvider(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{Name: "Naruto"}
	resolved := ResolvedSource{Kind: SourceAnimefire, Name: string(SourceAnimefire)}

	_, err := fetchEpisodesWithResolvedSource(anime, resolved, func(SourceKind) (SourceProvider, bool) {
		return nil, false
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no source provider registered")
}

func TestFetchStreamURLWithResolvedSourceRequiresRegisteredProvider(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{Name: "Naruto"}
	episode := &models.Episode{Number: "1"}
	resolved := ResolvedSource{Kind: SourceAnimefire, Name: string(SourceAnimefire)}

	_, err := fetchStreamURLWithResolvedSource(anime, episode, "best", resolved, func(SourceKind) (SourceProvider, bool) {
		return nil, false
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no source provider registered")
}

func TestScraperBackedSourceProviderFetchEpisodesUsesResolvedScraper(t *testing.T) {
	original := providerScraperForKind
	t.Cleanup(func() {
		providerScraperForKind = original
	})

	mockScraper := &mockUnifiedScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
	}
	providerScraperForKind = func(kind SourceKind) (scraper.UnifiedScraper, error) {
		if kind != SourceAnimeDrive {
			return nil, fmt.Errorf("unexpected kind %s", kind)
		}
		return mockScraper, nil
	}

	provider := scraperBackedSourceProvider{kind: SourceAnimeDrive, streamQualityMode: streamQualityAuto}
	anime := &models.Anime{URL: "https://example.com/anime"}

	episodes, err := provider.FetchEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, episodes, 1)
	assert.Equal(t, anime.URL, mockScraper.gotEpisodesURL)
}

func TestScraperBackedSourceProviderFetchStreamURLModes(t *testing.T) {
	testCases := []struct {
		name        string
		mode        streamQualityMode
		quality     string
		wantOptions []any
	}{
		{name: "requested quality", mode: streamQualityRequested, quality: "1080", wantOptions: []any{"1080"}},
		{name: "auto quality", mode: streamQualityAuto, quality: "720", wantOptions: []any{"auto"}},
		{name: "no quality option", mode: streamQualityNone, quality: "720", wantOptions: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := providerScraperForKind
			t.Cleanup(func() {
				providerScraperForKind = original
			})

			mockScraper := &mockUnifiedScraper{streamURL: "https://example.com/video.m3u8"}
			providerScraperForKind = func(kind SourceKind) (scraper.UnifiedScraper, error) {
				if kind != SourceAnimefire {
					return nil, fmt.Errorf("unexpected kind %s", kind)
				}
				return mockScraper, nil
			}

			provider := scraperBackedSourceProvider{kind: SourceAnimefire, streamQualityMode: tc.mode}
			streamURL, err := provider.FetchStreamURL(nil, &models.Episode{URL: "https://example.com/episode"}, tc.quality)
			require.NoError(t, err)
			assert.Equal(t, "https://example.com/video.m3u8", streamURL)
			assert.Equal(t, "https://example.com/episode", mockScraper.gotStreamURL)
			assert.Equal(t, tc.wantOptions, mockScraper.gotOptions)
		})
	}
}

func TestAllAnimeSourceProviderFetchStreamURLUsesNormalizedInputs(t *testing.T) {
	original := providerAllAnimeEpisodeURLDirect
	t.Cleanup(func() {
		providerAllAnimeEpisodeURLDirect = original
	})

	var gotEpisodeNumber string
	var gotQuality string
	providerAllAnimeEpisodeURLDirect = func(anime *models.Anime, episodeNumber, quality string) (string, map[string]string, error) {
		assert.Equal(t, "naruto123abc", anime.URL)
		gotEpisodeNumber = episodeNumber
		gotQuality = quality
		return "https://example.com/video.m3u8", nil, nil
	}

	provider := allAnimeSourceProvider{}
	streamURL, err := provider.FetchStreamURL(&models.Anime{URL: "naruto123abc"}, &models.Episode{Num: 7}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/video.m3u8", streamURL)
	assert.Equal(t, "7", gotEpisodeNumber)
	assert.Equal(t, "best", gotQuality)
}

func TestNormalizeStreamQualityAndProviderEpisodeNumber(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "best", normalizeStreamQuality(""))
	assert.Equal(t, "720", normalizeStreamQuality("720"))

	assert.Equal(t, "", providerEpisodeNumber(nil))
	assert.Equal(t, "12", providerEpisodeNumber(&models.Episode{Number: "12", Num: 99}))
	assert.Equal(t, "3", providerEpisodeNumber(&models.Episode{Num: 3}))
	assert.Equal(t, "1", providerEpisodeNumber(&models.Episode{}))
}
