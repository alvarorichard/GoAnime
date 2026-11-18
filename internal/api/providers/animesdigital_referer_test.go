package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesdigital"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// animesDigitalProviderFor wires the real provider and adapter to a client
// pointed at srv, which stands for the site and for its video host.
func animesDigitalProviderFor(srv *httptest.Server) *animesDigitalProvider {
	p := &animesDigitalProvider{}
	p.adapter.UnifiedScraper = scraper.NewAnimesDigitalAdapterWithClient(animesdigital.NewClientForTest(srv.Client(), srv.URL))
	p.once.Do(func() {})
	return p
}

// Player 1's video host answers 403 to every segment unless the Referer is the
// embedding player's, and the provider used to hand the player no Referer at
// all, so it fell back on a generic one the host refuses (issue #229: Wind
// Breaker never played). These tests walk provider, adapter and client.
func TestAnimesDigitalStreamCarriesThePlayersReferer(t *testing.T) {
	playlistStatus := http.StatusOK
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/video/a/1/":
			one := "https://api.anivideo.net/videohls.php?d=" + url.QueryEscape(srv.URL+"/hls/01.m3u8")
			_, _ = fmt.Fprintf(w, `<iframe class="metaframe rptss" src="%s"></iframe>`, strings.ReplaceAll(one, "&", "&#038;"))
		case "/hls/01.m3u8":
			w.WriteHeader(playlistStatus)
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:10,\nseg\n"))
		case "/hls/seg":
			// the video host's guard: pieces go only to the embedding player's Referer
			if r.Header.Get("Referer") != animesdigital.PlayerReferer {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	p := animesDigitalProviderFor(srv)
	episode := &models.Episode{URL: srv.URL + "/video/a/1/"}

	t.Run("the stream says which Referer the player must send", func(t *testing.T) {
		stream, err := p.FetchStream(context.Background(), episode, nil, "")
		require.NoError(t, err)
		assert.Equal(t, srv.URL+"/hls/01.m3u8", stream.URL)
		assert.Equal(t, animesdigital.PlayerReferer, stream.Referer)
	})

	t.Run("and the player is given it", func(t *testing.T) {
		util.ClearGlobalReferer()
		t.Cleanup(util.ClearGlobalReferer)
		_, err := p.FetchStreamURL(context.Background(), episode, &models.Anime{Source: "AnimesDigital"}, "")
		require.NoError(t, err)
		assert.Equal(t, animesdigital.PlayerReferer, util.GetGlobalReferer())
	})

	t.Run("a dead playlist with no player 2 plays nothing", func(t *testing.T) {
		playlistStatus = http.StatusNotFound
		_, err := p.FetchStream(context.Background(), episode, nil, "")
		require.Error(t, err)
	})
}
