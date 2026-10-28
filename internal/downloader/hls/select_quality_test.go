package hls

import "testing"

// TestSelectBestStreamPrefersHeight pins the variant choice: the tallest
// picture first, bandwidth only among equally tall variants or when the
// playlist gives no RESOLUTION at all.
func TestSelectBestStreamPrefersHeight(t *testing.T) {
	t.Parallel()
	d := &Downloader{}
	base := "https://cdn.test/hls/master.m3u8"
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{name: "1080p in an efficient codec beats a fatter 720p", lines: []string{
			"#EXTM3U",
			`#EXT-X-STREAM-INF:BANDWIDTH=6000000,RESOLUTION=1280x720,CODECS="avc1.64001f"`, "720.m3u8",
			`#EXT-X-STREAM-INF:BANDWIDTH=4500000,RESOLUTION=1920x1080,CODECS="hvc1.1.6.L120"`, "1080.m3u8",
		}, want: "https://cdn.test/hls/1080.m3u8"},
		{name: "same height takes the higher bandwidth", lines: []string{
			"#EXT-X-STREAM-INF:BANDWIDTH=3000000,RESOLUTION=1920x1080", "a.m3u8",
			"#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080", "b.m3u8",
			"#EXT-X-STREAM-INF:BANDWIDTH=9000000,RESOLUTION=1280x720", "c.m3u8",
		}, want: "https://cdn.test/hls/b.m3u8"},
		{name: "no resolution falls back to bandwidth", lines: []string{
			"#EXT-X-STREAM-INF:BANDWIDTH=800000", "low.m3u8",
			"#EXT-X-STREAM-INF:BANDWIDTH=4000000", "https://other.test/high.m3u8",
		}, want: "https://other.test/high.m3u8"},
	}
	for _, tt := range tests {
		if got := d.selectBestStream(tt.lines, base); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
