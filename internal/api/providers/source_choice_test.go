package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanChoices(t *testing.T) {
	t.Helper()
	t.Setenv("GOANIME_DISABLED_SOURCES", "")
	t.Setenv("GOANIME_ENABLED_SOURCES", "")
	util.ResetSourcePrefs()
	t.Cleanup(util.ResetSourcePrefs)
}

func TestGoyabuNote(t *testing.T) {
	assert.Empty(t, goyabuNote(true), "with the solver there is nothing to warn about")
	got := goyabuNote(false)
	assert.Contains(t, got, "cfsolve")
	assert.Contains(t, got, "docs/CFSOLVE.md", "the note says where to read more")
}

func TestGoyabuDescriptorCarriesTheNote(t *testing.T) {
	cleanChoices(t)
	withSolver(t, false)
	assert.Contains(t, goyabuDescriptor(t).Note, "cfsolve")
	withSolver(t, true)
	assert.Empty(t, goyabuDescriptor(t).Note)
}

func TestSearchOffReason(t *testing.T) {
	cleanChoices(t)
	d := source.Descriptor{Kind: source.Goyabu, DefaultDisabled: true}
	assert.Equal(t, "off by default", searchOffReason(d))
	util.SetSourcePref("Goyabu", false)
	assert.Equal(t, "turned off", searchOffReason(d), "the user's own choice is named as such")
}

// A source the user switched off in the SOURCES screen sits out the search and
// the SEARCHING panel says it was their choice; one they switched on joins
// even though it ships off.
func TestPlanLiveSearchHonoursSavedChoices(t *testing.T) {
	cleanChoices(t)
	fire := newSearchStub(source.AnimeFire, []*models.Anime{{Name: "X"}}, nil)
	hi := newSearchStub(source.HiAnime, []*models.Anime{{Name: "Y"}}, nil)
	goyabu := newSearchStub(source.Goyabu, nil, nil)
	goyabu.desc.DefaultDisabled = true
	restore := source.SwapRegistryForTesting(fire, hi, goyabu)
	t.Cleanup(restore)

	util.SetSourcePref("HiAnime", false)
	util.SetSourcePref("Goyabu", true)

	searchers, off, searched := planLiveSearch(nil)
	assert.Len(t, searchers, 2)
	assert.ElementsMatch(t, []string{"AnimeFire", "Goyabu"}, searched)
	require.Len(t, off, 1)
	assert.Equal(t, models.SearchOffEntry{Name: "HiAnime", Reason: "turned off"}, off[0])
}
