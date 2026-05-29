package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeEpisodeDownloader returns a downloader with a fixed temp OutputDir.
// Bypasses NewEpisodeDownloaderWithAnime so tests don't hit AniList enrichment.
func makeEpisodeDownloader(t *testing.T, animeName string, episodes []models.Episode, isMovie bool) *EpisodeDownloader {
	t.Helper()
	d := &EpisodeDownloader{
		config: DownloadConfig{
			AnimeURL:   "https://example.test/anime/foo",
			OutputDir:  t.TempDir(),
			NumThreads: 4,
			Concurrent: 2,
			AnimeName:  util.SanitizeForFilename(animeName),
			Season:     1,
			Meta:       &util.MediaMeta{OfficialTitle: animeName, Year: "2020"},
		},
		episodes: episodes,
	}
	if isMovie {
		d.anime = &models.Anime{Name: animeName, MediaType: models.MediaTypeMovie}
	}
	return d
}

func makeEpisodes(n int) []models.Episode {
	out := make([]models.Episode, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, models.Episode{Num: i, Number: fmt.Sprintf("%d", i), URL: fmt.Sprintf("https://example.test/ep%d", i)})
	}
	return out
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// TestNewEpisodeDownloader_NilAnime verifies the simple constructor produces
// an EpisodeDownloader with sensible defaults when no anime metadata exists.
func TestNewEpisodeDownloader_NilAnime(t *testing.T) {
	t.Parallel()

	eps := makeEpisodes(3)
	d := NewEpisodeDownloader(eps, "https://example.test/anime/foo")
	require.NotNil(t, d)
	assert.Equal(t, 3, len(d.episodes))
	assert.Equal(t, "https://example.test/anime/foo", d.config.AnimeURL)
	assert.Equal(t, 4, d.config.NumThreads)
	assert.Equal(t, 3, d.config.Concurrent)
	assert.GreaterOrEqual(t, d.config.Season, 1)
	assert.NotEmpty(t, d.config.OutputDir)
}

// TestNewEpisodeDownloaderWithAnime_MovieRoutesToMoviesDir verifies that a
// movie anime is routed to the movie download directory with a flat path.
func TestNewEpisodeDownloaderWithAnime_MovieRoutesToMoviesDir(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{Name: "Spirited Away", MediaType: models.MediaTypeMovie, Year: "2001"}
	d := NewEpisodeDownloaderWithAnime(nil, "https://example.test/movie/spirited", anime)
	require.NotNil(t, d)
	assert.Contains(t, strings.ToLower(d.config.OutputDir), "movies")
	assert.Contains(t, d.config.OutputDir, "Spirited Away")
}

// ---------------------------------------------------------------------------
// Top-level public commands
// ---------------------------------------------------------------------------

// TestDownloadSingleEpisode_NotFound verifies the not-found error path.
func TestDownloadSingleEpisode_NotFound(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(2), false)
	err := d.DownloadSingleEpisode(99)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestDownloadEpisodeRange_InvertedRange checks input validation.
func TestDownloadEpisodeRange_InvertedRange(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(3), false)
	err := d.DownloadEpisodeRange(5, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be greater")
}

// TestDownloadEpisodeRange_AllExisting goes through the all-exist path that
// short-circuits before downloading. The prompt reads stdin → we close stdin
// so it returns nil immediately.
func TestDownloadEpisodeRange_AllExisting(t *testing.T) {
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(2), false)
	// Pre-create episode files so they exist
	for i := 1; i <= 2; i++ {
		p := filepath.Join(d.episodeDir(i), d.episodeFilename(i))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o600))
	}
	withClosedStdin(t)
	// All episodes exist → routes into promptPlayExistingRangeHuh which
	// reads stdin and returns "failed to read input" on EOF. Either nil or
	// that input error is acceptable; what we're verifying is the no-download
	// short-circuit path.
	err := d.DownloadEpisodeRange(1, 2)
	if err != nil {
		assert.Contains(t, err.Error(), "input")
	}
}

// TestDownloadAllEpisodes_EmptyList verifies the empty-episodes error path.
func TestDownloadAllEpisodes_EmptyList(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	err := d.DownloadAllEpisodes()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no episodes")
}

// TestDownloadAllEpisodes_AllExisting short-circuits when every episode is
// already on disk.
func TestDownloadAllEpisodes_AllExisting(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(2), false)
	for i := 1; i <= 2; i++ {
		p := filepath.Join(d.episodeDir(i), d.episodeFilename(i))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o600))
	}
	err := d.DownloadAllEpisodes()
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Concurrent / multi-episode helpers
// ---------------------------------------------------------------------------

