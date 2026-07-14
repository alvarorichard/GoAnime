package scraper

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/allanime"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Adapter GetType tests — verify each adapter returns its registered ScraperType.
func TestAllAnimeAdapter_GetType(t *testing.T) {
	t.Parallel()
	a := &AllAnimeAdapter{}
	assert.Equal(t, AllAnimeType, a.GetType())
}

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
func TestAllAnimeAdapter_Client(t *testing.T) {
	t.Parallel()
	client := allanime.NewAllAnimeClient()
	a := &AllAnimeAdapter{client: client}
	assert.Same(t, client, a.Client())
}

func TestNewSuperFlixAdapterWithClient(t *testing.T) {
	t.Parallel()
	client := superflix.NewSuperFlixClient()
	a := NewSuperFlixAdapterWithClient(client)
	require.NotNil(t, a)
	assert.Same(t, client, a.GetClient())
	assert.Equal(t, SuperFlixType, a.GetType())
}
