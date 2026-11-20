package handlers

import (
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// lastHome is the last decision the SEARCH home screen made, with where the
// screen was: what to reopen it on when the user comes back to it from the
// results of that search. It lives here, beside the one function that opens the
// screen, so the loops between the two (the results screen, playback, the
// retry loop) pass nothing along and cannot lose it.
var lastHome struct {
	sync.Mutex
	searched string // the title the decision went on to search for
	restore  tui.HomeRestore
}

// rememberHomeSearch records res as the last decision. Every decision
// replaces the one before, including one with nothing to restore.
func rememberHomeSearch(res tui.SearchHomeResult) {
	lastHome.Lock()
	defer lastHome.Unlock()
	lastHome.searched = searchedByHome(res)
	lastHome.restore = res.Restore
}

// searchedByHome is the title a decision goes on to search for: the query, or
// for a CONTINUE row the title HandleContinue falls back to.
func searchedByHome(res tui.SearchHomeResult) string {
	if res.HasEntry {
		entry := res.Entry
		return cleanContinueTitle(&entry)
	}
	return res.Query
}

// homeRestoreFor is where the home screen reopens when it is opened with
// query kept in the box, the way back from a search. It restores only when
// query is the search the last decision made; anything else (a search edited
// on the way, a screen opened with no query) reopens as it always did.
func homeRestoreFor(query string) tui.HomeRestore {
	if strings.TrimSpace(query) == "" {
		return tui.HomeRestore{}
	}
	lastHome.Lock()
	defer lastHome.Unlock()
	if lastHome.searched == "" || !sameSearch(lastHome.searched, query) {
		return tui.HomeRestore{}
	}
	return lastHome.restore
}

// sameSearch reports whether a and b are the same search: main turns the typed
// "One Piece" into "one-piece" before searching, and that is what comes back.
func sameSearch(a, b string) bool {
	return util.TreatingAnimeName(strings.Join(strings.Fields(a), " ")) ==
		util.TreatingAnimeName(strings.Join(strings.Fields(b), " "))
}

// resetLastHome forgets the last decision, for tests.
func resetLastHome() {
	lastHome.Lock()
	defer lastHome.Unlock()
	lastHome.searched, lastHome.restore = "", tui.HomeRestore{}
}
