package util

import (
	"runtime"
	"slices"
	"strings"
	"testing"
)

// resetSubtitleState resets all subtitle-related global state between tests.
func resetSubtitleState() {
	ResetPlaybackState()
	GlobalNoSubs = false
	GlobalSubsLanguage = ""
	GlobalAudioLanguage = ""
}

// ─── Source Detection ────────────────────────────────────────────────────────

func TestSetGlobalAnimeSource(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalAnimeSource("9Anime")
	if got := GetGlobalAnimeSource(); got != "9Anime" {
		t.Errorf("GetGlobalAnimeSource() = %q, want %q", got, "9Anime")
	}

	SetGlobalAnimeSource("AllAnime")
	if got := GetGlobalAnimeSource(); got != "AllAnime" {
		t.Errorf("GetGlobalAnimeSource() = %q, want %q", got, "AllAnime")
	}

	SetGlobalAnimeSource("")
	if got := GetGlobalAnimeSource(); got != "" {
		t.Errorf("GetGlobalAnimeSource() = %q, want empty", got)
	}
}

func TestIs9AnimeSource(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	tests := []struct {
		source string
		want   bool
	}{
		{"9Anime", true},
		{"AllAnime", false},
		{"AnimeFire", false},
		{"FlixHQ", false},
		{"", false},
		{"9anime", false}, // case-sensitive
	}

	for _, tc := range tests {
		t.Run(tc.source, func(t *testing.T) {
			SetGlobalAnimeSource(tc.source)
			if got := Is9AnimeSource(); got != tc.want {
				t.Errorf("Is9AnimeSource() with source %q = %v, want %v", tc.source, got, tc.want)
			}
		})
	}
}

// ─── SetGlobalSubtitles / ClearGlobalSubtitles ──────────────────────────────

func TestSetGlobalSubtitles(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	subs := []SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
	}

	SetGlobalSubtitles(subs)

	if len(GlobalSubtitles) != 2 {
		t.Fatalf("Expected 2 subtitles stored, got %d", len(GlobalSubtitles))
	}
	if GlobalSubtitles[0].Label != "English" {
		t.Errorf("First subtitle label = %q, want %q", GlobalSubtitles[0].Label, "English")
	}
	if GlobalSubtitles[1].Language != "por" {
		t.Errorf("Second subtitle language = %q, want %q", GlobalSubtitles[1].Language, "por")
	}
}

func TestClearGlobalSubtitles(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})

	ClearGlobalSubtitles()

	if GlobalSubtitles != nil {
		t.Errorf("Expected GlobalSubtitles to be nil after clear, got %v", GlobalSubtitles)
	}
}

// ─── GetSubtitleArgs ─────────────────────────────────────────────────────────

func TestGetSubtitleArgs_NoSubs(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// No subtitles set
	args := GetSubtitleArgs()
	if args != nil {
		t.Errorf("Expected nil args with no subtitles, got %v", args)
	}
}

func TestGetSubtitleArgs_GlobalNoSubs(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})
	GlobalNoSubs = true

	args := GetSubtitleArgs()
	if args != nil {
		t.Errorf("Expected nil args when GlobalNoSubs=true, got %v", args)
	}
}

func TestGetSubtitleArgs_SingleSubtitle(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})

	args := GetSubtitleArgs()
	if len(args) != 1 {
		t.Fatalf("Expected 1 arg, got %d", len(args))
	}
	expected := "--sub-file=https://cdn.example.com/en.vtt"
	if args[0] != expected {
		t.Errorf("GetSubtitleArgs()[0] = %q, want %q", args[0], expected)
	}
}

func TestGetSubtitleArgs_MultipleSubtitles(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
		{URL: "https://cdn.example.com/es.vtt", Language: "spa", Label: "Spanish"},
	})

	args := GetSubtitleArgs()
	if len(args) != 1 {
		t.Fatalf("Expected 1 arg (--sub-files=...), got %d", len(args))
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	expected := "--sub-files=" + strings.Join([]string{
		"https://cdn.example.com/en.vtt",
		"https://cdn.example.com/pt.vtt",
		"https://cdn.example.com/es.vtt",
	}, separator)

	if args[0] != expected {
		t.Errorf("GetSubtitleArgs()[0] = %q, want %q", args[0], expected)
	}
}

