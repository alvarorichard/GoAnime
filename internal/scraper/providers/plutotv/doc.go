// Package plutotv reads Pluto TV's on-demand catalog: movies and series, free
// and without an account.
//
// The site is a client of its own JSON services, so this package is too:
//
//	boot    https://boot.pluto.tv/v4/start          anonymous session: a JWT, the
//	                                                service hosts, the stitcher query
//	search  <search>/v1/search?q=…                  movies, series, channels, timelines
//	series  <vod>/v4/vod/series/<id>/seasons        seasons and their episodes
//	stream  <stitcher>/v2/stitch/hls/episode/<id>/master.m3u8
//
// The catalog depends on the country the request comes from; nothing here
// tries to change that.
//
// # Playback
//
// The stitcher answers with an HLS playlist in which the title is cut into
// ranges (hls/0-498332/, hls/498333-941317/, …, hls/1295297-end/) with ad
// creatives spliced between them. The same directory also serves the title as
// one uninterrupted playlist, hls/0-end/hls_<bitrate>.m3u8, and that is the
// URL Stream returns: the title alone, audio and video in the same segments.
// When a title does not follow that layout the stitched playlist is returned
// instead, so it still plays, with the service's own breaks.
//
// Segments are AES-128 with the key served next to them, which every HLS
// player handles. A playlist that asks for anything else (SAMPLE-AES, a DRM
// key system) is refused with ErrProtected rather than worked around.
//
// # Identifiers
//
// The URLs carried in models.Anime and models.Episode are the package's own
// spelling of a title, built from the catalog ids:
//
//	series:  https://pluto.tv/on-demand/series/<seriesID>
//	movie:   https://pluto.tv/on-demand/movies/<movieID>
//	episode: https://pluto.tv/on-demand/series/<seriesID>/episode/<episodeID>
package plutotv
