package netx

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Clearing a Cloudflare challenge needs a real browser, and more than one source
// needs that now.
//
// SuperFlix has driven one for a while; Goyabu started serving a managed
// challenge too (403, cf-mitigated: challenge, "Just a moment…"), which no
// amount of header tuning gets past. The browser machinery is not
// SuperFlix-specific — it launches Chrome, waits out the gate and collects
// cf_clearance — but it lives in that package, and a scraper reaching across
// into a sibling provider would be the wrong shape.
//
// So the capability is declared here, where every scraper already looks, and
// whoever owns a browser registers it. netx keeps no dependency on any provider.

// ChallengeSolveResult is what clearing a gate yields: the cookies that prove it
// and the User-Agent they are bound to.
//
// The User-Agent matters as much as the cookies. Cloudflare binds cf_clearance
// to the UA that solved the challenge, so a client that keeps sending its own UA
// afterwards is re-challenged on every request.
type ChallengeSolveResult struct {
	Cookies   []*http.Cookie
	UserAgent string
	FinalURL  string
}

// RevealPolicy says when the solver window may appear on screen.
//
// A browser window is a cost paid by the user, so it is not a preference set
// once and forgotten — it is decided by what the gate actually does. Measured
// against goyabu.io on 2026-09-24, three consecutive solves:
//
//	cold profile, window hidden    never cleared in 90s
//	warm profile, window hidden    cleared in 1.8s
//	warm profile, window visible   cleared in 1.6s
//
// A hidden window is therefore enough almost always, and useless exactly when
// the profile has no clearance yet. That is the whole case for RevealWhenStuck:
// stay out of the way, and surface only once staying hidden is provably keeping
// a solvable challenge away from the only one who can solve it.
type RevealPolicy int

const (
	// RevealWhenStuck keeps the window off screen while the challenge is
	// self-solving and surfaces it only when it is not going to. This is what a
	// source should ask for unless it has a reason not to.
	RevealWhenStuck RevealPolicy = iota
	// RevealAlways puts the window on screen immediately, for a gate known to
	// need a human every time.
	RevealAlways
	// RevealNever keeps it hidden even when stuck, failing instead, for runs
	// where nobody is watching the screen anyway.
	RevealNever
)

func (p RevealPolicy) String() string {
	switch p {
	case RevealAlways:
		return "always"
	case RevealNever:
		return "never"
	default:
		return "when-stuck"
	}
}

// ChallengeSolver clears an interstitial for targetURL and returns the proof.
type ChallengeSolver interface {
	// SolveChallenge navigates to targetURL in a real browser and returns once
	// the challenge markup is gone, or fails within timeout.
	//
	// reveal says when the window may be shown; see RevealPolicy.
	SolveChallenge(ctx context.Context, targetURL string, timeout time.Duration, reveal RevealPolicy) (*ChallengeSolveResult, error)
}

var (
	challengeSolverMu sync.RWMutex
	challengeSolver   ChallengeSolver
)

// RegisterChallengeSolver installs the process-wide browser solver. Called from
// the package that owns the browser; passing nil clears it, which is how a test
// restores the previous state.
func RegisterChallengeSolver(s ChallengeSolver) {
	challengeSolverMu.Lock()
	challengeSolver = s
	challengeSolverMu.Unlock()
}

// SolveChallengeFor clears targetURL's gate through the registered solver.
//
// ok is false when no solver is registered — a build or test that never imported
// the package owning the browser. That is a plain "cannot help here", not an
// error to surface: the caller reports the challenge it already has.
func SolveChallengeFor(ctx context.Context, targetURL string, timeout time.Duration, reveal RevealPolicy) (res *ChallengeSolveResult, ok bool, err error) {
	challengeSolverMu.RLock()
	s := challengeSolver
	challengeSolverMu.RUnlock()

	if s == nil {
		return nil, false, nil
	}
	res, err = s.SolveChallenge(ctx, targetURL, timeout, reveal)
	return res, true, err
}

// ChallengeSolverAvailable reports whether a browser solver is registered, so a
// caller can skip work it could never finish.
func ChallengeSolverAvailable() bool {
	challengeSolverMu.RLock()
	defer challengeSolverMu.RUnlock()
	return challengeSolver != nil
}
