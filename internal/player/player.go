package player

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/downloader/hls"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/upscaler"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
	"github.com/pkg/errors"
	"golang.org/x/term"
)

// Cached mpv path — avoids repeated filesystem searches on every episode play
var (
	cachedMPVPath     string
	cachedMPVPathErr  error
	cachedMPVPathOnce sync.Once
)

// PreWarmMPVPath looks up the mpv binary path in the background.
// Call this early at startup so StartVideo doesn't block on the filesystem search.
func PreWarmMPVPath() {
	go cachedMPVPathOnce.Do(func() {
		cachedMPVPath, cachedMPVPathErr = findMPVPath()
		if cachedMPVPathErr == nil {
			util.Debugf("Pre-warmed mpv path: %s", cachedMPVPath)
		}
	})
}

// mediaState groups mutable per-session media metadata behind a lock so that
// concurrent goroutines (batch downloads, etc.) can safely read while the main
// flow writes.
type mediaState struct {
	mu          sync.RWMutex
	animeURL    string
	animeName   string
	animeSeason int
	isMovieOrTV bool
	mediaType   string                   // "movie", "tv", or "anime"
	seasonMap   []metadata.SeasonMapping // AniList-based absolute→season map
	meta        *util.MediaMeta          // External IDs and year for folder naming
}

var gMedia mediaState

// SetAnimeName sets the anime name and season for Plex-compatible download file naming.
// Call this before any download operations to ensure proper naming.
func SetAnimeName(name string, season int) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.animeName = name
	gMedia.animeSeason = max(season, 1)
}

// GetAnimeSeason returns the season downloads are currently filed under.
func GetAnimeSeason() int {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	return gMedia.animeSeason
}

// SetMediaType marks whether the current content is a movie/TV show (true) or anime (false).
// This determines whether downloads go to the movies or anime directory.
func SetMediaType(isMovieOrTV bool) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.isMovieOrTV = isMovieOrTV
}

// SetExactMediaType stores the exact media type ("movie", "tv", "anime") for
// intelligent download path organization. Movies get flat paths, TV shows and
// anime get season/episode structures.
func SetExactMediaType(mediaType string) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.mediaType = mediaType
	gMedia.isMovieOrTV = (mediaType == "movie" || mediaType == "tv")
}

// GetExactMediaType returns the current exact media type.
func GetExactMediaType() string {
	gMedia.mu.RLock()
	defer gMedia.mu.RUnlock()
	return gMedia.mediaType
}

// IsCurrentMediaMovie returns true if the current content is a standalone movie.
func IsCurrentMediaMovie() bool {
	gMedia.mu.RLock()
	defer gMedia.mu.RUnlock()
	return gMedia.mediaType == "movie"
}

// setLastAnimeURL stores the most recent anime URL for navigation support.
func setLastAnimeURL(u string) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.animeURL = u
}

// getLastAnimeURL returns the stored anime URL.
func getLastAnimeURL() string {
	gMedia.mu.RLock()
	defer gMedia.mu.RUnlock()
	return gMedia.animeURL
}

// mediaSnapshot is a read-only copy of all media metadata fields, obtained
// atomically under a single RLock so batch-download goroutines always see a
// consistent view of the state.
type mediaSnapshot struct {
	AnimeName   string
	AnimeSeason int
	IsMovieOrTV bool
	MediaType   string
	AnimeURL    string
	SeasonMap   []metadata.SeasonMapping
	Meta        *util.MediaMeta // External IDs and year for folder naming
}

// snapshotMedia returns a consistent point-in-time copy of the global media state.
func snapshotMedia() mediaSnapshot {
	gMedia.mu.RLock()
	defer gMedia.mu.RUnlock()
	return mediaSnapshot{
		AnimeName:   gMedia.animeName,
		AnimeSeason: gMedia.animeSeason,
		IsMovieOrTV: gMedia.isMovieOrTV,
		MediaType:   gMedia.mediaType,
		AnimeURL:    gMedia.animeURL,
		SeasonMap:   append([]metadata.SeasonMapping(nil), gMedia.seasonMap...),
		Meta:        cloneMediaMeta(gMedia.meta),
	}
}

func cloneMediaMeta(meta *util.MediaMeta) *util.MediaMeta {
	if meta == nil {
		return nil
	}
	cloned := *meta
	return &cloned
}

// SetSeasonMap stores the AniList-derived season mapping for per-episode
// season resolution. Call after metadata enrichment.
func SetSeasonMap(sm []metadata.SeasonMapping) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.seasonMap = append([]metadata.SeasonMapping(nil), sm...)
}

// SetMediaMeta stores external IDs (TMDB, IMDB, AniList, MAL) and year for
// Plex/Jellyfin-compatible folder naming. Call after metadata enrichment.
func SetMediaMeta(meta *util.MediaMeta) {
	gMedia.mu.Lock()
	defer gMedia.mu.Unlock()
	gMedia.meta = cloneMediaMeta(meta)
}

// GetMediaMeta returns the current media metadata (external IDs and year).
func GetMediaMeta() *util.MediaMeta {
	gMedia.mu.RLock()
	defer gMedia.mu.RUnlock()
	return cloneMediaMeta(gMedia.meta)
}

// resolveSeasonForEpisode returns the correct (season, episode) pair for an
// absolute episode number using the AniList season map. Falls back to
// (snap.AnimeSeason, absEp) when no map is available.
func resolveSeasonForEpisode(snap mediaSnapshot, absEp int) (season, ep int) {
	if snap.AnimeSeason > 1 {
		for _, sm := range snap.SeasonMap {
			if sm.Season == snap.AnimeSeason && absEp >= 1 && absEp <= sm.EpisodeCount {
				util.Debugf("resolveSeasonForEpisode: using selected season S%02d local episode %02d", snap.AnimeSeason, absEp)
				return snap.AnimeSeason, absEp
			}
		}
	}
	if len(snap.SeasonMap) == 0 {
		util.Debugf("resolveSeasonForEpisode: no season map, using default season=%d, ep=%d", max(snap.AnimeSeason, 1), absEp)
		return max(snap.AnimeSeason, 1), absEp
	}
	util.Debugf("resolveSeasonForEpisode: seasonMap has %d entries, absEp=%d", len(snap.SeasonMap), absEp)
	for _, sm := range snap.SeasonMap {
		if absEp >= sm.StartEp && absEp <= sm.EndEp {
			util.Debugf("resolveSeasonForEpisode: matched S%02d range [%d-%d] → s%02de%02d", sm.Season, sm.StartEp, sm.EndEp, sm.Season, absEp-sm.StartEp+1)
			return sm.Season, absEp - sm.StartEp + 1
		}
	}
	// Beyond known range: use last season
	last := snap.SeasonMap[len(snap.SeasonMap)-1]
	return last.Season, absEp - last.StartEp + 1
}

// tickMsg is a message for the tick command
type tickMsg time.Time

// statusMsg is a message to update the status
type statusMsg string

// model represents the Bubble Tea model for the progress bar and status
type model struct {
	progress   progress.Model
	totalBytes int64
	received   int64
	peakPct    float64 // highest progress percentage ever reached; ensures bar never goes backward
	done       bool
	doneFrames int // frames elapsed since done; allows 100% to render before quit
	status     string
	err        error // download error propagated from goroutine
	mu         sync.Mutex
	keys       keyMap

	parent       *model
	taskID       string
	taskTotals   map[string]int64
	taskReceived map[string]int64
	// taskEnded holds how each finished task of a batch ended: true when it
	// was downloaded, false when it failed.
	taskEnded map[string]bool

	// What the screen names and measures: the show, the episode or the
	// count of episodes, the terminal width, the speed and the start.
	title   string
	detail  string
	width   int
	rate    tui.RateMeter
	started time.Time
}

// downloadDetail names the episode a single download is for: "Episode 3".
func downloadDetail(episodeNumberStr string) string {
	if num := ExtractEpisodeNumber(episodeNumberStr); num != "" {
		return "Episode " + num
	}
	return strings.TrimSpace(episodeNumberStr)
}

// newDownloadModel is the progress screen of a download: title is the show,
// detail which part of it ("Episode 3", "12 episodes"), totalBytes the size
// expected so far (0 when unknown).
func newDownloadModel(title, detail string, totalBytes int64) *model {
	return &model{
		progress:   tui.NewTransferBar(),
		keys:       keyMap{quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
		totalBytes: totalBytes,
		title:      title,
		detail:     detail,
		started:    time.Now(),
	}
}

type keyMap struct {
	quit key.Binding
}

// Init initializes the Bubble Tea model
func (m *model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.progress.Init())
}

