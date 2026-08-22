package util

import (
	"os"
	"strings"
)

// Source kill-switch (ARCHITECTURE.md §7 S1).
//
// A user can disable a source WITHOUT a rebuild by listing it in an env var:
//
//	GOANIME_DISABLED_SOURCES="AllAnime,Goyabu"   # turn these off
//	GOANIME_ENABLED_SOURCES="Experimental"       # opt into a DefaultDisabled one
//
// It is the manual complement to the automatic per-source circuit breaker: the
// breaker reacts to failures at runtime; this lets a human pre-empt a known-bad
// source before it wastes an attempt. The parsing lives in util (a leaf package)
// so BOTH the dispatch layer (internal/api/source) and the search layer
// (internal/scraper) can honor it without an import cycle.
const (
	disabledSourcesEnv = "GOANIME_DISABLED_SOURCES"
	enabledSourcesEnv  = "GOANIME_ENABLED_SOURCES"
)

// canonSourceToken normalizes a source name to a comparison token so the env
// list is forgiving: "AllAnime", "allanime", "Animefire.io" and "animefire"
// all collapse to the same token. Case-insensitive; drops dots (and a trailing
// ".io") and surrounding whitespace.
func canonSourceToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ".io")
	s = strings.ReplaceAll(s, ".", "")
	return s
}

// sourceListed reports whether name matches any entry in the given env var's
// comma-separated list, using canonical-token comparison.
func sourceListed(env, name string) bool {
	raw := os.Getenv(env)
	if raw == "" {
		return false
	}
	want := canonSourceToken(name)
	if want == "" {
		return false
	}
	for part := range strings.SplitSeq(raw, ",") {
		if canonSourceToken(part) == want {
			return true
		}
	}
	return false
}

// SourceDisabled reports whether the named source is turned off via
// GOANIME_DISABLED_SOURCES. Matching is case-insensitive and dot-forgiving
// (see canonSourceToken).
func SourceDisabled(name string) bool {
	return sourceListed(disabledSourcesEnv, name)
}

// SourceForceEnabled reports whether the named source is explicitly opted into
// via GOANIME_ENABLED_SOURCES — used to turn on a source whose descriptor marks
// it DefaultDisabled (off unless requested).
func SourceForceEnabled(name string) bool {
	return sourceListed(enabledSourcesEnv, name)
}
