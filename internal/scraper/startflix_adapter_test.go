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
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sameServer sends every request to one test server whatever its host, so the
// adapter can be fed realistic URLs.
type sameServer struct{ target *url.URL }

func (s sameServer) RoundTrip(req *http.Request) (*http.Response, error) {
	out := req.Clone(req.Context())
	out.URL.Scheme, out.URL.Host, out.Host = s.target.Scheme, s.target.Host, s.target.Host
	return http.DefaultTransport.RoundTrip(out)
}

func newStartFlixTestAdapter(t *testing.T, h http.HandlerFunc) *StartFlixAdapter {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	client := startflix.NewClientForTest(&http.Client{Transport: sameServer{target}}, "https://www.startflix.test")
	return NewStartFlixAdapterWithClient(client)
}

func TestStartFlixAdapter_SearchAnime(t *testing.T) {
	t.Parallel()
	adapter := newStartFlixTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<div class="result-item"><article>
			<div class="image"><img src="https://image.tmdb.org/t/p/w92/x.jpg" /></div>
			<div class="details"><div class="title"><a href="https://www.startflix.test/filmes/oppenheimer/">Oppenheimer</a></div>
			<div class="meta"><span class="year">2023</span></div></div></article></div>`)
	})

	results, err := adapter.SearchAnime("oppenheimer")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Oppenheimer", results[0].Name)
	assert.Equal(t, "StartFlix", results[0].Source)
	assert.Equal(t, models.MediaTypeMovie, results[0].MediaType)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/x.jpg", results[0].ImageURL)
	assert.Equal(t, StartFlixType, adapter.GetType())
}

func TestStartFlixAdapter_GetAnimeEpisodesPointsAtTheAPIFlow(t *testing.T) {
	t.Parallel()
	_, err := (&StartFlixAdapter{}).GetAnimeEpisodes("https://www.startflix.test/series/x/")
	require.Error(t, err)
	_, err = (&StartFlixAdapter{}).GetAnimeEpisodesContext(context.Background(), "https://www.startflix.test/series/x/")
	require.ErrorContains(t, err, "GetStartFlixEpisodes", "the contextual form points at the api flow too")
}

// The metadata map is read by providers.applyPlaybackMetadata, which expects
// "subtitles" as a JSON array of {"url","language","label"}.
func TestStartFlixAdapter_GetStreamURLMetadataContract(t *testing.T) {
	t.Parallel()
	adapter := newStartFlixTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/movie.mp4" {
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://files.test/movie.mp4"
			data-subtitles="https://files.test/pt.vtt" data-type="jwplayer" data-id="1">Player #1</button>`)
	})

	streamURL, metadata, err := adapter.GetStreamURL("https://painel.test/filme/tt1")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/movie.mp4", streamURL)
	assert.Equal(t, "https://painel.test/", metadata["referer"])

	var subs []struct{ URL, Language, Label string }
	require.NoError(t, jsonx.Unmarshal([]byte(metadata["subtitles"]), &subs))
	require.Len(t, subs, 1)
	assert.Equal(t, "https://files.test/pt.vtt", subs[0].URL)
}

func TestStartFlixAdapter_GetStreamURLNoServer(t *testing.T) {
	t.Parallel()
	adapter := newStartFlixTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<button data-show-player="true" data-source="https://playembedapi.site/?v=x" data-type="iframe">P</button>`)
	})
	_, _, err := adapter.GetStreamURL("https://painel.test/episodio/1")
	require.ErrorIs(t, err, startflix.ErrNoSupportedServer)
}
