package tui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restoreFixture is a home with every list filled: three CONTINUE rows, two
// RECENT rows and three favorites.
func restoreFixture() SearchHomeOptions {
	continueAll, recent := homeTestEntries()
	return SearchHomeOptions{
		Continue: continueAll,
		Recent:   recent,
		Favorites: []HomeFavorite{
			{Key: "fav-frieren", Title: "Sousou no Frieren", Year: "2023", Media: "Anime", Source: "Animefire.io"},
			{Key: "fav-origem", Title: "A Origem", Year: "2010", Media: "Movie", Source: "StartFlix"},
			{Key: "fav-one-piece", Title: "One Piece", Year: "1999", Media: "Anime", Source: "HiAnime"},
		},
	}
}

// slugOpts is what the way back used to pass: a hyphenated query and the ALL tab.
func slugOpts(opts SearchHomeOptions, slug string) SearchHomeOptions {
	opts.Query = slug
	opts.ScopeIndex = 0
	return opts
}

func typeKeys(m *searchHomeModel, text string) {
	for _, r := range text {
		_, _ = m.Update(press(r, string(r)))
	}
}

func goToTab(t *testing.T, m *searchHomeModel, name string) {
	t.Helper()
	for range len(m.scopes) {
		if tabName(m.scopes[m.scopeIdx].Value) == name {
			return
		}
		_, _ = m.Update(press(tea.KeyTab, ""))
	}
	require.Equal(t, name, tabName(m.scopes[m.scopeIdx].Value))
}

// enter presses enter and returns the decision with where the screen was.
func enter(t *testing.T, m *searchHomeModel) SearchHomeResult {
	t.Helper()
	_, cmd := m.Update(press(tea.KeyEnter, ""))
	require.NotNil(t, cmd, "enter must leave the screen")
	return m.resultWithRestore()
}

// highlighted is the key of the highlighted row, "" when the box has the focus.
func highlighted(m *searchHomeModel) string {
	row, ok := m.currentRow()
	if !ok {
		return ""
	}
	return m.rowKey(row)
}

// ---- the value ------------------------------------------------------------------

func TestHomeRestoreZeroValue(t *testing.T) {
	t.Parallel()
	assert.True(t, HomeRestore{}.IsZero())
	assert.False(t, HomeRestore{Tab: TabAll}.IsZero())
	assert.False(t, HomeRestore{RowKey: "x"}.IsZero())
	assert.False(t, HomeRestore{Query: "x"}.IsZero())
}

func TestHomeRestoreTabsAreTheScreensTabsInOrder(t *testing.T) {
	t.Parallel()
	want := []string{TabAll, TabPTBR, TabEnglish, TabMoviesTV, TabFavorites}
	require.Len(t, homeScopes, len(want))
	for i, name := range want {
		idx, ok := HomeRestore{Tab: name}.tabIndex()
		require.True(t, ok, name)
		assert.Equal(t, i, idx, name)
		assert.Equal(t, name, tabName(homeScopes[i].Value), name)
	}
}

func TestHomeRestoreFavoritesIsNeverASearchScope(t *testing.T) {
	t.Parallel()
	for _, s := range searchScopes {
		assert.NotEqual(t, TabFavorites, tabName(s.Value))
		assert.NotEqual(t, favoritesScope, s.Value)
	}
	// ...and a restore never turns it into one: the result's Scope stays empty.
	m := newSearchHomeModel(restoreFixture())
	goToTab(t, m, TabFavorites)
	res := enter(t, m)
	assert.Equal(t, TabFavorites, res.Restore.Tab)
	assert.Empty(t, res.Scope)
	assert.NotEqual(t, favoritesScope, res.Scope)
}

func TestHomeRestoreUnknownTabIsIgnored(t *testing.T) {
	t.Parallel()
	for _, tab := range []string{"", "nope", "ALL", "Favorites", " all"} {
		r := HomeRestore{Tab: tab, RowKey: "favorite:fav-origem", Query: "zzz"}
		_, ok := r.tabIndex()
		assert.False(t, ok, "%q", tab)
		opts := SearchHomeOptions{Query: "keep-me", ScopeIndex: 2, Restore: r}
		assert.Equal(t, opts, opts.withRestore(), "%q", tab)
		m := newSearchHomeModel(restoreFixtureWith(opts))
		assert.Equal(t, 2, m.scopeIdx)
		assert.Equal(t, "keep-me", m.input.Value())
		assert.True(t, m.focusInput)
	}
}

