package tui

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- the result screen: ctrl+f -------------------------------------------------

var ctrlF = tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}

func resultsWithFavorites(t *testing.T) (*animeResultsModel, *favorites.Store) {
	t.Helper()
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	prev := newFavorites
	newFavorites = func() favoritesAPI { return store }
	t.Cleanup(func() { newFavorites = prev })
	m := newAnimeResultsModel([]*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime},
		{Name: "[PT-BR] Naruto", Source: "Animefire.io", Year: "2002", MediaType: models.MediaTypeAnime},
	})
	return m, store
}

func TestCtrlFAddsTheHighlightedWorkToFavoritos(t *testing.T) {
	m, store := resultsWithFavorites(t)

	_, cmd := m.Update(ctrlF)

	assert.Nil(t, cmd, "it stays on the screen")
	list, err := store.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Frieren", list[0].Title)
	assert.Equal(t, "frieren", list[0].Key)
	assert.Equal(t, "2023", list[0].Year)
	assert.Equal(t, "Anime", list[0].Media)
	assert.Equal(t, "AnimeFire", list[0].Source, "the chosen version is remembered")
	assert.Contains(t, m.detailsContent().title, "Frieren")
	assert.NotEqual(t, "Frieren", m.detailsContent().title, "the panel marks a favorite")
}

func TestCtrlFAgainTakesItOff(t *testing.T) {
	m, store := resultsWithFavorites(t)
	_, _ = m.Update(ctrlF)
	_, _ = m.Update(ctrlF)

	assert.False(t, store.Has("frieren"))
	assert.Equal(t, "Frieren", m.detailsContent().title)
}

func TestFavoriteIsPerWorkNotPerSource(t *testing.T) {
	m, store := resultsWithFavorites(t)
	// The English copy is grouped with the PT-BR one: one work, one favorite.
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	_, _ = m.Update(ctrlF)

	list, err := store.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "HiAnime", list[0].Source)
	assert.True(t, store.Has("frieren"))
}

func TestCtrlFWhileFilteringDoesNotFavorite(t *testing.T) {
	m, store := resultsWithFavorites(t)
	_, _ = m.Update(press('n', "n")) // typing opens the filter
	require.Equal(t, 1, int(m.results.FilterState()))

	_, _ = m.Update(ctrlF)

	assert.False(t, store.Has("naruto"))
	assert.False(t, store.Has("frieren"))
}

type failingFavorites struct{}

func (failingFavorites) Keys() map[string]bool                { return map[string]bool{} }
func (failingFavorites) Toggle(favorites.Entry) (bool, error) { return false, errors.New("disk full") }

func TestAFavoriteThatCannotBeSavedSaysWhy(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	m.favs = failingFavorites{}

	_, _ = m.Update(ctrlF)

	assert.Contains(t, m.notice, "Could not update FAVORITOS")
	assert.Contains(t, m.notice, "disk full")
}

func TestResultHintsOfferFavorite(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	var keys []string
	for _, h := range m.hints() {
		keys = append(keys, h.Key)
	}
	assert.Contains(t, keys, "ctrl+f")
}

// ---- the home screen: the FAVORITOS list ---------------------------------------

type favoritesFake struct {
	removed []string
	cleared int
	err     error
}

func (f *favoritesFake) options(rows []HomeFavorite, continueAll []HomeContinueEntry, recent []HomeRecentSearch) SearchHomeOptions {
	return SearchHomeOptions{
		Continue:  continueAll,
		Recent:    recent,
		Favorites: rows,
		DeleteFavorite: func(key string) error {
			f.removed = append(f.removed, key)
			return f.err
		},
		ClearFavorites: func() error { f.cleared++; return f.err },
	}
}

var sampleFavorites = []HomeFavorite{
	{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "Animefire.io"},
	{Key: "a origem", Title: "A Origem", Year: "2010", Media: "Movie", Source: "StartFlix"},
}

