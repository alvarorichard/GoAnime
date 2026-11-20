package suggest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLiveProviders asks the real Kitsu and IMDb endpoints for a partial
// word, the way the box does while the user types. It is skipped with -short
// (CI) and when the network is unreachable, and exists to catch an upstream
// that changed shape: the recorded fixtures in testdata/ then need updating.
func TestLiveProviders(t *testing.T) {
	if testing.Short() {
		t.Skip("live network test skipped with -short")
	}
	e := Default()

	anime, err := e.Suggest(t.Context(), "narut", []Kind{KindAnime})
	if err != nil {
		t.Skipf("network unavailable: %v", err)
	}
	require.NotEmpty(t, anime, "kitsu answered nothing for a partial word")
	assert.Equal(t, "Naruto", anime[0].Title)

	shows, err := e.Suggest(t.Context(), "breaking ba", []Kind{KindMovie, KindSeries})
	require.NoError(t, err)
	require.NotEmpty(t, shows)
	assert.Equal(t, "Breaking Bad", shows[0].Title)
	assert.Equal(t, KindSeries, shows[0].Kind)

	start := time.Now()
	_, err = e.Suggest(t.Context(), "Breaking  BA", []Kind{KindMovie, KindSeries})
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 50*time.Millisecond, "the repeat is served from the cache")
}
