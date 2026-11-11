package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// ---- a fake for the lists -----------------------------------------------------------

// homeListsFake is a FavoriteLists in memory with a log of its calls and errors
// that tests inject by method name.
type homeListsFake struct {
	order []HomeList // the lists, Favorites first; Count and Titles come from works
	works map[string][]HomeFavorite
	log   []string
	fail  map[string]error
	next  int
}

func (f *homeListsFake) record(call string) { f.log = append(f.log, call) }

// calls is the log without the reads, which the screens do freely.
func (f *homeListsFake) calls() []string {
	var out []string
	for _, c := range f.log {
		if !strings.HasPrefix(c, "Lists") && !strings.HasPrefix(c, "Works") {
			out = append(out, c)
		}
	}
	return out
}

func newHomeListsFake() *homeListsFake {
	return &homeListsFake{
		order: []HomeList{{ID: FavoritesListID, Name: "Favorites", Auto: true}},
		works: map[string][]HomeFavorite{FavoritesListID: {}},
		fail:  map[string]error{},
	}
}

var hlDay = time.Date(2026, 11, 10, 18, 30, 0, 0, time.UTC)

// withList adds a user's list holding n works named "<name> 0" ...
func (f *homeListsFake) withList(id, name string, n int) *homeListsFake {
	f.order = append(f.order, HomeList{ID: id, Name: name, Created: hlDay})
	f.works[id] = hlFakeWorks(name, n)
	return f
}

func (f *homeListsFake) withFavorites(n int) *homeListsFake {
	f.works[FavoritesListID] = hlFakeWorks("Fav", n)
	return f
}

func hlFakeWorks(prefix string, n int) []HomeFavorite {
	out := make([]HomeFavorite, n)
	for i := range out {
		out[i] = HomeFavorite{Key: fmt.Sprintf("%s-%d", strings.ToLower(prefix), i), Title: fmt.Sprintf("%s %d", prefix, i), Media: "Anime"}
	}
	return out
}

func (f *homeListsFake) Lists() ([]HomeList, error) {
	f.record("Lists")
	if err := f.fail["Lists"]; err != nil {
		return nil, err
	}
	out := make([]HomeList, len(f.order))
	for i, l := range f.order {
		l.Count = len(f.works[l.ID])
		l.Titles = nil
		for _, w := range f.works[l.ID] {
			l.Titles = append(l.Titles, w.Title)
		}
		out[i] = l
	}
	return out, nil
}

func (f *homeListsFake) Works(id string) ([]HomeFavorite, error) {
	f.record("Works(" + id + ")")
	if err := f.fail["Works"]; err != nil {
		return nil, err
	}
	return slices.Clone(f.works[id]), nil
}

func (f *homeListsFake) Create(name string) (HomeList, error) {
	f.record("Create(" + name + ")")
	if err := f.fail["Create"]; err != nil {
		return HomeList{}, err
	}
	name = strings.Join(strings.Fields(name), " ")
	switch {
	case name == "":
		return HomeList{}, errors.New("give your list a name")
	case len([]rune(name)) > homeListNameLimit:
		return HomeList{}, errors.New("that name is too long")
	case strings.EqualFold(name, "favorites"):
		return HomeList{}, errors.New("that name is reserved")
	}
	for _, l := range f.order {
		if strings.EqualFold(l.Name, name) {
			return HomeList{}, errors.New("you already have a list with that name")
		}
	}
	f.next++
	l := HomeList{ID: fmt.Sprintf("new%d", f.next), Name: name, Created: hlDay}
	f.order = append(f.order, l)
	f.works[l.ID] = nil
	return l, nil
}

func (f *homeListsFake) Rename(id, name string) error {
	f.record("Rename(" + id + "," + name + ")")
	if err := f.fail["Rename"]; err != nil {
		return err
	}
	name = strings.Join(strings.Fields(name), " ")
	if id == FavoritesListID {
		return errors.New("the favorites list cannot be renamed")
	}
	if name == "" {
		return errors.New("give your list a name")
	}
	at := -1
	for i, l := range f.order {
		if l.ID != id && strings.EqualFold(l.Name, name) {
			return errors.New("you already have a list with that name")
		}
		if l.ID == id {
			at = i
		}
	}
	if at >= 0 {
		f.order[at].Name = name
	}
	return nil
}

