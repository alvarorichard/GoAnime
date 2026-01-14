package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCleanTitle_BrazilianSources tests that CleanTitle properly handles
// anime titles from Brazilian sources (AnimeFire.plus) with various suffixes
// This test was added to verify the fix for the AniList enrichment failure
// that occurred with anime titles containing Portuguese metadata.
func TestCleanTitle_BrazilianSources(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// Language tag removal
		{
			name:     "removes [Portuguese] prefix",
			input:    "[Portuguese] Black Clover",
			expected: "Black Clover",
		},
		{
			name:     "removes [Português] prefix",
			input:    "[Português] Naruto Shippuden",
			expected: "Naruto Shippuden",
		},
		{
			name:     "removes [English] prefix",
			input:    "[English] One Piece",
			expected: "One Piece",
		},

		// Brazilian dub/sub indicators
		{
			name:     "removes Dublado suffix",
			input:    "Black Clover Dublado",
			expected: "Black Clover",
		},
		{
			name:     "removes Legendado suffix",
			input:    "One Piece Legendado",
			expected: "One Piece",
		},
		{
			name:     "removes Dual Áudio suffix",
			input:    "Naruto Dual Áudio",
			expected: "Naruto",
		},
		{
			name:     "removes (Dublado) in parentheses",
			input:    "Attack on Titan (Dublado)",
			expected: "Attack on Titan",
		},

		// Episode/complete indicators
		{
			name:     "removes Todos os Episódios with em-dash",
			input:    "Black Clover – Todos os Episódios",
			expected: "Black Clover",
		},
		{
			name:     "removes Todos os Episodios (no accent)",
			input:    "Bleach - Todos os Episodios",
			expected: "Bleach",
		},
		{
			name:     "removes Completo suffix",
			input:    "Death Note Completo",
			expected: "Death Note",
		},

		// Season indicators in Portuguese
		{
			name:     "removes X Temporada",
			input:    "My Hero Academia 5 Temporada",
			expected: "My Hero Academia",
		},
		{
			name:     "removes Xª Temporada",
			input:    "Demon Slayer 2ª Temporada",
			expected: "Demon Slayer",
		},
		{
			name:     "removes Temporada X",
			input:    "Jujutsu Kaisen Temporada 2",
			expected: "Jujutsu Kaisen",
		},

		// Part indicators
		{
			name:     "removes Parte X",
			input:    "Attack on Titan Parte 2",
			expected: "Attack on Titan",
		},

		// Episode count
		{
			name:     "removes (170 episodes)",
			input:    "Black Clover (170 episodes)",
			expected: "Black Clover",
		},
		{
			name:     "removes (171 episódios) Portuguese",
			input:    "Bleach (171 episódios)",
			expected: "Bleach",
		},

		// Combined Brazilian patterns (real-world examples)
		{
			name:     "full Brazilian title with language tag and dub indicator",
			input:    "[Portuguese] Black Clover Dublado",
			expected: "Black Clover",
		},
		{
			name:     "full Brazilian title with all metadata",
			input:    "[Portuguese] Black Clover – Todos os Episódios Dublado",
			expected: "Black Clover",
		},
		{
			name:     "AnimeFire style title with season",
			input:    "[Portuguese] My Hero Academia 6 Temporada Legendado",
			expected: "My Hero Academia",
		},
		{
			name:     "complex title with part and dub",
			input:    "[Português] Attack on Titan: The Final Season Parte 3 Dublado",
			expected: "Attack on Titan: The Final Season",
		},

		// Source tags
		{
			name:     "removes 🔥[AnimeFire] emoji tag",
			input:    "🔥[AnimeFire] One Piece",
			expected: "One Piece",
		},
		{
			name:     "removes [AnimeFire] tag",
			input:    "[AnimeFire] Naruto",
			expected: "Naruto",
		},
		{
			name:     "removes 🌐[AllAnime] emoji tag",
			input:    "🌐[AllAnime] Bleach",
			expected: "Bleach",
		},

		// URL-style names (hyphenated)
		{
			name:     "converts hyphenated URL names",
			input:    "black-clover",
			expected: "black clover",
		},

		// Edge cases
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "preserves clean title",
			input:    "Black Clover",
			expected: "Black Clover",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CleanTitle(tc.input)
			assert.Equal(t, tc.expected, result, "CleanTitle(%q) should equal %q", tc.input, tc.expected)
		})
	}
}