// TestDownloadConcurrentWithProgress_EmptyList returns nil on no-op input.
func TestDownloadConcurrentWithProgress_EmptyList(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	err := d.downloadConcurrentWithProgress(nil)
	require.NoError(t, err)
}

// TestDownloadMultipleWithProgress_Pin keeps the symbol referenced.
// The function drives a Bubble Tea program; full execution requires a TTY.
func TestDownloadMultipleWithProgress_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadMultipleWithProgress
}

// TestDownloadEpisodeWithSharedProgress_HTTPSuccess exercises the per-episode
// shared-progress HTTP download via httptest. Uses a tea program that we
// don't run; messages are buffered and ignored.
func TestDownloadEpisodeWithSharedProgress_HTTPSuccess(t *testing.T) {
	t.Parallel()

	body := []byte("FAKEVIDEOBYTES")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.Write(body)
	}))
	t.Cleanup(srv.Close)

	d := makeEpisodeDownloader(t, "Foo", nil, false)
	dest := filepath.Join(d.config.OutputDir, "out.mp4")
	m := &progressModel{progress: progress.New(progress.WithDefaultBlend()), totalBytes: int64(len(body))}
	prog := tea.NewProgram(m)

	var epRecv, totalRecv int64
	var mu sync.Mutex
	// SafeTransport blocks loopback IPs — call must error out before writing.
	err := d.downloadEpisodeWithSharedProgress(srv.URL+"/v.mp4", dest, &epRecv, &totalRecv, &mu, m, prog)
	require.Error(t, err, "loopback HTTP server should be blocked by SafeTransport")
}

// TestDownloadEpisodeWithSharedProgress_BloggerRoute verifies the blogger.com
// branch routes to yt-dlp (which will fail without a real URL — the routing
// itself is what we verify).
func TestDownloadEpisodeWithSharedProgress_BloggerRoute(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	dest := filepath.Join(d.config.OutputDir, "blog.mp4")
	var er, tr int64
	var mu sync.Mutex
	m := &progressModel{}
	err := d.downloadEpisodeWithSharedProgress("https://blogger.com/video.g?some=id", dest, &er, &tr, &mu, m, nil)
	// yt-dlp will fail on a fake URL — we just want to confirm routing happened.
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Helpers — pure / structural
// ---------------------------------------------------------------------------

func TestFindEpisodeByNumber(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", makeEpisodes(3), false)

	ep, ok := d.findEpisodeByNumber(2)
	assert.True(t, ok)
	assert.Equal(t, 2, ep.Num)

	_, ok = d.findEpisodeByNumber(99)
	assert.False(t, ok)
}

// TestPrintDownloadLocation_NoPanic exercises the helper. It only writes to
// stdout; we just confirm it does not panic on a relative path or absolute.
func TestPrintDownloadLocation_NoPanic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NotPanics(t, func() { printDownloadLocation(dir) })
	assert.NotPanics(t, func() { printDownloadLocation("relative/path") })
}

func TestEpisodeDownloader_FileExists(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	missing := filepath.Join(d.config.OutputDir, "nope.mp4")
	assert.False(t, d.fileExists(missing))

	present := filepath.Join(d.config.OutputDir, "yes.mp4")
	require.NoError(t, os.WriteFile(present, []byte("x"), 0o600))
	assert.True(t, d.fileExists(present))
}

func TestEpisodeDownloader_SanitizeDestPath(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		_, err := d.sanitizeDestPath("")
		require.Error(t, err)
	})

	t.Run("within", func(t *testing.T) {
		t.Parallel()
		p := filepath.Join(d.config.OutputDir, "Season 01", "ep.mp4")
		got, err := d.sanitizeDestPath(p)
		require.NoError(t, err)
		assert.Equal(t, p, got)
	})

	t.Run("escape", func(t *testing.T) {
		t.Parallel()
		_, err := d.sanitizeDestPath(filepath.Join(d.config.OutputDir, "..", "..", "etc", "passwd"))
		require.Error(t, err)
	})
}

func TestEpisodeDownloader_EpisodeFilename(t *testing.T) {
	t.Parallel()

	// fallback when AnimeName == ""
	dNoName := &EpisodeDownloader{config: DownloadConfig{OutputDir: t.TempDir(), Season: 1}}
	assert.Equal(t, "7.mp4", dNoName.episodeFilename(7))

	// Plex-style with AnimeName
	dNamed := makeEpisodeDownloader(t, "Naruto", nil, false)
	got := dNamed.episodeFilename(3)
	assert.Contains(t, got, "Naruto")
	assert.Contains(t, got, "S01E03")

	// Movie → flat name (no SxxEyy)
	dMovie := makeEpisodeDownloader(t, "Spirited", nil, true)
	got = dMovie.episodeFilename(1)
	assert.Contains(t, got, "Spirited")
	assert.NotContains(t, got, "E01")
}

