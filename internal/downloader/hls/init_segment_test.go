package hls

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fragmented-MP4 HLS carries its track definitions once, in the #EXT-X-MAP
// initialisation segment. Every media segment after it is a bare moof+mdat
// fragment with no headers of its own.
//
// This parser did not know the tag existed, so downloads concatenated the media
// segments alone. The result passed every check we had — 100+ MB on disk, well
// past the minimum-size guard — and could not be opened by anything:
//
//	trun track id unknown, no tfhd was found
//	error reading header  →  mpv: Failed to recognize file format (exit 2)
//
// AnimeFire's CDN serves exactly this shape, with the segments named .jpg.

func TestParsePlaylist_ReadsTheInitSegment(t *testing.T) {
	t.Parallel()
	const playlist = "#EXTM3U\n" +
		"#EXT-X-VERSION:6\n" +
		"#EXT-X-TARGETDURATION:14\n" +
		"#EXT-X-MAP:URI=\"i.jpg\"\n" +
		"#EXTINF:6.006,\n" +
		"1.jpg\n" +
		"#EXT-X-ENDLIST\n"

	got, err := new(Downloader).parseMediaPlaylistLines(strings.Split(playlist, "\n"), "https://cdn.example/i/token/p.jpg")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/i/token/i.jpg", got.InitSegmentURL,
		"the URI is relative to the playlist, like the segments beside it")
	require.Len(t, got.Segments, 1)
}

// The tag is an attribute list, so the URI is read by name — BYTERANGE may come
// first, and a future attribute must not shift what we read.
func TestParsePlaylist_InitSegmentURIIsReadByName(t *testing.T) {
	t.Parallel()
	const playlist = "#EXTM3U\n" +
		"#EXT-X-MAP:BYTERANGE=\"1234@0\",URI=\"init.mp4\"\n" +
		"#EXTINF:6,\n1.ts\n#EXT-X-ENDLIST\n"

	got, err := new(Downloader).parseMediaPlaylistLines(strings.Split(playlist, "\n"), "https://cdn.example/x/p.m3u8")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/x/init.mp4", got.InitSegmentURL)
}

func TestParsePlaylist_AbsoluteInitURIIsKept(t *testing.T) {
	t.Parallel()
	const playlist = "#EXTM3U\n" +
		"#EXT-X-MAP:URI=\"https://other.example/init.mp4\"\n" +
		"#EXTINF:6,\n1.ts\n#EXT-X-ENDLIST\n"

	got, err := new(Downloader).parseMediaPlaylistLines(strings.Split(playlist, "\n"), "https://cdn.example/x/p.m3u8")

	require.NoError(t, err)
	assert.Equal(t, "https://other.example/init.mp4", got.InitSegmentURL)
}

// MPEG-TS playlists are self-describing and carry no #EXT-X-MAP. They must not
// grow a phantom init segment, or every TS download would fetch a 404.
func TestParsePlaylist_TSPlaylistHasNoInitSegment(t *testing.T) {
	t.Parallel()
	const playlist = "#EXTM3U\n#EXTINF:6,\n1.ts\n#EXT-X-ENDLIST\n"

	got, err := new(Downloader).parseMediaPlaylistLines(strings.Split(playlist, "\n"), "https://cdn.example/x/p.m3u8")

	require.NoError(t, err)
	assert.Empty(t, got.InitSegmentURL)
}

// Order is the whole point: the init segment has to be the first bytes in the
// file, ahead of every media segment.
func TestDownload_WritesTheInitSegmentFirst(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var order []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		order = append(order, r.URL.Path)
		mu.Unlock()
		switch r.URL.Path {
		case "/p.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-MAP:URI=\"i.mp4\"\n" +
				"#EXTINF:6,\n1.ts\n#EXTINF:6,\n2.ts\n#EXT-X-ENDLIST\n"))
		case "/i.mp4":
			_, _ = w.Write([]byte("INIT"))
		case "/1.ts":
			_, _ = w.Write([]byte("AAAA"))
		case "/2.ts":
			_, _ = w.Write([]byte("BBBB"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "out.mp4")
	err := NewDownloaderWithClient(srv.Client()).DownloadWithProgress(context.Background(), srv.URL+"/p.m3u8", out, nil, nil)
	require.NoError(t, err)

	body, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "INITAAAABBBB", string(body),
		"the header must lead; media fragments alone are unopenable")
	assert.True(t, strings.HasPrefix(string(body), "INIT"))
}

// An init segment that cannot be fetched must fail the download. Continuing
// would write a file that looks complete and plays nowhere — which is how this
// went unnoticed in the first place.
func TestDownload_FailsWhenTheInitSegmentIsUnreachable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/p.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-MAP:URI=\"i.mp4\"\n#EXTINF:6,\n1.ts\n#EXT-X-ENDLIST\n"))
		case "/1.ts":
			_, _ = w.Write([]byte("AAAA"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "out.mp4")
	err := NewDownloaderWithClient(srv.Client()).DownloadWithProgress(context.Background(), srv.URL+"/p.m3u8", out, nil, nil)

	require.Error(t, err, "a missing header is a failed download, not a silent one")
	assert.Contains(t, err.Error(), "init segment")
}

func TestExtXMapURI_HandlesWhatThePlaylistsActuallyCarry(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct{ in, want string }{
		"plain":            {`URI="i.jpg"`, "i.jpg"},
		"lowercase attr":   {`uri="i.jpg"`, "i.jpg"},
		"byterange first":  {`BYTERANGE="500@0",URI="i.mp4"`, "i.mp4"},
		"absolute":         {`URI="https://h/i.mp4"`, "https://h/i.mp4"},
		"no uri attribute": {`BYTERANGE="500@0"`, ""},
		"empty":            {``, ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extXMapURI(tt.in))
		})
	}
}
