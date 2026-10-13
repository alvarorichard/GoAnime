package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/animedrive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// playerPage serves a two-server animeq player, rewriting media hosts to the
// test server. fhdAlive controls the FHD probe outcome; sd is always alive.
func animeDrivePlayerServer(t *testing.T, fhdAlive bool) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			if strings.Contains(r.URL.Path, "fhd") && !fhdAlive {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/sd.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/fhd.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
		`, server.URL, server.URL)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAnimeDriveAdapter_GetStreamURL_BestAndWorst(t *testing.T) {
	t.Parallel()

	server := animeDrivePlayerServer(t, true)
	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}

	best, meta, err := adapter.GetStreamURL(server.URL+"/episodio-1/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/fhd.mp4", best)
	assert.Equal(t, "animedrive", meta["source"])
	assert.Equal(t, server.URL+"/", meta["referer"])
	assert.Equal(t, "FHD", meta["audio_lang"])

	worst, _, err := adapter.GetStreamURL(server.URL+"/episodio-1/", "worst")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/sd.mp4", worst)
}

func TestAnimeDriveAdapter_GetStreamURL_DefaultQualityIsBest(t *testing.T) {
	t.Parallel()

	server := animeDrivePlayerServer(t, true)
	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}

	stream, _, err := adapter.GetStreamURL(server.URL + "/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/fhd.mp4", stream)
}

func TestAnimeDriveAdapter_GetStreamURL_SkipsDeadFHD(t *testing.T) {
	t.Parallel()

	server := animeDrivePlayerServer(t, false)
	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}

	stream, _, err := adapter.GetStreamURL(server.URL+"/episodio-1/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/sd.mp4", stream)
}

func TestAnimeDriveAdapter_GetStreamURL_HTTPErrorKeepsMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}
	_, metadata, err := adapter.GetStreamURL(server.URL + "/episodio-1/")
	require.Error(t, err)
	require.NotNil(t, metadata, "callers read metadata even on failure")
	assert.Equal(t, "animedrive", metadata["source"])
}

func TestAnimeDriveAdapter_GetStreamURL_NoServers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><p>no player</p></body></html>`)
	}))
	t.Cleanup(server.Close)

	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}
	_, _, err := adapter.GetStreamURL(server.URL + "/episodio-1/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks")
}

func TestAnimeDriveAdapter_SearchAndEpisodesDelegate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "s=") {
			_, _ = fmt.Fprintf(w, `<html><body><article class="item">
				<a href="/anime/test/"><img src="/t.jpg"/></a><h3>Test</h3>
			</article></body></html>`)
			return
		}
		_, _ = fmt.Fprint(w, `<html><body>
			<h1 class="entry-title">Test</h1>
			<ul class="episodios"><li><a href="/episodio-test-episodio-1/">Episode 1</a></li></ul>
		</body></html>`)
	}))
	t.Cleanup(server.Close)

	adapter := &AnimeDriveAdapter{client: animedrive.NewClientForTest(server.URL)}

	results, err := adapter.SearchAnime("test")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Test", results[0].Name)

	episodes, err := adapter.GetAnimeEpisodes(server.URL + "/anime/test/")
	require.NoError(t, err)
	require.Len(t, episodes, 1)
	assert.Equal(t, "1", episodes[0].Number)
}

func TestAnimeDriveAdapter_GetType(t *testing.T) {
	t.Parallel()

	adapter := &AnimeDriveAdapter{client: animedrive.NewAnimeDriveClient()}
	assert.Equal(t, AnimeDriveType, adapter.GetType())
}

func TestNewAnimeDriveAdapterForTest_PointsAtServer(t *testing.T) {
	t.Parallel()

	server := animeDrivePlayerServer(t, true)
	adapter, ok := NewAnimeDriveAdapterForTest(server.URL).(*AnimeDriveAdapter)
	require.True(t, ok, "must return the concrete adapter so cascade tests drive real glue")

	stream, _, err := adapter.GetStreamURL(server.URL+"/episodio-1/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/fhd.mp4", stream)
}
