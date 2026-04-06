package scraper

import (
	"errors"
	"fmt"
)

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable – for example, when an API endpoint returns an HTML page
// (block/challenge page) instead of the expected JSON payload.
var ErrSourceUnavailable = errors.New("source unavailable")

// checkHTMLResponse returns a wrapped ErrSourceUnavailable when body starts
// with '<', indicating the API endpoint returned HTML (e.g. a block page)
// instead of the expected JSON. source is used in the error message.
func checkHTMLResponse(body []byte, source string) error {
	if len(body) > 0 && body[0] == '<' {
		return fmt.Errorf("%s returned HTML instead of JSON (source blocked?): %w", source, ErrSourceUnavailable)
	}
	return nil
}
