package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	kind source.SourceKind
}

func (f *fakeProvider) Kind() source.SourceKind { return f.kind }
func (f *fakeProvider) HasSeasons() bool        { return false }
func (f *fakeProvider) FetchEpisodes(_ context.Context, _ *models.Anime) ([]models.Episode, error) {
	return nil, nil
}
func (f *fakeProvider) FetchStreamURL(_ context.Context, _ *models.Episode, _ *models.Anime, _ string) (string, error) {
	return "", nil
}

func TestHasProvider_RegisteredProviders(t *testing.T) {
	// init() registers built-ins — verify they're discoverable.
	assert.True(t, HasProvider(source.AllAnime))
	assert.True(t, HasProvider(source.AnimeFire))
	assert.True(t, HasProvider(source.Goyabu))
}

func TestHasProvider_UnknownReturnsFalse(t *testing.T) {
	assert.False(t, HasProvider(source.SourceKind("doesnotexist")))
}

func TestRegisterProvider_AddsCustomKind(t *testing.T) {
	const customKind source.SourceKind = "custom-test-kind"
	require.False(t, HasProvider(customKind))

	RegisterProvider(customKind, func(_ *scraper.ScraperManager) Provider {
		return &fakeProvider{kind: customKind}
	})
	t.Cleanup(func() {
		factoriesMu.Lock()
		delete(factories, customKind)
		factoriesMu.Unlock()
		ResetForTesting()
	})

	assert.True(t, HasProvider(customKind))
}

func TestForKind_UnregisteredReturnsError(t *testing.T) {
	ResetForTesting()
	_, err := ForKind(source.SourceKind("never-registered"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no provider registered")
}

func TestForKind_ReturnsCachedInstance(t *testing.T) {
	const k source.SourceKind = "cache-test-kind"
	RegisterProvider(k, func(_ *scraper.ScraperManager) Provider {
		return &fakeProvider{kind: k}
	})
	t.Cleanup(func() {
		factoriesMu.Lock()
		delete(factories, k)
		factoriesMu.Unlock()
		ResetForTesting()
	})

	p1, err := ForKind(k)
	require.NoError(t, err)
	p2, err := ForKind(k)
	require.NoError(t, err)
	assert.Same(t, p1, p2, "ForKind must return cached instance")
}

func TestResetForTesting_ClearsCache(t *testing.T) {
	const k source.SourceKind = "reset-test-kind"
	RegisterProvider(k, func(_ *scraper.ScraperManager) Provider {
		return &fakeProvider{kind: k}
	})
	t.Cleanup(func() {
		factoriesMu.Lock()
		delete(factories, k)
		factoriesMu.Unlock()
		ResetForTesting()
	})

	p1, err := ForKind(k)
	require.NoError(t, err)
	ResetForTesting()
	p2, err := ForKind(k)
	require.NoError(t, err)
	assert.NotSame(t, p1, p2, "ResetForTesting must drop cache so a fresh instance is built")
}

func TestForAnime_ResolvesAndReturnsProvider(t *testing.T) {
	anime := &models.Anime{
		Name:   "Naruto",
		URL:    "https://animefire.io/animes/naruto",
		Source: "AnimeFire",
	}
	p, resolved, err := ForAnime(anime)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, source.AnimeFire, p.Kind())
	assert.Equal(t, source.AnimeFire, resolved.BestEffortKind())
}

func TestForAnime_UnknownFallsBackToAllAnime(t *testing.T) {
	// Unresolved source.Resolve → Unknown → BestEffortKind() returns AllAnime.
	anime := &models.Anime{Name: "X", URL: "x"}
	p, resolved, err := ForAnime(anime)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, source.AllAnime, p.Kind())
	assert.Equal(t, source.AllAnime, resolved.BestEffortKind())
}
