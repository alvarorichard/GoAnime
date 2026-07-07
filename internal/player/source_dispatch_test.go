package player

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubSource is a minimal source.Source for dispatch wiring tests.
type stubSource struct {
	desc source.Descriptor
	url  string
	err  error

	gotEpisode *models.Episode
	gotAnime   *models.Anime
	gotQuality string
}

func (s *stubSource) Describe() source.Descriptor { return s.desc }

func (s *stubSource) FetchEpisodes(_ context.Context, _ *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (s *stubSource) FetchStreamURL(_ context.Context, e *models.Episode, a *models.Anime, q string) (string, error) {
	s.gotEpisode, s.gotAnime, s.gotQuality = e, a, q
	return s.url, s.err
}

// TestGetVideoURLForEpisodeEnhanced_DispatchesThroughSourceRegistry proves the
// wrapper routes through source.ResolveSource → Source.FetchStreamURL instead
// of the legacy per-source chain.
func TestGetVideoURLForEpisodeEnhanced_DispatchesThroughSourceRegistry(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: "stub-kind", Priority: 1, URLMatchers: []string{"stub.example"}},
		url:  "https://cdn.example/video.mp4",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Name: "X", URL: "https://stub.example/anime/1"}
	ep := &models.Episode{Number: "1", URL: "https://stub.example/ep/1"}

	url, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/video.mp4", url)
	assert.Same(t, ep, stub.gotEpisode)
	assert.Same(t, anime, stub.gotAnime)
}

// TestGetVideoURLForEpisodeEnhanced_AllAnimeErrorNotSilentlyFallenBack pins the
// legacy policy: a resolved-AllAnime failure surfaces as an error, never the
// silent legacy extraction.
func TestGetVideoURLForEpisodeEnhanced_AllAnimeErrorNotSilentlyFallenBack(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AllAnime, Priority: 1, Explicit: []string{"AllAnime"}, ShortID: true},
		err:  assert.AnError,
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "AllAnime", URL: "hHjXnUTda"}
	ep := &models.Episode{Number: "1", URL: "hHjXnUTda"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AllAnime stream URL")
}

// TestGetVideoURLForEpisodeEnhanced_MovieTVErrorKeepsSourceLabel pins the
// legacy movie/TV error policy: no legacy fallback, label from anime.Source.
func TestGetVideoURLForEpisodeEnhanced_MovieTVErrorKeepsSourceLabel(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.SuperFlix, Priority: 1, Explicit: []string{"SuperFlix"}},
		err:  assert.AnError,
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "SuperFlix", URL: "1234", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "1", URL: "1234"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SuperFlix stream URL")
}

// TestGetVideoURLForEpisodeEnhanced_UnknownFallsBackToRegisteredAllAnime pins
// the legacy best-effort default: an unrecognized anime dispatches to the
// registered AllAnime source (made explicit/configurable in Phase 2).
func TestGetVideoURLForEpisodeEnhanced_UnknownFallsBackToRegisteredAllAnime(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AllAnime, Priority: 1, Explicit: []string{"AllAnime"}},
		url:  "https://cdn.example/best-effort.mp4",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Name: "Mystery", URL: "https://unknown.example/x"}
	ep := &models.Episode{Number: "1", URL: "https://unknown.example/x/1"}

	url, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/best-effort.mp4", url)
	assert.Same(t, anime, stub.gotAnime, "best-effort dispatch must reach the AllAnime source")
}
