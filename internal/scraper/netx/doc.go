// Package netx holds the scraping layer's shared network plumbing: the SSRF
// guard (safe transports/dialers), HTTP/HTML response validation, and the
// typed SourceDiagnostic error surface with origin probing. It sits below the
// per-source scrapers and depends on nothing above util/models.
package netx
