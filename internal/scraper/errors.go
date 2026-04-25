// Package scraper provides shared scraper guards and error helpers.
package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable or blocked and cannot provide the expected payload.
var ErrSourceUnavailable = errors.New("source unavailable")

// ErrInvalidStreamURL is returned when a scraper extracts a value that is not a
// valid absolute playback URL.
var ErrInvalidStreamURL = errors.New("invalid stream url")

// checkHTTPStatus wraps blocking upstream statuses with ErrSourceUnavailable so
// callers can differentiate provider-side issues from local parsing failures.
func checkHTTPStatus(resp *http.Response, source string) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewHTTPStatusError(source, "http", resp.StatusCode)
	}
	return nil
}

// checkHTMLResponse detects HTML challenge or error pages where JSON payloads
// are expected.
func checkHTMLResponse(resp *http.Response, body []byte, source string) error {
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return NewBlockedChallengeError(source, "http", "returned HTML instead of JSON", nil)
	}

	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '<' {
		return NewBlockedChallengeError(source, "http", "returned HTML instead of JSON", nil)
	}

	return nil
}

// checkChallengeDocument detects common Cloudflare/challenge pages in HTML
// responses that should be classified as a source-unavailable condition.
func checkChallengeDocument(doc *goquery.Document, source string) error {
	title := strings.ToLower(strings.TrimSpace(doc.Find("title").First().Text()))
	if strings.Contains(title, "just a moment") {
		return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
	}

	if doc.Find("#cf-wrapper").Length() > 0 || doc.Find("#challenge-form").Length() > 0 {
		return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
	}

	body := strings.ToLower(doc.Text())
	if strings.Contains(body, "cf-error") || strings.Contains(body, "cloudflare") {
		return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
	}

	return nil
}

// validateStreamURL ensures extracted playback URLs are absolute HTTP(S) URLs.
func validateStreamURL(rawURL, source string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("%s returned malformed stream URL: %w", source, ErrInvalidStreamURL)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s returned unsupported stream URL scheme %q: %w", source, parsed.Scheme, ErrInvalidStreamURL)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("%s returned stream URL without host: %w", source, ErrInvalidStreamURL)
	}

	return parsed.String(), nil
}
