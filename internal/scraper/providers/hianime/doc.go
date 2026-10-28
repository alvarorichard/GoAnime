// Package hianime is the leaf scraper for hianime.at.
//
// It replaces the anidb.app path, which died the same way the AllAnime path
// before it did. On 2026-09-22 anidb.app answered 503 with an "Under
// Maintenance" body on EVERY path — the homepage, /browse and the
// /api/frontend/... endpoints alike — so nothing on that host is parseable at
// all. Upstream ani-cli had already moved to hianime.at; this package follows,
// written against the site's own traffic rather than ported from anything.
//
// The chain is four plain requests. No Cloudflare challenge, no rotating token
// and no per-request key derivation of the kind that killed AllAnime:
//
//	GET /search?keyword=<query>                   → anime cards (HTML)
//	GET /api/theme/episode/list/<animeID>         → {"status","totalItems","html"}
//	GET /api/theme/episode/servers?episodeId=<id> → {"status","html"}
//	GET <embed URL>                               → the player payload
//
// The two JSON endpoints answer with an HTML fragment inside a JSON envelope,
// which is why both a decoder and goquery appear in the same call.
//
// A server entry carries its embed URL as data-hash, which is plain base64. Of
// the servers on offer only ZokoAnime is read here: its page ships the whole
// player config in one obfuscated blob (see embed.go), while the MegaPlay
// family hands back an AES-encrypted "enc" field whose key lives in a
// third-party repository — the same shape of dependency that made AllAnime
// unmaintainable. ZokoAnime was present on every title sampled, for every
// audio track.
//
// Identity is carried in URLs so the source registry can route an anime back
// here without extra state:
//
//	anime:   https://hianime.at/<slug>-<animeID>
//	episode: https://hianime.at/watch/<slug>-<animeID>?ep=<episodeID>
//
// Audio: "sub" is subtitled and "dub" is dubbed; some titles also carry "hsub".
// Subtitled is preferred and the rest are fallbacks; GOANIME_HIANIME_AUDIO
// overrides the preference.
package hianime
