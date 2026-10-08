package player

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/downloader/hls"
)

// TestDirectHTTPRefusesNonMediaBodies pins that the direct-HTTP downloader
// never saves something that is not the video as if it were. It is the
// fallback after a failed native HLS download, so it gets the .m3u8 URL
// itself: the server answers 200 with the playlist text, and saving that
// "succeeded" — which skipped the yt-dlp fallback that could have worked.
// An HTML error page served with 200 is the same failure.
func TestDirectHTTPRefusesNonMediaBodies(t *testing.T) {
	home := t.TempDir() // the output path must live under $HOME; not parallel
	t.Setenv("HOME", home)
	bodies := map[string]struct {
		contentType string
		body        string
	}{
		"/index.m3u8":        {"application/vnd.apple.mpegurl", "#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:4.0,\nseg0.ts\n"},
		"/index-untyped":     {"application/octet-stream", "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\nhi.m3u8\n"},
		"/error.mp4":         {"text/html; charset=utf-8", "<!DOCTYPE html><html><body>Access denied</body></html>"},
		"/error-untyped.mp4": {"", "  <html><head><title>403</title></head></html>"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := bodies[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if b.contentType != "" {
			w.Header().Set("Content-Type", b.contentType)
		}
		_, _ = w.Write([]byte(b.body))
	}))
	t.Cleanup(srv.Close)

	for path := range bodies {
		out := filepath.Join(home, "downloads", "ep.mp4")
		err := downloadDirectHTTPWithClient(srv.URL+path, out, nil, srv.Client())
		if err == nil {
			t.Errorf("%s: a non-media body was saved as the video", path)
		} else if strings.Contains(err.Error(), "output path") {
			t.Fatalf("%s: refused for the wrong reason: %v", path, err)
		}
		if _, statErr := os.Stat(out); statErr == nil {
			t.Errorf("%s: the non-media body was left on disk", path)
		}
	}
}

// TestDirectHTTPStillSavesMedia: the refusal is about what the body is, so a
// real media body — even one served as application/octet-stream — is saved.
func TestDirectHTTPStillSavesMedia(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	media := append([]byte("\x00\x00\x00\x20ftypisom"), make([]byte, 4096)...)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(media)
	}))
	t.Cleanup(srv.Close)

	out := filepath.Join(home, "downloads", "ep.mp4")
	if err := downloadDirectHTTPWithClient(srv.URL+"/v.mp4", out, nil, srv.Client()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil || len(got) != len(media) {
		t.Errorf("saved %d bytes (err %v), want %d", len(got), err, len(media))
	}
}

// hlsChainCalls records which download steps a chain run used.
type hlsChainCalls struct{ native, direct, ytdlp int }

// useHLSChainMocks swaps the three batch HLS steps for mocks. Not parallel:
// the steps are package-level seams.
func useHLSChainMocks(t *testing.T, native, direct, ytdlp func(string, string, *model) error) *hlsChainCalls {
	t.Helper()
	calls := &hlsChainCalls{}
	prevN, prevD, prevY := nativeHLSDownloadFn, directHTTPDownloadFn, ytdlpDownloadFn
	nativeHLSDownloadFn = func(u, p string, m *model) error { calls.native++; return native(u, p, m) }
	directHTTPDownloadFn = func(u, p string, m *model) error { calls.direct++; return direct(u, p, m) }
	ytdlpDownloadFn = func(u, p string, m *model) error { calls.ytdlp++; return ytdlp(u, p, m) }
	t.Cleanup(func() { nativeHLSDownloadFn, directHTTPDownloadFn, ytdlpDownloadFn = prevN, prevD, prevY })
	return calls
}

func ok(string, string, *model) error { return nil }

// TestHLSChainReachesYtDlpWhenNativeFails is the end-to-end regression for
// the batch chain: native HLS fails, direct HTTP — the real one — is handed
// the .m3u8 and gets the playlist text with a 200, and the chain must still
// reach yt-dlp instead of stopping on a "successful" text file.
func TestHLSChainReachesYtDlpWhenNativeFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:4.0,\nseg0.ts\n"))
	}))
	t.Cleanup(srv.Close)
	out := filepath.Join(home, "dl", "ep.mp4")

	calls := useHLSChainMocks(t,
		func(string, string, *model) error { return errors.New("segment 3: HTTP 503") },
		func(u, p string, m *model) error { return downloadDirectHTTPWithClient(u, p, m, srv.Client()) },
		func(_, p string, _ *model) error { return os.WriteFile(p, []byte("video from yt-dlp"), 0o600) },
	)
	if err := downloadHLSWithFallbacks(srv.URL+"/index.m3u8", out, nil, 1); err != nil {
		t.Fatal(err)
	}
	if *calls != (hlsChainCalls{native: 1, direct: 1, ytdlp: 1}) {
		t.Errorf("steps used = %+v, want native, direct, then yt-dlp once each", *calls)
	}
	if got, _ := os.ReadFile(out); string(got) != "video from yt-dlp" {
		t.Errorf("file = %q, want yt-dlp's output", got)
	}
}

// TestHLSChainRouting pins the rest of the chain: streams the native
// downloader cannot handle go straight to yt-dlp, and the first step that
// succeeds ends the chain.
func TestHLSChainRouting(t *testing.T) {
	fail := func(err error) func(string, string, *model) error {
		return func(string, string, *model) error { return err }
	}
	tests := []struct {
		name                 string
		native, direct, ytdl func(string, string, *model) error
		want                 hlsChainCalls
		wantErr              bool
	}{
		{name: "separate audio goes straight to yt-dlp",
			native: fail(fmt.Errorf("parse: %w", hls.ErrSeparateAudioTracks)), direct: ok, ytdl: ok,
			want: hlsChainCalls{native: 1, ytdlp: 1}},
		{name: "SAMPLE-AES goes straight to yt-dlp",
			native: fail(fmt.Errorf("parse: %w: METHOD=SAMPLE-AES", hls.ErrUnsupportedEncryption)), direct: ok, ytdl: ok,
			want: hlsChainCalls{native: 1, ytdlp: 1}},
		{name: "native success ends the chain", native: ok, direct: ok, ytdl: ok,
			want: hlsChainCalls{native: 1}},
		{name: "direct success ends the chain", native: fail(errors.New("boom")), direct: ok, ytdl: ok,
			want: hlsChainCalls{native: 1, direct: 1}},
		{name: "every step failing reports the last error",
			native: fail(errors.New("a")), direct: fail(errors.New("b")), ytdl: fail(errors.New("yt-dlp: c")),
			want: hlsChainCalls{native: 1, direct: 1, ytdlp: 1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := useHLSChainMocks(t, tt.native, tt.direct, tt.ytdl)
			err := downloadHLSWithFallbacks("https://cdn.test/index.m3u8", "/unused", nil, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if *calls != tt.want {
				t.Errorf("steps used = %+v, want %+v", *calls, tt.want)
			}
		})
	}
}
