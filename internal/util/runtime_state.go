// Package util provides shared runtime state, configuration, and helper functions for GoAnime.
package util // nosemgrep

import "sync"

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
