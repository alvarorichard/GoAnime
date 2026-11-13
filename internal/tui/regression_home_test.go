package tui

// Regression net for the SEARCH home as it is TODAY (before the lists level
// lives beside Favorites): every screen state at four terminal sizes in rich
// and in plain mode, the footer and help text of every state, the scope tabs,
// and the SearchHomeResult of every action.
//
// The screens are frozen as golden files (testdata/regression/home_*.golden,
// see regression_helpers_test.go); the text users read in the footer, the
// notices and the questions is also asserted inline, exactly.
//
// Where today's behavior is arguably wrong the assertion still freezes it and
// carries a "frozen:" comment. The next steps may change a frozen assertion
// ONLY on purpose.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var regNow = time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC)

func regContinue() []HomeContinueEntry {
	return []HomeContinueEntry{
		{Key: "c1", AnimeTitle: "Mukashi Banashi", EpisodeTitle: "Urashima Tarou", EpisodeNum: 5, Position: 372, Duration: 580, Source: "Animefire.io", Audio: "Legendado", MediaType: "anime", UpdatedAt: regNow.Add(-24 * time.Hour)},
		{Key: "c2", AnimeTitle: "Kuro Nyago", EpisodeNum: 1, Position: 160, Duration: 600, Source: "Animefire.io", UpdatedAt: regNow.Add(-72 * time.Hour)},
		{Key: "c3", AnimeTitle: "Saru to Kani no Gassen", EpisodeNum: 1, Position: 590, Duration: 600, UpdatedAt: regNow.Add(-7 * 24 * time.Hour)},
	}
}

func regRecent() []HomeRecentSearch {
	return []HomeRecentSearch{{Query: "namakura gatana", Count: 5}, {Query: "momotarou", Count: 1}}
}

// regFavorites is the first n of a fixed list that exercises accents, long
// titles and missing year, type and source.
func regFavorites(n int) []HomeFavorite {
	all := []HomeFavorite{
		{Key: "frieren", Title: "Frieren: Beyond Journey's End", Year: "2023", Media: "Anime", Source: "Animefire.io"},
		{Key: "heat", Title: "Heat", Year: "1995", Media: "Movie", Source: "StartFlix"},
		{Key: "pokemon aventuras", Title: "Pokémon: Aventuras Pokémon", Year: "1999", Media: "Anime", Source: "Goyabu"},
		{Key: "dark", Title: "Dark", Year: "2017", Media: "TV", Source: "Pobreflix"},
		{Key: "mushishi", Title: "Mushishi", Year: "2005", Media: "Anime"},
		{Key: "spirited away", Title: "Spirited Away", Media: "Movie", Source: "TopCine"},
		{Key: "cowboy bebop", Title: "Cowboy Bebop", Year: "1998", Source: "HiAnime"},
		{Key: "fma", Title: "Fullmetal Alchemist: Brotherhood and the Extremely Long Subtitle That Never Ends", Year: "2009", Media: "Anime", Source: "AnimesDigital"},
		{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime", Source: "HiAnime"},
		{Key: "the matrix", Title: "The Matrix", Year: "1999", Media: "Movie", Source: "StartFlix"},
		{Key: "dandadan", Title: "Dandadan", Year: "2024", Media: "Anime", Source: "Animefire.io"},
		{Key: "severance", Title: "Severance", Year: "2022", Media: "TV", Source: "Pobreflix"},
		{Key: "akira", Title: "Akira", Year: "1988", Media: "Movie", Source: "AnimesROLL"},
	}
	return all[:n]
}

// regCalls records the host callbacks of a home screen.
type regCalls struct {
	log []string
	err error
}

func (c *regCalls) apply(o *SearchHomeOptions) {
	o.DeleteEntry = func(k string) error { c.log = append(c.log, "delete-continue:"+k); return c.err }
	o.DeleteRecent = func(q string) error { c.log = append(c.log, "delete-recent:"+q); return c.err }
	o.DeleteFavorite = func(k string) error { c.log = append(c.log, "delete-favorite:"+k); return c.err }
	o.ClearContinue = func() error { c.log = append(c.log, "clear-continue"); return c.err }
	o.ClearRecent = func() error { c.log = append(c.log, "clear-recent"); return c.err }
	o.ClearFavorites = func() error { c.log = append(c.log, "clear-favorites"); return c.err }
}

// regHome describes a home screen to build.
type regHome struct {
	continues bool // CONTINUE rows (3)
	recents   bool // RECENT rows (2)
	favs      int
	scope     int
	query     string
	notice    string
	calls     *regCalls
	sources   bool
	noHost    bool // no delete/clear callbacks at all
}

func (h regHome) build() *searchHomeModel {
	opts := SearchHomeOptions{
		Query: h.query, ScopeIndex: h.scope, SourcesLine: "6 sources on · Goyabu off",
		Favorites: regFavorites(h.favs), Notice: h.notice,
	}
	if h.continues {
		opts.Continue = regContinue()
	}
	if h.recents {
		opts.Recent = regRecent()
	}
	if !h.noHost {
		calls := h.calls
		if calls == nil {
			calls = &regCalls{}
		}
		calls.apply(&opts)
	}
	if h.sources {
		opts.Sources = func() []SourceRow { return []SourceRow{{ID: "a", Name: "Alpha", On: true}, {ID: "b", Name: "Beta"}} }
		opts.SetSource = func(string, bool) error { return nil }
	}
	m := newSearchHomeModel(opts)
	m.now = func() time.Time { return regNow }
	return m
}

type regScene struct {
	name string
	home regHome
	keys []string
}

// regHomeGolden renders every scene at every size in rich and in plain mode
// and compares each with its section of the golden file.
func regHomeGolden(t *testing.T, file string, scenes []regScene) {
	t.Helper()
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		for _, sc := range scenes {
			for _, size := range regSizes {
				m := sc.home.build()
				_, _ = m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
				regPress(m, sc.keys...)
				view := m.View()
				regGolden(t, file, fmt.Sprintf("%s %s %s", regMode(plain), size, sc.name), regFrame(t, view.Content, size))
			}
		}
	}
}

func regRender(m *searchHomeModel) string {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}

func stripANSI(s string) string { return ansi.Strip(s) }

// ---- the ALL tab -------------------------------------------------------------------

func TestRegressionHomeAllTabFavoriteCounts(t *testing.T) {
	var scenes []regScene
	for _, n := range []int{0, 1, 8, 12} {
		scenes = append(scenes,
			regScene{name: fmt.Sprintf("full favs=%d", n), home: regHome{continues: true, recents: true, favs: n}},
			regScene{name: fmt.Sprintf("bare favs=%d", n), home: regHome{favs: n}},
			regScene{name: fmt.Sprintf("favs-only favs=%d sources", n), home: regHome{favs: n, sources: true, query: "frieren"}},
		)
	}
	regHomeGolden(t, "home_all_counts", scenes)
}