// TestGenerateSearchVariations tests that the search variation generator
// produces appropriate fallback search terms for AniList queries.
// This is critical for Brazilian sources where titles may not match exactly.
func TestGenerateSearchVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		input             string
		expectedVariation string // A variation that must be present
		minVariations     int    // Minimum number of variations expected
	}{
		{
			name:              "basic title generates at least original",
			input:             "Black Clover",
			expectedVariation: "Black Clover",
			minVariations:     1,
		},
		{
			name:              "lowercase title gets title case variation",
			input:             "black clover",
			expectedVariation: "Black Clover",
			minVariations:     2,
		},
		{
			name:              "title with colon generates base title",
			input:             "Attack on Titan: The Final Season",
			expectedVariation: "Attack on Titan",
			minVariations:     2,
		},
		{
			name:              "title with Roman numeral gets base variation",
			input:             "Code Geass II",
			expectedVariation: "Code Geass",
			minVariations:     2,
		},
		{
			name:              "title with trailing number gets base variation",
			input:             "Jujutsu Kaisen 2",
			expectedVariation: "Jujutsu Kaisen",
			minVariations:     2,
		},
		{
			name:              "title with 'no' particle gets variation without it",
			input:             "Shingeki no Kyojin",
			expectedVariation: "Shingeki Kyojin",
			minVariations:     2,
		},
		{
			name:              "title starting with 'The' gets variation without it",
			input:             "The Seven Deadly Sins",
			expectedVariation: "Seven Deadly Sins",
			minVariations:     2,
		},
		{
			name:              "long title gets shortened variations",
			input:             "That Time I Got Reincarnated as a Slime",
			expectedVariation: "That Time I Got",
			minVariations:     3, // original + first 3 words + first 4 words
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			variations := generateSearchVariations(tc.input)

			assert.GreaterOrEqual(t, len(variations), tc.minVariations,
				"generateSearchVariations(%q) should generate at least %d variations, got %d",
				tc.input, tc.minVariations, len(variations))

			// Check that expected variation is present
			found := false
			for _, v := range variations {
				if v == tc.expectedVariation {
					found = true
					break
				}
			}
			assert.True(t, found,
				"generateSearchVariations(%q) should contain %q, got %v",
				tc.input, tc.expectedVariation, variations)

			// First variation should always be the original
			assert.Equal(t, tc.input, variations[0],
				"First variation should be the original input")
		})
	}
}

// TestCleanTitle_RealAnimefireExamples tests with real-world examples
// that were causing the AniList 404 Not Found error.
func TestCleanTitle_RealAnimefireExamples(t *testing.T) {
	t.Parallel()

	// These are actual title formats seen from AnimeFire.plus
	// that were causing the AniList enrichment to fail with 404
	realWorldCases := []struct {
		name           string
		animefireTitle string
		expectedClean  string
		description    string
	}{
		{
			name:           "Black Clover dubbed series (170 eps)",
			animefireTitle: "[Portuguese] Black Clover Dublado",
			expectedClean:  "Black Clover",
			description:    "The main test case from the reported issue",
		},
		{
			name:           "Black Clover with episode count",
			animefireTitle: "[Portuguese] Black Clover (170 episodes) Dublado",
			expectedClean:  "Black Clover",
			description:    "Title with episode count metadata",
		},
		{
			name:           "Naruto Shippuden subtitled",
			animefireTitle: "[Português] Naruto Shippuden Legendado",
			expectedClean:  "Naruto Shippuden",
			description:    "Long-running series with Portuguese accent in tag",
		},
		{
			name:           "My Hero Academia season 6",
			animefireTitle: "[Portuguese] Boku no Hero Academia 6ª Temporada Dublado",
			expectedClean:  "Boku no Hero Academia",
			description:    "Season-specific title with ordinal",
		},
		{
			name:           "Demon Slayer dubbed",
			animefireTitle: "[Portuguese] Kimetsu no Yaiba Dublado",
			expectedClean:  "Kimetsu no Yaiba",
			description:    "Japanese title with dub indicator",
		},
		{
			name:           "One Piece with todos os episodios",
			animefireTitle: "[Portuguese] One Piece – Todos os Episódios",
			expectedClean:  "One Piece",
			description:    "Title with em-dash episode indicator",
		},
		{
			name:           "Dragon Ball Super complete",
			animefireTitle: "[Portuguese] Dragon Ball Super Completo Dublado",
			expectedClean:  "Dragon Ball Super",
			description:    "Title with Completo suffix",
		},
		{
			name:           "Attack on Titan Final Season Part",
			animefireTitle: "[Portuguese] Shingeki no Kyojin: The Final Season Parte 2 Dublado",
			expectedClean:  "Shingeki no Kyojin: The Final Season",
			description:    "Complex title with part indicator",
		},
	}

	for _, tc := range realWorldCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := CleanTitle(tc.animefireTitle)
			assert.Equal(t, tc.expectedClean, result,
				"CleanTitle failed for %s: %s\nInput: %q\nExpected: %q\nGot: %q",
				tc.name, tc.description, tc.animefireTitle, tc.expectedClean, result)
		})
	}
}

