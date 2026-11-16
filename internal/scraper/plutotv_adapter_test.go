package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/plutotv"
)

// The adapter must stay a ContextualScraper: the registry's fan-out hands it
// the caller's deadline only through that capability.
var (
	_ UnifiedScraper    = (*PlutoTVAdapter)(nil)
	_ ContextualScraper = (*PlutoTVAdapter)(nil)
)

const (
	adapterMovieID   = "62b9ea9ec4eb8f00140cb633"
	adapterSeriesID  = "6206bba315c8610013683974"
	adapterEpisodeID = "6206bbaa15c8610013683a28"
)

// plutoAdapterServer serves a session, a search, and one episode with two
// rungs whose title playlists are told apart by their rendition.
func plutoAdapterServer(t *testing.T) *plutotv.Client {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/v4/start":
			_, _ = fmt.Fprintf(w, `{"sessionToken": "t", "stitcherParams": "sid=S", "servers": {"search": %q, "vod": %q, "stitcher": %q}}`, srv.URL, srv.URL, srv.URL)
		case path == "/v1/search":
			if r.URL.Query().Get("q") == "boom" {
				http.Error(w, "x", http.StatusInternalServerError)
				return
			}
			_, _ = fmt.Fprintf(w, `{"data": [
			 {"id": %q, "slug": "em-nome-do-pai-2022-1-1", "name": "Em Nome do Pai", "type": "movie", "language": "PT"},
			 {"id": %q, "slug": "death-note-br", "name": "Death Note", "type": "series", "language": "EN"},
			 {"id": "5f6df5a173d7340007c559f7", "slug": "naruto-ptv1", "name": "Naruto", "type": "channel", "language": "PT"}]}`, adapterMovieID, adapterSeriesID)
		case strings.HasSuffix(path, "/master.m3u8"):
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1042180\n1042180/playlist.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=3321280\n3321280/playlist.m3u8\n"))
		case strings.HasSuffix(path, "/playlist.m3u8"):
			rendition := "hls_600"
			if strings.Contains(path, "/3321280/") {
				rendition = "hls_2400"
			}
			var b strings.Builder
			for i := 0; i < 8; i++ {
				fmt.Fprintf(&b, "#EXTINF:5,\n%s/media/c/hls/0-end/%s-%05d.ts\n", srv.URL, rendition, i)
			}
			_, _ = w.Write([]byte(b.String()))
		case strings.HasPrefix(path, "/media/"):
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:5,\ns.ts\n#EXT-X-ENDLIST\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return plutotv.NewClientForTest(srv.Client(), srv.URL)
}

func TestPlutoTVAdapterSearchConvertsTitlesAndSkipsChannels(t *testing.T) {
	a := NewPlutoTVAdapterWithClient(plutoAdapterServer(t))
	results, err := a.SearchAnimeContext(t.Context(), "x")
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "[Movie] [PT-BR] Em Nome do Pai", results[0].Name)
	assert.Equal(t, plutotv.MovieURL(adapterMovieID), results[0].URL)
	assert.Equal(t, "movie", string(results[0].MediaType))
	assert.Equal(t, "2022", results[0].Year)
	assert.Equal(t, plutotv.SourceName, results[0].Source)

	assert.Equal(t, "[TV] [English] Death Note", results[1].Name)
	assert.Equal(t, plutotv.SeriesURL(adapterSeriesID), results[1].URL)
	assert.Equal(t, "tv", string(results[1].MediaType))
}

func TestPlutoTVAdapterSearchWithoutAContextIsTheSameSearch(t *testing.T) {
	a := NewPlutoTVAdapterWithClient(plutoAdapterServer(t))
	viaContext, err := a.SearchAnimeContext(t.Context(), "x")
	require.NoError(t, err)
	plain, err := a.SearchAnime("x")
	require.NoError(t, err)
	assert.Equal(t, viaContext, plain)
}

