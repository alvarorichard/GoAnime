package goanime

import "github.com/alvarorichard/Goanime/internal/scraper"

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable (e.g. rate-limited or behind a challenge page).
var ErrSourceUnavailable = scraper.ErrSourceUnavailable
