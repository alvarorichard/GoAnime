package tracking

import (
	"database/sql"
	"errors"
	"strings"
)

// ErrHistoryUnavailable is what clearing history reports when there is no
// history to clear: the build has no SQLite (no CGO) or there is no home
// directory to keep it in.
var ErrHistoryUnavailable = errors.New("watch history is unavailable (no SQLite in this build, or no home directory)")

// normalizeQuery is a search as it is stored: words joined by single spaces.
func normalizeQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

// affected is how many rows a statement touched.
func affected(res sql.Result, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

// DeleteRecentSearch forgets one past search and reports whether it was there.
// Nil-safe for tests.
func (t *LocalTracker) DeleteRecentSearch(query string) (bool, error) {
	if t == nil || t.db == nil {
		return false, ErrTrackerNotInited
	}
	n, err := affected(t.db.Exec(`DELETE FROM recent_searches WHERE query = ?`, normalizeQuery(query)))
	return n > 0, err
}

// ClearRecentSearches forgets every past search and returns how many there
// were. Nil-safe for tests.
func (t *LocalTracker) ClearRecentSearches() (int, error) {
	if t == nil || t.db == nil {
		return 0, ErrTrackerNotInited
	}
	return affected(t.db.Exec(`DELETE FROM recent_searches`))
}

// ClearProgress forgets every watched episode, which is the whole CONTINUE
// list, and returns how many rows there were. Nil-safe for tests.
func (t *LocalTracker) ClearProgress() (int, error) {
	if t == nil || t.db == nil {
		return 0, ErrTrackerNotInited
	}
	return affected(t.db.Exec(`DELETE FROM media_progress`))
}

// ForgetSearch removes one search from the SEARCH home screen's RECENT list.
// A search that is not there is not an error.
func ForgetSearch(query string) error {
	tracker := sharedTracker()
	if tracker == nil {
		return ErrHistoryUnavailable
	}
	_, err := tracker.DeleteRecentSearch(query)
	return err
}

// ClearRecent empties the RECENT list and returns how many searches it held.
func ClearRecent() (int, error) {
	tracker := sharedTracker()
	if tracker == nil {
		return 0, ErrHistoryUnavailable
	}
	return tracker.ClearRecentSearches()
}

// ClearContinue empties the CONTINUE list and returns how many episodes it
// held.
func ClearContinue() (int, error) {
	tracker := sharedTracker()
	if tracker == nil {
		return 0, ErrHistoryUnavailable
	}
	return tracker.ClearProgress()
}
