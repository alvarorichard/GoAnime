package tui

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingSuggest is a SuggestFunc that records each call and answers from
// a script.
type recordingSuggest struct {
	mu    sync.Mutex
	calls []suggestCall
	items []HomeSuggestion
	err   error
	ctxs  []context.Context
	// errAtCall is each call's ctx.Err() at the moment fn was entered.
	errAtCall []error
}

type suggestCall struct{ query, scope string }

func (r *recordingSuggest) fn(ctx context.Context, query, scope string) ([]HomeSuggestion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, suggestCall{query, scope})
	r.ctxs = append(r.ctxs, ctx)
	r.errAtCall = append(r.errAtCall, ctx.Err())
	return r.items, r.err
}

func (r *recordingSuggest) called() []suggestCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]suggestCall(nil), r.calls...)
}

func sugRows(titles ...string) []HomeSuggestion {
	rows := make([]HomeSuggestion, 0, len(titles))
	for _, title := range titles {
		rows = append(rows, HomeSuggestion{Title: title, Detail: "Anime · 2002"})
	}
	return rows
}

// suggestModel builds a home model whose debounce is instant.
func suggestModel(t *testing.T, rec *recordingSuggest, opts SearchHomeOptions) *searchHomeModel { //nolint:gocritic // SearchHomeOptions is a configuration bag handed over once when a screen opens
	t.Helper()
	opts.Suggest = rec.fn
	m := newSearchHomeModel(opts)
	m.debounce = 0
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	return m
}

// msgsOf runs cmd and flattens batches, keeping messages that arrive within
// a short window (a cursor blink timer is dropped).
func msgsOf(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if batch, ok := msg.(tea.BatchMsg); ok {
			var out []tea.Msg
			for _, c := range batch {
				out = append(out, msgsOf(c)...)
			}
			return out
		}
		if msg == nil {
			return nil
		}
		return []tea.Msg{msg}
	case <-time.After(150 * time.Millisecond):
		return nil
	}
}

// settle feeds the model every suggestDueMsg/suggestResultMsg chain that cmd
// leads to, as the Bubble Tea loop would.
func settle(m *searchHomeModel, cmd tea.Cmd) {
	for _, msg := range msgsOf(cmd) {
		switch msg.(type) {
		case suggestDueMsg, suggestResultMsg:
			_, next := m.Update(msg)
			settle(m, next)
		}
	}
}

// typeText types text key by key and lets only the last keystroke's debounce
// fire, as a real pause after typing would: earlier ticks are stale.
func typeText(m *searchHomeModel, text string) {
	var last tea.Cmd
	for _, r := range text {
		_, last = m.Update(press(r, string(r)))
	}
	settle(m, last)
}

func titlesOf(rows []HomeSuggestion) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Title)
	}
	return out
}

// TestSearchHomeSuggestCascade walks the user's path hop by hop: type, wait
// out the debounce, see the rows, pick one, search it.
func TestSearchHomeSuggestCascade(t *testing.T) {
	rec := &recordingSuggest{items: sugRows("Naruto", "Naruto: Shippuuden", "Boruto: Naruto Next Generations")}
	m := suggestModel(t, rec, SearchHomeOptions{})

	if !t.Run("1 typing schedules a lookup", func(t *testing.T) {
		_, cmd := m.Update(press('n', "n"))
		settle(m, cmd)
		assert.Empty(t, rec.called(), "one letter is below the minimum")
		_, cmd = m.Update(press('a', "a"))
		due := msgsOf(cmd)
		require.Len(t, due, 1)
		assert.IsType(t, suggestDueMsg{}, due[0])
		assert.Empty(t, rec.called(), "the lookup waits for the debounce message")
	}) {
		return
	}
	if !t.Run("2 the lookup answers into the dropdown", func(t *testing.T) {
		m = suggestModel(t, rec, SearchHomeOptions{})
		typeText(m, "nar")
		require.Equal(t, []suggestCall{{"nar", ""}}, rec.called(), "only the settled query is asked")
		assert.Equal(t, []string{"Naruto", "Naruto: Shippuuden", "Boruto: Naruto Next Generations"}, titlesOf(m.suggestions))
		assert.True(t, m.suggestionsOpen())
		assert.Equal(t, -1, m.sugCursor)
	}) {
		return
	}
	if !t.Run("3 the view shows the rows with their detail", func(t *testing.T) {
		view := ansi.Strip(m.View().Content)
		assert.Contains(t, view, "Naruto: Shippuuden")
		assert.Contains(t, view, "Anime · 2002")
		assert.Contains(t, view, "complete", "the footer teaches the right-arrow completion")
	}) {
		return
	}
	if !t.Run("4 down highlights, enter searches the pick", func(t *testing.T) {
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		require.Equal(t, 1, m.sugCursor)
		_, cmd := m.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		assert.Equal(t, SearchHomeResult{Action: HomeSearch, Query: "Naruto: Shippuuden", Scope: ""}, m.result)
	}) {
		return
	}
}

