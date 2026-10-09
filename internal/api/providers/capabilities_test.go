package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelC_NoSourceIsBrowserGated pins that GoAnime drives no browser: every
// registered source is plain HTTP, so none implements BrowserGated. A new
// source that needs one should make this test fail on purpose.
func TestModelC_NoSourceIsBrowserGated(t *testing.T) {
	t.Parallel()
	for _, kind := range []source.SourceKind{source.HiAnime, source.AnimeFire, source.Goyabu, source.StartFlix, source.TopCine} {
		s, ok := source.Registered(kind)
		require.True(t, ok, "source %s must be registered", kind)
		assert.False(t, source.IsBrowserGated(s), "%s is pure-HTTP and must not be browser-gated", kind)
	}
}

// TestModelC_MovieTVSourcesAreSeasoned pins that the movie/TV catalogs
// (StartFlix, TopCine) are seasoned; the anime sources report not-seasoned.
func TestModelC_MovieTVSourcesAreSeasoned(t *testing.T) {
	t.Parallel()

	for _, kind := range []source.SourceKind{source.StartFlix, source.TopCine} {
		s, ok := source.Registered(kind)
		require.True(t, ok)
		assert.True(t, source.IsSeasoned(s), "%s organizes content into seasons", kind)
	}

	for _, kind := range []source.SourceKind{source.HiAnime, source.AnimeFire, source.Goyabu} {
		s, ok := source.Registered(kind)
		require.True(t, ok)
		assert.False(t, source.IsSeasoned(s), "%s is a flat anime catalog, not seasoned", kind)
	}
}
