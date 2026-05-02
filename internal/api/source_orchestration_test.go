package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests mutate package-level seams and playback globals.
// Keep them sequential and do not add t.Parallel().

func resetSourceOrchestrationSeams() {
	orchestrateGetSuperFlixEpisodes = GetSuperFlixEpisodes
	orchestrateGetFlixHQEpisodes = GetFlixHQEpisodes
	orchestrateGetNineAnimeEpisodes = GetNineAnimeEpisodes
	orchestrateGetSuperFlixStreamURL = GetSuperFlixStreamURL
	orchestrateGetFlixHQStreamURL = GetFlixHQStreamURL
	orchestrateGetNineAnimeStreamURL = GetNineAnimeStreamURL
}

func resetSourceOrchestrationGlobals() {
	util.ClearGlobalSubtitles()
	util.ClearGlobalReferer()
	util.SetGlobalAnimeSource("")
	util.GlobalNoSubs = false
}

func TestGetEpisodesByResolvedSourceDedicatedSources(t *testing.T) {
	testCases := []struct {
		name string
		kind SourceKind
		set  func(t *testing.T, anime *models.Anime, want []models.Episode)
	}{
		{
			name: "FlixHQ",
			kind: SourceFlixHQ,
			set: func(t *testing.T, anime *models.Anime, want []models.Episode) {
				orchestrateGetFlixHQEpisodes = func(got *models.Anime) ([]models.Episode, error) {
					assert.Same(t, anime, got)
					return want, nil
				}
			},
		},
		{
			name: "9Anime",
			kind: SourceNineAnime,
			set: func(t *testing.T, anime *models.Anime, want []models.Episode) {
				orchestrateGetNineAnimeEpisodes = func(got *models.Anime) ([]models.Episode, error) {
					assert.Same(t, anime, got)
					return want, nil
				}
			},
		},
		{
			name: "SuperFlix",
			kind: SourceSuperFlix,
			set: func(t *testing.T, anime *models.Anime, want []models.Episode) {
				orchestrateGetSuperFlixEpisodes = func(got *models.Anime) ([]models.Episode, error) {
					assert.Same(t, anime, got)
					return want, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetSourceOrchestrationSeams()
			t.Cleanup(resetSourceOrchestrationSeams)

			anime := &models.Anime{Name: "Dexter"}
			want := []models.Episode{{Number: "1", Num: 1}}
			tc.set(t, anime, want)

			episodes, err := getEpisodesByResolvedSource(anime, ResolvedSource{Kind: tc.kind, Name: string(tc.kind)})
			require.NoError(t, err)
			assert.Equal(t, want, episodes)
			assert.Equal(t, string(tc.kind), anime.Source)
		})
	}
}

func TestGetEpisodesByResolvedSourceProviderBackedWrapsErrors(t *testing.T) {
	oldProvider, hadProvider := defaultSourceProviders[SourceAnimefire]
	t.Cleanup(func() {
		if hadProvider {
			defaultSourceProviders[SourceAnimefire] = oldProvider
		} else {
			delete(defaultSourceProviders, SourceAnimefire)
		}
	})

	defaultSourceProviders[SourceAnimefire] = &mockSourceProvider{
		kind:        SourceAnimefire,
		episodesErr: fmt.Errorf("upstream failed"),
	}

	anime := &models.Anime{Name: "Naruto"}
	_, err := getEpisodesByResolvedSource(anime, ResolvedSource{Kind: SourceAnimefire, Name: string(SourceAnimefire)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes from Animefire.io")
	assert.Contains(t, err.Error(), "upstream failed")
	assert.Equal(t, "Animefire.io", anime.Source)
}

func TestGetStreamURLByResolvedSourceRejectsNilInputs(t *testing.T) {
	_, err := getStreamURLByResolvedSource(nil, &models.Episode{Number: "1"}, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil anime")

	_, err = getStreamURLByResolvedSource(&models.Anime{Source: "Animefire.io"}, nil, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil episode")
}

func TestGetStreamURLByResolvedSourceFlixHQStoresSubtitles(t *testing.T) {
	resetSourceOrchestrationSeams()
	resetSourceOrchestrationGlobals()
	t.Cleanup(resetSourceOrchestrationSeams)
	t.Cleanup(resetSourceOrchestrationGlobals)

	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://stale/sub.vtt", Label: "Stale", Language: "eng"}})
	orchestrateGetFlixHQStreamURL = func(anime *models.Anime, episode *models.Episode, quality string) (string, []models.Subtitle, error) {
		assert.Equal(t, "1080", quality)
		assert.Equal(t, "FlixHQ", anime.Source)
		assert.Equal(t, "1", episode.Number)
		return "https://example.com/video.m3u8", []models.Subtitle{
			{URL: "https://example.com/sub.vtt", Language: "eng", Label: "English"},
		}, nil
	}

	streamURL, err := getStreamURLByResolvedSource(&models.Anime{Source: "FlixHQ"}, &models.Episode{Number: "1"}, "1080")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/video.m3u8", streamURL)
	assert.Equal(t, "FlixHQ", util.GetGlobalAnimeSource())
	require.Len(t, util.GlobalSubtitles, 1)
	assert.Equal(t, "English", util.GlobalSubtitles[0].Label)
}

func TestGetStreamURLByResolvedSourceProviderBackedClearsStaleSubtitles(t *testing.T) {
	resetSourceOrchestrationGlobals()
	t.Cleanup(resetSourceOrchestrationGlobals)

	oldProvider, hadProvider := defaultSourceProviders[SourceAnimefire]
	t.Cleanup(func() {
		if hadProvider {
			defaultSourceProviders[SourceAnimefire] = oldProvider
		} else {
			delete(defaultSourceProviders, SourceAnimefire)
		}
	})

	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://stale/sub.vtt", Label: "Stale", Language: "eng"}})
	provider := &mockSourceProvider{
		kind:      SourceAnimefire,
		streamURL: "https://example.com/provider.m3u8",
	}
	defaultSourceProviders[SourceAnimefire] = provider

	anime := &models.Anime{Name: "Naruto", Source: "Animefire.io"}
	episode := &models.Episode{Number: "3", URL: "https://example.com/episode/3"}

	streamURL, err := getStreamURLByResolvedSource(anime, episode, "720")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/provider.m3u8", streamURL)
	assert.Equal(t, 1, provider.streamHit)
	assert.Equal(t, "720", provider.gotQuality)
	assert.Empty(t, util.GlobalSubtitles)
	assert.Equal(t, "Animefire.io", util.GetGlobalAnimeSource())
}

func TestGetStreamURLByResolvedSourceProviderBackedErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		streamURL   string
		streamErr   error
		wantErrLike string
		wantIsBack  bool
	}{
		{
			name:        "wraps regular provider errors",
			streamErr:   fmt.Errorf("provider failed"),
			wantErrLike: "failed to get stream URL from Animefire.io",
		},
		{
			name:       "passes through back request",
			streamErr:  scraper.ErrBackRequested,
			wantIsBack: true,
		},
		{
			name:        "rejects empty stream url",
			streamURL:   "",
			wantErrLike: "empty stream URL returned from Animefire.io",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetSourceOrchestrationGlobals()
			t.Cleanup(resetSourceOrchestrationGlobals)

			oldProvider, hadProvider := defaultSourceProviders[SourceAnimefire]
			t.Cleanup(func() {
				if hadProvider {
					defaultSourceProviders[SourceAnimefire] = oldProvider
				} else {
					delete(defaultSourceProviders, SourceAnimefire)
				}
			})

			defaultSourceProviders[SourceAnimefire] = &mockSourceProvider{
				kind:      SourceAnimefire,
				streamURL: tc.streamURL,
				streamErr: tc.streamErr,
			}

			_, err := getStreamURLByResolvedSource(&models.Anime{Source: "Animefire.io"}, &models.Episode{Number: "1", URL: "https://example.com/episode/1"}, "best")
			require.Error(t, err)
			if tc.wantIsBack {
				assert.True(t, errors.Is(err, scraper.ErrBackRequested))
				return
			}
			assert.Contains(t, err.Error(), tc.wantErrLike)
		})
	}
}

func TestAnimefireHermeticProviderFlow(t *testing.T) {
	resetSourceOrchestrationGlobals()
	t.Cleanup(resetSourceOrchestrationGlobals)

	oldProvider, hadProvider := defaultSourceProviders[SourceAnimefire]
	t.Cleanup(func() {
		if hadProvider {
			defaultSourceProviders[SourceAnimefire] = oldProvider
		} else {
			delete(defaultSourceProviders, SourceAnimefire)
		}
	})

	provider := &mockSourceProvider{
		kind:      SourceAnimefire,
		episodes:  []models.Episode{{Number: "1", Num: 1, URL: "https://example.com/episode/1"}},
		streamURL: "https://example.com/provider-flow.m3u8",
	}
	defaultSourceProviders[SourceAnimefire] = provider

	anime := &models.Anime{
		Name:   "[PT-BR] Naruto",
		Source: "Animefire.io",
		URL:    "https://animefire.plus/anime/naruto",
	}

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	require.Len(t, episodes, 1)
	assert.Equal(t, 1, provider.episodesHit)
	assert.Equal(t, "Animefire.io", anime.Source)

	streamURL, err := GetEpisodeStreamURL(&episodes[0], anime, "best")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/provider-flow.m3u8", streamURL)
	assert.Equal(t, 1, provider.streamHit)
	assert.Equal(t, "Animefire.io", util.GetGlobalAnimeSource())
}
