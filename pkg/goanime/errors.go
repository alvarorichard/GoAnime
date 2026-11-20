// Package goanime exposes public client APIs and shared sentinel errors.
package goanime

import (
	"errors"

	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// ErrSourceUnavailable indicates that the selected upstream source is
// temporarily unavailable or blocked.
var ErrSourceUnavailable = netx.ErrSourceUnavailable

// ErrEpisodesUnsupported means the source can be searched but its episodes
// and streams are not reachable through this package yet: TopCine, Pobreflix
// and Pomfy are movie and series catalogues whose episodes are chosen through
// season and audio pickers in the command-line tool.
var ErrEpisodesUnsupported = errors.New("source does not support episode lists or stream URLs in this package")

// Errors of the movie and series calls. They match with errors.Is.
var (
	// ErrSeasonRequired is a series asked for without a season: list them with
	// GetSeasons and pick one for GetSeasonEpisodes.
	ErrSeasonRequired = scraper.ErrSeasonRequired
	// ErrSeasonNotFound is a season the series does not have.
	ErrSeasonNotFound = scraper.ErrSeasonNotFound
	// ErrAudioUnavailable is an audio the title does not offer, or offers on no
	// server that plays. The other audio is never played in its place.
	ErrAudioUnavailable = scraper.ErrAudioUnavailable
)

// ErrNotCatalog is a season call on a source that is not a movie and series
// catalogue (Source.IsCatalog): anime sources list episodes with
// GetAnimeEpisodes.
var ErrNotCatalog = errors.New("source has no seasons: use GetAnimeEpisodes")
