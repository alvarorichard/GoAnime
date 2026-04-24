package util

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/version"
	"github.com/ktr0731/go-fuzzyfinder"
)

// SubtitleInfo represents a single subtitle track
type SubtitleInfo struct {
	URL      string
	Language string
	Label    string
}

var (
	// IsDebug enables verbose debug logging when set to true.
	IsDebug       bool
	minNameLength = 4
	// ErrHelpRequested is returned when the user requests help or version info instead of running the program.
	ErrHelpRequested = errors.New("help requested")
	// GlobalSource is the selected anime source, set via --source flag (e.g. "allanime", "animefire").
	GlobalSource string
	// GlobalQuality is the selected video quality, set via --quality flag.
	GlobalQuality string
	// GlobalMediaType is the media type filter, set via --type flag (anime, movie, tv).
	GlobalMediaType string
	// GlobalSubsLanguage is the preferred subtitle language, set via --subs flag.
	GlobalSubsLanguage string
	// GlobalAudioLanguage is the preferred audio language, set via --audio flag.
	GlobalAudioLanguage string
	// GlobalSubtitles holds the subtitle tracks for the current playback session.
	GlobalSubtitles []SubtitleInfo
	// GlobalNoSubs disables subtitles when set to true via --no-subs flag.
	GlobalNoSubs bool
	// GlobalReferer is the HTTP Referer header sent with stream requests.
	GlobalReferer string
	// GlobalOutputDir is the custom download output directory, set via -o flag.
	GlobalOutputDir string
	// GlobalAnimeSource is the current anime scraper name (e.g. "9Anime", "AllAnime").
	GlobalAnimeSource string
)

// SessionConfig captures the user-facing runtime preferences parsed from flags.
type SessionConfig struct {
	Source        string
	Quality       string
	MediaType     string
	SubsLanguage  string
	AudioLanguage string
	NoSubs        bool
	OutputDir     string
}

var runtimeStateMu sync.RWMutex

// CurrentSessionConfig returns a snapshot of the current session preferences.
func CurrentSessionConfig() SessionConfig {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()

	return SessionConfig{
		Source:        GlobalSource,
		Quality:       GlobalQuality,
		MediaType:     GlobalMediaType,
		SubsLanguage:  GlobalSubsLanguage,
		AudioLanguage: GlobalAudioLanguage,
		NoSubs:        GlobalNoSubs,
		OutputDir:     GlobalOutputDir,
	}
}

// SetGlobalSource sets the preferred media source (e.g. "allanime", "animefire").
func SetGlobalSource(source string) {
	runtimeStateMu.Lock()
	GlobalSource = source
	runtimeStateMu.Unlock()
}

// GetGlobalSource returns the currently configured media source preference.
func GetGlobalSource() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalSource
}

// SetGlobalQuality sets the preferred stream quality (e.g. "1080", "720", "best").
func SetGlobalQuality(quality string) {
	runtimeStateMu.Lock()
	GlobalQuality = quality
	runtimeStateMu.Unlock()
}

// GetGlobalQuality returns the currently configured stream quality preference.
func GetGlobalQuality() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalQuality
}

// SetGlobalMediaType sets the media type filter (e.g. "movie", "tv").
func SetGlobalMediaType(mediaType string) {
	runtimeStateMu.Lock()
	GlobalMediaType = mediaType
	runtimeStateMu.Unlock()
}

// SetPreferredSubtitleLanguage sets the preferred subtitle language (e.g. "english", "portuguese").
func SetPreferredSubtitleLanguage(language string) {
	runtimeStateMu.Lock()
	GlobalSubsLanguage = language
	runtimeStateMu.Unlock()
}

// GetPreferredSubtitleLanguage returns the currently configured subtitle language.
func GetPreferredSubtitleLanguage() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalSubsLanguage
}

// SetPreferredAudioLanguage sets the preferred audio track language.
func SetPreferredAudioLanguage(language string) {
	runtimeStateMu.Lock()
	GlobalAudioLanguage = language
	runtimeStateMu.Unlock()
}

// GetPreferredAudioLanguage returns the currently configured audio language.
func GetPreferredAudioLanguage() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalAudioLanguage
}

// SetSubtitlesDisabled enables or disables subtitle loading globally.
func SetSubtitlesDisabled(disabled bool) {
	runtimeStateMu.Lock()
	GlobalNoSubs = disabled
	runtimeStateMu.Unlock()
}

// SubtitlesDisabled reports whether subtitle loading has been disabled by the user.
func SubtitlesDisabled() bool {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalNoSubs
}

// SetGlobalOutputDir sets the directory where downloaded files will be saved.
func SetGlobalOutputDir(outputDir string) {
	runtimeStateMu.Lock()
	GlobalOutputDir = outputDir
	runtimeStateMu.Unlock()
}

// GetGlobalOutputDir returns the configured output directory for downloads.
func GetGlobalOutputDir() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalOutputDir
}

func cloneDownloadRequest(req *DownloadRequest) *DownloadRequest {
	if req == nil {
		return nil
	}
	cloned := *req
	return &cloned
}

// SetGlobalDownloadRequest stores the current download request snapshot.
func SetGlobalDownloadRequest(req *DownloadRequest) {
	runtimeStateMu.Lock()
	GlobalDownloadRequest = cloneDownloadRequest(req)
	runtimeStateMu.Unlock()
}

// CurrentDownloadRequest returns a copy of the current download request.
func CurrentDownloadRequest() *DownloadRequest {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return cloneDownloadRequest(GlobalDownloadRequest)
}

// ClearGlobalDownloadRequest resets the stored download request to nil.
func ClearGlobalDownloadRequest() {
	SetGlobalDownloadRequest(nil)
}

func cloneUpscaleRequest(req *UpscaleRequest) *UpscaleRequest {
	if req == nil {
		return nil
	}
	cloned := *req
	return &cloned
}

