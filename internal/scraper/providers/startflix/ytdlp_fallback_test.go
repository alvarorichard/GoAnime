package startflix

import (
	"context"
	"testing"
)

// The generic yt-dlp fallback rescues titles offered only on hosts GoAnime
// has no dedicated resolver for: the first extracted direct URL wins.
func TestResolveStreamYtDlpFallback(t *testing.T) {
	t.Setenv("GOANIME_YTDLP_FALLBACK_TEST", "1")
	old := ytdlpResolveFn
	t.Cleanup(func() { ytdlpResolveFn = old })
	ytdlpResolveFn = func(_ context.Context, playerURL, _ string) string {
		if playerURL == "https://dood.test/e/abc123" {
			return "https://cdn.test/direct.mp4"
		}
		return ""
	}
	c := NewClientForTest(nil, "https://www.startflix.test")
	s, err := c.ResolveStream(context.Background(), []Player{
		{URL: "https://dood.test/e/abc123", Type: "iframe", Label: "HD Dublado"},
	})
	if err != nil {
		t.Fatalf("ResolveStream = %v, want the yt-dlp URL", err)
	}
	if s.URL != "https://cdn.test/direct.mp4" || s.Host != "dood.test" {
		t.Errorf("stream = %+v, want the yt-dlp direct URL on dood.test", s)
	}
}

// Without the fallback (unit-test mode) an unsupported-only listing still
// fails naming its hosts, so the user learns what was listed.
func TestResolveStreamUnsupportedNamesHosts(t *testing.T) {
	c := NewClientForTest(nil, "https://www.startflix.test")
	_, err := c.ResolveStream(context.Background(), []Player{
		{URL: "https://unknownhost.test/e/1", Type: "iframe"},
	})
	if err == nil {
		t.Fatal("want ErrNoSupportedServer for an unknown host")
	}
}