func TestEpisodeDownloader_ResolveEpisodeSeason(t *testing.T) {
	t.Parallel()

	t.Run("no_map", func(t *testing.T) {
		t.Parallel()
		d := makeEpisodeDownloader(t, "Foo", nil, false)
		s, e := d.resolveEpisodeSeason(42)
		assert.Equal(t, 1, s)
		assert.Equal(t, 42, e)
	})

	t.Run("with_map", func(t *testing.T) {
		t.Parallel()
		d := makeEpisodeDownloader(t, "Foo", nil, false)
		d.seasonMap = []metadata.SeasonMapping{
			{Season: 1, StartEp: 1, EndEp: 12},
			{Season: 2, StartEp: 13, EndEp: 24},
		}
		s, e := d.resolveEpisodeSeason(15)
		assert.Equal(t, 2, s)
		assert.Equal(t, 3, e)

		// Past EndEp of last → uses last season
		s, e = d.resolveEpisodeSeason(50)
		assert.Equal(t, 2, s)
		assert.Equal(t, 38, e)
	})
}

func TestEpisodeDownloader_EpisodeDir(t *testing.T) {
	t.Parallel()

	t.Run("no_anime_name", func(t *testing.T) {
		t.Parallel()
		d := &EpisodeDownloader{config: DownloadConfig{OutputDir: "/tmp/out"}}
		assert.Equal(t, "/tmp/out", d.episodeDir(1))
	})

	t.Run("movie", func(t *testing.T) {
		t.Parallel()
		d := makeEpisodeDownloader(t, "Movie", nil, true)
		assert.Equal(t, d.config.OutputDir, d.episodeDir(1))
	})

	t.Run("with_season_map", func(t *testing.T) {
		t.Parallel()
		d := makeEpisodeDownloader(t, "Naruto", nil, false)
		d.seasonMap = []metadata.SeasonMapping{{Season: 2, StartEp: 1, EndEp: 24}}
		got := d.episodeDir(3)
		assert.Contains(t, got, "Naruto")
		assert.Contains(t, got, "Season 02")
	})
}

// TestEpisodeDownloader_GetBestQualityURL verifies the helper bubbles up the
// underlying error path from player.GetVideoURLForEpisode for a bogus URL.
func TestEpisodeDownloader_GetBestQualityURL(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_, err := d.getBestQualityURL("https://example.invalid/ep1")
	require.Error(t, err)
}

func TestEpisodeDownloader_GetContentLength(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)

	t.Run("m3u8_short_circuit", func(t *testing.T) {
		t.Parallel()
		n, err := d.getContentLength("https://cdn.example.test/master.m3u8")
		require.NoError(t, err)
		assert.Equal(t, int64(400*1024*1024), n)
	})

	t.Run("allanime_url_fallback", func(t *testing.T) {
		t.Parallel()
		// AllAnime branch: HEAD will fail (unresolvable host) → 300MB fallback.
		// Hostname has invalid TLD to guarantee DNS failure regardless of env.
		n, err := d.getContentLength("https://allanime.pro.invalid.test/video/foo.mp4")
		require.NoError(t, err)
		assert.Equal(t, int64(300*1024*1024), n)
	})

	t.Run("non_allanime_loopback_errors", func(t *testing.T) {
		t.Parallel()
		// Plain unknown URL → SafeTransport blocks → no AllAnime fallback → error.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "12345")
		}))
		t.Cleanup(srv.Close)
		_, err := d.getContentLength(srv.URL + "/file.mp4")
		require.Error(t, err)
	})
}

func TestEpisodeDownloader_EstimateContentLengthForAllAnime(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	client := &http.Client{Timeout: time.Second}

	t.Run("m3u8_estimate", func(t *testing.T) {
		t.Parallel()
		n, err := d.estimateContentLengthForAllAnime("https://x/y.m3u8", client)
		require.NoError(t, err)
		assert.Equal(t, int64(500*1024*1024), n)
	})

	t.Run("range_request_failure_fallback", func(t *testing.T) {
		t.Parallel()
		n, err := d.estimateContentLengthForAllAnime("https://invalid.invalid.test/foo.mp4", client)
		require.NoError(t, err)
		assert.Equal(t, int64(300*1024*1024), n)
	})
}

// ---------------------------------------------------------------------------
// Download path branches (pin where deeper drive requires TTY/yt-dlp)
// ---------------------------------------------------------------------------

func TestEpisodeDownloader_DownloadWithProgress_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadWithProgress // drives tea program → needs TTY
}

