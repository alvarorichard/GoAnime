package player

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/allanime"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/upscaler"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// ErrUserQuit is returned when the user chooses to quit the application
var ErrUserQuit = errors.New("user requested to quit application")

// ErrChangeAnime is returned when the user chooses to change anime
var ErrChangeAnime = errors.New("user requested to change anime")

// ErrBackToDownloadOptions is returned when user wants to go back to download options
var ErrBackToDownloadOptions = errors.New("back to download options")

// errStayInPlayerMenu signals that a menu action was a no-op (playlist
// boundary hit or back from the episode picker). The in-player menu loop must
// keep running: returning nil here would unwind handleUserInput and surface
// the post-playback menu while mpv is still playing the current episode.
var errStayInPlayerMenu = errors.New("stay in player menu")

// dubSubTagRe strips parenthesized dub/sub tags from anime names for display
var dubSubTagRe = regexp.MustCompile(`\s*\((?i:Dublado|Legendado|SUB|DUB|Subbed|Dubbed)\)\s*`)

const defaultHLSReferer = "https://streameeeeee.site/"

func appendPlaybackRefererArgs(mpvArgs []string, videoURL string, isHLSStream bool) ([]string, string) {
	lowerURL := strings.ToLower(strings.TrimSpace(videoURL))
	if !strings.HasPrefix(lowerURL, "http://") && !strings.HasPrefix(lowerURL, "https://") {
		return mpvArgs, ""
	}

	referer := util.GetGlobalReferer()
	if referer == "" && isHLSStream {
		referer = defaultHLSReferer
	}
	if referer == "" {
		return mpvArgs, ""
	}

	return append(mpvArgs, fmt.Sprintf("--http-header-fields=Referer: %s", referer)), referer
}

// hlsAllowAllExtensionsArg relaxes ffmpeg's HLS segment-extension allowlist so
// mpv fetches segments served with disguised extensions (.js/.css/.woff, …).
// Some providers (e.g. SuperFlix's FirePlayer host) obfuscate their HLS
// segments this way AND serve the separate alternative-audio rendition's
// segments with the same disguised extensions. mpv's builtin ffmpeg only
// relaxes the check for the main variant (it auto-sets extension_picky=0), so
// the audio rendition stays blocked by the stricter `allowed_extensions`
// allowlist — the stream then plays video with NO audio. Forcing
// allowed_extensions=ALL lets the audio rendition load too. It only widens a
// check, never restricts, so it is safe for well-behaved HLS sources as well.
const hlsAllowAllExtensionsArg = "--demuxer-lavf-o=allowed_extensions=ALL"

// appendHLSDemuxerArgs adds HLS-specific demuxer options to mpvArgs when the
// stream is HLS, and returns mpvArgs unchanged otherwise.
func appendHLSDemuxerArgs(mpvArgs []string, isHLSStream bool) []string {
	if !isHLSStream {
		return mpvArgs
	}
	return append(mpvArgs, hlsAllowAllExtensionsArg)
}

// playbackArgsInput captures every decision that determines the mpv argument
// list. playVideo gathers these (some via interactive prompts, globals, and
// tracking side effects) and hands them to buildPlaybackArgs, so the actual
// argument assembly — including the HLS referer + allowed_extensions wiring that
// makes SuperFlix audio play — is deterministic and unit-testable without
// launching mpv.
type playbackArgsInput struct {
	VideoURL         string
	IsHLS            bool
	Is9Anime         bool
	UpscalingEnabled bool
	ShaderArgs       []string
	Wayland          bool
	CanImpersonate   bool
	Title            string // "" → no --force-media-title
	IsMovieOrTV      bool
	AudioLang        string
	SubsLang         string
	SubArgs          []string
	ResumeTime       int
}

// buildPlaybackArgs assembles the full mpv argument list from resolved inputs.
// It is pure (no I/O, no globals, no prompts) so the exact argument set — and in
// particular that HLS streams carry BOTH the Referer header and the
// allowed_extensions=ALL demuxer flag — can be pinned by tests. Keep it in sync
// with StartVideo's own base args; this covers only the per-playback args.
func buildPlaybackArgs(in playbackArgsInput) []string {
	mpvArgs := []string{
		"--cache=yes",
		"--demuxer-max-bytes=300M",
		"--demuxer-readahead-secs=20",
		"--audio-display=no",
	}

	if in.UpscalingEnabled {
		// Shader-based upscaling: gpu-next (libplacebo) + no hw decode so shaders
		// can process frames.
		mpvArgs = append(mpvArgs, "--vo=gpu-next", "--hwdec=no")
		mpvArgs = append(mpvArgs, in.ShaderArgs...)
	} else {
		mpvArgs = append(mpvArgs,
			"--no-config",
			"--hwdec=auto-safe",
			"--vo=gpu",
			"--profile=fast",
			"--video-latency-hacks=yes",
		)
	}

	if in.Wayland {
		mpvArgs = append(mpvArgs, "--gpu-context=wayland")
	}

	mpvArgs, playbackReferer := appendPlaybackRefererArgs(mpvArgs, in.VideoURL, in.IsHLS)
	// Relax the HLS segment-extension allowlist so alternative-audio renditions
	// with disguised segment extensions load (fixes video-plays-but-no-audio on
	// SuperFlix/FirePlayer streams).
	mpvArgs = appendHLSDemuxerArgs(mpvArgs, in.IsHLS)

	// For 9Anime (Cloudflare-protected CDNs), route playback through yt-dlp with
	// Chrome TLS impersonation so ffmpeg's TLS fingerprint is not rejected.
	if in.IsHLS && in.Is9Anime {
		referer := playbackReferer
		if referer == "" {
			referer = defaultHLSReferer
		}
		mpvArgs = append(mpvArgs,
			"--script-opts=ytdl_hook-try_ytdl_first=yes",
			fmt.Sprintf("--ytdl-raw-options-append=referer=%s", referer),
		)
		if in.CanImpersonate {
			mpvArgs = append(mpvArgs, "--ytdl-raw-options-append=impersonate=chrome")
		}
	}

	// For googlevideo.com URLs served through our local Blogger proxy, disable
	// yt-dlp so mpv fetches from 127.0.0.1 directly.
	if strings.Contains(in.VideoURL, "127.0.0.1") && strings.Contains(in.VideoURL, "blogger_proxy") {
		mpvArgs = append(mpvArgs, "--ytdl=no")
	}

	if in.Title != "" {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--force-media-title=%s", in.Title))
	}

	// Audio/subtitle language preferences only for movies/TV (FlixHQ etc.).
	if in.IsMovieOrTV {
		mpvArgs = append(mpvArgs,
			fmt.Sprintf("--alang=%s", in.AudioLang),
			fmt.Sprintf("--slang=%s", in.SubsLang),
		)
	}

	// External subtitle files (already gated + resolved by the caller).
	mpvArgs = append(mpvArgs, in.SubArgs...)

	// HLS resume is handled by seeking after start (--start is unreliable on HLS).
	if in.ResumeTime > 0 && !in.IsHLS {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", in.ResumeTime))
	}

	return mpvArgs
}

