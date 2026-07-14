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

// Resolve determines the source for an anime by scanning the registered
// sources' Describe() data, ordered by Priority. It is called ONCE per anime.
//
// Precedence (first match wins):
//
//  1. Explicit anime.Source field (checked across ALL sources first)
//  2. anime.MediaType
//  3. Tags in anime.Name
//  4. URL pattern / short ID
//
// If nothing matches, returns (nil, Kind=Unknown) with a warning log — the
// caller decides whether to fall back (BestEffortKind + Registered).
func Resolve(anime *models.Anime) (Source, ResolvedSource) {
	if anime == nil {
		return nil, ResolvedSource{Kind: Unknown, Reason: "nil anime"}
	}

	srcs := registeredByPriority()

	// Priority 1: Explicit Source field (highest priority, check all sources first)
	if anime.Source != "" {
		for _, s := range srcs {
			d := s.Describe()
			for _, e := range d.Explicit {
				if anime.Source == e {
					return s, ResolvedSource{Kind: d.Kind, Reason: "explicit Source=" + e}
				}
			}
		}
	}

	// Priority 2+: MediaType, tags, URL, shortID (lowest Priority wins)
	for _, s := range srcs {
		d := s.Describe()
		if reason, ok := d.matchNonExplicit(anime); ok {
			return s, ResolvedSource{Kind: d.Kind, Reason: reason}
		}
	}

	// PT-BR tag without specific source → default AnimeFire
	if anime.Name != "" {
		lower := strings.ToLower(anime.Name)
		if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portugu") {
			if s, ok := Registered(AnimeFire); ok {
				return s, ResolvedSource{Kind: AnimeFire, Reason: "PT-BR language tag (default AnimeFire)"}
			}
		}
	}

	util.Warn("source resolution fell through to Unknown", "anime", anime.Name, "url", anime.URL)
	return nil, ResolvedSource{Kind: Unknown, Reason: "no match, best-effort AllAnime"}
}

// ResolveURL resolves a source from a raw URL string only, scanning the
// registered sources. Used when no models.Anime context is available
// (e.g. direct URL playback).
func ResolveURL(rawURL string) (Source, ResolvedSource) {
	if rawURL == "" {
		return nil, ResolvedSource{Kind: Unknown, Reason: "empty URL"}
	}

	for _, s := range registeredByPriority() {
		d := s.Describe()
		if reason, ok := d.matchURL(rawURL); ok {
			return s, ResolvedSource{Kind: d.Kind, Reason: reason}
		}
	}

	return nil, ResolvedSource{Kind: Unknown, Reason: "URL not matched"}
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
