// Package anidb is the leaf scraper for anidb.app.
//
// It exists because the AllAnime path this project used went dark: mkissa.to
// removed the `epoch`/`partB` material that AllAnime's per-request AES key was
// derived from, so every episode-source query fails at key derivation even
// though search and the episode list still answer. Upstream ani-cli reached the
// same conclusion and moved to anidb.app in 5.0.0; this package follows it.
//
// The chain is four plain requests, with no crypto, no rotating token and no
// Cloudflare challenge:
//
//	GET /browse?q=<query>                        → anime cards (HTML)
//	GET /api/frontend/anime/<id>/episodes        → {"episodes":[{id,number,…}]}
//	GET /api/frontend/episode/<id>/languages     → per-language embed URLs
//	GET <embed_url>                              → the HLS master playlist URL
//
// Identity is carried in URLs so the source registry can route an anime back
// here without extra state:
//
//	anime:   https://anidb.app/anime/<slug>-<numeric id>
//	episode: https://anidb.app/episode/<numeric id>
//
// Language: "jpn" is subbed and "eng" is dubbed. Subbed is preferred and the
// other is used as a fallback; GOANIME_ANIDB_LANG overrides the preference.
package anidb