// waitForVideoReady waits for the HLS video to be ready for playback
// Returns true if video is ready, false if timeout
func waitForVideoReady(socketPath string) bool {
	util.Debugf("Waiting for HLS video to be ready...")

	maxWait := 45 * time.Second // Generous for slow HLS streams
	pollInterval := 50 * time.Millisecond
	maxPollInterval := 200 * time.Millisecond
	startTime := time.Now()

	// Alternate between two reliable properties each iteration
	// to reduce IPC overhead while covering both cases
	check := 0
	for time.Since(startTime) < maxWait {
		switch check % 2 {
		case 0:
			// Check duration (most reliable for HLS)
			if resp, err := mpvSendCommand(socketPath, []any{"get_property", "duration"}); err == nil {
				if d, ok := resp.(float64); ok && d > 0 {
					util.Debugf("HLS video ready (duration: %.0fs) after %.1fs", d, time.Since(startTime).Seconds())
					return true
				}
			}
		case 1:
			// Check time-pos (video is actually playing)
			if resp, err := mpvSendCommand(socketPath, []any{"get_property", "time-pos"}); err == nil {
				if pos, ok := resp.(float64); ok && pos >= 0 {
					util.Debugf("HLS video playing (pos: %.1fs) after %.1fs", pos, time.Since(startTime).Seconds())
					return true
				}
			}
		}
		check++

		time.Sleep(pollInterval)
		// Grow poll interval gradually to reduce IPC chatter
		if pollInterval < maxPollInterval {
			pollInterval = min(pollInterval*13/10, maxPollInterval)
		}
	}
	util.Debugf("Timeout waiting for HLS video (%.1fs)", time.Since(startTime).Seconds())
	return false
}

// seekToResumePosition seeks to the saved resume position for HLS streams
// This function is robust: it waits for video ready, seeks, and verifies the seek worked
func seekToResumePosition(socketPath string, resumeTime int) {
	if resumeTime <= 0 {
		return
	}

	util.Debugf("HLS resume: will seek to %d seconds", resumeTime)

	// Wait for video to be fully ready first
	if !waitForVideoReady(socketPath) {
		util.Debugf("HLS resume: video not ready after timeout, attempting seek anyway")
	}

	// Short stabilize pause after video is ready
	time.Sleep(100 * time.Millisecond)

	// Try seek with verification — 3 attempts is sufficient for HLS
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		util.Debugf("HLS resume: seek attempt %d/%d to %d seconds", attempt, maxAttempts, resumeTime)

		// Method 1: seek absolute (works best with HLS)
		_, err := mpvSendCommand(socketPath, []any{"seek", float64(resumeTime), "absolute"})
		if err != nil {
			util.Debugf("HLS resume: seek absolute failed: %v, trying set_property", err)
			// Method 2: set_property time-pos
			_, _ = mpvSendCommand(socketPath, []any{"set_property", "time-pos", float64(resumeTime)})
		}

		// Wait for seek to settle
		time.Sleep(time.Duration(200*attempt) * time.Millisecond)

		// Verify position
		posResp, posErr := mpvSendCommand(socketPath, []any{"get_property", "time-pos"})
		if posErr == nil {
			if pos, ok := posResp.(float64); ok {
				// Allow 10 second tolerance for HLS streams
				if pos >= float64(resumeTime-10) {
					util.Debugf("HLS resume: SUCCESS - position is %.0f seconds (target: %d)", pos, resumeTime)
					return
				}
				util.Debugf("HLS resume: position mismatch - got %.0f, want %d", pos, resumeTime)
			}
		} else {
			util.Debugf("HLS resume: could not get position: %v", posErr)
		}

		// Brief wait before retry
		if attempt < maxAttempts {
			time.Sleep(300 * time.Millisecond)
		}
	}

	util.Debugf("HLS resume: FAILED after %d attempts - video may start from beginning", maxAttempts)
}

// applySkipTimes applies skip times to an mpv instance
func applySkipTimes(socketPath string, episode *models.Episode) {
	var opts []string
	if episode.SkipTimes.Op.Start > 0 || episode.SkipTimes.Op.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_op=%d-%d", episode.SkipTimes.Op.Start, episode.SkipTimes.Op.End))
	}
	if episode.SkipTimes.Ed.Start > 0 || episode.SkipTimes.Ed.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_ed=%d-%d", episode.SkipTimes.Ed.Start, episode.SkipTimes.Ed.End))
	}

	if len(opts) > 0 {
		combinedOpts := strings.Join(opts, ",")
		_, cmdErr := mpvSendCommand(socketPath, []any{"set_property", "script-opts", combinedOpts})
		if cmdErr != nil {
			util.Debugf("Failed to apply skip times: %v. Command: set_property script-opts %s", cmdErr, combinedOpts)
		} else {
			util.Debugf("Skip times applied successfully: %s", combinedOpts)
		}
	} else {
		util.Debugf("No skip times available for episode %s", episode.Number)
	}
}

// showResumeDialog displays a compact dialog asking if user wants to resume playback
func showResumeDialog(episodeNum int, timeSeconds int) (bool, error) {
	var resume bool

	// Convert seconds to minutes and seconds for better readability
	minutes := timeSeconds / 60
	seconds := timeSeconds % 60

	var timeStr string
	if minutes > 0 {
		timeStr = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		timeStr = fmt.Sprintf("%ds", seconds)
	}

	confirm := huh.NewConfirm().
		Title(fmt.Sprintf("Resume episode %d from %s?", episodeNum, timeStr)).
		Description("You can continue watching from where you left off.").
		Affirmative("Yes, resume").
		Negative("No, start from beginning").
		Value(&resume)

	if err := tui.RunClean(confirm.Run); err != nil {
		return false, fmt.Errorf("error showing dialog: %w", err)
	}

	return resume, nil
}

// playVideo plays the video and manages interactions
// func playVideo(
// 	videoURL string,
// 	episodes []models.Episode,
// 	currentEpisodeNum int,
// 	anilistID int,
// 	updater *discord.RichPresenceUpdater,
// ) error {
// 	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)
// 	util.Debugf("Video URL: %s", videoURL)

