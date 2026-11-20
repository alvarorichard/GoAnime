package tui

// Smaller rules of the home screen's helpers that the screen goldens do not
// show by themselves, found by the mutation pass over search_home.go (see
// regression_home_test.go for the screens).

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegressionHomeRecentCountWording(t *testing.T) {
	regPin(t, false)
	m := regHome{}.build()
	m.recentAll = []HomeRecentSearch{{Query: "a", Count: 0}, {Query: "b", Count: 1}, {Query: "c", Count: 2}, {Query: "d", Count: 1000}}
	m.rebuildRows()
	lines := strings.Split(ansi.Strip(m.recentView(40)), "\n")
	require.Len(t, lines, 4)
	for i, want := range []string{"0 results", "1 result", "2 results", "1000 results"} {
		assert.True(t, strings.HasSuffix(strings.TrimRight(lines[i], " "), want), "%q ends with %q", lines[i], want)
	}
}

func TestRegressionHomeRowWithNoTitleIsUntitled(t *testing.T) {
	regPin(t, false)
	theme := CurrentTheme()
	line := ansi.Strip(renderHomeRow(&theme, false, homeRowItem{title: " \t "}, 40))
	assert.Equal(t, "     Untitled", strings.TrimRight(line, " "))
	line = ansi.Strip(renderHomeRow(&theme, true, homeRowItem{title: "Dark", note: "2017 · TV", key: "Pobreflix"}, 40))
	assert.Equal(t, " ▶   Dark          2017 · TV   Pobreflix", strings.TrimRight(line, " "))
	line = ansi.Strip(renderHomeRow(&theme, false, homeRowItem{title: "A very long title that cannot possibly fit", note: "note", key: "key"}, 30))
	assert.Equal(t, "     A very long t… note   key", strings.TrimRight(line, " "))
}

func TestRegressionHomeFailureNoticeIsOneLine(t *testing.T) {
	regPin(t, false)
	calls := &regCalls{err: errors.New("first line\n  second\tline\x1b[31m")}
	m := regHome{continues: true, calls: calls}.build()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	regPress(m, "down", "x")
	assert.Equal(t, " !  Could not remove the episode: first line second line", strings.TrimRight(strings.Split(ansi.Strip(m.questionView(100)), "\n")[0], " "))
}

func TestRegressionHomeEnterFromTheBoxOpensTheFirstMatchWhateverTheCursor(t *testing.T) {
	regPin(t, false)
	m := regHome{favs: 5, scope: 4}.build()
	m.cursor = 3 // frozen: unreachable by keys, but the box must win over a stale cursor
	m.focusInput = true
	res, ok := m.confirm()
	assert.True(t, ok)
	assert.Equal(t, "Frieren Beyond Journey s End", res.Query)
	m.focusInput = false
	res, ok = m.confirm()
	assert.True(t, ok)
	assert.Equal(t, "Dark", res.Query, "with a row highlighted, that row")
	m.cursor = 9
	_, ok = m.confirm()
	assert.False(t, ok)
	m.cursor = -1
	_, ok = m.confirm()
	assert.False(t, ok)
}

func TestRegressionHomeTypeKeys(t *testing.T) {
	for _, tt := range []struct {
		name string
		key  tea.KeyPressMsg
		want bool
	}{
		{"letter", tea.KeyPressMsg{Code: 'a', Text: "a"}, true},
		{"space", tea.KeyPressMsg{Code: ' ', Text: " "}, true},
		{"accent", tea.KeyPressMsg{Code: 'é', Text: "é"}, true},
		{"no text", tea.KeyPressMsg{Code: tea.KeyDown}, false},
		{"escape char", tea.KeyPressMsg{Code: 'x', Text: "\x1b"}, false},
		{"delete char", tea.KeyPressMsg{Code: 'x', Text: "\x7f"}, false},
		{"control char", tea.KeyPressMsg{Code: 'x', Text: "\u0085"}, false},
		{"tab char", tea.KeyPressMsg{Code: 'x', Text: "\t"}, false},
		{"first printable", tea.KeyPressMsg{Code: '!', Text: "!"}, true},
		{"just below space", tea.KeyPressMsg{Code: 'x', Text: "\x1f"}, false},
	} {
		assert.Equal(t, tt.want, isHomeTypeKey(tt.key), tt.name)
	}
	regPin(t, false)
	m := regHome{continues: true, recents: true, favs: 3}.build()
	regPress(m, "down", "space")
	assert.True(t, m.focusInput, "typing over the lists jumps back to the box with the character")
	assert.Equal(t, " ", m.input.Value())
	m = regHome{continues: true}.build()
	regPress(m, "down", "w")
	assert.True(t, m.focusInput)
	assert.Equal(t, "w", m.input.Value())
}

func TestRegressionHomeRunnerTrimsBeforeCallingItACancel(t *testing.T) {
	_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
		m := model.(*searchHomeModel)
		m.result = SearchHomeResult{Action: HomeSearch, Query: "   "}
		return m, nil
	})
	assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	res, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
		m := model.(*searchHomeModel)
		m.result = SearchHomeResult{Action: HomeResume, HasEntry: true, Entry: HomeContinueEntry{Key: "k"}}
		return m, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "k", res.Entry.Key, "a resume has no query and is not a cancel")
	_, err = searchHomeWithRunner(SearchHomeOptions{}, func(tea.Model) (tea.Model, error) { return nil, nil })
	assert.EqualError(t, err, "unexpected search home model <nil>")
	if !HasTerminal() {
		_, err = SearchHome(SearchHomeOptions{})
		assert.EqualError(t, err, "no terminal: stdin is not a TTY")
	}
}

func TestRegressionHomeFilterRules(t *testing.T) {
	regPin(t, false)
	rows := func(query string) []string {
		m := regHome{favs: 13, scope: 4, query: query}.build()
		var keys []string
		for _, r := range m.rows {
			keys = append(keys, m.favoriteAll[r.index].Key)
		}
		return keys
	}
	assert.Empty(t, rows("animefire.io"), "frozen: the source is matched as it reads on screen (AnimeFire), not as stored")
	assert.Equal(t, []string{"frieren", "dandadan"}, rows("ANIMEFIRE"))
	assert.Equal(t, []string{"pokemon aventuras"}, rows("POKEMON"))
	assert.Equal(t, []string{"pokemon aventuras"}, rows("Pokémon"))
	assert.Equal(t, []string{"pokemon aventuras"}, rows("Pokémon"), "a typed decomposed accent folds too")
	assert.Equal(t, []string{"heat"}, rows("1995")[:1])
	assert.Equal(t, []string{"heat", "spirited away", "the matrix", "akira"}, rows("movie"), "in the order of the list")
	assert.Equal(t, []string{"heat"}, rows("heat startflix 1995 movie"), "every word must match")
	assert.Empty(t, rows("heat startflix 1996"))
	assert.Equal(t, "Pokémon Aventuras", searchableTitle("Pokémon: Aventuras"), "combining marks stay with their letter")
	assert.Equal(t, "pokemon", foldForMatch("POKÉMON"))
	assert.Equal(t, "abc", foldForMatch("ÀBC"))
}
