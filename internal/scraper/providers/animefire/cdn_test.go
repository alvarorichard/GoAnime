package animefire

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The API hands out a stream URL whether or not its CDN still has the file.
// Measured 2026-10-12: of 20 One Piece tracks sampled across the run, only the
// newest episode's playlist answered; every other one was a 404 from
// akumast.net while the API still said is_offline: false. Fresh API answers
// (cf-cache-status MISS, age 0) 404'd as well, so this is not a stale cache.
// Handing such a URL to mpv fails as a player error with no word about the
// source, so the client checks the playlist first.

// cdnTrack is one audio track of the fake episode and how its playlist URL
// answers.
type cdnTrack struct {
	audio   string
	status  int    // what the CDN answers; 200 serves body
	body    string // defaults to a playlist for 200, a 404 page otherwise
	offline bool   // listed by the API as is_offline with a null url
}

// cdnFixture serves one episode's API record and the CDN it points at from a
// single httptest server, so a playlist check never leaves the machine. The
// episode record names each track's playlist on this server.
type cdnFixture struct {
	srv     *httptest.Server
	apiHits atomic.Int32
	cdnHits atomic.Int32

	mu     sync.Mutex
	agents []string
}

func newCDNFixture(t *testing.T, tracks ...cdnTrack) *cdnFixture {
	t.Helper()
	f := &cdnFixture{}
	byPath := map[string]cdnTrack{}
	for _, tr := range tracks {
		byPath["/i/"+tr.audio+"/h.jpg"] = tr
	}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/episode/") {
			f.apiHits.Add(1)
			streams := make([]map[string]any, 0, len(tracks))
			for _, tr := range tracks {
				s := map[string]any{"audio": tr.audio, "is_offline": tr.offline, "url": nil}
				if !tr.offline {
					s["url"] = "http://" + r.Host + "/i/" + tr.audio + "/h.jpg"
				}
				streams = append(streams, s)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"streams": streams}})
			return
		}
		tr, ok := byPath[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.cdnHits.Add(1)
		f.mu.Lock()
		f.agents = append(f.agents, r.Header.Get("User-Agent"))
		f.mu.Unlock()
		if r.Header.Get("User-Agent") == "" {
			// What akumast.net answers a request with no User-Agent.
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("erro"))
			return
		}
		body := tr.body
		if body == "" {
			body = "<html><head><title>404 Not Found</title></head></html>"
			if tr.status == http.StatusOK {
				body = "#EXTM3U\n#EXT-X-VERSION:6\n"
			}
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(tr.status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *cdnFixture) url(audio string) string { return f.srv.URL + "/i/" + audio + "/h.jpg" }

func (f *cdnFixture) stream(t *testing.T) (string, error) {
	t.Helper()
	return NewClientForTest(f.srv.URL).GetEpisodeStreamURL("https://animefire.one/episode/ep1")
}

func TestStreamAPI_DeadDubFallsBackToTheSubtitledTrack(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t,
		cdnTrack{audio: "dublado", status: http.StatusNotFound},
		cdnTrack{audio: "legendado", status: http.StatusOK},
	)

	got, err := f.stream(t)

	require.NoError(t, err)
	assert.Equal(t, f.url("legendado"), got, "the dub has no file, so the track that plays is the one to hand over")
}

func TestStreamAPI_EveryTrackDeadOnTheCDNIsOffline(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t,
		cdnTrack{audio: "dublado", status: http.StatusNotFound},
		cdnTrack{audio: "legendado", status: http.StatusNotFound},
	)

	_, err := f.stream(t)

	require.Error(t, err, "a URL the CDN answers 404 for must not reach the player")
	assert.ErrorIs(t, err, netx.ErrMediaOffline, "the player tells the user to try another source from this")
	assert.NotErrorIs(t, err, netx.ErrSourceUnavailable)
}

// A 200 that is not a playlist (a web page, an error record) is not a stream.
func TestStreamAPI_AnswerThatIsNotAPlaylistIsNotAStream(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t,
		cdnTrack{audio: "dublado", status: http.StatusOK, body: "<!doctype html><html></html>"},
		cdnTrack{audio: "legendado", status: http.StatusOK},
	)

	got, err := f.stream(t)

	require.NoError(t, err)
	assert.Equal(t, f.url("legendado"), got)
}

// The live CDN answers 403 to a request without a User-Agent; the check must
// send the client's.
func TestStreamAPI_PlaylistCheckSendsTheUserAgent(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t, cdnTrack{audio: "dublado", status: http.StatusOK})

	_, err := f.stream(t)

	require.NoError(t, err)
	f.mu.Lock()
	defer f.mu.Unlock()
	require.Len(t, f.agents, 1, "the playlist must be checked before the URL is handed over")
	assert.Equal(t, netx.UserAgent, f.agents[0])
}

// A CDN that cannot be reached or errors out says nothing about the file. The
// URL is still handed over rather than hiding a file that may be there.
func TestStreamAPI_InconclusiveCheckStillHandsTheURLOver(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t,
		cdnTrack{audio: "dublado", status: http.StatusServiceUnavailable},
		cdnTrack{audio: "legendado", status: http.StatusNotFound},
	)

	got, err := f.stream(t)

	require.NoError(t, err)
	assert.Equal(t, f.url("dublado"), got)
}

// Missing content does not appear on a retry 100ms later. Retrying it tripled
// the API calls (and now the playlist checks) for every offline episode.
func TestRetrying_DoesNotRetryOfflineContent(t *testing.T) {
	t.Parallel()
	f := newCDNFixture(t,
		cdnTrack{audio: "dublado", offline: true},
		cdnTrack{audio: "legendado", offline: true},
	)
	c := NewClientForTest(f.srv.URL)
	c.maxRetries = 2

	_, err := c.GetEpisodeStreamURL("https://animefire.one/episode/ep1")

	require.ErrorIs(t, err, netx.ErrMediaOffline)
	assert.Equal(t, int32(1), f.apiHits.Load(), "offline content was asked for more than once")
}