func TestFavoritosSitBetweenContinueAndRecent(t *testing.T) {
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, continueAll, recent))

	require.Len(t, m.rows, 7)
	kinds := make([]homeRowKind, len(m.rows))
	for i, r := range m.rows {
		kinds[i] = r.kind
	}
	assert.Equal(t, []homeRowKind{
		homeRowContinue, homeRowContinue, homeRowContinue,
		homeRowFavorite, homeRowFavorite,
		homeRowRecent, homeRowRecent,
	}, kinds, "the cursor walks the lists in the order they are drawn")

	view := screen(m)
	assert.Contains(t, view, "FAVORITOS")
	assert.Less(t, indexOf(view, "CONTINUE"), lastIndexOf(view, "FAVORITOS"))
	assert.Less(t, lastIndexOf(view, "FAVORITOS"), indexOf(view, "RECENT"))
	assert.Contains(t, view, "Frieren")
	assert.Contains(t, view, "2023 · Anime")
	assert.Contains(t, view, "A Origem")
}

func lastIndexOf(s, sub string) int { return strings.LastIndex(s, sub) }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestFavoritosIsOnTheMenuBeforeTheFirstFavorite(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})

	view := screen(m)
	assert.Contains(t, view, "FAVORITOS", "a new user can find it")
	assert.Contains(t, view, "Nothing here yet")
	assert.Contains(t, view, "ctrl+f", "and learns how to fill it")
	assert.Empty(t, m.rows, "an empty list has no rows to walk")
}

func TestTheEmptyFavoritosHintGoesAwayWithTheFirstOne(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: sampleFavorites[:1]})
	view := screen(m)
	assert.Contains(t, view, "FAVORITOS")
	assert.NotContains(t, view, "Nothing here yet")
}

func TestTheEmptyHintReturnsWhenTheLastFavoriteIsRemoved(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(sampleFavorites[:1], nil, nil))
	downTo(t, m, 0)
	_, _ = m.Update(press('x', "x"))
	assert.Contains(t, screen(m), "Nothing here yet")
	assert.True(t, m.focusInput)
}

func TestEnterOnAFavoriteSearchesItsTitleEverywhere(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	m.scopeIdx = 1 // PT-BR tab: a movie favorite must still be found
	downTo(t, m, 1)

	_, cmd := m.Update(press(tea.KeyEnter, ""))

	require.NotNil(t, cmd)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Equal(t, "A Origem", m.result.Query)
	assert.Equal(t, "", m.result.Scope, "every scope")
}

func TestXRemovesTheHighlightedFavorite(t *testing.T) {
	f := &favoritesFake{}
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel(f.options(sampleFavorites, continueAll, recent))
	downTo(t, m, 3)

	_, cmd := m.Update(press('x', "x"))

	assert.Nil(t, cmd)
	assert.Equal(t, []string{"frieren"}, f.removed)
	require.Len(t, m.favoriteAll, 1)
	assert.Equal(t, "A Origem", m.favoriteAll[0].Title)
	assert.Len(t, m.continueAll, 3)
	assert.Len(t, m.recentAll, 2)
}

func TestAFavoriteThatFailsToGoStays(t *testing.T) {
	f := &favoritesFake{err: errors.New("locked")}
	m := newSearchHomeModel(f.options(sampleFavorites, nil, nil))
	downTo(t, m, 0)

	_, _ = m.Update(press('x', "x"))

	assert.Len(t, m.favoriteAll, 2)
	assert.Contains(t, m.notice, "remove the favorite")
	assert.Contains(t, m.notice, "locked")
}

func TestXWithoutARemoverKeepsTheFavorite(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: sampleFavorites})
	downTo(t, m, 0)
	_, _ = m.Update(press('x', "x"))
	assert.Len(t, m.favoriteAll, 2)
}

func TestShiftXAsksThenYClearsOnlyFavoritos(t *testing.T) {
	f := &favoritesFake{}
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel(f.options(sampleFavorites, continueAll, recent))
	downTo(t, m, 3)

	_, _ = m.Update(press('X', "X"))
	assert.Equal(t, clearFavoriteRows, m.clearing)
	assert.Contains(t, screen(m), "Clear all 2 favorites from FAVORITOS?")
	assert.Equal(t, 0, f.cleared, "nothing is cleared before the answer")

	_, _ = m.Update(press('y', "y"))
	assert.Equal(t, 1, f.cleared)
	assert.Empty(t, m.favoriteAll)
	assert.Len(t, m.continueAll, 3)
	assert.Len(t, m.recentAll, 2)
}

