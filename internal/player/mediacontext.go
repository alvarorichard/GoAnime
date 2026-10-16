package player

import (
	"sync"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Media context and resume overrides for tracking writes.
//
// The CONTINUE list on the SEARCH home screen needs the show's name, the
// source and the audio word for every tracked episode, but the tracking
// writes happen deep in playVideo, which only sees the episode. The flow
// entry (HandleDownloadAndPlay, used by series and movies alike) records the
// context here; the writes below read it back. Snapshot fallbacks
// (SetAnimeName, GetGlobalAnimeSource) cover flows that predate it.

// mediaContext is one flow's show name, source and audio word.
type mediaContext struct {
	animeTitle string
	source     string
	audio      string
}

var mediaContextMu sync.RWMutex
var currentMediaContext mediaContext

// SetMediaContext records the show name, source and audio word for the
// tracking writes of this flow.
func SetMediaContext(animeTitle, source, audio string) {
	mediaContextMu.Lock()
	defer mediaContextMu.Unlock()
	currentMediaContext = mediaContext{animeTitle: animeTitle, source: source, audio: audio}
}

// snapshotMediaContext returns the recorded context, falling back to the
// older globals when a flow did not record one.
func snapshotMediaContext() mediaContext {
	mediaContextMu.RLock()
	ctx := currentMediaContext
	mediaContextMu.RUnlock()
	if ctx.animeTitle == "" {
		ctx.animeTitle = snapshotMedia().AnimeName
	}
	if ctx.source == "" {
		ctx.source = util.GetGlobalAnimeSource()
	}
	return ctx
}

// Resume overrides, consumed once by the next initTracking: the CONTINUE
// list's enter resumes without asking, and "r from start" skips the dialog
// at position zero.
var (
	resumeOverrideMu sync.Mutex
	resumeOverride   = -1
	fromStartNext    bool
)

// ForceResumeNext makes the next initTracking return position without
// asking whether to resume.
func ForceResumeNext(position int) {
	resumeOverrideMu.Lock()
	defer resumeOverrideMu.Unlock()
	resumeOverride = position
	fromStartNext = false
}

// ForceStartNext makes the next initTracking start from zero without asking.
func ForceStartNext() {
	resumeOverrideMu.Lock()
	defer resumeOverrideMu.Unlock()
	resumeOverride = -1
	fromStartNext = true
}

// takeResumeOverride consumes a pending override: position >= 0 resumes
// there, fromStart starts at zero. ok is false when nothing was pending.
func takeResumeOverride() (position int, fromStart, ok bool) {
	resumeOverrideMu.Lock()
	defer resumeOverrideMu.Unlock()
	if fromStartNext {
		fromStartNext = false
		return 0, true, true
	}
	if resumeOverride >= 0 {
		position = resumeOverride
		resumeOverride = -1
		return position, false, true
	}
	return 0, false, false
}

// ClearResumeOverride drops a pending override without consuming it, so a
// fallback search after a failed direct open behaves like a fresh one.
func ClearResumeOverride() {
	resumeOverrideMu.Lock()
	defer resumeOverrideMu.Unlock()
	resumeOverride = -1
	fromStartNext = false
}

// PeekResumeOverride reports the pending override without consuming it.
func PeekResumeOverride() (position int, fromStart, ok bool) {
	resumeOverrideMu.Lock()
	defer resumeOverrideMu.Unlock()
	if fromStartNext {
		return 0, true, true
	}
	if resumeOverride >= 0 {
		return resumeOverride, false, true
	}
	return 0, false, false
}