// SetGlobalUpscaleRequest stores the current upscale request snapshot.
func SetGlobalUpscaleRequest(req *UpscaleRequest) {
	runtimeStateMu.Lock()
	GlobalUpscaleRequest = cloneUpscaleRequest(req)
	runtimeStateMu.Unlock()
}

// CurrentUpscaleRequest returns a copy of the current upscale request.
func CurrentUpscaleRequest() *UpscaleRequest {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return cloneUpscaleRequest(GlobalUpscaleRequest)
}

// ClearGlobalUpscaleRequest resets the stored upscale request to nil.
func ClearGlobalUpscaleRequest() {
	SetGlobalUpscaleRequest(nil)
}

// SetGlobalSubtitles stores subtitles for the current playback session
func SetGlobalSubtitles(subs []SubtitleInfo) {
	playbackStateMu.Lock()
	GlobalSubtitles = cloneSubtitleInfos(subs)
	stored := cloneSubtitleInfos(GlobalSubtitles)
	playbackStateMu.Unlock()
	if len(stored) > 0 {
		Debugf("Stored %d subtitle track(s) for playback", len(stored))
		for i, sub := range stored {
			Debugf("  Subtitle %d: %s (%s)", i+1, sub.Label, sub.Language)
		}
	}
}

// ClearGlobalSubtitles clears stored subtitles
func ClearGlobalSubtitles() {
	playbackStateMu.Lock()
	GlobalSubtitles = nil
	playbackStateMu.Unlock()
}

// SetGlobalReferer stores the referer for stream requests
func SetGlobalReferer(referer string) {
	playbackStateMu.Lock()
	GlobalReferer = referer
	playbackStateMu.Unlock()
	if referer != "" {
		Debugf("Stored referer for stream requests: %s", referer)
	}
}

// GetGlobalReferer returns the stored referer
func GetGlobalReferer() string {
	playbackStateMu.RLock()
	defer playbackStateMu.RUnlock()
	return GlobalReferer
}

// ClearGlobalReferer clears the stored referer
func ClearGlobalReferer() {
	playbackStateMu.Lock()
	GlobalReferer = ""
	playbackStateMu.Unlock()
}

// SetGlobalAnimeSource stores the current anime source (e.g. "9Anime", "AllAnime")
func SetGlobalAnimeSource(source string) {
	playbackStateMu.Lock()
	GlobalAnimeSource = source
	playbackStateMu.Unlock()
	if source != "" {
		Debugf("Stored anime source: %s", source)
	}
}

// GetGlobalAnimeSource returns the stored anime source
func GetGlobalAnimeSource() string {
	playbackStateMu.RLock()
	defer playbackStateMu.RUnlock()
	return GlobalAnimeSource
}

// Is9AnimeSource returns true if the current stream is from 9Anime
func Is9AnimeSource() bool {
	return GetGlobalAnimeSource() == "9Anime"
}

// subtitleOption maps a display label to a sentinel value for subtitle selection.
type subtitleOption struct {
	Label string
	Value int
}

// SelectSubtitles displays an interactive menu for the user to choose which
// subtitle tracks to load. It updates GlobalSubtitles in-place so that the
// subsequent call to GetSubtitleArgs only includes the selected tracks.
// If there are 0 or 1 subtitles available, no menu is shown.
func SelectSubtitles() {
	tracks := GetGlobalSubtitles()
	if SubtitlesDisabled() || len(tracks) <= 1 {
		return
	}

	// Build options: "All", each individual track, "None"
	var items []subtitleOption
	items = append(items, subtitleOption{"All subtitles", -1})
	for i, sub := range tracks {
		label := sub.Label
		if label == "" {
			label = sub.Language
		}
		if label == "" {
			label = fmt.Sprintf("Subtitle %d", i+1)
		}
		items = append(items, subtitleOption{label, i})
	}
	items = append(items, subtitleOption{"No subtitles", -2})

	idx, err := tui.Find(items, func(i int) string {
		return items[i].Label
	}, fuzzyfinder.WithPromptString("Subtitles: "))
	if err != nil {
		// On error/cancel keep all subtitles
		return
	}

	selected := items[idx].Value
	switch selected {
	case -1:
		// Keep all — no change needed
		Debugf("User selected all %d subtitle track(s)", len(tracks))
	case -2:
		// Disable subtitles
		ClearGlobalSubtitles()
		Debugf("User disabled subtitles")
	default:
		if selected >= 0 && selected < len(tracks) {
			kept := tracks[selected]
			SetGlobalSubtitles([]SubtitleInfo{kept})
			Debugf("User selected subtitle: %s (%s)", kept.Label, kept.Language)
		}
	}
}