func restoreFixtureWith(o SearchHomeOptions) SearchHomeOptions {
	f := restoreFixture()
	f.Query, f.ScopeIndex, f.Restore = o.Query, o.ScopeIndex, o.Restore
	return f
}

func TestHomeRestoreOverridesTabAndBox(t *testing.T) {
	t.Parallel()
	opts := SearchHomeOptions{Query: "slug-here", ScopeIndex: 0, Restore: HomeRestore{Tab: TabEnglish, Query: "typed"}}
	got := opts.withRestore()
	assert.Equal(t, 2, got.ScopeIndex)
	assert.Equal(t, "typed", got.Query)

	// A restore from a row clears the slug the caller passed.
	opts.Restore = HomeRestore{Tab: TabPTBR, RowKey: "recent:x"}
	got = opts.withRestore()
	assert.Equal(t, 1, got.ScopeIndex)
	assert.Empty(t, got.Query)
}

// ---- producing it -----------------------------------------------------------------

func TestRestoreFromTheBoxOnEachSearchTab(t *testing.T) {
	t.Parallel()
	for _, tab := range []string{TabAll, TabPTBR, TabEnglish, TabMoviesTV} {
		t.Run(tab, func(t *testing.T) {
			t.Parallel()
			m := newSearchHomeModel(restoreFixture())
			goToTab(t, m, tab)
			typeKeys(m, "one piece")
			res := enter(t, m)
			assert.Equal(t, HomeSearch, res.Action)
			assert.Equal(t, HomeRestore{Tab: tab, Query: "one piece"}, res.Restore)
			assert.Equal(t, "one piece", res.Query, "the search itself is unchanged")
		})
	}
}

func TestRestoreFromTheBoxTakesTheSearchedQueryNotThePartialOne(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(restoreFixture())
	m.input.SetValue("  frieren  ")
	res := enter(t, m)
	assert.Equal(t, "frieren", res.Restore.Query, "what was searched, trimmed")
}

func TestRestoreFromEachSearchTabsRows(t *testing.T) {
	t.Parallel()
	for _, tab := range []string{TabAll, TabPTBR, TabEnglish, TabMoviesTV} {
		t.Run(tab, func(t *testing.T) {
			t.Parallel()
			// recent row: enter on it
			m := newSearchHomeModel(restoreFixture())
			goToTab(t, m, tab)
			downTo(t, m, 6) // three CONTINUE and three favorite rows come first
			row, _ := m.currentRow()
			require.Equal(t, homeRowRecent, row.kind)
			res := enter(t, m)
			assert.Equal(t, HomeRestore{Tab: tab, RowKey: "recent:namakura gatana"}, res.Restore)

			// continue row
			m = newSearchHomeModel(restoreFixture())
			goToTab(t, m, tab)
			downTo(t, m, 1)
			res = enter(t, m)
			assert.Equal(t, HomeResume, res.Action)
			assert.Equal(t, HomeRestore{Tab: tab, RowKey: "continue:t:kuro nyago\x00animefire.io"}, res.Restore)
		})
	}
}

func TestRestoreFromTheAllTabsFavoritesSection(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(restoreFixture())
	downTo(t, m, 3+1) // second favorite, after the three CONTINUE rows
	res := enter(t, m)
	assert.Equal(t, HomeSearch, res.Action)
	assert.Equal(t, "A Origem", res.Query)
	assert.Equal(t, HomeRestore{Tab: TabAll, RowKey: "favorite:fav-origem"}, res.Restore)
	assert.Empty(t, res.Restore.Query, "a row never refills the box with its slug")
}

func TestRestoreFromContinueRowActions(t *testing.T) {
	t.Parallel()
	for key, action := range map[string]HomeAction{"r": HomeFromStart, "e": HomeEpisodes} {
		m := newSearchHomeModel(restoreFixture())
		goToTab(t, m, TabPTBR)
		downTo(t, m, 0)
		_, cmd := m.Update(press(rune(key[0]), key))
		require.NotNil(t, cmd, key)
		res := m.resultWithRestore()
		assert.Equal(t, action, res.Action, key)
		assert.Equal(t, HomeRestore{Tab: TabPTBR, RowKey: "continue:t:mukashi banashi\x00animefire.io"}, res.Restore, key)
	}
}

