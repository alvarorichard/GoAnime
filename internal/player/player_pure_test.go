package player

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func TestSanitizeMediaTarget(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
		want    string
	}{
		{"empty", "", true, ""},
		{"newline middle", "http://x\ny", true, ""},
		{"null byte", "http://x\x00y", true, ""},
		{"dash prefix", "-attack", true, ""},
		{"https url", "https://example.com/v.mp4", false, "https://example.com/v.mp4"},
		{"http url", "http://example.com/v.mp4", false, "http://example.com/v.mp4"},
		{"file scheme rejected", "file:///etc/passwd", true, ""},
		{"ftp rejected", "ftp://x.com", true, ""},
		{"plain path cleaned", "/tmp/video.mp4", false, filepath.FromSlash("/tmp/video.mp4")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeMediaTarget(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"empty", "", true},
		{"null byte", "file\x00", true},
		{"newline", "file\n.mp4", true},
		{"dash prefix", "-output.mp4", true},
		{"home prefix ok", filepath.Join(homeOrEmpty(), "ok.mp4"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizeOutputPath(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func homeOrEmpty() string {
	d, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return d
}

// SetMediaType toggles the isMovieOrTV flag (see snapshot). IsCurrentMediaMovie
// reads the separate `mediaType == "movie"` field set via SetExactMediaType.
func TestSetMediaType_TogglesFlag(t *testing.T) {
	prev := snapshotMedia().IsMovieOrTV
	t.Cleanup(func() { SetMediaType(prev) })

	SetMediaType(true)
	assert.True(t, snapshotMedia().IsMovieOrTV)

	SetMediaType(false)
	assert.False(t, snapshotMedia().IsMovieOrTV)
}

func TestIsCurrentMediaMovie_DependsOnMediaType(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("movie")
	assert.True(t, IsCurrentMediaMovie())
	SetExactMediaType("tv")
	assert.False(t, IsCurrentMediaMovie())
	SetExactMediaType("anime")
	assert.False(t, IsCurrentMediaMovie())
}

func TestSetExactMediaType_RoundTrip(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("movie")
	assert.Equal(t, "movie", GetExactMediaType())

	SetExactMediaType("tv")
	assert.Equal(t, "tv", GetExactMediaType())

	SetExactMediaType("anime")
	assert.Equal(t, "anime", GetExactMediaType())
}

func TestSetSeasonMap_RoundTrip(t *testing.T) {
	sm := []metadata.SeasonMapping{{Season: 1, StartEp: 1, EndEp: 12, EpisodeCount: 12}}
	SetSeasonMap(sm)
	snap := snapshotMedia()
	assert.Equal(t, 1, snap.SeasonMap[0].Season)
}

func TestSetMediaMeta_RoundTrip(t *testing.T) {
	meta := &util.MediaMeta{IMDBID: "tt1234567", Year: "2020"}
	SetMediaMeta(meta)
	got := GetMediaMeta()
	assert.NotNil(t, got)
	assert.Equal(t, "tt1234567", got.IMDBID)
	assert.Equal(t, "2020", got.Year)
}

func TestPreWarmMPVPath_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { PreWarmMPVPath() })
}

func TestSnapshotMedia_ReturnsCopy(t *testing.T) {
	SetAnimeName("X", 3)
	snap := snapshotMedia()
	assert.Equal(t, "X", snap.AnimeName)
	assert.Equal(t, 3, snap.AnimeSeason)
}

func TestPrintDownloadLocation_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { printDownloadLocation("/tmp/some/file.mp4") })
}

func TestSetLastAnimeURL_RoundTrip(t *testing.T) {
	prev := getLastAnimeURL()
	t.Cleanup(func() { setLastAnimeURL(prev) })

	setLastAnimeURL("https://example.com/anime/x")
	assert.Equal(t, "https://example.com/anime/x", getLastAnimeURL())

	setLastAnimeURL("")
	assert.Equal(t, "", getLastAnimeURL())
}

func TestGetExactMediaType_DefaultEmpty(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("")
	assert.Equal(t, "", GetExactMediaType())
}

func TestGetMediaMeta_AfterClearReturnsNil(t *testing.T) {
	prev := GetMediaMeta()
	t.Cleanup(func() { SetMediaMeta(prev) })

	SetMediaMeta(nil)
	assert.Nil(t, GetMediaMeta())
}

// Mutates global util.GlobalSubtitles — keep serial (no parallel).
func TestDownloadSubtitleFiles_NoSubsIsNoop(t *testing.T) {
	prev := util.GlobalSubtitles
	util.ClearGlobalSubtitles()
	t.Cleanup(func() { util.GlobalSubtitles = prev })

	assert.NotPanics(t, func() { downloadSubtitleFiles("/tmp/x.mp4", nil) })
}

func TestStartVideo_InvalidLinkReturnsError(t *testing.T) {
	t.Parallel()
	if _, err := StartVideo("http://bad\nurl", nil); err == nil {
		t.Fatal("expected error from sanitize or missing-mpv path")
	}
}

// fuzzyfinder/tcell terminfo is package-level state — keep TUI-touching tests serial.
func TestHandleUpscaleFromMenu_DoesNotPanic(t *testing.T) {
	// handleUpscaleFromMenu opens an interactive fuzzy finder. Outside a TTY
	// it errors on Linux but blocks indefinitely on Windows (tcell winTty
	// getConsoleInput syscall), which deadlocks CI. Skip when no TTY.
	if os.Getenv("CI") != "" || !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("Skipping interactive fuzzy-finder test: requires a real TTY")
	}
	assert.NotPanics(t, func() { _ = handleUpscaleFromMenu() })
}

