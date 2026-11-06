package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sourceKey builds the key press a terminal would deliver for name.
func sourceKey(name string) tea.KeyPressMsg {
	switch name {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	}
	r := []rune(name)
	return tea.KeyPressMsg{Code: r[0], Text: name}
}

// fakeSources is the host the screen talks to: an in-memory catalogue that
// records every save, and can be told to fail one.
type fakeSources struct {
	rows   []SourceRow
	saved  []string // "id=on" / "id=off", in order
	setErr error
}

func (f *fakeSources) load() []SourceRow { return append([]SourceRow(nil), f.rows...) }

func (f *fakeSources) set(id string, on bool) error {
	if f.setErr != nil {
		return f.setErr
	}
	state := "off"
	if on {
		state = "on"
	}
	f.saved = append(f.saved, id+"="+state)
	for i := range f.rows {
		if f.rows[i].ID == id {
			f.rows[i].On = on
		}
	}
	return nil
}

func newFakeSources() *fakeSources {
	return &fakeSources{rows: []SourceRow{
		{ID: "AnimeFire", Name: "AnimeFire", On: true},
		{ID: "HiAnime", Name: "HiAnime", On: true},
		{ID: "Goyabu", Name: "Goyabu", On: false, Note: "off by default; needs cfsolve"},
		{ID: "TopCine", Name: "TopCine", On: false, Locked: true, Note: "locked by GOANIME_DISABLED_SOURCES"},
	}}
}

func newTestPanel(f *fakeSources) *sourcesPanel {
	theme := CurrentTheme()
	return newSourcesPanel(&theme, f.load, f.set)
}

func plainMode(t *testing.T, plain bool) {
	t.Helper()
	SetPlain(plain)
	t.Cleanup(func() { SetPlain(false) })
}

func TestSourcesSummary(t *testing.T) {
	plainMode(t, false)
	rows := []SourceRow{{Name: "A", On: true}, {Name: "B", On: true}, {Name: "Goyabu"}, {Name: "TopCine"}}

	assert.Equal(t, "2 sources on · Goyabu, TopCine off", SourcesSummary(rows))
	assert.Equal(t, "1 source on", SourcesSummary(rows[:1]), "one is singular, and nothing is off")
	assert.Equal(t, "0 sources on · Goyabu off", SourcesSummary(rows[2:3]))
	assert.Equal(t, "0 sources on", SourcesSummary(nil))

	plainMode(t, true)
	assert.Equal(t, "2 sources on - Goyabu, TopCine off", SourcesSummary(rows), "plain mode avoids the middle dot")
}

func TestNewSourcesPanelLoadsRows(t *testing.T) {
	p := newTestPanel(newFakeSources())
	assert.Len(t, p.rows, 4)
	assert.Equal(t, 0, p.cursor)
	assert.Equal(t, 2, p.onCount())
}

func TestSourcesPanelRefreshClampsTheCursor(t *testing.T) {
	f := newFakeSources()
	p := newTestPanel(f)
	p.cursor = 3
	f.rows = f.rows[:2]
	p.refresh()
	assert.Equal(t, 1, p.cursor, "the list shrank under the cursor")
	f.rows = nil
	p.refresh()
	assert.Equal(t, 0, p.cursor, "an empty list parks the cursor at 0")
}

func TestSourcesPanelMoveStopsAtTheEnds(t *testing.T) {
	p := newTestPanel(newFakeSources())
	p.move(-1)
	assert.Equal(t, 0, p.cursor)
	p.move(1)
	p.move(1)
	p.move(1)
	p.move(1)
	assert.Equal(t, 3, p.cursor, "stops on the last row, no wrap")

	empty := newTestPanel(&fakeSources{})
	empty.move(1)
	empty.move(-1)
	assert.Equal(t, 0, empty.cursor)
}

func TestSourcesPanelToggleSavesTheChoice(t *testing.T) {
	f := newFakeSources()
	p := newTestPanel(f)

	p.toggle() // AnimeFire on -> off
	assert.Equal(t, []string{"AnimeFire=off"}, f.saved)
	assert.False(t, p.rows[0].On, "the row follows what was saved")
	assert.Contains(t, p.status, "AnimeFire off")
	assert.False(t, p.statusBad)

	p.toggle() // back on
	assert.Equal(t, []string{"AnimeFire=off", "AnimeFire=on"}, f.saved)
	assert.True(t, p.rows[0].On)
	assert.Contains(t, p.status, "AnimeFire on")

	p.move(2) // Goyabu, off by default
	p.toggle()
	assert.Equal(t, "Goyabu=on", f.saved[len(f.saved)-1])
	assert.True(t, p.rows[2].On)
}