// 	currentEpisode, err := getCurrentEpisode(episodes, currentEpisodeNum)
// 	if err != nil {
// 		return fmt.Errorf("error getting current episode: %w", err)
// 	}

// 	mpvArgs := []string{
// 		"--hwdec=auto-safe",
// 		"--vo=gpu",
// 		"--profile=fast",
// 		"--cache=yes",
// 		"--demuxer-max-bytes=300M",
// 		"--demuxer-readahead-secs=20",
// 		"--no-config",
// 		"--video-latency-hacks=yes",
// 		"--audio-display=no",
// 	}

// 	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)
// 	if resumeTime > 0 {
// 		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
// 	}

// 	skipDataChan := fetchAniSkipAsync(anilistID, currentEpisodeNum, currentEpisode)
// 	socketPath, err := StartVideo(videoURL, mpvArgs)
// 	if err != nil {
// 		return fmt.Errorf("failed to start video: %w", err)
// 	}

// 	applyAniSkipResults(skipDataChan, socketPath, currentEpisode, currentEpisodeNum)

// 	if updater != nil {
// 		initDiscordPresence(updater, socketPath, tracker, anilistID, currentEpisode, currentEpisodeNum)
// 		defer updater.Stop()
// 	}

// 	currentEpisodeIndex := findEpisodeIndex(episodes, currentEpisodeNum)
// 	if currentEpisodeIndex == -1 {
// 		return fmt.Errorf("episode %d not found in list", currentEpisodeNum)
// 	}

// 	preloadNextEpisode(episodes, currentEpisodeIndex)

// 	stopTracking := startTrackingRoutine(tracker, socketPath, anilistID, currentEpisode, currentEpisodeNum, updater)
// 	defer close(stopTracking)

// 	return handleUserInput(
// 		socketPath,
// 		episodes,
// 		currentEpisodeIndex,
// 		currentEpisodeNum,
// 		anilistID,
// 		updater,
// 		stopTracking,
// 		currentEpisode,
// 	)
// }

// playVideo plays the video and manages interactions
// playVideo plays the video and manages interactions
//
// malID is the MyAnimeList ID (used by the AniSkip API).
// anilistID is the AniList ID (used as the tracking/resume key).
// They are distinct identifiers — passing one in place of the other
// causes resume data to collide across sources.
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	malID int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
) error {
	timer := util.StartTimer("playVideo:Total")
	defer timer.Stop()
	defer StopBloggerProxy() // Clean up any Blogger video proxy
	util.PerfCount("video_plays")

	// Log the episode number and URL for debugging
	util.Debugf("Playing video for episode %d, URL: %s", currentEpisodeNum, videoURL)

	// Resolve Blogger video URLs to actual video streams
	if strings.Contains(videoURL, "blogger.com/video.g") {
		util.Debugf("Resolving Blogger video URL before playback...")
		resolved, err := extractBloggerGoogleVideoURL(videoURL)
		if err == nil && resolved != videoURL {
			util.Debugf("Blogger URL resolved to: %s", resolved)
			videoURL = resolved
		}
	}

	// Normalize video URL if necessary
	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)

	// Get the current episode
	currentEpisode, err := getCurrentEpisode(episodes, currentEpisodeNum)
	if err != nil {
		return fmt.Errorf("error getting current episode: %w", err)
	}

	// Check if real-time upscaling is enabled
	shaderArgs := upscaler.GetMPVShaderArgs(upscaler.CurrentShaderMode)
	upscalingEnabled := len(shaderArgs) > 0

	// --- Gather every decision that drives the mpv argument list ------------
	// The heavy lifting (interactive prompts, globals, tracking) happens here;
	// the deterministic assembly is delegated to buildPlaybackArgs so the exact
	// arg set is unit-testable without launching mpv.

	isHLSStream := LooksLikeHLS(videoURL)

	// Determine the anime source for source-specific playback configuration.
	// Use the globally-stored anime source (set during stream resolution) so
	// that 9Anime is detected reliably even when Discord (and hence the
	// updater) is disabled.
	is9Anime := util.Is9AnimeSource()
	if !is9Anime && updater != nil && updater.GetAnime() != nil {
		is9Anime = updater.GetAnime().Source == "9Anime"
	}

	// MPV window title: clean anime name + season/episode (or just name for movies).
	titleSnap := snapshotMedia()
	title := ""
	if titleSnap.AnimeName != "" {
		// Display sanitizer, NOT the filename one: a window title may keep
		// colons etc. ("Need for Speed: O Filme" must not lose its ':').
		cleanName := util.SanitizeForDisplayTitle(titleSnap.AnimeName)
		// Also strip parenthesized dub/sub tags like (Dublado), (Legendado), (SUB)
		cleanName = dubSubTagRe.ReplaceAllString(cleanName, " ")
		cleanName = strings.TrimSpace(cleanName)
		if titleSnap.MediaType == "movie" {
			title = cleanName
		} else {
			title = fmt.Sprintf("%s S%02dE%02d", cleanName, titleSnap.AnimeSeason, currentEpisodeNum)
		}
	}

	// Audio/subtitle language preferences apply to movies/TV (FlixHQ) and to
	// SuperFlix.
	//
	// SuperFlix must be matched by SOURCE, not by media type: its streams are
	// multi-audio HLS with an external Portuguese subtitle track regardless of
	// whether the entry is a movie, a series, an anime or a dorama. Gating on
	// IsMovieOrTV alone silently threw away both the audio track the user picked
	// and the subtitles for every SuperFlix anime.
	isSuperFlix := util.IsSuperFlixSource()
	isMovieOrTV := false
	if updater != nil && updater.GetAnime() != nil {
		anime := updater.GetAnime()
		isMovieOrTV = anime.IsMovieOrTV() || strings.Contains(strings.ToLower(anime.Source), "flixhq")
		isSuperFlix = isSuperFlix || strings.EqualFold(anime.Source, "SuperFlix")
		// Update exact media type for download path organization.
		if anime.MediaType != "" && titleSnap.MediaType == "" {
			SetExactMediaType(string(anime.MediaType))
		}
	}
	wantsLangPrefs := isMovieOrTV || isSuperFlix

	audioLang, subsLang := "", ""
	if wantsLangPrefs {
		audioLang = util.GlobalAudioLanguage
		if audioLang == "" {
			// Default: prefer Portuguese (Brazil), Portuguese, Spanish, English.
			audioLang = "pt-BR,pt,por,pb,ptbr,portuguese,spa,es,spanish,eng,en,english"
		}
		subsLang = util.GlobalSubsLanguage
		if subsLang == "" {
			subsLang = "pt-BR,pt,por,pb,ptbr,portuguese,spa,es,spanish,eng,en,english"
		}
		util.Debugf("Applying language preferences: audio=%s, subs=%s", audioLang, subsLang)
	}

	// External subtitle files (FlixHQ / SuperFlix / 9Anime). For 9Anime, ALWAYS
	// prompt the user to pick a subtitle language after every episode selection.
	var subArgs []string
	if wantsLangPrefs || is9Anime {
		if is9Anime {
			util.PromptSubtitleLanguage()
		} else if len(util.GlobalSubtitles) > 1 {
			util.SelectSubtitles()
		}
		subArgs = util.GetSubtitleArgs()
		if len(subArgs) > 0 {
			util.Debugf("Added external subtitles: %v", subArgs)
		}
	}

	// YtdlpCanImpersonate is only consulted on the 9Anime HLS path; keep the call
	// gated so non-9Anime playback doesn't pay for it.
	canImpersonate := false
	if isHLSStream && is9Anime {
		canImpersonate = util.YtdlpCanImpersonate()
	}

	wayland := runtime.GOOS == "linux" && os.Getenv("WAYLAND_DISPLAY") != ""
	if upscalingEnabled {
		util.Infof("Real-time Anime4K upscaling enabled: %s", upscaler.GetShaderModeName(upscaler.CurrentShaderMode))
		util.Debugf("Shader args: %v", shaderArgs)
	}
	if wayland {
		util.Debugf("Wayland session detected — forcing gpu-context=wayland")
	}

	// Initialize tracking and check for resume time (may show a resume dialog).
	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)

	mpvArgs := buildPlaybackArgs(playbackArgsInput{
		VideoURL:         videoURL,
		IsHLS:            isHLSStream,
		Is9Anime:         is9Anime,
		UpscalingEnabled: upscalingEnabled,
		ShaderArgs:       shaderArgs,
		Wayland:          wayland,
		CanImpersonate:   canImpersonate,
		Title:            title,
		IsMovieOrTV:      isMovieOrTV,
		AudioLang:        audioLang,
		SubsLang:         subsLang,
		SubArgs:          subArgs,
		ResumeTime:       resumeTime,
	})

	// Fetch AniSkip data asynchronously (AniSkip API is keyed on MAL ID)
	skipDataChan := fetchAniSkipAsync(malID, currentEpisodeNum, currentEpisode)

	// Start the video with mpv
	mpvTimer := util.StartTimer("MPV:StartVideo")
	socketPath, err := StartVideo(videoURL, mpvArgs)
	mpvTimer.Stop()
	if err != nil {
		return fmt.Errorf("failed to start video: %w", err)
	}

	// For HLS streams, seek to resume position (includes waiting for video ready)
	if isHLSStream && resumeTime > 0 {
		util.Debugf("HLS stream detected with resume time %d seconds", resumeTime)
		hlsTimer := util.StartTimer("HLS:SeekToResume")
		seekToResumePosition(socketPath, resumeTime)
		hlsTimer.Stop()
	} else if isHLSStream {
		// Just wait for video ready without seeking
		hlsTimer := util.StartTimer("HLS:WaitForReady")
		waitForVideoReady(socketPath)
		hlsTimer.Stop()
	}

	// Apply AniSkip results to skip intros/outros
	applyAniSkipResults(skipDataChan, socketPath, currentEpisode, currentEpisodeNum)

	// Show OSD message if upscaling is enabled
	if upscalingEnabled {
		showShaderOSD(socketPath)
	}

	// Initialize Discord Rich Presence if updater is provided
	if updater != nil {
		initDiscordPresence(updater, socketPath, tracker, anilistID, currentEpisode, currentEpisodeNum)
		defer updater.Stop()
	}

	// Find the current episode index in the episode list
	currentEpisodeIndex := findEpisodeIndex(episodes, currentEpisodeNum)
	if currentEpisodeIndex == -1 {
		return fmt.Errorf("episode %d not found in list", currentEpisodeNum)
	}

	// Preload the next episode for seamless playback
	preloadNextEpisode(episodes, currentEpisodeIndex)

	// Start tracking routine if tracker is available
	stopTracking := startTrackingRoutine(tracker, socketPath, anilistID, currentEpisode, currentEpisodeNum, updater)

	// Handle user input for interactive controls
	err = handleUserInput(
		socketPath,
		episodes,
		currentEpisodeIndex,
		currentEpisodeNum,
		malID,
		anilistID,
		updater,
		stopTracking,
		currentEpisode,
	)

	// Close the tracking channel if it's still open
	select {
	case <-stopTracking:
		// Channel already closed
	default:
		close(stopTracking)
	}

	return err
}

