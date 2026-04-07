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

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	_, err := client.SearchAnime("One Piece")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable),
		"expected ErrSourceUnavailable, got: %v", err)
}

// TestAllAnimeSearchAnimeValidJSONParsesCorrectly confirms that a valid JSON
// response still passes through checkHTMLResponse and is parsed successfully.
func TestAllAnimeSearchAnimeValidJSONParsesCorrectly(t *testing.T) {
	t.Parallel()

	// Minimal valid GraphQL response that SearchAnime expects.
	const validJSON = `{"data":{"shows":{"edges":[{"_id":"abc","name":"One Piece","englishName":"One Piece","availableEpisodes":{"sub":1100}}]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, validJSON)
	}))
	defer server.Close()

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	results, err := client.SearchAnime("One Piece")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "One Piece")
}

// TestAllAnimeSearchAnimeClassifies403AsSourceUnavailable verifies that a 403
// Forbidden response is wrapped as ErrSourceUnavailable (source blocked).
func TestAllAnimeSearchAnimeClassifies403AsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	_, err := client.SearchAnime("One Piece")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable),
		"expected ErrSourceUnavailable for 403, got: %v", err)
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

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	_, err := client.GetEpisodesList("some-id", "sub")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable),
		"expected ErrSourceUnavailable, got: %v", err)
}
