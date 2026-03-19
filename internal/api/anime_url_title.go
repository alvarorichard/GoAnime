// Package api provides enhanced anime search and streaming capabilities.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/pkg/errors"
)

// ptbrURLSuffixes are noise suffixes commonly appended to anime URL slugs
// on Brazilian streaming sites like Goyabu and AnimeFire.
var ptbrURLSuffixes = regexp.MustCompile(
	`(?i)-(?:dublado|legendado|online|hd|completo|todos-os-episodios)(?:-(?:dublado|legendado|online|hd|completo|todos-os-episodios))*$`,
)

// extractRomajiFromURL extracts a romaji anime title from a Goyabu or AnimeFire URL slug.
//
// Brazilian anime sites use romaji-based URL slugs:
//
//	https://goyabu.io/anime/nanatsu-no-taizai          → "nanatsu no taizai"
//	https://goyabu.io/anime/shingeki-no-kyojin-dublado  → "shingeki no kyojin"
//	https://animefire.plus/animes/naruto-shippuuden-dublado-todos-os-episodios → "naruto shippuuden"
//
// This extracted romaji title can then be used to search AniList when
// the original PT-BR display title fails to match.
func extractRomajiFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path == "" {
		return ""
	}

	// Get the last path segment: /anime/nanatsu-no-taizai → nanatsu-no-taizai
	path := strings.TrimRight(parsed.Path, "/")
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return ""
	}
	slug := segments[len(segments)-1]

	// Ignore non-slug segments (e.g. bare domain, "anime", "animes")
	if slug == "" || slug == "anime" || slug == "animes" {
		return ""
	}

	// Remove known PT-BR noise suffixes
	slug = ptbrURLSuffixes.ReplaceAllString(slug, "")

	// Replace hyphens with spaces
	title := strings.ReplaceAll(slug, "-", " ")

	return strings.TrimSpace(title)
}

// generateSearchVariationsWithURL extends generateSearchVariations by adding
// a romaji title extracted from the anime's URL slug.
func generateSearchVariationsWithURL(cleanedName, animeURL string) []string {
	variations := generateSearchVariations(cleanedName)

	romaji := extractRomajiFromURL(animeURL)
	if romaji == "" || romaji == strings.ToLower(cleanedName) {
		return variations
	}

	// Check if romaji is already in variations (case-insensitive)
	for _, v := range variations {
		if strings.EqualFold(v, romaji) {
			return variations
		}
	}

	// Prepend romaji right after the original name
	result := make([]string, 0, len(variations)+2)
	result = append(result, variations[0]) // original cleaned name first
	result = append(result, romaji)        // romaji from URL second

	// Also add title-cased version
	titleCased := toTitleCase(romaji)
	if titleCased != romaji {
		result = append(result, titleCased)
	}

	result = append(result, variations[1:]...) // remaining variations
	return result
}

// toTitleCase capitalizes the first letter of each word.
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(string(w[0])) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// FetchAnimeFromAniListWithURL is like FetchAnimeFromAniList but also tries
// romaji title extracted from the anime URL when the display name fails.
func FetchAnimeFromAniListWithURL(animeName, animeURL string) (*models.AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	util.Debugf("Querying AniList for: '%s' (original: '%s', url: '%s')", cleanedName, animeName, animeURL)

	// Check cache first
	cache := util.GetAniListCache()
	cacheKey := "anilist:" + strings.ToLower(cleanedName)
	if cached, found := cache.Get(cacheKey); found {
		var result models.AniListResponse
		if err := json.Unmarshal(cached, &result); err == nil && result.Data.Media.ID != 0 {
			util.Debugf("AniList cache hit for: '%s'", cleanedName)
			return &result, nil
		}
	}

	// Generate search variations including romaji from URL
	searchVariations := generateSearchVariationsWithURL(cleanedName, animeURL)

	query := `query ($search: String) {
        Media(search: $search, type: ANIME) {
            id
            title { romaji english native }
            idMal
            coverImage { large }
            synonyms
        }
    }`

	var lastErr error
	for _, searchTerm := range searchVariations {
		util.Debugf("Trying AniList search with: '%s'", searchTerm)

		jsonData, err := json.Marshal(map[string]any{
			"query": query,
			"variables": map[string]any{
				"search": searchTerm,
			},
		})
		if err != nil {
			lastErr = fmt.Errorf("JSON marshal failed: %w", err)
			continue
		}

		resp, err := httpPostFast("https://graphql.anilist.co", jsonData)
		if err != nil {
			lastErr = fmt.Errorf("AniList request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		safeClose(resp.Body, "AniList response body")

		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			util.Debugf("AniList error response: %s", string(body))
			lastErr = fmt.Errorf("AniList returned: %s", resp.Status)
			continue
		}

		var result models.AniListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("JSON decode failed: %w", err)
			continue
		}

		if result.Data.Media.ID == 0 {
			lastErr = errors.New("no matching anime found on AniList")
			continue
		}

		cache.Set(cacheKey, body)

		util.Debugf("AniList found: ID=%d, MAL=%d, Title=%s (search term: '%s')",
			result.Data.Media.ID,
			result.Data.Media.IDMal,
			result.Data.Media.Title.Romaji,
			searchTerm)

		return &result, nil
	}

	return nil, lastErr
}