func TestRegressionHomeAllTabContentCombinations(t *testing.T) {
	regHomeGolden(t, "home_all_content", []regScene{
		{name: "continue only", home: regHome{continues: true, favs: 3}},
		{name: "recent only", home: regHome{recents: true, favs: 3}},
		{name: "continue and recent, no favorites", home: regHome{continues: true, recents: true}},
		{name: "continue only, row focus on the first", home: regHome{continues: true, favs: 3}, keys: []string{"down"}},
		{name: "recent only, row focus on the first", home: regHome{recents: true, favs: 3}, keys: []string{"down", "down", "down", "down"}},
		{name: "query kept in the box", home: regHome{continues: true, recents: true, favs: 2, query: "namak"}},
		{name: "a notice from the host", home: regHome{continues: true, recents: true, favs: 2, notice: "Your favorites file was damaged, so the list starts empty. A copy was kept as favorites.json.corrupt.20240310-120000."}},
	})
}

func TestRegressionHomeAllTabRowFocus(t *testing.T) {
	full := regHome{continues: true, recents: true, favs: 12}
	down := func(n int) []string {
		keys := make([]string, n)
		for i := range keys {
			keys[i] = "down"
		}
		return keys
	}
	regHomeGolden(t, "home_all_rowfocus", []regScene{
		{name: "continue 1", home: full, keys: down(1)},
		{name: "continue 3 (done)", home: full, keys: down(3)},
		{name: "favorite 1", home: full, keys: down(4)},
		{name: "favorite 8, the last drawn", home: full, keys: down(11)},
		{name: "recent 1", home: full, keys: down(12)},
		{name: "recent 2, the last row", home: full, keys: down(13)},
		{name: "past the end stays on the last", home: full, keys: down(20)},
		{name: "back up to the box", home: full, keys: append(down(1), "up")},
		{name: "pgdown", home: full, keys: append(down(1), "pgdown")},
		{name: "end then pgup", home: full, keys: append(down(1), "end", "pgup")},
		{name: "end then home", home: full, keys: append(down(1), "end", "home")},
		{name: "few favorites row focus", home: regHome{continues: true, recents: true, favs: 2}, keys: down(5)},
	})
}

func TestRegressionHomeClearQuestions(t *testing.T) {
	full := regHome{continues: true, recents: true, favs: 12}
	one := regHome{continues: true, recents: true, favs: 1}
	regHomeGolden(t, "home_questions", []regScene{
		{name: "CONTINUE", home: full, keys: []string{"down", "X"}},
		{name: "FAVORITES", home: full, keys: []string{"down", "down", "down", "down", "X"}},
		{name: "FAVORITES one", home: one, keys: []string{"down", "down", "down", "down", "X"}},
		{name: "RECENT", home: full, keys: []string{"end", "X"}},
		{name: "RECENT after the question", home: full, keys: []string{"end", "X", "n"}},
		{name: "a stray key behind the question", home: full, keys: []string{"down", "X", "type:zz", "x", "down"}},
		{name: "FAVORITES tab", home: regHome{favs: 5, scope: 4}, keys: []string{"down", "X"}},
	})
}

func TestRegressionHomeFailureNotices(t *testing.T) {
	boom := &regCalls{err: errors.New("disk full")}
	full := regHome{continues: true, recents: true, favs: 12, calls: boom}
	regHomeGolden(t, "home_failures", []regScene{
		{name: "remove an episode", home: full, keys: []string{"down", "x"}},
		{name: "remove a favorite", home: full, keys: []string{"down", "down", "down", "down", "x"}},
		{name: "remove a search", home: full, keys: []string{"end", "x"}},
		{name: "clear CONTINUE", home: full, keys: []string{"down", "X", "y"}},
		{name: "clear FAVORITES", home: full, keys: []string{"down", "down", "down", "down", "X", "y"}},
		{name: "clear RECENT", home: full, keys: []string{"end", "X", "y"}},
		{name: "remove a favorite on the tab", home: regHome{favs: 5, scope: 4, calls: boom}, keys: []string{"down", "x"}},
		{name: "the notice goes at the next key", home: full, keys: []string{"down", "x", "down"}},
	})
}

func TestRegressionHomeTheHostNoticeGoesAtTheFirstKey(t *testing.T) {
	regPin(t, false)
	m := regHome{favs: 1, notice: "  A\tline\nof   text "}.build()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	assert.Equal(t, "A line of text", m.notice, "a notice is one clean line")
	assert.Contains(t, m.View().Content, "A line of text")
	regPress(m, "type:a")
	assert.Equal(t, "", m.notice)
	assert.NotContains(t, m.View().Content, "A line of text")
}

// ---- tabs ----------------------------------------------------------------------------

func TestRegressionHomeScopeTabs(t *testing.T) {
	regPin(t, false)
	// frozen: the tab order, names and the value each runs the search with.
	wantNames := []string{"ALL", "PT-BR", "ENGLISH", "MOVIES/TV", "FAVORITES"}
	wantValues := []string{"", "ptbr", "hianime", "movietv", "favorites"}
	require.Len(t, homeScopes, 5)
	for i, s := range homeScopes {
		assert.Equal(t, wantNames[i], s.Name)
		assert.Equal(t, wantValues[i], s.Value)
	}
	require.Len(t, searchScopes, 4)
	for i, s := range searchScopes {
		assert.Equal(t, wantNames[i], s.Name)
		assert.Equal(t, wantValues[i], s.Value)
	}
	assert.Equal(t, "favorites", favoritesScope)

	m := regHome{continues: true, recents: true, favs: 3}.build()
	assert.Equal(t, 0, m.scopeIdx)
	order := []int{}
	for range 6 {
		regPress(m, "tab")
		order = append(order, m.scopeIdx)
	}
	assert.Equal(t, []int{1, 2, 3, 4, 0, 1}, order, "tab goes right and wraps from FAVORITES to ALL")

	for _, idx := range []int{-3, -1, 5, 6, 99} {
		assert.Equal(t, 0, regHome{scope: idx}.build().scopeIdx, "scope %d is out of range and falls on ALL", idx)
	}
	for idx := range 5 {
		assert.Equal(t, idx, regHome{scope: idx}.build().scopeIdx)
	}
}

