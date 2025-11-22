package types

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/scraper"
)

// Source represents an anime scraper source
type Source int

const (
	// SourceAllAnime represents the AllAnime source
	SourceAllAnime Source = iota
	// SourceAnimeFire represents the AnimeFire source
	SourceAnimeFire
)

// String returns the string representation of the source
func (s Source) String() string {
	switch s {
	case SourceAllAnime:
		return "AllAnime"
	case SourceAnimeFire:
		return "AnimeFire"
	default:
		return "Unknown"
	}
}

// ToScraperType converts the public Source type to internal ScraperType
func (s Source) ToScraperType() scraper.ScraperType {
	switch s {
	case SourceAllAnime:
		return scraper.AllAnimeType
	case SourceAnimeFire:
		return scraper.AnimefireType
	default:
		return scraper.AllAnimeType
	}
}

// ParseSource parses a string into a Source type
func ParseSource(s string) (Source, error) {
	switch s {
	case "AllAnime", "allanime", "all":
		return SourceAllAnime, nil
	case "AnimeFire", "animefire", "fire":
		return SourceAnimeFire, nil
	default:
		return SourceAllAnime, fmt.Errorf("unknown source: %s", s)
	}
}
