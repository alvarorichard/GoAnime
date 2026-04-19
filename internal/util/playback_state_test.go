package util

import "testing"

func TestCurrentPlaybackStateReturnsCopy(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalAnimeSource("9Anime")
	SetGlobalReferer("https://rapid-cloud.co/")
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})

	state := CurrentPlaybackState()
	if state.AnimeSource != "9Anime" {
		t.Fatalf("AnimeSource = %q, want %q", state.AnimeSource, "9Anime")
	}
	if state.Referer != "https://rapid-cloud.co/" {
		t.Fatalf("Referer = %q, want %q", state.Referer, "https://rapid-cloud.co/")
	}
	if len(state.Subtitles) != 1 {
		t.Fatalf("len(Subtitles) = %d, want 1", len(state.Subtitles))
	}

	state.Subtitles[0].Label = "Mutated"
	current := GetGlobalSubtitles()
	if current[0].Label != "English" {
		t.Fatalf("stored subtitles were mutated through snapshot: %q", current[0].Label)
	}
}

func TestResetPlaybackStateClearsSessionFields(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalAnimeSource("9Anime")
	SetGlobalReferer("https://rapid-cloud.co/")
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})

	ResetPlaybackState()

	state := CurrentPlaybackState()
	if state.AnimeSource != "" {
		t.Fatalf("AnimeSource = %q, want empty", state.AnimeSource)
	}
	if state.Referer != "" {
		t.Fatalf("Referer = %q, want empty", state.Referer)
	}
	if len(state.Subtitles) != 0 {
		t.Fatalf("len(Subtitles) = %d, want 0", len(state.Subtitles))
	}
}
