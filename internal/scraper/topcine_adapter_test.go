package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTopCineTestAdapter(t *testing.T, h http.HandlerFunc) *TopCineAdapter {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: sameServer{target}}
	return NewTopCineAdapterWithClients(
		topcine.NewClientForTest(hc, "https://topcine.test"),
		startflix.NewClientForTest(hc, "https://www.startflix.test"),
	)
}

func TestTopCineAdapter_SearchAnime(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<article class="pagina-busca-card"><a href="/serie/breaking-bad">
			<img src="https://image.tmdb.org/t/p/w342/x.jpg"><h2>Breaking Bad</h2><small>2008 • Série</small></a></article>`)
	})

	results, err := adapter.SearchAnime("breaking bad")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Breaking Bad", results[0].Name)
	assert.Equal(t, "TopCine", results[0].Source)
	assert.Equal(t, "https://topcine.test/serie/breaking-bad", results[0].URL)
	assert.Equal(t, models.MediaTypeTV, results[0].MediaType)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/x.jpg", results[0].ImageURL)
	assert.Equal(t, TopCineType, adapter.GetType())
}

func TestTopCineAdapter_GetAnimeEpisodesPointsAtTheAPIFlow(t *testing.T) {
	t.Parallel()
	_, err := (&TopCineAdapter{}).GetAnimeEpisodesContext(context.Background(), "https://topcine.test/serie/x")
	require.ErrorContains(t, err, "GetTopCineEpisodes")
}

// TopCine episodes stream from the StartFlix panel, with StartFlix's metadata
// contract and TopCine named as the source.
func TestTopCineAdapter_GetStreamURLThroughThePanel(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/movie.mp4" {
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://files.test/movie.mp4" data-type="jwplayer" data-id="1">Player #1</button>`)
	})

	streamURL, metadata, err := adapter.GetStreamURL("https://painel.test/filme/tt34385135")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/movie.mp4", streamURL)
	assert.Equal(t, "topcine", metadata["source"])
	assert.Equal(t, "https://painel.test/", metadata["referer"])
}

func TestTopCineAdapter_GetStreamURLNoServer(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://playembedapi.site/?v=x" data-type="iframe">P</button>`)
	})
	_, metadata, err := adapter.GetStreamURL("https://painel.test/episodio/1")
	require.ErrorIs(t, err, startflix.ErrNoSupportedServer)
	assert.Nil(t, metadata)
}

func TestNewAdapter_TopCine(t *testing.T) {
	t.Parallel()
	ad, err := NewAdapter(TopCineType)
	require.NoError(t, err)
	assert.Equal(t, TopCineType, ad.GetType())
	assert.Equal(t, "TopCine", scraperDisplayName(TopCineType))
	assert.Equal(t, "[PT-BR]", scraperLanguageTag(TopCineType))
}

func TestTopCineAdapter_SearchError(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	_, err := adapter.SearchAnime("x")
	require.Error(t, err)
}

func TestTopCineAdapter_SearchEmptyQuery(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(http.ResponseWriter, *http.Request) {
		t.Error("a blank query was sent")
	})
	results, err := adapter.SearchAnime("  ")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestTopCineAdapter_GetAnimeEpisodes(t *testing.T) {
	t.Parallel()
	_, err := (&TopCineAdapter{}).GetAnimeEpisodes("https://topcine.test/serie/x")
	require.ErrorContains(t, err, "GetTopCineEpisodes")
}

func TestTopCineAdapter_GetClient(t *testing.T) {
	t.Parallel()
	client := topcine.NewClientForTest(http.DefaultClient, "https://topcine.test")
	adapter := NewTopCineAdapterWithClients(client, nil)
	assert.Same(t, client, adapter.GetClient())
}

func TestTopCineAdapter_StreamSubtitlesMetadata(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/movie.mp4" {
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://files.test/movie.mp4"
			data-subtitles="https://files.test/pt.vtt" data-type="jwplayer" data-id="1">Player #1</button>`)
	})
	_, metadata, err := adapter.GetStreamURLContext(context.Background(), "https://painel.test/filme/tt1")
	require.NoError(t, err)
	assert.Contains(t, metadata["subtitles"], "https://files.test/pt.vtt")
	assert.Equal(t, "files.test", metadata["host"])
}

func TestNewTopCineAdapterWithClients(t *testing.T) {
	t.Parallel()
	tc := topcine.NewClientForTest(http.DefaultClient, "https://topcine.test")
	sfx := startflix.NewClientForTest(http.DefaultClient, "https://www.startflix.test")
	a := NewTopCineAdapterWithClients(tc, sfx)
	assert.Same(t, tc, a.client)
	assert.Same(t, sfx, a.panel, "streams resolve on the StartFlix client given")
	var _ ContextualScraper = a
	var _ UnifiedScraper = a
}

func TestTopCineAdapter_GetType(t *testing.T) {
	t.Parallel()
	assert.Equal(t, TopCineType, (&TopCineAdapter{}).GetType())
	assert.NotEqual(t, StartFlixType, TopCineType, "TopCine is its own scraper type")
}

func TestTopCineAdapter_SearchAnimeContext(t *testing.T) {
	t.Parallel()
	var gotQuery string
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		fmt.Fprint(w, `<article class="pagina-busca-card"><a href="/filme/zona-zero"><h2>Zona Zero</h2><small>2026 • Filme</small></a></article>`)
	})

	results, err := adapter.SearchAnimeContext(context.Background(), "zona-zero")
	require.NoError(t, err)
	assert.Equal(t, "zona zero", gotQuery)
	require.Len(t, results, 1)
	assert.Equal(t, models.MediaTypeMovie, results[0].MediaType)
	assert.Equal(t, "2026", results[0].Year)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = adapter.SearchAnimeContext(ctx, "x")
	require.ErrorIs(t, err, context.Canceled, "the fan-out deadline reaches the request")
}

func TestTopCineAdapter_GetAnimeEpisodesContext(t *testing.T) {
	t.Parallel()
	eps, err := (&TopCineAdapter{}).GetAnimeEpisodesContext(context.Background(), "https://topcine.test/serie/x")
	require.ErrorContains(t, err, "GetTopCineEpisodes")
	assert.Nil(t, eps)
}

func TestTopCineAdapter_GetStreamURL(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ep.mp4" {
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://files.test/ep.mp4" data-type="jwplayer" data-id="1">P</button>`)
	})
	streamURL, metadata, err := adapter.GetStreamURL("https://painel.test/episodio/7")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep.mp4", streamURL)
	assert.Equal(t, "topcine", metadata["source"])
}

func TestTopCineAdapter_GetStreamURLContext(t *testing.T) {
	t.Parallel()
	adapter := newTopCineTestAdapter(t, func(http.ResponseWriter, *http.Request) {
		t.Error("requested with a cancelled context")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, metadata, err := adapter.GetStreamURLContext(ctx, "https://painel.test/episodio/7")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, metadata, "no metadata without a stream")
}