func TestRegressionHomeTabRowText(t *testing.T) {
	// frozen: the active tab is padded by its tag style (one blank each side),
	// the others are not, and the "tab" keycap is padded too.
	for _, tt := range []struct {
		plain bool
		idx   int
		want  string
	}{
		{false, 0, "in  ALL    PT-BR   ENGLISH   MOVIES/TV   FAVORITES      tab  change"},
		{false, 4, "in ALL   PT-BR   ENGLISH   MOVIES/TV    FAVORITES       tab  change"},
		{true, 0, "in  [ALL]    PT-BR   ENGLISH   MOVIES/TV   FAVORITES      tab  change"},
		{true, 1, "in ALL    [PT-BR]    ENGLISH   MOVIES/TV   FAVORITES      tab  change"},
		{true, 2, "in ALL   PT-BR    [ENGLISH]    MOVIES/TV   FAVORITES      tab  change"},
		{true, 3, "in ALL   PT-BR   ENGLISH    [MOVIES/TV]    FAVORITES      tab  change"},
		{true, 4, "in ALL   PT-BR   ENGLISH   MOVIES/TV    [FAVORITES]       tab  change"},
	} {
		regPin(t, tt.plain)
		m := regHome{scope: tt.idx}.build()
		assert.Equal(t, tt.want, stripANSI(m.scopeTabs()), "%s tab %d", regMode(tt.plain), tt.idx)
	}
}

func TestRegressionHomeRichTabIsMarkedByStyleNotBrackets(t *testing.T) {
	regPin(t, false)
	for idx, name := range []string{"ALL", "PT-BR", "ENGLISH", "MOVIES/TV", "FAVORITES"} {
		m := regHome{scope: idx}.build()
		raw := m.scopeTabs()
		assert.Contains(t, raw, m.theme.Tag.Render(name), "the active tab %s is drawn as a tag", name)
		for j, other := range []string{"ALL", "PT-BR", "ENGLISH", "MOVIES/TV", "FAVORITES"} {
			if j != idx {
				assert.Contains(t, raw, m.theme.Muted.Render(other))
				assert.NotContains(t, raw, m.theme.Tag.Render(other))
			}
		}
		assert.NotContains(t, stripANSI(raw), "[")
	}
}

func TestRegressionHomeBoxPlaceholders(t *testing.T) {
	for _, tt := range []struct {
		plain bool
		want  string
	}{
		{false, "title, or ↓ to pick from the lists below"},
		{true, "title, or down to pick from the lists below"},
	} {
		regPin(t, tt.plain)
		m := regHome{}.build()
		assert.Equal(t, tt.want, m.input.Placeholder)
		assert.Equal(t, tt.want, m.searchPlaceholder)
		assert.Equal(t, map[bool]string{false: "❯ ", true: "> "}[tt.plain], m.input.Prompt)
		regPress(m, "tab", "tab", "tab", "tab")
		assert.Equal(t, "type to filter your favorites, enter opens the first one", m.input.Placeholder)
		regPress(m, "tab")
		assert.Equal(t, tt.want, m.input.Placeholder, "back on ALL the search hint returns")
		assert.Equal(t, "type to filter your favorites, enter opens the first one", favoritesPlaceholder)
		assert.Equal(t, "type to filter your favorites, enter opens the first one", regHome{scope: 4}.build().input.Placeholder)
	}
}

// ---- the FAVORITES tab (legacy single list) -----------------------------------------------

func TestRegressionHomeFavoritesTab(t *testing.T) {
	down := func(n int) []string {
		keys := make([]string, n)
		for i := range keys {
			keys[i] = "down"
		}
		return keys
	}
	tab := func(n int) regHome {
		return regHome{continues: true, recents: true, favs: n, scope: 4}
	}
	regHomeGolden(t, "home_favorites_tab", []regScene{
		{name: "empty", home: tab(0)},
		{name: "empty with a query", home: regHome{scope: 4, query: "frieren"}},
		{name: "1", home: tab(1)},
		{name: "5", home: tab(5)},
		{name: "5 row 3", home: tab(5), keys: down(3)},
		{name: "13", home: tab(13)},
		{name: "13 row 1", home: tab(13), keys: down(1)},
		{name: "13 row 7", home: tab(13), keys: down(7)},
		{name: "13 row 12", home: tab(13), keys: down(12)},
		{name: "13 last", home: tab(13), keys: []string{"down", "end"}},
		{name: "13 last then home", home: tab(13), keys: []string{"down", "end", "home"}},
		{name: "13 pgdown", home: tab(13), keys: []string{"down", "pgdown"}},
		{name: "13 pgdown pgdown pgup", home: tab(13), keys: []string{"down", "pgdown", "pgdown", "pgup"}},
		{name: "filter movie", home: regHome{favs: 13, scope: 4, query: "movie"}},
		{name: "filter movie row 2", home: regHome{favs: 13, scope: 4, query: "movie"}, keys: down(2)},
		{name: "filter accent-less pokemon", home: regHome{favs: 13, scope: 4, query: "pokemon"}},
		{name: "filter two words", home: regHome{favs: 13, scope: 4, query: "anime  1999"}},
		{name: "filter by source", home: regHome{favs: 13, scope: 4, query: "animefire"}},
		{name: "filter typed", home: tab(13), keys: []string{"type:dan"}},
		{name: "filter typed then row", home: tab(13), keys: []string{"type:a", "down", "down"}},
		{name: "no match", home: regHome{favs: 13, scope: 4, query: "zzz  qq"}},
		{name: "no match typed", home: tab(13), keys: []string{"type:zzz"}},
		{name: "damaged-file notice", home: regHome{scope: 4, notice: "Your favorites file was damaged, so the list starts empty. A copy was kept as favorites.json.corrupt.20240310-120000."}},
		{name: "damaged-file notice with favorites", home: regHome{favs: 5, scope: 4, notice: "Your favorites file was damaged, so the list starts empty. A copy was kept as favorites.json.corrupt.20240310-120000."}},
		{name: "clear question", home: tab(13), keys: []string{"down", "X"}},
		{name: "clear question after yes", home: tab(13), keys: []string{"down", "X", "y"}},
		{name: "remove the first", home: tab(5), keys: []string{"down", "x"}},
		{name: "remove the last of 5", home: tab(5), keys: []string{"end", "x"}},
		{name: "remove while scrolled", home: tab(13), keys: []string{"down", "end", "x"}},
		{name: "from the tab back to ALL", home: tab(5), keys: []string{"tab"}},
		{name: "tab keeps the typed query", home: regHome{favs: 5, scope: 3, query: "heat"}, keys: []string{"tab"}},
	})
}