// TestCleanTitle_And_GenerateVariations_Integration tests the full flow
// of cleaning a title and generating search variations, simulating
// the actual AniList search process.
func TestCleanTitle_And_GenerateVariations_Integration(t *testing.T) {
	t.Parallel()

	// This simulates the exact flow that was failing:
	// 1. User selects anime from AnimeFire
	// 2. Title is cleaned with CleanTitle
	// 3. Search variations are generated for AniList query

	testCases := []struct {
		name              string
		rawTitle          string
		mustContainSearch string // The final search variations must contain this
	}{
		{
			name:              "Brazilian dubbed anime",
			rawTitle:          "[Portuguese] Black Clover Dublado",
			mustContainSearch: "Black Clover",
		},
		{
			name:              "Brazilian subtitled anime",
			rawTitle:          "[Português] Naruto Shippuden Legendado",
			mustContainSearch: "Naruto Shippuden",
		},
		{
			name:              "Anime with season in Portuguese",
			rawTitle:          "[Portuguese] Boku no Hero Academia 6 Temporada",
			mustContainSearch: "Boku no Hero Academia",
		},
		{
			name:              "AllAnime English source",
			rawTitle:          "[English] One Piece",
			mustContainSearch: "One Piece",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Clean the title (as done in FetchAnimeFromAniList)
			cleanedTitle := CleanTitle(tc.rawTitle)

			// Step 2: Generate search variations
			variations := generateSearchVariations(cleanedTitle)

			// Step 3: Verify the expected search term is in the variations
			found := false
			for _, v := range variations {
				if v == tc.mustContainSearch {
					found = true
					break
				}
			}

			assert.True(t, found,
				"Integration test failed for %q:\nCleaned title: %q\nVariations: %v\nMust contain: %q",
				tc.rawTitle, cleanedTitle, variations, tc.mustContainSearch)
		})
	}
}

// TestCleanTitle_PreservesValidTitles ensures CleanTitle doesn't break
// titles that are already clean and valid for AniList search.
func TestCleanTitle_PreservesValidTitles(t *testing.T) {
	t.Parallel()

	validTitles := []string{
		"Black Clover",
		"One Piece",
		"Naruto",
		"Naruto Shippuden",
		"Attack on Titan",
		"Demon Slayer: Kimetsu no Yaiba",
		"My Hero Academia",
		"Jujutsu Kaisen",
		"Chainsaw Man",
		"Spy x Family",
		"Bocchi the Rock!",
		"Mob Psycho 100",
		"Dr. Stone",
		"Re:Zero",
		"Steins;Gate",
	}

	for _, title := range validTitles {
		title := title
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			result := CleanTitle(title)
			assert.Equal(t, title, result,
				"CleanTitle should preserve valid title %q, but got %q", title, result)
		})
	}
}