func TestNKeepsTheFavorites(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(sampleFavorites, nil, nil))
	downTo(t, m, 0)
	_, _ = m.Update(press('X', "X"))
	_, _ = m.Update(press('n', "n"))
	assert.Equal(t, 0, f.cleared)
	assert.Len(t, m.favoriteAll, 2)
}

func TestRAndEDoNothingOnAFavoriteRow(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: sampleFavorites})
	downTo(t, m, 0)
	_, cmd := m.Update(press('r', "r"))
	assert.Nil(t, cmd)
	_, cmd = m.Update(press('e', "e"))
	assert.Nil(t, cmd)
	assert.Len(t, m.favoriteAll, 2)
}

// ---- feedback on the row, at any width -------------------------------------------

func TestTheStarShowsOnTheRowAtEveryWidth(t *testing.T) {
	for _, width := range []int{60, 80, 140} {
		m, _ := resultsWithFavorites(t)
		_, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: 30})
		assert.NotContains(t, screen2(m), "★", "width %d: nothing is a favorite yet", width)

		_, _ = m.Update(ctrlF)
		view := screen2(m)
		assert.Contains(t, view, "★", "width %d: the row tells the user it worked", width)

		_, _ = m.Update(ctrlF)
		assert.NotContains(t, screen2(m), "★", "width %d: and that it is off again", width)
	}
}

func screen2(m *animeResultsModel) string { return ansi.Strip(m.View().Content) }

func TestOnlyTheFavoriteRowGetsTheStar(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	_, _ = m.Update(ctrlF) // Frieren

	count := 0
	for line := range strings.SplitSeq(screen2(m), "\n") {
		if strings.Contains(line, "★") && !strings.Contains(line, "Details") {
			count++
			assert.Contains(t, line, "Frieren")
		}
	}
	assert.GreaterOrEqual(t, count, 1)
	assert.NotContains(t, rowOf(m, "Naruto"), "★")
}

func rowOf(m *animeResultsModel, title string) string {
	for line := range strings.SplitSeq(screen2(m), "\n") {
		if strings.Contains(line, title) {
			return line
		}
	}
	return ""
}

func TestFavoritesFromEarlierSessionsShowOnTheirRows(t *testing.T) {
	store := favorites.Open(filepath.Join(t.TempDir(), "favorites.json"))
	_, err := store.Toggle(favorites.Entry{Key: "naruto", Title: "Naruto"})
	require.NoError(t, err)
	prev := newFavorites
	newFavorites = func() favoritesAPI { return store }
	t.Cleanup(func() { newFavorites = prev })

	m := newAnimeResultsModel([]*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io"},
		{Name: "[PT-BR] Naruto", Source: "Animefire.io"},
	})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})

	assert.Contains(t, rowOf(m, "Naruto"), "★")
	assert.NotContains(t, rowOf(m, "Frieren"), "★")
}

// countingFavorites counts how often the store is read.
type countingFavorites struct {
	keys  map[string]bool
	reads int
}

func (c *countingFavorites) Keys() map[string]bool { c.reads++; return c.keys }
func (c *countingFavorites) Toggle(e favorites.Entry) (bool, error) {
	if c.keys[e.Key] {
		delete(c.keys, e.Key)
		return false, nil
	}
	c.keys[e.Key] = true
	return true, nil
}

func TestRenderingNeverReadsTheStore(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	c := &countingFavorites{keys: map[string]bool{}}
	m.favs = c
	for range 20 {
		_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		_ = m.View()
		_ = m.detailsContent()
	}
	assert.Zero(t, c.reads, "the keys are read once, when the screen opens")
}

func TestTheErrorNoticeGoesAwayWhenAToggleWorksAgain(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	good := &countingFavorites{keys: map[string]bool{}}
	m.favs = failingFavorites{}
	_, _ = m.Update(ctrlF)
	require.NotEmpty(t, m.notice)

	m.favs = good
	_, _ = m.Update(ctrlF)

	assert.Empty(t, m.notice)
	assert.NotContains(t, screen2(m), "Could not update")
	assert.Contains(t, screen2(m), "★")
}

// ---- more than the home screen draws ---------------------------------------------

