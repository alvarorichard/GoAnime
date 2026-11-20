package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func homeWithSources(f *fakeSources) *searchHomeModel {
	return newSearchHomeModel(SearchHomeOptions{
		SourcesLine: SourcesSummary(f.load()),
		Sources:     f.load,
		SetSource:   f.set,
	})
}

func homeKey(m *searchHomeModel, name string) (*searchHomeModel, tea.Cmd) {
	model, cmd := m.Update(sourceKey(name))
	return model.(*searchHomeModel), cmd
}

func hintKeys(m *searchHomeModel) map[string]bool {
	out := map[string]bool{}
	for _, h := range m.hints() {
		out[h.Key] = true
	}
	return out
}

func TestSearchHomeCtrlSOpensAndEscCloses(t *testing.T) {
	plainMode(t, false)
	f := newFakeSources()
	m := homeWithSources(f)
	require.Nil(t, m.sources)

	m, cmd := homeKey(m, "ctrl+s")
	require.NotNil(t, m.sources, "ctrl+s opens the screen")
	assert.Nil(t, cmd)
	assert.Contains(t, ansi.Strip(m.View().Content), "SOURCES")

	m, _ = homeKey(m, "esc")
	assert.Nil(t, m.sources, "esc is done")
	assert.NotContains(t, ansi.Strip(m.View().Content), "SOURCES")

	m, _ = homeKey(m, "ctrl+s")
	m, _ = homeKey(m, "ctrl+s")
	assert.Nil(t, m.sources, "ctrl+s again also closes")
}

func TestSearchHomeSourcesAbsentWithoutAHost(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	m, cmd := homeKey(m, "ctrl+s")
	assert.Nil(t, m.sources, "no host, no screen")
	assert.Nil(t, cmd)
	assert.False(t, hintKeys(m)["ctrl+s"], "the footer does not advertise a key that does nothing")
}

func TestSearchHomeSourcesAreAdvertised(t *testing.T) {
	plainMode(t, false)
	m := homeWithSources(newFakeSources())
	assert.True(t, hintKeys(m)["ctrl+s"], "the footer says how to reach the screen")
	var essential bool
	for _, h := range m.hints() {
		if h.Key == "ctrl+s" {
			essential = h.Essential
		}
	}
	assert.True(t, essential, "the hint survives a narrow footer")

	for _, h := range m.hints() {
		if h.Key == "tab" {
			assert.Equal(t, "scope", h.Desc, "tab changes the scope; the word 'sources' now belongs to ctrl+s")
		}
	}

	help := ansi.Strip(m.helpView(100, 30))
	assert.Contains(t, help, "ctrl+s")
	assert.Contains(t, help, "choose sources")
}

func TestSearchHomeToggleUpdatesTheHeaderAtOnce(t *testing.T) {
	plainMode(t, false)
	f := newFakeSources()
	m := homeWithSources(f)
	assert.Equal(t, "2 sources on · Goyabu, TopCine off", m.shell.Context)

	m, _ = homeKey(m, "ctrl+s")
	m, _ = homeKey(m, "space") // AnimeFire off
	assert.Equal(t, "1 source on · AnimeFire, Goyabu, TopCine off", m.shell.Context)
	assert.Equal(t, []string{"AnimeFire=off"}, f.saved)

	m, _ = homeKey(m, "esc")
	assert.Equal(t, "1 source on · AnimeFire, Goyabu, TopCine off", m.shell.Context, "the header keeps the new state after closing")
}

func TestSearchHomeSourcesScreenOwnsTheKeys(t *testing.T) {
	f := newFakeSources()
	m := homeWithSources(f)
	m.input.SetValue("naruto")
	m, _ = homeKey(m, "ctrl+s")

	m, _ = homeKey(m, "x") // flips; must not type into the box
	m, _ = homeKey(m, "j")
	m, _ = homeKey(m, "tab") // must not change the scope
	m, cmd := homeKey(m, "enter")
	assert.Nil(t, cmd, "enter flips a source here, it does not search")
	assert.Equal(t, "naruto", m.input.Value(), "typing never leaks into the search box")
	assert.Equal(t, 0, m.scopeIdx)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Empty(t, m.result.Query, "nothing was searched")
}

func TestSearchHomeSourcesCtrlCStillQuits(t *testing.T) {
	m := homeWithSources(newFakeSources())
	m, _ = homeKey(m, "ctrl+s")
	m, cmd := homeKey(m, "ctrl+c")
	require.NotNil(t, cmd, "ctrl+c quits from anywhere")
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}

func TestSearchHomeOpeningSourcesDropsOpenSuggestions(t *testing.T) {
	m := homeWithSources(newFakeSources())
	m.suggestions = []HomeSuggestion{{Title: "Naruto"}}
	m, _ = homeKey(m, "ctrl+s")
	assert.Empty(t, m.suggestions, "a dropdown must not reappear under the screen")
}

func TestSearchHomeSourcesHintsReplaceTheHomeHints(t *testing.T) {
	m := homeWithSources(newFakeSources())
	m, _ = homeKey(m, "ctrl+s")
	keys := hintKeys(m)
	assert.True(t, keys["space"])
	assert.True(t, keys["esc"])
	assert.False(t, keys["enter"], "the home's enter-to-search hint would mislead here")
}

func TestSearchHomeSourcesSurviveResize(t *testing.T) {
	m := homeWithSources(newFakeSources())
	m, _ = homeKey(m, "ctrl+s")
	model, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = model.(*searchHomeModel)
	for _, line := range splitLines(m.View().Content) {
		assert.LessOrEqual(t, ansi.StringWidth(line), 40, "%q", ansi.Strip(line))
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := range s {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// openSources and updateSources, called directly.
func TestOpenSourcesBuildsThePanelFromTheHost(t *testing.T) {
	f := newFakeSources()
	m := homeWithSources(f)
	m.openSources()
	require.NotNil(t, m.sources)
	assert.Len(t, m.sources.rows, 4, "the panel lists what the host's load returned")

	none := newSearchHomeModel(SearchHomeOptions{})
	none.openSources()
	assert.Nil(t, none.sources, "no host, nothing to open")
}

func TestUpdateSourcesRoutesKeysToThePanel(t *testing.T) {
	plainMode(t, false)
	f := newFakeSources()
	m := homeWithSources(f)
	m.openSources()

	model, cmd := m.updateSources(sourceKey("space"))
	assert.Same(t, m, model)
	assert.Nil(t, cmd)
	assert.Equal(t, []string{"AnimeFire=off"}, f.saved)
	assert.Equal(t, "1 source on · AnimeFire, Goyabu, TopCine off", m.shell.Context)
	require.NotNil(t, m.sources, "a flip keeps the screen open")

	_, _ = m.updateSources(sourceKey("esc"))
	assert.Nil(t, m.sources, "done closes it")

	m.openSources()
	_, cmd = m.updateSources(sourceKey("ctrl+c"))
	require.NotNil(t, cmd)
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}