// getCurrentEpisode gets the current episode
// func getCurrentEpisode(episodes []models.Episode, num int) (*models.Episode, error) {
// 	if num < 1 || num > len(episodes) {
// 		return nil, fmt.Errorf("invalid episode number: %d", num)
// 	}
// 	return &episodes[num-1], nil
// }

// getCurrentEpisode retrieves the current episode based on the episode number
func getCurrentEpisode(episodes []models.Episode, num int) (*models.Episode, error) {
	numStr := fmt.Sprintf("%d", num)

	// First try: match by extracted episode number
	for i := range episodes {
		if ExtractEpisodeNumber(episodes[i].Number) == numStr {
			return &episodes[i], nil
		}
	}

	// Second try: match by Num field directly
	for i := range episodes {
		if episodes[i].Num == num {
			return &episodes[i], nil
		}
	}

	// Third try: match by Number field directly (for simple numeric cases)
	for i := range episodes {
		if episodes[i].Number == numStr {
			return &episodes[i], nil
		}
	}

	// Fourth try: check if we're within the bounds and use index-based access
	// This handles cases where episode numbering doesn't match exactly
	if num > 0 && num <= len(episodes) {
		return &episodes[num-1], nil
	}

	return nil, fmt.Errorf("episode %d not found", num)
}

// // initTracking initializes the tracking system
// func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
// 	if !tracking.IsCgoEnabled {
// 		if util.IsDebug {
// 			util.Debug("Tracking disabled: CGO not available")
// 		}
// 		return nil, 0
// 	}

// 	currentUser, err := user.Current()
// 	if err != nil {
// 		util.Errorf("Failed to get current user: %v", err)
// 		return nil, 0
// 	}

// 	var dbPath string
// 	if runtime.GOOS == "windows" {
// 		dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
// 	} else {
// 		dbPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
// 	}

// 	tracker := tracking.NewLocalTracker(dbPath)
// 	if tracker == nil {
// 		return nil, 0
// 	}

// 	progress, err := tracker.GetAnime(anilistID, episode.URL)
// 	if err != nil || progress == nil || progress.PlaybackTime <= 0 {
// 		return tracker, 0
// 	}

