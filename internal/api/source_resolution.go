// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	sourcepkg "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceKind is an alias for source.SourceKind, the canonical media source identifier.
type SourceKind = sourcepkg.SourceKind

// ResolvedSource is an alias for source.ResolvedSource, the immutable result of source resolution.
type ResolvedSource = sourcepkg.ResolvedSource

// Source kind constants re-exported for callers that import the api package directly.
const (
	SourceUnknown    SourceKind = sourcepkg.Unknown
	SourceAllAnime   SourceKind = sourcepkg.AllAnime
	SourceAnimefire  SourceKind = sourcepkg.AnimeFire
	SourceAnimeDrive SourceKind = sourcepkg.AnimeDrive
	SourceFlixHQ     SourceKind = sourcepkg.FlixHQ
	SourceSFlix      SourceKind = sourcepkg.SFlix
	SourceNineAnime  SourceKind = sourcepkg.NineAnime
	SourceGoyabu     SourceKind = sourcepkg.Goyabu
	SourceSuperFlix  SourceKind = sourcepkg.SuperFlix
)

// ResolveSource determines the canonical source for an anime.
func ResolveSource(anime *models.Anime) (ResolvedSource, error) {
	return sourcepkg.Resolve(anime)
}

// ResolveURL resolves a source from a raw URL string only.
func ResolveURL(url string) ResolvedSource {
	return sourcepkg.ResolveURL(url)
}

// IsAllAnimeSource reports whether the anime's resolved source is AllAnime.
func IsAllAnimeSource(anime *models.Anime) bool {
	resolved, err := ResolveSource(anime)
	return err == nil && resolved.Kind == SourceAllAnime
}

// IsAllAnimeShortID returns true if value looks like an AllAnime short ID.
func IsAllAnimeShortID(value string) bool {
	return sourcepkg.IsAllAnimeShortID(value)
}

// ExtractAllAnimeID extracts the AllAnime ID from a URL or bare string.
func ExtractAllAnimeID(value string) string {
	return sourcepkg.ExtractAllAnimeID(value)
}
