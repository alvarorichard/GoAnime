package tui

import (
	"errors"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
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

func (failingFavorites) Has(string) bool                      { return false }
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
	assert.Less(t, indexOf(view, "CONTINUE"), indexOf(view, "FAVORITOS"))
	assert.Less(t, indexOf(view, "FAVORITOS"), indexOf(view, "RECENT"))
	assert.Contains(t, view, "Frieren")
	assert.Contains(t, view, "2023 · Anime")
	assert.Contains(t, view, "A Origem")
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestNoFavoritosNoSection(t *testing.T) {
	m := newSearchHomeModel(SearchHomeOptions{})
	assert.NotContains(t, screen(m), "FAVORITOS")
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