func (m *model) childProgress(taskID string, estimatedTotal int64) *model {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskTotals == nil {
		m.taskTotals = make(map[string]int64)
	}
	if m.taskReceived == nil {
		m.taskReceived = make(map[string]int64)
	}
	if estimatedTotal > 0 {
		if old := m.taskTotals[taskID]; old != estimatedTotal {
			m.totalBytes += estimatedTotal - old
			m.taskTotals[taskID] = estimatedTotal
		}
	}
	return &model{parent: m, taskID: taskID}
}

// endTask records how a task of a batch ended. A downloaded task's total
// becomes what it actually received, since its size was only estimated up
// front, so the overall bar ends where the bytes do.
func (m *model) endTask(taskID string, downloaded bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskEnded == nil {
		m.taskEnded = make(map[string]bool)
	}
	m.taskEnded[taskID] = downloaded
	if received := m.taskReceived[taskID]; downloaded && received > 0 {
		if m.taskTotals == nil {
			m.taskTotals = make(map[string]int64)
		}
		m.totalBytes += received - m.taskTotals[taskID]
		m.taskTotals[taskID] = received
	}
}

func (m *model) setProgressTotal(total int64) {
	if m == nil || total <= 0 {
		return
	}
	if m.parent != nil {
		m.parent.setTaskTotal(m.taskID, total)
		return
	}
	m.mu.Lock()
	m.totalBytes = total
	m.mu.Unlock()
}

func (m *model) progressTotal() int64 {
	if m == nil {
		return 0
	}
	if m.parent != nil {
		return m.parent.taskTotal(m.taskID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalBytes
}

func (m *model) taskTotal(taskID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.taskTotals[taskID]
}

func (m *model) shouldGrowProgressTotal(total int64) bool {
	if total <= 0 {
		return false
	}
	current := m.progressTotal()
	return total > current
}

func (m *model) setTaskTotal(taskID string, total int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskTotals == nil {
		m.taskTotals = make(map[string]int64)
	}
	old := m.taskTotals[taskID]
	if old == total {
		return
	}
	m.taskTotals[taskID] = total
	m.totalBytes += total - old
}

func (m *model) addProgressReceived(delta int64) {
	if m == nil || delta <= 0 {
		return
	}
	if m.parent != nil {
		m.parent.addTaskReceived(m.taskID, delta)
		return
	}
	m.mu.Lock()
	m.received += delta
	m.mu.Unlock()
}

func (m *model) addTaskReceived(taskID string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskReceived == nil {
		m.taskReceived = make(map[string]int64)
	}
	m.taskReceived[taskID] += delta
	m.received += delta
}

func (m *model) setProgressReceived(received int64) {
	if m == nil || received < 0 {
		return
	}
	if m.parent != nil {
		m.parent.setTaskReceived(m.taskID, received)
		return
	}
	m.mu.Lock()
	m.received = received
	m.mu.Unlock()
}

func (m *model) setTaskReceived(taskID string, received int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskReceived == nil {
		m.taskReceived = make(map[string]int64)
	}
	old := m.taskReceived[taskID]
	if received < old {
		received = old
	}
	m.taskReceived[taskID] = received
	m.received += received - old
}

func (m *model) resetProgressReceived() {
	if m == nil {
		return
	}
	if m.parent != nil {
		m.parent.resetTaskReceived(m.taskID)
		return
	}
	m.mu.Lock()
	m.received = 0
	m.peakPct = 0
	m.mu.Unlock()
}

func (m *model) resetTaskReceived(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.taskReceived == nil {
		return
	}
	old := m.taskReceived[taskID]
	delete(m.taskReceived, taskID)
	m.received -= old
	if m.received < 0 {
		m.received = 0
	}
}

func (m *model) setProgressPeak(pct float64) {
	if m == nil || pct <= 0 {
		return
	}
	if m.parent != nil {
		m.parent.setProgressPeak(pct)
		return
	}
	m.mu.Lock()
	if pct > m.peakPct {
		m.peakPct = pct
	}
	m.mu.Unlock()
}

// StartVideo opens mpv with a socket for IPC
// Modify the StartVideo function in player.go
func StartVideo(link string, args []string) (string, error) {
	// Verify MPV is installed (cached after first lookup to avoid repeated filesystem searches)
	cachedMPVPathOnce.Do(func() {
		cachedMPVPath, cachedMPVPathErr = findMPVPath()
	})
	if cachedMPVPathErr != nil {
		return "", fmt.Errorf("mpv not found: %w\nPlease install mpv: https://mpv.io/installation/", cachedMPVPathErr)
	}
	mpvPath := cachedMPVPath

	randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
	var socketPath string

	if runtime.GOOS == "windows" {
		// Keep pipe name short — Windows named-pipe path limit is generous, but
		// shorter names avoid rare third-party filter issues on minimal VMs.
		pipeID := randomNumber
		if len(pipeID) > 16 {
			pipeID = pipeID[:16]
		}
		socketPath = fmt.Sprintf(`\\.\pipe\goanime_mpv_%s`, pipeID)
	} else {
		// Use os.TempDir() for cross-platform compatibility
		// macOS uses /var/folders/... accessed via $TMPDIR
		// filepath.Join handles trailing slashes correctly (fixes macOS double-slash issue)
		socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("goanime_mpvsocket_%s", randomNumber))
	}

	mpvArgs := []string{
		"--no-terminal",
		"--force-window=yes",
		fmt.Sprintf("--input-ipc-server=%s", socketPath),
	}
	// Validate and filter any additional args before passing to mpv
	mpvArgs = append(mpvArgs, filterMPVArgs(args)...)

	// Sanitize media target (URL or local file path)
	safeLink, err := sanitizeMediaTarget(link)
	if err != nil {
		return "", fmt.Errorf("invalid media target: %w", err)
	}
	mpvArgs = append(mpvArgs, safeLink)

	util.Debugf("Starting mpv with arguments: %v", mpvArgs)

	// #nosec G204: mpvArgs are validated via filterMPVArgs and sanitizeMediaTarget
	cmd := exec.Command(mpvPath, mpvArgs...)
	setProcessGroup(cmd) // Handle OS-specific process groups

	// Capture stderr so crash/GPU errors are visible when IPC never appears.
	// Intentionally not using --quiet: silent exits on VMs hide the real cause.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mpv: %w (stderr: %s)", err, stderr.String())
	}

	// Reap the process in the background so we can detect early exit (common on
	// Windows VMs when --vo=gpu fails / bundled mpv is missing DLLs). Without
	// this, StartVideo waits the full timeout then Process.Kill returns
	// "Access is denied" because the process is already dead.
	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- cmd.Wait()
	}()

	util.Debugf("mpv started (pid probe), waiting for socket creation: %s", socketPath)

	// Wait for socket creation with adaptive timeout and exponential backoff.
	// VMs / cold media opens can take longer than a local GPU host.
	maxWaitTime := 12 * time.Second
	initialInterval := 5 * time.Millisecond
	maxInterval := 100 * time.Millisecond
	currentInterval := initialInterval

	for time.Since(startTime) < maxWaitTime {
		// Detect early mpv death before wasting the full timeout.
		select {
		case waitErr := <-waitErrCh:
			return "", formatMPVEarlyExitError(waitErr, stderr.String(), mpvPath)
		default:
		}

		// Try to connect to the socket instead of checking file existence
		// This works for both Unix sockets and Windows named pipes
		conn, err := dialMPVSocket(socketPath)
		if err == nil {
			_ = conn.Close() // Close immediately, we just wanted to verify connectivity
			util.Debugf("Socket connected successfully after %.2fs", time.Since(startTime).Seconds())
			return socketPath, nil
		}

		if cmd.Process == nil {
			return "", fmt.Errorf("mpv process not started properly: %s", stderr.String())
		}

		time.Sleep(currentInterval)
		// Apply exponential backoff with faster growth — reaches max sooner
		currentInterval = min(currentInterval*2, maxInterval)
	}

	elapsed := time.Since(startTime)
	util.Debugf("Timeout after %.2fs waiting for mpv socket", elapsed.Seconds())
	if stderr.Len() > 0 {
		util.Debugf("mpv stderr during timeout: %s", stderr.String())
	}

	// Best-effort cleanup. On Windows Kill returns "Access is denied" if the
	// process already exited between the last probe and here — ignore that.
	if cmd.Process != nil {
		if killErr := cmd.Process.Kill(); killErr != nil {
			util.Debugf("Failed to kill mpv process (often already exited): %v", killErr)
		}
	}
	// Drain Wait so we don't leak the reaper goroutine's result unnoticed.
	select {
	case waitErr := <-waitErrCh:
		if waitErr != nil {
			util.Debugf("mpv wait after timeout: %v", waitErr)
		}
	case <-time.After(500 * time.Millisecond):
	}

	stderrHint := strings.TrimSpace(stderr.String())
	if stderrHint != "" {
		return "", fmt.Errorf("timeout waiting for mpv socket after %.1fs.\nmpv stderr: %s\nmpv path: %s\nPossible issues:\n1. MPV crashed or GPU/video output failed (common on VMs — try updating GPU drivers or reinstall mpv)\n2. MPV installation corrupted / missing DLLs\n3. Invalid video URL\nCheck debug logs with -debug flag", elapsed.Seconds(), stderrHint, mpvPath)
	}
	return "", fmt.Errorf("timeout waiting for mpv socket after %.1fs.\nmpv path: %s\nPossible issues:\n1. MPV hung before creating IPC (GPU/driver issue on VMs)\n2. MPV installation corrupted / missing DLLs\n3. Invalid video URL\nRun: \"%s\" --version\nCheck debug logs with -debug flag", elapsed.Seconds(), mpvPath, mpvPath)
}