func TestGetSubtitleArgs_EmptyURLsFiltered(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "", Language: "eng", Label: "English"},
		{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
	})

	args := GetSubtitleArgs()
	if len(args) != 1 {
		t.Fatalf("Expected 1 arg, got %d", len(args))
	}
	expected := "--sub-file=https://cdn.example.com/pt.vtt"
	if args[0] != expected {
		t.Errorf("GetSubtitleArgs()[0] = %q, want %q", args[0], expected)
	}
}

// ─── PromptSubtitleLanguage (non-interactive paths) ──────────────────────────

func TestPromptSubtitleLanguage_NoSubs_ClearsState(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// When GlobalNoSubs is true, PromptSubtitleLanguage must clear subtitles
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})
	GlobalNoSubs = true

	PromptSubtitleLanguage()

	if GlobalSubtitles != nil {
		t.Errorf("Expected GlobalSubtitles to be nil when GlobalNoSubs=true, got %v", GlobalSubtitles)
	}
	// GetSubtitleArgs must return nil too
	if args := GetSubtitleArgs(); args != nil {
		t.Errorf("Expected nil args after GlobalNoSubs prompt, got %v", args)
	}
}

func TestPromptSubtitleLanguage_EmptyTracks_NoError(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// No tracks — should not panic and should leave state empty
	GlobalSubtitles = nil

	PromptSubtitleLanguage() // must not panic

	if GlobalSubtitles != nil {
		t.Errorf("Expected GlobalSubtitles to remain nil, got %v", GlobalSubtitles)
	}
	if args := GetSubtitleArgs(); args != nil {
		t.Errorf("Expected nil args with no tracks, got %v", args)
	}
}

// ─── SelectSubtitles (non-interactive paths) ─────────────────────────────────

func TestSelectSubtitles_NoSubs_Skips(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	GlobalNoSubs = true
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
	})

	SelectSubtitles()

	// Should be unchanged because GlobalNoSubs skips
	if len(GlobalSubtitles) != 2 {
		t.Errorf("Expected subtitles unchanged when GlobalNoSubs=true, got %d", len(GlobalSubtitles))
	}
}

func TestSelectSubtitles_SingleTrack_Skips(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
	})

	SelectSubtitles()

	// Single track — no menu shown, subtitles unchanged
	if len(GlobalSubtitles) != 1 {
		t.Errorf("Expected 1 subtitle (unchanged), got %d", len(GlobalSubtitles))
	}
	if GlobalSubtitles[0].Label != "English" {
		t.Errorf("Expected subtitle label %q, got %q", "English", GlobalSubtitles[0].Label)
	}
}

// ─── Referer management ──────────────────────────────────────────────────────

func TestGlobalReferer(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalReferer("https://rapid-cloud.co/")
	if got := GetGlobalReferer(); got != "https://rapid-cloud.co/" {
		t.Errorf("GetGlobalReferer() = %q, want %q", got, "https://rapid-cloud.co/")
	}

	ClearGlobalReferer()
	if got := GetGlobalReferer(); got != "" {
		t.Errorf("Expected empty referer after clear, got %q", got)
	}
}

// ─── Full Flow: Source → Subtitles → Args ────────────────────────────────────
// These tests simulate the complete subtitle selection pipeline used in the
// actual codebase: api/enhanced.go sets the source and subtitles, then
// player/playvideo.go reads them to build mpv arguments.

func TestFullFlow_9Anime_SubtitlesStoredAndRetrieved(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// Step 1: Simulate what api.GetNineAnimeStreamURL does
	ClearGlobalSubtitles()
	SetGlobalAnimeSource("9Anime")
	SetGlobalReferer("https://rapid-cloud.co/")

	// Simulate subtitle tracks from rapid-cloud response
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/01/01/abc123.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/01/01/def456.vtt", Language: "por", Label: "Portuguese - Brazilian Portuguese"},
		{URL: "https://cc.2cdns.com/01/01/ghi789.vtt", Language: "spa", Label: "Spanish"},
		{URL: "https://cc.2cdns.com/01/01/jkl012.vtt", Language: "jpn", Label: "Japanese"},
	})

	// Step 2: Verify source detection (used in player/playvideo.go)
	if !Is9AnimeSource() {
		t.Fatal("Expected Is9AnimeSource() to be true after SetGlobalAnimeSource(\"9Anime\")")
	}

	// Step 3: Verify all subtitles are available for the prompt
	if len(GlobalSubtitles) != 4 {
		t.Fatalf("Expected 4 subtitle tracks stored, got %d", len(GlobalSubtitles))
	}

	// Step 4: Verify GetSubtitleArgs returns correct mpv args (before user selection)
	args := GetSubtitleArgs()
	if len(args) != 1 {
		t.Fatalf("Expected 1 arg (--sub-files), got %d", len(args))
	}
	if !strings.HasPrefix(args[0], "--sub-files=") {
		t.Errorf("Expected --sub-files= prefix, got %q", args[0])
	}
	// All 4 URLs should be present
	for _, url := range []string{"abc123.vtt", "def456.vtt", "ghi789.vtt", "jkl012.vtt"} {
		if !strings.Contains(args[0], url) {
			t.Errorf("Expected args to contain %q, got %q", url, args[0])
		}
	}
}

