package providers_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/models"
)

func TestResolveSourceName_BySourceField(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"Goyabu", "goyabu"},
		{"Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			anime := &models.Anime{Source: tt.source}
			got := providers.ResolveSourceName(anime)
			if got != tt.expected {
				t.Errorf("ResolveSourceName(Source=%q) = %q, want %q", tt.source, got, tt.expected)
			}
		})
	}
}

func TestResolveSourceName_ByTags(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"[Unknown] Naruto", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{Name: tt.name}
			got := providers.ResolveSourceName(anime)
			if got != tt.expected {
				t.Errorf("ResolveSourceName(Name=%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestResolveSourceName_PTBR(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"[PT-BR] Dragon Ball", "https://goyabu.com/video/123", "goyabu"},
		{"[PT-BR] Naruto", "https://animefire.plus/video/123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anime := &models.Anime{Name: tt.name, URL: tt.url}
			got := providers.ResolveSourceName(anime)
			if got != tt.expected {
				t.Errorf("ResolveSourceName(Name=%q, URL=%q) = %q, want %q", tt.name, tt.url, got, tt.expected)
			}
		})
	}
}

func TestForSource_Unmigrated(t *testing.T) {
	anime := &models.Anime{Source: "AllAnime"}
	p := providers.ForSource(anime)
	if p != nil {
		t.Errorf("ForSource(AllAnime) = %v, want nil (not migrated yet)", p)
	}
}

func TestForSource_Migrated_Goyabu(t *testing.T) {
	anime := &models.Anime{Source: "Goyabu"}
	p := providers.ForSource(anime)
	if p == nil {
		t.Fatal("ForSource(Goyabu) = nil, want non-nil provider")
	}
	if p.Name() != "Goyabu" {
		t.Errorf("Provider.Name() = %q, want %q", p.Name(), "Goyabu")
	}
}