// formatMPVEarlyExitError builds a clear error when mpv dies before the IPC
// socket appears. This is the common failure mode on minimal Windows VMs.
func formatMPVEarlyExitError(waitErr error, stderrOut, mpvPath string) error {
	stderrOut = strings.TrimSpace(stderrOut)
	util.Debugf("mpv exited before IPC socket was ready: %v stderr=%q path=%s", waitErr, stderrOut, mpvPath)
	if stderrOut != "" {
		return fmt.Errorf("mpv exited before IPC socket was ready: %v\nmpv stderr: %s\nmpv path: %s\nHint: on Windows VMs OpenGL often fails — GoAnime uses a VO fallback chain; if this persists, reinstall mpv or run mpv manually", waitErr, stderrOut, mpvPath)
	}
	return fmt.Errorf("mpv exited before IPC socket was ready: %v\nmpv path: %s (no stderr captured)\nHint: bundled mpv may be missing DLLs, or video output failed. Run: \"%s\" --version", waitErr, mpvPath, mpvPath)
}

// MpvSendCommand is a wrapper function to expose mpvSendCommand to other packages
func MpvSendCommand(socketPath string, command []any) (any, error) {
	return mpvSendCommand(socketPath, command)
}

// filterMPVArgs whitelists allowed mpv flags to avoid passing unexpected parameters.
func filterMPVArgs(args []string) []string {
	allowedNoValue := map[string]struct{}{
		"--no-config": {},
	}
	allowedWithValuePrefixes := []string{
		"--hwdec=",
		"--vo=",
		"--gpu-context=",
		"--profile=",
		"--cache=",
		"--demuxer-max-bytes=",
		"--demuxer-readahead-secs=",
		"--video-latency-hacks=",
		"--keep-open=",       // user-config guards (see userConfigGuardArgs)
		"--idle=",            //
		"--loop-file=",       //
		"--loop-playlist=",   //
		"--resume-playback=", //
		"--audio-display=",
		"--start=",
		"--alang=",                     // Audio language preference
		"--slang=",                     // Subtitle language preference
		"--aid=",                       // Audio track ID
		"--sid=",                       // Subtitle track ID
		"--sub-file=",                  // External subtitle file (one flag per track; never colon-join https URLs)
		"--sub-files=",                 // legacy mpv multi-file form (unused — colon breaks https://)
		"--audio-file=",                // External audio file
		"--http-header-fields=",        // HTTP headers for HLS streams
		"--http-header-fields-append=", // one header per option; required for values containing a comma
		"--stream-lavf-o=",             // FFmpeg/lavf options for streaming protocols
		"--demuxer-lavf-o=",            // FFmpeg/lavf demuxer options (e.g. allowed_extensions=ALL so HLS audio renditions with disguised segment extensions load)
		"--demuxer-lavf-format=",       // Force a lavf demuxer format (e.g. hls)
		"--referrer=",                  // HTTP referrer for streaming
		"--user-agent=",                // HTTP user agent for streaming
		// Anime4K real-time upscaling shaders
		"--glsl-shader=",          // GLSL shader for video processing
		"--glsl-shaders=",         // Multiple GLSL shaders (colon-separated)
		"--gpu-shader-cache-dir=", // Shader cache directory
		"--gpu-api=",              // GPU API selection (auto, opengl, vulkan, d3d11)
		// yt-dlp integration for Cloudflare-protected streams (9Anime, etc.)
		"--script-opts=",             // mpv script options (e.g. ytdl_hook-try_ytdl_first)
		"--ytdl-raw-options-append=", // Pass raw options to yt-dlp backend
		"--ytdl-format=",             // yt-dlp format / quality selection
		"--ytdl=",                    // Enable/disable yt-dlp (e.g. --ytdl=no for local proxy URLs)
		"--force-media-title=",       // Override media title shown in MPV window
		// Add more allowed prefixes here if needed in the future
	}

	var filtered []string
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			// ignore positional args; media target is handled separately
			continue
		}
		if _, ok := allowedNoValue[a]; ok {
			filtered = append(filtered, a)
			continue
		}
		for _, p := range allowedWithValuePrefixes {
			if strings.HasPrefix(a, p) {
				filtered = append(filtered, a)
				break
			}
		}
	}
	return filtered
}

// sanitizeMediaTarget ensures the media target is a safe http(s) URL or a cleaned file path
func sanitizeMediaTarget(link string) (string, error) {
	l := strings.TrimSpace(link)
	if l == "" {
		return "", fmt.Errorf("empty link")
	}
	if strings.ContainsAny(l, "\x00\n\r") {
		return "", fmt.Errorf("invalid control characters in link")
	}
	if strings.HasPrefix(l, "-") {
		return "", fmt.Errorf("media target must not start with '-' (looks like a flag)")
	}
	// Treat as URL only if it contains "://". This avoids misclassifying Windows
	// paths like "C:\\..." as having scheme "c".
	if strings.Contains(l, "://") {
		u, err := neturl.Parse(l)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			return l, nil
		default:
			return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
		}
	}
	// Treat as local path
	cleaned := filepath.Clean(l)
	return cleaned, nil
}

// sanitizeOutputPath validates an output path to avoid directory traversal and disallow leading '-'
func sanitizeOutputPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty output path")
	}
	if strings.ContainsAny(p, "\x00\n\r") {
		return "", fmt.Errorf("invalid control characters in output path")
	}
	if strings.HasPrefix(p, "-") {
		return "", fmt.Errorf("output path must not start with '-' (looks like a flag)")
	}
	cleaned := filepath.Clean(p)
	// Verify the resolved path stays within user home to prevent path traversal
	userHome, err := os.UserHomeDir()
	if err == nil {
		abs, absErr := filepath.Abs(cleaned)
		if absErr == nil && !strings.HasPrefix(abs, userHome) {
			return "", fmt.Errorf("output path escapes user home directory")
		}
	}
	return cleaned, nil
}

// mpvDebugf is where the mpv conversation writes its diagnostics; tests replace
// it. Only an unexpected answer is written: the position, duration, pause state
// and speed are asked for several times a second, and logging each answer put
// 70 lines (and as many blank ones) in a 15 second session.
var mpvDebugf = util.Debugf

// mpvSendCommand sends a JSON command to MPV via the IPC socket and receives the response.
func mpvSendCommand(socketPath string, command []any) (any, error) {
	conn, err := dialMPVSocket(socketPath)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing mpv socket")
		}
	}(conn)

	commandJSON, err := json.Marshal(map[string]any{
		"command": command,
	})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(commandJSON, '\n'))
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	// Tratar múltiplos JSONs na mesma resposta
	responses := bytes.SplitSeq(buffer[:n], []byte("\n"))
	for resp := range responses {
		if len(bytes.TrimSpace(resp)) == 0 {
			continue
		}
		var response map[string]any
		err = jsonx.Unmarshal(resp, &response)
		if err != nil {
			util.Debugf("Error when unmarshaling: %v", err)
			continue
		}
		if errStr, ok := response["error"].(string); ok && errStr == "property unavailable" {
			// Propriedade ainda não disponível, ignore sem erro. Expected
			// before a file is loaded, so it is not logged.
			continue
		}
		if errStr, ok := response["error"].(string); ok && errStr != "success" {
			mpvDebugf("mpv answered %q to %v", errStr, command)
		}
		// Check for success response (set_property returns {"error":"success"} without data)
		if errStr, ok := response["error"].(string); ok && errStr == "success" {
			// Command succeeded, return nil data
			if data, exists := response["data"]; exists {
				return data, nil
			}
			return nil, nil
		}
		if data, exists := response["data"]; exists {
			return data, nil
		}
	}
	return nil, errors.New("no data field in mpv response")
}

// windows
// dialMPVSocket creates a connection to mpv's socket.
// func dialMPVSocket(socketPath string) (net.Conn, error) {
//	if runtime.GOOS == "windows" {
//		// Attempt named pipe on Windows
//		return net.Dial("unix", socketPath)
//	} else {
//		// Unix-like system uses Unix sockets
//		return net.Dial("unix", socketPath)
//	}
// }

// Funções de download extraídas de player.go
// downloadPart, combineParts, DownloadVideo, downloadWithYtDlp, ExtractVideoSources, getBestQualityURL, ExtractVideoSourcesWithPrompt, HandleBatchDownload, getEpisodeRange, findEpisode, createEpisodePath, fileExists
// As implementações completas estão agora em download.go