func TestRestoreFromTheFavoritesTab(t *testing.T) {
	t.Parallel()
	t.Run("a highlighted row", func(t *testing.T) {
		t.Parallel()
		m := newSearchHomeModel(restoreFixture())
		goToTab(t, m, TabFavorites)
		downTo(t, m, 2)
		res := enter(t, m)
		assert.Equal(t, HomeRestore{Tab: TabFavorites, RowKey: "favorite:fav-one-piece"}, res.Restore)
		assert.Equal(t, "One Piece", res.Query)
		assert.Empty(t, res.Scope)
	})
	t.Run("enter from the box opens the first match", func(t *testing.T) {
		t.Parallel()
		m := newSearchHomeModel(restoreFixture())
		goToTab(t, m, TabFavorites)
		res := enter(t, m)
		assert.Equal(t, HomeRestore{Tab: TabFavorites, RowKey: "favorite:fav-frieren"}, res.Restore)
	})
	t.Run("enter from the box with a filter", func(t *testing.T) {
		t.Parallel()
		m := newSearchHomeModel(restoreFixture())
		goToTab(t, m, TabFavorites)
		typeKeys(m, "orig")
		res := enter(t, m)
		assert.Equal(t, HomeRestore{Tab: TabFavorites, RowKey: "favorite:fav-origem"}, res.Restore)
		assert.Equal(t, "A Origem", res.Query)
	})
	t.Run("the box with nothing to open does not leave", func(t *testing.T) {
		t.Parallel()
		m := newSearchHomeModel(restoreFixture())
		goToTab(t, m, TabFavorites)
		typeKeys(m, "zzzz")
		_, cmd := m.Update(press(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
		assert.True(t, m.restore.IsZero())
	})
}

func TestRestoreIsNotRecordedWithoutADecision(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(restoreFixture())
	typeKeys(m, "x") // too short: enter does nothing
	_, cmd := m.Update(press(tea.KeyEnter, ""))
	assert.Nil(t, cmd)
	assert.True(t, m.restore.IsZero())
	assert.True(t, m.resultWithRestore().Restore.IsZero())
}

// The decision itself is the one it always was: the restore rides beside it.
func TestRestoreLeavesTheModelsResultUntouched(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(restoreFixture())
	goToTab(t, m, TabPTBR)
	typeKeys(m, "frieren")
	_ = enter(t, m)
	assert.Equal(t, SearchHomeResult{Action: HomeSearch, Query: "frieren", Scope: "ptbr"}, m.result)
}

// ---- consuming it: the real model, there and back --------------------------------

// there presses enter on a home built from opts and returns the options the
// way back would build, with the slug and the ALL tab the old way back passed.
func there(t *testing.T, opts SearchHomeOptions, drive func(*searchHomeModel)) (SearchHomeOptions, SearchHomeResult) {
	t.Helper()
	m := newSearchHomeModel(opts)
	drive(m)
	res := enter(t, m)
	back := slugOpts(opts, "one-piece")
	back.Restore = res.Restore
	return back, res
}

func TestHomeComesBackWhereItWas(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		drive     func(t *testing.T, m *searchHomeModel)
		tab       string
		row       string // highlighted row key; "" for the box
		box       string
		wantInBox bool
	}{
		{"box on ALL", func(t *testing.T, m *searchHomeModel) { typeKeys(m, "one piece") }, TabAll, "", "one piece", true},
		{"box on PT-BR", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabPTBR); typeKeys(m, "one piece") }, TabPTBR, "", "one piece", true},
		{"box on ENGLISH", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabEnglish); typeKeys(m, "one piece") }, TabEnglish, "", "one piece", true},
		{"box on MOVIES/TV", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabMoviesTV); typeKeys(m, "one piece") }, TabMoviesTV, "", "one piece", true},
		{"favorite on ALL", func(t *testing.T, m *searchHomeModel) { downTo(t, m, 4) }, TabAll, "favorite:fav-origem", "", false},
		{"recent on ENGLISH", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabEnglish); downTo(t, m, 7) }, TabEnglish, "recent:momotarou", "", false},
		{"continue on MOVIES/TV", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabMoviesTV); downTo(t, m, 2) }, TabMoviesTV, "continue:t:saru to kani no gassen\x00", "", false},
		{"favorites tab row", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabFavorites); downTo(t, m, 1) }, TabFavorites, "favorite:fav-origem", "", false},
		{"favorites tab box", func(t *testing.T, m *searchHomeModel) { goToTab(t, m, TabFavorites) }, TabFavorites, "favorite:fav-frieren", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			back, _ := there(t, restoreFixture(), func(m *searchHomeModel) { tc.drive(t, m) })
			m := newSearchHomeModel(back)
			assert.Equal(t, tc.tab, tabName(m.scopes[m.scopeIdx].Value), "tab")
			assert.Equal(t, tc.row, highlighted(m), "row")
			assert.Equal(t, tc.row == "", m.focusInput, "focus")
			assert.Equal(t, tc.row == "", m.input.Focused(), "the box takes keys only when it has the focus")
			assert.Equal(t, tc.box, m.input.Value(), "box")
			if tc.row != "" {
				assert.NotContains(t, m.input.Value(), "-", "no slug in the box when the search came from a row")
			}
			// what the user sees is the tab they left and, if a row, that row marked
			view := screen(m)
			for _, s := range homeScopes {
				assert.Contains(t, view, s.Name)
			}
		})
	}
}

