package util

import "sync"

// PlaybackState captures the stream-scoped playback metadata that providers
// and player/download flows exchange during resolution.
type PlaybackState struct {
	Subtitles   []SubtitleInfo
	Referer     string
	AnimeSource string
}

var playbackStateMu sync.RWMutex

func cloneSubtitleInfos(subs []SubtitleInfo) []SubtitleInfo {
	if len(subs) == 0 {
		return nil
	}
	cloned := make([]SubtitleInfo, len(subs))
	copy(cloned, subs)
	return cloned
}

// CurrentPlaybackState returns a copy of the current playback-scoped state.
func CurrentPlaybackState() PlaybackState {
	playbackStateMu.RLock()
	defer playbackStateMu.RUnlock()

	return PlaybackState{
		Subtitles:   cloneSubtitleInfos(GlobalSubtitles),
		Referer:     GlobalReferer,
		AnimeSource: GlobalAnimeSource,
	}
}

// GetGlobalSubtitles returns a defensive copy of the currently stored subtitles.
func GetGlobalSubtitles() []SubtitleInfo {
	playbackStateMu.RLock()
	defer playbackStateMu.RUnlock()
	return cloneSubtitleInfos(GlobalSubtitles)
}

// SubtitleCount returns the number of currently stored subtitle tracks.
func SubtitleCount() int {
	playbackStateMu.RLock()
	defer playbackStateMu.RUnlock()
	return len(GlobalSubtitles)
}

// HasGlobalSubtitles reports whether there are subtitle tracks stored for playback.
func HasGlobalSubtitles() bool {
	return SubtitleCount() > 0
}

// ResetPlaybackState clears the stream-scoped playback state before resolving a new stream.
func ResetPlaybackState() {
	playbackStateMu.Lock()
	defer playbackStateMu.Unlock()
	GlobalSubtitles = nil
	GlobalReferer = ""
	GlobalAnimeSource = ""
}

// PreparePlaybackSubtitles applies the source-specific subtitle selection flow.
func PreparePlaybackSubtitles() {
	if Is9AnimeSource() {
		PromptSubtitleLanguage()
		return
	}
	if HasGlobalSubtitles() {
		SelectSubtitles()
	}
}
