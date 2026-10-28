package hls

import "testing"

// TestPlaylistReferencesResolvePerRFC3986 pins how variant, segment and init
// references resolve against the playlist URL: as a browser or ffmpeg would
// (RFC 3986 §5.2), not by gluing strings at the last slash. Root-relative
// and dot-segment paths are common in real playlists, and a signed playlist
// URL can carry a "/" inside its query string.
func TestPlaylistReferencesResolvePerRFC3986(t *testing.T) {
	t.Parallel()
	d := &Downloader{}
	const playlist = "https://cdn.test/a/b/index.m3u8"

	segments := []struct{ ref, want string }{
		{"seg0.ts", "https://cdn.test/a/b/seg0.ts"},
		{"/root/seg1.ts", "https://cdn.test/root/seg1.ts"},
		{"../up/seg2.ts", "https://cdn.test/a/up/seg2.ts"},
		{"//cdn2.test/x/seg3.ts", "https://cdn2.test/x/seg3.ts"},
		{"http-seg4.ts", "https://cdn.test/a/b/http-seg4.ts"},
		{"https://abs.test/seg5.ts", "https://abs.test/seg5.ts"},
		{"seg6.ts?t=1", "https://cdn.test/a/b/seg6.ts?t=1"},
	}
	var lines []string
	for _, s := range segments {
		lines = append(lines, "#EXTINF:4.0,", s.ref)
	}
	pl, err := d.parseMediaPlaylistLines(append([]string{"#EXTM3U", "#EXT-X-MAP:URI=\"/init/main.mp4\""}, lines...), playlist)
	if err != nil {
		t.Fatal(err)
	}
	if pl.InitSegmentURL != "https://cdn.test/init/main.mp4" {
		t.Errorf("init segment = %q, want https://cdn.test/init/main.mp4", pl.InitSegmentURL)
	}
	if len(pl.Segments) != len(segments) {
		t.Fatalf("parsed %d segments, want %d", len(pl.Segments), len(segments))
	}
	for i, s := range segments {
		if got := pl.Segments[i].URL; got != s.want {
			t.Errorf("segment %q resolved to %q, want %q", s.ref, got, s.want)
		}
	}

	// A slash inside the playlist URL's query is not a directory.
	pl, err = d.parseMediaPlaylistLines([]string{"#EXTINF:4.0,", "seg.ts"}, "https://cdn.test/hls/index.m3u8?sig=ab/cd")
	if err != nil {
		t.Fatal(err)
	}
	if got := pl.Segments[0].URL; got != "https://cdn.test/hls/seg.ts" {
		t.Errorf("segment under a signed playlist URL = %q, want https://cdn.test/hls/seg.ts", got)
	}

	variants := []struct{ ref, want string }{
		{"/v/720.m3u8", "https://cdn.test/v/720.m3u8"},
		{"../v/720.m3u8", "https://cdn.test/a/v/720.m3u8"},
		{"720.m3u8", "https://cdn.test/a/b/720.m3u8"},
	}
	for _, v := range variants {
		got := d.selectBestStream([]string{"#EXT-X-STREAM-INF:BANDWIDTH=1,RESOLUTION=1280x720", v.ref}, playlist)
		if got != v.want {
			t.Errorf("variant %q resolved to %q, want %q", v.ref, got, v.want)
		}
	}
}
