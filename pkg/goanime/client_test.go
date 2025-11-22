package goanime_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func TestNewClient(t *testing.T) {
	client := goanime.NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestGetAvailableSources(t *testing.T) {
	client := goanime.NewClient()
	sources := client.GetAvailableSources()

	if len(sources) == 0 {
		t.Fatal("No sources available")
	}

	// Check if expected sources are present
	hasAllAnime := false
	hasAnimeFire := false

	for _, source := range sources {
		if source == types.SourceAllAnime {
			hasAllAnime = true
		}
		if source == types.SourceAnimeFire {
			hasAnimeFire = true
		}
	}

	if !hasAllAnime {
		t.Error("AllAnime source not found")
	}
	if !hasAnimeFire {
		t.Error("AnimeFire source not found")
	}
}

func TestSourceString(t *testing.T) {
	tests := []struct {
		source   types.Source
		expected string
	}{
		{types.SourceAllAnime, "AllAnime"},
		{types.SourceAnimeFire, "AnimeFire"},
	}

	for _, tt := range tests {
		if got := tt.source.String(); got != tt.expected {
			t.Errorf("Source.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestParseSource(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Source
		hasError bool
	}{
		{"AllAnime", types.SourceAllAnime, false},
		{"allanime", types.SourceAllAnime, false},
		{"all", types.SourceAllAnime, false},
		{"AnimeFire", types.SourceAnimeFire, false},
		{"animefire", types.SourceAnimeFire, false},
		{"fire", types.SourceAnimeFire, false},
		{"invalid", types.SourceAllAnime, true},
	}

	for _, tt := range tests {
		got, err := types.ParseSource(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseSource(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseSource(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseSource(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		}
	}
}

// Integration test - requires network access
func TestSearchAnime_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := goanime.NewClient()

	// Test search across all sources
	results, err := client.SearchAnime("Naruto", nil)
	if err != nil {
		t.Fatalf("SearchAnime failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("SearchAnime returned no results")
	}

	// Verify result structure
	anime := results[0]
	if anime.Name == "" {
		t.Error("Anime name is empty")
	}
	if anime.URL == "" {
		t.Error("Anime URL is empty")
	}
	if anime.Source == "" {
		t.Error("Anime source is empty")
	}
}

// Integration test for specific source
func TestSearchAnimeSpecificSource_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := goanime.NewClient()
	source := types.SourceAllAnime

	results, err := client.SearchAnime("One Piece", &source)
	if err != nil {
		t.Fatalf("SearchAnime with specific source failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("SearchAnime returned no results for specific source")
	}

	// All results should be from the specified source
	for _, anime := range results {
		if anime.Source != source.String() {
			t.Errorf("Expected source %s, got %s", source.String(), anime.Source)
		}
	}
}