// 	// Always use the selected episodeNum for the dialog, but use the tracked PlaybackTime if available
// 	if ok, _ := showResumeDialog(episodeNum, progress.PlaybackTime); ok {
// 		util.Debugf("Resuming from saved time: %d seconds for episode %d", progress.PlaybackTime, episodeNum)
// 		return tracker, progress.PlaybackTime
// 	}

// 	return tracker, 0
// }

// cachedDBPath stores the database path to avoid repeated user.Current() calls
var cachedDBPath string

// getTrackerDBPath returns the cached database path
func getTrackerDBPath() string {
	if cachedDBPath != "" {
		return cachedDBPath
	}

	currentUser, err := user.Current()
	if err != nil {
		util.Errorf("Failed to get current user: %v", err)
		return ""
	}

	if runtime.GOOS == "windows" {
		cachedDBPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
	} else {
		cachedDBPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
	}

	return cachedDBPath
}

// trackingKey builds an episode-specific tracking key so that each episode
// gets its own row in the database. Without this, sources like AllAnime
// (where episode.URL is the anime ID, not the episode ID) would share a
// single tracking row across all episodes, causing stale resume times.
func trackingKey(episodeURL string, episodeNum int) string {
	return fmt.Sprintf("%s:ep%d", episodeURL, episodeNum)
}

// initTracking inicializa o sistema de rastreamento
func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
	if !tracking.IsCgoEnabled {
		if util.IsDebug {
			util.Debug("Tracking disabled: CGO not available")
		}
		return nil, 0
	}

	dbPath := getTrackerDBPath()
	if dbPath == "" {
		return nil, 0
	}

	// First check if we have a cached tracker (fast path)
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		// Need to initialize - this is only slow the first time
		tracker = tracking.NewLocalTracker(dbPath)
		if tracker == nil {
			return nil, 0
		}
	}

	key := trackingKey(episode.URL, episodeNum)
	util.Debugf("Tracking lookup: anilistID=%d, key=%s", anilistID, key)

	// Try episode-specific key first
	progress, err := tracker.GetAnime(anilistID, key)
	if err != nil {
		util.Debugf("Tracking lookup error: %v", err)
		return tracker, 0
	}

	// Fallback: try legacy key (without episode number) for backward compatibility
	if progress == nil {
		progress, err = tracker.GetAnime(anilistID, episode.URL)
		if err != nil {
			util.Debugf("Tracking legacy lookup error: %v", err)
			return tracker, 0
		}
	}

	if progress == nil {
		util.Debugf("Tracking lookup: no progress found")
		return tracker, 0
	}
	if progress.PlaybackTime <= 0 {
		util.Debugf("Tracking lookup: playback time is 0 or negative")
		return tracker, 0
	}

	// Safety check: verify the saved progress belongs to the current episode.
	// This prevents offering a resume from episode 3's position when the user
	// has switched to episode 4 (e.g., AllAnime shares the same URL across episodes).
	if progress.EpisodeNumber != episodeNum {
		util.Debugf("Tracking: saved progress is for episode %d, current is %d - skipping resume",
			progress.EpisodeNumber, episodeNum)
		return tracker, 0
	}

	util.Debugf("Tracking found: PlaybackTime=%d seconds for episode %d", progress.PlaybackTime, episodeNum)

	if ok, _ := showResumeDialog(episodeNum, progress.PlaybackTime); ok {
		util.Debugf("Resuming from saved time: %d seconds for episode %d", progress.PlaybackTime, episodeNum)
		return tracker, progress.PlaybackTime
	}

	util.Debugf("User declined to resume")
	return tracker, 0
}

// InitTrackerAsync initializes the tracker in the background.
// Call this early in the application lifecycle to avoid delays later.
func InitTrackerAsync() {
	if !tracking.IsCgoEnabled {
		return
	}

	go func() {
		dbPath := getTrackerDBPath()
		if dbPath != "" {
			tracking.NewLocalTracker(dbPath)
		}
	}()
}

// aniSkipFetcher is the function used to fetch AniSkip skip-times data.
// It is exposed as a package-level variable so tests can swap it for a spy
// to verify which ID (MAL vs AniList) is being routed to the AniSkip API
// without making real network calls.
var aniSkipFetcher = api.GetAndParseAniSkipData

// fetchAniSkipAsync fetches AniSkip data in parallel.
// The AniSkip API is keyed by MyAnimeList (MAL) ID, not AniList ID.
func fetchAniSkipAsync(malID, episodeNum int, episode *models.Episode) chan error {
	ch := make(chan error, 1)
	go func() {
		err := aniSkipFetcher(malID, episodeNum, episode)
		ch <- err
	}()
	return ch
}

// showShaderOSD displays an OSD message confirming shaders are active
func showShaderOSD(socketPath string) {
	// Run in background to avoid blocking the main playback flow
	go func() {
		// Give mpv a moment to initialize its OSD subsystem
		time.Sleep(300 * time.Millisecond)

		modeName := upscaler.GetShaderModeName(upscaler.CurrentShaderMode)
		message := fmt.Sprintf("Anime4K Upscaling: %s\\nPress Shift+I twice for stats", modeName)

		// Show OSD message for 4 seconds
		_, err := mpvSendCommand(socketPath, []any{
			"show-text", message, 4000,
		})
		if err != nil {
			util.Debugf("Failed to show shader OSD: %v", err)
		}
	}()
}

// applyAniSkipResults applies AniSkip results asynchronously so MPV is not blocked.
// Skip times are applied as soon as the AniSkip data arrives (or after 2s timeout).
func applyAniSkipResults(ch chan error, socketPath string, episode *models.Episode, episodeNum int) {
	go func() {
		select {
		case err := <-ch:
			if err == nil {
				applySkipTimes(socketPath, episode)

				// For AllAnime episodes, also try to set chapter markers (like Curd does)
				if strings.Contains(episode.URL, "kibfyvtiFpKC") || len(episode.URL) < 30 {
					allAnimeClient := allanime.NewAllAnimeClient()
					if chapterErr := allAnimeClient.SendSkipTimesToMPV(episode, socketPath, MpvSendCommand); chapterErr != nil {
						util.Debugf("Failed to set chapter markers: %v", chapterErr)
					}
				}
			} else {
				util.Debugf("AniSkip data unavailable for episode %d: %v", episodeNum, err)
			}
		case <-time.After(2 * time.Second):
			util.Debugf("Timeout fetching AniSkip data for episode %d", episodeNum)
		}
	}()
}

