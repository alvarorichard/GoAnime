package netx

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSourceUnavailable is returned when an upstream source is temporarily
// unavailable or blocked and cannot provide the expected payload.
var ErrSourceUnavailable = errors.New("source unavailable")

// ErrInvalidStreamURL is returned when a scraper extracts a value that is not a
// valid absolute playback URL.
var ErrInvalidStreamURL = errors.New("invalid stream url")

// CheckHTTPStatus wraps blocking upstream statuses with ErrSourceUnavailable so
// callers can differentiate provider-side issues from local parsing failures.
func CheckHTTPStatus(resp *http.Response, source string) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewHTTPStatusError(source, "http", resp.StatusCode)
	}
	return nil
}

// CheckHTMLResponse detects HTML challenge or error pages where JSON payloads
// are expected.
func CheckHTMLResponse(resp *http.Response, body []byte, source string) error {
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return NewBlockedChallengeError(source, "http", "returned HTML instead of JSON", nil)
	}

	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '<' {
		return NewBlockedChallengeError(source, "http", "returned HTML instead of JSON", nil)
	}

	return nil
}

// challengeBodyPhrases are phrases that, when present in a page's rendered
// text, strongly indicate a Cloudflare/anti-bot interstitial rather than the
// real content. The bare word "cloudflare" is intentionally excluded: real
// pages routinely mention it in footers, terms of service, or user comments
// without being challenges (issue #166: Goyabu episode pages were being
// misclassified, breaking stream extraction even when playersData was valid).
var challengeBodyPhrases = []string{
	"cf-error",
	"checking your browser before accessing",
	"ddos protection by cloudflare",
	"please enable cookies and reload",
	"performance & security by cloudflare",
	"ray id:",
	"please complete the security check to access",
	"attention required! | cloudflare",
}

// CheckChallengeDocument detects common Cloudflare/challenge pages in HTML
// responses that should be classified as a source-unavailable condition.
// Returns the matching marker via debug log when triggered so misclassifications
// are diagnosable from logs.
func CheckChallengeDocument(doc *goquery.Document, source string) error {
	title := strings.ToLower(strings.TrimSpace(doc.Find("title").First().Text()))
	if strings.Contains(title, "just a moment") ||
		strings.Contains(title, "attention required") ||
		strings.Contains(title, "access denied") {
		util.Debug("challenge detected via title", "source", source, "title", title)
		return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
	}

	if doc.Find("#cf-wrapper").Length() > 0 || doc.Find("#challenge-form").Length() > 0 ||
		doc.Find("#cf-error-details").Length() > 0 || doc.Find(".cf-browser-verification").Length() > 0 {
		util.Debug("challenge detected via cf marker element", "source", source)
		return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
	}

	body := strings.ToLower(doc.Text())
	for _, phrase := range challengeBodyPhrases {
		if strings.Contains(body, phrase) {
			util.Debug("challenge detected via body phrase", "source", source, "phrase", phrase)
			return NewBlockedChallengeError(source, "http", "returned a challenge page", nil)
		}
	}

	return nil
}

// ValidateStreamURL ensures extracted playback URLs are absolute HTTP(S) URLs.
func ValidateStreamURL(rawURL, source string) (string, error) {
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
