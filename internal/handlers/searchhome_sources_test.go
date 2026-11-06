package handlers

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubSource is a registry entry with just a descriptor.
type stubSource struct{ desc source.Descriptor }

func (s *stubSource) Describe() source.Descriptor { return s.desc }
func (s *stubSource) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}
func (s *stubSource) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	return "", nil
}

func stub(kind source.SourceKind, priority int, mutate ...func(*source.Descriptor)) *stubSource {
	d := source.Descriptor{Kind: kind, Priority: priority, Explicit: []string{string(kind)}}
	for _, m := range mutate {
		m(&d)
	}
	return &stubSource{desc: d}
}

// sourcesWorld gives a test three sources (one ships off), no env lists, no
// saved choices, and the config folder redirected to a temp dir.
func sourcesWorld(t *testing.T) (prefsFile string) {
	t.Helper()
	t.Setenv("GOANIME_DISABLED_SOURCES", "")
	t.Setenv("GOANIME_ENABLED_SOURCES", "")
	util.ResetSourcePrefs()
	restore := source.SwapRegistryForTesting(
		stub(source.AnimeFire, 10),
		stub(source.HiAnime, 20),
		stub(source.Goyabu, 30, func(d *source.Descriptor) { d.DefaultDisabled = true; d.Note = "needs cfsolve" }),
	)
	prefsFile = filepath.Join(t.TempDir(), "goanime", "sources.json")
	oldPath := sourcePrefsPath
	sourcePrefsPath = func() (string, error) { return prefsFile, nil }
	t.Cleanup(func() {
		restore()
		sourcePrefsPath = oldPath
		util.ResetSourcePrefs()
	})
	return prefsFile
}

func TestHomeSourceRowsAdaptsStates(t *testing.T) {
	rows := homeSourceRows([]source.State{
		{Kind: source.AnimeFire, On: true},
		{Kind: source.Goyabu, Note: "off by default"},
		{Kind: source.TopCine, Locked: true, Note: "locked"},
	})
	assert.Equal(t, []tui.SourceRow{
		{ID: "AnimeFire", Name: "AnimeFire", On: true},
		{ID: "Goyabu", Name: "Goyabu", Note: "off by default"},
		{ID: "TopCine", Name: "TopCine", Note: "locked", Locked: true},
	}, rows)
	assert.Empty(t, homeSourceRows(nil))
}

func TestHomeSourcesReadsTheRegistry(t *testing.T) {
	sourcesWorld(t)
	rows := homeSources()
	require.Len(t, rows, 3)
	assert.Equal(t, "AnimeFire", rows[0].ID)
	assert.True(t, rows[0].On)
	assert.Equal(t, "Goyabu", rows[2].ID)
	assert.False(t, rows[2].On)
	assert.Contains(t, rows[2].Note, "off by default")
	assert.Contains(t, rows[2].Note, "needs cfsolve")
}

func TestSourcesLineCountsWhatSearches(t *testing.T) {
	sourcesWorld(t)
	assert.Equal(t, "2 sources on · Goyabu off", sourcesLine(),
		"a source that ships off is listed as off, not hidden")
	util.SetSourcePref("Goyabu", true)
	assert.Equal(t, "3 sources on", sourcesLine())
}

func TestSetHomeSource(t *testing.T) {
	t.Run("saves through the persist seam at the config path", func(t *testing.T) {
		file := sourcesWorld(t)
		oldPersist := persistSourcePref
		t.Cleanup(func() { persistSourcePref = oldPersist })
		var gotPath, gotName string
		var gotOn bool
		persistSourcePref = func(path, name string, on bool) error { gotPath, gotName, gotOn = path, name, on; return nil }

		require.NoError(t, setHomeSource("Goyabu", true))
		assert.Equal(t, file, gotPath)
		assert.Equal(t, "Goyabu", gotName)
		assert.True(t, gotOn)
	})

	t.Run("a missing config folder is reported", func(t *testing.T) {
		sourcesWorld(t)
		sourcePrefsPath = func() (string, error) { return "", errors.New("no home") }
		err := setHomeSource("Goyabu", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no home")
	})

	t.Run("a failed save is reported", func(t *testing.T) {
		sourcesWorld(t)
		oldPersist := persistSourcePref
		t.Cleanup(func() { persistSourcePref = oldPersist })
		persistSourcePref = func(string, string, bool) error { return errors.New("disk full") }
		require.EqualError(t, setHomeSource("Goyabu", true), "disk full")
	})
}

func TestOpenSearchHomeHandsTheScreenItsSources(t *testing.T) {
	sourcesWorld(t)
	var got tui.SearchHomeOptions
	_, err := openSearchHome("naruto", func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		got = o
		return tui.SearchHomeResult{}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, got.Sources, "without these the ctrl+s screen would not exist")
	require.NotNil(t, got.SetSource)
	assert.Len(t, got.Sources(), 3)
	assert.Equal(t, "2 sources on · Goyabu off", got.SourcesLine)
}
