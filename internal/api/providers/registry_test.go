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
		{"AllAnime", "allanime"},
		{"Animefire.io", "animefire"},
		{"AnimeDrive", "animedrive"},
		{"FlixHQ", "flixhq"},
		{"9Anime", "9anime"},
		{"Goyabu", "goyabu"},
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
		{"[English] Naruto", "allanime"},
		{"[Multilanguage] One Piece", "9anime"},
		{"[Movie] Avengers", "flixhq"},
		{"[TV] Breaking Bad", "flixhq"},
		{"[Movies/TV] Spider-Man", "flixhq"},
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

func TestResolveSourceName_ByURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://animefire.io/video/naruto", "animefire"},
		{"https://animesdrive.com/naruto-ep-1", "animedrive"},
		{"https://goyabu.to/naruto", "goyabu"},
		{"https://flixhq.to/movie/avengers-12345", "flixhq"},
		{"https://allanime.to/anime/naruto", "allanime"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			anime := &models.Anime{URL: tt.url}
			got := providers.ResolveSourceName(anime)
			if got != tt.expected {
				t.Errorf("ResolveSourceName(URL=%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

func TestResolveSourceName_AllAnimeIDFallback(t *testing.T) {
	anime := &models.Anime{URL: "abc123XYZ"}
	got := providers.ResolveSourceName(anime)
	if got != "allanime" {
		t.Errorf("ResolveSourceName(URL=%q) = %q, want %q", anime.URL, got, "allanime")
	}
}

func TestResolveSourceName_Priority_SourceOverTags(t *testing.T) {
	anime := &models.Anime{
		Source: "AnimeDrive",
		Name:   "[English] Some Anime",
		URL:    "https://animefire.io/video/test",
	}
	got := providers.ResolveSourceName(anime)
	if got != "animedrive" {
		t.Errorf("Source field should take priority, got %q want %q", got, "animedrive")
	}
}

func TestResolveSourceName_Priority_TagsOverURL(t *testing.T) {
	anime := &models.Anime{
		Name: "[English] Some Anime",
		URL:  "https://animefire.io/video/test",
	}
	got := providers.ResolveSourceName(anime)
	if got != "allanime" {
		t.Errorf("[English] tag should resolve to allanime, got %q", got)
	}
}

func TestResolveSourceName_PTBR_DisambiguationByURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"[PT-BR] Naruto", "https://animesdrive.com/naruto", "animedrive"},
		{"[PT-BR] Naruto", "https://goyabu.to/naruto", "goyabu"},
		{"[PT-BR] Naruto", "https://animefire.io/naruto", "animefire"},
		{"[PT-BR] Naruto", "https://unknown.com/naruto", "animefire"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			anime := &models.Anime{Name: tt.name, URL: tt.url}
			got := providers.ResolveSourceName(anime)
			if got != tt.expected {
				t.Errorf("ResolveSourceName(Name=%q, URL=%q) = %q, want %q", tt.name, tt.url, got, tt.expected)
			}
		})
	}
}

func TestResolveSourceName_MediaType(t *testing.T) {
	anime := &models.Anime{MediaType: models.MediaTypeMovie}
	got := providers.ResolveSourceName(anime)
	if got != "flixhq" {
		t.Errorf("MediaTypeMovie should resolve to flixhq, got %q", got)
	}

	anime = &models.Anime{MediaType: models.MediaTypeTV}
	got = providers.ResolveSourceName(anime)
	if got != "flixhq" {
		t.Errorf("MediaTypeTV should resolve to flixhq, got %q", got)
	}
}

func TestResolveSourceName_NilAnime(t *testing.T) {
	got := providers.ResolveSourceName(nil)
	if got != "allanime" {
		t.Errorf("nil anime should fallback to allanime, got %q", got)
	}
}

func TestForSource_NeverReturnsNil(t *testing.T) {
	cases := []*models.Anime{
		nil,
		{},
		{Source: "UnknownSource"},
		{URL: "something-totally-random"},
		{Source: "AllAnime"},
		{Source: "FlixHQ"},
		{Source: "9Anime"},
	}

	for _, anime := range cases {
		p := providers.ForSource(anime)
		if p == nil {
			t.Errorf("ForSource(%+v) returned nil", anime)
		}
	}
}

func TestForSource_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		anime        *models.Anime
		expectedName string
	}{
		{&models.Anime{Source: "AllAnime"}, "AllAnime"},
		{&models.Anime{Source: "Animefire.io"}, "Animefire.io"},
		{&models.Anime{Source: "AnimeDrive"}, "AnimeDrive"},
		{&models.Anime{Source: "FlixHQ"}, "FlixHQ"},
		{&models.Anime{Source: "9Anime"}, "9Anime"},
		{&models.Anime{Source: "Goyabu"}, "Goyabu"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			p := providers.ForSource(tt.anime)
			if p.Name() != tt.expectedName {
				t.Errorf("ForSource(Source=%q).Name() = %q, want %q", tt.anime.Source, p.Name(), tt.expectedName)
			}
		})
	}
}

func TestForSourceName(t *testing.T) {
	tests := []struct {
		source       string
		expectedName string
	}{
		{"allanime", "AllAnime"},
		{"animefire", "Animefire.io"},
		{"animedrive", "AnimeDrive"},
		{"flixhq", "FlixHQ"},
		{"9anime", "9Anime"},
		{"goyabu", "Goyabu"},
		{"", "AllAnime"},
		{"unknown", "AllAnime"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			p := providers.ForSourceName(tt.source)
			if p.Name() != tt.expectedName {
				t.Errorf("ForSourceName(%q).Name() = %q, want %q", tt.source, p.Name(), tt.expectedName)
			}
		})
	}
}

func TestConvenienceHelpers(t *testing.T) {
	if !providers.IsAllAnime(&models.Anime{Source: "AllAnime"}) {
		t.Error("IsAllAnime should return true for Source=AllAnime")
	}
	if !providers.IsAnimeFire(&models.Anime{Source: "Animefire.io"}) {
		t.Error("IsAnimeFire should return true for Source=Animefire.io")
	}
	if !providers.IsAnimeDrive(&models.Anime{Source: "AnimeDrive"}) {
		t.Error("IsAnimeDrive should return true for Source=AnimeDrive")
	}
	if !providers.IsFlixHQ(&models.Anime{Source: "FlixHQ"}) {
		t.Error("IsFlixHQ should return true for Source=FlixHQ")
	}
	if !providers.Is9Anime(&models.Anime{Source: "9Anime"}) {
		t.Error("Is9Anime should return true for Source=9Anime")
	}
	if !providers.IsGoyabu(&models.Anime{Source: "Goyabu"}) {
		t.Error("IsGoyabu should return true for Source=Goyabu")
	}

	if providers.IsFlixHQ(&models.Anime{Source: "AllAnime"}) {
		t.Error("IsFlixHQ should return false for Source=AllAnime")
	}
	if providers.Is9Anime(&models.Anime{Source: "AllAnime"}) {
		t.Error("Is9Anime should return false for Source=AllAnime")
	}
}

func TestProviderInterfaceCompliance(t *testing.T) {
	providerNames := []string{"allanime", "animefire", "animedrive", "flixhq", "9anime", "goyabu"}
	for _, name := range providerNames {
		p := providers.ForSourceName(name)
		if p == nil {
			t.Fatalf("ForSourceName(%q) returned nil", name)
		}
		if p.Name() == "" {
			t.Errorf("Provider for %q has empty Name()", name)
		}
	}
}
