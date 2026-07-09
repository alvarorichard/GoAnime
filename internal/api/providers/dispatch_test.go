package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// epStubSource records FetchEpisodes calls for dispatch tests.
type epStubSource struct {
	desc     source.Descriptor
	eps      []models.Episode
	err      error
	gotAnime *models.Anime
}

func (s *epStubSource) Describe() source.Descriptor { return s.desc }
func (s *epStubSource) FetchEpisodes(_ context.Context, a *models.Anime) ([]models.Episode, error) {
	s.gotAnime = a
	return s.eps, s.err
}
func (s *epStubSource) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	return "", nil
}

func TestFetchEpisodes_DispatchesThroughRegistry(t *testing.T) {
	// Swaps the global registry — not parallel.
	stub := &epStubSource{
		desc: source.Descriptor{Kind: source.Goyabu, Priority: 1, URLMatchers: []string{"goyabu"}},
		eps:  []models.Episode{{Number: "1"}, {Number: "2"}},
	}
	restore := source.SwapRegistryForTesting(stub)
	t.Cleanup(restore)

	eps, err := FetchEpisodes(context.Background(), &models.Anime{URL: "https://goyabu.io/x"})
	require.NoError(t, err)
	assert.Len(t, eps, 2)
	require.NotNil(t, stub.gotAnime)
	assert.Equal(t, "Goyabu", stub.gotAnime.Source, "empty source must be canonicalized to the resolved kind")
}

func TestFetchEpisodes_NilAnime(t *testing.T) {
	t.Parallel()
	_, err := FetchEpisodes(context.Background(), nil)
	require.Error(t, err)
}

func TestFetchEpisodes_UnknownFallsBackToBestEffort(t *testing.T) {
	// Swaps the global registry — not parallel.
	allAnime := &epStubSource{
		desc: source.Descriptor{Kind: source.AllAnime, Priority: 1, Explicit: []string{"AllAnime"}},
		eps:  []models.Episode{{Number: "1"}},
	}
	restore := source.SwapRegistryForTesting(allAnime)
	t.Cleanup(restore)

	// An unrecognized anime dispatches best-effort to the registered AllAnime.
	eps, err := FetchEpisodes(context.Background(), &models.Anime{Name: "Mystery", URL: "https://unknown.example/x"})
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Same(t, allAnime.gotAnime, allAnime.gotAnime)
}

func TestFetchEpisodes_ExplicitSourceNotOverwritten(t *testing.T) {
	// Swaps the global registry — not parallel.
	af := &epStubSource{
		desc: source.Descriptor{Kind: source.AnimeFire, Priority: 1, Explicit: []string{"Animefire.io", "AnimeFire"}},
	}
	restore := source.SwapRegistryForTesting(af)
	t.Cleanup(restore)

	anime := &models.Anime{Source: "Animefire.io", URL: "https://animefire.plus/x"}
	_, err := FetchEpisodes(context.Background(), anime)
	require.NoError(t, err)
	assert.Equal(t, "Animefire.io", anime.Source, "an explicitly-set source must be left untouched")
}
