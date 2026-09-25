package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Goyabu ships off, for now.
//
// It is the only source that needs a browser, and on a network Cloudflare has
// flagged there is currently no way past its gate — measured 2026-09-25, the
// gate cleared in the browser and the clearance was refused on replay by every
// transport available to us. So it costs a Chrome launch and returns nothing.
//
// These tests pin the DEFAULT, not the capability, and they are written that
// way on purpose: the scraper is complete and correct, the source works on a
// network that is not flagged, and GOANIME_ENABLED_SOURCES=goyabu brings it
// back the moment a bypass exists. See the descriptor in source_providers.go
// for what has already been ruled out.
//
// When that day comes, the change is one line — DefaultDisabled goes away — and
// TestGoyabu_IsOffByDefault is the test that should fail and be deleted with
// it. It is a marker, not a verdict.

func goyabuDescriptor(t *testing.T) source.Descriptor {
	t.Helper()
	src, ok := source.Registered(source.Goyabu)
	require.True(t, ok, "the Goyabu provider is not registered any more")
	return src.Describe()
}

func TestGoyabu_IsOffByDefault(t *testing.T) {
	d := goyabuDescriptor(t)
	assert.True(t, d.DefaultDisabled,
		"Goyabu must ship off: it is the only source that launches a browser, for results AnimeFire already covers")
	assert.False(t, source.IsSearchEnabled(d),
		"with no opt-in, Goyabu must not join the search")
	// But it must still be RESOLVABLE, or an anime already saved as Goyabu
	// stops matching its own Kind and falls through to the URL — which is how
	// one of them resolved to AnimeFire while this was being written.
	assert.True(t, source.IsEnabled(d),
		"turning a source off must not make its existing entries route somewhere else")
}

func TestGoyabu_ComesBackWhenAskedFor(t *testing.T) {
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	assert.True(t, source.IsSearchEnabled(goyabuDescriptor(t)),
		"the opt-in must work, or this is a removal wearing a default's clothes")
}

// Disabled by default must still mean disable-able: the kill-switch has to win
// over an opt-in, so someone who enabled it globally can still turn it off for
// one run.
func TestGoyabu_KillSwitchStillWins(t *testing.T) {
	t.Setenv("GOANIME_ENABLED_SOURCES", "goyabu")
	t.Setenv("GOANIME_DISABLED_SOURCES", "goyabu")
	d := goyabuDescriptor(t)
	assert.False(t, source.IsSearchEnabled(d))
	assert.False(t, source.IsEnabled(d), "the kill-switch is the one that makes it unusable")
}

// The sources that carry the catalogue must NOT have picked up the same
// default. This is the test that fails if someone copies the descriptor.
func TestTheWorkingSourcesStayOn(t *testing.T) {
	for _, kind := range []source.SourceKind{source.AnimeFire, source.HiAnime, source.SuperFlix} {
		src, ok := source.Registered(kind)
		require.Truef(t, ok, "%s is not registered", kind)
		d := src.Describe()
		assert.Falsef(t, d.DefaultDisabled, "%s must stay on by default", kind)
		assert.Truef(t, source.IsSearchEnabled(d), "%s must join the search with no opt-in", kind)
	}
}
