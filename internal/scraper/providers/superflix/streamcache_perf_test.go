package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The cache replay must overlap the CDN liveness probe with the player-extras
// fetch instead of running it as a third serial round-trip. Verified
// deterministically (no wall-clock assertions): the extras handler blocks until
// it observes the probe request — if the probe only ran after ParallelExecute
// returned (the old serial shape), it could never arrive while extras is still
// pending and the flag stays false.
func TestStreamFromCache_ProbeOverlapsExtrasFetch(t *testing.T) {
	withFreshStreamCache(t)

	probeArrived := make(chan struct{})
	var probeSeenDuringExtras atomic.Bool

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"securedLink":"%s/cdn/master.m3u8"}`, srv.URL)
	})
	mux.HandleFunc("/cdn/master.m3u8", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Range"), "bytes=") {
			select {
			case <-probeArrived:
			default:
				close(probeArrived)
			}
		}
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte{0, 0})
	})
	mux.HandleFunc("/video/hash123", func(w http.ResponseWriter, _ *http.Request) {
		select {
		case <-probeArrived:
			probeSeenDuringExtras.Store(true)
		case <-time.After(3 * time.Second):
			// Old serial shape: the probe cannot fire while we block here.
		}
		_, _ = fmt.Fprint(w, realPlayerPage)
	})

	// Entry without cached extras, so the extras fetch actually runs.
	defaultStreamCache.put(streamCacheKey("serie", "42821", "1", "3"),
		streamCacheEntry{Host: srv.URL, Hash: "hash123"})

	c := NewClientForTest(srv.URL)
	res, ok := c.TryCachedStream(context.Background(), "serie", "42821", "1", "3")
	require.True(t, ok, "cache replay must succeed")
	assert.Contains(t, res.StreamURL, "master.m3u8")
	assert.True(t, probeSeenDuringExtras.Load(),
		"the CDN probe must run concurrently with the extras fetch, not after it")
}
