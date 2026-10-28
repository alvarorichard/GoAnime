// Package topcine scrapes the TopCine movie/TV catalog (topcine3.site): search,
// title pages (kind, TMDB id, seasons and episode names) and the language
// check of the player those pages embed. It is a leaf provider package — it
// depends only on netx/util/models, never on the dispatch layers above it.
//
// TopCine does not host video. Every title page iframes azullog.top, keyed by
// TMDB id, and azullog hands both audio tracks to a single host, 1take.top,
// which sits behind a Cloudflare Turnstile (checked 2026-10-09). GoAnime ships
// no browser and does not solve challenges, so this package stops at the
// catalog: the api layer plays TopCine titles by their TMDB id through the
// StartFlix video panel, whose player hosts resolve over plain HTTP.
package topcine
