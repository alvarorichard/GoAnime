// Package pobreflix scrapes the Pobreflix HD catalog (pobreflixtv.sale):
// search over the site's own sitemap shards, title pages and the series'
// season/episode lists. It is a leaf provider package — it depends only on
// netx/util/models, never on the dispatch layers above it.
//
// The site is an Invision Community (IPS4) install with a custom "videobox"
// app. Everything a title needs is server-rendered — the movie's page carries
// <div id="view" data-video-id="…" data-tmdb-id="…">, a series page lists its
// seasons and episodes with Portuguese names and TMDB posters, and every record
// URL carries its audio (-dublado/-legendado). Two things the site does not
// offer to anonymous visitors, though:
//
//   - Search. /buscar runs client-side and the IPS quick-search endpoint is
//     members-only (checked 2026-10-16). The catalog does expose the whole grid
//     through content-sitemap shards (sitemap.php lists 74 shard files of 1000
//     URLs each, newest record first), so a search builds an index of those
//     shards — bounded parallel fetch, a disk cache with a refresh window, and
//     partial results survive a timed-out first build.
//   - Playback. The player is assembled client-side by an obfuscated scripts
//     bundle; no host URL is in the HTML. Like TopCine, this package stops at
//     the catalog: every page names the title's TMDB id, and the api layer
//     plays the title through the StartFlix panel keyed by that id.
package pobreflix
