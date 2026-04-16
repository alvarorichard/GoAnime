// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// SourceKind represents the canonical scraper/media source for an anime entry.
type SourceKind string

const (
	// SourceUnknown marks entries whose source could not be resolved safely.
	SourceUnknown    SourceKind = "Unknown"
	SourceAllAnime   SourceKind = "AllAnime"
	SourceAnimefire  SourceKind = "Animefire.io"
	SourceAnimeDrive SourceKind = "AnimeDrive"
	SourceGoyabu     SourceKind = "Goyabu"
	SourceNineAnime  SourceKind = "9Anime"
	SourceFlixHQ     SourceKind = "FlixHQ"
	SourceSuperFlix  SourceKind = "SuperFlix"
)

// ResolvedSource is the normalized source resolution result for an anime entry.
type ResolvedSource struct {
	Kind   SourceKind
	Name   string
	Reason string
}

// Apply normalizes the source field on the provided anime when a source was resolved.
func (r ResolvedSource) Apply(anime *models.Anime) {
	if anime == nil || r.Name == "" {
		return
	}
	anime.Source = r.Name
}

// IsProviderBacked returns true when this source should dispatch through the source-provider registry.
func (r ResolvedSource) IsProviderBacked() bool {
	return r.Kind.IsProviderBacked()
}

// IsProviderBacked returns true when this source should dispatch through the source-provider registry.
func (k SourceKind) IsProviderBacked() bool {
	switch k {
	case SourceAllAnime, SourceAnimefire, SourceAnimeDrive, SourceGoyabu:
		return true
	default:
		return false
	}
}

// ScraperType returns the unified scraper type when the source maps directly to one.
func (k SourceKind) ScraperType() (scraper.ScraperType, bool) {
	switch k {
	case SourceAllAnime:
		return scraper.AllAnimeType, true
	case SourceAnimefire:
		return scraper.AnimefireType, true
	case SourceAnimeDrive:
		return scraper.AnimeDriveType, true
	case SourceGoyabu:
		return scraper.GoyabuType, true
	case SourceNineAnime:
		return scraper.NineAnimeType, true
	case SourceFlixHQ:
		return scraper.FlixHQType, true
	case SourceSuperFlix:
		return scraper.SuperFlixType, true
	default:
		return 0, false
	}
}

// ResolveSource infers the canonical source once so downstream code can share the same rules.
func ResolveSource(anime *models.Anime) (ResolvedSource, error) {
	if anime == nil {
		return ResolvedSource{}, fmt.Errorf("cannot resolve source for nil anime")
	}

	if resolved, ok := resolveSourceFromExplicit(anime.Source); ok {
		return resolved, nil
	}

	if anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		return newResolvedSource(SourceFlixHQ, "media type"), nil
	}

	if IsAllAnimeShortID(anime.URL) && !hasPTBRTag(anime.Name) {
		return newResolvedSource(SourceAllAnime, "AllAnime short ID"), nil
	}

	if resolved, handled, err := resolveSourceFromTags(anime.Name, anime.URL); handled {
		if err != nil {
			return ResolvedSource{}, err
		}
		return resolved, nil
	}

	if resolved, ok := resolveSourceFromURL(anime.URL); ok {
		return resolved, nil
	}

	return ResolvedSource{}, fmt.Errorf(
		"could not resolve source for %q (source=%q mediaType=%q url=%q)",
		anime.Name,
		anime.Source,
		anime.MediaType,
		anime.URL,
	)
}

// IsAllAnimeSource reports whether the given anime resolves to the AllAnime source.
func IsAllAnimeSource(anime *models.Anime) bool {
	resolved, err := ResolveSource(anime)
	return err == nil && resolved.Kind == SourceAllAnime
}

// IsAllAnimeShortID reports whether a string matches the short identifier style used by AllAnime.
func IsAllAnimeShortID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(strings.ToLower(value), "http") {
		return false
	}
	if isNumericLike(value) {
		return false
	}
	return len(value) >= 6 && len(value) < 30 && containsASCIILetter(value)
}

// ExtractAllAnimeID extracts the AllAnime ID from either a short ID or a full URL.
func ExtractAllAnimeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if !strings.Contains(strings.ToLower(value), "http") && len(value) < 30 {
		return value
	}

	if strings.Contains(strings.ToLower(value), "allanime") {
		parts := strings.SplitSeq(value, "/")
		for part := range parts {
			if IsAllAnimeShortID(part) {
				return part
			}
		}
	}

	return value
}

func newResolvedSource(kind SourceKind, reason string) ResolvedSource {
	return ResolvedSource{
		Kind:   kind,
		Name:   string(kind),
		Reason: reason,
	}
}

