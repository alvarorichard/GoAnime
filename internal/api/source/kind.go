// Package source provides canonical source resolution for all media providers.
// It is the single source of truth for determining which scraper handles a given anime/media.
package source

import "github.com/alvarorichard/Goanime/v2/internal/scraper"

// SourceKind is the canonical type-safe identifier for a media source.
// Unlike scraper.ScraperType (iota int), SourceKind is human-readable and safe for logging.
type SourceKind string

const (
	AnimeFire     SourceKind = "AnimeFire"
	Goyabu        SourceKind = "Goyabu"
	HiAnime       SourceKind = "HiAnime"
	StartFlix     SourceKind = "StartFlix"
	TopCine       SourceKind = "TopCine"
	AnimesDigital SourceKind = "AnimesDigital"
	AnimeDrive    SourceKind = "AnimeDrive"
	// Pobreflix is the catalogue whose /api search answers directly and
	// whose current pages embed public MP4 servers. Old-theme pages keep
	// the StartFlix panel fallback by TMDB id.
	Pobreflix SourceKind = "Pobreflix"
	Anroll    SourceKind = "AnimesROLL"
	// Pomfy is the Byse-CDN movie/series catalog keyed by TMDB id; its
	// playback runs the proof-of-work captcha and AES-256-GCM seal flow.
	Pomfy SourceKind = "Pomfy"

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

// KindFor is the inverse of ScraperTypeFor: the SourceKind of a ScraperType,
// or ("", false) when none has it.
func KindFor(st scraper.ScraperType) (SourceKind, bool) {
	for kind, t := range scraperTypeMap {
		if t == st {
			return kind, true
		}
	}
	return "", false
}

var scraperTypeMap = map[SourceKind]scraper.ScraperType{
	AnimeFire:     scraper.AnimefireType,
	Goyabu:        scraper.GoyabuType,
	HiAnime:       scraper.HiAnimeType,
	StartFlix:     scraper.StartFlixType,
	TopCine:       scraper.TopCineType,
	AnimesDigital: scraper.AnimesDigitalType,
	AnimeDrive:    scraper.AnimeDriveType,
	Pobreflix:     scraper.PobreflixType,
	Anroll:        scraper.AnrollType,
	Pomfy:         scraper.PomfyType,
}