// initDiscordPresence initializes Discord presence
func initDiscordPresence(updater *discord.RichPresenceUpdater, socketPath string, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	updater.SetSocketPath(socketPath)
	updater.Start()

	go func() {
		waitForPlaybackStart(socketPath, updater)
		updateEpisodeDuration(socketPath, updater, tracker, anilistID, episode, episodeNum)
	}()
}

// waitForPlaybackStart waits for playback to start
func waitForPlaybackStart(socketPath string, updater *discord.RichPresenceUpdater) {
	maxAttempts := 30 // Max 30 seconds to wait
	for range maxAttempts {
		timePos, err := mpvSendCommand(socketPath, []any{"get_property", "time-pos"})
		if err == nil && timePos != nil && !updater.IsEpisodeStarted() {
			updater.SetEpisodeStarted(true)
			return
		}
		time.Sleep(1 * time.Second)
	}
	// Set as started anyway to avoid infinite loop
	updater.SetEpisodeStarted(true)
}

// updateEpisodeDuration updates the episode duration
func updateEpisodeDuration(socketPath string, updater *discord.RichPresenceUpdater, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	maxAttempts := 10 // Only try 10 times to get duration
	for range maxAttempts {
		if !updater.IsEpisodeStarted() {
			time.Sleep(2 * time.Second)
			continue
		}

		// If duration is already set, just update tracking and exit
		if updater.GetEpisodeDuration() > 0 {
			dur := updater.GetEpisodeDuration()
			updateTrackingWithDuration(tracker, anilistID, episode, episodeNum, dur)
			return
		}

		durationPos, err := mpvSendCommand(socketPath, []any{"get_property", "duration"})
		if err != nil || durationPos == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		duration, ok := durationPos.(float64)
		if !ok || duration <= 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		dur := time.Duration(duration * float64(time.Second))
		if dur < time.Second {
			dur = 24 * time.Minute
		}

		updater.SetEpisodeDuration(dur)
		updateTrackingWithDuration(tracker, anilistID, episode, episodeNum, dur)
		return
	}
}

// updateTrackingWithDuration updates the local tracker with episode info and duration
func updateTrackingWithDuration(tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int, dur time.Duration) {
	if tracker != nil && dur > 0 {
		anime := tracking.Anime{
			AnilistID:     anilistID,
			AllanimeID:    episode.URL,
			EpisodeNumber: episodeNum,
			Duration:      int(dur.Seconds()),
			Title:         getEpisodeTitle(episode.Title),
			LastUpdated:   time.Now(),
		}
		if err := tracker.UpdateProgress(anime); err != nil {
			util.Errorf("Failed to update tracking: %v", err)
		}
	}
}

// getEpisodeTitle gets the episode title
func getEpisodeTitle(title models.TitleDetails) string {
	if title.English != "" {
		return title.English
	}
	if title.Romaji != "" {
		return title.Romaji
	}
	if title.Japanese != "" {
		return title.Japanese
	}
	return "No title"
}

// findEpisodeIndex finds the episode index
// func findEpisodeIndex(episodes []models.Episode, num int) int {
// 	episodeStr := strconv.Itoa(num)
// 	for i, ep := range episodes {
// 		if ExtractEpisodeNumber(ep.Number) == episodeStr {
// 			return i
// 		}
// 	}
// 	return -1
// }

// findEpisodeIndex finds the episode index based on the episode number
func findEpisodeIndex(episodes []models.Episode, num int) int {
	episodeStr := fmt.Sprintf("%d", num)

	// First try: match by extracted episode number
	for i, ep := range episodes {
		extractedNum := ExtractEpisodeNumber(ep.Number)
		if extractedNum == episodeStr {
			return i
		}
	}

	// Second try: match by Num field directly
	for i, ep := range episodes {
		if ep.Num == num {
			return i
		}
	}

	// Third try: match by Number field directly
	for i, ep := range episodes {
		if ep.Number == episodeStr {
			return i
		}
	}

	// Fourth try: use index-based access if within bounds
	if num > 0 && num <= len(episodes) {
		return num - 1
	}

	return -1 // Episode not found
}

// preloadNextEpisode preloads the next episode
func preloadNextEpisode(episodes []models.Episode, currentIndex int) {
	if currentIndex+1 >= len(episodes) {
		return
	}

	// Skip preloading for AllAnime episodes (they use IDs, not HTTP URLs)
	nextEpisodeURL := episodes[currentIndex+1].URL
	if len(nextEpisodeURL) < 30 && !strings.Contains(nextEpisodeURL, "http") {
		return
	}

	// NOTE: We intentionally do NOT call GetVideoURLForEpisode here because it
	// may invoke extractActualVideoURL which contains an interactive (huh.Select)
	// quality picker. Running that in a background goroutine while the foreground
	// showPlayerMenu is also using huh.Select causes terminal corruption and
	// "user aborted" errors.  Preloading is a best-effort optimisation; if we
	// can't do it safely we just skip it.
	go func() {
		// Only extract the intermediate video source URL (no interactive prompt).
		url, err := extractVideoURL(nextEpisodeURL)
		if err == nil && url != "" {
			util.Debugf("Preloaded next episode source URL: %s", url)
		}
	}()
}

// startTrackingRoutine starts the tracking routine
// Optimized with adaptive update interval to reduce overhead
func startTrackingRoutine(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) chan struct{} {
	stopChan := make(chan struct{})
	if tracker == nil {
		return stopChan
	}

	go func() {
		// Start with faster updates, then slow down for efficiency
		ticker := time.NewTicker(3 * time.Second) // Increased from 2s to reduce overhead
		defer ticker.Stop()

		updateCount := 0
		for {
			select {
			case <-ticker.C:
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
				updateCount++
				// After 30 updates (~90 seconds), slow down to every 5 seconds
				if updateCount == 30 {
					ticker.Reset(5 * time.Second)
				}
			case <-stopChan:
				// Final update before stopping
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
				return
			}
		}
	}()

	return stopChan
}

// updateTracking updates tracking
func updateTracking(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) {
	timePos, err := mpvSendCommand(socketPath, []any{"get_property", "time-pos"})
	if err != nil || timePos == nil {
		return
	}

	position, ok := timePos.(float64)
	if !ok {
		return
	}

	duration := 1440 // Default duration in seconds (24 minutes)
	if updater != nil {
		episodeDur := updater.GetEpisodeDuration()
		if episodeDur > 0 {
			duration = int(episodeDur.Seconds())
		}
	}

	// Ensure duration is valid before updating tracking
	if duration <= 0 {
		duration = 1440 // Fallback to default
	}

	anime := tracking.Anime{
		AnilistID:     anilistID,
		AllanimeID:    trackingKey(episode.URL, episodeNum),
		EpisodeNumber: episodeNum,
		PlaybackTime:  int(position),
		Duration:      duration,
		Title:         getEpisodeTitle(episode.Title),
		LastUpdated:   time.Now(),
	}

	if err := tracker.UpdateProgress(anime); err != nil {
		util.Errorf("Error updating tracking: %v", err)
	}
}