// PromptSubtitleLanguage always prompts the user to select a subtitle language
// for multi-language platforms (e.g., 9Anime). This MUST be called after every
// episode selection, without exception. Unlike SelectSubtitles, this function
// always shows the prompt regardless of the number of available tracks.
// It updates GlobalSubtitles in-place so that GetSubtitleArgs returns
// the correct arguments for mpv.
func PromptSubtitleLanguage() {
	if SubtitlesDisabled() {
		Debugf("Subtitles disabled by user (--no-subs), skipping subtitle prompt")
		ClearGlobalSubtitles()
		return
	}

	tracks := GetGlobalSubtitles()

	// No tracks available — inform the user and continue without subtitles
	if len(tracks) == 0 {
		fmt.Println("\nNo subtitle tracks available for this episode.")
		return
	}

	// Single track — still ask the user if they want it
	if len(tracks) == 1 {
		label := tracks[0].Label
		if label == "" {
			label = tracks[0].Language
		}
		if label == "" {
			label = "Unknown"
		}

		fmt.Printf("\n1 subtitle track available: %s\n", label)

		items := []subtitleOption{
			{label, 0},
			{"No subtitles", -2},
		}

		idx, err := tui.Find(items, func(i int) string {
			return items[i].Label
		}, fuzzyfinder.WithPromptString("Select subtitle language: "))
		if err != nil {
			// On error/cancel keep the track
			fmt.Printf("Subtitles: %s\n", label)
			return
		}

		if items[idx].Value == -2 {
			ClearGlobalSubtitles()
			fmt.Println("Subtitles: disabled")
			Debugf("User disabled subtitles")
		} else {
			fmt.Printf("Subtitles: %s\n", label)
			Debugf("User selected subtitle: %s", label)
		}
		return
	}

	// Multiple tracks — show full selection menu
	fmt.Printf("\n%d subtitle language(s) available:\n", len(tracks))

	var items []subtitleOption
	items = append(items, subtitleOption{"All subtitles", -1})
	for i, sub := range tracks {
		label := sub.Label
		if label == "" {
			label = sub.Language
		}
		if label == "" {
			label = fmt.Sprintf("Subtitle %d", i+1)
		}
		items = append(items, subtitleOption{label, i})
	}
	items = append(items, subtitleOption{"No subtitles", -2})

	idx, err := tui.Find(items, func(i int) string {
		return items[i].Label
	}, fuzzyfinder.WithPromptString("Select subtitle language: "))
	if err != nil {
		// On error/cancel keep all subtitles
		fmt.Println("Subtitles: all (default)")
		return
	}

	selected := items[idx].Value
	switch selected {
	case -1:
		// Keep all — no change needed
		fmt.Printf("Subtitles: all (%d tracks)\n", len(tracks))
		Debugf("User selected all %d subtitle track(s)", len(tracks))
	case -2:
		// Disable subtitles
		ClearGlobalSubtitles()
		fmt.Println("Subtitles: disabled")
		Debugf("User disabled subtitles")
	default:
		if selected >= 0 && selected < len(tracks) {
			kept := tracks[selected]
			SetGlobalSubtitles([]SubtitleInfo{kept})
			fmt.Printf("Subtitles: %s\n", kept.Label)
			Debugf("User selected subtitle: %s (%s)", kept.Label, kept.Language)
		}
	}
}

// GetSubtitleArgs returns mpv arguments for subtitles
// Based on lobster.sh implementation:
// - Single subtitle: --sub-file='URL'
// - Multiple subtitles: --sub-files='URL1:URL2:...'
func GetSubtitleArgs() []string {
	tracks := GetGlobalSubtitles()
	if SubtitlesDisabled() || len(tracks) == 0 {
		return nil
	}

	// Collect all subtitle URLs
	var urls []string
	for _, sub := range tracks {
		if sub.URL != "" {
			urls = append(urls, sub.URL)
		}
	}

	if len(urls) == 0 {
		return nil
	}

	if len(urls) == 1 {
		// Single subtitle file
		return []string{fmt.Sprintf("--sub-file=%s", urls[0])}
	}

	// Multiple subtitle files - join with appropriate separator
	// Unix uses : as separator, Windows uses ; (following lobster.sh implementation)
	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}
	joined := strings.Join(urls, separator)
	return []string{fmt.Sprintf("--sub-files=%s", joined)}
}

// Cleanup function to be called on program exit
var (
	cleanupFuncs []func()
	cleanupMu    sync.Mutex
)

// RegisterCleanup registers a function to be called on program exit
func RegisterCleanup(fn func()) {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	cleanupFuncs = append(cleanupFuncs, fn)
}

// RunCleanup runs all registered cleanup functions
func RunCleanup() {
	cleanupMu.Lock()
	funcs := make([]func(), len(cleanupFuncs))
	copy(funcs, cleanupFuncs)
	cleanupMu.Unlock()
	for _, fn := range funcs {
		fn()
	}
	// Print performance report if enabled
	if PerfEnabled {
		GetPerfTracker().PrintReport()
	}
}

// ErrorHandler returns a string with the error message, if debug mode is enabled, it will return the full error with details.
func ErrorHandler(err error) string {
	if IsDebug {
		if LogFilePath != "" {
			return fmt.Sprintf("%+v\n\nDebug log saved to: %s", err, LogFilePath)
		}
		return fmt.Sprintf("%+v", err)
	}
	return fmt.Sprintf("%v -- run the program with --debug to see details and save a log file", err)
}

// Helper prints the beautiful help message
func Helper() {
	ShowBeautifulHelp()
}

// Custom error types for different exit conditions
var (
	ErrUpdateRequested        = errors.New("update requested")
	ErrDownloadRequested      = errors.New("download requested")
	ErrUpscaleRequested       = errors.New("upscale requested")
	ErrMovieDownloadRequested = errors.New("movie download requested")
)

// DownloadRequest holds download command parameters
type DownloadRequest struct {
	AnimeName     string
	EpisodeNum    int
	IsRange       bool
	StartEpisode  int
	EndEpisode    int
	Source        string // Added source field for specifying anime source
	Quality       string // Added quality field for video quality
	AllAnimeSmart bool   // Enable AllAnime Smart Range (auto-skip intros/credits and preferred mirrors)
	// Movie/TV specific fields
	IsMovie      bool   // True if downloading a movie from FlixHQ/SFlix
	IsTV         bool   // True if downloading a TV show from FlixHQ/SFlix
	SeasonNum    int    // Season number for TV shows
	SubsLanguage string // Subtitle language preference
	OutputDir    string // Custom output directory for downloads
	// Download-all mode
	IsAll bool // True to download ALL episodes (anime) or ALL seasons+episodes (TV/series/dorama)
}

// UpscaleRequest holds upscale command parameters
type UpscaleRequest struct {
	InputPath        string  // Input video or image file path
	OutputPath       string  // Output file path (optional, defaults to input_upscaled.ext)
	ScaleFactor      int     // Upscale multiplier (default: 2)
	Passes           int     // Number of processing passes (default: 2)
	StrengthColor    float64 // Line thinning strength 0-1 (default: 0.333)
	StrengthGradient float64 // Sharpening strength 0-1 (default: 1.0)
	FastMode         bool    // Use fast mode (lower quality)
	HighQuality      bool    // Use high quality mode (slower)
	PreserveAudio    bool    // Preserve original audio track
	UseGPU           bool    // Use GPU encoding if available
	VideoBitrate     string  // Video bitrate (default: 8M)
	Workers          int     // Number of parallel workers
}

