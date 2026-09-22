package providers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// A failed search has to be reportable without dumping the whole error chain
// at the user.
//
// The chain is genuinely useful — errors.Is/As need it, and the debug log is
// where a maintainer looks — but rendering it is not. Flattening the
// diagnostics into the message alongside the raw errors produced exactly one
// line of signal wrapped in three layers of repetition:
//
//	Search failed for "dexter": failed to search: no results for "dexter" (all
//	sources failed) — SuperFlix blocked the request: HTTP 429/challenge; AniDB
//	temporarily unavailable: HTTP 503: SuperFlix: server returned: 429 Too Many
//	Requests
//	AniDB: AniDB search: upstream unavailable with HTTP 503
//
// So the failure is a VALUE, not a string. It keeps one short line per source
// for display and the untouched causes underneath for errors.Is, and the UI
// layer decides how much to show.

// SourceFailure is one source's reason for not answering a search.
type SourceFailure struct {
	// Kind names the source, e.g. "SuperFlix".
	Kind source.SourceKind
	// Reason is one short, user-facing clause: "rate limited (HTTP 429)".
	// It does NOT repeat the source name — the caller prints that.
	Reason string
	// RateLimited marks a source that refused us rather than failed, so the
	// caller can say "wait" instead of "something broke".
	RateLimited bool
	// Err is the untouched cause, kept for errors.Is/As.
	Err error
}

// SearchFailure reports that every source declined a search.
type SearchFailure struct {
	Query   string
	Sources []SourceFailure
}

// Error is the one-line form, used when something prints the error directly.
// Callers that can do better should use errors.As and render Sources.
func (f *SearchFailure) Error() string {
	if f == nil {
		return "search failed"
	}
	if len(f.Sources) == 0 {
		return fmt.Sprintf("no results for %q", f.Query)
	}
	parts := make([]string, 0, len(f.Sources))
	for _, s := range f.Sources {
		parts = append(parts, fmt.Sprintf("%s %s", s.Kind, s.Reason))
	}
	return fmt.Sprintf("no results for %q (all sources failed): %s",
		f.Query, strings.Join(parts, "; "))
}

// Detail is the full form for the debug log: every short reason followed by the
// untouched cause. Error() deliberately omits the causes — they are noise on a
// terminal — so this is where a maintainer reading the log gets them back.
func (f *SearchFailure) Detail() string {
	if f == nil {
		return "search failed"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "no results for %q", f.Query)
	for _, s := range f.Sources {
		fmt.Fprintf(&b, "\n  %s %s", s.Kind, s.Reason)
		if s.Err != nil {
			fmt.Fprintf(&b, " | cause: %v", s.Err)
		}
	}
	return b.String()
}

// Unwrap exposes every cause, so errors.Is and errors.As still reach through a
// SearchFailure to the original errors.
func (f *SearchFailure) Unwrap() []error {
	if f == nil {
		return nil
	}
	errs := make([]error, 0, len(f.Sources))
	for _, s := range f.Sources {
		if s.Err != nil {
			errs = append(errs, s.Err)
		}
	}
	return errs
}

// RateLimited reports whether any source refused us for talking too much. That
// is the one failure mode with an action attached — wait — so it is worth
// telling the user apart from a source being broken.
func (f *SearchFailure) RateLimited() bool {
	if f == nil {
		return false
	}
	for _, s := range f.Sources {
		if s.RateLimited {
			return true
		}
	}
	return false
}

// errSearchRateLimited matches a request GoAnime suppressed itself because the
// host had asked it to back off (superflix.ErrRateLimited). Declared as a
// variable so this package does not import the scraper just to name it; the
// sentinel is matched by message because errors.Is needs the concrete value and
// pulling it in would invert the dependency.
var errSearchRateLimited = errors.New("superflix: rate limited, backing off")

// describeFailure reduces a source's error to one short clause plus the
// rate-limited flag. The diagnostic machinery already classifies the error for
// the circuit breaker; this is the same classification, phrased for a person.
func describeFailure(diag *netx.SourceDiagnostic, err error) (reason string, rateLimited bool) {
	// A request the back-off suppressed never reached the host, so calling it
	// a failure of that host would be wrong.
	if err != nil && strings.Contains(err.Error(), errSearchRateLimited.Error()) {
		return "is rate limiting this network — wait a few minutes", true
	}
	if diag == nil {
		return "failed", false
	}

	switch {
	case diag.StatusCode == http.StatusTooManyRequests:
		return "is rate limiting this network — wait a few minutes", true
	case diag.Kind == netx.DiagnosticBlockedChallenge:
		return "blocked the request (captcha/challenge)", false
	case diag.Kind == netx.DiagnosticSourceUnavailable && diag.StatusCode > 0:
		return fmt.Sprintf("is temporarily unavailable (HTTP %d)", diag.StatusCode), false
	case diag.Kind == netx.DiagnosticSourceUnavailable:
		return "is temporarily unavailable", false
	case diag.Kind == netx.DiagnosticParserBroken:
		return "answered, but GoAnime could not read it", false
	}
	return "failed", false
}
