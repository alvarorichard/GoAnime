package util

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

func SetGlobalSource(source string) {
	runtimeStateMu.Lock()
	GlobalSource = source
	runtimeStateMu.Unlock()
}

func GetGlobalSource() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalSource
}

func SetGlobalQuality(quality string) {
	runtimeStateMu.Lock()
	GlobalQuality = quality
	runtimeStateMu.Unlock()
}

func GetGlobalQuality() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalQuality
}

func SetGlobalMediaType(mediaType string) {
	runtimeStateMu.Lock()
	GlobalMediaType = mediaType
	runtimeStateMu.Unlock()
}

func GetGlobalMediaType() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalMediaType
}

func SetPreferredSubtitleLanguage(language string) {
	runtimeStateMu.Lock()
	GlobalSubsLanguage = language
	runtimeStateMu.Unlock()
}

func GetPreferredSubtitleLanguage() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalSubsLanguage
}

func SetPreferredAudioLanguage(language string) {
	runtimeStateMu.Lock()
	GlobalAudioLanguage = language
	runtimeStateMu.Unlock()
}

func GetPreferredAudioLanguage() string {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalAudioLanguage
}

func SetSubtitlesDisabled(disabled bool) {
	runtimeStateMu.Lock()
	GlobalNoSubs = disabled
	runtimeStateMu.Unlock()
}

func SubtitlesDisabled() bool {
	runtimeStateMu.RLock()
	defer runtimeStateMu.RUnlock()
	return GlobalNoSubs
}

func SetGlobalOutputDir(outputDir string) {
	runtimeStateMu.Lock()
	GlobalOutputDir = outputDir
	runtimeStateMu.Unlock()
}

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

func ClearGlobalUpscaleRequest() {
	SetGlobalUpscaleRequest(nil)
}
