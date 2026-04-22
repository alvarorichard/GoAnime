// Package source provides canonical source resolution for all media providers.
// It is the single source of truth for determining which scraper handles a given anime/media.
package source

import "github.com/alvarorichard/Goanime/internal/scraper"

// SourceKind is the canonical type-safe identifier for a media source.
// Unlike scraper.ScraperType (iota int), SourceKind is human-readable and safe for logging.
//
//revive:disable-next-line:exported
type SourceKind string

// Media source kind constants used throughout the application for source identification.
const (
	AllAnime   SourceKind = "AllAnime"
	AnimeFire  SourceKind = "Animefire.io"
	AnimeDrive SourceKind = "AnimeDrive"
	FlixHQ     SourceKind = "FlixHQ"
	SFlix      SourceKind = "SFlix"
	NineAnime  SourceKind = "9Anime"
	Goyabu     SourceKind = "Goyabu"
	SuperFlix  SourceKind = "SuperFlix"
	Unknown    SourceKind = "Unknown"
)

// ScraperTypeMap maps a SourceKind to the corresponding scraper.ScraperType.
var ScraperTypeMap = map[SourceKind]scraper.ScraperType{
	AllAnime:   scraper.AllAnimeType,
	AnimeFire:  scraper.AnimefireType,
	AnimeDrive: scraper.AnimeDriveType,
	FlixHQ:     scraper.FlixHQType,
	SFlix:      scraper.SFlixType,
	NineAnime:  scraper.NineAnimeType,
	Goyabu:     scraper.GoyabuType,
	SuperFlix:  scraper.SuperFlixType,
}

// ScraperType returns the matching scraper type when one exists.
func (k SourceKind) ScraperType() (scraper.ScraperType, bool) {
	st, ok := ScraperTypeMap[k]
	return st, ok
}

// ScraperTypeFor maps a SourceKind to the corresponding scraper.ScraperType.
func ScraperTypeFor(kind SourceKind) (scraper.ScraperType, bool) {
	return kind.ScraperType()
}