func TestSearchHomeSuggestDisabled(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(SearchHomeOptions{})
	_, cmd := m.Update(press('n', "n"))
	_, cmd2 := m.Update(press('a', "a"))
	_, cmd3 := m.Update(press('r', "r"))
	for _, c := range []tea.Cmd{cmd, cmd2, cmd3} {
		for _, msg := range msgsOf(c) {
			assert.IsNotType(t, suggestDueMsg{}, msg, "no Suggest, no lookups")
		}
	}
	assert.False(t, m.suggestionsOpen())
	assert.Nil(t, m.startLookup(m.seq), "a stray tick with no engine does nothing")
}

func TestSearchHomeSuggestStaleAnswersAreDropped(t *testing.T) {
	t.Parallel()
	rec := &recordingSuggest{items: sugRows("Naruto")}
	m := suggestModel(t, rec, SearchHomeOptions{})

	typeText(m, "nar")
	require.Len(t, m.suggestions, 1)

	t.Run("an old tick does not start a lookup", func(t *testing.T) {
		before := len(rec.called())
		_, cmd := m.Update(suggestDueMsg{seq: m.seq - 1})
		assert.Nil(t, cmd)
		assert.Len(t, rec.called(), before)
	})
	t.Run("an old answer does not replace newer rows", func(t *testing.T) {
		_, _ = m.Update(suggestResultMsg{seq: m.seq - 1, items: sugRows("Stale")})
		assert.Equal(t, []string{"Naruto"}, titlesOf(m.suggestions))
	})
	t.Run("typing on clears the rows and cancels the lookup", func(t *testing.T) {
		_, cmd := m.Update(press('u', "u"))
		assert.Empty(t, m.suggestions, "rows for the old text are never shown under new text")
		_, due := m.Update(suggestDueMsg{seq: m.seq})
		require.NotNil(t, due)
		_ = cmd
	})
	t.Run("a lookup in flight is cancelled by new typing", func(t *testing.T) {
		rec := &recordingSuggest{items: sugRows("X")}
		m := suggestModel(t, rec, SearchHomeOptions{})
		_, _ = m.Update(press('a', "a"))
		_, _ = m.Update(press('b', "b"))
		run := m.startLookup(m.seq)
		require.NotNil(t, run)
		require.NotNil(t, m.cancel)
		_, _ = m.Update(press('c', "c"))
		assert.Nil(t, m.cancel, "the in-flight lookup was cancelled")
		msg := run().(suggestResultMsg)
		require.NoError(t, msg.err)
		_, _ = m.Update(msg)
		assert.Empty(t, m.suggestions, "its late answer is dropped by seq")
		require.NotEmpty(t, rec.errAtCall)
		assert.Error(t, rec.errAtCall[0], "the lookup saw a cancelled ctx: nobody waits on it any more")
		assert.Error(t, rec.ctxs[0].Err(), "ctx is released when the lookup returns")
	})
}

func TestSearchHomeSuggestFailureShowsNothing(t *testing.T) {
	t.Parallel()
	rec := &recordingSuggest{err: errors.New("offline"), items: sugRows("Naruto")}
	m := suggestModel(t, rec, SearchHomeOptions{})
	typeText(m, "nar")
	assert.Equal(t, []suggestCall{{"nar", ""}}, rec.called())
	assert.Empty(t, m.suggestions)
	assert.NotContains(t, ansi.Strip(m.View().Content), "offline", "the box works without autocomplete, silently")

	_, cmd := m.Update(press(tea.KeyEnter, ""))
	require.NotNil(t, cmd, "enter still searches what was typed")
	assert.Equal(t, "nar", m.result.Query)
}