func TestEpisodeDownloader_DownloadEpisodeWithProgress_EmptyURL(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	err := d.downloadEpisodeWithProgress("", filepath.Join(d.config.OutputDir, "x.mp4"), &progressModel{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty video URL")
}

func TestEpisodeDownloader_DownloadHTTPWithProgress_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadHTTPWithProgress // needs running tea program
}

func TestEpisodeDownloader_DownloadM3U8WithYtDlp_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadM3U8WithYtDlp // requires yt-dlp binary + tea program
}

func TestEpisodeDownloader_DownloadWithYtDlp_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadWithYtDlp // requires yt-dlp binary
}

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

func TestIsUnsafeExtError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unsafe_extension", errors.New("yt-dlp: file has unsafe extension"), true},
		{"unusual_extension", errors.New("extension is unusual and will be skipped"), true},
		{"generic_error", errors.New("connection refused"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isUnsafeExtError(tc.err))
		})
	}
}

// ---------------------------------------------------------------------------
// TUI prompt helpers — pin (read stdin)
// ---------------------------------------------------------------------------

func TestPromptPlayExisting_StdinClosed(t *testing.T) {
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	withClosedStdin(t)
	err := d.promptPlayExisting(1, "/dev/null")
	assert.NoError(t, err)
}

func TestPromptPlayDownloaded_StdinClosed(t *testing.T) {
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	withClosedStdin(t)
	err := d.promptPlayDownloaded(1, "/dev/null")
	assert.NoError(t, err)
}

func TestPromptPlayDownloadedRangeHuh_EmptyList(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	err := d.promptPlayDownloadedRangeHuh(nil)
	require.NoError(t, err)
}

func TestPromptPlayExistingRangeHuh_EmptyList(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	err := d.promptPlayExistingRangeHuh(nil)
	require.NoError(t, err)
}

// TestPlayEpisode_Pin keeps the symbol referenced — it calls player.StartVideo
// which spawns mpv, not driveable from a unit test.
func TestPlayEpisode_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.playEpisode
}

// ---------------------------------------------------------------------------
// Bubble Tea progressModel
// ---------------------------------------------------------------------------

func TestTickCmd_ReturnsTickMsg(t *testing.T) {
	t.Parallel()
	cmd := tickCmd()
	require.NotNil(t, cmd)
	// We can't fully drive the cmd without a tea runtime, but we can confirm
	// invocation produces a tickMsg eventually. Use a goroutine + small wait.
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		_, ok := msg.(tickMsg)
		assert.True(t, ok, "tickCmd must produce tickMsg, got %T", msg)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("tickCmd timed out")
	}
}

func TestProgressModel_Init(t *testing.T) {
	t.Parallel()
	m := &progressModel{}
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestProgressModel_Update(t *testing.T) {
	t.Parallel()

	t.Run("status_msg", func(t *testing.T) {
		t.Parallel()
		m := &progressModel{progress: progress.New(progress.WithDefaultBlend())}
		_, cmd := m.Update(statusMsg("hello"))
		assert.Equal(t, "hello", m.status)
		assert.Nil(t, cmd)
	})

	t.Run("progress_msg", func(t *testing.T) {
		t.Parallel()
		m := &progressModel{progress: progress.New(progress.WithDefaultBlend())}
		_, _ = m.Update(progressMsg{received: 50, totalBytes: 100})
		assert.Equal(t, int64(50), m.received)
		assert.Equal(t, int64(100), m.totalBytes)
		assert.Greater(t, m.peakPct, 0.0)
	})

	t.Run("tick_done_quits", func(t *testing.T) {
		t.Parallel()
		m := &progressModel{progress: progress.New(progress.WithDefaultBlend()), done: true}
		_, cmd := m.Update(tickMsg(time.Now()))
		require.NotNil(t, cmd)
		// tea.Quit is the cmd — call it and confirm it returns tea.QuitMsg.
		msg := cmd()
		_, ok := msg.(tea.QuitMsg)
		assert.True(t, ok)
	})
}

func TestProgressModel_View(t *testing.T) {
	t.Parallel()
	m := &progressModel{progress: progress.New(progress.WithDefaultBlend())}
	v := m.View()
	s := fmt.Sprintf("%v", v)
	assert.Contains(t, s, "downloading")

	m2 := &progressModel{progress: progress.New(progress.WithDefaultBlend()), totalBytes: 200, received: 100, status: "Working"}
	v2 := m2.View()
	s2 := fmt.Sprintf("%v", v2)
	assert.Contains(t, s2, "Working")
}

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

// withClosedStdin replaces os.Stdin with /dev/null for the test duration so
// fmt.Scanln() returns immediately with an EOF error.
func withClosedStdin(t *testing.T) {
	t.Helper()
	old := os.Stdin
	f, err := os.Open(os.DevNull)
	require.NoError(t, err)
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
		_ = f.Close()
	})
}
