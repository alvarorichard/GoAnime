package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable – for example, when an API endpoint returns an HTML page
// (block/challenge page) instead of the expected JSON payload.
var ErrSourceUnavailable = errors.New("source unavailable")

// checkHTMLResponse returns a wrapped ErrSourceUnavailable when the response
// looks like an HTML page instead of the expected JSON.  It checks the
// Content-Type header first (most reliable), then falls back to inspecting the
// first non-whitespace byte of body. source is used in the error message.
func checkHTMLResponse(resp *http.Response, body []byte, source string) error {
	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return fmt.Errorf("%s returned HTML instead of JSON (source blocked?): %w", source, ErrSourceUnavailable)
	}
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '<' {
		return fmt.Errorf("%s returned HTML instead of JSON (source blocked?): %w", source, ErrSourceUnavailable)
	}
	return nil
}
