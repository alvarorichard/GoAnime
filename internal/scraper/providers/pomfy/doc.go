// Package pomfy integrates the Pomfy movie/TV catalog (pomfy.my) and the
// Byse frontend its player runs on (pomfy-cdn.shop). It is a leaf provider
// package, like topcine: it depends only on netx/util/api-movie/models, never
// on the dispatch layers above it.
//
// Pomfy does not offer a text search of its own. Its API is keyed by TMDB id:
//
//	pomfy.my/api/catalogo-ids            every registered id, movies and series
//	pomfy.my/api/watch-sources?id=…&tipo=filme|serie[&temporada&s&episodio]
//	                                     the players listed for a title
//	pomfy.my/api/filme-data|serie-data   title metadata (TMDB dump) per id
//	api.pomfy.stream/api/tmdb-season     per-season episode numbers and names
//
// so text queries are matched through the TMDB search client (internal/api/movie)
// and filtered by catalog membership. The user's TMDB_API_KEY runs first;
// without it the search falls back to the public key pool the watching site
// ships in its own browser bundle, rotated the same way, and a bare TMDB id
// skips the name search entirely. The watch pages then hand over a Byse
// video code, and the Byse captcha adds a proof-of-work gate around the
// AES-256-GCM sealed playback record that byse.go opens.
//
// What plays: each decrypted source entry is an HLS master on the Byse edge.
// Series entries carry their audio language in the URL (lang/por, lang/eng);
// the player site refers to them as "Servidor 1" (byse) and sometimes a
// "Servidor 2" (flyfile), which this package does not resolve — it surfaces
// what watch-sources lists byse-wise.
package pomfy