// regFavWindow is the FAVORITES tab's window as one line: the "more above" and
// "more below" lines, and the keys of the favorites drawn, in order.
func regFavWindow(m *searchHomeModel) string {
	text := regRender(m)
	var above, below string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "↑ ") {
			above = trimmed
		}
		if strings.HasPrefix(trimmed, "↓ ") {
			below = trimmed
		}
	}
	var drawn []string
	for _, f := range regFavorites(13) {
		if strings.Contains(text, f.Title) || strings.Contains(text, f.Title[:min(len(f.Title), 20)]+" ") {
			drawn = append(drawn, f.Key)
		}
	}
	return fmt.Sprintf("above=%q below=%q drawn=%v", above, below, drawn)
}

func TestRegressionHomeFavoritesTabWindowMath(t *testing.T) {
	regPin(t, false)
	var sb strings.Builder
	for _, h := range []int{6, 7, 8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 30, 40} {
		for _, row := range []int{0, 2, 5, 6, 9, 12} {
			m := regHome{favs: 13, scope: 4}.build()
			_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: h})
			keys := make([]string, 0, row+1)
			for range row + 1 {
				keys = append(keys, "down")
			}
			regPress(m, keys...)
			fmt.Fprintf(&sb, "h=%d row=%d %s\n", h, row, regFavWindow(m))
		}
	}
	regGolden(t, "home_favorites_window", "window math", strings.TrimSuffix(sb.String(), "\n"))
}

// ---- hints and help -------------------------------------------------------------------------

func TestRegressionHomeHints(t *testing.T) {
	rich := func(h regHome, keys ...string) string {
		regPin(t, false)
		m := h.build()
		regPress(m, keys...)
		return regHintLine(m.hints())
	}
	plain := func(h regHome, keys ...string) string {
		regPin(t, true)
		m := h.build()
		regPress(m, keys...)
		return regHintLine(m.hints())
	}
	full := regHome{continues: true, recents: true, favs: 3}

	// ALL tab, box focus.
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | ?:help | ctrl+c:quit!", rich(full))
	assert.Equal(t, "enter:search! | up/dn:move | tab:scope | ?:help | ctrl+c:quit!", plain(full))
	// Row focus: x and X come after tab.
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(full, "down"))
	assert.Equal(t, "enter:search! | up/dn:move | tab:scope | x:remove! | X:clear list | ?:help | ctrl+c:quit!", plain(full, "down"))
	// With the sources screen, ctrl+s goes after tab.
	withSources := regHome{continues: true, recents: true, favs: 3, sources: true}
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | ctrl+s:sources! | ?:help | ctrl+c:quit!", rich(withSources))
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | ctrl+s:sources! | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(withSources, "down"))
	// FAVORITES tab: enter opens.
	assert.Equal(t, "enter:open! | ↑↓:move | tab:scope | ?:help | ctrl+c:quit!", rich(full, "tab", "tab", "tab", "tab"))
	assert.Equal(t, "enter:open! | up/dn:move | tab:scope | ?:help | ctrl+c:quit!", plain(full, "tab", "tab", "tab", "tab"))
	assert.Equal(t, "enter:open! | ↑↓:move | tab:scope | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(full, "tab", "tab", "tab", "tab", "down"))
	assert.Equal(t, "enter:open! | ↑↓:move | tab:scope | ctrl+s:sources! | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(withSources, "tab", "tab", "tab", "tab", "down"))
	// The tab after FAVORITES is ALL again.
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | ?:help | ctrl+c:quit!", rich(full, "tab", "tab", "tab", "tab", "tab"))
	// The other search tabs search too.
	for n := 1; n <= 3; n++ {
		keys := make([]string, n)
		for i := range keys {
			keys[i] = "tab"
		}
		assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | ?:help | ctrl+c:quit!", rich(full, keys...))
	}
	// An empty FAVORITES tab has no rows to focus: the box keeps the focus.
	assert.Equal(t, "enter:open! | ↑↓:move | tab:scope | ?:help | ctrl+c:quit!", rich(regHome{scope: 4}, "down"))
	// The clear question.
	const question = "y:clear all! | n:keep! | ctrl+c:quit"
	assert.Equal(t, question, rich(full, "down", "X"))
	assert.Equal(t, question, plain(full, "down", "X"))
	assert.Equal(t, question, rich(regHome{favs: 3, scope: 4}, "down", "X"))
	// The help overlay.
	const help = "?:close! | esc:close!"
	assert.Equal(t, help, rich(full, "down", "?"))
	assert.Equal(t, help, plain(full, "down", "?"))
	assert.Equal(t, help, rich(regHome{favs: 3, scope: 4}, "down", "?"))
	// The overlay is closed by ? and by esc, and the footer returns.
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(full, "down", "?", "?"))
	assert.Equal(t, "enter:search! | ↑↓:move | tab:scope | x:remove! | X:clear list | ?:help | ctrl+c:quit!", rich(full, "down", "?", "esc"))
}

func TestRegressionHomeHintsWithACompletion(t *testing.T) {
	for _, tt := range []struct {
		plain bool
		want  string
	}{
		{false, "enter:search! | ↑↓:move | →:complete | tab:scope | ?:help | ctrl+c:quit!"},
		{true, "enter:search! | up/dn:move | right:complete | tab:scope | ?:help | ctrl+c:quit!"},
	} {
		regPin(t, tt.plain)
		m := regHome{continues: true, favs: 2}.build()
		regPress(m, "type:fri")
		m.suggestions = []HomeSuggestion{{Title: "Frieren", Detail: "Anime · 2023"}, {Title: "Friends", Detail: "Series · 1994"}}
		m.syncGhost()
		assert.Equal(t, tt.want, regHintLine(m.hints()))
		// With sources: the order is enter, move, complete, tab, ctrl+s.
		m.loadSources = func() []SourceRow { return nil }
		assert.Equal(t, strings.Replace(tt.want, "tab:scope", "tab:scope | ctrl+s:sources!", 1), regHintLine(m.hints()))
	}
}

func TestRegressionHomeDropdownAndGhost(t *testing.T) {
	size := regSize{100, 30}
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		m := regHome{continues: true, recents: true, favs: 3}.build()
		_, _ = m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
		regPress(m, "type:fri")
		m.suggestions = []HomeSuggestion{{Title: "Frieren", Detail: "Anime · 2023"}, {Title: "Friends", Detail: "Series · 1994"}, {Title: "Fringe", Detail: "Series · 2008"}}
		regGolden(t, "home_dropdown", regMode(plain)+" closed highlight", regFrame(t, m.View().Content, size))
		regPress(m, "down")
		assert.Equal(t, 0, m.sugCursor)
		regGolden(t, "home_dropdown", regMode(plain)+" first highlighted", regFrame(t, m.View().Content, size))
		regPress(m, "down", "down")
		regGolden(t, "home_dropdown", regMode(plain)+" third highlighted", regFrame(t, m.View().Content, size))
	}
}

