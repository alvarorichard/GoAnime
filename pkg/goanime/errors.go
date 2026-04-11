// Package goanime provides a high-level client for searching and streaming anime.
package goanime

import "github.com/alvarorichard/Goanime/internal/scraper"

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable (e.g. rate-limited or behind a challenge page).
// Callers can use errors.Is to detect and handle this case gracefully.
var ErrSourceUnavailable = scraper.ErrSourceUnavailable