// GlobalDownloadRequest holds the pending download request parsed from CLI flags.
var GlobalDownloadRequest *DownloadRequest

// GlobalUpscaleRequest holds the pending upscale request parsed from CLI flags.
var GlobalUpscaleRequest *UpscaleRequest

// FlagParser parses the -flags and returns the anime name
func FlagParser() (string, error) {
	// Override the default flag.Usage to show our custom help
	flag.Usage = func() {
		Helper()
	}

	// Use a custom FlagSet to avoid conflicts with library flags (e.g., Anime4KGo)
	fs := flag.NewFlagSet("goanime", flag.ContinueOnError)

	// Define flags
	debug := fs.Bool("debug", false, "enable debug mode")
	perf := fs.Bool("perf", false, "enable performance profiling")
	help := fs.Bool("help", false, "show help message")
	altHelp := fs.Bool("h", false, "show help message")
	versionFlag := fs.Bool("version", false, "show version information")
	updateFlag := fs.Bool("update", false, "check for updates and update if available")
	downloadFlag := fs.Bool("d", false, "download mode")
	rangeFlag := fs.Bool("r", false, "download episode range (use with -d)")
	allFlag := fs.Bool("a", false, "download ALL episodes/seasons (use with -d or -dm)")
	movieDownloadFlag := fs.Bool("dm", false, "download movie/TV from FlixHQ/SFlix")
	sourceFlag := fs.String("source", "", "specify anime source (allanime, animefire, ptbr, flixhq)")
	qualityFlag := fs.String("quality", "best", "specify video quality (best, worst, 720p, 1080p, etc.)")
	allanimeSmartFlag := fs.Bool("allanime-smart", false, "enable AllAnime Smart Range: auto-skip intros/outros and use priority mirrors")
	mediaTypeFlag := fs.String("type", "", "specify media type (anime, movie, tv)")
	subsLanguageFlag := fs.String("subs", "english", "specify subtitle language for movies/TV (FlixHQ only)")
	audioLanguageFlag := fs.String("audio", "pt-BR,pt,english", "specify preferred audio language for movies/TV (FlixHQ only)")
	noSubsFlag := fs.Bool("no-subs", false, "disable subtitles for movies/TV (FlixHQ only)")
	outputDirFlag := fs.String("o", "", "output directory for downloads (default: ~/.local/goanime/downloads/anime/)")

	// Upscale flags
	upscaleFlag := fs.Bool("upscale", false, "upscale mode - enhance video/image quality using Anime4K algorithm")
	upscaleOutputFlag := fs.String("upscale-output", "", "output path for upscaled file (default: input_upscaled.ext)")
	upscaleScaleFlag := fs.Int("upscale-scale", 2, "upscale factor (default: 2x)")
	upscalePassesFlag := fs.Int("upscale-passes", 2, "number of processing passes (default: 2)")
	upscaleFastFlag := fs.Bool("upscale-fast", false, "use fast mode (lower quality but faster)")
	upscaleHQFlag := fs.Bool("upscale-hq", false, "use high quality mode (slower but better results)")
	upscaleGPUFlag := fs.Bool("upscale-gpu", false, "use GPU encoding for video output")
	upscaleBitrateFlag := fs.String("upscale-bitrate", "8M", "video bitrate for output (default: 8M)")
	upscaleWorkersFlag := fs.Int("upscale-workers", 0, "number of parallel workers (default: CPU cores)")

	// Set custom usage for our FlagSet
	fs.Usage = func() {
		Helper()
	}

	// Parse the flags early before any manipulation of os.Args
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			Helper()
			return "", ErrHelpRequested
		}
		return "", err
	}

	// Set debug mode based on flag (set unconditionally for consistency)
	IsDebug = *debug

	// Set performance profiling mode
	PerfEnabled = *perf
	if PerfEnabled {
		// Also enable debug for performance mode to see detailed logs
		IsDebug = true
		Debug("Performance profiling enabled")
	}

	// Store global configurations
	SetGlobalSource(*sourceFlag)
	SetGlobalQuality(*qualityFlag)
	SetGlobalMediaType(*mediaTypeFlag)
	SetPreferredSubtitleLanguage(*subsLanguageFlag)
	SetPreferredAudioLanguage(*audioLanguageFlag)
	SetSubtitlesDisabled(*noSubsFlag)
	SetGlobalOutputDir(*outputDirFlag)

	if *noSubsFlag {
		Debug("Subtitles disabled by user")
	}

	if *versionFlag || version.HasVersionArg() {
		version.ShowVersion()
		return "", ErrHelpRequested // Signal version instead of exiting
	}

	if *help || *altHelp {
		Helper()
		return "", ErrHelpRequested // Signal help instead of exiting
	}

	if *updateFlag {
		return "", ErrUpdateRequested // Signal update request
	}

	// Handle download mode
	if *downloadFlag {
		return handleDownloadModeWithSmart(fs.Args(), *rangeFlag, *allFlag, *sourceFlag, *qualityFlag, *allanimeSmartFlag)
	}

	// Handle movie/TV download mode (FlixHQ/SFlix)
	if *movieDownloadFlag {
		return handleMovieDownloadMode(fs.Args(), *rangeFlag, *allFlag, *qualityFlag, *subsLanguageFlag, *mediaTypeFlag)
	}

	// Handle upscale mode
	if *upscaleFlag {
		return handleUpscaleMode(
			fs,
			*upscaleOutputFlag,
			*upscaleScaleFlag,
			*upscalePassesFlag,
			*upscaleFastFlag,
			*upscaleHQFlag,
			*upscaleGPUFlag,
			*upscaleBitrateFlag,
			*upscaleWorkersFlag,
		)
	}

	if *debug {
		Debug("Debug mode is enabled")
	}

	// If the user has provided an anime name as an argument, we use it.
	var animeName string
	if len(fs.Args()) > 0 {
		animeName = strings.Join(fs.Args(), " ")
		// Check if it has some flags and remove them
		if strings.Contains(animeName, "-") {
			animeName = strings.Split(animeName, "-")[0]
		}
		Debug("Anime name", "name", animeName)
		if len(animeName) < minNameLength {
			return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
		}
		return TreatingAnimeName(animeName), nil
	}
	animeName, err := getUserInput("Enter anime/movie name")
	return TreatingAnimeName(animeName), err
}

