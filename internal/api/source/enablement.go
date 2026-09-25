package source

import (
	"slices"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Source enablement (ARCHITECTURE.md §7 S1 — manual kill-switch).
//
// A registered source still participates in resolution ONLY if it is enabled.
// Disabling is a config decision (no rebuild): a source is off when it is
// listed in GOANIME_DISABLED_SOURCES, or when its Descriptor is DefaultDisabled
// and it was NOT opted in via GOANIME_ENABLED_SOURCES. The explicit-list
// parsing lives in util so the search layer honors the same switch without an
// import cycle (see util.SourceDisabled).

// "Disabled" comes in two strengths, because "do not search this" and "I cannot
// play what I already have" are different requests.
//
// GOANIME_DISABLED_SOURCES is the hard one: the source is gone, and anything
// pointing at it stops resolving. DefaultDisabled is the soft one: the source
// stays known, so an entry already tagged with it still routes home, but it does
// not join the search unless asked for.
//
// The distinction is not theoretical. Marking Goyabu DefaultDisabled with one
// predicate covering both made an anime saved as Goyabu resolve to AnimeFire —
// the Kind no longer matched, so resolution fell through to the URL, which
// happened to contain "animefire". Turning a source off had quietly started
// playing other sources' links for it.

// IsEnabled reports whether the source described by d may be used at all: it is
// false only for the explicit kill-switch. Resolution asks this, so a source
// that ships off by default still recognises its own entries.
func IsEnabled(d Descriptor) bool {
	return !util.SourceDisabled(string(d.Kind))
}

// IsSearchEnabled reports whether d joins the search fan-out. A source that
// ships off by default has to be asked for by name.
func IsSearchEnabled(d Descriptor) bool {
	if !IsEnabled(d) {
		return false
	}
	if d.DefaultDisabled {
		return util.SourceForceEnabled(string(d.Kind))
	}
	return true
}

// filterEnabled returns only the currently-enabled sources, logging each
// skipped one so a disabled source is visible, never silently dropped (R5).
func filterEnabled(srcs []Source) []Source {
	out := make([]Source, 0, len(srcs))
	for _, s := range srcs {
		if IsEnabled(s.Describe()) {
			out = append(out, s)
			continue
		}
		util.Debug("source disabled by config; skipped in resolution", "kind", s.Describe().Kind)
	}
	return out
}

// Enabled returns the registered source for kind only if it is currently
// enabled. Dispatch paths use this (instead of Registered) so a disabled source
// is never selected — including the best-effort fallback.
func Enabled(kind SourceKind) (Source, bool) {
	s, ok := Registered(kind)
	if !ok || !IsEnabled(s.Describe()) {
		return nil, false
	}
	return s, true
}

// DisabledSources returns the registered sources that are currently disabled by
// config, sorted by Kind. Callers (e.g. app startup) use it to log a one-time,
// visible confirmation that a manual kill-switch took effect.
func DisabledSources() []SourceKind {
	registryMu.RLock()
	var out []SourceKind
	for k, s := range registry {
		if !IsEnabled(s.Describe()) {
			out = append(out, k)
		}
	}
	registryMu.RUnlock()

	slices.Sort(out)
	return out
}