// HandleDownloadAndPlay handles the download and playback of the video
func HandleDownloadAndPlay(
	videoURL string,
	episodes []models.Episode,
	selectedEpisodeNum int,
	animeURL string,
	episodeNumberStr string,
	animeMalID int,
	animeAnilistID int,
	updater *discord.RichPresenceUpdater,
	animeName string,
	animeSeason int,
	anime *models.Anime,
) error {
	util.Debug("HandleDownloadAndPlay called", "videoURL", videoURL, "episodeNum", selectedEpisodeNum)

	// Persist the anime URL/ID to aid episode switching when updater is nil (e.g., Discord disabled)
	setLastAnimeURL(animeURL)

	// Store anime name for Plex-compatible download file naming
	if animeName != "" {
		season := max(animeSeason, 1)
		if util.GlobalDownloadRequest != nil && util.GlobalDownloadRequest.SeasonNum > 0 {
			season = util.GlobalDownloadRequest.SeasonNum
		}
		SetAnimeName(animeName, season)
	}

	// Record the show context for tracking writes: the SEARCH home screen's
	// CONTINUE list reads the name, source and audio back from the database.
	if anime != nil {
		SetMediaContext(tui.SingleLine(animeName), tui.SingleLine(anime.Source), tui.SingleLine(anime.Audio))
	} else {
		SetMediaContext(tui.SingleLine(animeName), "", "")
	}

	// Check if this is an HLS stream (for proper handling later)
	isHLSStream := LooksLikeHLS(videoURL)
	util.Debug("Stream type", "isHLS", isHLSStream)

	for {
		downloadOption := downloadMenuOption(selectedEpisodeNum, len(episodes), animeName)
		switch downloadOption {
		case 0:
			// User wants to go back to server selection
			return ErrBackToEpisodeSelection
		case 1:
			// Download the current episode
			if isHLSStream {
				// HLS streams need special download handling
				util.Debugf("HLS download requested - using stream URL")
			}
			err := downloadAndPlayEpisode(
				videoURL,
				episodes,
				selectedEpisodeNum,
				animeURL,
				episodeNumberStr,
				animeMalID,
				animeAnilistID,
				updater,
			)
			if err != nil {
				if errors.Is(err, ErrBackToDownloadOptions) {
					// Go back to download options menu
					RequestEpisodeMenu(EpisodeMenuRequest{ShowMenu: true})
					continue
				}
				return err
			}
			return nil
		case 2:
			// Download episodes in a range
			if start := takeRangeStart(); start > 0 {
				if err := downloadRangeFrom(start, episodes, anime); err != nil {
					return err
				}
				return nil
			}
			if err := HandleBatchDownload(episodes, anime); err != nil {
				return err
			}
			return nil
		case 3:
			// Upscale video with Anime4K
			if err := handleUpscaleFromMenu(); err != nil {
				util.Errorf("Upscale error: %v", err)
			}
			continue // Return to menu after upscaling
		case 5:
			// Download ALL episodes
			if len(episodes) == 0 {
				util.Errorf("No episodes available to download")
				continue
			}
			if !confirmDownloadAll(len(episodes)) {
				continue
			}
			if err := HandleDownloadAll(episodes, anime); err != nil {
				if errors.Is(err, ErrUserQuit) {
					return nil
				}
				return err
			}
			return nil
		default:
			// Play online - determine the best approach based on URL type
			videoURLToPlay := ""

			switch {
			case isHLSStream:
				// HLS streams are already resolved, play directly
				videoURLToPlay = videoURL
				if util.IsDebug {
					util.Debugf("HLS stream detected, playing directly: %s", videoURLToPlay)
				}
			case videoURL != "" && needsVideoExtraction(videoURL):
				// Intermediate URL (e.g. animefire.io/video/) needs resolution
				// to obtain the final CDN video URL.
				if util.IsDebug {
					util.Debugf("Intermediate URL detected, resolving: %s", videoURL)
				}
				if resolved, err := extractActualVideoURL(videoURL); err == nil && resolved != "" {
					videoURLToPlay = resolved
				}
			case videoURL != "" && strings.HasPrefix(videoURL, "http"):
				// The enhanced API already resolved a direct stream URL (CDN,
				// mp4, etc.). Use it directly — re-extracting may trigger
				// duplicate quality prompts or cause CDN URLs to expire.
				videoURLToPlay = videoURL
				if util.IsDebug {
					util.Debugf("Using resolved stream URL directly: %s", videoURLToPlay)
				}
			case videoURL != "":
				// Non-HTTP URL (e.g. episode ID). Try legacy extraction.
				if len(episodes) > 0 && selectedEpisodeNum > 0 {
					selectedEp, found := findEpisode(episodes, selectedEpisodeNum)
					if found {
						if util.IsDebug {
							util.Debugf("Extracting URL from episode page: %s", selectedEp.URL)
						}
						url, err := ExtractVideoSourcesWithPrompt(selectedEp.URL)
						if errors.Is(err, ErrBackRequested) {
							continue
						}
						if err == nil && url != "" {
							videoURLToPlay = url
						}
					}
				}
				// Fallback: try to extract from original videoURL
				if videoURLToPlay == "" {
					if util.IsDebug {
						util.Debugf("Fallback: extracting from original URL: %s", videoURL)
					}
					url, err := ExtractVideoSourcesWithPrompt(videoURL)
					if errors.Is(err, ErrBackRequested) {
						continue
					}
					if err == nil && url != "" {
						videoURLToPlay = url
					}
				}
			}

			// Final validation
			if videoURLToPlay == "" {
				util.Debugf("No valid video URL found")
				return fmt.Errorf("no valid video URL found")
			}

			if util.IsDebug {
				util.Debugf("Final video URL: %s", videoURLToPlay)
			}

			err := playVideo(
				videoURLToPlay,
				episodes,
				selectedEpisodeNum,
				animeMalID,
				animeAnilistID,
				updater,
			)
			if err != nil {
				if errors.Is(err, ErrBackToDownloadOptions) {
					// Go back to download options menu
					RequestEpisodeMenu(EpisodeMenuRequest{ShowMenu: true})
					continue
				}
				return err
			}
			return nil
		}
	}
}