// showPlayerMenu displays an interactive menu for player controls
func showPlayerMenu(animeName string, currentEpisodeNum int) (string, error) {

	// Build title and options based on media type
	var title string

	type menuOption struct {
		Label string
		Value string
	}
	var menuItems []menuOption

	isMovie := IsCurrentMediaMovie()

	if isMovie {
		// Movie: show movie name without episode number
		title = "GoAnime Player Controls"
		if animeName != "" {
			title = fmt.Sprintf("Now playing: %s", animeName)
		}
		menuItems = []menuOption{
			{"← Back", "download_options"},
			{"Replay movie", "next"},
			{"Change movie", "change"},
			{"Exit", "quit"},
		}
	} else {
		// TV series / anime: show episode navigation
		title = "GoAnime Player Controls"
		if animeName != "" {
			title = fmt.Sprintf("Now playing: %s - Episode %d", animeName, currentEpisodeNum)
		}
		menuItems = []menuOption{
			{"← Back", "download_options"},
			{"Next episode", "next"},
			{"Previous episode", "previous"},
			{"Select episode", "select"},
			{"Change anime", "change"},
			{"Skip intro", "skip"},
			{"Exit", "quit"},
		}
	}

	_ = title // title is informational for the prompt string
	idx, err := tui.Find(menuItems, func(i int) string {
		return menuItems[i].Label
	}, fuzzyfinder.WithPromptString(title+": "))

	if err != nil {
		return "", fmt.Errorf("error showing menu: %w", err)
	}

	return menuItems[idx].Value, nil
}

// handleUserInput manages user input
func handleUserInput(
	socketPath string,
	episodes []models.Episode,
	currentIndex int,
	currentEpisodeNum int,
	malID int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
	stopTracking chan struct{},
	currentEpisode *models.Episode,
) error {
	// Get anime name for display
	var animeName string
	if updater != nil && updater.GetAnime() != nil {
		animeName = updater.GetAnime().Name
	}

	for {
		// Check if mpv is still running before showing the menu.
		// If mpv exited (e.g., due to a bad URL), quit gracefully instead of
		// showing a menu that the user cannot meaningfully interact with.
		if _, pingErr := mpvSendCommand(socketPath, []any{"get_property", "pid"}); pingErr != nil {
			util.Debugf("mpv process appears to have exited, returning to caller")
			return ErrBackToDownloadOptions
		}

		choice, err := showPlayerMenu(animeName, currentEpisodeNum)
		if err != nil {
			// If the menu was disrupted (e.g., by a concurrent terminal writer),
			// treat it as "go back" rather than a fatal error.
			util.Debugf("Player menu interrupted: %v", err)
			_, _ = mpvSendCommand(socketPath, []any{"quit"})
			return ErrBackToDownloadOptions
		}

		switch choice {
		case "download_options":
			// Stop playback and go back to download options
			_, _ = mpvSendCommand(socketPath, []any{"quit"})
			return ErrBackToDownloadOptions
		case "next":
			if err := playNextEpisode(currentIndex+1, episodes, malID, anilistID, updater, stopTracking, socketPath); !errors.Is(err, errStayInPlayerMenu) {
				return err
			}
		case "previous":
			if err := playPreviousEpisode(currentIndex-1, episodes, malID, anilistID, updater, stopTracking, socketPath); !errors.Is(err, errStayInPlayerMenu) {
				return err
			}
		case "quit":
			_, _ = mpvSendCommand(socketPath, []any{"quit"})
			return ErrUserQuit
		case "change":
			_, _ = mpvSendCommand(socketPath, []any{"quit"})
			return ErrChangeAnime
		case "select":
			if err := selectEpisode(episodes, malID, anilistID, updater, stopTracking, socketPath); !errors.Is(err, errStayInPlayerMenu) {
				return err
			}
		case "audio":
			selectAudioTrack(socketPath)
		case "subtitle":
			selectSubtitleTrack(socketPath)
		case "skip":
			skipIntro(socketPath, currentEpisode)
		}
	}
}

// playNextEpisode plays next episode
func playNextEpisode(newIndex int, episodes []models.Episode, malID, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex >= len(episodes) {
		fmt.Println("You are on the last episode")
		return errStayInPlayerMenu
	}
	return switchEpisode(newIndex, episodes, malID, anilistID, updater, stopTracking, socketPath)
}

// playPreviousEpisode plays previous episode
func playPreviousEpisode(newIndex int, episodes []models.Episode, malID, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex < 0 {
		fmt.Println("You are on the first episode")
		return errStayInPlayerMenu
	}
	return switchEpisode(newIndex, episodes, malID, anilistID, updater, stopTracking, socketPath)
}

// selectEpisode allows selecting an episode
func selectEpisode(episodes []models.Episode, malID, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	selectedURL, selectedNumStr, err := SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		// User selected back: keep the player menu open, mpv untouched
		if errors.Is(err, ErrBackRequested) {
			return errStayInPlayerMenu
		}
		return fmt.Errorf("failed to select episode: %w", err)
	}

	if idx := findSelectedEpisodeIndex(episodes, selectedURL, selectedNumStr); idx >= 0 {
		return switchEpisode(idx, episodes, malID, anilistID, updater, stopTracking, socketPath)
	}

	return fmt.Errorf("episode %s not found", selectedNumStr)
}

// findSelectedEpisodeIndex resolves a fuzzy-finder selection to a slice index.
// The episode number is matched before the URL because sources like AllAnime
// use the anime ID as the URL of every episode, so a URL-only match always
// resolved to the first episode regardless of what the user picked.
func findSelectedEpisodeIndex(episodes []models.Episode, url, numStr string) int {
	for i := range episodes {
		if episodes[i].Number == numStr && episodes[i].URL == url {
			return i
		}
	}
	for i := range episodes {
		if episodes[i].Number == numStr {
			return i
		}
	}
	for i := range episodes {
		if episodes[i].URL == url {
			return i
		}
	}
	return -1
}

