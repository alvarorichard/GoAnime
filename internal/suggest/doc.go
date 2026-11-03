// Package suggest powers the search box's autocomplete for anime, movies and
// series. It asks keyless public endpoints (Kitsu for anime, IMDb's suggestion
// feed for movies and series) while the user types, merges and ranks the
// answers, and caches them, so typing stays instant on repeated prefixes.
//
// Nothing here needs an API key, account or configuration, and a provider that
// is down never breaks the box: the engine answers with whatever the others
// returned and only reports an error when every provider failed.
package suggest