func TestPlutoTVAdapterSearchPassesAFailureOn(t *testing.T) {
	a := NewPlutoTVAdapterWithClient(plutoAdapterServer(t))
	results, err := a.SearchAnimeContext(t.Context(), "boom")
	require.Error(t, err)
	assert.Nil(t, results)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = a.SearchAnimeContext(ctx, "x")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPlutoTVAdapterListsNothingItself(t *testing.T) {
	a := NewPlutoTVAdapterWithClient(plutoAdapterServer(t))
	for name, call := range map[string]func() ([]string, error){
		"plain": func() ([]string, error) { eps, err := a.GetAnimeEpisodes("u"); return make([]string, len(eps)), err },
		"context": func() ([]string, error) {
			eps, err := a.GetAnimeEpisodesContext(t.Context(), "u")
			return make([]string, len(eps)), err
		},
	} {
		t.Run(name, func(t *testing.T) {
			eps, err := call()
			require.ErrorContains(t, err, "GetPlutoTVEpisodes", "the error says where the listing is")
			assert.Empty(t, eps)
		})
	}
}

func TestPlutoTVAdapterStreamHonoursTheQualityOption(t *testing.T) {
	client := plutoAdapterServer(t)
	a := NewPlutoTVAdapterWithClient(client)
	episode := plutotv.EpisodeURL(adapterSeriesID, adapterEpisodeID)

	cases := []struct {
		options []any
		want    string
	}{
		{nil, "hls_2400"},
		{[]any{"best"}, "hls_2400"},
		{[]any{"worst"}, "hls_600"},
		{[]any{"360p"}, "hls_600"},
		{[]any{""}, "hls_2400"},
		{[]any{42}, "hls_2400"},
	}
	for _, tc := range cases {
		url, meta, err := a.GetStreamURL(episode, tc.options...)
		require.NoError(t, err, "%v", tc.options)
		assert.Contains(t, url, "/"+tc.want+".m3u8", "%v", tc.options)
		assert.Equal(t, map[string]string{"source": "plutotv"}, meta)
	}

	url, _, err := a.GetStreamURLContext(t.Context(), episode, "worst")
	require.NoError(t, err)
	assert.Contains(t, url, "/hls_600.m3u8")
}

func TestPlutoTVAdapterStreamFailureCarriesNoURL(t *testing.T) {
	a := NewPlutoTVAdapterWithClient(plutoAdapterServer(t))
	url, meta, err := a.GetStreamURL(plutotv.SeriesURL(adapterSeriesID))
	require.Error(t, err)
	assert.Empty(t, url)
	assert.Nil(t, meta)
}

func TestPlutoTVAdapterIdentity(t *testing.T) {
	client := plutoAdapterServer(t)
	a := NewPlutoTVAdapterWithClient(client)
	assert.Equal(t, PlutoTVType, a.GetType())
	assert.Same(t, client, a.GetClient())
}

func TestPlutoTVIsBuiltByNewAdapterAndNamedLikeTheOthers(t *testing.T) {
	a, err := NewAdapter(PlutoTVType)
	require.NoError(t, err)
	assert.IsType(t, &PlutoTVAdapter{}, a)
	assert.Equal(t, PlutoTVType, a.GetType())
	assert.Same(t, plutotv.Shared(), a.(*PlutoTVAdapter).GetClient(), "one client, one session, for the process")

	assert.Equal(t, "PlutoTV", scraperDisplayName(PlutoTVType))
	assert.Equal(t, plutotv.SourceName, scraperDisplayName(PlutoTVType))
	assert.Equal(t, "[Multilanguage]", scraperLanguageTag(PlutoTVType))
}

func TestPlutoTVTypeIsAppendedNotInserted(t *testing.T) {
	// The values are stitched into saved history: reordering one moves the
	// others.
	assert.Equal(t, AnimesFHDType+1, PlutoTVType)
	assert.Equal(t, ScraperType(0), AnimefireType)
}

func TestPlutoTVIsHealthChecked(t *testing.T) {
	assert.Contains(t, healthTargets(), PlutoTVType)
	assert.Equal(t, "dexter", DefaultHealthCheckQuery(PlutoTVType), "a title the movie sources are probed with, not an anime")
	assert.Equal(t, "naruto", DefaultHealthCheckQuery(AnimesFHDType), "the anime sources keep theirs")
}

func TestEveryScraperTypeHasANameAndAnAdapter(t *testing.T) {
	for st := AnimefireType; st <= PlutoTVType; st++ {
		assert.NotEqual(t, "Desconhecido", scraperDisplayName(st), "type %d", st)
		assert.NotEqual(t, "[Unknown]", scraperLanguageTag(st), "type %d", st)
	}
}