func TestRegressionHomeHelpOverlay(t *testing.T) {
	full := regHome{continues: true, recents: true, favs: 3}
	regHomeGolden(t, "home_help", []regScene{
		{name: "ALL tab", home: full, keys: []string{"down", "?"}},
		{name: "FAVORITES tab", home: regHome{favs: 3, scope: 4}, keys: []string{"down", "?"}},
		{name: "ALL tab with sources", home: regHome{continues: true, favs: 3, sources: true}, keys: []string{"down", "?"}},
	})
}

func TestRegressionHomeHelpRowsAreExact(t *testing.T) {
	regPin(t, false)
	m := regHome{continues: true, recents: true, favs: 3}.build()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	regPress(m, "down", "?")
	text := regRender(m)
	for _, want := range []string{
		"THIS SCREEN · SEARCH",
		"search, resume",
		"change scope; FAVORITES lists your favorites",
		"choose sources",
		"from the start",
		"episodes",
		"remove",
		"clear the list",
		"esc never quits. ctrl+c always does.",
	} {
		assert.Contains(t, text, want)
	}
	// The rows of THIS SCREEN come in this order.
	order := []string{"search, resume", "change scope; FAVORITES lists your favorites", "choose sources", "from the start", "episodes", "remove", "clear the list"}
	last := -1
	for _, o := range order {
		at := strings.Index(text, o)
		require.Greater(t, at, last, o)
		last = at
	}
	// ? is not a help key while the box has the focus: it types.
	m2 := regHome{continues: true}.build()
	regPress(m2, "?")
	assert.False(t, m2.showHelp)
	assert.Equal(t, "?", m2.input.Value())
	// Keys other than ? esc ctrl+c do nothing behind the overlay.
	regPress(m, "x", "X", "r", "e", "down", "tab", "enter")
	assert.True(t, m.showHelp)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.scopeIdx)
	assert.Equal(t, clearNothing, m.clearing)
	assert.True(t, regPress(m, "ctrl+c"), "ctrl+c quits from the help")
	assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
}

func TestRegressionHomeSourcesScreenIsHandedOverIntact(t *testing.T) {
	regPin(t, false)
	m := regHome{continues: true, favs: 2, sources: true}.build()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	regPress(m, "ctrl+s")
	require.NotNil(t, m.sources)
	assert.Equal(t, regHintLine(m.sources.hints()), regHintLine(m.hints()))
	assert.Equal(t, "6 sources on · Goyabu off", m.shell.Context)
	regPress(m, "esc")
	assert.Nil(t, m.sources)
	// Without the host's loader the key does nothing.
	m2 := regHome{continues: true}.build()
	regPress(m2, "ctrl+s")
	assert.Nil(t, m2.sources)
}

// ---- results of every action ----------------------------------------------------------------

