// Package scraper resolves media IDs from episode URLs for supported scrapers.
package scraper

import "strings"

// ExtractMediaID pulls the trailing numeric/media identifier from FlixHQ/SFlix-style URLs.
// It also works when the input is already just the raw ID.
func ExtractMediaID(urlStr string) string {
	parts := strings.Split(urlStr, "-")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
