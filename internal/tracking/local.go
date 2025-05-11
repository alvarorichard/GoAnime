package tracking

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Anime represents the structure for tracking anime progress
type Anime struct {
	AnilistID     int
	AllanimeID    string
	EpisodeNumber int
	PlaybackTime  int
	Duration      int
	Title         string
	LastUpdated   time.Time
}

// LocalTracker handles local tracking of anime progress
type LocalTracker struct {
	databaseFile string
	cache        map[string]Anime
	mu           sync.RWMutex
	writeMu      sync.Mutex
	writeQueue   chan Anime
	stopChan     chan struct{}
	forceWrite   chan struct{} // Channel to force immediate write
}

// NewLocalTracker creates a new instance of LocalTracker
func NewLocalTracker(databaseFile string) *LocalTracker {
	// Ensure the directory exists
	dir := filepath.Dir(databaseFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil
	}

	tracker := &LocalTracker{
		databaseFile: databaseFile,
		cache:        make(map[string]Anime),
		writeQueue:   make(chan Anime, 100),
		stopChan:     make(chan struct{}),
		forceWrite:   make(chan struct{}, 1),
	}

	// Create file with headers if it doesn't exist
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		if err := tracker.writeHeaders(); err != nil {
			return nil
		}
	}

	// Load existing data into cache
	if entries, err := tracker.readAll(); err == nil {
		for _, entry := range entries {
			key := getCacheKey(entry.AnilistID, entry.AllanimeID)
			tracker.cache[key] = entry
		}
	}

	// Start background writer
	go tracker.backgroundWriter()

	return tracker
}

// writeHeaders writes the CSV headers to the file
func (lt *LocalTracker) writeHeaders() error {
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	file, err := os.Create(lt.databaseFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"anilist_id", "allanime_id", "episode_number", "playback_time", "duration", "title", "last_updated"}
	return writer.Write(headers)
}

// backgroundWriter handles writing updates to the file
func (lt *LocalTracker) backgroundWriter() {
	ticker := time.NewTicker(1 * time.Second) // Reduced to 1 second
	defer ticker.Stop()

	var pendingUpdates []Anime
	lastWrite := time.Now()

	for {
		select {
		case anime := <-lt.writeQueue:
			pendingUpdates = append(pendingUpdates, anime)
			// If we have more than 5 updates or it's been more than 2 seconds, write immediately
			if len(pendingUpdates) >= 5 || time.Since(lastWrite) > 2*time.Second {
				lt.writeMu.Lock()
				if err := lt.writeAll(pendingUpdates); err != nil {
					fmt.Printf("Failed to write updates: %v\n", err)
				}
				lt.writeMu.Unlock()
				pendingUpdates = nil
				lastWrite = time.Now()
			}
		case <-ticker.C:
			if len(pendingUpdates) > 0 {
				lt.writeMu.Lock()
				if err := lt.writeAll(pendingUpdates); err != nil {
					fmt.Printf("Failed to write updates: %v\n", err)
				}
				lt.writeMu.Unlock()
				pendingUpdates = nil
				lastWrite = time.Now()
			}
		case <-lt.forceWrite:
			if len(pendingUpdates) > 0 {
				lt.writeMu.Lock()
				if err := lt.writeAll(pendingUpdates); err != nil {
					fmt.Printf("Failed to write updates: %v\n", err)
				}
				lt.writeMu.Unlock()
				pendingUpdates = nil
				lastWrite = time.Now()
			}
		case <-lt.stopChan:
			// Final write before stopping
			if len(pendingUpdates) > 0 {
				lt.writeMu.Lock()
				_ = lt.writeAll(pendingUpdates)
				lt.writeMu.Unlock()
			}
			return
		}
	}
}

// getCacheKey returns a unique key for an anime entry
func getCacheKey(anilistID int, allanimeID string) string {
	return fmt.Sprintf("%d:%s", anilistID, allanimeID)
}

// UpdateProgress updates the progress of an anime in the local database
func (lt *LocalTracker) UpdateProgress(anime Anime) error {
	lt.mu.Lock()
	key := getCacheKey(anime.AnilistID, anime.AllanimeID)
	lt.cache[key] = anime
	lt.mu.Unlock()

	// Queue the update for writing
	select {
	case lt.writeQueue <- anime:
		// Try to force an immediate write if possible
		select {
		case lt.forceWrite <- struct{}{}:
		default:
		}
	default:
		// If queue is full, force an immediate write
		lt.writeMu.Lock()
		if err := lt.writeAll([]Anime{anime}); err != nil {
			lt.writeMu.Unlock()
			return fmt.Errorf("failed to write update: %w", err)
		}
		lt.writeMu.Unlock()
	}

	return nil
}

// GetAnime retrieves a specific anime entry by Anilist ID and Allanime ID
func (lt *LocalTracker) GetAnime(anilistID int, allanimeID string) (*Anime, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	key := getCacheKey(anilistID, allanimeID)
	if anime, exists := lt.cache[key]; exists {
		return &anime, nil
	}

	return nil, nil
}

// GetAllAnime retrieves all anime entries from the local database
func (lt *LocalTracker) GetAllAnime() ([]Anime, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	entries := make([]Anime, 0, len(lt.cache))
	for _, entry := range lt.cache {
		entries = append(entries, entry)
	}

	return entries, nil
}

// DeleteAnime removes an anime entry from the local database
func (lt *LocalTracker) DeleteAnime(anilistID int, allanimeID string) error {
	lt.mu.Lock()
	key := getCacheKey(anilistID, allanimeID)
	delete(lt.cache, key)
	lt.mu.Unlock()

	// Force an immediate write to update the file
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	entries, err := lt.GetAllAnime()
	if err != nil {
		return err
	}

	return lt.writeAll(entries)
}

// Close stops the background writer and performs final writes
func (lt *LocalTracker) Close() {
	// Force a final write before closing
	lt.forceWrite <- struct{}{}
	close(lt.stopChan)
}

// readAll reads all entries from the CSV file
func (lt *LocalTracker) readAll() ([]Anime, error) {
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	file, err := os.OpenFile(lt.databaseFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Skip header row
	if len(records) <= 1 {
		return nil, nil
	}

	var entries []Anime
	for _, record := range records[1:] {
		if len(record) < 6 {
			continue
		}

		anilistID, _ := strconv.Atoi(record[0])
		episodeNumber, _ := strconv.Atoi(record[2])
		playbackTime, _ := strconv.Atoi(record[3])
		duration, _ := strconv.Atoi(record[4])

		anime := Anime{
			AnilistID:     anilistID,
			AllanimeID:    record[1],
			EpisodeNumber: episodeNumber,
			PlaybackTime:  playbackTime,
			Duration:      duration,
			Title:         record[5],
		}

		// Parse last updated time if available
		if len(record) > 6 {
			if lastUpdated, err := time.Parse(time.RFC3339, record[6]); err == nil {
				anime.LastUpdated = lastUpdated
			}
		}

		entries = append(entries, anime)
	}

	return entries, nil
}

// writeAll writes all entries to the CSV file
func (lt *LocalTracker) writeAll(entries []Anime) error {
	file, err := os.Create(lt.databaseFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"anilist_id", "allanime_id", "episode_number", "playback_time", "duration", "title", "last_updated"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		record := []string{
			strconv.Itoa(entry.AnilistID),
			entry.AllanimeID,
			strconv.Itoa(entry.EpisodeNumber),
			strconv.Itoa(entry.PlaybackTime),
			strconv.Itoa(entry.Duration),
			entry.Title,
			entry.LastUpdated.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}