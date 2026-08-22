package providers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
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

// searchStubSource implements source.Searchable for SearchAll tests.
type searchStubSource struct {
	epStubSource
	results []*models.Anime
	sErr    error
	delay   time.Duration
	calls   atomic.Int32
}

func (s *searchStubSource) Search(ctx context.Context, _ string) ([]*models.Anime, error) {
	s.calls.Add(1)
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return s.results, s.sErr
}

func newSearchStub(kind source.SourceKind, results []*models.Anime, err error) *searchStubSource {
	return &searchStubSource{
		desc:    source.Descriptor{Kind: kind, Priority: 1},
		results: results,
		sErr:    err,
	}
}

func TestSearchAll_FansOutOverRegistry(t *testing.T) {
	// Swaps the global registry — not parallel.
	a := newSearchStub(source.AllAnime, []*models.Anime{{Name: "[English] Naruto"}}, nil)
	g := newSearchStub(source.Goyabu, []*models.Anime{{Name: "[PT-BR] Naruto"}}, nil)
	restore := source.SwapRegistryForTesting(a, g)
	t.Cleanup(restore)

	got, err := SearchAll(context.Background(), "naruto")
	require.NoError(t, err)
	assert.Len(t, got, 2, "results from all searchable sources must be aggregated")
	assert.Equal(t, int32(1), a.calls.Load())
	assert.Equal(t, int32(1), g.calls.Load())
}

func TestSearchAll_SpecificKindFilter(t *testing.T) {
	// Swaps the global registry — not parallel.
	a := newSearchStub(source.AllAnime, []*models.Anime{{Name: "AA"}}, nil)
	g := newSearchStub(source.Goyabu, []*models.Anime{{Name: "GY"}}, nil)
	restore := source.SwapRegistryForTesting(a, g)
	t.Cleanup(restore)

	got, err := SearchAll(context.Background(), "x", source.Goyabu)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "GY", got[0].Name)
	assert.Equal(t, int32(0), a.calls.Load(), "non-selected source must not be searched")
	assert.Equal(t, int32(1), g.calls.Load())
}

func TestSearchAll_ToleratesPerSourceFailure(t *testing.T) {
	// Swaps the global registry — not parallel.
	ok := newSearchStub(source.AllAnime, []*models.Anime{{Name: "AA"}}, nil)
	bad := newSearchStub(source.Goyabu, nil, assert.AnError)
	restore := source.SwapRegistryForTesting(ok, bad)
	t.Cleanup(restore)

	got, err := SearchAll(context.Background(), "x")
	require.NoError(t, err, "one failing source must not fail the whole search")
	assert.Len(t, got, 1)
}

func TestSearchAll_AllFailReturnsError(t *testing.T) {
	// Swaps the global registry — not parallel.
	b1 := newSearchStub(source.AllAnime, nil, assert.AnError)
	b2 := newSearchStub(source.Goyabu, nil, assert.AnError)
	restore := source.SwapRegistryForTesting(b1, b2)
	t.Cleanup(restore)

	_, err := SearchAll(context.Background(), "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all sources failed")
}

// hangingSearchSource models the real adapters: its Search ignores ctx and
// blocks until released, so the only thing that ends the wait is the
// per-source deadline enforced by searchOneWithTimeout.
type hangingSearchSource struct {
	epStubSource
	release chan struct{}
}

func (s *hangingSearchSource) Search(context.Context, string) ([]*models.Anime, error) {
	<-s.release
	return nil, nil
}

func TestSearchOneWithTimeout_EnrichesWithOriginProbe(t *testing.T) {
	// Mutates the package-level perSourceSearchTimeout — not parallel.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(522) // Cloudflare origin down
	}))
	t.Cleanup(srv.Close)

	prev := perSourceSearchTimeout
	perSourceSearchTimeout = 20 * time.Millisecond
	t.Cleanup(func() { perSourceSearchTimeout = prev })

	stub := &hangingSearchSource{
		desc:    source.Descriptor{Kind: source.Goyabu, Priority: 1},
		release: make(chan struct{}),
	}
	t.Cleanup(func() { close(stub.release) }) // let the abandoned goroutine exit

	got := searchOneWithTimeout(context.Background(), activeSearcher{
		sr:       stub,
		kind:     source.Goyabu,
		probeURL: srv.URL,
	}, "naruto")

	require.Error(t, got.err)
	var diag *netx.SourceDiagnostic
	require.True(t, errors.As(got.err, &diag), "timeout must be enriched into a *netx.SourceDiagnostic")
	assert.Equal(t, 522, diag.StatusCode)
}

func TestSearchAll_NoSearchableSource(t *testing.T) {
	// Swaps the global registry with a non-searchable source — not parallel.
	restore := source.SwapRegistryForTesting(&epStubSource{desc: source.Descriptor{Kind: "plain", Priority: 1}})
	t.Cleanup(restore)

	_, err := SearchAll(context.Background(), "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no searchable source")
}