func TestSearchHomeSuggestRowCleaning(t *testing.T) {
	t.Parallel()
	rows := []HomeSuggestion{
		{Title: "naruto"},               // exactly what is typed
		{Title: "  "},                   // empty
		{Title: "Naruto\n: Shippuuden"}, // multi-line
	}
	for i := range 8 {
		rows = append(rows, HomeSuggestion{Title: "Naruto " + string(rune('A'+i))})
	}
	rec := &recordingSuggest{items: rows}
	m := suggestModel(t, rec, SearchHomeOptions{})
	typeText(m, "naruto")
	require.Len(t, m.suggestions, maxHomeSuggestions, "the dropdown is capped")
	assert.NotContains(t, titlesOf(m.suggestions), "naruto")
	assert.Equal(t, "Naruto : Shippuuden", m.suggestions[0].Title, "rows are one line")
}

func TestSearchHomeSuggestKeys(t *testing.T) {
	t.Parallel()
	newOpen := func(t *testing.T, opts SearchHomeOptions) (*searchHomeModel, *recordingSuggest) {
		rec := &recordingSuggest{items: sugRows("Naruto", "Naruto: Shippuuden")}
		m := suggestModel(t, rec, opts)
		typeText(m, "nar")
		require.True(t, m.suggestionsOpen())
		return m, rec
	}

	t.Run("up and down walk the rows and back to the box", func(t *testing.T) {
		t.Parallel()
		m, _ := newOpen(t, SearchHomeOptions{})
		_, _ = m.Update(press(tea.KeyUp, ""))
		assert.Equal(t, -1, m.sugCursor, "up from the box stays in the box")
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		assert.Equal(t, 1, m.sugCursor)
		_, _ = m.Update(press(tea.KeyUp, ""))
		_, _ = m.Update(press(tea.KeyUp, ""))
		assert.Equal(t, -1, m.sugCursor)
		assert.True(t, m.focusInput)
	})

	t.Run("down past the last row enters the lists", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		m, _ := newOpen(t, SearchHomeOptions{Continue: continueAll})
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		assert.False(t, m.focusInput, "the lists below the box take over")
		assert.False(t, m.suggestionsOpen())
		assert.NotContains(t, ansi.Strip(m.View().Content), "Naruto: Shippuuden", "the dropdown hides while the lists have focus")
	})

	t.Run("right at the end completes with the first row", func(t *testing.T) {
		t.Parallel()
		m, rec := newOpen(t, SearchHomeOptions{})
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "Naruto", m.input.Value())
		assert.False(t, m.suggestionsOpen())
		assert.Len(t, rec.called(), 1, "completing never re-asks for the completed title")
		_, cmd := m.Update(press(tea.KeyLeft, ""))
		for _, msg := range msgsOf(cmd) {
			assert.IsNotType(t, suggestDueMsg{}, msg, "a cursor move after completing is not a new query")
		}
		_, cmd = m.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		assert.Equal(t, "Naruto", m.result.Query)
	})

	t.Run("right completes with the highlighted row", func(t *testing.T) {
		t.Parallel()
		m, _ := newOpen(t, SearchHomeOptions{})
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "Naruto: Shippuuden", m.input.Value())
	})

	t.Run("right inside the text just moves the cursor", func(t *testing.T) {
		t.Parallel()
		m, _ := newOpen(t, SearchHomeOptions{})
		_, _ = m.Update(press(tea.KeyLeft, ""))
		_, _ = m.Update(press(tea.KeyLeft, ""))
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "nar", m.input.Value())
		assert.True(t, m.suggestionsOpen())
	})

	t.Run("right with no rows still moves the cursor", func(t *testing.T) {
		t.Parallel()
		m := newSearchHomeModel(SearchHomeOptions{Query: "abc"})
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "abc", m.input.Value())
	})

	t.Run("esc closes the rows, then clears the box", func(t *testing.T) {
		t.Parallel()
		m, _ := newOpen(t, SearchHomeOptions{})
		_, _ = m.Update(press(tea.KeyEscape, ""))
		assert.False(t, m.suggestionsOpen())
		assert.Equal(t, "nar", m.input.Value(), "the first esc keeps the text")
		_, _ = m.Update(press(tea.KeyEscape, ""))
		assert.Empty(t, m.input.Value())
	})

	t.Run("esc on an empty box does nothing", func(t *testing.T) {
		t.Parallel()
		rec := &recordingSuggest{}
		m := suggestModel(t, rec, SearchHomeOptions{})
		_, cmd := m.Update(press(tea.KeyEscape, ""))
		assert.Nil(t, cmd)
		assert.Empty(t, rec.called())
	})

	t.Run("enter on the box (no highlight) searches the typed text", func(t *testing.T) {
		t.Parallel()
		m, _ := newOpen(t, SearchHomeOptions{})
		_, cmd := m.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		assert.Equal(t, "nar", m.result.Query)
	})

	t.Run("backspacing to a short query closes the rows without asking", func(t *testing.T) {
		t.Parallel()
		m, rec := newOpen(t, SearchHomeOptions{})
		before := len(rec.called())
		_, _ = m.Update(press(tea.KeyBackspace, ""))
		_, cmd := m.Update(press(tea.KeyBackspace, ""))
		settle(m, cmd)
		assert.False(t, m.suggestionsOpen())
		assert.Len(t, rec.called(), before)
	})

	t.Run("typing over the lists asks for the new text", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		rec := &recordingSuggest{items: sugRows("Frieren")}
		m := suggestModel(t, rec, SearchHomeOptions{Continue: continueAll})
		_, _ = m.Update(press(tea.KeyDown, ""))
		require.False(t, m.focusInput)
		typeText(m, "fr")
		assert.True(t, m.focusInput)
		assert.Equal(t, []suggestCall{{"fr", ""}}, rec.called())
	})
}

