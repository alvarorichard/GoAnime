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

// TestGetVideoURLForEpisodeEnhanced_RegistrySourceErrorNotSilentlyFallenBack
// pins the policy: a failure from a registry-backed source surfaces as an
// error, never the silent legacy extraction. The guard used to name AllAnime;
// AniDB inherited that position when AllAnime was removed.
func TestGetVideoURLForEpisodeEnhanced_RegistrySourceErrorNotSilentlyFallenBack(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AniDB, Priority: 1, Explicit: []string{"AniDB"}},
		err:  assert.AnError,
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "AniDB", URL: "https://anidb.app/anime/x-1"}
	ep := &models.Episode{Number: "1", URL: "https://anidb.app/episode/1"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get AniDB stream URL")
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
		desc: source.Descriptor{Kind: source.AniDB, Priority: 1, Explicit: []string{"AllAnime"}},
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

// TestGetVideoURLForEpisodeEnhanced_UnknownIsReportedNotGuessed pins the policy
// after AllAnime's removal: an unrecognized anime has no plausible source to be
// dispatched to, so it fails with a reason instead of being handed to whichever
// source happens to be registered.
func TestGetVideoURLForEpisodeEnhanced_UnknownIsReportedNotGuessed(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &stubSource{
		desc: source.Descriptor{Kind: source.AniDB, Priority: 1, Explicit: []string{"AllAnime"}},
		url:  "https://cdn.example/best-effort.mp4",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Name: "Mystery", URL: "https://unknown.example/x"}
	ep := &models.Episode{Number: "1", URL: "https://unknown.example/x/1"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err, "an unrecognized source must not be guessed at")
	assert.Nil(t, stub.gotAnime, "no source must be reached for an Unknown resolution")
}

// gatedStubSource is a stubSource that also implements source.BrowserGated so
// the dispatch's WarmUp step can be exercised.
type gatedStubSource struct {
	stubSource
	warmUpErr error
	warmedUp  bool
}

func (g *gatedStubSource) WarmUp(_ context.Context) error {
	g.warmedUp = true
	return g.warmUpErr
}

// TestGetVideoURLForEpisodeEnhanced_WarmsUpBrowserGatedBeforeFetch pins the
// Model C consumer: a browser-gated source is warmed up before FetchStreamURL,
// and a WarmUp failure fails the dispatch fast — FetchStreamURL is never reached.
func TestGetVideoURLForEpisodeEnhanced_WarmsUpBrowserGatedBeforeFetch(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &gatedStubSource{
		desc:      source.Descriptor{Kind: source.SuperFlix, Priority: 1, Explicit: []string{"SuperFlix"}},
		url:       "https://cdn.example/should-not-be-reached.m3u8",
		warmUpErr: assert.AnError,
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "SuperFlix", URL: "1234", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "1", URL: "1234"}

	_, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.Error(t, err, "a failed warm-up must fail the dispatch")
	assert.True(t, stub.warmedUp, "WarmUp must run for a browser-gated source")
	assert.Nil(t, stub.gotAnime, "FetchStreamURL must NOT be reached when warm-up fails")
}

// TestGetVideoURLForEpisodeEnhanced_WarmUpSuccessProceedsToFetch pins the happy
// path: a browser-gated source that warms up cleanly proceeds to FetchStreamURL.
func TestGetVideoURLForEpisodeEnhanced_WarmUpSuccessProceedsToFetch(t *testing.T) {
	// Swaps the global source registry — not parallel.
	stub := &gatedStubSource{
		desc: source.Descriptor{Kind: source.SuperFlix, Priority: 1, Explicit: []string{"SuperFlix"}},
		url:  "https://cdn.example/sf.m3u8",
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "SuperFlix", URL: "1234", MediaType: models.MediaTypeTV}
	ep := &models.Episode{Number: "1", URL: "1234"}

	url, err := GetVideoURLForEpisodeEnhanced(context.Background(), ep, anime)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/sf.m3u8", url)
	assert.True(t, stub.warmedUp, "WarmUp must run before the fetch")
	assert.Same(t, anime, stub.gotAnime, "FetchStreamURL must run after a clean warm-up")
}
