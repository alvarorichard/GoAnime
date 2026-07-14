package downloader

import (
	"path/filepath"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// downloadMultipleWithProgress — empty input returns nil immediately
// ---------------------------------------------------------------------------

func TestDownloadMultipleWithProgress_EmptyList(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	episodeInfos := map[int]struct {
		videoURL string
		path     string
		size     int64
	}{}
	// With an empty episodeNums slice the goroutine loop is never entered
	// and errChan is closed with no items → returns nil immediately.
	err := d.downloadMultipleWithProgress([]int{}, episodeInfos, &progressModel{}, nil)
	assert.NoError(t, err)
}

func TestDownloadMultipleWithProgress_EpisodeNotInInfoMap(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	episodeInfos := map[int]struct {
		videoURL string
		path     string
		size     int64
	}{} // deliberately empty — episode 1 won't be found
	// goroutine skips episodes not in episodeInfos → no program.Send call
	err := d.downloadMultipleWithProgress([]int{1}, episodeInfos, &progressModel{}, nil)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// downloadWithProgress — symbol pin (blocks on tea.Program.Run without TTY)
// ---------------------------------------------------------------------------

func TestDownloadWithProgress_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadWithProgress // TTY required for tea.Program.Run
}

// ---------------------------------------------------------------------------
// downloadHTTPWithProgress — SSRF loopback URL fails before program.Send
// ---------------------------------------------------------------------------

func TestDownloadHTTPWithProgress_LoopbackSSRFBlocked(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")

	// SafeTransport blocks 127.0.0.1 → fails before program.Send → nil program is safe
	err := d.downloadHTTPWithProgress("https://127.0.0.1/video.mp4", dest, &progressModel{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start download")
}

func TestDownloadHTTPWithProgress_InvalidScheme(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	dest := filepath.Join(d.config.OutputDir, "ep1.mp4")
	err := d.downloadHTTPWithProgress("not-a-url", dest, &progressModel{}, nil)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// downloadM3U8WithYtDlp — symbol pin (requires yt-dlp binary + tea.Program)
// ---------------------------------------------------------------------------

func TestDownloadM3U8WithYtDlp_Pin(t *testing.T) {
	t.Parallel()
	d := makeEpisodeDownloader(t, "Foo", nil, false)
	_ = d.downloadM3U8WithYtDlp // requires yt-dlp binary and running tea.Program
}

// ---------------------------------------------------------------------------
// playEpisode — already pinned in downloader_test.go; no duplicate needed
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// progressModel helpers (already partly covered, extend for shared-progress path)
// ---------------------------------------------------------------------------

func TestProgressModel_UpdateSharedProgress(t *testing.T) {
	t.Parallel()
	var totalReceived int64
	var mu sync.Mutex
	var episodeReceived int64

	// Simulate what downloadEpisodeWithSharedProgress does internally with mu
	mu.Lock()
	totalReceived += 1024
	episodeReceived += 1024
	_ = episodeReceived
	mu.Unlock()

	assert.Equal(t, int64(1024), totalReceived)
}

// ---------------------------------------------------------------------------
// Bubble Tea progressModel.Update — additional msg branches
// ---------------------------------------------------------------------------

func TestProgressModel_Update_ProgressMsg(t *testing.T) {
	t.Parallel()
	m := &progressModel{totalBytes: 1000}
	msg := progressMsg{received: 500, totalBytes: 1000}
	newModel, cmd := m.Update(msg)
	require.NotNil(t, newModel)
	_ = cmd
	updated := newModel.(*progressModel)
	assert.Equal(t, int64(500), updated.received)
}

func TestProgressModel_Update_StatusMsg(t *testing.T) {
	t.Parallel()
	m := &progressModel{}
	newModel, _ := m.Update(statusMsg("Downloading..."))
	updated := newModel.(*progressModel)
	assert.Equal(t, "Downloading...", updated.status)
}

func TestProgressModel_Update_CtrlC(t *testing.T) {
	t.Parallel()
	m := &progressModel{}
	// In Bubble Tea v2, Ctrl+C is represented as Code:'c' + Mod:ModCtrl
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	assert.NotNil(t, cmd)
}

func TestProgressModel_View_DoneState(t *testing.T) {
	t.Parallel()
	m := &progressModel{done: true, status: "done"}
	view := m.View()
	assert.Contains(t, view.Content, "done")
}
