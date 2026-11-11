package tui

import (
	"slices"
	"strings"
)

// HomeRestore is where the search home screen reopens after the user comes
// back to it from a search that it started: the same tab, the same row, and
// the box as the user left it. Without it a beginner who opened a favorite and
// pressed esc on the results landed on the ALL tab with a slug in the box and
// lost their place.
//
// It is a plain value, level-agnostic and independent of the scope a search
// runs in: Tab names a tab of the screen (TabFavorites is a tab, never a
// search scope, so a restore must not be read as one) and RowKey names a row by
// what it stands for, not by its position, so it survives the lists being
// reordered by the search itself (RECENT gains a row on top).
//
// The zero value restores nothing.
type HomeRestore struct {
	// Tab is one of the Tab* names.
	Tab string
	// RowKey is the row to highlight; empty when the search came from the box.
	RowKey string
	// Query refills the box; empty when the search came from a row.
	Query string
}

// The tab names a HomeRestore carries. They are the screen's tabs, in order,
// named by the scope they search with; TabFavorites searches nothing.
const (
	TabAll       = "all"
	TabPTBR      = "ptbr"
	TabEnglish   = "hianime"
	TabMoviesTV  = "movietv"
	TabFavorites = "favorites"
)

// Row key prefixes: the list a row belongs to.
const (
	rowKeyContinue = "continue:"
	rowKeyRecent   = "recent:"
	rowKeyFavorite = "favorite:"
)

// IsZero reports whether r restores nothing.
func (r HomeRestore) IsZero() bool { return r == HomeRestore{} }

// tabIndex is the position of r.Tab among the screen's tabs.
func (r HomeRestore) tabIndex() (int, bool) {
	for i, s := range homeScopes {
		if tabName(s.Value) == r.Tab {
			return i, true
		}
	}
	return 0, false
}

// tabName names the tab whose scope value is value ("" is ALL).
func tabName(value string) string {
	if value == "" {
		return TabAll
	}
	return value
}

// withRestore is o with the restore applied: its tab replaces ScopeIndex and
// its box replaces Query (a slug the caller passes is never shown when the
// search came from a row). A restore naming no known tab is ignored, leaving o
// exactly as it was.
func (o SearchHomeOptions) withRestore() SearchHomeOptions {
	idx, ok := o.Restore.tabIndex()
	if !ok {
		return o
	}
	o.ScopeIndex = idx
	o.Query = o.Restore.Query
	return o
}

// placeRestore highlights the row r names, on the tab already selected. A row
// that is no longer there (removed, or beyond what the tab draws) leaves the
// box focused.
func (m *searchHomeModel) placeRestore(r HomeRestore) {
	if _, ok := r.tabIndex(); !ok || r.RowKey == "" {
		return
	}
	pos := slices.IndexFunc(m.rows, func(row homeRow) bool { return m.rowKey(row) == r.RowKey })
	if pos < 0 {
		return
	}
	m.cursor = pos
	m.focusInput = false
	m.input.Blur()
}

// rowKey names a row by what it stands for.
func (m *searchHomeModel) rowKey(row homeRow) string {
	switch row.kind {
	case homeRowContinue:
		return rowKeyContinue + continueRowKey(m.continueAll[row.index])
	case homeRowRecent:
		return rowKeyRecent + foldForMatch(m.recentAll[row.index].Query)
	default:
		return rowKeyFavorite + favoriteRowKey(m.favoriteAll[row.index])
	}
}

// continueRowKey names a CONTINUE row by its show, as the row itself groups
// episodes: the key of its latest episode changes with every episode watched,
// the show does not. Rows with no show title fall back to the episode's
// address without its ":epN" suffix.
func continueRowKey(e HomeContinueEntry) string {
	if title := foldForMatch(singleLine(e.AnimeTitle)); title != "" {
		return "t:" + title + "\x00" + foldForMatch(singleLine(e.Source))
	}
	base := e.Key
	if i := strings.LastIndex(base, ":ep"); i >= 0 && isDigits(base[i+3:]) {
		base = base[:i]
	}
	return "u:" + strings.ToLower(strings.TrimSpace(base))
}

// favoriteRowKey names a favorite by its stored key, else by its title.
func favoriteRowKey(f HomeFavorite) string {
	if f.Key != "" {
		return f.Key
	}
	return "title:" + foldForMatch(singleLine(f.Title))
}

// isDigits reports whether s is one or more ASCII digits.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// markRestore records where the screen is, for the decision just made. It is
// kept beside the result, not in it: the decision itself is unchanged.
func (m *searchHomeModel) markRestore() {
	m.restore = m.restorePoint(m.result)
}

// resultWithRestore is the decision with where the screen was attached.
func (m *searchHomeModel) resultWithRestore() SearchHomeResult {
	res := m.result
	res.Restore = m.restore
	return res
}

// restorePoint is what to reopen the screen on after res: the tab it was on
// and, when res came from a row, that row; when it came from the box, the
// query searched, since the box is where the user was.
func (m *searchHomeModel) restorePoint(res SearchHomeResult) HomeRestore {
	r := HomeRestore{Tab: tabName(m.scopes[m.scopeIdx].Value)}
	if m.onFavorites() {
		// enter opens the highlighted favorite, or the first match from the box.
		pos := m.cursor
		if m.focusInput {
			pos = 0
		}
		if pos >= 0 && pos < len(m.rows) {
			r.RowKey = m.rowKey(m.rows[pos])
		}
		return r
	}
	if row, ok := m.currentRow(); ok {
		r.RowKey = m.rowKey(row)
		return r
	}
	if res.Action == HomeSearch {
		r.Query = res.Query
	}
	return r
}
