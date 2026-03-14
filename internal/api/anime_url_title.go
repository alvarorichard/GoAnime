package api

import (
	"net/url"
	"regexp"
	"strings"
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