func TestSearchHomeSuggestFollowsScope(t *testing.T) {
	t.Parallel()
	rec := &recordingSuggest{items: sugRows("Dune")}
	m := suggestModel(t, rec, SearchHomeOptions{})
	typeText(m, "dun")
	require.Equal(t, []suggestCall{{"dun", ""}}, rec.called())

	for range 3 { // ALL -> PT-BR -> ENGLISH -> MOVIES/TV
		_, cmd := m.Update(press(tea.KeyTab, ""))
		settle(m, cmd)
	}
	assert.Equal(t, []suggestCall{
		{"dun", ""}, {"dun", "ptbr"}, {"dun", "hianime"}, {"dun", "movietv"},
	}, rec.called(), "every scope change asks again, in the new scope")

	t.Run("tab with no text asks for nothing", func(t *testing.T) {
		rec := &recordingSuggest{}
		m := suggestModel(t, rec, SearchHomeOptions{})
		_, cmd := m.Update(press(tea.KeyTab, ""))
		settle(m, cmd)
		assert.Empty(t, rec.called())
	})
}

func TestSearchHomePrefilledQueryDoesNotAutoAsk(t *testing.T) {
	t.Parallel()
	rec := &recordingSuggest{items: sugRows("Naruto")}
	m := suggestModel(t, rec, SearchHomeOptions{Query: "naruto"})
	assert.Empty(t, rec.called(), "a prefilled box shows no dropdown until the user types")
	assert.Equal(t, m.suggestKey(), m.asked)
	_, cmd := m.Update(press(tea.KeyLeft, ""))
	assert.Empty(t, msgsOf(cmd), "moving the cursor is not a change")
}

func TestSearchHomeSuggestHintsAndPlain(t *testing.T) {
	rec := &recordingSuggest{items: sugRows("Naruto")}
	m := suggestModel(t, rec, SearchHomeOptions{})
	keysOf := func() []string {
		var out []string
		for _, h := range m.hints() {
			out = append(out, h.Key)
		}
		return out
	}
	assert.NotContains(t, keysOf(), "→")
	typeText(m, "nar")
	assert.Equal(t, []string{"enter", "↑↓", "→", "tab", "?", "ctrl+c"}, keysOf())

	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	assert.Equal(t, []string{"enter", "up/dn", "right", "tab", "?", "ctrl+c"}, keysOf())
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "Naruto")
	assert.Contains(t, view, "right")
}

func TestSearchHomeSuggestResultLeavesCancelClean(t *testing.T) {
	t.Parallel()
	rec := &recordingSuggest{items: sugRows("Naruto")}
	m := suggestModel(t, rec, SearchHomeOptions{})
	typeText(m, "nar")
	assert.Nil(t, m.cancel, "a finished lookup holds no cancel func")
	m.dropSuggestions() // with nothing in flight
	assert.Empty(t, m.suggestions)
}

func ghostOf(m *searchHomeModel) []string { return m.input.MatchedSuggestions() }