func resolveSourceFromExplicit(source string) (ResolvedSource, bool) {
	source = strings.TrimSpace(source)
	if source == "" {
		return ResolvedSource{}, false
	}

	lower := strings.ToLower(source)
	switch {
	case lower == strings.ToLower(string(SourceAllAnime)) || strings.Contains(lower, "allanime"):
		return newResolvedSource(SourceAllAnime, "explicit source"), true
	case lower == strings.ToLower(string(SourceAnimefire)) || strings.Contains(lower, "animefire"):
		return newResolvedSource(SourceAnimefire, "explicit source"), true
	case lower == strings.ToLower(string(SourceAnimeDrive)) || strings.Contains(lower, "animedrive"):
		return newResolvedSource(SourceAnimeDrive, "explicit source"), true
	case lower == strings.ToLower(string(SourceGoyabu)) || strings.Contains(lower, "goyabu"):
		return newResolvedSource(SourceGoyabu, "explicit source"), true
	case lower == strings.ToLower(string(SourceNineAnime)) || strings.Contains(lower, "9anime") || strings.Contains(lower, "nineanime"):
		return newResolvedSource(SourceNineAnime, "explicit source"), true
	case lower == strings.ToLower(string(SourceFlixHQ)) || lower == "movie" || lower == "tv" || strings.Contains(lower, "flixhq"):
		return newResolvedSource(SourceFlixHQ, "explicit source"), true
	case lower == strings.ToLower(string(SourceSuperFlix)) || strings.Contains(lower, "superflix"):
		return newResolvedSource(SourceSuperFlix, "explicit source"), true
	default:
		return ResolvedSource{}, false
	}
}

func resolveSourceFromTags(name, url string) (ResolvedSource, bool, error) {
	lowerName := strings.ToLower(name)
	switch {
	case strings.Contains(lowerName, "[allanime]") || strings.Contains(lowerName, "[english]"):
		return newResolvedSource(SourceAllAnime, "language tag"), true, nil
	case strings.Contains(lowerName, "[animefire]"):
		return newResolvedSource(SourceAnimefire, "legacy AnimeFire tag"), true, nil
	case strings.Contains(lowerName, "[multilanguage]"):
		return newResolvedSource(SourceNineAnime, "language tag"), true, nil
	case strings.Contains(lowerName, "[pt-br]") ||
		strings.Contains(lowerName, "[portuguese]") ||
		strings.Contains(lowerName, "[português]"):
		if resolved, ok := resolvePTBRSource(url); ok {
			resolved.Reason = "PT-BR tag + URL"
			return resolved, true, nil
		}
		return ResolvedSource{}, true, fmt.Errorf(
			"could not resolve PT-BR source for %q without an explicit source or recognizable URL",
			name,
		)
	default:
		return ResolvedSource{}, false, nil
	}
}

func resolveSourceFromURL(url string) (ResolvedSource, bool) {
	lowerURL := strings.ToLower(strings.TrimSpace(url))
	switch {
	case strings.Contains(lowerURL, "animefire"):
		return newResolvedSource(SourceAnimefire, "URL"), true
	case strings.Contains(lowerURL, "animesdrive"):
		return newResolvedSource(SourceAnimeDrive, "URL"), true
	case strings.Contains(lowerURL, "goyabu"):
		return newResolvedSource(SourceGoyabu, "URL"), true
	case strings.Contains(lowerURL, "allanime"):
		return newResolvedSource(SourceAllAnime, "URL"), true
	case strings.Contains(lowerURL, "flixhq") || strings.Contains(lowerURL, "sflix"):
		return newResolvedSource(SourceFlixHQ, "URL"), true
	case strings.Contains(lowerURL, "superflix"):
		return newResolvedSource(SourceSuperFlix, "URL"), true
	case strings.Contains(lowerURL, "9anime"):
		return newResolvedSource(SourceNineAnime, "URL"), true
	default:
		return ResolvedSource{}, false
	}
}

func resolvePTBRSource(url string) (ResolvedSource, bool) {
	resolved, ok := resolveSourceFromURL(url)
	if !ok {
		return ResolvedSource{}, false
	}

	switch resolved.Kind {
	case SourceAnimefire, SourceAnimeDrive, SourceGoyabu, SourceSuperFlix:
		return resolved, true
	default:
		return ResolvedSource{}, false
	}
}

func isNumericLike(value string) bool {
	if value == "" {
		return false
	}

	dotCount := 0
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r == '.':
			dotCount++
			if dotCount > 1 {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func containsASCIILetter(value string) bool {
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func hasPTBRTag(name string) bool {
	lowerName := strings.ToLower(name)
	return strings.Contains(lowerName, "[pt-br]") ||
		strings.Contains(lowerName, "[portuguese]") ||
		strings.Contains(lowerName, "[portugu\u00eas]") ||
		strings.Contains(lowerName, "[animefire]")
}