func TestFullFlow_9Anime_SingleSubtitleSelection(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// Simulate the 9anime flow
	ClearGlobalSubtitles()
	SetGlobalAnimeSource("9Anime")

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/pt.vtt", Language: "por", Label: "Portuguese"},
		{URL: "https://cc.2cdns.com/es.vtt", Language: "spa", Label: "Spanish"},
	})

	// Simulate user selecting Portuguese (index 1) — this is what
	// PromptSubtitleLanguage does internally when the user picks a single track
	kept := GlobalSubtitles[1]
	GlobalSubtitles = []SubtitleInfo{kept}

	// Verify only Portuguese remains
	if len(GlobalSubtitles) != 1 {
		t.Fatalf("Expected 1 subtitle after selection, got %d", len(GlobalSubtitles))
	}
	if GlobalSubtitles[0].Label != "Portuguese" {
		t.Errorf("Expected label %q, got %q", "Portuguese", GlobalSubtitles[0].Label)
	}

	// Verify mpv args
	args := GetSubtitleArgs()
	if len(args) != 1 {
		t.Fatalf("Expected 1 arg, got %d", len(args))
	}
	if args[0] != "--sub-file=https://cc.2cdns.com/pt.vtt" {
		t.Errorf("Expected --sub-file for Portuguese, got %q", args[0])
	}
}

func TestFullFlow_9Anime_DisableSubtitles(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// Simulate the 9anime flow
	ClearGlobalSubtitles()
	SetGlobalAnimeSource("9Anime")

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/en.vtt", Language: "eng", Label: "English"},
	})

	// Simulate user selecting "No subtitles" — PromptSubtitleLanguage sets nil
	GlobalSubtitles = nil

	// Verify no subtitle args
	args := GetSubtitleArgs()
	if args != nil {
		t.Errorf("Expected nil args after disabling subtitles, got %v", args)
	}
}

func TestFullFlow_9Anime_SubtitlesClearedBetweenEpisodes(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// --- Episode 1 ---
	ClearGlobalSubtitles()
	SetGlobalAnimeSource("9Anime")
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/ep1-en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/ep1-pt.vtt", Language: "por", Label: "Portuguese"},
	})

	if !Is9AnimeSource() {
		t.Fatal("Expected 9Anime source")
	}
	if len(GlobalSubtitles) != 2 {
		t.Fatalf("Episode 1: expected 2 tracks, got %d", len(GlobalSubtitles))
	}

	// Simulate user selects English for ep 1
	GlobalSubtitles = []SubtitleInfo{GlobalSubtitles[0]}
	args1 := GetSubtitleArgs()
	if len(args1) != 1 || !strings.Contains(args1[0], "ep1-en.vtt") {
		t.Errorf("Episode 1: expected English subtitle, got %v", args1)
	}

	// --- Episode 2 ---
	// ClearGlobalSubtitles is called at the start of GetNineAnimeStreamURL
	ClearGlobalSubtitles()
	// New subtitles from the server for episode 2 (might differ)
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/ep2-en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/ep2-pt.vtt", Language: "por", Label: "Portuguese"},
		{URL: "https://cc.2cdns.com/ep2-fr.vtt", Language: "fre", Label: "French"},
	})

	// Episode 2 has a different track list — user MUST be prompted again
	if len(GlobalSubtitles) != 3 {
		t.Fatalf("Episode 2: expected 3 tracks, got %d", len(GlobalSubtitles))
	}

	// Simulate user selects French for ep 2
	GlobalSubtitles = []SubtitleInfo{GlobalSubtitles[2]}
	args2 := GetSubtitleArgs()
	if len(args2) != 1 || !strings.Contains(args2[0], "ep2-fr.vtt") {
		t.Errorf("Episode 2: expected French subtitle, got %v", args2)
	}

	// Verify episode 1 subtitles are NOT leaking into episode 2
	if strings.Contains(args2[0], "ep1") {
		t.Errorf("Episode 2 subtitle args contain episode 1 URL: %v", args2)
	}
}