func manyFavorites(n int) []HomeFavorite {
	rows := make([]HomeFavorite, n)
	for i := range rows {
		rows[i] = HomeFavorite{Key: "k" + strconv.Itoa(i), Title: "Work " + strconv.Itoa(i), Media: "Anime"}
	}
	return rows
}

func TestOnlyTheFirstFavoritesAreDrawnAndTheRestAreCounted(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: manyFavorites(13)})

	assert.Len(t, m.rows, homeFavoritesShown, "the cursor only reaches what is drawn")
	view := screen(m)
	assert.Contains(t, view, "Work 0")
	assert.Contains(t, view, "Work 7")
	assert.NotContains(t, view, "Work 8")
	assert.Contains(t, view, "+5 more")
}

func TestRemovingAFavoriteSlidesTheNextOneIn(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(manyFavorites(10), nil, nil))
	downTo(t, m, 0)

	_, _ = m.Update(press('x', "x"))

	assert.Equal(t, []string{"k0"}, f.removed)
	assert.Len(t, m.rows, homeFavoritesShown, "the hidden ones backfill")
	view := screen(m)
	assert.Contains(t, view, "Work 8", "it was behind the +N line")
	assert.Contains(t, view, "Work 1")
	assert.NotContains(t, view, "Work 0")
	assert.Contains(t, view, "+1 more", "9 favorites are left, 8 drawn")
}

func TestClearCountsEveryFavoriteIncludingHiddenOnes(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(manyFavorites(13), nil, nil))
	downTo(t, m, 0)

	_, _ = m.Update(press('X', "X"))

	assert.Contains(t, screen(m), "Clear all 13 favorites from FAVORITOS?")
}

func TestTheScreenDoesNotScribbleOnTheCallersFavorites(t *testing.T) {
	rows := manyFavorites(3)
	before := slices.Clone(rows)
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(rows, nil, nil))
	downTo(t, m, 0)
	_, _ = m.Update(press('x', "x"))

	assert.Equal(t, before, rows, "removing a row must not shift the caller's slice")
}

// ---- plain words for beginners ---------------------------------------------------

func TestTheDetailsSayInWordsWhatCtrlFDoes(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	assert.Contains(t, screen2(m), "ctrl+f adds it to FAVORITOS")

	_, _ = m.Update(ctrlF)
	view := screen2(m)
	assert.Contains(t, view, "ctrl+f takes it off FAVORITOS")
	assert.NotContains(t, view, "ctrl+f adds it to FAVORITOS")
}

func TestTheFooterNamesTheActionItWillDo(t *testing.T) {
	m, _ := resultsWithFavorites(t)
	hint := func() string {
		for _, h := range m.hints() {
			if h.Key == "ctrl+f" {
				return h.Desc
			}
		}
		return ""
	}
	assert.Equal(t, "favorite", hint())
	_, _ = m.Update(ctrlF)
	assert.Equal(t, "unfavorite", hint())
}

func TestThePlainPanelSaysItToo(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	m, _ := resultsWithFavorites(t)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	assert.Contains(t, screen2(m), "ctrl+f adds it to FAVORITOS")
	_, _ = m.Update(ctrlF)
	assert.Contains(t, screen2(m), "* Frieren", "a star drawn as text, for screen readers")
}

// ---- the FAVORITOS tab -------------------------------------------------------------

func tabTo(t *testing.T, m *searchHomeModel, name string) {
	t.Helper()
	for range len(m.scopes) {
		if m.scopes[m.scopeIdx].Name == name {
			return
		}
		_, _ = m.Update(press(tea.KeyTab, ""))
	}
	t.Fatalf("no tab %s", name)
}

func TestFavoritosIsATabBesideMoviesTV(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	names := make([]string, len(m.scopes))
	for i, s := range m.scopes {
		names[i] = s.Name
	}
	assert.Equal(t, []string{"ALL", "PT-BR", "ENGLISH", "MOVIES/TV", "FAVORITOS"}, names)
	assert.Contains(t, screen(m), "FAVORITOS", "on the tab row, where a new user looks")
}

