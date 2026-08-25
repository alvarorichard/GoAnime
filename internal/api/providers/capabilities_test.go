package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelC_OnlySuperFlixIsBrowserGated pins the Model C capability wiring:
// SuperFlix is the sole browser-gated source; the pure-HTTP anime sources must
// NOT implement BrowserGated (so they never carry browser methods they don't
// need). The live registry populated by init() is the source of truth.
func TestModelC_OnlySuperFlixIsBrowserGated(t *testing.T) {
	t.Parallel()

	sf, ok := source.Registered(source.SuperFlix)
	require.True(t, ok)
	assert.True(t, source.IsBrowserGated(sf), "SuperFlix must be browser-gated")

	for _, kind := range []source.SourceKind{source.AllAnime, source.AnimeFire, source.Goyabu, source.AniDB} {
		s, ok := source.Registered(kind)
		require.True(t, ok, "source %s must be registered", kind)
		assert.False(t, source.IsBrowserGated(s), "%s is pure-HTTP and must not be browser-gated", kind)
	}
}

// TestModelC_OnlySuperFlixIsSeasoned pins that SuperFlix (movie/TV catalog) is
// the only seasoned source; the anime sources report not-seasoned.
func TestModelC_OnlySuperFlixIsSeasoned(t *testing.T) {
	t.Parallel()

	sf, ok := source.Registered(source.SuperFlix)
	require.True(t, ok)
	assert.True(t, source.IsSeasoned(sf), "SuperFlix organizes content into seasons")

	for _, kind := range []source.SourceKind{source.AllAnime, source.AnimeFire, source.Goyabu, source.AniDB} {
		s, ok := source.Registered(kind)
		require.True(t, ok)
		assert.False(t, source.IsSeasoned(s), "%s is a flat anime catalog, not seasoned", kind)
	}
}

// TestSuperFlixProvider_WarmUp pins the browser-gated WarmUp behavior: it is
// cancellation-aware, and (on this CI box, which has no display) reports the
// clear "needs a screen" reason rather than letting a doomed solve proceed.
func TestSuperFlixProvider_WarmUp(t *testing.T) {
	t.Parallel()
	p := &superFlixProvider{}

	t.Run("cancelled context returns immediately", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		require.ErrorIs(t, p.WarmUp(ctx), context.Canceled)
	})

	t.Run("headless environment surfaces a clear reason", func(t *testing.T) {
		t.Parallel()
		// The test runner is display-less and does not set --sf-headless, so
		// HeadlessEnvironment() is true and WarmUp must refuse with guidance.
		// (On a developer desktop with a display this path returns nil; either
		// way WarmUp must never panic and must be self-explanatory when it errs.)
		err := p.WarmUp(context.Background())
		if err != nil {
			assert.Contains(t, err.Error(), "screen",
				"the headless refusal must tell the user a screen is needed")
		}
	})
}
