package tracking

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// DBPath is where watch history and recent searches live. It follows HOME
// (USERPROFILE on Windows), like the log and download directories.
func DBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "GoAnime", "tracking", "progress.db")
	}
	return filepath.Join(home, ".local", "goanime", "tracking", "progress.db")
}

// sharedTracker opens (once, globally cached) the database for callers that
// only record or read, such as recent searches. It returns nil without SQLite
// (no CGO) or without a home directory.
func sharedTracker() *LocalTracker {
	if !IsCgoEnabled {
		return nil
	}
	if tracker := GetGlobalTracker(); tracker != nil {
		return tracker
	}
	path := DBPath()
	if path == "" {
		return nil
	}
	return NewLocalTracker(path)
}

// RecordSearch remembers a successful search and its result count for the
// SEARCH home screen's RECENT list. Best-effort: never fails the search.
func RecordSearch(query string, resultCount int) error {
	query = strings.Join(strings.Fields(query), " ")
	if query == "" {
		return nil
	}
	tracker := sharedTracker()
	if tracker == nil || tracker.db == nil || tracker.recentUpsertPS == nil {
		return nil
	}
	_, err := tracker.recentUpsertPS.Exec(query, resultCount, time.Now().Unix())
	return err
}

// RecordRecentSearch stores one search on a tracker. Nil-safe for tests.
func (t *LocalTracker) RecordRecentSearch(query string, resultCount int) error {
	if t == nil || t.db == nil || t.recentUpsertPS == nil {
		return ErrTrackerNotInited
	}
	query = strings.Join(strings.Fields(query), " ")
	if query == "" {
		return nil
	}
	_, err := t.recentUpsertPS.Exec(query, resultCount, time.Now().Unix())
	return err
}

// RecentSearches returns the latest successful searches, newest first.
// Empty without SQLite, without history, or beyond limit <= 0.
func RecentSearches(limit int) ([]RecentSearch, error) {
	tracker := sharedTracker()
	if tracker == nil {
		return nil, nil
	}
	return tracker.ListRecentSearches(limit)
}

// ListRecentSearches lists past searches, newest first. Nil-safe for tests.
func (t *LocalTracker) ListRecentSearches(limit int) ([]RecentSearch, error) {
	if t == nil || t.db == nil || t.recentAllPS == nil {
		return nil, ErrTrackerNotInited
	}
	if limit <= 0 {
		return nil, nil
	}
	rows, err := t.recentAllPS.Query(limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []RecentSearch
	for rows.Next() {
		var r RecentSearch
		var ts int64
		if err := rows.Scan(&r.Query, &r.ResultCount, &ts); err != nil {
			return nil, err
		}
		r.SearchedAt = time.Unix(ts, 0)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ContinueEntries returns every tracked episode, most recently watched
// first, for the SEARCH home screen's CONTINUE list. Rows watched to the
// end stay: they read as done (✓) and resume offers the next episode.
func ContinueEntries() ([]Anime, error) {
	tracker := sharedTracker()
	if tracker == nil {
		return nil, nil
	}
	entries, err := tracker.GetAllAnime()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].LastUpdated.After(entries[j].LastUpdated)
	})
	return entries, nil
}