func downloadAndPlayEpisode(
	videoURL string,
	episodes []models.Episode,
	selectedEpisodeNum int,
	animeURL string,
	episodeNumberStr string,
	animeMalID int,
	animeAnilistID int,
	updater *discord.RichPresenceUpdater,
) error {
	// Check if video URL is valid
	if videoURL == "" {
		return fmt.Errorf("empty video URL provided for episode %s", episodeNumberStr)
	}
	util.SetSubtitleContext(snapshotMedia().AnimeName, util.GetGlobalAnimeSource(), "", "")

	// Resolve intermediate URLs (e.g., animefire.io/video/ JSON API) to actual CDN URLs
	if strings.Contains(videoURL, "animefire.io/video/") {
		util.Debug("Resolving AnimeFire video API URL before download", "url", videoURL)
		resolved, err := extractActualVideoURL(videoURL)
		if err != nil {
			return fmt.Errorf("failed to resolve AnimeFire video URL: %w", err)
		}
		if resolved != "" {
			util.Debug("Resolved AnimeFire URL", "resolved", resolved)
			videoURL = resolved
		}
	}

	// Use Plex-compatible naming when anime name is available
	var downloadPath, episodePath string
	snap := snapshotMedia()
	if snap.AnimeName != "" {
		// Route to the correct base directory: movies/ for movies/TV, anime/ for anime
		var baseDir string
		if snap.IsMovieOrTV {
			baseDir = util.DefaultMovieDownloadDir()
		} else {
			baseDir = util.DefaultDownloadDir()
		}

		// Check if this is a standalone movie (flat path) vs TV/anime (season structure)
		if snap.MediaType == "movie" {
			// Movies: flat structure <baseDir>/<MovieName (Year) {ids}>/
			downloadPath = util.FormatPlexMovieDir(baseDir, snap.AnimeName, snap.Meta)
			episodePath = util.FormatPlexMoviePath(baseDir, snap.AnimeName, "", snap.Meta)
		} else {
			// TV Shows and Anime: season/episode structure
			// Use the int episode number directly; fall back to parsing the string only if needed
			epNum := selectedEpisodeNum
			if epNum < 1 {
				parsed, _ := strconv.Atoi(episodeNumberStr)
				if parsed > 0 {
					epNum = parsed
				} else {
					epNum = 1
				}
			}
			season, relEp := resolveSeasonForEpisode(snap, epNum)
			downloadPath = util.FormatPlexEpisodeDir(baseDir, snap.AnimeName, season, snap.Meta)
			episodePath = util.FormatPlexEpisodePath(baseDir, snap.AnimeName, season, relEp, snap.Meta)
		}
		util.Debugf("Download routing: mediaType=%s, isMovieOrTV=%v, baseDir=%s, path=%s", snap.MediaType, snap.IsMovieOrTV, baseDir, episodePath)
	} else {
		// Fallback: route based on media type even without anime name
		fallbackBase := util.DefaultDownloadDir()
		if snap.IsMovieOrTV {
			fallbackBase = util.DefaultMovieDownloadDir()
		}
		downloadPath = filepath.Join(fallbackBase, DownloadFolderFormatter(animeURL))
		episodePath = filepath.Join(downloadPath, episodeNumberStr+".mp4")
	}

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadPath, 0o700); err != nil {
			return fmt.Errorf("failed to create download directory: %w", err)
		}
	}

	// Prompt user to select subtitle language BEFORE download starts
	// (stdin is free here — no Bubble Tea running yet)
	// For 9Anime, ALWAYS use the mandatory language prompt regardless of track count.
	if util.Is9AnimeSource() {
		util.PromptSubtitleLanguage()
	} else if len(util.GetGlobalSubtitles()) > 0 {
		util.SelectSubtitles()
	}

	if _, err := os.Stat(episodePath); os.IsNotExist(err) {
		numThreads := 4 // Define the number of threads for downloading

		// Check URL type and use appropriate download method
		switch {
		case isBloggerProxyURL(videoURL):
			// Download directly from googlevideo CDN, bypassing the proxy.
			// Uses independent surf clients per chunk for parallel download
			// with Chrome TLS and automatic resume on connection drops.
			directURL := GetBloggerVideoURL()
			if directURL == "" {
				return fmt.Errorf("blogger video URL not available")
			}

			// Fail fast without a TTY: the progress program blocks on Windows
			// console input headless instead of erroring like Linux.
			if !tui.HasTerminal() {
				return fmt.Errorf("no terminal: progress UI requires a TTY")
			}
			m := newDownloadModel(snapshotMedia().AnimeName, downloadDetail(episodeNumberStr), 0)
			p := tui.NewProgram(m)

			go func() {
				dlErr := downloadBloggerDirect(directURL, episodePath, numThreads, m)
				if dlErr != nil {
					m.mu.Lock()
					m.err = fmt.Errorf("failed to download video: %w", dlErr)
					m.done = true
					m.mu.Unlock()
					p.Send(statusMsg("Download failed"))
					return
				}
				// Set final size + done in a single lock to prevent the tick
				// handler from seeing 100% with done=false (causes visual jump).
				m.mu.Lock()
				if fi, statErr := os.Stat(episodePath); statErr == nil && fi.Size() > 0 {
					m.totalBytes = fi.Size()
					m.received = fi.Size()
				}
				m.done = true
				m.mu.Unlock()
				p.Send(statusMsg("Download complete"))
			}()

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running progress bar: %w", err)
			}
			if m.err != nil {
				return m.err
			}

			if _, err := os.Stat(episodePath); os.IsNotExist(err) {
				return fmt.Errorf("download failed: file was not created")
			}

			const minEpisodeSize int64 = 10 * 1024 * 1024
			if stat, err := os.Stat(episodePath); err == nil && stat.Size() < minEpisodeSize {
				_ = os.Remove(episodePath)
				return fmt.Errorf("download incomplete: file is only %d bytes (%.1f MB), expected at least %.0f MB",
					stat.Size(), float64(stat.Size())/(1024*1024), float64(minEpisodeSize)/(1024*1024))
			}

			fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
			printDownloadLocation(episodePath)
			downloadSubtitleFiles(episodePath, util.GetGlobalSubtitles(), func(format string, a ...any) {
				fmt.Printf(format, a...)
			})

		case strings.Contains(videoURL, "blogger.com") ||
			LooksLikeHLS(videoURL) ||
			strings.Contains(videoURL, "wixmp.com") ||
			strings.Contains(videoURL, "sharepoint.com"):
			// Fail fast without a TTY (see above): progress UI hangs on Windows.
			if !tui.HasTerminal() {
				return fmt.Errorf("no terminal: progress UI requires a TTY")
			}
			// Use yt-dlp with progress bar
			m := newDownloadModel(snapshotMedia().AnimeName, downloadDetail(episodeNumberStr), 0)
			p := tui.NewProgram(m)

			// Estimate/obtain total size for progress percentage.
			// For HLS streams, do NOT pre-seed a large estimate (like 500 MB)
			// because the native HLS or yt-dlp callbacks will set the real
			// total dynamically. A wrong initial value makes the bar stuck.
			if LooksLikeHLS(videoURL) {
				m.totalBytes = 0 // let download callbacks set the real value
			} else {
				httpClient := &http.Client{Transport: downloadTransport(10 * time.Second)}
				if sz, err := getContentLength(videoURL, httpClient); err == nil && sz > 0 {
					m.totalBytes = sz
				} else {
					m.totalBytes = 150 * 1024 * 1024
				}
			}

			go func() {
				// Native HLS first for .m3u8 — handles obfuscated segment extensions
				// (.js, .html, .jpg) that break yt-dlp/ffmpeg.
				// SharePoint URLs (.aspx) may serve HLS or direct video; yt-dlp rejects the extension.
				var dlErr error
				switch {
				case LooksLikeHLS(videoURL) || hasUnsafeExtension(videoURL):
					dlErr = downloadWithNativeHLS(videoURL, episodePath, m)
					if dlErr != nil && stderrors.Is(dlErr, hls.ErrSeparateAudioTracks) {
						// Separate audio tracks need yt-dlp for proper audio/video merging
						util.Debugf("HLS has separate audio tracks, using yt-dlp: %v", dlErr)
						// Reset progress — native HLS didn't produce output
						m.mu.Lock()
						m.received = 0
						m.totalBytes = 0
						m.peakPct = 0
						m.mu.Unlock()
						dlErr = downloadWithYtDlp(videoURL, episodePath, m)
					} else if dlErr != nil {
						// Native HLS failed for another reason — try yt-dlp, then direct HTTP
						util.Debugf("Native HLS failed, trying yt-dlp: %v", dlErr)
						// Reset progress for yt-dlp fallback
						m.mu.Lock()
						m.received = 0
						m.totalBytes = 0
						m.peakPct = 0
						m.mu.Unlock()
						dlErr = downloadWithYtDlp(videoURL, episodePath, m)
						if dlErr != nil {
							util.Debugf("yt-dlp failed, trying direct HTTP: %v", dlErr)
							dlErr = downloadDirectHTTP(videoURL, episodePath, m)
						}
					}
				default:
					dlErr = downloadWithYtDlp(videoURL, episodePath, m)
				}
				if dlErr == nil {
					dlErr = validateDownloadedVideo(episodePath)
					if dlErr != nil {
						_ = os.Remove(episodePath)
					}
				}
				if dlErr != nil {
					m.mu.Lock()
					m.err = fmt.Errorf("failed to download video: %w", dlErr)
					m.done = true
					m.mu.Unlock()
					p.Send(statusMsg("Download failed"))
					return
				}
				// Set final size + done in a single lock to prevent the tick
				// handler from seeing 100% with done=false (causes visual jump).
				m.mu.Lock()
				if fi, statErr := os.Stat(episodePath); statErr == nil && fi.Size() > 0 {
					m.totalBytes = fi.Size()
					m.received = fi.Size()
				}
				m.done = true
				m.mu.Unlock()
				p.Send(statusMsg("Download complete"))
			}()

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running progress bar: %w", err)
			}
			if m.err != nil {
				return m.err
			}

			// Verify the file was actually downloaded
			if _, err := os.Stat(episodePath); os.IsNotExist(err) {
				return fmt.Errorf("download failed: file was not created")
			}

			// Verify the file is a reasonable size for a video episode.
			// HLS episodes are typically at least 20 MB; anything below 10 MB
			// almost certainly indicates a truncated or failed download.
			fmt.Printf("Download of episode %s completed!\n", episodeNumberStr)
			printDownloadLocation(episodePath)

			// Download selected subtitles alongside the video file
			downloadSubtitleFiles(episodePath, util.GetGlobalSubtitles(), func(format string, a ...any) {
				fmt.Printf(format, a...)
			})

		default:
			// Fail fast without a TTY (see above): progress UI hangs on Windows.
			if !tui.HasTerminal() {
				return fmt.Errorf("no terminal: progress UI requires a TTY")
			}
			// Initialize progress model
			m := newDownloadModel(snapshotMedia().AnimeName, downloadDetail(episodeNumberStr), 0)
			p := tui.NewProgram(m)

			// Get content length
			httpClient := &http.Client{
				Transport: downloadTransport(10 * time.Second),
			}
			contentLength, err := getContentLength(videoURL, httpClient)
			if err != nil {
				util.Warnf("Failed to get content length: %v, using fallback estimate", err)
				contentLength = 200 * 1024 * 1024 // 200MB fallback
			}
			m.totalBytes = contentLength

			// Start the download in a separate goroutine
			go func() {
				// Update status

				if dlErr := DownloadVideo(videoURL, episodePath, numThreads, m); dlErr != nil {
					m.mu.Lock()
					m.err = fmt.Errorf("failed to download video: %w", dlErr)
					m.done = true
					m.mu.Unlock()
					p.Send(statusMsg("Download failed"))
					return
				}

				m.mu.Lock()
				m.done = true
				m.mu.Unlock()

				// Final status update
				p.Send(statusMsg("Download complete"))
			}()

			// Run the Bubble Tea program in the main goroutine
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running progress bar: %w", err)
			}
			if m.err != nil {
				return m.err
			}

			// Download selected subtitles alongside the video file
			printDownloadLocation(episodePath)
			downloadSubtitleFiles(episodePath, util.GetGlobalSubtitles(), func(format string, a ...any) {
				fmt.Printf(format, a...)
			})
		}
	} else {
		fmt.Println("Video already downloaded.")
		// Check if the file is actually valid (not empty)
		if stat, err := os.Stat(episodePath); err == nil {
			if stat.Size() < 1024 {
				fmt.Println("File is too small, re-downloading...")
				if removeErr := os.Remove(episodePath); removeErr != nil {
					util.Warnf("Failed to remove invalid file: %v", removeErr)
				}
				return downloadAndPlayEpisode(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr, animeMalID, animeAnilistID, updater)
			}
		}
	}

	if askForPlayOffline() {
		if err := playVideo(episodePath, episodes, selectedEpisodeNum, animeMalID, animeAnilistID, updater); err != nil {
			return err
		}
		return nil
	}
	// User chose not to watch; terminate flow cleanly
	return ErrUserQuit
}