func TestRegressionHomeEnterResults(t *testing.T) {
	regPin(t, false)
	cont := regContinue()
	downN := func(n int) []string {
		keys := make([]string, n)
		for i := range keys {
			keys[i] = "down"
		}
		return keys
	}
	tests := []struct {
		name string
		home regHome
		keys []string
		want SearchHomeResult
		quit bool
	}{
		{name: "box, ALL", home: regHome{continues: true, query: "frieren"}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "frieren", Scope: ""}, quit: true},
		{name: "box, PT-BR", home: regHome{query: "frieren", scope: 1}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "frieren", Scope: "ptbr"}, quit: true},
		{name: "box, ENGLISH", home: regHome{query: "frieren", scope: 2}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "frieren", Scope: "hianime"}, quit: true},
		{name: "box, MOVIES/TV", home: regHome{query: "heat", scope: 3}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "heat", Scope: "movietv"}, quit: true},
		{name: "box, tab to MOVIES/TV", home: regHome{query: "heat"}, keys: []string{"tab", "tab", "tab", "enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "heat", Scope: "movietv"}, quit: true},
		{name: "box, typed", home: regHome{}, keys: []string{"type:  one piece ", "enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "one piece", Scope: ""}, quit: true},
		{name: "box, two letters is enough", home: regHome{query: "ab"}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "ab"}, quit: true},
		{name: "box, one letter stays", home: regHome{query: "a"}, keys: []string{"enter"}, quit: false},
		{name: "box, empty stays", home: regHome{continues: true}, keys: []string{"enter"}, quit: false},
		{name: "box, blanks stay", home: regHome{query: "    "}, keys: []string{"enter"}, quit: false},
		{name: "continue 1", home: regHome{continues: true, recents: true, favs: 2}, keys: []string{"down", "enter"},
			want: SearchHomeResult{Action: HomeResume, Entry: cont[0], HasEntry: true, Scope: ""}, quit: true},
		{name: "continue 3 with scope PT-BR", home: regHome{continues: true, recents: true, favs: 2, scope: 1}, keys: append(downN(3), "enter"),
			want: SearchHomeResult{Action: HomeResume, Entry: cont[2], HasEntry: true, Scope: "ptbr"}, quit: true},
		{name: "continue 2 with scope MOVIES/TV", home: regHome{continues: true, scope: 3}, keys: []string{"down", "down", "enter"},
			want: SearchHomeResult{Action: HomeResume, Entry: cont[1], HasEntry: true, Scope: "movietv"}, quit: true},
		{name: "favorite 1 searches every scope", home: regHome{continues: true, recents: true, favs: 3}, keys: append(downN(4), "enter"),
			want: SearchHomeResult{Action: HomeSearch, Query: "Frieren Beyond Journey s End", Scope: ""}, quit: true},
		{name: "favorite 3 with accents and punctuation", home: regHome{recents: true, favs: 3, scope: 1}, keys: append(downN(3), "enter"),
			want: SearchHomeResult{Action: HomeSearch, Query: "Pokémon Aventuras Pokémon", Scope: ""}, quit: true},
		{name: "recent 1", home: regHome{continues: true, recents: true, favs: 3}, keys: append(downN(7), "enter"),
			want: SearchHomeResult{Action: HomeSearch, Query: "namakura gatana", Scope: ""}, quit: true},
		{name: "recent 2 keeps the scope", home: regHome{recents: true, scope: 2}, keys: []string{"down", "down", "enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "momotarou", Scope: "hianime"}, quit: true},
		{name: "recent after eight favorites", home: regHome{recents: true, favs: 12}, keys: append(downN(9), "enter"),
			want: SearchHomeResult{Action: HomeSearch, Query: "namakura gatana"}, quit: true},
		{name: "FAVORITES box opens the first", home: regHome{favs: 5, scope: 4}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "Frieren Beyond Journey s End", Scope: ""}, quit: true},
		{name: "FAVORITES box opens the first match", home: regHome{favs: 13, scope: 4, query: "movie"}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "Heat", Scope: ""}, quit: true},
		{name: "FAVORITES row", home: regHome{favs: 5, scope: 4}, keys: []string{"down", "down", "down", "enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "Pokémon Aventuras Pokémon", Scope: ""}, quit: true},
		{name: "FAVORITES filtered row", home: regHome{favs: 13, scope: 4, query: "tv"}, keys: []string{"down", "down", "enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "Severance", Scope: ""}, quit: true},
		{name: "FAVORITES no match stays", home: regHome{favs: 5, scope: 4, query: "zzz"}, keys: []string{"enter"}, quit: false},
		{name: "FAVORITES empty stays", home: regHome{scope: 4}, keys: []string{"enter"}, quit: false},
		// frozen: the two-letter minimum is for searching; on FAVORITES one letter opens the first match.
		{name: "FAVORITES one letter is enough", home: regHome{favs: 5, scope: 4, query: "h"}, keys: []string{"enter"},
			want: SearchHomeResult{Action: HomeSearch, Query: "Heat"}, quit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.home.build()
			quit := regPress(m, tt.keys...)
			assert.Equal(t, tt.quit, quit)
			assert.Equal(t, tt.want, m.result)
			assert.NoError(t, m.err)
		})
	}
}

func TestRegressionHomeRowKeysResults(t *testing.T) {
	regPin(t, false)
	cont := regContinue()
	full := func(scope int) regHome {
		return regHome{continues: true, recents: true, favs: 3, scope: scope}
	}
	t.Run("r from start", func(t *testing.T) {
		m := full(1).build()
		assert.True(t, regPress(m, "down", "down", "r"))
		assert.Equal(t, SearchHomeResult{Action: HomeFromStart, Entry: cont[1], HasEntry: true, Scope: "ptbr"}, m.result)
	})
	t.Run("e episodes", func(t *testing.T) {
		m := full(3).build()
		assert.True(t, regPress(m, "down", "e"))
		assert.Equal(t, SearchHomeResult{Action: HomeEpisodes, Entry: cont[0], HasEntry: true, Scope: "movietv"}, m.result)
	})
	t.Run("r and e do nothing on a favorite or a recent search", func(t *testing.T) {
		for _, steps := range []int{4, 8} {
			m := full(0).build()
			keys := make([]string, steps)
			for i := range keys {
				keys[i] = "down"
			}
			regPress(m, keys...)
			assert.False(t, regPress(m, "r"))
			assert.False(t, regPress(m, "e"))
			assert.Equal(t, SearchHomeResult{}, m.result)
		}
	})
	t.Run("in the box the letters type", func(t *testing.T) {
		m := full(0).build()
		assert.False(t, regPress(m, "r", "e", "x", "X"))
		assert.Equal(t, "rexX", m.input.Value())
		assert.Equal(t, SearchHomeResult{}, m.result)
	})
	t.Run("on a row, another letter jumps back to the box and types", func(t *testing.T) {
		m := full(0).build()
		regPress(m, "down", "q")
		assert.True(t, m.focusInput)
		assert.Equal(t, "q", m.input.Value())
	})
	t.Run("ctrl+c", func(t *testing.T) {
		m := full(0).build()
		assert.True(t, regPress(m, "ctrl+c"))
		assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
		assert.Equal(t, SearchHomeResult{}, m.result)
	})
	t.Run("esc never quits", func(t *testing.T) {
		m := full(0).build()
		assert.False(t, regPress(m, "type:abc", "esc"))
		assert.Equal(t, "", m.input.Value(), "esc clears the box")
		assert.False(t, regPress(m, "esc"))
		assert.NoError(t, m.err)
	})
}

func TestRegressionHomeRemovalsCallTheHost(t *testing.T) {
	regPin(t, false)
	down := func(n int) []string {
		keys := make([]string, n)
		for i := range keys {
			keys[i] = "down"
		}
		return keys
	}
	t.Run("x on each kind of row", func(t *testing.T) {
		calls := &regCalls{}
		m := regHome{continues: true, recents: true, favs: 3, calls: calls}.build()
		regPress(m, "down", "x") // the first episode
		regPress(m, "x")         // the next one slides under the cursor
		assert.Equal(t, []string{"delete-continue:c1", "delete-continue:c2"}, calls.log)
		assert.Len(t, m.continueAll, 1)
		regPress(m, "down", "x") // first favorite
		assert.Equal(t, "delete-favorite:frieren", calls.log[2])
		assert.Len(t, m.favoriteAll, 2)
		regPress(m, "end", "x") // last recent search
		assert.Equal(t, "delete-recent:momotarou", calls.log[3])
		assert.Len(t, m.recentAll, 1)
	})
	t.Run("a failed removal keeps the row and says so", func(t *testing.T) {
		calls := &regCalls{err: errors.New("disk full")}
		m := regHome{continues: true, recents: true, favs: 3, calls: calls}.build()
		regPress(m, "down", "x")
		assert.Equal(t, "Could not remove the episode: disk full", m.notice)
		assert.Len(t, m.continueAll, 3)
		regPress(m, "down", "down", "down", "x")
		assert.Equal(t, "Could not remove the favorite: disk full", m.notice)
		regPress(m, "end", "x")
		assert.Equal(t, "Could not remove the search: disk full", m.notice)
		assert.Len(t, m.recentAll, 2)
	})
	t.Run("without a host callback x does nothing", func(t *testing.T) {
		m := regHome{continues: true, recents: true, favs: 3, noHost: true}.build()
		regPress(m, "down", "x", "X")
		assert.Len(t, m.continueAll, 3)
		assert.Equal(t, clearNothing, m.clearing)
		regPress(m, "end", "x", "X")
		assert.Len(t, m.recentAll, 2)
		assert.Equal(t, clearNothing, m.clearing)
		regPress(m, "up", "up", "up", "up", "up", "up", "up", "up", "up", "up", "up", "up")
		regPress(m, "down", "down", "down", "down", "x", "X")
		assert.Len(t, m.favoriteAll, 3)
		assert.Equal(t, clearNothing, m.clearing)
	})
	t.Run("X asks, y clears, n and esc keep", func(t *testing.T) {
		for _, tt := range []struct {
			steps int
			text  string
			call  string
			after func(m *searchHomeModel) int
		}{
			{1, "Clear all 3 episodes from CONTINUE? y to clear, n to keep.", "clear-continue", func(m *searchHomeModel) int { return len(m.continueAll) }},
			{4, "Clear all 3 favorites from FAVORITES? y to clear, n to keep.", "clear-favorites", func(m *searchHomeModel) int { return len(m.favoriteAll) }},
			{7, "Clear all 2 searches from RECENT? y to clear, n to keep.", "clear-recent", func(m *searchHomeModel) int { return len(m.recentAll) }},
		} {
			calls := &regCalls{}
			m := regHome{continues: true, recents: true, favs: 3, calls: calls}.build()
			_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
			regPress(m, down(tt.steps)...)
			regPress(m, "X")
			assert.Contains(t, regRender(m), tt.text)
			assert.Equal(t, " !  "+tt.text, strings.TrimRight(stripANSI(m.questionView(100)), "\n"))
			regPress(m, "n")
			assert.Empty(t, calls.log)
			assert.Equal(t, 3-boolInt(tt.steps == 7), tt.after(m))
			regPress(m, "X", "esc")
			assert.Empty(t, calls.log)
			assert.NotContains(t, regRender(m), tt.text)
			regPress(m, "X", "y")
			assert.Equal(t, []string{tt.call}, calls.log)
			assert.Equal(t, 0, tt.after(m))
		}
	})
	t.Run("the questions count one in the singular", func(t *testing.T) {
		m := regHome{continues: true, recents: true, favs: 1}.build()
		regPress(m, "down", "down", "down", "down", "X")
		assert.Contains(t, m.questionView(100), "Clear all 1 favorite from FAVORITES? y to clear, n to keep.")
		m2 := regHome{recents: true}.build()
		m2.recentAll = m2.recentAll[:1]
		m2.rebuildRows()
		regPress(m2, "down", "X")
		assert.Contains(t, m2.questionView(100), "Clear all 1 search from RECENT? y to clear, n to keep.")
		m3 := regHome{continues: true}.build()
		m3.continueAll = m3.continueAll[:1]
		m3.rebuildRows()
		regPress(m3, "down", "X")
		assert.Contains(t, m3.questionView(100), "Clear all 1 episode from CONTINUE? y to clear, n to keep.")
	})
	t.Run("ctrl+c quits from the question", func(t *testing.T) {
		m := regHome{continues: true}.build()
		regPress(m, "down", "X")
		assert.True(t, regPress(m, "ctrl+c"))
		assert.ErrorIs(t, m.err, ErrSearchHomeCancelled)
	})
	t.Run("a failed clear keeps the list and says so", func(t *testing.T) {
		calls := &regCalls{err: errors.New("locked")}
		m := regHome{continues: true, recents: true, favs: 3, calls: calls}.build()
		regPress(m, "down", "X", "y")
		assert.Equal(t, "Could not clear CONTINUE: locked", m.notice)
		assert.Len(t, m.continueAll, 3)
		regPress(m, "down", "down", "down", "X", "y")
		assert.Equal(t, "Could not clear FAVORITES: locked", m.notice)
		assert.Len(t, m.favoriteAll, 3)
		regPress(m, "end", "X", "y")
		assert.Equal(t, "Could not clear RECENT: locked", m.notice)
		assert.Len(t, m.recentAll, 2)
	})
	t.Run("clearing the last list row hands the focus to the box", func(t *testing.T) {
		m := regHome{favs: 2}.build()
		regPress(m, "down", "X", "y")
		assert.True(t, m.focusInput)
		assert.Empty(t, m.favoriteAll)
		assert.Empty(t, m.rows)
	})
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestRegressionHomeFavoritesTabKeys(t *testing.T) {
	regPin(t, false)
	t.Run("the tab starts on the box with the cursor at the top", func(t *testing.T) {
		m := regHome{favs: 5, scope: 4}.build()
		assert.True(t, m.focusInput)
		assert.Equal(t, 0, m.cursor)
		assert.Len(t, m.rows, 5)
		for _, r := range m.rows {
			assert.Equal(t, homeRowFavorite, r.kind)
		}
	})
	t.Run("the rows are only the favorites that match", func(t *testing.T) {
		m := regHome{continues: true, recents: true, favs: 13, scope: 4, query: "pokemon"}.build()
		require.Len(t, m.rows, 1)
		assert.Equal(t, 2, m.rows[0].index)
		m = regHome{favs: 13, scope: 4, query: "animefire"}.build()
		keys := []string{}
		for _, r := range m.rows {
			keys = append(keys, m.favoriteAll[r.index].Key)
		}
		assert.Equal(t, []string{"frieren", "dandadan"}, keys, "the source reads AnimeFire, Animefire.io included")
		m = regHome{favs: 13, scope: 4, query: "1999 movie"}.build()
		require.Len(t, m.rows, 1)
		assert.Equal(t, "the matrix", m.favoriteAll[m.rows[0].index].Key)
		m = regHome{favs: 13, scope: 4, query: "  "}.build()
		assert.Len(t, m.rows, 13, "blanks match everything")
		m = regHome{favs: 13, scope: 4, query: "ANIME"}.build()
		assert.NotEmpty(t, m.rows)
		m = regHome{favs: 13, scope: 4, query: "é"}.build()
		assert.NotEmpty(t, m.rows, "accents fold: é finds Pokémon")
	})
	t.Run("filter changes send the highlight back to the top", func(t *testing.T) {
		m := regHome{favs: 13, scope: 4}.build()
		regPress(m, "down", "down", "down", "down")
		assert.Equal(t, 3, m.cursor)
		regPress(m, "type:a")
		assert.Equal(t, 0, m.cursor)
		assert.False(t, m.focusInput == false && len(m.rows) == 0)
	})
	t.Run("a no-op key keeps the highlight", func(t *testing.T) {
		m := regHome{favs: 13, scope: 4}.build()
		regPress(m, "down", "down", "down")
		assert.Equal(t, 2, m.cursor)
		regPress(m, "pgdown")
		assert.Equal(t, 12, m.cursor)
		regPress(m, "pgup")
		assert.Equal(t, 2, m.cursor)
		regPress(m, "end")
		assert.Equal(t, 12, m.cursor)
		regPress(m, "home")
		assert.Equal(t, 0, m.cursor)
	})
	t.Run("pgdown and pgup move ten rows", func(t *testing.T) {
		m := regHome{favs: 13, scope: 4}.build()
		regPress(m, "down", "pgdown")
		assert.Equal(t, 10, m.cursor)
		regPress(m, "pgdown")
		assert.Equal(t, 12, m.cursor, "clamped to the last row")
		regPress(m, "pgup")
		assert.Equal(t, 2, m.cursor)
		regPress(m, "pgup")
		assert.Equal(t, 0, m.cursor, "clamped to the first row")
	})
	t.Run("pgup and the like only move rows when a row has the focus", func(t *testing.T) {
		m := regHome{favs: 13, scope: 4, query: "d"}.build()
		regPress(m, "pgdown", "end", "home")
		assert.True(t, m.focusInput)
		assert.Equal(t, 0, m.cursor)
	})
	t.Run("removing the highlighted favorite", func(t *testing.T) {
		calls := &regCalls{}
		m := regHome{favs: 5, scope: 4, calls: calls}.build()
		regPress(m, "down", "end", "x")
		assert.Equal(t, []string{"delete-favorite:mushishi"}, calls.log)
		assert.Len(t, m.rows, 4)
		assert.Equal(t, 3, m.cursor, "the highlight stays on the last row")
		assert.False(t, m.focusInput)
	})
	t.Run("removing every favorite hands the box the focus", func(t *testing.T) {
		m := regHome{favs: 2, scope: 4}.build()
		regPress(m, "down", "x", "x")
		assert.True(t, m.focusInput)
		assert.Empty(t, m.rows)
	})
	t.Run("nothing is searched while on the tab", func(t *testing.T) {
		m := regHome{favs: 5, scope: 4}.build()
		asked := 0
		m.suggest = func(_ context.Context, _, _ string) ([]HomeSuggestion, error) { asked++; return nil, nil }
		m.input.SetValue("abc")
		assert.Nil(t, m.queryChanged(), "typing on FAVORITES schedules no lookup")
		assert.Equal(t, 0, asked)
		assert.Empty(t, m.suggestions)
		// On a search tab a changed query schedules the lookup.
		m.scopeIdx = 0
		assert.NotNil(t, m.scopeChanged())
	})
}

// ---- helpers' own text ------------------------------------------------------------------------

func TestRegressionSearchableTitle(t *testing.T) {
	for in, want := range map[string]string{
		"Frieren":                       "Frieren",
		"Pokémon: Aventuras Pokémon":    "Pokémon Aventuras Pokémon",
		"Re:Zero - Starting Life":       "Re Zero Starting Life",
		"  Spaced   Out  ":              "Spaced Out",
		"Fate/Zero (2011)":              "Fate Zero 2011",
		"Dr. STONE":                     "Dr STONE",
		"Mob Psycho 100 II":             "Mob Psycho 100 II",
		"!!!":                           "!!!",
		"":                              "",
		"Bad\x1b[31m\ntitle":            "Bad title",
		"東京 ミュウミュウ":                     "東京 ミュウミュウ",
		"Attack on Titan: The Final...": "Attack on Titan The Final",
	} {
		assert.Equal(t, want, searchableTitle(in), "%q", in)
	}
}

func TestRegressionRelativeDay(t *testing.T) {
	now := regNow
	for _, tt := range []struct {
		ago  time.Duration
		want string
	}{
		{-time.Hour, "today"},
		{0, "today"},
		{23 * time.Hour, "today"},
		{24 * time.Hour, "yesterday"},
		{47 * time.Hour, "yesterday"},
		{48 * time.Hour, "2 days ago"},
		{6 * 24 * time.Hour, "6 days ago"},
		{7 * 24 * time.Hour, "last week"},
		{13 * 24 * time.Hour, "last week"},
		{14 * 24 * time.Hour, "2 weeks ago"},
		{20 * 24 * time.Hour, "2 weeks ago"},
		{21 * 24 * time.Hour, "3 weeks ago"},
		{59 * 24 * time.Hour, "8 weeks ago"},
		{60 * 24 * time.Hour, "2024-01-10"},
		{400 * 24 * time.Hour, "2023-02-04"},
	} {
		assert.Equal(t, tt.want, relativeDay(now.Add(-tt.ago), now), "%v ago", tt.ago)
	}
}

func TestRegressionHomeContinueRowText(t *testing.T) {
	for _, tt := range []struct {
		e        HomeContinueEntry
		title    string
		finished bool
	}{
		{HomeContinueEntry{AnimeTitle: "Show", EpisodeTitle: "Ep"}, "Show", false},
		{HomeContinueEntry{EpisodeTitle: "Only Episode"}, "Only Episode", false},
		{HomeContinueEntry{AnimeTitle: " \t "}, "Unknown title", false},
		{HomeContinueEntry{Position: 539, Duration: 600}, "Unknown title", false},
		{HomeContinueEntry{Position: 540, Duration: 600}, "Unknown title", true},
		{HomeContinueEntry{Position: 600, Duration: 0}, "Unknown title", false},
	} {
		assert.Equal(t, tt.title, tt.e.DisplayTitle())
		assert.Equal(t, tt.finished, tt.e.Finished())
	}
	assert.Equal(t, "1 episode", countOf(1, "episode", "episodes"))
	assert.Equal(t, "0 episodes", countOf(0, "episode", "episodes"))
	assert.Equal(t, "2 episodes", countOf(2, "episode", "episodes"))
}

func TestRegressionHomeConstants(t *testing.T) {
	assert.Equal(t, 8, homeFavoritesShown)
	assert.Equal(t, 10, homePageRows)
	assert.Equal(t, 2, minHomeQuery)
	assert.Equal(t, 6, maxHomeSuggestions)
	assert.Equal(t, 180*time.Millisecond, suggestDebounce)
	assert.Equal(t, 4*time.Second, suggestTimeout)
	assert.Equal(t, "back from search home", ErrSearchHomeBack.Error())
	assert.Equal(t, "search home cancelled", ErrSearchHomeCancelled.Error())
	assert.Equal(t, HomeAction(0), HomeSearch)
	assert.Equal(t, HomeAction(1), HomeResume)
	assert.Equal(t, HomeAction(2), HomeFromStart)
	assert.Equal(t, HomeAction(3), HomeEpisodes)
}

func TestRegressionHomeViewMetadata(t *testing.T) {
	regPin(t, false)
	m := regHome{continues: true}.build()
	v := m.View()
	assert.True(t, v.AltScreen)
	assert.Equal(t, "GoAnime", v.WindowTitle)
}

func TestRegressionSearchHomeWithRunnerAnswers(t *testing.T) {
	regPin(t, false)
	t.Run("a result comes back", func(t *testing.T) {
		res, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchHomeModel)
			m.result = SearchHomeResult{Action: HomeSearch, Query: "x1", Scope: "ptbr"}
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, SearchHomeResult{Action: HomeSearch, Query: "x1", Scope: "ptbr"}, res)
	})
	t.Run("an empty search is a cancel", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) { return model, nil })
		assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	})
	t.Run("the runner failing is wrapped", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(tea.Model) (tea.Model, error) { return nil, errors.New("tty gone") })
		assert.EqualError(t, err, "run search home screen: tty gone")
	})
	t.Run("no runner", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, nil)
		assert.EqualError(t, err, "search home runner not configured")
	})
	t.Run("an error of the model wins", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchHomeModel)
			m.err = ErrSearchHomeCancelled
			m.result = SearchHomeResult{Action: HomeSearch, Query: "zz"}
			return m, nil
		})
		assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	})
}