func TestTheSearchPromptKeepsOnlyTheSearchTabs(t *testing.T) {
	names := make([]string, len(searchScopes))
	for i, s := range searchScopes {
		names[i] = s.Name
	}
	assert.Equal(t, []string{"ALL", "PT-BR", "ENGLISH", "MOVIES/TV"}, names, "a bare search prompt has no favorites tab")
}

func TestTheTabListsEveryFavoriteWithoutTheOtherLists(t *testing.T) {
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(13), continueAll, recent))
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	tabTo(t, m, "FAVORITOS")

	view := screen(m)
	assert.Contains(t, view, "Work 0")
	assert.Contains(t, view, "Work 12", "none is hidden behind a +N line: the list scrolls")
	assert.Len(t, m.rows, 13)
	assert.NotContains(t, view, "CONTINUE")
	assert.NotContains(t, view, "RECENT")
	assert.NotContains(t, view, "Mukashi Banashi")
	assert.True(t, m.focusInput)
	assert.Contains(t, m.input.Placeholder, "filter your favorites")
}

func TestLeavingTheTabBringsTheListsBack(t *testing.T) {
	continueAll, recent := homeTestEntries()
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, continueAll, recent))
	tabTo(t, m, "FAVORITOS")
	_, _ = m.Update(press(tea.KeyTab, "")) // wraps to ALL

	assert.Equal(t, "ALL", m.scopes[m.scopeIdx].Name)
	view := screen(m)
	assert.Contains(t, view, "CONTINUE")
	assert.Contains(t, view, "RECENT")
	assert.Len(t, m.rows, 3+2+2)
	assert.Contains(t, m.input.Placeholder, "title, or")
}

func TestTypingOnTheTabFiltersTheFavorites(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	tabTo(t, m, "FAVORITOS")

	typeText(m, "orig")

	require.Len(t, m.rows, 1)
	view := screen(m)
	assert.Contains(t, view, "A Origem")
	assert.NotContains(t, view, "Frieren")
	assert.Contains(t, view, "1 of 2")

	typeText(m, "zzz")
	assert.Contains(t, screen(m), "No favorite matches")
	assert.Contains(t, screen(m), "Delete a letter")
}

func TestTheFilterMatchesYearTypeAndSourceToo(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	tabTo(t, m, "FAVORITOS")
	typeText(m, "movie")
	require.Len(t, m.rows, 1)
	assert.Equal(t, 1, m.rows[0].index, "A Origem is the movie")
}

func TestEnterInTheBoxOpensTheFirstMatchAndNeverSearchesTheTab(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	tabTo(t, m, "FAVORITOS")
	typeText(m, "orig")

	_, cmd := m.Update(press(tea.KeyEnter, ""))

	require.NotNil(t, cmd)
	assert.Equal(t, HomeSearch, m.result.Action)
	assert.Equal(t, "A Origem", m.result.Query)
	assert.Equal(t, "", m.result.Scope, "the tab is not a search scope")
}

func TestEnterOnAnEmptyTabDoesNothing(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	tabTo(t, m, "FAVORITOS")
	_, cmd := m.Update(press(tea.KeyEnter, ""))
	assert.Nil(t, cmd)
	assert.Contains(t, screen(m), "You have no favorites yet")
	assert.Contains(t, screen(m), "ctrl+f")
}

func TestEnterOnAHighlightedFavoriteOnTheTab(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	tabTo(t, m, "FAVORITOS")
	downTo(t, m, 1)
	_, cmd := m.Update(press(tea.KeyEnter, ""))
	require.NotNil(t, cmd)
	assert.Equal(t, "A Origem", m.result.Query)
	assert.Equal(t, "", m.result.Scope)
}

func TestTheTabScrollsAndKeepsTheCursorInView(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(60), nil, nil))
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	tabTo(t, m, "FAVORITOS")
	downTo(t, m, 40)

	view := screen(m)
	assert.Contains(t, view, "Work 40", "the highlighted row is on screen")
	assert.NotContains(t, view, "Work 0 ", "the top scrolled away")
	assert.NotContains(t, view, "Work 1 ", "and the rows after it")
	assert.Contains(t, view, "more above")
	assert.Contains(t, view, "more below")
	assert.LessOrEqual(t, strings.Count(view, "\n")+1, 24, "it fits the terminal")
}