// downloadMenuOption decides how playback proceeds: it plays directly
// unless the OPTIONS menu was requested (m key, --download-menu) or an
// episodes-screen download was picked (d/D/A auto-confirm).
func downloadMenuOption(episodeNum, totalEpisodes int, animeName string) int {
	req, _ := TakeEpisodeMenu()
	currentMenuRequest = req
	if req.DownloadCode != 0 {
		return req.DownloadCode
	}
	if req.ShowMenu || util.GlobalDownloadMenu {
		return askForDownload(episodeNum, totalEpisodes, animeName)
	}
	return 4 // play online
}

// currentMenuRequest is the request downloadMenuOption consumed for this
// pass: case 2 reads its range start from it.
var currentMenuRequest EpisodeMenuRequest

// takeRangeStart consumes the D key's preset range start, if any.
func takeRangeStart() int {
	start := currentMenuRequest.RangeStart
	currentMenuRequest.RangeStart = 0
	return start
}

// The codes askForDownload returns, which the playback loop switches on.
const (
	downloadBack    = 0 // back out of the options
	downloadThis    = 1 // download this episode
	downloadRange   = 2 // download a range
	downloadUpscale = 3 // open the upscaling menu
	downloadWatch   = 4 // watch now
	downloadAll     = 5 // download every episode
)

// Seams for the interactive menus in this file. The defaults open the real
// terminal UI; tests replace them, so no menu needs a TTY to be exercised.
var (
	stdinIsTerminal     = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	pickItemsFn         = tui.Pick
	confirmFn           = confirmWithHuh
	shaderModeNameFn    = func() string { return upscaler.GetShaderModeName(upscaler.GetShaderMode()) }
	shadersInstalledFn  = upscaler.ShadersInstalled
	ganInstalledFn      = upscaler.GANShadersInstalled
	setShaderModeFn     = upscaler.SetShaderMode
	installShadersFn    = upscaler.InstallShaders
	installGANShadersFn = upscaler.InstallGANShaders
	infoFn              = util.Info
	warnFn              = util.Warn
)

// confirmWithHuh asks a yes/no question in the terminal.
func confirmWithHuh(title, description, yes, no string) (bool, error) {
	var confirmed bool
	confirm := huh.NewConfirm().
		Title(title).
		Description(description).
		Affirmative(yes).
		Negative(no).
		Value(&confirmed)
	if err := tui.RunClean(confirm.WithTheme(tui.FormTheme()).Run); err != nil {
		return false, err
	}
	return confirmed, nil
}

// downloadMenuItems are the rows of the episode options menu. Enter watches
// right away: downloading is a shortcut (d/D/A), never the default. The
// download-all row exists only when there is more than one episode.
func downloadMenuItems(totalEpisodes int, upscaleStatus string) []tui.PickItem {
	items := []tui.PickItem{
		{Label: "Watch now"},
		{Label: "Download this episode", Key: "d"},
		{Label: "Download a range...", Key: "D"},
		{Label: "Real-time upscale", Details: upscaleStatus, Key: "u"},
	}
	if totalEpisodes > 1 {
		items = append(items, tui.PickItem{
			Label: fmt.Sprintf("Download all %d episodes...", totalEpisodes),
			Key:   "A",
		})
	}
	return items
}

// downloadMenuHeader is the breadcrumb and the one-line context of the menu.
func downloadMenuHeader(episodeNum int, animeName string) (breadcrumb, context string) {
	breadcrumb = "Playback > Options"
	if episodeNum > 0 {
		breadcrumb = fmt.Sprintf("Episode %d > Options", episodeNum)
	}
	context = tui.SingleLine(animeName)
	if episodeNum > 0 && context != "" {
		context = fmt.Sprintf("%s · Episode %d", context, episodeNum)
	}
	return breadcrumb, context
}

// downloadChoiceFor maps the picked row to its code. A row that does not exist
// plays the episode: the one thing an unexpected index must never do is start
// downloading the whole series.
func downloadChoiceFor(idx int, hasDownloadAll bool) int {
	switch {
	case idx == 1:
		return downloadThis
	case idx == 2:
		return downloadRange
	case idx == 3:
		return downloadUpscale
	case idx == 4 && hasDownloadAll:
		return downloadAll
	default:
		return downloadWatch
	}
}

// askForDownload presents the episode options and returns a download* code.
func askForDownload(episodeNum, totalEpisodes int, animeName string) int {
	if !stdinIsTerminal() {
		return downloadWatch // no TTY: default to play online
	}
	breadcrumb, context := downloadMenuHeader(episodeNum, animeName)
	idx, err := pickItemsFn(downloadMenuItems(totalEpisodes, shaderModeNameFn()), tui.PickOptions{
		Breadcrumb:   breadcrumb,
		WindowTitle:  "GoAnime - Options",
		Context:      context,
		ItemSingular: "option",
		ItemPlural:   "options",
		Action:       "watch",
	})
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) {
			return downloadBack
		}
		util.Errorf("Error showing options menu: %v", err)
		return downloadWatch // Default to play online on error
	}
	return downloadChoiceFor(idx, totalEpisodes > 1)
}

// confirmDownloadAll asks before downloading the whole series: the old menu
// put "Download ALL" one arrow from enter, so one wrong tap fetched the
// entire series.
func confirmDownloadAll(totalEpisodes int) bool {
	if !stdinIsTerminal() {
		return true
	}
	confirmed, err := confirmFn(
		fmt.Sprintf("Download all %d episodes?", totalEpisodes),
		"This downloads the whole series.",
		"Yes, download all",
		"No, back",
	)
	if err != nil {
		util.Errorf("Error showing download confirmation: %v", err)
		return false
	}
	return confirmed
}

func askForPlayOffline() bool {
	if !stdinIsTerminal() {
		return false // no TTY: default to no
	}
	idx, err := pickLabelsFn([]string{"Yes", "No"}, tui.PickOptions{
		Breadcrumb:   "Playback > Offline",
		WindowTitle:  "GoAnime - Play Offline",
		ItemSingular: "option",
		ItemPlural:   "options",
	})
	if err != nil {
		util.Errorf("Error showing offline playback menu: %v", err)
		return false // Default to no on error
	}

	return idx == 0 // "Yes" is index 0
}

// playVideo has been moved to playvideo.go

// ToggleSubtitle toggles subtitle visibility
func ToggleSubtitle(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"sub-visibility",
	})
	return err
}

// GetPlaybackStats returns current playback statistics
func GetPlaybackStats(socketPath string) (map[string]any, error) {
	stats := make(map[string]any)

	// Get various playback properties
	properties := []string{
		"time-pos",
		"duration",
		"speed",
		"volume",
		"pause",
		"filename",
	}

	for _, prop := range properties {
		value, err := mpvSendCommand(socketPath, []any{"get_property", prop})
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", prop, err)
		}
		stats[prop] = value
	}

	return stats, nil
}

// SetPlaybackSpeed sets the video playback speed
func SetPlaybackSpeed(socketPath string, speed float64) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"speed",
		speed,
	})
	return err
}

