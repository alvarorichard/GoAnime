// Package superflix scrapes the SuperFlix movie/TV source: token bootstrap,
// player-embed resolution, Cloudflare Turnstile clearance via a headed
// browser, stream caching, and TVmaze-backed episode listing. It is a leaf
// provider package — it depends only on netx/util/models, never on the
// dispatch layers above it.
package superflix
