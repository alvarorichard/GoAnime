// Package miruro is the leaf scraper for miruro.cx, the English source
// upstream ani-cli moved to on 2026-11-25 (pystardust/ani-cli#1993). GoAnime
// keeps HiAnime beside it: both still worked when this was written, and two
// English sources survive one of them going dark.
//
// The site is a SvelteKit app, and every page has a data endpoint that answers
// with the page's load data in devalue form (see devalue.go). Three requests,
// none of them behind a key or a token:
//
//	GET /search/__data.json?q=<query>           → items (AniList metadata, ids)
//	GET /watch/<id>/<slug>/__data.json?ep=1     → the episode list
//	GET /watch/<id>/<slug>/__data.json?ep=<n>   → that episode's sources
//
// A watch URL with any slug works: the app answers with a redirect to the
// canonical one, which getData follows. GetAnimeEpisodes builds the episode
// URLs with the canonical slug, so playing an episode is one request.
//
// The sources come per audio track: "ssub" (subtitles as separate files),
// "sub" (burned in) and "dub". Each track lists several providers (sun, bee,
// hop, ...) in the site's own preference order, each provider several
// servers, each server an HLS URL and the Referer/Origin its CDN checks.
// Servers that only offer an embed page are skipped. The first stream whose
// playlist answers is played; stream.go explains the order.
//
// Identity is carried in URLs so the source registry can route a title back
// here without extra state:
//
//	anime:   https://miruro.cx/info/<id>
//	episode: https://miruro.cx/watch/<id>/<slug>?ep=<number>
//
// Audio: subtitled is preferred, dubbed is the fallback; GOANIME_MIRURO_AUDIO
// overrides the preference.
package miruro
