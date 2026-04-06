package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAllAnimeClient creates a fresh (non-singleton) client for unit tests
// so that overriding apiBase does not leak into other tests.
func newTestAllAnimeClient(apiBase string) *AllAnimeClient {
	return &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   apiBase,
		userAgent: UserAgent,
	}
}

// TestAllAnimeSearchAnimeClassifiesHTMLPayloadAsSourceUnavailable verifies that
// when AllAnime returns an HTML page (block / challenge page) instead of JSON,
// the error is wrapped as ErrSourceUnavailable rather than a raw parse error.
func TestAllAnimeSearchAnimeClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Just a moment...</title></head><body><div id="cf-wrapper">Cloudflare block</div></body></html>`)
	}))
	defer server.Close()

	client := newTestAllAnimeClient(server.URL)

	_, err := client.SearchAnime("One Piece")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable),
		"expected ErrSourceUnavailable, got: %v", err)
}

// TestAllAnimeGetEpisodesListClassifiesHTMLPayloadAsSourceUnavailable verifies
// the same classification for the episodes-list endpoint.
func TestAllAnimeGetEpisodesListClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body>Access Denied</body></html>`)
	}))
	defer server.Close()

	client := newTestAllAnimeClient(server.URL)

	_, err := client.GetEpisodesList("some-id", "sub")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable),
		"expected ErrSourceUnavailable, got: %v", err)
}
