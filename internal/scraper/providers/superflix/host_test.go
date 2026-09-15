package superflix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetHostDiscovery restores the package-global discovery state so a test can
// drive it from scratch without leaking a resolved host into its neighbours.
// Takes the same mutex the production path does, so it stays race-free next to
// tests that call base().
func resetHostDiscovery(t *testing.T) {
	t.Helper()
	set := func(host string, probed bool) {
		hostMu.Lock()
		defer hostMu.Unlock()
		hostFound, hostProbed = host, probed
	}
	hostMu.Lock()
	savedHost, savedProbed := hostFound, hostProbed
	hostMu.Unlock()

	set("", false)
	t.Cleanup(func() { set(savedHost, savedProbed) })
}

func TestLiveEmbedHost_FallsBackToCompiledDefault(t *testing.T) {
	resetHostDiscovery(t)
	assert.Equal(t, SuperFlixEmbedHost, liveEmbedHost(), "undiscovered → compiled default")
	assert.Equal(t, SuperFlixBase, liveBase())
}

func TestEnsureLiveHost_EnvOverrideWins(t *testing.T) {
	resetHostDiscovery(t)
	// A full origin and a bare host must both normalize to the bare host, so
	// users can paste either form.
	for _, pinned := range []string{"superflixapi.example", "https://superflixapi.example/"} {
		resetHostDiscovery(t)
		t.Setenv(hostEnvOverride, pinned)
		ensureLiveHost(context.Background())
		assert.Equalf(t, "superflixapi.example", liveEmbedHost(), "pinned %q", pinned)
	}
}

func TestEnsureLiveHost_DiscoveryRunsOnlyOnce(t *testing.T) {
	resetHostDiscovery(t)
	t.Setenv(hostEnvOverride, "superflixapi.first")
	ensureLiveHost(context.Background())
	require.Equal(t, "superflixapi.first", liveEmbedHost())

	// A second call must not re-probe: discovery is one round trip per process,
	// not one per request (base() calls it on every request).
	t.Setenv(hostEnvOverride, "superflixapi.second")
	ensureLiveHost(context.Background())
	assert.Equal(t, "superflixapi.first", liveEmbedHost(), "sync.Once must hold")
}

// TestBase_TestClientNeverDiscovers pins the guard that keeps unit tests
// offline: a client pointed at an httptest server, or carrying a scripted
// solver, must return its own base URL untouched.
func TestBase_TestClientNeverDiscovers(t *testing.T) {
	resetHostDiscovery(t)

	c := NewClientForTest("https://sf.test")
	assert.Equal(t, "https://sf.test", c.base(), "explicit test base is returned verbatim")

	scripted := NewSuperFlixClient()
	scripted.browserSolver = &scriptedSolver{}
	assert.Equal(t, SuperFlixBase, scripted.base(),
		"a scripted solver means no live site, so no discovery")
	hostMu.Lock()
	probed := hostProbed
	hostMu.Unlock()
	assert.False(t, probed, "no probe may have run")
}

// TestNextHopHost covers the walk's decision table: what counts as "arrived",
// what counts as "keep walking", and what aborts.
func TestNextHopHost(t *testing.T) {
	const cur = "superflixapi.sbs"

	tests := []struct {
		name      string
		status    int
		location  string
		wantHost  string
		wantDone  bool
		wantError bool
	}{
		{"200 is the live host", 200, "", cur, true, false},
		{"301 to a sibling rotates", 301, "https://superflixapi.baby/", "superflixapi.baby", false, false},
		{"302 rotates too", 302, "https://superflixapi.baby/", "superflixapi.baby", false, false},
		{"403 challenge still means live", 403, "", cur, true, false},
		{"503 challenge still means live", 503, "", cur, true, false},
		{"404 is terminal, not a rotation", 404, "", cur, true, false},
		{"3xx without Location is terminal", 301, "", cur, true, false},
		{"relative Location is not a rotation", 301, "/filme/550", cur, true, false},
		{"same-host Location is not a rotation", 301, "https://" + cur + "/x", cur, true, false},
		{"off-family redirect aborts", 301, "https://evil.test/", "", false, true},
		{"lookalike subdomain aborts", 301, "https://superflixapi.baby.evil.test/", "", false, true},
		{"unparseable Location aborts", 301, "://%%zz", "", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			host, done, err := nextHopHost(cur, tc.status, tc.location)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantHost, host)
			assert.Equal(t, tc.wantDone, done)
		})
	}
}

// TestSuperflixHostRe_RejectsOffFamilyRedirects pins the safety check: a
// retired SuperFlix domain can be re-registered by anyone, so a Location that
// leaves the family must abort discovery rather than hand a stranger our
// requests and Cloudflare cookies.
func TestSuperflixHostRe_RejectsOffFamilyRedirects(t *testing.T) {
	rejected := []string{
		"evil.test",
		"superflixapi.baby.evil.test",
		"notsuperflixapi.baby",
		"sub.superflixapi.baby",
		"superflixapi.",
		"superflixapi.baby:8080",
		"",
	}
	for _, h := range rejected {
		assert.Falsef(t, superflixHostRe.MatchString(h), "%q must be rejected", h)
	}
}

// TestDiscoveryAllowed_NeverProbesUnderGoTest is the regression for a CI
// failure that had nothing to do with the domain rotation it appeared to be
// about.
//
// Host discovery used to be gated on a guess — "this client still has the
// default base URL and the real solver, so it must be production". A unit test
// that built a client with NewSuperFlixClient() and only swapped its
// http.Client for an httptest one slipped through: `go test -short` made a live
// network call, resolved the rotated host into a package global, and an
// unrelated parallel test comparing that global against the compiled constant
// went red on three platforms at once.
//
// The gate is now testing.Testing(), which no client construction can defeat.
func TestDiscoveryAllowed_NeverProbesUnderGoTest(t *testing.T) {
	t.Setenv("GOANIME_LIVE", "")
	assert.False(t, discoveryAllowed(),
		"a test binary must never spend a network round trip discovering the host")
}

// Live tests opt back in explicitly, so the real rotation path stays covered.
func TestDiscoveryAllowed_OptInForLiveTests(t *testing.T) {
	t.Setenv("GOANIME_LIVE", "1")
	assert.True(t, discoveryAllowed())
}

// TestEnsureLiveHost_NoNetworkInTests proves the gate end to end: even a
// production-shaped client must leave the host on the compiled default while
// under `go test`, so no global state leaks between tests.
func TestEnsureLiveHost_NoNetworkInTests(t *testing.T) {
	resetHostDiscovery(t)
	t.Setenv("GOANIME_LIVE", "")
	t.Setenv(hostEnvOverride, "")

	c := NewSuperFlixClient() // default base URL, real solver: the shape that used to probe
	assert.Equal(t, SuperFlixBase, c.base(),
		"base() must not reach the network from a test binary")
	assert.Equal(t, SuperFlixEmbedHost, liveEmbedHost(),
		"and must not leave a discovered host behind for other tests to trip on")
}

// The manual pin costs no network, so it must keep working in tests — it is the
// escape hatch when discovery is unavailable.
func TestEnsureLiveHost_EnvPinWorksEvenInTests(t *testing.T) {
	resetHostDiscovery(t)
	t.Setenv("GOANIME_LIVE", "")
	t.Setenv(hostEnvOverride, "superflixapi.pinned")

	ensureLiveHost(t.Context())
	assert.Equal(t, "superflixapi.pinned", liveEmbedHost())
}