// getUserInput prompts the user for input the anime name and returns it
func getUserInput(label string) (string, error) {
	var animeName string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(label).
				Description("Type the anime title and press Enter").
				Value(&animeName).
				Validate(func(v string) error {
					if len(strings.TrimSpace(v)) < minNameLength {
						return fmt.Errorf("name must have at least %d characters", minNameLength)
					}
					return nil
				}),
		),
	)

	if err := tui.RunClean(form.Run); err != nil {
		return "", err
	}
	return animeName, nil
}

// TreatingAnimeName removes special characters and spaces from the anime name.
func TreatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	return strings.ReplaceAll(loweredName, " ", "-")
}

// handleDownloadModeWithSmart processes download args with AllAnime Smart option
func handleDownloadModeWithSmart(args []string, isRange bool, isAll bool, source, quality string, allanimeSmart bool) (string, error) {

	if len(args) == 0 {
		return "", fmt.Errorf("download mode requires anime name and episode number/range")
	}

	// Download-all mode: goanime -d -a "anime name"
	if isAll {
		animeName := strings.Join(args, " ")
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			IsAll:         true,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
			OutputDir:     GlobalOutputDir,
		}
		return TreatingAnimeName(animeName), ErrDownloadRequested
	}

	if isRange {
		// Range download: goanime -d -r "anime name" start-end
		if len(args) < 2 {
			return "", fmt.Errorf("range download requires anime name and episode range (e.g., '1-5')")
		}

		animeName := strings.Join(args[:len(args)-1], " ")
		rangeStr := args[len(args)-1]

		// Parse range (e.g., "1-5")
		rangeParts := strings.Split(rangeStr, "-")
		if len(rangeParts) != 2 {
			return "", fmt.Errorf("invalid range format. Use 'start-end' (e.g., '1-5')")
		}

		startEp, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
		if err != nil {
			return "", fmt.Errorf("invalid start episode number: %s", rangeParts[0])
		}

		endEp, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
		if err != nil {
			return "", fmt.Errorf("invalid end episode number: %s", rangeParts[1])
		}

		if startEp > endEp {
			return "", fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
		}

		if startEp < 1 {
			return "", fmt.Errorf("episode numbers must be positive")
		}

		// Store download request
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			IsRange:       true,
			StartEpisode:  startEp,
			EndEpisode:    endEp,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
			OutputDir:     GlobalOutputDir,
		}

		return TreatingAnimeName(animeName), ErrDownloadRequested
	}

	// No episode number provided — show interactive download mode menu
	// This covers: goanime -d "anime name"
	animeName := strings.Join(args, " ")

	// Try parsing last arg as episode number first
	if len(args) >= 2 {
		episodeStr := args[len(args)-1]
		if episodeNum, err := strconv.Atoi(episodeStr); err == nil && episodeNum >= 1 {
			// Last arg is a valid episode number
			animeName = strings.Join(args[:len(args)-1], " ")
			GlobalDownloadRequest = &DownloadRequest{
				AnimeName:     animeName,
				EpisodeNum:    episodeNum,
				IsRange:       false,
				Source:        source,
				Quality:       quality,
				AllAnimeSmart: allanimeSmart,
				OutputDir:     GlobalOutputDir,
			}
			return TreatingAnimeName(animeName), ErrDownloadRequested
		}
	}

	// No episode number — show interactive menu
	var downloadMode string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Download mode for: "+animeName).
				Options(
					huh.NewOption("Download ALL episodes", "all"),
					huh.NewOption("Download a single episode", "single"),
					huh.NewOption("Download a range of episodes", "range"),
				).
				Value(&downloadMode),
		),
	)

	if err := tui.RunClean(form.Run); err != nil {
		return "", fmt.Errorf("download mode selection cancelled: %w", err)
	}

	switch downloadMode {
	case "all":
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			IsAll:         true,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
			OutputDir:     GlobalOutputDir,
		}
		return TreatingAnimeName(animeName), ErrDownloadRequested

	case "single":
		var episodeStr string
		inputForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Episode number").
					Description("Enter the episode number to download").
					Value(&episodeStr).
					Validate(func(v string) error {
						if n, err := strconv.Atoi(v); err != nil || n < 1 {
							return fmt.Errorf("enter a valid positive number")
						}
						return nil
					}),
			),
		)
		if err := tui.RunClean(inputForm.Run); err != nil {
			return "", fmt.Errorf("episode input cancelled: %w", err)
		}
		episodeNum, _ := strconv.Atoi(episodeStr)
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			EpisodeNum:    episodeNum,
			IsRange:       false,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
			OutputDir:     GlobalOutputDir,
		}
		return TreatingAnimeName(animeName), ErrDownloadRequested

	case "range":
		var startStr, endStr string
		rangeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Start episode").
					Description("First episode number").
					Value(&startStr).
					Validate(func(v string) error {
						if n, err := strconv.Atoi(v); err != nil || n < 1 {
							return fmt.Errorf("enter a valid positive number")
						}
						return nil
					}),
				huh.NewInput().
					Title("End episode").
					Description("Last episode number").
					Value(&endStr).
					Validate(func(v string) error {
						if n, err := strconv.Atoi(v); err != nil || n < 1 {
							return fmt.Errorf("enter a valid positive number")
						}
						return nil
					}),
			),
		)
		if err := tui.RunClean(rangeForm.Run); err != nil {
			return "", fmt.Errorf("range input cancelled: %w", err)
		}
		startEp, _ := strconv.Atoi(startStr)
		endEp, _ := strconv.Atoi(endStr)
		if startEp > endEp {
			return "", fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
		}
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			IsRange:       true,
			StartEpisode:  startEp,
			EndEpisode:    endEp,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
			OutputDir:     GlobalOutputDir,
		}
		return TreatingAnimeName(animeName), ErrDownloadRequested

	default:
		return "", fmt.Errorf("unknown download mode selected")
	}
}