func TestSourcesPanelToggleOnALockedRowSavesNothing(t *testing.T) {
	f := newFakeSources()
	p := newTestPanel(f)
	p.move(3)
	p.toggle()
	assert.Empty(t, f.saved, "the environment holds this one")
	assert.True(t, p.statusBad)
	assert.Contains(t, p.status, "TopCine")
	assert.Contains(t, p.status, "GOANIME_DISABLED_SOURCES")
}

func TestSourcesPanelToggleSaveFailureIsShownAndKeepsTheRow(t *testing.T) {
	f := newFakeSources()
	f.setErr = errors.New("disk full\nsecond line")
	p := newTestPanel(f)

	p.toggle()
	assert.True(t, p.rows[0].On, "a choice that did not save is not shown as made")
	assert.True(t, p.statusBad)
	assert.Contains(t, p.status, "Could not save")
	assert.Contains(t, p.status, "disk full")
	assert.NotContains(t, p.status, "\n", "the status stays on one line")
}

func TestSourcesPanelToggleWithoutAHostOrRows(t *testing.T) {
	theme := CurrentTheme()
	f := newFakeSources()
	noSet := newSourcesPanel(&theme, f.load, nil)
	noSet.toggle() // must not panic
	assert.True(t, noSet.rows[0].On)
	assert.Empty(t, noSet.status)

	empty := newTestPanel(&fakeSources{})
	empty.toggle() // must not panic
	assert.Empty(t, empty.status)
}

func TestSourcesPanelKeys(t *testing.T) {
	t.Run("navigation", func(t *testing.T) {
		p := newTestPanel(newFakeSources())
		for _, step := range []struct {
			key  string
			want int
		}{{"down", 1}, {"j", 2}, {"up", 1}, {"k", 0}, {"end", 3}, {"home", 0}} {
			assert.False(t, p.update(sourceKey(step.key)), step.key)
			assert.Equal(t, step.want, p.cursor, step.key)
		}
		p.update(sourceKey("G"))
		assert.Equal(t, 3, p.cursor)
		p.update(sourceKey("g"))
		assert.Equal(t, 0, p.cursor)
	})

	t.Run("space enter and x flip the highlighted source", func(t *testing.T) {
		for _, key := range []string{"space", "enter", "x"} {
			f := newFakeSources()
			p := newTestPanel(f)
			assert.False(t, p.update(sourceKey(key)), key)
			assert.Equal(t, []string{"AnimeFire=off"}, f.saved, key)
		}
	})

	t.Run("a literal space also flips", func(t *testing.T) {
		f := newFakeSources()
		p := newTestPanel(f)
		p.update(tea.KeyPressMsg{Code: ' ', Text: " "})
		assert.Equal(t, []string{"AnimeFire=off"}, f.saved)
	})

	t.Run("done keys close it", func(t *testing.T) {
		for _, key := range []string{"esc", "ctrl+s", "q"} {
			f := newFakeSources()
			p := newTestPanel(f)
			assert.True(t, p.update(sourceKey(key)), key)
			assert.Empty(t, f.saved, "closing saves nothing extra: every flip was already saved")
		}
	})

	t.Run("other keys do nothing", func(t *testing.T) {
		f := newFakeSources()
		p := newTestPanel(f)
		for _, key := range []string{"z", "tab", "?"} {
			assert.False(t, p.update(sourceKey(key)), key)
		}
		assert.Empty(t, f.saved)
		assert.Equal(t, 0, p.cursor)
	})

	t.Run("an empty list survives every key", func(t *testing.T) {
		p := newTestPanel(&fakeSources{})
		for _, key := range []string{"down", "up", "end", "home", "space"} {
			assert.NotPanics(t, func() { p.update(sourceKey(key)) }, key)
		}
	})
}