func TestHomeComesBackToTheFavoritesTabWithEveryFavoriteListed(t *testing.T) {
	t.Parallel()
	back, _ := there(t, restoreFixture(), func(m *searchHomeModel) { goToTab(t, m, TabFavorites); downTo(t, m, 2) })
	m := newSearchHomeModel(back)
	require.True(t, m.onFavorites())
	assert.Len(t, m.rows, 3)
	assert.Equal(t, 2, m.cursor)
	assert.Equal(t, favoritesPlaceholder, m.input.Placeholder, "the favorites hint, not the search one")
	view := screen(m)
	assert.Contains(t, view, "One Piece")
	assert.Contains(t, view, "A Origem")
}

func TestHomeComesBackTwiceTheSame(t *testing.T) {
	t.Parallel()
	opts := restoreFixture()
	back, _ := there(t, opts, func(m *searchHomeModel) { goToTab(t, m, TabPTBR); downTo(t, m, 4+0) })
	// the user presses enter again on the restored row: same place again
	m := newSearchHomeModel(back)
	res := enter(t, m)
	assert.Equal(t, back.Restore, res.Restore)
}

func TestRestoredHomeAllowsTheUsualKeys(t *testing.T) {
	t.Parallel()
	back, _ := there(t, restoreFixture(), func(m *searchHomeModel) { goToTab(t, m, TabPTBR); downTo(t, m, 3) })
	m := newSearchHomeModel(back)
	before := m.cursor
	_, _ = m.Update(press(tea.KeyDown, ""))
	assert.Equal(t, before+1, m.cursor)
	_, _ = m.Update(press(tea.KeyUp, ""))
	_, _ = m.Update(press(tea.KeyUp, ""))
	assert.Equal(t, before-1, m.cursor)
	// up from the first row returns to the box and typing works
	m = newSearchHomeModel(back)
	for i := 0; i < len(m.rows)+1 && !m.focusInput; i++ {
		_, _ = m.Update(press(tea.KeyUp, ""))
	}
	require.True(t, m.focusInput)
	typeKeys(m, "ab")
	assert.Equal(t, "ab", m.input.Value())
}

// ---- when the row is not there anymore ----------------------------------------------

func TestRestoreWithAMissingRowKeepsTheTabAndFocusesTheBox(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"favorite:gone", "recent:gone", "continue:t:gone\x00", "garbage"} {
		opts := restoreFixture()
		opts.Restore = HomeRestore{Tab: TabPTBR, RowKey: key}
		m := newSearchHomeModel(opts)
		assert.Equal(t, 1, m.scopeIdx, key)
		assert.True(t, m.focusInput, key)
		assert.Equal(t, 0, m.cursor, key)
		assert.Empty(t, m.input.Value(), key)
	}
}

func TestRestoreFavoriteBeyondWhatTheAllTabDrawsFallsBackToTheBox(t *testing.T) {
	t.Parallel()
	opts := restoreFixture()
	for i := range 12 {
		opts.Favorites = append(opts.Favorites, HomeFavorite{Key: fmt.Sprintf("extra-%d", i), Title: fmt.Sprintf("Extra %d", i)})
	}
	last := "favorite:extra-11"
	opts.Restore = HomeRestore{Tab: TabAll, RowKey: last}
	m := newSearchHomeModel(opts)
	assert.True(t, m.focusInput, "not drawn on ALL")
	opts.Restore = HomeRestore{Tab: TabFavorites, RowKey: last}
	m = newSearchHomeModel(opts)
	assert.Equal(t, last, highlighted(m), "the FAVORITES tab draws them all")
}