// handleUpscaleMode processes upscale command arguments
func handleUpscaleMode(fs *flag.FlagSet, outputPath string, scaleFactor, passes int, fastMode, hqMode, useGPU bool, bitrate string, workers int) (string, error) {
	args := fs.Args()

	if len(args) == 0 {
		return "", fmt.Errorf("upscale mode requires an input file path\nUsage: goanime --upscale <input_file> [options]")
	}

	inputPath := args[0]

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input file not found: %s", inputPath)
	}

	// Set default values
	if scaleFactor < 1 || scaleFactor > 4 {
		scaleFactor = 2
	}
	if passes < 1 || passes > 8 {
		passes = 2
	}

	// Determine strength values based on mode
	strengthColor := 1.0 / 3.0
	strengthGradient := 1.0

	if hqMode {
		passes = 4
		strengthColor = 0.4
	}

	// Store upscale request
	GlobalUpscaleRequest = &UpscaleRequest{
		InputPath:        inputPath,
		OutputPath:       outputPath,
		ScaleFactor:      scaleFactor,
		Passes:           passes,
		StrengthColor:    strengthColor,
		StrengthGradient: strengthGradient,
		FastMode:         fastMode,
		HighQuality:      hqMode,
		PreserveAudio:    true,
		UseGPU:           useGPU,
		VideoBitrate:     bitrate,
		Workers:          workers,
	}

	return inputPath, ErrUpscaleRequested
}

// handleMovieDownloadMode processes movie/TV download arguments for FlixHQ/SFlix
func handleMovieDownloadMode(args []string, isRange bool, isAll bool, quality, subsLanguage, mediaType string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("movie download mode requires movie/TV name\nUsage: goanime -dm \"Movie Name\" (for movies)\n       goanime -dm -r \"TV Show\" season episode-range (for TV episodes)\n       goanime -dm -a \"TV Show\" (download all seasons and episodes)")
	}

	// Determine if it's a movie or TV download
	isTV := mediaType == "tv" || isRange || isAll

	// Download-all mode for TV/series/dorama: goanime -dm -a "Show Name"
	if isAll {
		showName := strings.Join(args, " ")
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:    showName,
			IsAll:        true,
			IsTV:         true,
			Quality:      quality,
			SubsLanguage: subsLanguage,
			OutputDir:    GlobalOutputDir,
		}
		return TreatingAnimeName(showName), ErrMovieDownloadRequested
	}

	if isTV && isRange {
		// TV episode range download: goanime -dm -r "TV Show" season start-end
		if len(args) < 3 {
			return "", fmt.Errorf("TV episode range download requires show name, season number, and episode range\nUsage: goanime -dm -r \"TV Show\" 1 1-5")
		}

		showName := strings.Join(args[:len(args)-2], " ")
		seasonStr := args[len(args)-2]
		rangeStr := args[len(args)-1]

		// Parse season number
		seasonNum, err := strconv.Atoi(strings.TrimSpace(seasonStr))
		if err != nil {
			return "", fmt.Errorf("invalid season number: %s", seasonStr)
		}

		// Parse range (e.g., "1-5")
		rangeParts := strings.Split(rangeStr, "-")
		if len(rangeParts) != 2 {
			return "", fmt.Errorf("invalid episode range format. Use 'start-end' (e.g., '1-5')")
		}

		startEp, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
		if err != nil {
			return "", fmt.Errorf("invalid start episode number: %s", rangeParts[0])
		}

		endEp, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
		if err != nil {
			return "", fmt.Errorf("invalid end episode number: %s", rangeParts[1])
		}

		if startEp > endEp {
			return "", fmt.Errorf("start episode (%d) cannot be greater than end episode (%d)", startEp, endEp)
		}

		if startEp < 1 || seasonNum < 1 {
			return "", fmt.Errorf("season and episode numbers must be positive")
		}

		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:    showName,
			IsRange:      true,
			IsTV:         true,
			SeasonNum:    seasonNum,
			StartEpisode: startEp,
			EndEpisode:   endEp,
			Quality:      quality,
			SubsLanguage: subsLanguage,
			OutputDir:    GlobalOutputDir,
		}

		return TreatingAnimeName(showName), ErrMovieDownloadRequested
	}

	if isTV {
		// Single TV episode download: goanime -dm --type tv "TV Show" season episode
		if len(args) < 3 {
			return "", fmt.Errorf("TV episode download requires show name, season number, and episode number\nUsage: goanime -dm --type tv \"TV Show\" 1 5")
		}

		showName := strings.Join(args[:len(args)-2], " ")
		seasonStr := args[len(args)-2]
		episodeStr := args[len(args)-1]

		seasonNum, err := strconv.Atoi(strings.TrimSpace(seasonStr))
		if err != nil {
			return "", fmt.Errorf("invalid season number: %s", seasonStr)
		}

		episodeNum, err := strconv.Atoi(strings.TrimSpace(episodeStr))
		if err != nil {
			return "", fmt.Errorf("invalid episode number: %s", episodeStr)
		}

		if seasonNum < 1 || episodeNum < 1 {
			return "", fmt.Errorf("season and episode numbers must be positive")
		}

		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:    showName,
			IsTV:         true,
			SeasonNum:    seasonNum,
			EpisodeNum:   episodeNum,
			IsRange:      false,
			Quality:      quality,
			SubsLanguage: subsLanguage,
			OutputDir:    GlobalOutputDir,
		}

		return TreatingAnimeName(showName), ErrMovieDownloadRequested
	}

	// Movie download: goanime -dm "Movie Name"
	movieName := strings.Join(args, " ")

	GlobalDownloadRequest = &DownloadRequest{
		AnimeName:    movieName,
		IsMovie:      true,
		IsRange:      false,
		Quality:      quality,
		SubsLanguage: subsLanguage,
		OutputDir:    GlobalOutputDir,
	}

	return TreatingAnimeName(movieName), ErrMovieDownloadRequested
}

