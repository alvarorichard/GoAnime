package player

import (
	"strings"
	"sync"
)

// Episodes-screen decisions (tela 5) travel to playback in these one-shot
// globals: the picker returns a selection, and the request alongside it
// tells HandleDownloadAndPlay whether to play directly, open the OPTIONS
// menu, or auto-confirm a download. Consumed once; clearing on every fresh
// picker keeps a stale request out of the next flow.

// EpisodeMenuRequest is what the episodes screen asked for.
type EpisodeMenuRequest struct {
	// ShowMenu opens the OPTIONS menu instead of playing directly (m key).
	ShowMenu bool
	// DownloadCode auto-confirms an OPTIONS choice: 1 this episode, 2 a
	// range, 5 all. Zero plays directly.
	DownloadCode int
	// RangeStart presets the range's start episode for code 2 (D key).
	// Zero prompts for the whole range.
	RangeStart int
}

var episodeMenuMu sync.Mutex
var pendingEpisodeMenu EpisodeMenuRequest
var episodeMenuPending bool

// RequestEpisodeMenu records the episodes screen's decision for the next
// HandleDownloadAndPlay.
func RequestEpisodeMenu(req EpisodeMenuRequest) {
	episodeMenuMu.Lock()
	defer episodeMenuMu.Unlock()
	pendingEpisodeMenu = req
	episodeMenuPending = true
}

// TakeEpisodeMenu consumes the pending request, if any.
func TakeEpisodeMenu() (EpisodeMenuRequest, bool) {
	episodeMenuMu.Lock()
	defer episodeMenuMu.Unlock()
	if !episodeMenuPending {
		return EpisodeMenuRequest{}, false
	}
	req := pendingEpisodeMenu
	pendingEpisodeMenu = EpisodeMenuRequest{}
	episodeMenuPending = false
	return req, true
}

// ClearEpisodeMenu discards a pending request: a fresh picker means fresh
// intent, and an abandoned one must never leak into the next flow.
func ClearEpisodeMenu() {
	episodeMenuMu.Lock()
	defer episodeMenuMu.Unlock()
	pendingEpisodeMenu = EpisodeMenuRequest{}
	episodeMenuPending = false
}

// episodeListContext carries the show for the episodes screen's header
// (tela 5): "Search > Mukashi Banashi > Episodes" with "AnimeFire ·
// Legendado" at the right. Recorded where the anime is known.
type episodeListContext struct {
	title  string
	source string
	audio  string
}

var episodeListContextMu sync.RWMutex
var currentEpisodeListContext episodeListContext

// SetEpisodeListContext records the show the next episode list belongs to.
func SetEpisodeListContext(title, source, audio string) {
	episodeListContextMu.Lock()
	defer episodeListContextMu.Unlock()
	currentEpisodeListContext = episodeListContext{title: title, source: source, audio: audio}
}

// snapshotEpisodeListContext returns the recorded show, if any.
func snapshotEpisodeListContext() episodeListContext {
	episodeListContextMu.RLock()
	defer episodeListContextMu.RUnlock()
	return currentEpisodeListContext
}

// displayListSource is how a source reads in the episodes header.
func displayListSource(source string) string {
	trimmed := strings.TrimSpace(source)
	if strings.EqualFold(strings.TrimSuffix(trimmed, ".io"), "animefire") {
		return "AnimeFire"
	}
	return trimmed
}