// TestSearchHomeGhostText covers the editor-style inline suggestion: the rest
// of the best completion drawn after the typed text, beside the dropdown.
func TestSearchHomeGhostText(t *testing.T) {
	t.Parallel()
	open := func(t *testing.T, items ...string) (*searchHomeModel, *recordingSuggest) {
		rec := &recordingSuggest{items: sugRows(items...)}
		m := suggestModel(t, rec, SearchHomeOptions{})
		typeText(m, "nar")
		return m, rec
	}

	t.Run("shows the first completion after the typed text", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Like Naruto", "Naruto", "Naruto: Shippuuden")
		assert.Equal(t, []string{"Naruto"}, ghostOf(m), "rows that do not continue the text give no ghost")
		view := ansi.Strip(m.View().Content)
		assert.Contains(t, view, "naruto", "typed 'nar' + ghost 'uto' reads as one word (the rows are capitalised)")
	})

	t.Run("follows the highlighted row and returns to the best one", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Naruto", "Naruto: Shippuuden")
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyDown, ""))
		assert.Equal(t, []string{"Naruto: Shippuuden"}, ghostOf(m))
		assert.Contains(t, ansi.Strip(m.View().Content), "naruto: Shippuuden", "typed text plus the rest in the row's own case")
		_, _ = m.Update(press(tea.KeyUp, ""))
		_, _ = m.Update(press(tea.KeyUp, ""))
		assert.Equal(t, []string{"Naruto"}, ghostOf(m))
	})

	t.Run("matches ignoring case", func(t *testing.T) {
		t.Parallel()
		rec := &recordingSuggest{items: sugRows("Naruto")}
		m := suggestModel(t, rec, SearchHomeOptions{})
		typeText(m, "NAR")
		assert.Equal(t, []string{"Naruto"}, ghostOf(m))
	})

	t.Run("no row continues the text: no ghost, no completion", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Like Naruto")
		assert.Empty(t, ghostOf(m))
		_, ok := m.completionPick()
		assert.False(t, ok)
		for _, h := range m.hints() {
			assert.NotEqual(t, "→", h.Key, "no completion to teach")
		}
		_, _ = m.Update(press(tea.KeyLeft, ""))
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "nar", m.input.Value(), "right only moves the cursor")
		assert.True(t, m.suggestionsOpen(), "the dropdown still shows")
	})

	t.Run("a highlighted row completes even when it does not continue the text", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Like Naruto")
		_, _ = m.Update(press(tea.KeyDown, ""))
		_, _ = m.Update(press(tea.KeyRight, ""))
		assert.Equal(t, "Like Naruto", m.input.Value())
	})

	t.Run("goes away with typing", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Naruto")
		require.NotEmpty(t, ghostOf(m))
		_, _ = m.Update(press('x', "x"))
		assert.Empty(t, ghostOf(m), "a ghost for old text never sits under new text")
	})

	t.Run("goes away with esc", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Naruto")
		require.NotEmpty(t, ghostOf(m))
		_, _ = m.Update(press(tea.KeyEscape, ""))
		assert.Empty(t, ghostOf(m), "esc closes the dropdown and the ghost together")
	})

	t.Run("goes away when focus leaves the box", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		rec := &recordingSuggest{items: sugRows("Naruto")}
		m := suggestModel(t, rec, SearchHomeOptions{Continue: continueAll})
		typeText(m, "nar")
		require.NotEmpty(t, ghostOf(m))
		_, _ = m.Update(press(tea.KeyDown, "")) // onto the dropdown row
		_, _ = m.Update(press(tea.KeyDown, "")) // past it, into the lists
		require.False(t, m.focusInput)
		assert.Empty(t, ghostOf(m), "no ghost while the lists have focus")
	})

	t.Run("the widget's own keys stay off", func(t *testing.T) {
		t.Parallel()
		m, _ := open(t, "Naruto", "Naruto: Shippuuden")
		_, _ = m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
		assert.Equal(t, "nar", m.input.Value())
		assert.Equal(t, []string{"Naruto"}, ghostOf(m), "ctrl+n does not cycle the ghost away from the pick")
		_, _ = m.Update(press(tea.KeyTab, ""))
		assert.Equal(t, "nar", m.input.Value(), "tab changes the scope; it never accepts the ghost")
		assert.Equal(t, 1, m.scopeIdx)
	})
}
