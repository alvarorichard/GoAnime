package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Goyabu's default follows whether a challenge solver is registered.
//
// It is the only source behind a bot gate (a Cloudflare managed challenge).
// Without a solver it can only report itself blocked on a flagged network, so
// it stays out of the search rather than slowing every query. With one — the
// optional cfsolve tool, registered at startup when installed — it joins the
// search with no setup. Measured 2026-11-01, cfsolve clears the gate in about
// 8s and GoAnime's own client accepts the resulting clearance.
//
// GOANIME_ENABLED_SOURCES=goyabu still forces it on for an unchallenged
// network, and GOANIME_DISABLED_SOURCES=goyabu still forces it off.

// stubSolver is a registered solver that is never asked to solve anything.
type stubSolver struct{}

func (stubSolver) SolveChallenge(context.Context, string, time.Duration, netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	return nil, errors.New("stub")
}

func withSolver(t *testing.T, on bool) {
	t.Helper()
	var s netx.ChallengeSolver
	if on {
		s = stubSolver{}
	}
	netx.RegisterChallengeSolver(s)
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
}

func goyabuDescriptor(t *testing.T) source.Descriptor {
	t.Helper()
	src, ok := source.Registered(source.Goyabu)
	require.True(t, ok, "the Goyabu provider is not registered any more")
	return src.Describe()
}

func TestGoyabu_IsOffByDefaultWithoutASolver(t *testing.T) {
	withSolver(t, false)
	d := goyabuDescriptor(t)
	assert.True(t, d.DefaultDisabled,
		"without a solver Goyabu can only be blocked on a challenged network, so it must not join the search")
	assert.False(t, source.IsSearchEnabled(d),
		"with no opt-in and no solver, Goyabu must not join the search")
	// But it must still be RESOLVABLE, or an anime already saved as Goyabu
	// stops matching its own Kind and falls through to the URL — which is how
	// one of them resolved to AnimeFire while this was being written.
	assert.True(t, source.IsEnabled(d),
		"turning a source off must not make its existing entries route somewhere else")
}

func TestGoyabu_IsOnByDefaultWithASolver(t *testing.T) {
	withSolver(t, true)
	d := goyabuDescriptor(t)
	assert.False(t, d.DefaultDisabled, "a registered solver is the whole reason to search Goyabu by default")
	assert.True(t, source.IsSearchEnabled(d), "Goyabu must join the search with no opt-in once a solver exists")
}

func TestGoyabu_ComesBackWhenAskedFor(t *testing.T) {
	withSolver(t, false)
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	assert.True(t, source.IsSearchEnabled(goyabuDescriptor(t)),
		"the opt-in must work, or this is a removal wearing a default's clothes")
}

// Disabled by default must still mean disable-able: the kill-switch has to win
// over an opt-in, so someone who enabled it globally can still turn it off for
// one run.
func TestGoyabu_KillSwitchStillWins(t *testing.T) {
	withSolver(t, true)
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	t.Setenv("GOANIME_DISABLED_SOURCES", "goyabu")
	d := goyabuDescriptor(t)
	assert.False(t, source.IsSearchEnabled(d))
	assert.False(t, source.IsEnabled(d), "the kill-switch is the one that makes it unusable")
}

// The sources that carry the catalogue must NOT have picked up the same
// default. This is the test that fails if someone copies the descriptor.
func TestTheWorkingSourcesStayOn(t *testing.T) {
	for _, kind := range []source.SourceKind{source.AnimeFire, source.HiAnime, source.StartFlix} {
		src, ok := source.Registered(kind)
		require.Truef(t, ok, "%s is not registered", kind)
		d := src.Describe()
		assert.Falsef(t, d.DefaultDisabled, "%s must stay on by default", kind)
		assert.Truef(t, source.IsSearchEnabled(d), "%s must join the search with no opt-in", kind)
	}
}
