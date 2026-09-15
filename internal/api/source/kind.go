// Package source provides canonical source resolution for all media providers.
// It is the single source of truth for determining which scraper handles a given anime/media.
package source

import "github.com/alvarorichard/Goanime/internal/scraper"

// SourceKind is the canonical type-safe identifier for a media source.
// Unlike scraper.ScraperType (iota int), SourceKind is human-readable and safe for logging.
type SourceKind string

const (
	AnimeFire SourceKind = "AnimeFire"
	Goyabu    SourceKind = "Goyabu"
	SuperFlix SourceKind = "SuperFlix"
	AniDB     SourceKind = "AniDB"

	// Unknown is returned when no definition matches. It is surfaced with a
	// warning rather than guessed at: the best-effort fallback used to be
	// AllAnime, and that source no longer exists.
	Unknown SourceKind = "Unknown"
)

// ScraperTypeFor maps a SourceKind to the corresponding scraper.ScraperType.
// Returns the ScraperType and true if found, or (0, false) for Unknown/unregistered kinds.
func ScraperTypeFor(kind SourceKind) (scraper.ScraperType, bool) {
	st, ok := scraperTypeMap[kind]
	return st, ok
}

var scraperTypeMap = map[SourceKind]scraper.ScraperType{
	AnimeFire: scraper.AnimefireType,
	Goyabu:    scraper.GoyabuType,
	SuperFlix: scraper.SuperFlixType,
	AniDB:     scraper.AniDBType,
}
