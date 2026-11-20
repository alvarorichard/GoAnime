package source

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// savedChoice runs with no env lists and an empty saved-choice table, and
// leaves both clean. Not parallel: it owns process-global state.
func savedChoice(t *testing.T) {
	t.Helper()
	t.Setenv(disabledSourcesEnvForTest, "")
	t.Setenv(enabledSourcesEnvForTest, "")
	util.ResetSourcePrefs()
	t.Cleanup(util.ResetSourcePrefs)
}

// IsSearchEnabled: env kill-switch > env opt-in > saved choice > default.
func TestIsSearchEnabledSavedChoice(t *testing.T) {
	plain := Descriptor{Kind: HiAnime}
	shipsOff := Descriptor{Kind: Goyabu, DefaultDisabled: true}

	t.Run("a saved off keeps a normal source out of the search", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("HiAnime", false)
		assert.False(t, IsSearchEnabled(plain))
		assert.True(t, IsEnabled(plain), "off in the screen is a SEARCH choice: history from it still resolves")
	})

	t.Run("a saved on brings a source that ships off", func(t *testing.T) {
		savedChoice(t)
		assert.False(t, IsSearchEnabled(shipsOff))
		util.SetSourcePref("Goyabu", true)
		assert.True(t, IsSearchEnabled(shipsOff))
	})

	t.Run("no saved choice means the default", func(t *testing.T) {
		savedChoice(t)
		assert.True(t, IsSearchEnabled(plain))
		assert.False(t, IsSearchEnabled(shipsOff))
	})

	t.Run("the env opt-in beats a saved off", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("HiAnime", false)
		t.Setenv(enabledSourcesEnvForTest, "HiAnime")
		assert.True(t, IsSearchEnabled(plain))
	})

	t.Run("the env kill-switch beats a saved on", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("Goyabu", true)
		t.Setenv(disabledSourcesEnvForTest, "Goyabu")
		assert.False(t, IsSearchEnabled(shipsOff))
	})
}

func TestStateOf(t *testing.T) {
	plain := Descriptor{Kind: HiAnime}
	caveat := Descriptor{Kind: Goyabu, DefaultDisabled: true, Note: "needs cfsolve"}

	t.Run("on by default has no note", func(t *testing.T) {
		savedChoice(t)
		st := stateOf(plain)
		assert.Equal(t, State{Kind: HiAnime, On: true}, st)
	})

	t.Run("on keeps the source's own caveat", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("Goyabu", true)
		st := stateOf(caveat)
		assert.True(t, st.On)
		assert.Equal(t, "needs cfsolve", st.Note)
	})

	t.Run("off by default says so, with the caveat", func(t *testing.T) {
		savedChoice(t)
		st := stateOf(caveat)
		assert.False(t, st.On)
		assert.False(t, st.Locked)
		assert.Equal(t, "off by default; needs cfsolve", st.Note)
	})

	t.Run("off by the user says so", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("HiAnime", false)
		st := stateOf(plain)
		assert.False(t, st.On)
		assert.Equal(t, "turned off by you", st.Note)
	})

	t.Run("locked by the environment cannot be changed", func(t *testing.T) {
		savedChoice(t)
		t.Setenv(disabledSourcesEnvForTest, "HiAnime")
		st := stateOf(plain)
		assert.False(t, st.On)
		assert.True(t, st.Locked)
		assert.Contains(t, st.Note, "GOANIME_DISABLED_SOURCES")
	})

	t.Run("an env opt-in over a saved off is explained", func(t *testing.T) {
		savedChoice(t)
		util.SetSourcePref("HiAnime", false)
		t.Setenv(enabledSourcesEnvForTest, "HiAnime")
		st := stateOf(plain)
		assert.True(t, st.On)
		assert.Contains(t, st.Note, "GOANIME_ENABLED_SOURCES")
	})
}

func TestJoinNote(t *testing.T) {
	assert.Equal(t, "off by default", joinNote("off by default", ""))
	assert.Equal(t, "off by default; needs cfsolve", joinNote("off by default", "needs cfsolve"))
}

func TestSourceStatesFollowsRegistryAndPriority(t *testing.T) {
	savedChoice(t)
	restore := SwapRegistryForTesting(
		newFake(Goyabu, 30, func(d *Descriptor) { d.DefaultDisabled = true }),
		newFake(HiAnime, 10),
		newFake(StartFlix, 20),
	)
	t.Cleanup(restore)
	util.SetSourcePref("StartFlix", false)

	states := SourceStates()
	require.Len(t, states, 3)
	assert.Equal(t, []SourceKind{HiAnime, StartFlix, Goyabu}, []SourceKind{states[0].Kind, states[1].Kind, states[2].Kind}, "resolution order")
	assert.True(t, states[0].On)
	assert.False(t, states[1].On, "saved off")
	assert.False(t, states[2].On, "ships off")

	// The screen must never disagree with the search.
	for _, st := range states {
		src, ok := Registered(st.Kind)
		require.True(t, ok)
		assert.Equal(t, IsSearchEnabled(src.Describe()), st.On, "state of %s", st.Kind)
	}
}

func TestSourceStatesEmptyRegistry(t *testing.T) {
	savedChoice(t)
	restore := SwapRegistryForTesting()
	t.Cleanup(restore)
	assert.Empty(t, SourceStates())
}

// The kill-switch removes a source from resolution, but the screen must still
// list it, locked: a source that simply vanished would leave the user guessing.
func TestSourceStatesKeepsLockedSourcesVisible(t *testing.T) {
	savedChoice(t)
	restore := SwapRegistryForTesting(newFake(HiAnime, 10), newFake(Goyabu, 20))
	t.Cleanup(restore)
	t.Setenv(disabledSourcesEnvForTest, "Goyabu")

	states := SourceStates()
	require.Len(t, states, 2)
	assert.Equal(t, Goyabu, states[1].Kind)
	assert.True(t, states[1].Locked)
	assert.False(t, states[1].On)
	assert.Len(t, registeredByPriority(), 1, "resolution still skips it")
}

func TestAllByPriorityOrdersAndKeepsEverything(t *testing.T) {
	savedChoice(t)
	restore := SwapRegistryForTesting(newFake(Goyabu, 30), newFake(HiAnime, 10), newFake(StartFlix, 10))
	t.Cleanup(restore)
	t.Setenv(disabledSourcesEnvForTest, "HiAnime")

	var kinds []SourceKind
	for _, s := range allByPriority() {
		kinds = append(kinds, s.Describe().Kind)
	}
	assert.Equal(t, []SourceKind{HiAnime, StartFlix, Goyabu}, kinds, "priority, then kind; the disabled one stays")
}
