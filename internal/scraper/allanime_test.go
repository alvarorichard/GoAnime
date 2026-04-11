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

func TestAllAnimeSearchAnimeClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable, got: %v", err)
}

func TestAllAnimeSearchAnimeValidJSONParsesCorrectly(t *testing.T) {
	t.Parallel()

	const validJSON = `{"data":{"shows":{"edges":[{"_id":"abc","name":"One Piece","englishName":"One Piece","availableEpisodes":{"sub":1100}}]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

func TestAllAnimeSearchAnimeClassifies403AsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable for 403, got: %v", err)
}

func TestAllAnimeGetEpisodesListClassifiesHTMLPayloadAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable, got: %v", err)
}

func TestAllAnimeGetEpisodeURLClassifiesHTMLAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body>Rate limited</body></html>`)
	}))
	defer server.Close()

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	_, _, err := client.GetEpisodeURL("anime-id", "1", "sub", "best")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable for episode URL, got: %v", err)
}

func TestAllAnimeGetEpisodeURL503ClassifiesAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   AllAnimeReferer,
		apiBase:   server.URL,
		userAgent: UserAgent,
	}

	_, _, err := client.GetEpisodeURL("anime-id", "1", "sub", "best")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable for 503, got: %v", err)
}

func TestCheckHTMLResponseByteFallback(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: make(http.Header)}
	body := []byte("\r\n<!DOCTYPE html><html><body>blocked</body></html>")

	err := checkHTMLResponse(resp, body, "test-source")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "expected ErrSourceUnavailable from byte fallback, got: %v", err)
}

func TestCheckHTTPStatusNonBlockingCodeReturnsPlainError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := checkHTTPStatus(resp, "test-source")
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrSourceUnavailable), "404 should not be ErrSourceUnavailable, got: %v", err)
	assert.Contains(t, err.Error(), "404")
}
