// Package pobreflix scrapes the Pobreflix HD catalog (pobreflixtv.gay):
// search through the site's /api catalog and its sitemap fallback, title
// pages, series' season/episode lists, and the embedded MP4 servers. It is a
// leaf provider package — it depends only on netx/util/models, never on the
// dispatch layers above it.
//
// The site is an Invision Community (IPS4) install with a custom "videobox"
// app. It moved to a JSON-first theme in October 2026:
//
//   - search: GET /api?action=catalog&type=movie|series&search=… answers with
//     the record's real name, slug, TMDB id and audio. The sitemap corpus
//     (/sitemap-movies.xml and /sitemap-series.xml) remains as a browserless
//     fallback when the API is unavailable.
//   - title pages: /filme/<slug>.html and /serie/<slug>.html embed
//     initialMovieData / initialSeriesData with the authoritative TMDB id,
//     name, poster, servers and episode list.
//   - episode pages: /assistir/<slug>-t<season>-ep<episode>.html embed
//     initialEpisodeData with the episode's servers; their download endpoint
//     redirects to the final public MP4.
//
// Old-theme pages (/filmes/online/…, server-rendered #view) are still parsed
// so captured fixtures and mirrors keep working. Those titles have no
// embedded servers and play through the StartFlix panel in the api layer.
package pobreflix
