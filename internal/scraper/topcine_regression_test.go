package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/topcine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. TopCine searches its own catalog but
// streams from the StartFlix panel, so the stub serves both: the TopCine
// search page and a panel players page with direct-file players.

// topcineCascadeStub serves a TopCine search hit plus a panel endpoint whose
// players resolve without any embed host, and records the referers the probes
// carry.
func topcineCascadeStub(t *testing.T) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var refs []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".mp4"):
			mu.Lock()
			refs = append(refs, r.Header.Get("Referer"))
			mu.Unlock()
			if strings.Contains(r.URL.Path, "dead") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			return
		case r.URL.Path == "/buscar":
			_, _ = fmt.Fprint(w, `<html><body>
				<article class="pagina-busca-card"><a href="/serie/dexter"><img src="https://image.tmdb.org/t/p/w342/d.jpg"><h2>Dexter</h2><small>2006 • Série</small></a></article>
			</body></html>`)
			return
		case r.URL.Path == "/episodio/7":
			_, _ = fmt.Fprintf(w, `
				<button data-show-player="true" data-source="%s/ok.mp4" data-type="jwplayer" data-id="1">Player #1 HD</button>
				<button data-show-player="true" data-source="%s/dead.mp4" data-type="jwplayer" data-id="2">Player #2 SD</button>`,
				srv.URL, srv.URL)
			return
		case r.URL.Path == "/episodio/unsupported":
			_, _ = fmt.Fprint(w, `<button data-show-player="true" data-source="https://vidsrcme.su/x" data-type="iframe">P</button>`)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), refs...)
	}
}

func topcineCascadeAdapter(t *testing.T, srv *httptest.Server) *TopCineAdapter {
	t.Helper()
	target, _ := url.Parse(srv.URL)
	hc := &http.Client{Transport: sameServer{target}}
	return NewTopCineAdapterWithClients(
		topcine.NewClientForTest(hc, "https://topcine.test"),
		startflix.NewClientForTest(hc, "https://www.startflix.test"),
	)
}

// TestTopCineAdapterCascade_SearchToStream drives the real TopCine search
// plus the real panel stream resolution: the search hit routes through the
// TopCine catalog, the episode URL through the shared panel, and the metadata
// is restamped to topcine.
func TestTopCineAdapterCascade_SearchToStream(t *testing.T) {
	t.Parallel()

	srv, refs := topcineCascadeStub(t)
	adapter := topcineCascadeAdapter(t, srv)

	results, err := adapter.SearchAnime("dexter")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Dexter", results[0].Name)
	assert.Equal(t, "TopCine", results[0].Source)

	stream, meta, err := adapter.GetStreamURL(srv.URL + "/episodio/7")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/ok.mp4", stream, "the dead direct file must be skipped")
	assert.Equal(t, "topcine", meta["source"], "the panel stamps startflix; the adapter restamps topcine")
	assert.Equal(t, srv.URL+"/", meta["referer"])
	require.NotEmpty(t, refs(), "the direct-file probe must reach the server")
}

// TestTopCineAdapterCascade_UnsupportedOnly surfaces the panel's no-resolver
// error instead of an empty success.
func TestTopCineAdapterCascade_UnsupportedOnly(t *testing.T) {
	t.Parallel()

	srv, _ := topcineCascadeStub(t)
	adapter := topcineCascadeAdapter(t, srv)

	_, meta, err := adapter.GetStreamURL(srv.URL + "/episodio/unsupported")
	require.ErrorIs(t, err, startflix.ErrNoSupportedServer)
	assert.Nil(t, meta)
}
