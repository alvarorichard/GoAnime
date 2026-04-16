package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
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
}

func (m *mockSourceProvider) Kind() SourceKind {
	return m.kind
}

func (m *mockSourceProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	m.episodesHit++
	m.gotAnime = anime
	return m.episodes, nil
}

func (m *mockSourceProvider) FetchStreamURL(anime *models.Anime, episode *models.Episode, quality string) (string, error) {
	m.streamHit++
	m.gotAnime = anime
	m.gotEpisode = episode
	m.gotQuality = quality
	return m.streamURL, nil
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