// Pre-compiled regexes for SanitizeForFilename and related functions (hot path)
var (
	bracketTagRe = regexp.MustCompile(`\[(?i:English|Portuguese|Português|PT-BR|Movies?(?:/TV)?|TV|MoviesTV|Unknown|Multilanguage|Multi[ _-]?Subs?|HD|9Anime|SUB|DUB)\]`)
	ageClassRe   = regexp.MustCompile(`\s+(A\d{1,2}|AL|L)\s*$`)
	scoreRe      = regexp.MustCompile(`\s+\d{1,2}\.\d{1,2}\s*$`)
)

// SanitizeForFilename removes characters that are not allowed in file/directory names
// and returns a cleaned version of the name suitable for Plex/Jellyfin media libraries.
// It also strips ratings (e.g. "7.27"), age classifications (e.g. "A14", "L"),
// and language/source/metadata tags that many anime sources append to titles.
func SanitizeForFilename(name string) string {
	// Remove bracketed tags: [English], [Multilanguage], [Movie], [9Anime], [HD], etc.
	name = bracketTagRe.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	// Remove trailing parenthesized 9anime/source metadata.
	// e.g. "Boruto (HD SUB DUB Ep 293/293)" → "Boruto"
	// Matches if the parenthesized suffix contains episode numbers, SUB, DUB,
	// HD, or Multilanguage — i.e. metadata, not a real subtitle like "(Shippuuden)".
	name = strip9AnimeParenMeta(name)

	// Remove trailing anime source metadata: ratings like "7.27" and age
	// classifications like "A14", "A12", "A16", "A18", "L", "AL".
	// These are commonly appended by AllAnime/AnimeFire sources.
	// Pattern: strip trailing tokens that look like scores or classifications.
	name = stripTrailingAnimeMetadata(name)

	// Remove characters not allowed in filenames across platforms
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "")
	}
	// Remove trailing dots and spaces (problematic on Windows)
	name = strings.TrimRight(name, ". ")
	// Collapse multiple spaces
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}
	return strings.TrimSpace(name)
}

// strip9AnimeParenMeta removes trailing parenthesized metadata appended by 9anime
// search results, e.g. "(HD SUB DUB Ep 293/293)" or "(Multilanguage SUB Ep 100)".
// It only strips when the parenthesized part is at the END of the string and looks
// like metadata (contains keywords like SUB, DUB, HD, Multilanguage, or episode numbers),
// preserving legitimate subtitle parentheses like "(Shippuuden)" or "(Dublado)".
func strip9AnimeParenMeta(name string) string {
	idx := strings.LastIndex(name, " (")
	if idx <= 0 {
		return name
	}
	suffix := name[idx:]
	// Only strip if the closing paren is at the very end of the string
	closeIdx := strings.LastIndex(suffix, ")")
	if closeIdx < 0 || closeIdx != len(suffix)-1 {
		return name
	}
	candidate := strings.ToUpper(suffix)
	isMetadata := strings.Contains(candidate, "SUB") ||
		strings.Contains(candidate, "DUB") ||
		strings.Contains(candidate, "HD") ||
		strings.Contains(candidate, "MULTILANGUAGE") ||
		strings.Contains(candidate, "MULTI") ||
		strings.Contains(candidate, "EP ")
	if isMetadata {
		return strings.TrimSpace(name[:idx])
	}
	return name
}

// stripTrailingAnimeMetadata removes common metadata that anime sources append
// to titles, such as scores (e.g. "7.27"), age classifications (e.g. "A14", "L"),
// and other trailing tokens that don't belong in a Plex-style filename.
//
// Example: "Black Clover (Dublado) 7.27 A14" → "Black Clover (Dublado)"
func stripTrailingAnimeMetadata(name string) string {
	changed := true
	for changed {
		changed = false
		// Strip trailing age classification
		if loc := ageClassRe.FindStringIndex(name); loc != nil {
			name = strings.TrimSpace(name[:loc[0]])
			changed = true
		}
		// Strip trailing decimal rating
		if loc := scoreRe.FindStringIndex(name); loc != nil {
			name = strings.TrimSpace(name[:loc[0]])
			changed = true
		}
	}
	return name
}

// MediaMeta carries external IDs, year, and official title for
// Plex/Jellyfin-compatible folder naming. Pass nil when metadata is
// unavailable — all helpers treat a nil *MediaMeta the same as an empty one.
type MediaMeta struct {
	OfficialTitle string // Official title from TMDB/AniList (English or Romaji)
	Year          string // Release year, e.g. "2003"
	TMDBID        int    // TheMovieDB ID
	IMDBID        string // IMDB ID, e.g. "tt0369179"
	AnilistID     int    // AniList ID
	MalID         int    // MyAnimeList ID
}

// resolveTitle returns the best available title: OfficialTitle from metadata
// databases (TMDB, AniList), falling back to the sanitized scraper name.
func resolveTitle(scraperName string, meta *MediaMeta) string {
	if meta != nil && meta.OfficialTitle != "" {
		safe := SanitizeForFilename(meta.OfficialTitle)
		if safe != "" {
			return safe
		}
	}
	safe := SanitizeForFilename(scraperName)
	if safe != "" {
		return safe
	}
	return "Unknown"
}

