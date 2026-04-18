package source

import (
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ResolvedSource is the immutable result of source resolution.
type ResolvedSource struct {
	Kind   SourceKind // The resolved source type.
	Reason string     // Human-readable explanation for debugging.
}

// Resolve determines the canonical source for an anime by iterating sourceDefs.
// It is called ONCE per anime; the result is passed to downstream layers.
//
// Precedence per definition (first match wins):
//  1. Explicit anime.Source field
//  2. anime.MediaType
//  3. Tags in anime.Name
//  4. URL pattern / short ID
//
// If nothing matches, returns Kind=Unknown with a warning log.
func Resolve(anime *models.Anime) ResolvedSource {
	if anime == nil {
		return ResolvedSource{Kind: Unknown, Reason: "nil anime"}
	}

	// Priority 1: Explicit Source field (highest priority, check all defs first)
	if anime.Source != "" {
		for i := range sourceDefs {
			for _, s := range sourceDefs[i].Explicit {
				if anime.Source == s {
					return ResolvedSource{Kind: sourceDefs[i].Kind, Reason: "explicit Source=" + s}
				}
			}
		}
	}

	// Priority 2+: MediaType, tags, URL, shortID (first def match wins)
	for i := range sourceDefs {
		if reason, ok := sourceDefs[i].matchNonExplicit(anime); ok {
			return ResolvedSource{Kind: sourceDefs[i].Kind, Reason: reason}
		}
	}

	// PT-BR tag without specific source → default AnimeFire
	if anime.Name != "" {
		lower := strings.ToLower(anime.Name)
		if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portugu") {
			return ResolvedSource{Kind: AnimeFire, Reason: "PT-BR language tag (default AnimeFire)"}
		}
	}

	util.Warn("source resolution fell through to Unknown", "anime", anime.Name, "url", anime.URL)
	return ResolvedSource{Kind: Unknown, Reason: "no match, best-effort AllAnime"}
}

// ResolveURL resolves a source from a raw URL string only.
// Used when no models.Anime context is available (e.g. direct URL playback).
func ResolveURL(rawURL string) ResolvedSource {
	if rawURL == "" {
		return ResolvedSource{Kind: Unknown, Reason: "empty URL"}
	}

	for i := range sourceDefs {
		if reason, ok := sourceDefs[i].matchURL(rawURL); ok {
			return ResolvedSource{Kind: sourceDefs[i].Kind, Reason: reason}
		}
	}

	return ResolvedSource{Kind: Unknown, Reason: "URL not matched"}
}

// BestEffortKind returns the effective SourceKind for dispatch.
// Unknown is treated as AllAnime for backward compatibility.
func (r ResolvedSource) BestEffortKind() SourceKind {
	if r.Kind == Unknown {
		return AllAnime
	}
	return r.Kind
}

// IsAllAnimeShortID returns true if s looks like an AllAnime short ID:
// alphanumeric, <30 chars, not a URL, not purely numeric.
func IsAllAnimeShortID(s string) bool {
	if s == "" || len(s) > 30 || strings.Contains(s, "http") || strings.Contains(s, "/") {
		return false
	}
	hasLetter := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			// ok
		default:
			return false // non-alphanumeric
		}
	}
	return hasLetter
}

// ExtractAllAnimeID extracts the AllAnime ID from a URL or bare string.
func ExtractAllAnimeID(s string) string {
	if IsAllAnimeShortID(s) {
		return s
	}
	if idx := strings.LastIndex(s, "/"); idx >= 0 && idx < len(s)-1 {
		candidate := s[idx+1:]
		if IsAllAnimeShortID(candidate) {
			return candidate
		}
	}
	return s
}
