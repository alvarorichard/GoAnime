package source

import (
	"context"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Optional capabilities (Model C, ARCHITECTURE.md §2).
//
// A Source implements one of the interfaces below ONLY when it genuinely has
// that capability. The dispatch layer discovers them by type assertion, so a
// simple anime source never carries movie/season or browser methods it doesn't
// need. They were introduced with SuperFlix — the first source to require them
// — per §4's incremental rule: "introduce those interfaces when wiring
// SuperFlix, not before."
//
// The type assertion IS the capability signal; helpers below wrap the assertion
// so callers get an explicit, logged decision (R5: a missing capability is
// visible, never a silent no-op).

// Seasoned marks a source whose catalog is organized into seasons (movie/TV,
// e.g. SuperFlix). It generalizes the old providers.HasSeasons() bool: instead
// of every provider carrying a HasSeasons() that returns false, dispatch asks
// IsSeasoned and only genuinely-seasoned sources answer true.
type Seasoned interface {
	// HasSeasons reports the source organizes content into seasons.
	HasSeasons() bool
}

// Searchable marks a source that can be searched by free-text query. Every
// live source is searchable today, but keeping it optional means a
// stream-only or URL-only source could exist without carrying a Search method.
// The registry-wide search (providers.SearchAll) fans out over the sources
// that implement this.
type Searchable interface {
	// Search returns matches for the query, already language-tagged for the
	// source. Implementations should be safe to call concurrently.
	Search(ctx context.Context, query string) ([]*models.Anime, error)
}

// BrowserGated marks a source that must drive a headed browser to clear a bot
// gate (e.g. SuperFlix's Cloudflare Turnstile). Pure-HTTP anime sources don't
// implement it, so they never carry browser methods.
type BrowserGated interface {
	// WarmUp readies the headed-browser machinery and reports whether this
	// environment can run it. It returns a non-nil error when the browser
	// cannot be driven here (e.g. no graphical display), letting dispatch fail
	// fast with a clear reason instead of deep inside the solve. Implementations
	// must stay cheap and side-effect-light — no eager solve on the happy path.
	WarmUp(ctx context.Context) error
}

// ActiveSources returns the enabled registered sources ordered by Priority.
// Config-disabled sources (S1 kill-switch) are excluded. Registry-wide
// operations — like the search fan-out — iterate this instead of a hardcoded
// source list, so adding a source needs no edit here.
func ActiveSources() []Source {
	return registeredByPriority()
}

// IsSeasoned reports whether src advertises season organization. The type
// assertion is the capability signal; a source that doesn't implement Seasoned
// is simply not seasoned.
func IsSeasoned(src Source) bool {
	s, ok := src.(Seasoned)
	return ok && s.HasSeasons()
}

// IsBrowserGated reports whether src needs a headed browser at all — a cheap
// predicate (no WarmUp) for callers that only need to decide, e.g., whether to
// skip the source in a non-interactive/headless batch.
func IsBrowserGated(src Source) bool {
	_, ok := src.(BrowserGated)
	return ok
}

// WarmUp prepares src's browser machinery when it is BrowserGated, and is an
// explicit (debug-logged) no-op otherwise — never a silent skip (R5). A
// non-nil return means a browser-gated source cannot run in this environment;
// the caller should surface it rather than attempt a doomed fetch.
func WarmUp(ctx context.Context, src Source) error {
	gated, ok := src.(BrowserGated)
	if !ok {
		util.Debug("source is not browser-gated; no warm-up needed", "kind", src.Describe().Kind)
		return nil
	}
	util.Debug("warming up browser-gated source", "kind", src.Describe().Kind)
	return gated.WarmUp(ctx)
}