func TestSourcesPanelHints(t *testing.T) {
	plainMode(t, false)
	p := newTestPanel(newFakeSources())
	keys := map[string]bool{}
	for _, h := range p.hints() {
		keys[h.Key] = h.Essential
	}
	assert.True(t, keys["space"], "how to flip is always shown")
	assert.True(t, keys["esc"], "how to leave is always shown")
	assert.Contains(t, keys, "↑↓")

	plainMode(t, true)
	found := false
	for _, h := range p.hints() {
		found = found || h.Key == "up/dn"
	}
	assert.True(t, found, "plain mode spells the arrows")
}

func TestSourcesPanelView(t *testing.T) {
	plainMode(t, false)
	f := newFakeSources()
	p := newTestPanel(f)
	out := ansi.Strip(p.view(80))

	assert.Contains(t, out, "SOURCES")
	assert.Contains(t, out, "saved as you make them", "tells the user there is no apply step")
	assert.Contains(t, out, "[x] AnimeFire")
	assert.Contains(t, out, "[ ] Goyabu")
	assert.Contains(t, out, "off by default; needs cfsolve")
	assert.Contains(t, out, "locked by GOANIME_DISABLED_SOURCES")
	assert.Contains(t, out, "2 of 4 on")
	assert.NotContains(t, out, "No source is on")

	t.Run("the pointer marks the highlighted row", func(t *testing.T) {
		lines := strings.Split(out, "\n")
		var pointed []string
		for _, l := range lines {
			if strings.Contains(l, "▶") || strings.HasPrefix(strings.TrimSpace(l), ">") {
				pointed = append(pointed, l)
			}
		}
		require.Len(t, pointed, 1)
		assert.Contains(t, pointed[0], "AnimeFire")
	})

	t.Run("plain mode uses an ASCII pointer", func(t *testing.T) {
		plainMode(t, true)
		assert.Contains(t, ansi.Strip(p.view(80)), "> [x] AnimeFire")
	})

	t.Run("warns when nothing is on", func(t *testing.T) {
		plainMode(t, false)
		for i := range f.rows {
			f.rows[i].On = false
		}
		p.refresh()
		assert.Contains(t, ansi.Strip(p.view(80)), "No source is on")
	})

	t.Run("shows the status line", func(t *testing.T) {
		p.status, p.statusBad = "Could not save: disk full", true
		assert.Contains(t, ansi.Strip(p.view(80)), "Could not save: disk full")
	})

	t.Run("an empty catalogue draws without the warning", func(t *testing.T) {
		out := ansi.Strip(newTestPanel(&fakeSources{}).view(80))
		assert.Contains(t, out, "0 of 0 on")
		assert.NotContains(t, out, "No source is on")
	})
}

// No line may run past the terminal, however narrow, or the layout tears.
func TestSourcesPanelViewFitsNarrowTerminals(t *testing.T) {
	plainMode(t, false)
	f := newFakeSources()
	f.rows[0].Name = "A very long source name\nwith a break"
	f.rows[0].Note = strings.Repeat("note ", 30)
	p := newTestPanel(f)
	p.status = strings.Repeat("status ", 30)
	for _, width := range []int{20, 40, 60, 100} {
		for _, line := range strings.Split(p.view(width), "\n") {
			assert.LessOrEqual(t, ansi.StringWidth(line), width, "width %d: %q", width, ansi.Strip(line))
		}
	}
	assert.NotContains(t, ansi.Strip(p.view(80)), "with a break\n", "a name with a newline stays on one line")
}

// row is the unit of the checklist: pointer, box, name, note.
func TestSourcesPanelRow(t *testing.T) {
	plainMode(t, false)
	p := newTestPanel(newFakeSources())
	on := SourceRow{Name: "HiAnime", On: true}
	off := SourceRow{Name: "Goyabu", Note: "off by default"}

	got := ansi.Strip(p.row(0, on, 7, 60))
	assert.Contains(t, got, "[x] HiAnime")
	assert.Contains(t, got, "▶", "row 0 is under the cursor")

	got = ansi.Strip(p.row(1, off, 7, 60))
	assert.Contains(t, got, "[ ] Goyabu")
	assert.Contains(t, got, "off by default")
	assert.NotContains(t, got, "▶")

	narrow := ansi.Strip(p.row(1, off, 7, 14))
	assert.LessOrEqual(t, ansi.StringWidth(narrow), 14)
	assert.NotContains(t, narrow, "off by default", "a note that does not fit is dropped, not wrapped")
}