func TestXAndShiftXWorkOnTheTab(t *testing.T) {
	f := &favoritesFake{}
	m := newSearchHomeModel(f.options(manyFavorites(12), nil, nil))
	tabTo(t, m, "FAVORITOS")
	typeText(m, "work 1") // Work 1, Work 10, Work 11
	require.Len(t, m.rows, 3)
	downTo(t, m, 1)

	_, _ = m.Update(press('x', "x"))

	assert.Equal(t, []string{"k10"}, f.removed, "the filtered row, not the one at that position overall")
	assert.Len(t, m.favoriteAll, 11)
	assert.Len(t, m.rows, 2)

	_, _ = m.Update(press('X', "X"))
	assert.Contains(t, screen(m), "Clear all 11 favorites from FAVORITOS?")
}

func TestTheTabHintsSayOpen(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	enter := func() string {
		for _, h := range m.hints() {
			if h.Key == "enter" {
				return h.Desc
			}
		}
		return ""
	}
	assert.Equal(t, "search", enter())
	tabTo(t, m, "FAVORITOS")
	assert.Equal(t, "open", enter())
}

func TestNoSuggestionsAreAskedOnTheTab(t *testing.T) {
	asked := 0
	m := newSearchHomeModel(SearchHomeOptions{
		Favorites: sampleFavorites,
		Suggest: func(context.Context, string, string) ([]HomeSuggestion, error) {
			asked++
			return nil, nil
		},
	})
	m.debounce = 0
	tabTo(t, m, "FAVORITOS")
	typeText(m, "frieren")
	assert.Nil(t, m.queryChanged())
	assert.Zero(t, asked)
	assert.Empty(t, m.suggestions)
}

// ---- review follow-ups -------------------------------------------------------------

func TestTheFilterIgnoresAccentsCaseAndExtraSpaces(t *testing.T) {
	rows := []HomeFavorite{
		{Key: "a", Title: "Pokémon", Year: "1997", Media: "Anime"},
		{Key: "b", Title: "ÉCOLE de Magie", Year: "2020", Media: "Anime"},
		{Key: "c", Title: "Dragon Ball Z", Year: "1989", Media: "Anime"},
		{Key: "d", Title: "Ação Total", Year: "2001", Media: "Movie"},
	}
	for query, want := range map[string]string{
		"pokemon":      "a",
		"POKÉMON":      "a",
		"ecole":        "b",
		"dragon  ball": "c",
		"ball dragon":  "c",
		"acao":         "d",
		"ação":         "d",
	} {
		m := newSearchHomeModel(SearchHomeOptions{Favorites: rows})
		tabTo(t, m, "FAVORITOS")
		m.input.SetValue(query)
		_ = m.queryChanged()
		require.Len(t, m.rows, 1, "query %q", query)
		assert.Equal(t, want, rows[m.rows[0].index].Key, "query %q", query)
	}
}

func TestEscOnAHighlightedFavoriteKeepsTheHighlight(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(10), nil, nil))
	tabTo(t, m, "FAVORITOS")
	downTo(t, m, 5)

	_, _ = m.Update(press(tea.KeyEscape, ""))

	assert.False(t, m.focusInput)
	assert.Equal(t, 5, m.cursor, "esc with an empty box must not throw the highlight to the top")
}

func TestTabbingFromAHighlightedFavoriteLandsOnTheBox(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(10), nil, nil))
	tabTo(t, m, "FAVORITOS")
	downTo(t, m, 5)

	_, _ = m.Update(press(tea.KeyTab, ""))

	assert.True(t, m.focusInput)
	assert.Equal(t, 0, m.cursor)
}

func TestTheHighlightedFavoriteIsAlwaysOnScreen(t *testing.T) {
	for _, size := range []struct{ w, h int }{{80, 24}, {60, 20}, {50, 24}, {80, 15}, {100, 40}} {
		for _, question := range []bool{false, true} {
			m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(40), nil, nil))
			_, _ = m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
			tabTo(t, m, "FAVORITOS")
			downTo(t, m, 25)
			if question {
				_, _ = m.Update(press('X', "X"))
			}
			body := ansi.Strip(m.homeView(m.shellWidth(), m.shellHeight()))
			assert.Contains(t, body, "Work 25", "%dx%d question=%v", size.w, size.h, question)
			_, h := m.shell.ContentSize()
			assert.LessOrEqual(t, strings.Count(body, "\n")+1, h, "%dx%d question=%v: the body fits", size.w, size.h, question)
		}
	}
}

