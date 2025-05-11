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

type Anime struct {
	AnilistID     int
	AllanimeID    string
	EpisodeNumber int
	PlaybackTime  int
	Duration      int
	Title         string
	LastUpdated   time.Time
}

type LocalTracker struct {
	databaseFile string
	cache        map[string]Anime
	mu           sync.RWMutex
	writeMu      sync.Mutex
	writeRequest chan struct{}
	stopChan     chan struct{}
	timer        *time.Timer
}

func NewLocalTracker(databaseFile string) *LocalTracker {
	dir := filepath.Dir(databaseFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil
	}

	tracker := &LocalTracker{
		databaseFile: databaseFile,
		cache:        make(map[string]Anime),
		writeRequest: make(chan struct{}, 100),
		stopChan:     make(chan struct{}),
		timer:        time.NewTimer(0),
	}
	tracker.timer.Stop()

	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		if err := tracker.writeHeaders(); err != nil {
			return nil
		}
	}

	if entries, err := tracker.readAll(); err == nil {
		for _, entry := range entries {
			key := getCacheKey(entry.AnilistID, entry.AllanimeID)
			tracker.cache[key] = entry
		}
	}

	go tracker.backgroundWriter()
	return tracker
}

func (lt *LocalTracker) backgroundWriter() {
	const debounceTime = 100 * time.Millisecond

	for {
		select {
		case <-lt.writeRequest:
			lt.handleWriteRequest(debounceTime)

		case <-lt.timer.C:
			lt.handleTimerExpiration()

		case <-lt.stopChan:
			lt.handleStop()
			return
		}
	}
}

func (lt *LocalTracker) handleWriteRequest(debounceTime time.Duration) {
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	if !lt.timer.Stop() {
		select {
		case <-lt.timer.C:
		default:
		}
	}
	lt.timer.Reset(debounceTime)
}

func (lt *LocalTracker) handleTimerExpiration() {
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	if err := lt.flush(); err != nil {
		fmt.Printf("Erro na escrita: %v\n", err)
	}
}

func (lt *LocalTracker) handleStop() {
	lt.writeMu.Lock()
	defer lt.writeMu.Unlock()

	if !lt.timer.Stop() {
		select {
		case <-lt.timer.C:
		default:
		}
	}
	if err := lt.flush(); err != nil {
		fmt.Printf("Erro final: %v\n", err)
	}
}

func (lt *LocalTracker) flush() error {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	file, err := os.Create(lt.databaseFile)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		"anilist_id",
		"allanime_id",
		"episode_number",
		"playback_time",
		"duration",
		"title",
		"last_updated",
	}); err != nil {
		return err
	}

	for _, entry := range lt.cache {
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
			return err
		}
	}
	return nil
}

func (lt *LocalTracker) writeHeaders() error {
	file, err := os.Create(lt.databaseFile)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.Write([]string{
		"anilist_id",
		"allanime_id",
		"episode_number",
		"playback_time",
		"duration",
		"title",
		"last_updated",
	})
}

func (lt *LocalTracker) UpdateProgress(anime Anime) error {
	key := getCacheKey(anime.AnilistID, anime.AllanimeID)

	lt.mu.Lock()
	lt.cache[key] = anime
	lt.mu.Unlock()

	select {
	case lt.writeRequest <- struct{}{}:
	default:
	}

	return nil
}

func (lt *LocalTracker) GetAnime(anilistID int, allanimeID string) (*Anime, error) {
	key := getCacheKey(anilistID, allanimeID)

	lt.mu.RLock()
	defer lt.mu.RUnlock()

	if anime, exists := lt.cache[key]; exists {
		return &anime, nil
	}
	return nil, nil
}

func (lt *LocalTracker) GetAllAnime() ([]Anime, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	entries := make([]Anime, 0, len(lt.cache))
	for _, entry := range lt.cache {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (lt *LocalTracker) DeleteAnime(anilistID int, allanimeID string) error {
	key := getCacheKey(anilistID, allanimeID)

	lt.mu.Lock()
	delete(lt.cache, key)
	lt.mu.Unlock()

	select {
	case lt.writeRequest <- struct{}{}:
	default:
	}

	return nil
}

func (lt *LocalTracker) Close() {
	close(lt.stopChan)
}

func (lt *LocalTracker) readAll() ([]Anime, error) {
	file, err := os.Open(lt.databaseFile)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler CSV: %w", err)
	}

	var entries []Anime
	for i, record := range records {
		if i == 0 || len(record) < 7 {
			continue
		}

		anilistID, _ := strconv.Atoi(record[0])
		episodeNumber, _ := strconv.Atoi(record[2])
		playbackTime, _ := strconv.Atoi(record[3])
		duration, _ := strconv.Atoi(record[4])
		lastUpdated, _ := time.Parse(time.RFC3339, record[6])

		entries = append(entries, Anime{
			AnilistID:     anilistID,
			AllanimeID:    record[1],
			EpisodeNumber: episodeNumber,
			PlaybackTime:  playbackTime,
			Duration:      duration,
			Title:         record[5],
			LastUpdated:   lastUpdated,
		})
	}

	return entries, nil
}

func getCacheKey(anilistID int, allanimeID string) string {
	return fmt.Sprintf("%d|%s", anilistID, allanimeID)
}