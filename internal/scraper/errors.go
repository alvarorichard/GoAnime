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

// checkHTTPStatus returns a wrapped ErrSourceUnavailable for HTTP status codes
// that indicate the upstream source is blocking access (403 Forbidden, 429 Too
// Many Requests, 503 Service Unavailable). Other non-2xx codes are returned as
// plain errors so callers can distinguish them.
func checkHTTPStatus(resp *http.Response, source string) error {
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return fmt.Errorf("%s returned status %d (source blocked?): %w", source, resp.StatusCode, ErrSourceUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned status %d", source, resp.StatusCode)
	}
	return nil
}

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
