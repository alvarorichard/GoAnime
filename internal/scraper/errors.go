package scraper

import "errors"

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable – for example, when an API endpoint returns an HTML page
// (block/challenge page) instead of the expected JSON payload.
var ErrSourceUnavailable = errors.New("source unavailable")
