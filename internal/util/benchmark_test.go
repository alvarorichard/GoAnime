package util

import "testing"

func BenchmarkCurrentSessionConfig(b *testing.B) {
	SetGlobalSource("goyabu")
	SetGlobalQuality("best")
	SetGlobalMediaType("anime")
	SetPreferredSubtitleLanguage("pt-BR")
	SetPreferredAudioLanguage("ja")
	SetSubtitlesDisabled(false)
	SetGlobalOutputDir("C:/anime")

	b.ReportAllocs()
	for range b.N {
		_ = CurrentSessionConfig()
	}
}

func BenchmarkGetGlobalSource(b *testing.B) {
	SetGlobalSource("goyabu")

	b.ReportAllocs()
	for range b.N {
		_ = GetGlobalSource()
	}
}

func BenchmarkCurrentDownloadRequest(b *testing.B) {
	SetGlobalDownloadRequest(&DownloadRequest{
		SeasonNum:     2,
		AllAnimeSmart: true,
		StartEpisode:  1,
		EndEpisode:    12,
	})
	b.Cleanup(ClearGlobalDownloadRequest)

	b.ReportAllocs()
	for range b.N {
		_ = CurrentDownloadRequest()
	}
}

func BenchmarkGetGlobalSubtitles(b *testing.B) {
	playbackStateMu.Lock()
	GlobalSubtitles = []SubtitleInfo{
		{URL: "https://subs.example.com/pt.vtt", Language: "pt-BR", Label: "Portuguese"},
		{URL: "https://subs.example.com/en.vtt", Language: "en", Label: "English"},
	}
	playbackStateMu.Unlock()
	b.Cleanup(ResetPlaybackState)

	b.ReportAllocs()
	for range b.N {
		_ = GetGlobalSubtitles()
	}
}

func BenchmarkCurrentPlaybackStateEmpty(b *testing.B) {
	ResetPlaybackState()

	b.ReportAllocs()
	for range b.N {
		_ = CurrentPlaybackState()
	}
}

func BenchmarkCurrentPlaybackStateWithSubtitles(b *testing.B) {
	playbackStateMu.Lock()
	GlobalSubtitles = []SubtitleInfo{
		{URL: "https://subs.example.com/pt.vtt", Language: "pt-BR", Label: "Portuguese"},
		{URL: "https://subs.example.com/en.vtt", Language: "en", Label: "English"},
	}
	GlobalReferer = "https://goyabu.to/"
	GlobalAnimeSource = "Goyabu"
	playbackStateMu.Unlock()
	b.Cleanup(ResetPlaybackState)

	b.ReportAllocs()
	for range b.N {
		_ = CurrentPlaybackState()
	}
}

func BenchmarkSetGlobalQuality(b *testing.B) {
	values := []string{"best", "720p"}
	b.ReportAllocs()
	for i := range b.N {
		SetGlobalQuality(values[i%len(values)])
	}
}
