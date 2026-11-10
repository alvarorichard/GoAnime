package types

import (
	"fmt"

	"github.com/alvarorichard/Goanime/v2/internal/scraper"
)

// Source represents an anime scraper source
type Source int

// The values are append-only: SourceAnimeFire stays 0 and a new source goes at
// the end, so a Source stored by a caller keeps its meaning across releases.
const (
	// SourceAnimeFire is AnimeFire (PT-BR anime).
	SourceAnimeFire Source = iota
	// SourceGoyabu is Goyabu (PT-BR anime). Off by default; enable it with
	// GOANIME_ENABLED_SOURCES.
	SourceGoyabu
	// SourceHiAnime is HiAnime (English, subbed and dubbed).
	SourceHiAnime
	// SourceStartFlix is StartFlix (PT-BR movies and series).
	SourceStartFlix
	// SourceTopCine is TopCine (PT-BR movies and series).
	SourceTopCine
	// SourceAnimesDigital is AnimesDigital (PT-BR anime).
	SourceAnimesDigital
	// SourceAnimeDrive is AnimeDrive (PT-BR anime).
	SourceAnimeDrive
	// SourcePobreflix is Pobreflix (PT-BR movies and series).
	SourcePobreflix
	// SourceAnroll is AnimesROLL (PT-BR anime).
	SourceAnroll
	// SourcePomfy is Pomfy (movies and series).
	SourcePomfy

	sourceCount // not a source: the number of them
)

// sourceInfo is what is known about one Source. name is the spelling the
// registry stamps onto Anime.Source, so String and ParseSource round-trip.
type sourceInfo struct {
	name    string
	scraper scraper.ScraperType
	aliases []string
	// episodes is whether the source's episodes and streams can be reached
	// without the CLI's season and audio pickers: true for anime sources.
	episodes bool
	// catalog is whether its episodes hang off seasons and audio tracks
	// (movies and series): the Season calls apply to it.
	catalog bool
}

var sourceTable = [sourceCount]sourceInfo{
	SourceAnimeFire:     {"Animefire.io", scraper.AnimefireType, []string{"AnimeFire", "animefire", "fire", "animefire.io"}, true, false},
	SourceGoyabu:        {"Goyabu", scraper.GoyabuType, []string{"goyabu"}, true, false},
	SourceHiAnime:       {"HiAnime", scraper.HiAnimeType, []string{"hianime"}, true, false},
	SourceStartFlix:     {"StartFlix", scraper.StartFlixType, []string{"startflix"}, true, true},
	SourceTopCine:       {"TopCine", scraper.TopCineType, []string{"topcine"}, false, false},
	SourceAnimesDigital: {"AnimesDigital", scraper.AnimesDigitalType, []string{"animesdigital"}, true, false},
	SourceAnimeDrive:    {"AnimeDrive", scraper.AnimeDriveType, []string{"animedrive"}, true, false},
	SourcePobreflix:     {"Pobreflix", scraper.PobreflixType, []string{"pobreflix"}, false, false},
	SourceAnroll:        {"AnimesROLL", scraper.AnrollType, []string{"Anroll", "anroll", "animesroll"}, true, false},
	SourcePomfy:         {"Pomfy", scraper.PomfyType, []string{"pomfy"}, false, false},
}

func (s Source) valid() bool { return s >= 0 && s < sourceCount }

// SupportsEpisodes reports whether the source's episode lists and stream URLs
// can be fetched through the client. TopCine, Pobreflix and Pomfy can only be
// searched for now.
func (s Source) SupportsEpisodes() bool { return s.valid() && sourceTable[s].episodes }

// IsCatalog reports whether the source is a movie and series catalogue, whose
// episodes the client lists through GetSeasons and GetSeasonEpisodes and whose
// streams take an audio (see StreamOptions.Audio).
func (s Source) IsCatalog() bool { return s.valid() && sourceTable[s].catalog }

// Sources returns every source this package knows, in declaration order: all
// of them can be searched.
func Sources() []Source {
	out := make([]Source, sourceCount)
	for i := range out {
		out[i] = Source(i)
	}
	return out
}

// String returns the string representation of the source, the spelling the
// registry stamps onto Anime.Source.
func (s Source) String() string {
	if !s.valid() {
		return "Unknown"
	}
	return sourceTable[s].name
}

// ToScraperType converts the public Source type to internal ScraperType. An
// unknown Source falls back to AnimeFire, the original default.
func (s Source) ToScraperType() scraper.ScraperType {
	if !s.valid() {
		return scraper.AnimefireType
	}
	return sourceTable[s].scraper
}

// ParseSource parses a string into a Source type. It accepts the String()
// spelling and the lowercase name of each source; AnimeFire also answers to
// "fire".
func ParseSource(s string) (Source, error) {
	for i, info := range sourceTable {
		if s == info.name {
			return Source(i), nil
		}
		for _, a := range info.aliases {
			if s == a {
				return Source(i), nil
			}
		}
	}
	return SourceAnimeFire, fmt.Errorf("unknown source: %s", s)
}
