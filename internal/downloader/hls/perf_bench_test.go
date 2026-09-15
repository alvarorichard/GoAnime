package hls

import (
	"fmt"
	"testing"
)

func BenchmarkSelectBestStream(b *testing.B) {
	dl := &Downloader{}
	baseURL := "https://cdn.example.com/video/master.m3u8"

	// Master playlist with 8 quality variants — typical for large CDNs.
	var lines []string
	lines = append(lines, "#EXTM3U")
	for i := 1; i <= 8; i++ {
		lines = append(lines,
			fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=1920x1080", i*500000),
			fmt.Sprintf("variant%d.m3u8", i),
		)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		dl.selectBestStream(lines, baseURL)
	}
}