// BuildMediaFolderName returns a Plex/Jellyfin-compatible folder name.
// Format: "<OfficialTitle> (<Year>) {tmdb-123} {imdb-tt456}"
// Prefers the official title from TMDB/AniList over the scraped name.
// External IDs use the {source-id} syntax recognised by both Plex and Jellyfin.
func BuildMediaFolderName(name string, meta *MediaMeta) string {
	result := resolveTitle(name, meta)
	if meta == nil {
		return result
	}

	// Append year
	if meta.Year != "" {
		result += " (" + meta.Year + ")"
	}

	// Append external IDs in priority order (Plex/Jellyfin {source-id} syntax)
	if meta.TMDBID > 0 {
		result += fmt.Sprintf(" {tmdb-%d}", meta.TMDBID)
	}
	if meta.IMDBID != "" {
		result += fmt.Sprintf(" {imdb-%s}", meta.IMDBID)
	}
	if meta.AnilistID > 0 {
		result += fmt.Sprintf(" {anilist-%d}", meta.AnilistID)
	}
	if meta.MalID > 0 {
		result += fmt.Sprintf(" {mal-%d}", meta.MalID)
	}

	return result
}

// BuildMediaFileName returns a Plex/Jellyfin-compatible base name for files.
// Format: "<OfficialTitle> (<Year>)" — external IDs are only on the folder, not the file.
// Prefers the official title from TMDB/AniList over the scraped name.
func BuildMediaFileName(name string, meta *MediaMeta) string {
	title := resolveTitle(name, meta)
	if meta != nil && meta.Year != "" {
		return title + " (" + meta.Year + ")"
	}
	return title
}

// DefaultDownloadDir returns the base download directory for anime content.
// If the user specified a custom directory via -o flag, that is returned.
// Otherwise returns the default ~/.local/goanime/downloads/anime/ path.
func DefaultDownloadDir() string {
	if outputDir := GetGlobalOutputDir(); outputDir != "" {
		return outputDir
	}
	userHome, _ := os.UserHomeDir()
	return filepath.Join(userHome, ".local", "goanime", "downloads", "anime")
}

// DefaultMovieDownloadDir returns the base download directory for movie/TV content.
// If the user specified a custom directory via -o flag, that is returned.
// Otherwise returns the default ~/.local/goanime/downloads/movies/ path.
func DefaultMovieDownloadDir() string {
	if outputDir := GetGlobalOutputDir(); outputDir != "" {
		return outputDir
	}
	userHome, _ := os.UserHomeDir()
	return filepath.Join(userHome, ".local", "goanime", "downloads", "movies")
}

// FormatPlexMoviePath builds a Plex/Jellyfin-compatible file path for a movie.
// Format: <baseDir>/<MovieName (Year) {ids}>/<MovieName (Year)>.mp4
// The folder includes external IDs; the filename includes only name and year.
func FormatPlexMoviePath(baseDir, movieName string, year string, meta ...*MediaMeta) string {
	var m *MediaMeta
	if len(meta) > 0 {
		m = meta[0]
	}
	// Ensure year is populated from meta if not passed directly
	if year == "" && m != nil {
		year = m.Year
	}
	// Build a consistent meta for helpers (merge year param)
	effectiveMeta := &MediaMeta{}
	if m != nil {
		*effectiveMeta = *m
	}
	if year != "" {
		effectiveMeta.Year = year
	}

	folderName := BuildMediaFolderName(movieName, effectiveMeta)
	fileName := BuildMediaFileName(movieName, effectiveMeta)
	return filepath.ToSlash(filepath.Join(baseDir, folderName, fileName+".mp4"))
}

// FormatPlexMovieDir returns the directory path for a Plex-compatible movie.
// Format: <baseDir>/<MovieName (Year) {ids}>/
func FormatPlexMovieDir(baseDir, movieName string, meta ...*MediaMeta) string {
	var m *MediaMeta
	if len(meta) > 0 {
		m = meta[0]
	}
	folderName := BuildMediaFolderName(movieName, m)
	return filepath.ToSlash(filepath.Join(baseDir, folderName))
}

// FormatPlexEpisodePath builds a Plex/Jellyfin-compatible file path for an episode.
// Format: <baseDir>/<Name (Year) {ids}>/Season XX/<Name (Year)> - SXXeXX.mp4
// The folder includes external IDs; the filename includes name, year, and episode info.
func FormatPlexEpisodePath(baseDir, animeName string, season, episodeNum int, meta ...*MediaMeta) string {
	var m *MediaMeta
	if len(meta) > 0 {
		m = meta[0]
	}
	folderName := BuildMediaFolderName(animeName, m)
	fileName := BuildMediaFileName(animeName, m)
	if season < 1 {
		season = 1
	}
	seasonDir := fmt.Sprintf("Season %02d", season)
	filename := fmt.Sprintf("%s - S%02dE%02d.mp4", fileName, season, episodeNum)
	return filepath.ToSlash(filepath.Join(baseDir, folderName, seasonDir, filename))
}

// FormatPlexEpisodeDir returns the directory path for a Plex-compatible anime season.
// Format: <baseDir>/<Name (Year) {ids}>/Season XX/
func FormatPlexEpisodeDir(baseDir, animeName string, season int, meta ...*MediaMeta) string {
	var m *MediaMeta
	if len(meta) > 0 {
		m = meta[0]
	}
	folderName := BuildMediaFolderName(animeName, m)
	if season < 1 {
		season = 1
	}
	seasonDir := fmt.Sprintf("Season %02d", season)
	return filepath.ToSlash(filepath.Join(baseDir, folderName, seasonDir))
}

// PlexEpisodeFilename returns just the filename part in Plex format.
// Format: <Name (Year)> - SXXeXX.mp4
func PlexEpisodeFilename(animeName string, season, episodeNum int, meta ...*MediaMeta) string {
	var m *MediaMeta
	if len(meta) > 0 {
		m = meta[0]
	}
	fileName := BuildMediaFileName(animeName, m)
	if season < 1 {
		season = 1
	}
	return fmt.Sprintf("%s - S%02dE%02d.mp4", fileName, season, episodeNum)
}