// CycleAudioTrack cycles through available audio tracks
func CycleAudioTrack(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"aid",
	})
	return err
}

// CycleSubtitleTrack cycles through available subtitle tracks
func CycleSubtitleTrack(socketPath string) error {
	_, err := mpvSendCommand(socketPath, []any{
		"cycle",
		"sid",
	})
	return err
}

// SetAudioTrack sets a specific audio track by ID
func SetAudioTrack(socketPath string, trackID int) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"aid",
		trackID,
	})
	return err
}

// SetSubtitleTrack sets a specific subtitle track by ID
func SetSubtitleTrack(socketPath string, trackID int) error {
	_, err := mpvSendCommand(socketPath, []any{
		"set_property",
		"sid",
		trackID,
	})
	return err
}

// GetAudioTracks returns list of available audio tracks
func GetAudioTracks(socketPath string) ([]map[string]any, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "track-list"})
	if err != nil {
		return nil, err
	}

	tracks, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected track-list format")
	}

	var audioTracks []map[string]any
	for _, t := range tracks {
		track, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if trackType, ok := track["type"].(string); ok && trackType == "audio" {
			audioTracks = append(audioTracks, track)
		}
	}
	return audioTracks, nil
}

// GetSubtitleTracks returns list of available subtitle tracks
func GetSubtitleTracks(socketPath string) ([]map[string]any, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "track-list"})
	if err != nil {
		return nil, err
	}

	tracks, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected track-list format")
	}

	var subTracks []map[string]any
	for _, t := range tracks {
		track, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if trackType, ok := track["type"].(string); ok && trackType == "sub" {
			subTracks = append(subTracks, track)
		}
	}
	return subTracks, nil
}

// GetCurrentAudioTrack returns the current audio track ID
func GetCurrentAudioTrack(socketPath string) (int, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "aid"})
	if err != nil {
		return 0, err
	}
	if id, ok := result.(float64); ok {
		return int(id), nil
	}
	return 0, fmt.Errorf("unexpected aid format")
}

// GetCurrentSubtitleTrack returns the current subtitle track ID
func GetCurrentSubtitleTrack(socketPath string) (int, error) {
	result, err := mpvSendCommand(socketPath, []any{"get_property", "sid"})
	if err != nil {
		return 0, err
	}
	if id, ok := result.(float64); ok {
		return int(id), nil
	}
	// "no" means no subtitle is selected
	if _, ok := result.(string); ok {
		return 0, nil
	}
	return 0, fmt.Errorf("unexpected sid format")
}

// upscaleEntry is one row of the upscaling menu.
type upscaleEntry struct{ label, value string }

// upscaleMenuEntries are the rows of the upscaling menu. The modes appear only
// once the Anime4K shaders are installed, and the GAN row offers installing
// them when they are not.
func upscaleMenuEntries(shadersInstalled, ganInstalled bool) []upscaleEntry {
	items := []upscaleEntry{
		{"Back", "back"},
		{"Off (no upscaling)", "off"},
	}
	if shadersInstalled {
		items = append(items,
			upscaleEntry{"Performance (weak GPU)", "performance"},
			upscaleEntry{"Fast (Mode A - text-heavy)", "fast"},
			upscaleEntry{"Balanced (Mode B - general)", "balanced"},
			upscaleEntry{"Quality (Mode C - films)", "quality"},
			upscaleEntry{"Ultra (Max Enhancement - SD sources)", "ultra"},
			upscaleEntry{"--- Advanced Modes ---", "separator"},
			upscaleEntry{"A+A (Max Perceptual - 1080p)", "advanced_aa"},
			upscaleEntry{"B+B (720p Optimized)", "advanced_bb"},
			upscaleEntry{"C+A (Upscaled/Downscaled Content)", "advanced_ca"},
		)
		if ganInstalled {
			items = append(items, upscaleEntry{"GAN UUL (360p to 4K - HEAVY)", "gan_uul"})
		} else {
			items = append(items, upscaleEntry{"GAN UUL (not installed)", "setup_gan"})
		}
	}
	return append(items, upscaleEntry{"Setup shaders (download)", "setup"})
}

// upscaleBreadcrumb names the menu after the current mode, or asks for the
// shaders when there are none.
func upscaleBreadcrumb(shadersInstalled bool, currentMode string) string {
	if !shadersInstalled {
		return "Upscaling (install shaders first)"
	}
	return tui.SingleLine(fmt.Sprintf("Upscaling [%s]", currentMode))
}

// upscaleModes maps a menu value to the shader mode it selects and the
// message that announces it.
var upscaleModes = map[string]struct {
	mode upscaler.ShaderMode
	info string
}{
	"off":         {upscaler.ShaderModeOff, "Real-time upscaling disabled"},
	"performance": {upscaler.ShaderModePerformance, "Real-time upscaling: Performance mode (minimal shaders)"},
	"fast":        {upscaler.ShaderModeFast, "Real-time upscaling: Fast mode (Mode A - good for subtitled anime)"},
	"balanced":    {upscaler.ShaderModeBalanced, "Real-time upscaling: Balanced mode (Mode B - general purpose)"},
	"quality":     {upscaler.ShaderModeQuality, "Real-time upscaling: Quality mode (Mode C - best for films)"},
	"ultra":       {upscaler.ShaderModeUltra, "Real-time upscaling: Ultra mode (Maximum enhancement for SD sources)"},
	"advanced_aa": {upscaler.ShaderModeAdvancedAA, "Real-time upscaling: Advanced A+A (highest perceptual quality, may cause ringing)"},
	"advanced_bb": {upscaler.ShaderModeAdvancedBB, "Real-time upscaling: Advanced B+B (optimized for 720p with aliasing)"},
	"advanced_ca": {upscaler.ShaderModeAdvancedCA, "Real-time upscaling: Advanced C+A (quality + restore for downscaled content)"},
	"gan_uul":     {upscaler.ShaderModeGAN_UUL, "Real-time upscaling: GAN UUL mode (360p→4K - requires powerful GPU!)"},
}

// applyUpscaleChoice carries out one pick of the upscaling menu. again is true
// when the menu should be shown once more (the separator row was chosen).
func applyUpscaleChoice(choice string) (again bool, err error) {
	switch choice {
	case "back":
		return false, nil
	case "separator":
		return true, nil
	case "setup_gan":
		infoFn("Setting up experimental GAN UUL shaders...")
		if err := installGANShadersFn(); err != nil {
			return false, fmt.Errorf("failed to install GAN shaders: %w", err)
		}
		infoFn("GAN UUL shaders installed! Select 'GAN UUL' to enable 360p→4K upscaling.")
		return false, nil
	case "setup":
		infoFn("Setting up Anime4K shaders...")
		if err := installShadersFn(); err != nil {
			return false, fmt.Errorf("failed to install shaders: %w", err)
		}
		infoFn("Anime4K shaders installed! Select a mode to enable upscaling.")
		return false, nil
	}
	if m, ok := upscaleModes[choice]; ok {
		setShaderModeFn(m.mode)
		infoFn(m.info)
		if choice == "gan_uul" {
			warnFn("This mode is VERY heavy! If you experience lag, switch to a lighter mode.")
		}
	}
	return false, nil
}

// handleUpscaleFromMenu shows the real-time upscaling options menu
func handleUpscaleFromMenu() error {
	if !stdinIsTerminal() {
		return fmt.Errorf("no TTY available for interactive menu")
	}
	for {
		shadersInstalled := shadersInstalledFn()
		items := upscaleMenuEntries(shadersInstalled, ganInstalledFn())
		labels := make([]string, len(items))
		for i, it := range items {
			labels[i] = it.label
		}
		idx, err := pickLabelsFn(labels, tui.PickOptions{
			Breadcrumb:   upscaleBreadcrumb(shadersInstalled, shaderModeNameFn()),
			WindowTitle:  "GoAnime - Upscale",
			ItemSingular: "mode",
			ItemPlural:   "modes",
		})
		if err != nil {
			if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
				return nil
			}
			return fmt.Errorf("cancelled: %w", err)
		}
		if idx < 0 || idx >= len(items) {
			return nil
		}
		again, err := applyUpscaleChoice(items[idx].value)
		if err != nil || !again {
			return err
		}
	}
}

// printDownloadLocation prints the absolute path of the downloaded file so
// the user knows where the file was saved (works on macOS, Linux, Windows).
func printDownloadLocation(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	util.PrintSavedLocation("File saved at:", absPath)
}

// subtitleHTTPClientFn builds the SSRF-guarded client subtitle tracks are
// fetched with; a seam for tests.
var subtitleHTTPClientFn = func() *http.Client {
	return &http.Client{
		Transport: api.SafeTransport(30 * time.Second),
		Timeout:   60 * time.Second,
	}
}