func (f *homeListsFake) Delete(id string) error {
	f.record("Delete(" + id + ")")
	if err := f.fail["Delete"]; err != nil {
		return err
	}
	if id == FavoritesListID {
		return errors.New("the favorites list cannot be deleted")
	}
	f.order = slices.DeleteFunc(f.order, func(l HomeList) bool { return l.ID == id })
	delete(f.works, id)
	return nil
}

func (f *homeListsFake) ToggleWork(id string, w HomeFavorite) (bool, error) {
	f.record("ToggleWork(" + id + "," + w.Key + ")")
	if err := f.fail["ToggleWork"]; err != nil {
		return false, err
	}
	if !slices.ContainsFunc(f.works[id], func(x HomeFavorite) bool { return x.Key == w.Key }) {
		f.works[id] = append(f.works[id], w)
	}
	return true, nil
}

func (f *homeListsFake) RemoveWork(id, key string) error {
	f.record("RemoveWork(" + id + "," + key + ")")
	if err := f.fail["RemoveWork"]; err != nil {
		return err
	}
	f.works[id] = slices.DeleteFunc(f.works[id], func(w HomeFavorite) bool { return w.Key == key })
	return nil
}

func (f *homeListsFake) Clear(id string) error {
	f.record("Clear(" + id + ")")
	if err := f.fail["Clear"]; err != nil {
		return err
	}
	f.works[id] = nil
	return nil
}

func (f *homeListsFake) Membership(key string) map[string]bool {
	out := map[string]bool{}
	for id, ws := range f.works {
		for _, w := range ws {
			if w.Key == key {
				out[id] = true
			}
		}
	}
	return out
}

// ---- helpers --------------------------------------------------------------------------

var (
	hlEnter = press(tea.KeyEnter, "")
	hlEsc   = press(tea.KeyEscape, "")
	hlTab   = press(tea.KeyTab, "")
)

// homeListsScreen is a home screen of the given size on the FAVORITES tab, first level.
func homeListsScreen(t *testing.T, f *homeListsFake, w, h int) *searchHomeModel {
	t.Helper()
	m := newSearchHomeModel(SearchHomeOptions{Lists: f, Favorites: slices.Clone(f.works[FavoritesListID])})
	_, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	tabTo(t, m, "FAVORITES")
	return m
}

// hlRowIs puts the highlight on the first row that draws title.
func hlRowIs(t *testing.T, m *searchHomeModel, title string) {
	t.Helper()
	for i, row := range m.rows {
		var got string
		switch row.kind {
		case homeRowNewList, homeRowList:
			got = m.listItem(row, 100).title
		case homeRowFavorite:
			got = m.tabFavorites()[row.index].Title
		}
		if strings.Contains(got, title) {
			hlGoto(t, m, i)
			return
		}
	}
	t.Fatalf("no row %q in %v", title, m.rows)
}

// hlGoto moves the highlight to row with up and down, from wherever it is.
func hlGoto(t *testing.T, m *searchHomeModel, row int) {
	t.Helper()
	for i := 0; i < 2*len(m.rows)+4 && (m.focusInput || m.cursor != row); i++ {
		key := tea.KeyDown
		if !m.focusInput && m.cursor > row {
			key = tea.KeyUp
		}
		_, _ = m.Update(tea.KeyPressMsg{Code: key})
	}
	require.False(t, m.focusInput)
	require.Equal(t, row, m.cursor)
}

func hlSend(m *searchHomeModel, msgs ...tea.Msg) {
	for _, msg := range msgs {
		_, _ = m.Update(msg)
	}
}

func hlTypeName(m *searchHomeModel, text string) {
	for _, r := range text {
		_, _ = m.Update(press(r, string(r)))
	}
}

// body is the home view as plain text, the way the shell cuts it.
func hlBody(m *searchHomeModel) string {
	return ansi.Strip(m.homeView(m.shellWidth(), m.shellHeight()))
}

// hlFocusBox moves the highlight up into the filter box.
func hlFocusBox(m *searchHomeModel) {
	for i := 0; i < len(m.rows)+2 && !m.focusInput; i++ {
		_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
}

// hlType types text into the filter box, taking the focus there first, as a
// user does before typing a name to find.
func hlType(m *searchHomeModel, text string) {
	hlFocusBox(m)
	typeText(m, text)
}