func (m *searchHomeModel) shellWidth() int  { w, _ := m.shell.ContentSize(); return w }
func (m *searchHomeModel) shellHeight() int { _, h := m.shell.ContentSize(); return h }

func TestOpeningOnTheFavoritesTabByIndex(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: sampleFavorites, ScopeIndex: 4})
	assert.Equal(t, "FAVORITOS", m.scopes[m.scopeIdx].Name)
	assert.Len(t, m.rows, 2)
	assert.Contains(t, m.input.Placeholder, "filter your favorites")
}

func TestTheHelpMentionsTheFavoritesTab(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m.showHelp = true
	assert.Contains(t, screen(m), "FAVORITOS")
}

// ---- the beginner round: footer, paging, queries, notices ---------------------------

func hintsByKey(m *searchHomeModel) map[string]Hint {
	out := map[string]Hint{}
	for _, h := range m.hints() {
		out[h.Key] = h
	}
	return out
}

func TestRemoveStaysInTheFooterWhenARowIsHighlighted(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(sampleFavorites, nil, nil))
	assert.NotContains(t, hintsByKey(m), "x", "nothing to remove while the box has the focus")
	downTo(t, m, 0)
	h, ok := hintsByKey(m)["x"]
	require.True(t, ok)
	assert.True(t, h.Essential, "a narrow terminal must not hide how to remove")
}

func TestPagingAndEndsMoveThroughALongList(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(60), nil, nil))
	tabTo(t, m, "FAVORITOS")
	downTo(t, m, 0)

	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, homePageRows, m.cursor)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 2*homePageRows, m.cursor)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, homePageRows, m.cursor)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	assert.Equal(t, 59, m.cursor)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 59, m.cursor, "clamped at the end")
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	assert.Equal(t, 0, m.cursor)
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 0, m.cursor, "clamped at the top, and still on the list")
	assert.False(t, m.focusInput)
}

func TestHomeAndEndStillMoveTheCaretInTheBox(t *testing.T) {
	m := newSearchHomeModel((&favoritesFake{}).options(manyFavorites(5), nil, nil))
	typeText(m, "abc")
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	assert.True(t, m.focusInput)
	assert.Equal(t, 0, m.input.Position())
}

func TestAFavoriteIsSearchedByItsWordsNotItsPunctuation(t *testing.T) {
	rows := []HomeFavorite{
		{Key: "a", Title: "Pokémon: Aventuras"},
		{Key: "b", Title: "Re:Zero - Starting Life (2016)"},
		{Key: "c", Title: "!!!"},
	}
	for want, i := range map[string]int{"Pokémon Aventuras": 0, "Re Zero Starting Life 2016": 1, "!!!": 2} {
		m := newSearchHomeModel(SearchHomeOptions{Favorites: rows})
		downTo(t, m, i)
		_, cmd := m.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		assert.Equal(t, want, m.result.Query)
	}
}

func TestThePlainTabRowMarksTheActiveTabInText(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })
	m := newSearchHomeModel(SearchHomeOptions{})
	assert.Contains(t, ansi.Strip(m.scopeTabs()), "[ALL]")
	tabTo(t, m, "FAVORITOS")
	row := ansi.Strip(m.scopeTabs())
	assert.Contains(t, row, "[FAVORITOS]")
	assert.NotContains(t, row, "[ALL]")
}

func TestTheRichTabRowHasNoBrackets(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	assert.NotContains(t, ansi.Strip(m.scopeTabs()), "[")
}

func TestAnOpeningNoticeShowsThenGoesAtTheFirstKey(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Notice: "Your favorites file was damaged\nsecond line"})
	assert.Contains(t, screen(m), "Your favorites file was damaged second line")
	_, _ = m.Update(press('a', "a"))
	assert.NotContains(t, screen(m), "damaged")
}

func TestTheMoreLineSaysHowToSeeThemAll(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{Favorites: manyFavorites(12)})
	assert.Contains(t, screen(m), "press tab until FAVORITOS to see them all")
}
