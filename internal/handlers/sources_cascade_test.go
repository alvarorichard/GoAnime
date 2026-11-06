package handlers

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// searching returns the kinds that join the search fan-out right now.
func searching() []source.SourceKind {
	var kinds []source.SourceKind
	for _, s := range source.ActiveSources() {
		kinds = append(kinds, s.Describe().Kind)
	}
	return kinds
}

// rowFor finds a source's row and fails if the screen does not list it: a
// check inside a loop over the rows would pass silently when the row is gone.
func rowFor(t *testing.T, rows []tui.SourceRow, id string) tui.SourceRow {
	t.Helper()
	for _, r := range rows {
		if r.ID == id {
			return r
		}
	}
	require.FailNow(t, "the screen does not list this source", id)
	return tui.SourceRow{}
}

// restart is what quitting and launching GoAnime does to the saved choices:
// memory is wiped and the file is read again.
func restart(t *testing.T, file string) {
	t.Helper()
	util.ResetSourcePrefs()
	require.NoError(t, util.LoadSourcePrefs(file))
}

// TestSourcesChoiceCascade walks the user's path end to end against stubs, one
// stage per hop, and stops at the first hop that breaks so the failure names
// it: a choice made in the SOURCES screen is saved, survives a restart, and
// changes who is searched.
func TestSourcesChoiceCascade(t *testing.T) {
	file := sourcesWorld(t)
	var opts tui.SearchHomeOptions
	_, err := openSearchHome("", func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		opts = o
		return tui.SearchHomeResult{}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, opts.SetSource)

	stages := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"1 first run: Goyabu ships off and nothing is saved", func(t *testing.T) {
			assert.Equal(t, []source.SourceKind{source.AnimeFire, source.HiAnime}, searching())
			_, statErr := os.Stat(file)
			assert.True(t, os.IsNotExist(statErr), "no file until the user chooses")
		}},
		{"2 the user switches Goyabu on in the screen", func(t *testing.T) {
			require.NoError(t, opts.SetSource("Goyabu", true))
			assert.Equal(t, []source.SourceKind{source.AnimeFire, source.HiAnime, source.Goyabu}, searching(), "takes effect without a restart")
			assert.True(t, rowFor(t, opts.Sources(), "Goyabu").On, "the screen shows what was chosen")
		}},
		{"3 the choice is on disk", func(t *testing.T) {
			raw, readErr := os.ReadFile(file)
			require.NoError(t, readErr)
			assert.Contains(t, string(raw), `"goyabu":true`)
		}},
		{"4 after a restart Goyabu still searches", func(t *testing.T) {
			restart(t, file)
			assert.Contains(t, searching(), source.Goyabu)
		}},
		{"5 the user switches HiAnime off", func(t *testing.T) {
			require.NoError(t, opts.SetSource("HiAnime", false))
			assert.Equal(t, []source.SourceKind{source.AnimeFire, source.Goyabu}, searching())
			assert.Equal(t, "2 sources on · HiAnime off", sourcesLine())
		}},
		{"6 history from HiAnime still resolves", func(t *testing.T) {
			// Off in the screen is a search choice: a saved episode must still open.
			src, resolved := source.Resolve(&models.Anime{Source: "HiAnime"})
			require.NotNil(t, src)
			assert.Equal(t, source.HiAnime, resolved.Kind)
		}},
		{"7 after another restart both choices hold", func(t *testing.T) {
			restart(t, file)
			assert.Equal(t, []source.SourceKind{source.AnimeFire, source.Goyabu}, searching())
		}},
		{"8 the environment outranks the saved choices", func(t *testing.T) {
			t.Setenv("GOANIME_DISABLED_SOURCES", "Goyabu")
			assert.Equal(t, []source.SourceKind{source.AnimeFire}, searching())
			assert.True(t, rowFor(t, opts.Sources(), "Goyabu").Locked, "the screen says it is not the user's to flip")
		}},
	}
	for _, st := range stages {
		if !t.Run(st.name, st.run) {
			t.Fatalf("cascade stopped: stage %q failed", st.name)
		}
	}
}
