package types

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

// Source represents an anime scraper source
type Source int

const (
	// SourceAnimeFire represents the AnimeFire source.
	//
	// NOTE: this enum predates the registry and still lists only one source.
	// Goyabu, SuperFlix and AniDB are reachable through the CLI and the
	// registry but were never added here; SourceAllAnime was removed when the
	// AllAnime source was deleted. Extending this enum is a separate, breaking
	// SDK change.
	SourceAnimeFire Source = iota
)

// String returns the string representation of the source
func (s Source) String() string {
	switch s {
	case SourceAnimeFire:
		// "Animefire.io" is the canonical spelling the registry stamps onto
		// models.Anime.Source. String() used to answer "AnimeFire", which meant
		// results from SearchAnime(&SourceAnimeFire) never matched
		// source.String() — a mismatch masked while AllAnime (whose label and
		// stamp agreed) was the example source.
		return "Animefire.io"
	default:
		return "Unknown"
	}
}

// ToScraperType converts the public Source type to internal ScraperType
func (s Source) ToScraperType() scraper.ScraperType {
	switch s {
	case SourceAnimeFire:
		return scraper.AnimefireType
	default:
		return scraper.AnimefireType
	}
}

// ParseSource parses a string into a Source type
func ParseSource(s string) (Source, error) {
	switch s {
	case "AnimeFire", "animefire", "fire", "Animefire.io", "animefire.io":
		return SourceAnimeFire, nil
	default:
		return SourceAnimeFire, fmt.Errorf("unknown source: %s", s)
	}
}
