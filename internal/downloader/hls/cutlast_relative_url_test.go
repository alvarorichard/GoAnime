package hls

import (
	"strings"
	"testing"
)

// legacyBaseDir is the pre-Go-1.27 relative-URL base computation
// (strings.LastIndex arithmetic) used as the oracle for the differential test.
func legacyBaseDir(baseURL string) (string, bool) {
	if idx := strings.LastIndex(baseURL, "/"); idx != -1 {
		return baseURL[:idx+1], true
	}
	return "", false
}

// cutLastBaseDir mirrors the Go 1.27 rewrite used in selectBestStream and
// parseMediaPlaylistLines.
func cutLastBaseDir(baseURL string) (string, bool) {
	if base, _, ok := strings.CutLast(baseURL, "/"); ok {
		return base + "/", true
	}
	return "", false
}

func TestRelativeURLBaseMatchesLegacyImplementation(t *testing.T) {
	corpus := []string{
		"", "/", "a", "a/", "/a", "a/b", "a/b/",
		"https://cdn.example.com/hls/master.m3u8",
		"https://cdn.example.com/hls/",
		"https://cdn.example.com",
		"https://cdn.example.com/",
		"https://cdn.example.com/a/b/c/index.m3u8?token=x",
		"master.m3u8",
		"//cdn.example.com/x.m3u8",
	}
	for _, in := range corpus {
		wantBase, wantOK := legacyBaseDir(in)
		gotBase, gotOK := cutLastBaseDir(in)
		if gotBase != wantBase || gotOK != wantOK {
			t.Errorf("base(%q) = (%q,%v), legacy = (%q,%v)", in, gotBase, gotOK, wantBase, wantOK)
		}
	}
}

func TestSelectBestStream_ResolvesRelativeVariant(t *testing.T) {
	d := NewDownloader()
	lines := []string{
		"#EXTM3U",
		"#EXT-X-STREAM-INF:BANDWIDTH=800000",
		"480p/index.m3u8",
		"#EXT-X-STREAM-INF:BANDWIDTH=2400000",
		"1080p/index.m3u8",
	}
	got := d.selectBestStream(lines, "https://cdn.example.com/hls/master.m3u8")
	want := "https://cdn.example.com/hls/1080p/index.m3u8"
	if got != want {
		t.Errorf("selectBestStream() = %q, want %q", got, want)
	}
}

func TestSelectBestStream_RelativeVariantWithoutSlashInBaseIsSkipped(t *testing.T) {
	d := NewDownloader()
	lines := []string{
		"#EXTM3U",
		"#EXT-X-STREAM-INF:BANDWIDTH=800000",
		"480p.m3u8",
	}
	// No "/" in the base URL: the variant cannot be resolved, so nothing is
	// selected. This matches the pre-1.27 behaviour exactly.
	if got := d.selectBestStream(lines, "master.m3u8"); got != "" {
		t.Errorf("selectBestStream() = %q, want empty string", got)
	}
}

func TestParseMediaPlaylistLines_ResolvesRelativeSegments(t *testing.T) {
	d := NewDownloader()
	lines := []string{
		"#EXTM3U",
		"#EXT-X-TARGETDURATION:10",
		"#EXTINF:9.9,",
		"seg0.ts",
		"#EXTINF:9.9,",
		"sub/seg1.ts",
		"#EXTINF:9.9,",
		"https://other.example.com/seg2.ts",
		"#EXT-X-ENDLIST",
	}
	pl, err := d.parseMediaPlaylistLines(lines, "https://cdn.example.com/hls/index.m3u8")
	if err != nil {
		t.Fatalf("parseMediaPlaylistLines: %v", err)
	}
	want := []string{
		"https://cdn.example.com/hls/seg0.ts",
		"https://cdn.example.com/hls/sub/seg1.ts",
		"https://other.example.com/seg2.ts",
	}
	if len(pl.Segments) != len(want) {
		t.Fatalf("got %d segments, want %d", len(pl.Segments), len(want))
	}
	for i, w := range want {
		if pl.Segments[i].URL != w {
			t.Errorf("segment %d = %q, want %q", i, pl.Segments[i].URL, w)
		}
	}
	if !pl.EndList {
		t.Error("expected EndList to be true")
	}
}

func TestParseMediaPlaylistLines_BaseWithoutSlashGetsSuffixed(t *testing.T) {
	d := NewDownloader()
	lines := []string{"#EXTM3U", "#EXTINF:4,", "seg0.ts"}
	pl, err := d.parseMediaPlaylistLines(lines, "playlist")
	if err != nil {
		t.Fatalf("parseMediaPlaylistLines: %v", err)
	}
	if len(pl.Segments) != 1 || pl.Segments[0].URL != "playlist/seg0.ts" {
		t.Fatalf("got %+v, want single segment playlist/seg0.ts", pl.Segments)
	}
}
