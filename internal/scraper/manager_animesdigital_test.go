package scraper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesdigital"
)

// The adapter has to say which Referer player 1's video host checks: the
// player would otherwise send a generic one and the host refuses it
// (issue #229). Player 2's Blogger video needs none.
func TestAnimesDigitalAdapterTellsThePlayerTheReferer(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/video/"):
			_, _ = w.Write([]byte(`<iframe class="metaframe rptss" src="https://api.anivideo.net/videohls.php?d=` + srv.URL + `/hls/x.m3u8"></iframe>`))
		case r.URL.Path == "/hls/x.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:10,\npiece.ts\n"))
		default:
			w.WriteHeader(http.StatusPartialContent)
		}
	}))
	defer srv.Close()
	a := NewAnimesDigitalAdapterWithClient(animesdigital.NewClientForTest(srv.Client(), srv.URL))

	stream, meta, err := a.GetStreamURL(srv.URL + "/video/a/1/")
	if err != nil || stream != srv.URL+"/hls/x.m3u8" {
		t.Fatalf("stream = %q, err = %v", stream, err)
	}
	if meta["referer"] != animesdigital.PlayerReferer {
		t.Fatalf("referer = %q, want %q", meta["referer"], animesdigital.PlayerReferer)
	}
	if meta["source"] != "animesdigital" {
		t.Fatalf("source = %q", meta["source"])
	}
}
