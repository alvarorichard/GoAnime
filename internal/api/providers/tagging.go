package providers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
)

// Per-source result tagging (ported from the ScraperManager, re-keyed by
// SourceKind). A source knows its own language, so tagging its own results is
// the Model B home for this logic — each provider applies it in Search.

var (
	ptbrSpaceRe      = regexp.MustCompile(`\s+`)
	ptbrAgeRatingRe  = regexp.MustCompile(`\bA\d{2}\b`)
	ptbrNumRatingRe  = regexp.MustCompile(`\b\d+[.,]\d+\b|\bN/A\b`)
	ptbrTypeSuffixRe = regexp.MustCompile(`(?i)\s*\((TV\s*Short|TV|Movie|OVA|ONA|Special|Filme|Especial|Longa-?Metragem)\)`)
	ptbrDubLegRe     = regexp.MustCompile(`(?i)\s*[\(\[]?(dublado|legendado)[\)\]]?`)
)

// sourceDisplayName is the canonical Source string stamped onto results. Note
// AnimeFire uses "Animefire.io" (the scraper's canonical spelling) so downstream
// source matching stays consistent.
func sourceDisplayName(kind source.SourceKind) string {
	switch kind {
	case source.AllAnime:
		return "AllAnime"
	case source.AnimeFire:
		return "Animefire.io"
	case source.Goyabu:
		return "Goyabu"
	case source.SuperFlix:
		return "SuperFlix"
	case source.AniDB:
		return "AniDB"
	default:
		return string(kind)
	}
}

func languageTag(kind source.SourceKind) string {
	if kind == source.AllAnime || kind == source.AniDB {
		return "[English]"
	}
	return "[PT-BR]"
}

// cleanPTBRTitle strips dub/leg labels, age/numeric ratings, and media-type
// suffixes from a PT-BR title so tagResults can re-add a consistent tag.
func cleanPTBRTitle(title string) string {
	title = ptbrDubLegRe.ReplaceAllString(title, "")
	title = ptbrSpaceRe.ReplaceAllString(strings.TrimSpace(title), " ")
	title = ptbrAgeRatingRe.ReplaceAllString(title, "")
	title = ptbrNumRatingRe.ReplaceAllString(title, "")
	title = ptbrTypeSuffixRe.ReplaceAllString(title, "")
	return strings.TrimSpace(ptbrSpaceRe.ReplaceAllString(title, " "))
}

// tagResults language-tags a source's search results in place and stamps the
// canonical Source field, matching the legacy ScraperManager.tagResults exactly.
func tagResults(results []*models.Anime, kind source.SourceKind) {
	name := sourceDisplayName(kind)
	isPTBR := kind == source.AnimeFire || kind == source.Goyabu

	for _, anime := range results {
		if isPTBR {
			anime.Name = cleanPTBRTitle(anime.Name)
		}

		hasLanguageTag := strings.Contains(anime.Name, "[English]") ||
			strings.Contains(anime.Name, "[PT-BR]") ||
			strings.Contains(anime.Name, "[Portuguese]") ||
			strings.Contains(anime.Name, "[Português]") ||
			strings.Contains(anime.Name, "[Multilanguage]") ||
			strings.Contains(anime.Name, "[Movie]") ||
			strings.Contains(anime.Name, "[TV]")

		if !hasLanguageTag {
			if kind == source.SuperFlix {
				switch anime.MediaType {
				case models.MediaTypeMovie:
					anime.Name = fmt.Sprintf("[Movie] [PT-BR] %s", anime.Name)
				case models.MediaTypeTV:
					anime.Name = fmt.Sprintf("[TV] [PT-BR] %s", anime.Name)
				default:
					anime.Name = fmt.Sprintf("[PT-BR] %s", anime.Name)
				}
			} else {
				anime.Name = fmt.Sprintf("%s %s", languageTag(kind), anime.Name)
			}
		}

		if isPTBR {
			lowerURL := strings.ToLower(anime.URL)
			lowerName := strings.ToLower(anime.Name)
			switch {
			case strings.Contains(lowerName, "dublado") || strings.Contains(lowerURL, "dublado"):
				if !strings.Contains(anime.Name, "(Dublado)") {
					anime.Name += " (Dublado)"
				}
			case strings.Contains(lowerName, "legendado") || strings.Contains(lowerURL, "legendado"):
				if !strings.Contains(anime.Name, "(Legendado)") {
					anime.Name += " (Legendado)"
				}
			}
		}

		anime.Source = name
	}
}
