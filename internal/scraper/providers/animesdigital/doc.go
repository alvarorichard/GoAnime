// Package animesdigital implements the plain-HTTP scraper for
// animesdigital.org.
//
// The site keeps subtitled and dubbed releases on separate title pages. Search
// merges matching pairs and stores the dubbed URL in an opaque fragment on the
// subtitled URL; the API layer asks which audio to list only when both exist.
//
// Episode pages currently expose two iframes. The supported player is
// api.anivideo.net/videohls.php: its d query parameter is the direct HLS URL.
// The second player encodes a Blogspot URL in its path, redirects to an
// api.anivideo.net/file endpoint, and returned HTTP 404 when checked on
// 2026-10-09. It is intentionally not treated as playable until it can be
// resolved reliably without a browser.
package animesdigital
