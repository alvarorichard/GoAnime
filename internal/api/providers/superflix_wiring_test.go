package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fan-out gives each source perSourceSearchTimeout and abandons the
// goroutine when it fires. If the adapter drops the context, that goroutine
// keeps working — and for SuperFlix "keeps working" meant the transport's
// Retry-After loop sleeping and re-requesting against a host that was already
// rate-limiting us, so consecutive searches stacked up and the source could not
// recover ("dexter", 2026-09-20).
func TestSuperFlix_DeclaresTheContextualCapability(t *testing.T) {
	t.Parallel()
	adapter, err := scraper.NewAdapter(scraper.SuperFlixType)
	require.NoError(t, err)

	_, ok := adapter.(scraper.ContextualScraper)
	assert.True(t, ok,
		"SuperFlixAdapter must implement ContextualScraper, or the provider silently "+
			"falls back to the non-cancelable path and abandoned searches keep hitting the host")
}

// A cancelled context must stop the provider before it reaches the network.
func TestSuperFlix_SearchHonorsCancelledContext(t *testing.T) {
	t.Parallel()
	s, ok := source.Registered(source.SuperFlix)
	require.True(t, ok)
	sr := s.(source.Searchable)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sr.Search(ctx, "dexter")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
