package tracking

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLocalTracker(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_tracker.db")

	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
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
	// Configuração inicial
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	defer func() {
		if err := tracker.Close(); err != nil {
			t.Logf("Error closing tracker: %v", err)
		}
	}()

	// Updated test data
	testAnime := Anime{
		AnilistID:     1,
		AllanimeID:    "allanime123",
		EpisodeNumber: 5,
		PlaybackTime:  120,
		Duration:      1500,
		Title:         "Test Anime",
		LastUpdated:   time.Now().UTC(), // Ensures current timestamp
	}

	// Teste de criação
	if err := tracker.UpdateProgress(testAnime); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Corrected verification
	retrieved, err := tracker.GetAnime(testAnime.AnilistID, testAnime.AllanimeID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Anime not found after update")
		return // This will never execute but satisfies linter
	}

	// Verifica todos os campos
	if retrieved.EpisodeNumber != testAnime.EpisodeNumber {
		t.Errorf("EpisodeNumber mismatch: got %d, want %d", retrieved.EpisodeNumber, testAnime.EpisodeNumber)
	}

	if retrieved.PlaybackTime != testAnime.PlaybackTime {
		t.Errorf("PlaybackTime mismatch: got %d, want %d", retrieved.PlaybackTime, testAnime.PlaybackTime)
	}

	if retrieved.Title != testAnime.Title {
		t.Errorf("Title mismatch: got %s, want %s", retrieved.Title, testAnime.Title)
	}

	// Tolerant timestamp verification (±2 seconds)
	now := time.Now().UTC()
	if retrieved.LastUpdated.After(now) || retrieved.LastUpdated.Before(now.Add(-2*time.Second)) {
		t.Errorf("LastUpdated out of range: got %v, expected ~%v", retrieved.LastUpdated, now)
	}
}

func TestLocalTracker_GetAnime(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_get_anime.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
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
		return // This will never execute but satisfies linter
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
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
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
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
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

func TestLocalTracker_EpisodeSpecificKeys(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_episode_keys.db")
	tracker := NewLocalTracker(dbPath)
	if tracker == nil {
		t.Skip("tracking unavailable (CGO/SQLite not enabled in this build)")
	}
	defer func() {
		if err := tracker.Close(); err != nil {
			t.Logf("Error closing tracker: %v", err)
		}
	}()

	// Simulate AllAnime where all episodes share the same base URL.
	// The tracking key should be "animeID:ep<N>" so each episode
	// gets its own row and does not overwrite other episodes.
	baseURL := "shared-anime-id-123"

	ep3 := Anime{
		AnilistID:     100,
		AllanimeID:    baseURL + ":ep3",
		EpisodeNumber: 3,
		PlaybackTime:  900,
		Duration:      1440,
		Title:         "Breaking Bad S2E3",
		LastUpdated:   time.Now(),
	}
	ep4 := Anime{
		AnilistID:     100,
		AllanimeID:    baseURL + ":ep4",
		EpisodeNumber: 4,
		PlaybackTime:  450,
		Duration:      1440,
		Title:         "Breaking Bad S2E4",
		LastUpdated:   time.Now(),
	}

	if err := tracker.UpdateProgress(ep3); err != nil {
		t.Fatalf("UpdateProgress ep3: %v", err)
	}
	if err := tracker.UpdateProgress(ep4); err != nil {
		t.Fatalf("UpdateProgress ep4: %v", err)
	}

	// Each episode should have its own independent tracking entry
	got3, err := tracker.GetAnime(100, baseURL+":ep3")
	if err != nil {
		t.Fatalf("GetAnime ep3: %v", err)
	}
	if got3 == nil {
		t.Fatal("ep3 tracking not found")
	}
	if got3.EpisodeNumber != 3 || got3.PlaybackTime != 900 {
		t.Errorf("ep3 mismatch: got episode=%d time=%d, want episode=3 time=900",
			got3.EpisodeNumber, got3.PlaybackTime)
	}

	got4, err := tracker.GetAnime(100, baseURL+":ep4")
	if err != nil {
		t.Fatalf("GetAnime ep4: %v", err)
	}
	if got4 == nil {
		t.Fatal("ep4 tracking not found")
	}
	if got4.EpisodeNumber != 4 || got4.PlaybackTime != 450 {
		t.Errorf("ep4 mismatch: got episode=%d time=%d, want episode=4 time=450",
			got4.EpisodeNumber, got4.PlaybackTime)
	}

	// Looking up with the bare base URL should NOT return either episode's data
	gotBare, err := tracker.GetAnime(100, baseURL)
	if err != nil {
		t.Fatalf("GetAnime bare: %v", err)
	}
	if gotBare != nil {
		t.Errorf("bare base URL should not match any episode-specific entry, got: %+v", gotBare)
	}
}