func TestRestoreFindsARecentRowThatMovedDown(t *testing.T) {
	t.Parallel()
	opts := restoreFixture()
	opts.Restore = HomeRestore{Tab: TabAll, RowKey: "recent:momotarou"}
	// the search itself put a new row on top of RECENT
	opts.Recent = append([]HomeRecentSearch{{Query: "brand new", Count: 1}}, opts.Recent...)
	m := newSearchHomeModel(opts)
	assert.Equal(t, "recent:momotarou", highlighted(m))
}

func TestRestoreFindsAContinueRowWhoseEpisodeAdvanced(t *testing.T) {
	t.Parallel()
	opts := restoreFixture()
	opts.Restore = HomeRestore{Tab: TabAll, RowKey: "continue:" + continueRowKey(opts.Continue[1])}
	opts.Continue = append([]HomeContinueEntry(nil), opts.Continue...)
	opts.Continue[1].Key = "a-newer-episode-key"
	opts.Continue[1].EpisodeNum = 9
	m := newSearchHomeModel(opts)
	assert.Equal(t, 1, m.cursor)
	assert.False(t, m.focusInput)
}

func TestRestoreRecentKeyIgnoresCaseAndAccents(t *testing.T) {
	t.Parallel()
	opts := SearchHomeOptions{
		Recent:  []HomeRecentSearch{{Query: "Pokémon", Count: 2}},
		Restore: HomeRestore{Tab: TabAll, RowKey: "recent:pokemon"},
	}
	m := newSearchHomeModel(opts)
	assert.False(t, m.focusInput)
}

func TestRestoreRowKeysAreDistinctAcrossLists(t *testing.T) {
	t.Parallel()
	// the same text in three lists names three different rows
	opts := SearchHomeOptions{
		Continue:  []HomeContinueEntry{{Key: "x", AnimeTitle: "same"}},
		Recent:    []HomeRecentSearch{{Query: "same"}},
		Favorites: []HomeFavorite{{Key: "same", Title: "same"}},
	}
	m := newSearchHomeModel(opts)
	seen := map[string]bool{}
	for _, row := range m.rows {
		k := m.rowKey(row)
		assert.False(t, seen[k], k)
		seen[k] = true
	}
	assert.Len(t, seen, 3)
}

func TestContinueRowKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b HomeContinueEntry
		same bool
	}{
		{"next episode of the same show", HomeContinueEntry{Key: "u1:ep1", AnimeTitle: "Naruto", Source: "X"}, HomeContinueEntry{Key: "u1:ep2", AnimeTitle: "Naruto", Source: "X"}, true},
		{"case and spacing of the title", HomeContinueEntry{AnimeTitle: "Naruto"}, HomeContinueEntry{AnimeTitle: "  naruto "}, true},
		{"another source is another row", HomeContinueEntry{AnimeTitle: "Naruto", Source: "X"}, HomeContinueEntry{AnimeTitle: "Naruto", Source: "Y"}, false},
		{"another show", HomeContinueEntry{AnimeTitle: "Naruto"}, HomeContinueEntry{AnimeTitle: "Bleach"}, false},
		{"no title groups by address without the episode", HomeContinueEntry{Key: "https://s/x:ep1"}, HomeContinueEntry{Key: "https://s/x:ep22"}, true},
		{"no title, the episode suffix right at the start", HomeContinueEntry{Key: ":ep1"}, HomeContinueEntry{Key: ":ep2"}, true},
		{"no title, address case", HomeContinueEntry{Key: "HTTPS://S/X:ep1"}, HomeContinueEntry{Key: "https://s/x:ep2"}, true},
		{"no title, different address", HomeContinueEntry{Key: "https://s/x:ep1"}, HomeContinueEntry{Key: "https://s/y:ep1"}, false},
		{"no title, :ep that is not a number stays", HomeContinueEntry{Key: "https://s/x:epic"}, HomeContinueEntry{Key: "https://s/x"}, false},
		{"no title, bare :ep stays", HomeContinueEntry{Key: "https://s/x:ep"}, HomeContinueEntry{Key: "https://s/x"}, false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.same, continueRowKey(tc.a) == continueRowKey(tc.b), tc.name)
	}
}

func TestFavoriteRowKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "k", favoriteRowKey(HomeFavorite{Key: "k", Title: "T"}))
	assert.Equal(t, "title:one piece", favoriteRowKey(HomeFavorite{Title: " One  Piece "}))
	assert.NotEqual(t, favoriteRowKey(HomeFavorite{Title: "a"}), favoriteRowKey(HomeFavorite{Title: "b"}))
}

func TestIsDigits(t *testing.T) {
	t.Parallel()
	assert.True(t, isDigits("0"))
	assert.True(t, isDigits("0123456789"))
	assert.False(t, isDigits(""))
	assert.False(t, isDigits("1a"))
	assert.False(t, isDigits("/"), "just below 0")
	assert.False(t, isDigits(":"), "just above 9")
	assert.False(t, isDigits("٣"), "non-ASCII digits are not episode numbers")
}

// ---- without a restore nothing changes ---------------------------------------------

func TestNoRestoreIsByteIdenticalToToday(t *testing.T) {
	t.Parallel()
	cases := []SearchHomeOptions{
		{},
		{Query: "naruto-shippuden", ScopeIndex: 2},
		{Query: "x", ScopeIndex: 4},
		{Query: "x", ScopeIndex: 99},
		{Query: "x", ScopeIndex: -1},
	}
	for i, base := range cases {
		withFixture := restoreFixture()
		withFixture.Query, withFixture.ScopeIndex = base.Query, base.ScopeIndex
		a := newSearchHomeModel(withFixture)
		withFixture.Restore = HomeRestore{}
		b := newSearchHomeModel(withFixture)
		assert.Equal(t, a.View().Content, b.View().Content, "case %d", i)
		assert.Equal(t, a.scopeIdx, b.scopeIdx)
		assert.Equal(t, a.cursor, b.cursor)
		assert.Equal(t, a.focusInput, b.focusInput)
		assert.Equal(t, a.input.Value(), b.input.Value())
		assert.Equal(t, base.Query, a.input.Value(), "the passed query is still what the box shows")
	}
}

func TestNoRestoreKeepsTheOldWayBack(t *testing.T) {
	t.Parallel()
	opts := slugOpts(restoreFixture(), "one-piece")
	opts.ScopeIndex = 2
	m := newSearchHomeModel(opts)
	assert.Equal(t, 2, m.scopeIdx)
	assert.Equal(t, "one-piece", m.input.Value())
	assert.True(t, m.focusInput)
}

func TestResultWithoutADecisionCarriesNoRestore(t *testing.T) {
	t.Parallel()
	_, err := searchHomeWithRunner(restoreFixture(), func(m tea.Model) (tea.Model, error) { return m, nil })
	assert.ErrorIs(t, err, ErrSearchHomeCancelled)
}

func TestSearchHomeWithRunnerCarriesTheRestore(t *testing.T) {
	t.Parallel()
	res, err := searchHomeWithRunner(restoreFixture(), func(m tea.Model) (tea.Model, error) {
		hm := m.(*searchHomeModel)
		goToTab(t, hm, TabEnglish)
		downTo(t, hm, 7)
		_, _ = hm.Update(press(tea.KeyEnter, ""))
		return hm, nil
	})
	require.NoError(t, err)
	assert.Equal(t, HomeRestore{Tab: TabEnglish, RowKey: "recent:momotarou"}, res.Restore)
	assert.Equal(t, "momotarou", res.Query)
	assert.Equal(t, "hianime", res.Scope)
}

// a time-dependent fixture must not leak into keys
func TestRestoreKeysDoNotDependOnTime(t *testing.T) {
	t.Parallel()
	e := HomeContinueEntry{Key: "k", AnimeTitle: "A", UpdatedAt: time.Now()}
	f := e
	f.UpdatedAt = time.Time{}
	f.Position = 99
	assert.Equal(t, continueRowKey(e), continueRowKey(f))
}

// The box restores what was searched: a highlighted suggestion is what was
// searched, not the partial text typed before it.
func TestRestoreFromTheBoxTakesThePickedSuggestion(t *testing.T) {
	t.Parallel()
	m := newSearchHomeModel(restoreFixture())
	m.input.SetValue("naru")
	m.suggestions = []HomeSuggestion{{Title: "Naruto: Shippuuden"}}
	m.sugCursor = 0
	res := enter(t, m)
	assert.Equal(t, "Naruto: Shippuuden", res.Query)
	assert.Equal(t, HomeRestore{Tab: TabAll, Query: "Naruto: Shippuuden"}, res.Restore)
}
