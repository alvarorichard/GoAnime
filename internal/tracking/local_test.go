package tracking

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLocalTracker(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_tracker.db")

	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}

	// Check if DB file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("DB file was not created: %v", err)
	}

	// Check if the tracker can be closed without error
	if err := tracker.Close(); err != nil {
		t.Errorf("tracker.Close() returned error: %v", err)
	}
}

func TestLocalTracker_UpdateProgress(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_update_progress.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}
	defer func(tracker *LocalTracker) {
		err := tracker.Close()
		if err != nil {
			t.Fatalf("tracker.Close() returned error: %v", err)
		}
	}(tracker)

	anime := Anime{
		AnilistID:     123,
		AllanimeID:    "abc",
		EpisodeNumber: 5,
		PlaybackTime:  120,
		Duration:      1440,
		Title:         "Test Anime",
	}

	err := tracker.UpdateProgress(anime)
	if err != nil {
		t.Fatalf("UpdateProgress returned error: %v", err)
	}

	// Retrieve and check
	got, err := tracker.GetAnime(anime.AnilistID, anime.AllanimeID)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}
	if got == nil {
		t.Fatal("GetAnime returned nil after UpdateProgress")
	}
	if got.EpisodeNumber != anime.EpisodeNumber || got.PlaybackTime != anime.PlaybackTime || got.Duration != anime.Duration || got.Title != anime.Title {
		t.Errorf("Anime fields do not match after UpdateProgress: got %+v, want %+v", got, anime)
	}
	if got.LastUpdated.IsZero() {
		t.Error("LastUpdated was not set")
	}
}

func TestLocalTracker_GetAnime(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_get_anime.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}
	defer func(tracker *LocalTracker) {
		err := tracker.Close()
		if err != nil {
			t.Fatalf("tracker.Close() returned error: %v", err)
		}
	}(tracker)

	// Should return nil for non-existent anime
	got, err := tracker.GetAnime(999, "notfound")
	if err != nil {
		t.Fatalf("GetAnime returned error for non-existent: %v", err)
	}
	if got != nil {
		t.Errorf("GetAnime should return nil for non-existent anime, got: %+v", got)
	}

	// Insert and retrieve
	anime := Anime{
		AnilistID:     321,
		AllanimeID:    "def",
		EpisodeNumber: 2,
		PlaybackTime:  60,
		Duration:      600,
		Title:         "Another Test",
	}
	err = tracker.UpdateProgress(anime)
	if err != nil {
		t.Fatalf("UpdateProgress returned error: %v", err)
	}
	got, err = tracker.GetAnime(anime.AnilistID, anime.AllanimeID)
	if err != nil {
		t.Fatalf("GetAnime returned error: %v", err)
	}
	if got == nil {
		t.Fatal("GetAnime returned nil after insert")
	}
	if got.EpisodeNumber != anime.EpisodeNumber || got.PlaybackTime != anime.PlaybackTime || got.Duration != anime.Duration || got.Title != anime.Title {
		t.Errorf("Anime fields do not match after GetAnime: got %+v, want %+v", got, anime)
	}
}
func TestLocalTracker_GetAllAnime(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_get_all_anime.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}
	defer func(tracker *LocalTracker) {
		err := tracker.Close()
		if err != nil {
			t.Fatalf("tracker.Close() returned error: %v", err)
		}
	}(tracker)

	// Initially, should be empty
	all, err := tracker.GetAllAnime()
	if err != nil {
		t.Fatalf("GetAllAnime returned error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("Expected 0 anime, got %d", len(all))
	}

	// Insert some anime
	anime1 := Anime{
		AnilistID:     1,
		AllanimeID:    "a1",
		EpisodeNumber: 1,
		PlaybackTime:  10,
		Duration:      100,
		Title:         "Anime One",
	}
	anime2 := Anime{
		AnilistID:     2,
		AllanimeID:    "a2",
		EpisodeNumber: 2,
		PlaybackTime:  20,
		Duration:      200,
		Title:         "Anime Two",
	}
	if err := tracker.UpdateProgress(anime1); err != nil {
		t.Fatalf("UpdateProgress anime1 error: %v", err)
	}
	if err := tracker.UpdateProgress(anime2); err != nil {
		t.Fatalf("UpdateProgress anime2 error: %v", err)
	}

	all, err = tracker.GetAllAnime()
	if err != nil {
		t.Fatalf("GetAllAnime returned error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("Expected 2 anime, got %d", len(all))
	}
	// Optionally, check contents
	found := map[string]bool{}
	for _, a := range all {
		found[a.AllanimeID] = true
	}
	if !found["a1"] || !found["a2"] {
		t.Errorf("GetAllAnime missing expected anime: %+v", all)
	}
}

func TestLocalTracker_DeleteAnime(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_delete_anime.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Fatal("NewLocalTracker returned nil")
	}
	defer func(tracker *LocalTracker) {
		err := tracker.Close()
		if err != nil {
			t.Fatalf("tracker.Close() returned error: %v", err)
		}
	}(tracker)

	anime := Anime{
		AnilistID:     10,
		AllanimeID:    "delme",
		EpisodeNumber: 3,
		PlaybackTime:  30,
		Duration:      300,
		Title:         "Delete Me",
	}
	if err := tracker.UpdateProgress(anime); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	// Confirm exists
	got, err := tracker.GetAnime(anime.AnilistID, anime.AllanimeID)
	if err != nil {
		t.Fatalf("GetAnime error: %v", err)
	}
	if got == nil {
		t.Fatal("Anime should exist before deletion")
	}

	// Delete
	if err := tracker.DeleteAnime(anime.AnilistID, anime.AllanimeID); err != nil {
		t.Fatalf("DeleteAnime error: %v", err)
	}

	// Confirm deleted
	got, err = tracker.GetAnime(anime.AnilistID, anime.AllanimeID)
	if err != nil {
		t.Fatalf("GetAnime after delete error: %v", err)
	}
	if got != nil {
		t.Error("Anime was not deleted")
	}
}

