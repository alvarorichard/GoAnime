//go:build live

package player

import (
	"io"
	"net/http"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/goyabu"
)

// TestGoyabuBlackCloverEpisode3Live reproduces the production failure reported
// on 2026-07-23 without launching mpv. It resolves the same WordPress episode,
// unwraps Blogger, starts the local proxy, and verifies that the CDN is serving
// the video through it.
func TestGoyabuBlackCloverEpisode3Live(t *testing.T) {
	client := goyabu.NewGoyabuClient()
	sourceURL, err := client.GetEpisodeStreamURL("https://goyabu.io/?p=2366")
	if err != nil {
		t.Fatalf("resolve Goyabu episode: %v", err)
	}

	proxyURL, err := extractActualVideoURL(sourceURL)
	if err != nil {
		t.Fatalf("resolve Blogger/googlevideo stream: %v", err)
	}
	t.Cleanup(StopBloggerProxy)

	resp, err := http.Head(proxyURL) //nolint:noctx // live diagnostic against local proxy
	if err != nil {
		t.Fatalf("probe local Blogger proxy: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("local Blogger proxy returned %s", resp.Status)
	}

	req, err := http.NewRequest(http.MethodGet, proxyURL, http.NoBody)
	if err != nil {
		t.Fatalf("create streaming probe: %v", err)
	}
	req.Header.Set("Range", "bytes=0-1023")
	streamResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stream video bytes through local Blogger proxy: %v", err)
	}
	defer func() { _ = streamResp.Body.Close() }()
	if streamResp.StatusCode < 200 || streamResp.StatusCode >= 300 {
		t.Fatalf("video streaming probe returned %s", streamResp.Status)
	}
	videoPrefix, err := io.ReadAll(io.LimitReader(streamResp.Body, 1024))
	if err != nil {
		t.Fatalf("read streamed video bytes: %v", err)
	}
	if len(videoPrefix) == 0 {
		t.Fatal("video streaming probe returned an empty body")
	}
}
