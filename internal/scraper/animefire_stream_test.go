package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func episodePageFixture(sources []struct{ src, quality string }) string {
	body := `<html><head><title>Anime Episode</title></head><body>`
	for _, s := range sources {
		body += fmt.Sprintf(`<div data-video-src="%s" data-quality="%s"></div>`, s.src, s.quality)
	}
	body += `</body></html>`
	return body
}

func TestAnimefireGetEpisodeStreamURLSelectsHighestQuality(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, episodePageFixture([]struct{ src, quality string }{
			{"https://cdn.example.com/ep1_480p.mp4", "480p"},
			{"https://cdn.example.com/ep1_720p.mp4", "720p"},
			{"https://cdn.example.com/ep1_360p.mp4", "360p"},
		}))
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/anime/1/episode/1")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/ep1_720p.mp4", streamURL)
}

func TestAnimefireGetEpisodeStreamURL1080pWins(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, episodePageFixture([]struct{ src, quality string }{
			{"https://cdn.example.com/ep_720p.mp4", "720p"},
			{"https://cdn.example.com/ep_1080p.mp4", "1080p"},
		}))
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL

	streamURL, err := client.GetEpisodeStreamURL(server.URL + "/anime/2/episode/1")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/ep_1080p.mp4", streamURL)
}

func TestAnimefireGetEpisodeStreamURLSingleSource(t *testing.T) {
	t.Parallel()

	const streamURL = "https://cdn.example.com/ep_only.mp4"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, episodePageFixture([]struct{ src, quality string }{
			{streamURL, ""},
		}))
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL

	resolvedURL, err := client.GetEpisodeStreamURL(server.URL + "/anime/3/episode/1")
	require.NoError(t, err)
	assert.Equal(t, streamURL, resolvedURL)
}

func TestAnimefireGetEpisodeStreamURLErrorsWhenNoSource(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body><div class="episode-player"></div></body></html>`)
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL

	_, err := client.GetEpisodeStreamURL(server.URL + "/anime/4/episode/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video source")
}

func TestAnimefireGetEpisodeStreamURLBlockedPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><head><title>Just a moment...</title></head><body><div id="cf-wrapper"></div></body></html>`)
	}))
	defer server.Close()

	client := NewAnimefireClient()
	client.baseURL = server.URL

	_, err := client.GetEpisodeStreamURL(server.URL + "/anime/5/episode/1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourceUnavailable), "challenge page should yield ErrSourceUnavailable, got: %v", err)
}
