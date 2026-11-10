// Package goanime exposes public client APIs and shared sentinel errors.
package goanime

import (
	"errors"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// ErrSourceUnavailable indicates that the selected upstream source is
// temporarily unavailable or blocked.
var ErrSourceUnavailable = netx.ErrSourceUnavailable

// ErrEpisodesUnsupported means the source can be searched but its episodes
// and streams are not reachable through this package: StartFlix, TopCine,
// Pobreflix and Pomfy are movie and series catalogues whose episodes are
// chosen through season and audio pickers in the command-line tool.
var ErrEpisodesUnsupported = errors.New("source does not support episode lists or stream URLs in this package")
