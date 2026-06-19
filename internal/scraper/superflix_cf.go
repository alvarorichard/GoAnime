package scraper

import (
	"bytes"
	"net/http"
)

// cfChallengeMarkers are byte substrings that, when present in a response body,
// indicate the page is a Cloudflare interactive challenge (Turnstile / managed
// challenge / JS challenge) rather than the real content.
//
// Why bytes (not string Contains): the body is read once into a []byte buffer
// in the HTTP path; converting to string on every check would allocate a copy
// for a hot fallback decision.
var cfChallengeMarkers = [][]byte{
	[]byte("/cdn-cgi/challenge-platform"),
	[]byte("challenges.cloudflare.com/turnstile"),
	[]byte("cf_chl_opt"),
	[]byte("__cf_chl_"),
	[]byte("Just a moment..."),
	[]byte("Checking your browser before accessing"),
	[]byte("cf-mitigated"),
	// SuperFlix's host serves a custom "Verificação" gate that embeds a
	// Turnstile widget (form id cf-turnstile-form) and GET-submits the token.
	// These two appear only on that gate, never on the real player page, so
	// they double as the "still gated" signal for the browser solver.
	[]byte("cf-turnstile-form"),
	[]byte("<title>Verificação</title>"),
}

// restrictedEmbedMarkers identify SuperFlix's post-gate "Visualização Externa /
// Acesso Restrito" page — served (200, no CF markers) once the gate is already
// cleared. The real content is reachable only by following its embed iframe, so
// this page must ALSO be routed through the browser solver (which performs the
// embed read), not returned to the caller as-is.
var restrictedEmbedMarkers = [][]byte{
	[]byte("Visualiza\xc3\xa7\xc3\xa3o Externa"), // "Visualização Externa" (UTF-8)
	[]byte("Acesso Restrito"),
	[]byte("ACESSO RESTRITO"),
}

// isRestrictedEmbedPage reports whether body is the post-gate restricted embed
// page (content only available via its embed iframe).
func isRestrictedEmbedPage(body []byte) bool {
	for _, m := range restrictedEmbedMarkers {
		if bytes.Contains(body, m) {
			return true
		}
	}
	return false
}

// bodyHasChallengeMarker reports whether body contains any known CF/Turnstile
// interstitial marker. Shared by IsCloudflareChallenge (HTTP path) and the
// browser solver (which polls a live page's HTML to know when the gate clears).
func bodyHasChallengeMarker(body []byte) bool {
	for _, m := range cfChallengeMarkers {
		if bytes.Contains(body, m) {
			return true
		}
	}
	return false
}

// IsCloudflareChallenge reports whether the response looks like a Cloudflare
// interstitial that requires a real browser to solve.
//
// Logic:
//   - status 403/503 with HTML body → very likely a challenge
//   - any status with a known marker substring in body → challenge
//   - presence of "cf-mitigated: challenge" response header → challenge
//
// SuperFlix's real API can legitimately return 403 (rate limit) without a
// challenge — those bodies are JSON, not HTML, so the marker check filters
// them out.
func IsCloudflareChallenge(resp *http.Response, body []byte) bool {
	if resp == nil {
		return false
	}

	if mit := resp.Header.Get("cf-mitigated"); mit == "challenge" {
		return true
	}

	if bodyHasChallengeMarker(body) {
		return true
	}

	// Fallback: 403/503 + HTML body (very short JSON 403s already excluded
	// above by the marker check failing).
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusServiceUnavailable {
		trimmed := bytes.TrimLeft(body, " \t\r\n\xef\xbb\xbf")
		if len(trimmed) > 0 && trimmed[0] == '<' {
			// HTML 403/503 without explicit CF markers — still likely a CF
			// edge page (some CF deployments strip the markers). Treat as
			// challenge so we retry through the browser.
			return true
		}
	}

	return false
}
