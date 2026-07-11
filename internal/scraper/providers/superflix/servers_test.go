package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SuperFlix answers the player URL with one of two pages. The shell carries no
// tokens (so /player/bootstrap — and with it the whole server list — is
// unreachable); the real player carries PAGE_TOKEN and the content id. Note
// CSRF_TOKEN is EMPTY on the real page: it was retired, and the old code's
// insistence on it rejected every real page.
const (
	sfShellPage = `<html><head><title>Embed | Teerã</title></head><body><iframe src="x"></iframe></body></html>`

	sfRealPlayerPage = `<html><head><title>Player | Teerã</title></head><body><script>
var CSRF_TOKEN = "";
var PAGE_TOKEN = "eyJleHAiOjE3ODM4MDg3MzJ9";
var INITIAL_CONTENT_ID = 127972;
var CONTENT_TYPE = "serie";
var CURRENT_EPISODE = 1;
</script></body></html>`

	// The real bootstrap payload for Tehran S1E1: two servers, both Dublado, one
	// of them a direct MP4.
	sfBootstrapJSON = `{"data":{"options":[
		{"ID":159462,"type":1,"name":"Servidor 159462","is_file":false,"can_download":false},
		{"ID":"native_media:233831","type":1,"name":"MP4 Dublado","is_file":true,"can_download":false}
	]}}`
)

func TestSuperFlixServer_IDStringAndFallback(t *testing.T) {
	t.Parallel()
	// SuperFlix mixes numeric and string ids in the same list.
	assert.Equal(t, "159462", SuperFlixServer{ID: []byte(`159462`)}.IDString())
	assert.Equal(t, "native_media:233831", SuperFlixServer{ID: []byte(`"native_media:233831"`)}.IDString())
	assert.Empty(t, SuperFlixServer{ID: []byte(`{}`)}.IDString())

	assert.True(t, SuperFlixServer{ID: []byte(`"fallback_leg"`)}.IsFallback())
	assert.False(t, SuperFlixServer{ID: []byte(`159462`)}.IsFallback())
}

// TestGetServers_RetriesPastTheShell pins the fix that makes server selection
// possible at all: the site serves a token-less shell most of the time, so the
// player page must be retried until the real one arrives.
func TestGetServers_RetriesPastTheShell(t *testing.T) {
	t.Parallel()

	var pageHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/bootstrap") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, sfBootstrapJSON)
			return
		}
		// Shell on the first two hits, the real player on the third.
		if pageHits.Add(1) < 3 {
			_, _ = fmt.Fprint(w, sfShellPage)
			return
		}
		_, _ = fmt.Fprint(w, sfRealPlayerPage)
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.URL)
	servers, tokens, err := c.GetServers(context.Background(), "serie", "103913", "1", "1")
	require.NoError(t, err)

	assert.GreaterOrEqual(t, pageHits.Load(), int32(3), "must retry past the shell")
	require.NotNil(t, tokens)
	assert.Equal(t, "127972", tokens.ContentID)
	assert.Empty(t, tokens.CSRF, "CSRF is retired; requiring it is what broke this path")

	require.Len(t, servers, 2)
	assert.Equal(t, "Servidor 159462", servers[0].Name)
	assert.Equal(t, SuperFlixAudioDubbed, servers[0].Type)
	assert.False(t, servers[0].IsFile)
	assert.Equal(t, "MP4 Dublado", servers[1].Name)
	assert.True(t, servers[1].IsFile, "the MP4 mirror must be distinguishable — it is what makes a pick stable across episodes")
}

// "Too many requests" arrives as a 200 with a plain-text body, so it looks like a
// page. Retrying into it only earns a longer ban, so it must abort at once.
func TestGetServers_RateLimitAbortsImmediately(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = fmt.Fprint(w, "Too many requests")
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.URL)
	_, _, err := c.GetServers(context.Background(), "serie", "103913", "1", "1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSuperFlixRateLimited)
	assert.Equal(t, int32(1), hits.Load(), "a rate limit must not be retried into")
}

func TestGetServers_DropsFallbackPlaceholders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/bootstrap") {
			w.Header().Set("Content-Type", "application/json")
			// The site's own player filters these out before rendering its list.
			_, _ = fmt.Fprint(w, `{"data":{"options":[
				{"ID":"fallback_leg","type":2,"name":"Legendado"},
				{"ID":42,"type":2,"name":"Servidor 42"}
			]}}`)
			return
		}
		_, _ = fmt.Fprint(w, sfRealPlayerPage)
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.URL)
	servers, _, err := c.GetServers(context.Background(), "serie", "1", "1", "1")
	require.NoError(t, err)

	require.Len(t, servers, 1, "placeholder options are not playable sources")
	assert.Equal(t, "Servidor 42", servers[0].Name)
	assert.Equal(t, SuperFlixAudioSubtitled, servers[0].Type)
}

func TestGetServers_EmptyListIsNoServers(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/bootstrap") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"data":{"options":[]}}`)
			return
		}
		_, _ = fmt.Fprint(w, sfRealPlayerPage)
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.URL)
	_, _, err := c.GetServers(context.Background(), "serie", "1", "1", "1")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSuperFlixNoServers)
}

func TestIsRateLimited(t *testing.T) {
	t.Parallel()
	assert.True(t, isRateLimited("Too many requests"))
	assert.True(t, isRateLimited("too many requests"))
	assert.False(t, isRateLimited(sfRealPlayerPage))
	assert.False(t, isRateLimited(""))
	// A long page that happens to mention the phrase is a page, not a notice.
	assert.False(t, isRateLimited(strings.Repeat("x", 300)+"too many requests"))
}
