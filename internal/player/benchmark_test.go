package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
)

func BenchmarkResolveSeasonForEpisode(b *testing.B) {
	benchCases := []struct {
		name string
		snap mediaSnapshot
		ep   int
	}{
		{
			name: "NoSeasonMap",
			snap: mediaSnapshot{
				AnimeSeason: 1,
			},
			ep: 12,
		},
		{
			name: "SelectedSeasonLocalEpisode",
			snap: mediaSnapshot{
				AnimeSeason: 2,
				SeasonMap: []metadata.SeasonMapping{
					{Season: 1, StartEp: 1, EndEp: 23, EpisodeCount: 23},
					{Season: 2, StartEp: 24, EndEp: 46, EpisodeCount: 23},
				},
			},
			ep: 5,
		},
		{
			name: "AbsoluteEpisodeMapping",
			snap: mediaSnapshot{
				AnimeSeason: 1,
				SeasonMap: []metadata.SeasonMapping{
					{Season: 1, StartEp: 1, EndEp: 23, EpisodeCount: 23},
					{Season: 2, StartEp: 24, EndEp: 46, EpisodeCount: 23},
				},
			},
			ep: 30,
		},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = resolveSeasonForEpisode(benchCase.snap, benchCase.ep)
			}
		})
	}
}

func BenchmarkLooksLikeHLS(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"HLS", "https://cdn.example.com/video/master.m3u8"},
		{"HLSPath", "https://cdn.example.com/hls/token"},
		{"MP4", "https://cdn.example.com/video.mp4"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = LooksLikeHLS(benchCase.url)
			}
		})
	}
}

func BenchmarkHasUnsafeExtension(b *testing.B) {
	benchCases := []struct {
		name string
		url  string
	}{
		{"Unsafe", "https://cdn.example.com/video.aspx"},
		{"Safe", "https://cdn.example.com/video.mp4"},
	}

	for _, benchCase := range benchCases {
		benchCase := benchCase
		b.Run(benchCase.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = hasUnsafeExtension(benchCase.url)
			}
		})
	}
}
