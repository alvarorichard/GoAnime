package source

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Env var names honored by util.SourceDisabled / util.SourceForceEnabled.
const (
	disabledSourcesEnvForTest = "GOANIME_DISABLED_SOURCES"
	enabledSourcesEnvForTest  = "GOANIME_ENABLED_SOURCES"
)

func TestIsEnabled(t *testing.T) {
	// Uses t.Setenv — not parallel.
	t.Run("plain source enabled by default", func(t *testing.T) {
		t.Setenv(disabledSourcesEnvForTest, "")
		assert.True(t, IsEnabled(Descriptor{Kind: AniDB}))
	})

	t.Run("disabled via config", func(t *testing.T) {
		t.Setenv(disabledSourcesEnvForTest, "AniDB")
		assert.False(t, IsEnabled(Descriptor{Kind: AniDB}))
		assert.True(t, IsEnabled(Descriptor{Kind: Goyabu}), "only the listed source is off")
	})

	t.Run("DefaultDisabled is off unless opted in", func(t *testing.T) {
		d := Descriptor{Kind: "Experimental", DefaultDisabled: true}
		t.Setenv(enabledSourcesEnvForTest, "")
		assert.False(t, IsEnabled(d), "DefaultDisabled source is off by default")

		t.Setenv(enabledSourcesEnvForTest, "Experimental")
		assert.True(t, IsEnabled(d), "opting in via GOANIME_ENABLED_SOURCES turns it on")
	})

	t.Run("explicit disable overrides opt-in", func(t *testing.T) {
		d := Descriptor{Kind: "Experimental", DefaultDisabled: true}
		t.Setenv(disabledSourcesEnvForTest, "Experimental")
		t.Setenv(enabledSourcesEnvForTest, "Experimental")
		assert.False(t, IsEnabled(d), "the kill-switch wins over opt-in")
	})
}

func TestEnabled(t *testing.T) {
	// Swaps registry + env — not parallel.
	restore := SwapRegistryForTesting(newFake(AniDB, 1), newFake(Goyabu, 2))
	t.Cleanup(restore)

	t.Run("enabled source is returned", func(t *testing.T) {
		t.Setenv(disabledSourcesEnvForTest, "")
		s, ok := Enabled(AniDB)
		require.True(t, ok)
		assert.Equal(t, AniDB, s.Describe().Kind)
	})

	t.Run("disabled source is not returned", func(t *testing.T) {
		t.Setenv(disabledSourcesEnvForTest, "AniDB")
		_, ok := Enabled(AniDB)
		assert.False(t, ok, "a disabled source must not be selectable via Enabled")
		_, ok = Enabled(Goyabu)
		assert.True(t, ok, "other sources stay enabled")
	})

	t.Run("unregistered kind is not enabled", func(t *testing.T) {
		_, ok := Enabled("nope")
		assert.False(t, ok)
	})
}

func TestResolve_SkipsDisabledSource(t *testing.T) {
	// Swaps registry + env — not parallel.
	animeFire := newFake(AnimeFire, 10, func(d *Descriptor) { d.URLMatchers = []string{"animefire"} })
	goyabu := newFake(Goyabu, 20, func(d *Descriptor) { d.URLMatchers = []string{"goyabu"} })
	restore := SwapRegistryForTesting(animeFire, goyabu)
	t.Cleanup(restore)

	// With AnimeFire disabled, an animefire URL must not resolve to it.
	t.Setenv(disabledSourcesEnvForTest, "AnimeFire")

	src, resolved := Resolve(&models.Anime{URL: "https://animefire.plus/x"})
	assert.Nil(t, src, "a disabled source must not be returned by Resolve")
	assert.Equal(t, Unknown, resolved.Kind, "resolution falls through when the only match is disabled")

	// Goyabu still resolves.
	src, resolved = Resolve(&models.Anime{URL: "https://goyabu.to/x"})
	require.NotNil(t, src)
	assert.Equal(t, Goyabu, resolved.Kind)
}

func TestDisabledSources(t *testing.T) {
	// Swaps registry + env — not parallel.
	restore := SwapRegistryForTesting(newFake(AniDB, 1), newFake(Goyabu, 2), newFake(SuperFlix, 3))
	t.Cleanup(restore)

	t.Setenv(disabledSourcesEnvForTest, "Goyabu,SuperFlix")
	got := DisabledSources()
	assert.Equal(t, []SourceKind{Goyabu, SuperFlix}, got, "must list disabled registered sources, sorted")

	t.Setenv(disabledSourcesEnvForTest, "")
	assert.Empty(t, DisabledSources(), "nothing disabled ⇒ empty")
}
