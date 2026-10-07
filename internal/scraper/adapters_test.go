package scraper

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Adapter GetType tests — verify each adapter returns its registered ScraperType.
func TestAnimefireAdapter_GetType(t *testing.T) {
	t.Parallel()
	a := &AnimefireAdapter{}
	assert.Equal(t, AnimefireType, a.GetType())
}

func TestGoyabuAdapter_GetType(t *testing.T) {
	t.Parallel()
	a := &GoyabuAdapter{}
	assert.Equal(t, GoyabuType, a.GetType())
}

// Adapter GetClient / Client tests
func TestNewStartFlixAdapterWithClient(t *testing.T) {
	t.Parallel()
	client := startflix.NewClient()
	a := NewStartFlixAdapterWithClient(client)
	require.NotNil(t, a)
	assert.Same(t, client, a.GetClient())
	assert.Equal(t, StartFlixType, a.GetType())
}