// downloadSubtitleFiles fetches the given subtitle tracks and muxes them into
// the downloaded video file. Single downloads pass util.GlobalSubtitles
// (already filtered by util.SelectSubtitles); a batch passes each episode's
// own tracks, captured when that episode's stream was resolved.
func downloadSubtitleFiles(videoPath string, subs []util.SubtitleInfo, printFn func(format string, a ...any)) {
	if printFn == nil {
		printFn = func(format string, a ...any) {
			fmt.Printf(format, a...)
		}
	}
	if len(subs) == 0 {
		return
	}

	// ffmpeg is required for muxing subtitles into the video.
	ffmpegPath, err := resolveFFmpeg()
	if err != nil {
		util.Debugf("%v", err)
		if errors.Is(err, errFFmpegMissing) {
			printFn("Warning: ffmpeg not found — cannot embed subtitles\n")
		} else if errors.Is(err, errFFmpegUnresolvable) {
			printFn("Warning: failed to resolve ffmpeg path\n")
		}
		return
	}

	entries := fetchSubtitleFiles(subtitleHTTPClientFn(), filepath.Dir(videoPath), subs)
	if len(entries) == 0 {
		return
	}

	printFn("Embedding %d subtitle(s) into video...\n", len(entries))
	run := func(args []string) error { return runFFmpegFn(ffmpegPath, args) }
	switch muxSubtitles(run, videoPath, entries) {
	case muxedMP4:
		printFn("Subtitles embedded successfully!\n")
	case muxedMKV:
		printFn("Note: saved as .mkv for better subtitle compatibility\n")
		printFn("Subtitles embedded successfully!\n")
	default:
		util.Debugf("Could not embed subtitles — both MP4 and MKV muxing failed")
		printFn("Warning: Could not embed subtitles — both MP4 and MKV muxing failed\n")
	}

	// Clean up temp subtitle files
	for _, e := range entries {
		_ = os.Remove(e.tmpPath)
	}
}

var (
	errFFmpegMissing      = errors.New("ffmpeg not found — cannot embed subtitles into the video file")
	errFFmpegUnresolvable = errors.New("failed to resolve the ffmpeg path")
	errFFmpegUnsafe       = errors.New("ffmpeg is not a regular absolute file — refusing to execute")
)

// Seams for the subtitle muxing; tests replace them so no ffmpeg is looked up
// or run.
var (
	lookPathFn     = exec.LookPath
	evalSymlinksFn = filepath.EvalSymlinks
	statFn         = os.Stat
	runFFmpegFn    = func(path string, args []string) error {
		util.Debugf("ffmpeg mux cmd: %s %v", path, args)
		cmd := exec.Command(path, args...) // #nosec G204 -- path is validated by resolveFFmpeg: resolved via EvalSymlinks, confirmed absolute and a regular file
		var stderr bytes.Buffer
		cmd.Stdout = nil
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			util.Debugf("ffmpeg mux failed: %v\nstderr: %s", err, stderr.String())
			return err
		}
		return nil
	}
)

// resolveFFmpeg finds ffmpeg and checks it before anything runs it: symlinks
// are resolved, and the result must be an absolute path to a regular file, so
// a poisoned PATH entry cannot turn into an arbitrary command.
func resolveFFmpeg() (string, error) {
	path, err := lookPathFn("ffmpeg")
	if err != nil {
		return "", errFFmpegMissing
	}
	path, err = evalSymlinksFn(path)
	if err != nil {
		return "", errFFmpegUnresolvable
	}
	if !filepath.IsAbs(path) {
		return "", errFFmpegUnsafe
	}
	if fi, statErr := statFn(path); statErr != nil || fi.IsDir() {
		return "", errFFmpegUnsafe
	}
	return path, nil
}

// subtitleEntry is a subtitle track downloaded next to the video, waiting to
// be muxed into it.
type subtitleEntry struct {
	tmpPath  string
	label    string
	langCode string
}

// subtitleExt is the file extension a subtitle URL points at; WebVTT unless it
// names SRT or ASS.
func subtitleExt(subURL string) string {
	lower := strings.ToLower(subURL)
	switch {
	case strings.Contains(lower, ".srt"):
		return "srt"
	case strings.Contains(lower, ".ass"):
		return "ass"
	default:
		return "vtt"
	}
}

// subtitleLang names a track for a file name: its label, else its language,
// else "unknown".
func subtitleLang(label, language string) string {
	if lang := util.SanitizeForFilename(label); lang != "" {
		return lang
	}
	if lang := util.SanitizeForFilename(language); lang != "" {
		return lang
	}
	return "unknown"
}

// fetchSubtitleFiles downloads each track with a URL into dir and returns the
// ones that arrived; a track that fails is skipped, not fatal.
func fetchSubtitleFiles(client *http.Client, dir string, subs []util.SubtitleInfo) []subtitleEntry {
	var entries []subtitleEntry
	for _, sub := range subs {
		if sub.URL == "" {
			continue
		}
		lang := subtitleLang(sub.Label, sub.Language)
		tmpPath := filepath.Join(dir, fmt.Sprintf(".tmp_sub_%s.%s", lang, subtitleExt(sub.URL)))
		req, reqErr := http.NewRequest("GET", sub.URL, http.NoBody)
		if reqErr != nil {
			util.Warnf("Failed to create subtitle request (%s): %v", sub.Label, reqErr)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, respErr := client.Do(req) // #nosec G107 G704
		if respErr != nil {
			util.Warnf("Failed to download subtitle (%s): %v", sub.Label, respErr)
			continue
		}
		func() {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				util.Warnf("Subtitle download failed (%s): HTTP %d", sub.Label, resp.StatusCode)
				return
			}
			out, oErr := os.Create(filepath.Clean(tmpPath))
			if oErr != nil {
				util.Warnf("Failed to create temp subtitle file (%s): %v", sub.Label, oErr)
				return
			}
			defer func() { _ = out.Close() }()
			if _, cpErr := io.Copy(out, resp.Body); cpErr != nil {
				util.Warnf("Failed to write subtitle (%s): %v", sub.Label, cpErr)
				return
			}
			langCode := sub.Language
			if langCode == "" {
				langCode = lang
			}
			entries = append(entries, subtitleEntry{tmpPath: tmpPath, label: sub.Label, langCode: langCode})
		}()
	}
	return entries
}

// subtitleMuxArgs is the ffmpeg command line that copies the video's video and
// audio streams, adds each subtitle track with subCodec and writes outPath.
// Only video and audio are mapped from the input: data streams such as
// timed_id3, present in the MPEG-TS of HLS downloads, crash MP4/MKV muxing.
func subtitleMuxArgs(videoPath string, entries []subtitleEntry, subCodec, outPath string) []string {
	a := []string{"-y", "-fflags", "+genpts", "-i", filepath.Clean(videoPath)}
	for _, e := range entries {
		a = append(a, "-i", filepath.Clean(e.tmpPath))
	}
	a = append(a, "-map", "0:v", "-map", "0:a")
	for i := range entries {
		a = append(a, "-map", fmt.Sprintf("%d", i+1))
	}
	a = append(a, "-c:v", "copy", "-c:a", "copy", "-c:s", subCodec)
	for i, e := range entries {
		a = append(a,
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("language=%s", e.langCode),
			fmt.Sprintf("-metadata:s:s:%d", i), fmt.Sprintf("title=%s", e.label),
		)
	}
	return append(a, filepath.Clean(outPath))
}

// muxResult is how embedding the subtitles ended.
type muxResult int

const (
	muxFailed muxResult = iota
	muxedMP4            // replaced the video in place
	muxedMKV            // saved next to it as .mkv
)

// muxSubtitles embeds the subtitles: first as an MP4 with mov_text, replacing
// the video; if that fails, as an MKV, which tolerates more subtitle formats
// and TS inputs. A failed attempt leaves no partial output behind.
func muxSubtitles(run func(args []string) error, videoPath string, entries []subtitleEntry) muxResult {
	attempt := func(subCodec, outPath string) error {
		args := subtitleMuxArgs(videoPath, entries, subCodec, outPath)
		if err := run(args); err != nil {
			_ = os.Remove(outPath)
			return err
		}
		return nil
	}

	tmpMP4 := videoPath + ".muxing.mp4"
	if err := attempt("mov_text", tmpMP4); err == nil {
		renErr := os.Rename(tmpMP4, videoPath)
		if renErr == nil {
			return muxedMP4
		}
		util.Warnf("Failed to replace video: %v", renErr)
		_ = os.Remove(tmpMP4)
	} else {
		util.Debugf("MP4 mux failed: %v — trying MKV fallback", err)
	}

	mkvPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".mkv"
	tmpMKV := mkvPath + ".tmp.mkv"
	if err := attempt("srt", tmpMKV); err == nil {
		renErr := os.Rename(tmpMKV, mkvPath)
		if renErr == nil {
			return muxedMKV
		}
		util.Warnf("Failed to save MKV: %v", renErr)
		_ = os.Remove(tmpMKV)
	}
	return muxFailed
}
