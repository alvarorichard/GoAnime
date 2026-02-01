package util

import (
	"errors"
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/alvarorichard/Goanime/internal/version"
	"github.com/charmbracelet/huh"
)

// SubtitleInfo represents a single subtitle track
type SubtitleInfo struct {
	URL      string
	Language string
	Label    string
}

var (
	IsDebug             bool
	minNameLength       = 4
	ErrHelpRequested    = errors.New("help requested") // Custom error for help
	GlobalSource        string                         // Global variable to store selected source
	GlobalQuality       string                         // Global variable to store selected quality
	GlobalMediaType     string                         // Global variable to store media type (anime, movie, tv)
	GlobalSubsLanguage  string                         // Global variable to store subtitle language
	GlobalAudioLanguage string                         // Global variable to store preferred audio language
	GlobalSubtitles     []SubtitleInfo                 // Global variable to store current subtitles for playback
	GlobalNoSubs        bool                           // Global flag to disable subtitles
)

// SetGlobalSubtitles stores subtitles for the current playback session
func SetGlobalSubtitles(subs []SubtitleInfo) {
	GlobalSubtitles = subs
	if len(subs) > 0 {
		Debugf("Stored %d subtitle track(s) for playback", len(subs))
		for i, sub := range subs {
			Debugf("  Subtitle %d: %s (%s)", i+1, sub.Label, sub.Language)
		}
	}
}

// ClearGlobalSubtitles clears stored subtitles
func ClearGlobalSubtitles() {
	GlobalSubtitles = nil
}

// GetSubtitleArgs returns mpv arguments for subtitles
// Based on lobster.sh implementation:
// - Single subtitle: --sub-file='URL'
// - Multiple subtitles: --sub-files='URL1:URL2:...'
func GetSubtitleArgs() []string {
	if GlobalNoSubs || len(GlobalSubtitles) == 0 {
		return nil
	}

	// Collect all subtitle URLs
	var urls []string
	for _, sub := range GlobalSubtitles {
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
var cleanupFuncs []func()

// RegisterCleanup registers a function to be called on program exit
func RegisterCleanup(fn func()) {
	cleanupFuncs = append(cleanupFuncs, fn)
}

// RunCleanup runs all registered cleanup functions
func RunCleanup() {
	for _, fn := range cleanupFuncs {
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
		return fmt.Sprintf("%+v", err)
	} else {
		return fmt.Sprintf("%v -- run the program with -debug to see details", err)
	}
}

// Helper prints the beautiful help message
func Helper() {
	ShowBeautifulHelp()
}

// Custom error types for different exit conditions
var (
	ErrUpdateRequested   = errors.New("update requested")
	ErrDownloadRequested = errors.New("download requested")
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
}

// Global variable to store download request
var GlobalDownloadRequest *DownloadRequest

// FlagParser parses the -flags and returns the anime name
func FlagParser() (string, error) {
	// Define flags
	debug := flag.Bool("debug", false, "enable debug mode")
	perf := flag.Bool("perf", false, "enable performance profiling")
	help := flag.Bool("help", false, "show help message")
	altHelp := flag.Bool("h", false, "show help message")
	versionFlag := flag.Bool("version", false, "show version information")
	updateFlag := flag.Bool("update", false, "check for updates and update if available")
	downloadFlag := flag.Bool("d", false, "download mode")
	rangeFlag := flag.Bool("r", false, "download episode range (use with -d)")
	sourceFlag := flag.String("source", "", "specify anime source (allanime, animefire)")
	qualityFlag := flag.String("quality", "best", "specify video quality (best, worst, 720p, 1080p, etc.)")
	allanimeSmartFlag := flag.Bool("allanime-smart", false, "enable AllAnime Smart Range: auto-skip intros/outros and use priority mirrors")
	mediaTypeFlag := flag.String("type", "", "specify media type (anime, movie, tv)")
	subsLanguageFlag := flag.String("subs", "english", "specify subtitle language for movies/TV (FlixHQ only)")
	audioLanguageFlag := flag.String("audio", "pt-BR,pt,english", "specify preferred audio language for movies/TV (FlixHQ only)")
	noSubsFlag := flag.Bool("no-subs", false, "disable subtitles for movies/TV (FlixHQ only)")

	// Parse the flags early before any manipulation of os.Args
	flag.Parse()

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
	GlobalSource = *sourceFlag
	GlobalQuality = *qualityFlag
	GlobalMediaType = *mediaTypeFlag
	GlobalSubsLanguage = *subsLanguageFlag
	GlobalAudioLanguage = *audioLanguageFlag
	GlobalNoSubs = *noSubsFlag

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
		return handleDownloadModeWithSmart(*rangeFlag, *sourceFlag, *qualityFlag, *allanimeSmartFlag)
	}

	if *debug {
		Debug("Debug mode is enabled")
	}

	// If the user has provided an anime name as an argument, we use it.
	var animeName string
	if len(flag.Args()) > 0 {
		animeName = strings.Join(flag.Args(), " ")
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
	animeName, err := getUserInput("Enter anime name")
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

	if err := form.Run(); err != nil {
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
func handleDownloadModeWithSmart(isRange bool, source, quality string, allanimeSmart bool) (string, error) {
	args := flag.Args()

	if len(args) == 0 {
		return "", fmt.Errorf("download mode requires anime name and episode number/range")
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
		}

		return TreatingAnimeName(animeName), ErrDownloadRequested

	} else {
		// Single episode download: goanime -d "anime name" episode_number
		if len(args) < 2 {
			return "", fmt.Errorf("single episode download requires anime name and episode number")
		}

		animeName := strings.Join(args[:len(args)-1], " ")
		episodeStr := args[len(args)-1]

		episodeNum, err := strconv.Atoi(episodeStr)
		if err != nil {
			return "", fmt.Errorf("invalid episode number: %s", episodeStr)
		}

		if episodeNum < 1 {
			return "", fmt.Errorf("episode number must be positive")
		}

		// Store download request
		GlobalDownloadRequest = &DownloadRequest{
			AnimeName:     animeName,
			EpisodeNum:    episodeNum,
			IsRange:       false,
			Source:        source,
			Quality:       quality,
			AllAnimeSmart: allanimeSmart,
		}

		return TreatingAnimeName(animeName), ErrDownloadRequested
	}
}