func TestFullFlow_NonNineAnime_SubtitlePromptNotTriggered(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// For AllAnime, Is9AnimeSource should be false
	SetGlobalAnimeSource("AllAnime")

	if Is9AnimeSource() {
		t.Error("Is9AnimeSource() should be false for AllAnime")
	}

	// For FlixHQ, same thing
	SetGlobalAnimeSource("FlixHQ")

	if Is9AnimeSource() {
		t.Error("Is9AnimeSource() should be false for FlixHQ")
	}
}

// ─── Subtitle Trigger Gating ─────────────────────────────────────────────────
// These tests verify the exact conditions from player/playvideo.go that
// determine whether the subtitle prompt fires.

func TestSubtitlePromptTrigger_9Anime_AlwaysTriggered(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// Simulate different subtitle states: 0 tracks, 1 track, multiple tracks.
	// For 9Anime, the prompt condition in playVideo is:
	//   if is9Anime { util.PromptSubtitleLanguage() }
	// which is unconditional — no check on len(GlobalSubtitles).

	testCases := []struct {
		name   string
		tracks []SubtitleInfo
	}{
		{"no_tracks", nil},
		{"single_track", []SubtitleInfo{
			{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		}},
		{"multiple_tracks", []SubtitleInfo{
			{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
			{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetSubtitleState()
			SetGlobalAnimeSource("9Anime")
			if tc.tracks != nil {
				SetGlobalSubtitles(tc.tracks)
			}

			// The gate condition from playVideo:
			//   is9Anime := util.Is9AnimeSource()
			//   if isMovieOrTV || is9Anime { ... if is9Anime { PromptSubtitleLanguage() } }
			is9Anime := Is9AnimeSource()

			// Verify the trigger is ALWAYS true for 9Anime
			if !is9Anime {
				t.Errorf("9Anime subtitle prompt gate should be true, regardless of track count (%d)",
					len(GlobalSubtitles))
			}
		})
	}
}

func TestSubtitlePromptTrigger_FlixHQ_OnlyMultipleTracks(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalAnimeSource("FlixHQ")

	// For non-9Anime (FlixHQ), the condition is:
	//   } else if len(util.GlobalSubtitles) > 1 { util.SelectSubtitles() }
	// So the prompt only triggers with >1 track.

	testCases := []struct {
		name            string
		tracks          []SubtitleInfo
		shouldSelectRun bool
	}{
		{"no_tracks", nil, false},
		{"single_track", []SubtitleInfo{
			{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		}, false},
		{"multiple_tracks", []SubtitleInfo{
			{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
			{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
		}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetSubtitleState()
			SetGlobalAnimeSource("FlixHQ")
			if tc.tracks != nil {
				SetGlobalSubtitles(tc.tracks)
			}

			// Reproduce the gate condition from playVideo
			is9Anime := Is9AnimeSource()
			wouldCallSelect := !is9Anime && len(GlobalSubtitles) > 1

			if wouldCallSelect != tc.shouldSelectRun {
				t.Errorf("FlixHQ select gate = %v, want %v (tracks: %d)",
					wouldCallSelect, tc.shouldSelectRun, len(GlobalSubtitles))
			}
		})
	}
}

// ─── Subtitle State Applied to Playback Args ────────────────────────────────

func TestSubtitleArgs_AppliedAfterUserSelection(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	// Full realistic flow: source set → subtitles loaded → user picks one → args built
	SetGlobalAnimeSource("9Anime")
	SetGlobalReferer("https://rapid-cloud.co/")

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/pt-br.vtt", Language: "por", Label: "Portuguese - Brazilian Portuguese"},
		{URL: "https://cc.2cdns.com/ar.vtt", Language: "ara", Label: "Arabic"},
	})

	// User selects "Portuguese - Brazilian Portuguese" (index 1)
	kept := GlobalSubtitles[1]
	GlobalSubtitles = []SubtitleInfo{kept}

	// Build mpv args — this is exactly what playVideo does
	subArgs := GetSubtitleArgs()

	// Simulate: mpvArgs = append(mpvArgs, subArgs...)
	mpvArgs := []string{"--cache=yes", "--demuxer-max-bytes=300M"}
	mpvArgs = append(mpvArgs, subArgs...)

	// Verify the subtitle arg is in the final mpv command
	found := slices.Contains(mpvArgs, "--sub-file=https://cc.2cdns.com/pt-br.vtt")
	if !found {
		t.Errorf("Expected mpv args to contain Portuguese subtitle file, got %v", mpvArgs)
	}
}

func TestSubtitleArgs_AllSubtitlesPassedToMpv(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetGlobalAnimeSource("9Anime")

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "https://cc.2cdns.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cc.2cdns.com/pt.vtt", Language: "por", Label: "Portuguese"},
	})

	// User selects "All" — GlobalSubtitles remains unchanged
	subArgs := GetSubtitleArgs()

	if len(subArgs) != 1 {
		t.Fatalf("Expected 1 arg, got %d", len(subArgs))
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	if !strings.HasPrefix(subArgs[0], "--sub-files=") {
		t.Errorf("Expected --sub-files= prefix, got %q", subArgs[0])
	}

	// Both URLs present, separated correctly
	expectedSuffix := "https://cc.2cdns.com/en.vtt" + separator + "https://cc.2cdns.com/pt.vtt"
	if !strings.HasSuffix(subArgs[0], expectedSuffix) {
		t.Errorf("Expected suffix %q, got %q", expectedSuffix, subArgs[0])
	}
}

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

func resetRuntimeState() {
	SetGlobalSource("")
	SetGlobalQuality("")
	SetGlobalMediaType("")
	SetPreferredSubtitleLanguage("")
	SetPreferredAudioLanguage("")
	SetSubtitlesDisabled(false)
	SetGlobalOutputDir("")
	ClearGlobalDownloadRequest()
	ClearGlobalUpscaleRequest()
}

func TestCurrentSessionConfig(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalSource("goyabu")
	SetGlobalQuality("720p")
	SetGlobalMediaType("anime")
	SetPreferredSubtitleLanguage("por")
	SetPreferredAudioLanguage("jpn")
	SetSubtitlesDisabled(true)
	SetGlobalOutputDir("C:\\downloads")

	cfg := CurrentSessionConfig()
	if cfg.Source != "goyabu" || cfg.Quality != "720p" || cfg.MediaType != "anime" {
		t.Fatalf("unexpected config snapshot: %+v", cfg)
	}
	if cfg.SubsLanguage != "por" || cfg.AudioLanguage != "jpn" {
		t.Fatalf("unexpected language snapshot: %+v", cfg)
	}
	if !cfg.NoSubs || cfg.OutputDir != "C:\\downloads" {
		t.Fatalf("unexpected output/no-subs snapshot: %+v", cfg)
	}
}

func TestCurrentDownloadRequestReturnsCopy(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalDownloadRequest(&DownloadRequest{
		AnimeName:     "Naruto",
		SeasonNum:     2,
		EpisodeNum:    7,
		AllAnimeSmart: true,
	})

	req := CurrentDownloadRequest()
	if req == nil {
		t.Fatal("CurrentDownloadRequest() returned nil")
	}
	req.AnimeName = "Mutated"

	current := CurrentDownloadRequest()
	if current.AnimeName != "Naruto" {
		t.Fatalf("stored request was mutated through snapshot: %+v", current)
	}
}

func TestCurrentUpscaleRequestReturnsCopy(t *testing.T) {
	t.Cleanup(resetRuntimeState)

	SetGlobalUpscaleRequest(&UpscaleRequest{
		InputPath:   "input.mp4",
		OutputPath:  "output.mp4",
		ScaleFactor: 2,
	})

	req := CurrentUpscaleRequest()
	if req == nil {
		t.Fatal("CurrentUpscaleRequest() returned nil")
	}
	req.OutputPath = "mutated.mp4"

	current := CurrentUpscaleRequest()
	if current.OutputPath != "output.mp4" {
		t.Fatalf("stored upscale request was mutated through snapshot: %+v", current)
	}
}