// switchEpisode switches between episodes
func switchEpisode(newIndex int, episodes []models.Episode, malID, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	target := episodes[newIndex]
	targetNum, err := strconv.Atoi(ExtractEpisodeNumber(target.Number))
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	var anime *models.Anime
	if updater != nil {
		anime = updater.GetAnime()
	}

	// If no updater/anime context, try to synthesize from lastAnimeURL
	storedURL := getLastAnimeURL()
	if anime == nil && storedURL != "" {
		guessedSource := ""
		// Check global anime source first (most reliable, set during stream resolution)
		if src := util.GetGlobalAnimeSource(); src != "" {
			guessedSource = src
		} else if ref := util.GetGlobalReferer(); strings.Contains(ref, "rapid-cloud") {
			// Check global referer to detect 9Anime (uses rapid-cloud referer)
			guessedSource = "9Anime"
		} else if (len(storedURL) < 30 && !strings.Contains(storedURL, "http")) || strings.Contains(storedURL, "allanime") {
			guessedSource = "AllAnime"
		}
		anime = &models.Anime{URL: storedURL, Source: guessedSource}
	}

	// switchEpisode runs inside the mpv key-event loop, which has no context
	// today; Background is the sanctioned placeholder until that loop is
	// context-aware.
	targetURL, err := GetVideoURLForEpisodeEnhanced(context.Background(), &target, anime)
	if err != nil {
		return fmt.Errorf("failed to get video URL: %w", err)
	}

	if updater != nil {
		updater.Stop()
	}

	close(stopTracking)
	_, _ = mpvSendCommand(socketPath, []any{"quit"})

	var newUpdater *discord.RichPresenceUpdater
	if updater != nil {
		duration := time.Duration(target.Duration) * time.Second
		newUpdater = discord.NewRichPresenceUpdater(
			updater.GetAnime(),
			updater.GetIsPaused(),
			updater.GetAnimeMutex(),
			updater.GetUpdateFreq(),
			duration,
			"",
			MpvSendCommand,
		)
		newUpdater.SetEpisodeStarted(false)
	}

	return playVideo(targetURL, episodes, targetNum, malID, anilistID, newUpdater)
}

// skipIntro skips the intro
func skipIntro(socketPath string, episode *models.Episode) {
	if episode.SkipTimes.Op.End > 0 {
		_, _ = mpvSendCommand(socketPath, []any{"seek", episode.SkipTimes.Op.End, "absolute"})
		fmt.Printf("Intro skipped to %ds\n", episode.SkipTimes.Op.End)
	} else {
		fmt.Println("Intro skip data not available")
	}
}

// selectAudioTrack shows a menu to select audio track
func selectAudioTrack(socketPath string) {
	tracks, err := GetAudioTracks(socketPath)
	if err != nil {
		fmt.Printf("Error getting audio tracks: %v\n", err)
		return
	}

	if len(tracks) == 0 {
		fmt.Println("No audio tracks available")
		return
	}

	currentID, _ := GetCurrentAudioTrack(socketPath)

	type trackOption struct {
		Label string
		ID    int
	}
	var trackItems []trackOption
	for _, track := range tracks {
		id := int(track["id"].(float64))

		// Build label with all available info
		var parts []string

		// Language
		lang := ""
		if l, ok := track["lang"].(string); ok && l != "" {
			lang = l
		}

		// Title (often contains language info)
		title := ""
		if t, ok := track["title"].(string); ok && t != "" {
			title = t
		}

		// Codec
		codec := ""
		if c, ok := track["codec"].(string); ok && c != "" {
			codec = c
		}

		// Channels (stereo, 5.1, etc.)
		channels := ""
		if ch, ok := track["demux-channels"].(string); ok && ch != "" {
			channels = ch
		} else if ch, ok := track["audio-channels"].(float64); ok {
			channels = fmt.Sprintf("%.0fch", ch)
		}

		// Build the label
		switch {
		case title != "":
			parts = append(parts, title)
		case lang != "":
			parts = append(parts, lang)
		default:
			parts = append(parts, fmt.Sprintf("Audio %d", id))
		}

		if codec != "" {
			parts = append(parts, codec)
		}
		if channels != "" {
			parts = append(parts, channels)
		}

		label := fmt.Sprintf("Track %d: %s", id, strings.Join(parts, " | "))
		if id == currentID {
			label = "* " + label + " (current)"
		}
		trackItems = append(trackItems, trackOption{Label: label, ID: id})
	}

	idx, err := tui.Find(trackItems, func(i int) string {
		return trackItems[i].Label
	}, fuzzyfinder.WithPromptString("Select Audio Track: "))
	if err != nil {
		return
	}

	selected := trackItems[idx].ID
	if err := SetAudioTrack(socketPath, selected); err != nil {
		fmt.Printf("Error setting audio track: %v\n", err)
	} else {
		fmt.Printf("Audio track changed to %d\n", selected)
	}
}

// selectSubtitleTrack shows a menu to select subtitle track
func selectSubtitleTrack(socketPath string) {
	tracks, err := GetSubtitleTracks(socketPath)
	if err != nil {
		fmt.Printf("Error getting subtitle tracks: %v\n", err)
		return
	}

	currentID, _ := GetCurrentSubtitleTrack(socketPath)

	type trackOption struct {
		Label string
		ID    int
	}
	var trackItems []trackOption
	// Add option to disable subtitles
	disableLabel := "Disable subtitles"
	if currentID == 0 {
		disableLabel = "* " + disableLabel + " (current)"
	}
	trackItems = append(trackItems, trackOption{Label: disableLabel, ID: 0})

	for _, track := range tracks {
		id := int(track["id"].(float64))

		// Build label with all available info
		var parts []string

		// Language
		lang := ""
		if l, ok := track["lang"].(string); ok && l != "" {
			lang = l
		}

		// Title (often contains language info)
		title := ""
		if t, ok := track["title"].(string); ok && t != "" {
			title = t
		}

		// Build the label
		switch {
		case title != "":
			parts = append(parts, title)
		case lang != "":
			parts = append(parts, lang)
		default:
			parts = append(parts, fmt.Sprintf("Subtitle %d", id))
		}

		label := fmt.Sprintf("Track %d: %s", id, strings.Join(parts, " | "))
		if id == currentID {
			label = "* " + label + " (current)"
		}
		trackItems = append(trackItems, trackOption{Label: label, ID: id})
	}

	idx, err := tui.Find(trackItems, func(i int) string {
		return trackItems[i].Label
	}, fuzzyfinder.WithPromptString("Select Subtitle Track: "))
	if err != nil {
		return
	}

	selected := trackItems[idx].ID
	if selected == 0 {
		// Disable subtitles
		_, _ = mpvSendCommand(socketPath, []any{"set_property", "sid", "no"})
		fmt.Println("Subtitles disabled")
	} else {
		if err := SetSubtitleTrack(socketPath, selected); err != nil {
			fmt.Printf("Error setting subtitle track: %v\n", err)
		} else {
			fmt.Printf("Subtitle track changed to %d\n", selected)
		}
	}
}
