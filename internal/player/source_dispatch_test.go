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
// wrapper routes through source.Resolve → Source.FetchStreamURL instead
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

// TestGetVideoURLForEpisodeEnhanced_NilAnimeShortIDRoutesThroughRegistry pins
// Phase 2.1: URL-only resolution goes through source.ResolveURL and the
// minimal anime context comes from the resolution — no fake-AllAnime synthesis.
func TestGetVideoURLForEpisodeEnhanced_NilAnimeShortIDRoutesThroughRegistry(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AllAnime, Priority: 1, ShortID: true},
		url:  "https://cdn.example/from-id.mp4",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	ep := &models.Episode{Num: 2, URL: "hHjXnUTda"}
	url, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/from-id.mp4", url)
	require.NotNil(t, stub.gotAnime)
	assert.Equal(t, "AllAnime", stub.gotAnime.Source, "context must come from the resolution, not a hardcoded guess")
	assert.Equal(t, "hHjXnUTda", stub.gotAnime.URL)
	assert.Equal(t, "2", stub.gotEpisode.Number, "episode number must be derived from Num")
}

// TestGetVideoURLForEpisodeEnhanced_NilAnimeUnmatchedIDErrors pins Phase 2.1:
// a bare value no registered source claims is surfaced as an error, not guessed.
func TestGetVideoURLForEpisodeEnhanced_NilAnimeUnmatchedIDErrors(t *testing.T) {
	// Swaps the global source registry — not parallel.
	restore := source.SwapRegistryForTesting() // empty registry: nothing matches
	t.Cleanup(restore)

	ep := &models.Episode{Number: "1", URL: "shortid"}
	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot resolve stream without anime context")
}

// TestGetVideoURLForEpisodeEnhanced_StrictSourceDisablesBestEffort pins Phase
// 2.2: with GOANIME_STRICT_SOURCE set, an unrecognized source fails loudly
// instead of dispatching best-effort AllAnime.
func TestGetVideoURLForEpisodeEnhanced_StrictSourceDisablesBestEffort(t *testing.T) {
	// Swaps the global source registry and env — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AllAnime, Priority: 1, Explicit: []string{"AllAnime"}},
		url:  "https://cdn.example/never.mp4",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)
	t.Setenv("GOANIME_STRICT_SOURCE", "1")

	anime := &models.Anime{Name: "Mystery", URL: "https://unknown.example/x"}
	ep := &models.Episode{Number: "1", URL: "https://unknown.example/x/1"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized source")
	assert.Nil(t, stub.gotAnime, "no source must be reached in strict mode")
}

// TestGetVideoURLForEpisodeEnhanced_UnknownFallsBackToRegisteredAllAnime pins
// the default best-effort policy: an unrecognized anime dispatches to the
// registered AllAnime source (with a loud warning; disabled by GOANIME_STRICT_SOURCE).
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
