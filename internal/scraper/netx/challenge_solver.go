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

// ChallengeSolver clears an interstitial for targetURL and returns the proof.
type ChallengeSolver interface {
	// SolveChallenge navigates to targetURL in a real browser and returns once
	// the challenge markup is gone, or fails within timeout.
	//
	// visible asks for an on-screen window. It is not decoration: measured
	// 2026-09-21 against goyabu.io from a challenged IP, with the profile wiped
	// between runs, a minimized window never cleared in 70s while a visible one
	// cleared every time (~1min cold, ~6s warm). Headless never cleared at all.
	// SuperFlix's own gate is happier and passes minimized, so the caller says
	// what its source needs.
	SolveChallenge(ctx context.Context, targetURL string, timeout time.Duration, visible bool) (*ChallengeSolveResult, error)
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
func SolveChallengeFor(ctx context.Context, targetURL string, timeout time.Duration, visible bool) (res *ChallengeSolveResult, ok bool, err error) {
	challengeSolverMu.RLock()
	s := challengeSolver
	challengeSolverMu.RUnlock()

	if s == nil {
		return nil, false, nil
	}
	res, err = s.SolveChallenge(ctx, targetURL, timeout, visible)
	return res, true, err
}

// ChallengeSolverAvailable reports whether a browser solver is registered, so a
// caller can skip work it could never finish.
func ChallengeSolverAvailable() bool {
	challengeSolverMu.RLock()
	defer challengeSolverMu.RUnlock()
	return challengeSolver != nil
}
