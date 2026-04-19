package streaming

import "testing"

func BenchmarkShouldUseYtDLPDownload(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"HLS", "https://cdn.example.com/video/master.m3u8"},
		{"Blogger", "https://www.blogger.com/video.g?token=abc"},
		{"MP4", "https://cdn.example.com/video.mp4"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = ShouldUseYtDLPDownload(benchCase.url)
			}
		})
	}
}

func BenchmarkShouldUseNativeHLSDownload(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"HLS", "https://cdn.example.com/video/master.m3u8"},
		{"UnsafeExtension", "https://cdn.example.com/video.aspx"},
		{"MP4", "https://cdn.example.com/video.mp4"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = ShouldUseNativeHLSDownload(benchCase.url)
			}
		})
	}
}

func BenchmarkEstimatedProgressSizeBytes(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"Wix", "https://repackager.wixmp.com/video"},
		{"HLS", "https://cdn.example.com/video/master.m3u8"},
		{"PlainMP4", "https://cdn.example.com/video.mp4"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = EstimatedProgressSizeBytes(benchCase.url)
			}
		})
	}
}

func BenchmarkDeriveReferer(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"HTTP", "https://cdn.example.com/video/master.m3u8?token=abc"},
		{"Invalid", "not-a-valid-url"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = DeriveReferer(benchCase.url)
			}
		})
	}
}
