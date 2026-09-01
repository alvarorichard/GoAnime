package superflix

import (
	"net/http"
	"runtime"
)

// SuperFlix's player CDN (the punycode FirePlayer hosts) hotlink-protects the
// signed media URLs with a rule that rejects any client which does not look
// like a modern Chromium tab. Confirmed live 2026-08-31 by bisecting a freshly
// signed master.txt down to the minimum header set that still answers 200:
//
//	Referer            required — the player's own /video/<hash> page
//	User-Agent         required — EXACTLY the UA that obtained the signed URL
//	Accept-Language    required — EXACTLY "en-US,en;q=0.9"
//	Sec-CH-UA-Mobile   required
//	Sec-CH-UA-Platform required
//
// Everything else the browser sends is optional, and extra headers (Range,
// Origin) are harmless: Cookie (cf_clearance and the player session included),
// Sec-Fetch-*, Accept, Accept-Encoding, Priority, and even Sec-CH-UA itself all
// make no difference. Dropping any ONE of the five above turns the same URL
// into a 403.
//
// Two of the five are matched by VALUE, not just presence — measured against
// one live URL in a single run:
//
//	Accept-Language: en-US,en;q=0.9   -> 200      User-Agent: <solving browser>   -> 200
//	Accept-Language: en-US            -> 403      User-Agent: Chrome/140 (Linux)  -> 403
//	Accept-Language: en               -> 403      User-Agent: Chrome/120 (Win)    -> 403
//	Accept-Language: pt-BR,pt;q=0.9   -> 403      User-Agent: Firefox             -> 403
//	Accept-Language: *                -> 403      User-Agent: libmpv              -> 403
//
// So the User-Agent cannot be a constant here: it is whichever UA fetched the
// URL, carried alongside it (SuperFlixStreamResult.UserAgent) and replayed by
// every consumer.
//
// This is why SuperFlix "stopped working" without any visible error: the signed
// URL was fine and the browser could refetch it all day, but every consumer
// outside the browser — the liveness probe, mpv, ffmpeg — sent Referer and
// User-Agent only and got a 403. The probe then classified the stream as a dead
// host and discarded a perfectly good solve before mpv ever ran.
const (
	// cdnAcceptLanguage must match Chromium's default byte for byte; "en-US",
	// "en" and "*" are all rejected. The embedded comma is why the player uses
	// mpv's --http-header-fields-append (one header per option) instead of the
	// comma-joined --http-header-fields, which would split this value into
	// "Accept-Language: en-US" plus a junk field named "en;q=0.9".
	cdnAcceptLanguage = "en-US,en;q=0.9"
	// cdnSecCHUAMobile is the client hint for "not a mobile device". The `?0`
	// form is a structured-header boolean, not a typo.
	cdnSecCHUAMobile = "?0"
)

// cdnSecCHUAPlatform reports the Sec-CH-UA-Platform value for the host OS. The
// value must be a quoted string per the client-hints spec; an unquoted one is
// not a valid structured header.
func cdnSecCHUAPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return `"Windows"`
	case "darwin":
		return `"macOS"`
	default:
		return `"Linux"`
	}
}

// CDNPlaybackHeaders returns the headers every consumer of a signed SuperFlix
// media URL must send, keyed for direct use as an http.Header.
//
// referer must be the player's /video/<hash> page (see playerRefererFor); the
// bare player origin is rejected. userAgent must be the UA that obtained the
// signed URL — it falls back to the package default only so a caller that has
// lost track of it still sends a well-formed request, not because any UA works.
func CDNPlaybackHeaders(referer, userAgent string) http.Header {
	if userAgent == "" {
		userAgent = SuperFlixUserAgent
	}
	h := http.Header{}
	if referer != "" {
		h.Set("Referer", referer)
	}
	h.Set("User-Agent", userAgent)
	h.Set("Accept-Language", cdnAcceptLanguage)
	h.Set("Sec-CH-UA-Mobile", cdnSecCHUAMobile)
	h.Set("Sec-CH-UA-Platform", cdnSecCHUAPlatform())
	return h
}

// CDNPlaybackHeaderFields returns the same contract as "Name: value" strings.
//
// One header per element, because Accept-Language's value contains a comma:
// callers must pass each element to its own mpv --http-header-fields-append (or
// join them with CRLF for ffmpeg's -headers), never comma-join them into a
// single --http-header-fields.
//
// Exported because the player and downloader are the consumers that actually
// fetch the media; keeping the list here rather than in each of them means the
// next CDN rule change is one edit, not three.
func CDNPlaybackHeaderFields(referer, userAgent string) []string {
	h := CDNPlaybackHeaders(referer, userAgent)
	// Fixed order so the mpv argument is stable and testable.
	names := []string{"Referer", "User-Agent", "Accept-Language", "Sec-Ch-Ua-Mobile", "Sec-Ch-Ua-Platform"}
	fields := make([]string, 0, len(names))
	for _, n := range names {
		if v := h.Get(n); v != "" {
			fields = append(fields, canonicalCDNHeaderName(n)+": "+v)
		}
	}
	return fields
}

// canonicalCDNHeaderName restores the client-hint capitalization. Go's
// textproto canonicalization lowercases "CH" to "Ch", which is valid on the
// wire (header names are case-insensitive) but confusing to read in an mpv
// command line next to the browser's own request.
func canonicalCDNHeaderName(n string) string {
	switch n {
	case "Sec-Ch-Ua-Mobile":
		return "Sec-CH-UA-Mobile"
	case "Sec-Ch-Ua-Platform":
		return "Sec-CH-UA-Platform"
	default:
		return n
	}
}

// applyCDNPlaybackHeaders sets the CDN contract on an outgoing request,
// replacing whatever decorateRequest put there.
func applyCDNPlaybackHeaders(req *http.Request, referer, userAgent string) {
	for name, values := range CDNPlaybackHeaders(referer, userAgent) {
		req.Header.Set(name, values[0])
	}
}
