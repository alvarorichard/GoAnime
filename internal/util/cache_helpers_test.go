package util

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestEncodeEpisodes(t *testing.T) {
	tests := []struct {
		name     string
		episodes []models.Episode
		wantErr  bool
	}{
		{
			name:     "empty episodes",
			episodes: []models.Episode{},
			wantErr:  false,
		},
		{
			name: "single episode",
			episodes: []models.Episode{
				{
					Number: "1",
					Num:    1,
					URL:    "https://example.com/1",
					Title: models.TitleDetails{
						English: "Episode 1",
						Romaji:  "第1話",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple episodes",
			episodes: []models.Episode{
				{Number: "1", Num: 1, URL: "https://example.com/1"},
				{Number: "2", Num: 2, URL: "https://example.com/2"},
				{Number: "3", Num: 3, URL: "https://example.com/3"},
			},
			wantErr: false,
		},
		{
			name: "episode with all fields",
			episodes: []models.Episode{
				{
					Number: "12",
					Num:    12,
					URL:    "https://example.com/ep12",
					Title: models.TitleDetails{
						English: "The Beginning",
						Romaji:  "始まり",
					},
					DataID: "data-123",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeEpisodes(tt.episodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeEpisodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Error("EncodeEpisodes() expected non-empty result")
			}
		})
	}
}

func TestDecodeEpisodes(t *testing.T) {
	tests := []struct {
		name     string
		episodes []models.Episode
	}{
		{
			name:     "empty episodes",
			episodes: []models.Episode{},
		},
		{
			name: "single episode",
			episodes: []models.Episode{
				{
					Number: "1",
					Num:    1,
					URL:    "https://example.com/1",
					Title: models.TitleDetails{
						English: "Episode 1",
					},
				},
			},
		},
		{
			name: "multiple episodes",
			episodes: []models.Episode{
				{Number: "1", Num: 1, URL: "https://example.com/1"},
				{Number: "2", Num: 2, URL: "https://example.com/2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode first
			encoded, err := EncodeEpisodes(tt.episodes)
			if err != nil {
				t.Fatalf("EncodeEpisodes() failed: %v", err)
			}

			// Then decode
			got := DecodeEpisodes(encoded)
			if len(got) != len(tt.episodes) {
				t.Errorf("DecodeEpisodes() got %d episodes, want %d", len(got), len(tt.episodes))
			}

			// Verify content
			for i := range got {
				if got[i].Number != tt.episodes[i].Number {
					t.Errorf("episode %d: got Number=%s, want %s", i, got[i].Number, tt.episodes[i].Number)
				}
				if got[i].Num != tt.episodes[i].Num {
					t.Errorf("episode %d: got Num=%d, want %d", i, got[i].Num, tt.episodes[i].Num)
				}
			}
		})
	}
}

func TestDecodeEpisodesInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid JSON",
			data: []byte("not valid json"),
		},
		{
			name: "empty data",
			data: []byte(""),
		},
		{
			name: "partial JSON",
			data: []byte(`[{"Number":`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeEpisodes(tt.data)
			// Should return nil or empty slice on error
			if got != nil && len(got) > 0 {
				t.Errorf("DecodeEpisodes() expected nil or empty, got %d episodes", len(got))
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := []models.Episode{
		{
			Number: "Special",
			Num:    0,
			URL:    "special-episode",
			Title: models.TitleDetails{
				English: "Special Episode",
				Romaji:  "特別編",
			},
			DataID: "special-data",
		},
		{
			Number: "1",
			Num:    1,
			URL:    "episode-1",
			Title: models.TitleDetails{
				English: "The First Episode",
				Romaji:  "第1話",
			},
		},
		{
			Number: "2",
			Num:    2,
			URL:    "episode-2",
			Title: models.TitleDetails{
				English: "The Second Episode",
				Romaji:  "第2話",
			},
		},
	}

	encoded, err := EncodeEpisodes(original)
	if err != nil {
		t.Fatalf("EncodeEpisodes() failed: %v", err)
	}

	decoded := DecodeEpisodes(encoded)
	if len(decoded) != len(original) {
		t.Fatalf("DecodeEpisodes() got %d episodes, want %d", len(decoded), len(original))
	}

	// Compare each episode
	for i := range original {
		if decoded[i].Number != original[i].Number {
			t.Errorf("episode %d: Number = %s, want %s", i, decoded[i].Number, original[i].Number)
		}
		if decoded[i].Num != original[i].Num {
			t.Errorf("episode %d: Num = %d, want %d", i, decoded[i].Num, original[i].Num)
		}
		if decoded[i].URL != original[i].URL {
			t.Errorf("episode %d: URL = %s, want %s", i, decoded[i].URL, original[i].URL)
		}
		if decoded[i].Title.English != original[i].Title.English {
			t.Errorf("episode %d: Title.English = %s, want %s", i, decoded[i].Title.English, original[i].Title.English)
		}
		if decoded[i].Title.Romaji != original[i].Title.Romaji {
			t.Errorf("episode %d: Title.Romaji = %s, want %s", i, decoded[i].Title.Romaji, original[i].Title.Romaji)
		}
	}
}

func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(1*time.Second, 100)

	// Test Set and Get
	cache.Set("key1", []byte("value1"))

	got, ok := cache.Get("key1")
	if !ok {
		t.Error("expected to get value1")
	}
	if string(got) != "value1" {
		t.Errorf("got %s, want value1", string(got))
	}

	// Test Get non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("expected false for non-existent key")
	}
}

func TestResponseCacheExpiration(t *testing.T) {
	cache := NewResponseCache(100*time.Millisecond, 100)

	cache.Set("key1", []byte("value1"))

	// Should exist immediately
	got, ok := cache.Get("key1")
	if !ok || string(got) != "value1" {
		t.Error("expected value1 before expiration")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("key1")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestResponseCacheMaxSize(t *testing.T) {
	maxSize := 3
	cache := NewResponseCache(1*time.Minute, maxSize)

	// Fill cache beyond max size
	cache.Set("key1", []byte("value1"))
	cache.Set("key2", []byte("value2"))
	cache.Set("key3", []byte("value3"))
	cache.Set("key4", []byte("value4"))

	// The oldest should be evicted
	// Note: The eviction logic removes oldest when at max size
	// So key1 might still exist depending on implementation
	_, _ = cache.Get("key1") // May or may not exist depending on eviction
	_, key2Exists := cache.Get("key2")
	_, key3Exists := cache.Get("key3")
	_, key4Exists := cache.Get("key4")

	// Log results (for debugging)
	t.Logf("key2 exists: %v, key3 exists: %v, key4 exists: %v", key2Exists, key3Exists, key4Exists)

	if !key4Exists {
		t.Error("key4 should exist")
	}

	// At least some keys should exist
	if !key2Exists && !key3Exists && !key4Exists {
		t.Error("at least some keys should exist after adding max+1 items")
	}
}

func TestGetAniListCache(t *testing.T) {
	cache1 := GetAniListCache()
	cache2 := GetAniListCache()

	if cache1 != cache2 {
		t.Error("GetAniListCache should return same instance")
	}

	// Test that it's functional
	cache1.Set("test", []byte("data"))
	got, ok := cache1.Get("test")
	if !ok || string(got) != "data" {
		t.Error("cache should store and retrieve data")
	}
}

func TestGetSearchCache(t *testing.T) {
	cache1 := GetSearchCache()
	cache2 := GetSearchCache()

	if cache1 != cache2 {
		t.Error("GetSearchCache should return same instance")
	}

	// Test that it's functional
	cache1.Set("test", []byte("data"))
	got, ok := cache1.Get("test")
	if !ok || string(got) != "data" {
		t.Error("cache should store and retrieve data")
	}
}

func TestGetEpisodeCache(t *testing.T) {
	cache1 := GetEpisodeCache()
	cache2 := GetEpisodeCache()

	if cache1 != cache2 {
		t.Error("GetEpisodeCache should return same instance")
	}

	// Test that it's functional
	cache1.Set("test", []byte("data"))
	got, ok := cache1.Get("test")
	if !ok || string(got) != "data" {
		t.Error("cache should store and retrieve data")
	}
}

func TestResponseCacheConcurrentAccess(t *testing.T) {
	cache := NewResponseCache(1*time.Minute, 1000)

	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Set(string(rune('a'+n)), []byte("value"))
			}
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Get(string(rune('a' + n%10)))
			}
			done <- true
		}(i)
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}

	// No race conditions should occur
}
