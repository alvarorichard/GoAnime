package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSearchVariations_IncludesURLRomaji(t *testing.T) {
	t.Parallel()

	variations := generateSearchVariationsWithURL(
		"Os Sete Pecados Capitais",
		"https://goyabu.io/anime/nanatsu-no-taizai",
	)

	assert.Contains(t, variations, "Os Sete Pecados Capitais", "should include original title")
	assert.Contains(t, variations, "nanatsu no taizai", "should include romaji from URL")
}

func TestGenerateSearchVariationsWithURL_EmptyURL(t *testing.T) {
	t.Parallel()

	variations := generateSearchVariationsWithURL("Naruto", "")
	assert.Contains(t, variations, "Naruto")
}

func TestGenerateSearchVariationsWithURL_NoSlug(t *testing.T) {
	t.Parallel()

	variations := generateSearchVariationsWithURL("Naruto", "https://example.com/")
	assert.Contains(t, variations, "Naruto")
}

func TestExtractRomajiFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "goyabu simple slug",
			url:      "https://goyabu.io/anime/nanatsu-no-taizai",
			expected: "nanatsu no taizai",
		},
		{
			name:     "goyabu slug with dublado suffix",
			url:      "https://goyabu.io/anime/shingeki-no-kyojin-dublado",
			expected: "shingeki no kyojin",
		},
		{
			name:     "goyabu slug with online suffix",
			url:      "https://goyabu.io/anime/boku-no-hero-academia-online",
			expected: "boku no hero academia",
		},
		{
			name:     "goyabu slug with dublado-online suffix",
			url:      "https://goyabu.io/anime/one-piece-dublado-online",
			expected: "one piece",
		},
		{
			name:     "goyabu slug with legendado suffix",
			url:      "https://goyabu.io/anime/bleach-legendado",
			expected: "bleach",
		},
		{
			name:     "animefire URL with /animes/ prefix",
			url:      "https://animefire.plus/animes/shingeki-no-kyojin-todos-os-episodios",
			expected: "shingeki no kyojin",
		},
		{
			name:     "animefire URL with dublado",
			url:      "https://animefire.plus/animes/naruto-shippuuden-dublado-todos-os-episodios",
			expected: "naruto shippuuden",
		},
		{
			name:     "trailing slash",
			url:      "https://goyabu.io/anime/death-note/",
			expected: "death note",
		},
		{
			name:     "empty URL returns empty",
			url:      "",
			expected: "",
		},
		{
			name:     "URL with no anime path returns empty",
			url:      "https://example.com/",
			expected: "",
		},
		{
			name:     "slug with season number preserved",
			url:      "https://goyabu.io/anime/boku-no-hero-academia-2",
			expected: "boku no hero academia 2",
		},
		{
			name:     "slug with hd suffix removed",
			url:      "https://goyabu.io/anime/naruto-hd",
			expected: "naruto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractRomajiFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