func TestAskForDownload_ReturnsValidMarker(t *testing.T) {
	// askForDownload opens an interactive fuzzy finder. Outside a TTY it
	// errors on Linux but blocks indefinitely on Windows (tcell winTty
	// getConsoleInput syscall), which deadlocks CI. Skip when no TTY.
	if os.Getenv("CI") != "" || !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("Skipping interactive fuzzy-finder test: requires a real TTY")
	}
	got := askForDownload()
	assert.GreaterOrEqual(t, got, 1)
}

func TestAskForPlayOffline_DoesNotPanic(t *testing.T) {
	// askForPlayOffline opens an interactive fuzzy finder. Outside a TTY it
	// errors on Linux but blocks indefinitely on Windows (tcell winTty
	// getConsoleInput syscall), which deadlocks CI. Skip when no TTY.
	if os.Getenv("CI") != "" || !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("Skipping interactive fuzzy-finder test: requires a real TTY")
	}
	assert.NotPanics(t, func() { _ = askForPlayOffline() })
}

// HandleDownloadAndPlay loops on askForDownload (huh.NewSelect). Without a
// TTY, askForDownload errors and returns its sentinel "4" code which routes
// to the play (default) branch. Each test below pins a different sub-branch
// of the play path. All run serial — fuzzyfinder/tcell terminfo is package-
// level and races with parallel tests.
func TestHandleDownloadAndPlay_EmptyURLReturnsNoValidVideoURL(t *testing.T) {
	SetAnimeName("HDP_EmptyURLTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	anime := &models.Anime{URL: "https://example.com/x", Source: "AllAnime"}
	err := HandleDownloadAndPlay(
		"", nil, 1, "https://example.com/x", "1", 0, 0, nil, "HDP_EmptyURLTest", 1, anime,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid video URL found")
}

func TestHandleDownloadAndPlay_AnimeFireURLHitsExtractionBranch(t *testing.T) {
	SetAnimeName("HDP_AnimeFireTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	anime := &models.Anime{URL: "https://example.com/x", Source: "AllAnime"}
	err := HandleDownloadAndPlay(
		"https://animefire.io/video/x", // forces the extraction branch
		nil, 1, "https://example.com/x", "1", 0, 0, nil, "HDP_AnimeFireTest", 1, anime,
	)
	// The extraction goes through SafeGet; loopback rejection leaves the
	// resolved URL empty and exercises the final validation branch.
	require.Error(t, err)
}

// (HLS / plain-HTTP play branches: HandleDownloadAndPlay loops on
// askForDownload → on TUI failure returns the play sentinel → playVideo →
// catches ErrBackToDownloadOptions → continues the loop → infinite. Cannot
// drive non-error play branches without a TTY. The dispatched routines
// (playVideo, extractActualVideoURL, ExtractVideoSourcesWithPrompt) are
// covered by sibling tests.)

func TestDownloadAndPlayEpisode_EmptyURLReturnsError(t *testing.T) {
	t.Parallel()
	err := downloadAndPlayEpisode("", nil, 1, "https://x", "1", 0, 0, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty video URL")
}

func TestDownloadAndPlayEpisode_AnimeFireResolveFails(t *testing.T) {
	// animefire.io/video/ → extractActualVideoURL calls api.SafeGet which
	// rejects loopback (no real-network), so the resolve step fails fast.
	SetAnimeName("DAPE_AnimeFireResolveTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	err := downloadAndPlayEpisode(
		"https://animefire.io/video/loopback-blocked",
		nil, 1, "https://x", "1", 0, 0, nil,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve AnimeFire video URL")
}

func TestDownloadAndPlayEpisode_BloggerProxyURLBranch(t *testing.T) {
	// isBloggerProxyURL → tries to use GetBloggerVideoURL which is empty
	// (no proxy started) → error path.
	SetAnimeName("DAPE_BloggerProxyTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	StopBloggerProxy() // ensure empty
	err := downloadAndPlayEpisode(
		"http://127.0.0.1:8080/blogger_proxy/x",
		nil, 1, "https://x", "1", 0, 0, nil,
	)
	require.Error(t, err)
}

// (HLS branch coverage for downloadAndPlayEpisode is exercised indirectly
// via TestDownloadWithNativeHLS_* — calling downloadAndPlayEpisode here
// spawned a download goroutine that outlived the test and raced with other
// tests on util.GlobalReferer.)

// (TestDownloadAndPlayEpisode_ExistingFileSkipsDownloadButFailsAtPlay was
// removed: with a host-side mpv present the test launched a real window and
// without mpv the spawned DownloadVideo goroutine outlived the test and
// raced with subsequent tests that mutate util.GlobalReferer. The exercised
// branches (file-exists check, playVideo dispatch) are covered separately
// by TestFileExists and TestPlayVideo_*.)
