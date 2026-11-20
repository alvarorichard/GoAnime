package tui

// Regression net for the RESULTS screen as it is TODAY (before the ctrl+l
// picker and the "lists" wording): the rows, the star, the details panel with
// its title, fields, lines and tip, the footer and the help overlay, ctrl+f,
// at four terminal sizes in rich and in plain mode.
//
// The screens are frozen as golden files (testdata/regression/results_*.golden,
// see regression_helpers_test.go); the text of the panel, the footer and the
// notices is also asserted inline, exactly.
//
// Where today's behavior is arguably wrong the assertion still freezes it and
// carries a "frozen:" comment. The next steps may change a frozen assertion
// ONLY on purpose.

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/favorites"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regFavAPI is the favorites the result screen talks to, in memory.
type regFavAPI struct {
	keys    map[string]bool
	err     error
	toggles []favorites.Entry
}

func (f *regFavAPI) Keys() map[string]bool { return maps.Clone(f.keys) }

func (f *regFavAPI) Toggle(e favorites.Entry) (bool, error) {
	f.toggles = append(f.toggles, e)
	if f.err != nil {
		return false, f.err
	}
	if f.keys[e.Key] {
		delete(f.keys, e.Key)
		return false, nil
	}
	f.keys[e.Key] = true
	return true, nil
}

func regAnimes() []*models.Anime {
	return []*models.Anime{
		{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime, Audio: "Legendado"},
		{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime, Audio: "Sub"},
		{Name: "Heat", Source: "StartFlix", Year: "1995", MediaType: models.MediaTypeMovie},
		{Name: "[PT-BR] Dark (Dublado)", Source: "Pobreflix", Year: "2017", MediaType: models.MediaTypeTV, Audio: "Dublado"},
		{Name: "Mushishi", Source: "Goyabu"},
		{Name: "Naruto", Source: "Animefire.io", Year: "2002", MediaType: models.MediaTypeAnime, Quality: "1080p"},
	}
}

// regUseFavorites points the result screen at api (nil: no favorites) until
// the test ends.
func regUseFavorites(t *testing.T, api favoritesAPI) {
	t.Helper()
	prev := newFavorites
	newFavorites = func() favoritesAPI { return api }
	t.Cleanup(func() { newFavorites = prev })
}

type regResScene struct {
	name  string
	fav   string // "ok" (default), "fail" (every toggle fails), "nil" (no favorites)
	have  []string
	query string
	keys  []string
	after func(t *testing.T, m *animeResultsModel)
}

func (sc regResScene) build(t *testing.T, size regSize) (*animeResultsModel, *regFavAPI) {
	t.Helper()
	api := &regFavAPI{keys: map[string]bool{}}
	for _, k := range sc.have {
		api.keys[k] = true
	}
	switch sc.fav {
	case "fail":
		api.err = errors.New("disk full")
		regUseFavorites(t, api)
	case "nil":
		regUseFavorites(t, nil)
		api = nil
	default:
		regUseFavorites(t, api)
	}
	m := newAnimeResultsModelWithQuery(regAnimes(), sc.query)
	_, _ = m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
	regPress(m, sc.keys...)
	if sc.after != nil {
		sc.after(t, m)
	}
	return m, api
}

func regResultsGolden(t *testing.T, file string, scenes []regResScene) {
	t.Helper()
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		for _, sc := range scenes {
			for _, size := range regSizes {
				m, _ := sc.build(t, size)
				regGolden(t, file, fmt.Sprintf("%s %s %s", regMode(plain), size, sc.name), regFrame(t, m.View().Content, size))
			}
		}
	}
}

func TestRegressionResultsScreens(t *testing.T) {
	regResultsGolden(t, "results_screens", []regResScene{
		{name: "fresh", query: "frieren"},
		{name: "fresh without a query"},
		{name: "second row, a movie", keys: []string{"down"}},
		{name: "third row, TV with an audio word", keys: []string{"down", "down"}},
		{name: "fourth row, no year and no type", keys: []string{"down", "down", "down"}},
		{name: "last row", keys: []string{"end"}},
		{name: "version right", keys: []string{"right"}},
		{name: "version right twice wraps", keys: []string{"right", "right"}},
		{name: "version left wraps", keys: []string{"left"}},
		{name: "single version ignores right", keys: []string{"down", "right"}},
		{name: "favorite", keys: []string{"ctrl+f"}},
		{name: "favorite then second row", keys: []string{"ctrl+f", "down"}},
		{name: "favorite on the second version", keys: []string{"right", "ctrl+f"}},
		{name: "favorite then off", keys: []string{"ctrl+f", "ctrl+f"}},
		{name: "already a favorite on open", have: []string{"frieren", "heat"}},
		{name: "already a favorite, second row", have: []string{"frieren", "heat"}, keys: []string{"down"}},
		{name: "no favorites available", fav: "nil", query: "frieren"},
		{name: "no favorites available, ctrl+f", fav: "nil", keys: []string{"ctrl+f"}},
		{name: "favorite cannot be saved", fav: "fail", keys: []string{"ctrl+f"}},
		{name: "favorite cannot be saved, moving on", fav: "fail", keys: []string{"ctrl+f", "down"}},
		{name: "help", keys: []string{"?"}},
		{name: "help with a favorite", have: []string{"frieren"}, keys: []string{"?"}},
		{name: "help closed again", keys: []string{"?", "?"}},
		{name: "filter being typed", keys: []string{"type:na"}},
		{name: "filter applied", keys: []string{"down"}, after: func(_ *testing.T, m *animeResultsModel) { m.results.SetFilterText("hea"); m.reservePlainPanel() }},
		{name: "filter with no match", after: func(_ *testing.T, m *animeResultsModel) { m.results.SetFilterText("zzzz"); m.reservePlainPanel() }},
		{name: "filter applied on a favorite", have: []string{"heat"}, after: func(_ *testing.T, m *animeResultsModel) { m.results.SetFilterText("hea"); m.reservePlainPanel() }},
	})
}

func TestRegressionResultsPendingNotice(t *testing.T) {
	NoticeOnNextResults("")
	t.Cleanup(func() { NoticeOnNextResults("") })
	notices := map[string]string{
		"short": "That title has no episodes yet.",
		"long":  "  Some   very long\tnotice that goes on and on about what happened and what the user could try next, which wraps onto a second line and is then cut\nbeyond that, because a notice never takes more than two lines of the screen.  ",
	}
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		regUseFavorites(t, &regFavAPI{keys: map[string]bool{}})
		for _, name := range []string{"short", "long"} {
			for _, size := range regSizes {
				NoticeOnNextResults(notices[name])
				m := newAnimeResultsModelWithQuery(regAnimes(), "x")
				_, _ = m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
				regGolden(t, "results_notice", fmt.Sprintf("%s %s %s", regMode(plain), size, name), regFrame(t, m.View().Content, size))
			}
		}
	}
	// The notice is taken by the screen that opens next: once.
	NoticeOnNextResults("once")
	assert.Equal(t, "once", newAnimeResultsModel(regAnimes()).notice)
	assert.Equal(t, "", newAnimeResultsModel(regAnimes()).notice)
	NoticeOnNextResults(" a \n b \x1b[31m ")
	assert.Equal(t, "a b", takeResultsNotice(), "a notice is one clean line")
	assert.Equal(t, "", takeResultsNotice())
}

// ---- the footer -------------------------------------------------------------------------------

func TestRegressionResultsHints(t *testing.T) {
	scene := func(sc regResScene) string {
		m, _ := sc.build(t, regSize{100, 30})
		return regHintLine(m.hints())
	}
	for _, tt := range []struct {
		plain bool
		sc    regResScene
		want  string
	}{
		{false, regResScene{}, "enter:open! | ←→:version | ctrl+f:favorite | ↑↓:move | esc:edit search! | ?:help | q:quit"},
		{true, regResScene{}, "enter:open! | <->:version | ctrl+f:favorite | up/dn:move | esc:edit search! | ?:help | q:quit"},
		{false, regResScene{have: []string{"frieren"}}, "enter:open! | ←→:version | ctrl+f:unfavorite | ↑↓:move | esc:edit search! | ?:help | q:quit"},
		{true, regResScene{have: []string{"frieren"}}, "enter:open! | <->:version | ctrl+f:unfavorite | up/dn:move | esc:edit search! | ?:help | q:quit"},
		{false, regResScene{have: []string{"frieren"}, keys: []string{"down"}}, "enter:open! | ←→:version | ctrl+f:favorite | ↑↓:move | esc:edit search! | ?:help | q:quit"},
		{false, regResScene{keys: []string{"ctrl+f"}}, "enter:open! | ←→:version | ctrl+f:unfavorite | ↑↓:move | esc:edit search! | ?:help | q:quit"},
		// frozen: with no favorites available the footer still offers ctrl+f.
		{false, regResScene{fav: "nil"}, "enter:open! | ←→:version | ctrl+f:favorite | ↑↓:move | esc:edit search! | ?:help | q:quit"},
		{false, regResScene{keys: []string{"?"}}, "?:close! | esc:close!"},
		{true, regResScene{keys: []string{"?"}}, "?:close! | esc:close!"},
		{false, regResScene{keys: []string{"type:na"}}, "↑↓:move | enter:open! | esc:clear filter!"},
		{true, regResScene{keys: []string{"type:na"}}, "up/dn:move | enter:open! | esc:clear filter!"},
	} {
		regPin(t, tt.plain)
		assert.Equal(t, tt.want, scene(tt.sc), "%s %v", regMode(tt.plain), tt.sc)
	}
	regPin(t, false)
	m, _ := regResScene{after: func(_ *testing.T, m *animeResultsModel) { m.results.SetFilterText("hea") }}.build(t, regSize{100, 30})
	assert.Equal(t, "↑↓:move | enter:open! | esc:clear filter!", regHintLine(m.hints()))
}

func TestRegressionResultsFooterDropsTheNonEssentialFirst(t *testing.T) {
	regPin(t, false)
	for _, tt := range []struct {
		w    int
		want string
	}{
		{140, "enter  open    ←→  version    ctrl+f  favorite    ↑↓  move    esc  edit search    ?  help    q  quit"},
		{101, "enter  open    ←→  version    ctrl+f  favorite    ↑↓  move    esc  edit search    ?  help    q  quit"},
		{100, "enter  open    ←→  version    ctrl+f  favorite    ↑↓  move    esc  edit search    ?  help"},
		{90, "enter  open    ←→  version    ctrl+f  favorite    ↑↓  move    esc  edit search    ?  help"},
		{80, "enter  open    ←→  version    ctrl+f  favorite    ↑↓  move    esc  edit search"},
		{60, "enter  open    ←→  version    esc  edit search"},
		{40, "enter  open    esc  edit search"},
		{20, "enter  open"},
	} {
		m, _ := regResScene{}.build(t, regSize{tt.w, 30})
		lines := strings.Split(regRenderRes(m), "\n")
		assert.Equal(t, " "+tt.want, strings.TrimRight(lines[len(lines)-1], " "), "width %d", tt.w)
	}
}

func regRenderRes(m *animeResultsModel) string {
	lines := strings.Split(stripANSI(m.View().Content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}

// ---- the details panel -----------------------------------------------------------------------------

func TestRegressionResultsDetailsContent(t *testing.T) {
	t.Run("rich", func(t *testing.T) {
		regPin(t, false)
		m, _ := regResScene{}.build(t, regSize{140, 40})
		c := m.detailsContent()
		assert.Equal(t, "Details", c.label)
		assert.Equal(t, "詳細", c.gloss)
		assert.Equal(t, "Frieren", c.title)
		assert.Equal(t, []detailField{{"Year", "2023"}, {"Type", "Anime"}}, c.fields)
		assert.Equal(t, []string{
			"VERSIONS  2 · ←→ to choose",
			"",
			"▶ AnimeFire  PT-BR · Legendado",
			"  HiAnime  English · Sub",
			"",
			"enter opens it on AnimeFire",
		}, c.lines)
		assert.Equal(t, "ctrl+f adds it to FAVORITES", c.tip)
		assert.Equal(t, "", c.command)
		assert.Nil(t, c.progress)
		assert.Empty(t, c.keys)

		regPress(m, "right")
		c = m.detailsContent()
		assert.Equal(t, []string{
			"VERSIONS  2 · ←→ to choose",
			"",
			"  AnimeFire  PT-BR · Legendado",
			"▶ HiAnime  English · Sub",
			"",
			"enter opens it on HiAnime",
		}, c.lines)

		regPress(m, "ctrl+f")
		c = m.detailsContent()
		assert.Equal(t, "★ Frieren", c.title)
		assert.Equal(t, "ctrl+f takes it off FAVORITES", c.tip)
	})
	t.Run("plain", func(t *testing.T) {
		regPin(t, true)
		m, _ := regResScene{have: []string{"frieren"}}.build(t, regSize{140, 40})
		c := m.detailsContent()
		assert.Equal(t, "* Frieren", c.title)
		assert.Equal(t, []string{
			"VERSIONS  2 · left/right to choose",
			"",
			"> AnimeFire  PT-BR · Legendado",
			"  HiAnime  English · Sub",
			"",
			"enter opens it on AnimeFire",
		}, c.lines)
		assert.Equal(t, "ctrl+f takes it off FAVORITES", c.tip)
	})
	t.Run("one version, no year, no type, no favorites", func(t *testing.T) {
		regPin(t, false)
		m, _ := regResScene{fav: "nil", keys: []string{"down", "down", "down"}}.build(t, regSize{140, 40})
		c := m.detailsContent()
		assert.Equal(t, "Mushishi", c.title)
		assert.Empty(t, c.fields)
		assert.Equal(t, []string{"VERSIONS  1", "", "▶ Goyabu  PT-BR", "", "enter opens it on Goyabu"}, c.lines)
		assert.Equal(t, "", c.tip, "no tip when favorites are not available")
	})
	t.Run("a movie and a TV row", func(t *testing.T) {
		regPin(t, false)
		m, _ := regResScene{keys: []string{"down"}}.build(t, regSize{140, 40})
		c := m.detailsContent()
		assert.Equal(t, "Heat", c.title)
		assert.Equal(t, []detailField{{"Year", "1995"}, {"Type", "Movie"}}, c.fields)
		assert.Equal(t, []string{"VERSIONS  1", "", "▶ StartFlix  PT-BR", "", "enter opens it on StartFlix"}, c.lines)
		regPress(m, "down")
		c = m.detailsContent()
		assert.Equal(t, "Dark", c.title)
		assert.Equal(t, []detailField{{"Year", "2017"}, {"Type", "TV"}}, c.fields)
		assert.Equal(t, []string{"VERSIONS  1", "", "▶ Pobreflix  PT-BR · Dublado", "", "enter opens it on Pobreflix"}, c.lines)
	})
	t.Run("no result", func(t *testing.T) {
		regPin(t, false)
		regUseFavorites(t, &regFavAPI{keys: map[string]bool{}})
		m := &animeResultsModel{}
		m.results = list.New(nil, rowDelegate{theme: &m.theme}, 10, 10)
		c := m.detailsContent()
		assert.Equal(t, &panelContent{label: "Details", gloss: "詳細", title: "No result selected"}, c)
	})
}

func TestRegressionResultsTipIsTheFirstToGoOnAShortScreen(t *testing.T) {
	var sb strings.Builder
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		for h := 7; h <= 44; h++ {
			for _, favorite := range []bool{false, true} {
				sc := regResScene{}
				if favorite {
					sc.have = []string{"frieren"}
				}
				m, _ := sc.build(t, regSize{140, h})
				text := regRenderRes(m)
				fmt.Fprintf(&sb, "%s h=%d favorite=%v panel=%v tip=%v versions=%v enter-line=%v lines=%d\n", regMode(plain), h, favorite,
					strings.Contains(text, "DETAILS") || strings.Contains(text, "Details"), strings.Contains(text, "ctrl+f adds it to FAVORITES") || strings.Contains(text, "ctrl+f takes it off FAVORITES"),
					strings.Contains(text, "VERSIONS"), strings.Contains(text, "enter opens it on"), len(strings.Split(text, "\n")))
			}
		}
	}
	regGolden(t, "results_panel_heights", "tip and panel by height", strings.TrimSuffix(sb.String(), "\n"))
}

func TestRegressionResultsWideOrNarrow(t *testing.T) {
	regPin(t, false)
	for _, tt := range []struct {
		w    int
		wide bool
	}{{140, true}, {101, true}, {100, true}, {99, false}, {80, false}, {60, false}} {
		m, _ := regResScene{}.build(t, regSize{tt.w, 40})
		text := regRenderRes(m)
		assert.Equal(t, tt.wide, strings.Contains(text, "DETAILS"), "width %d", tt.w)
		assert.Equal(t, tt.wide, strings.Contains(text, "enter opens it on"), "width %d", tt.w)
	}
	assert.Equal(t, 100, wideResultsBreakpoint)
	assert.Equal(t, 60, resultsListWidth(100))
	assert.Equal(t, 84, resultsListWidth(140))
	assert.Equal(t, 2, resultsGutter)
	regPin(t, true)
	for _, w := range []int{140, 100, 99, 60} {
		m, _ := regResScene{}.build(t, regSize{w, 40})
		text := regRenderRes(m)
		assert.Contains(t, text, "enter opens it on AnimeFire", "plain shows the details as text below the list at width %d", w)
		lines := strings.Split(text, "\n")
		listAt, detailsAt := -1, -1
		for i, l := range lines {
			if strings.Contains(l, "01  Frieren") && listAt < 0 {
				listAt = i
			}
			if strings.Contains(l, "VERSIONS") && detailsAt < 0 {
				detailsAt = i
			}
		}
		assert.Greater(t, detailsAt, listAt, "below the list")
	}
}

// ---- rows and the star ---------------------------------------------------------------------------

func TestRegressionResultsRowsAndStar(t *testing.T) {
	regPin(t, false)
	groups := groupAnimeResults(regAnimes())
	require.Len(t, groups, 5)
	favs := map[string]bool{"frieren": true, "dark": true}
	rows := make([]row, len(groups))
	for i := range groups {
		rows[i] = animeGroupItem{group: &groups[i], pos: i, number: fmt.Sprintf("%02d", i+1), favs: favs}.row()
	}
	assert.Equal(t, row{number: "01", title: "Frieren", desc: "2023 · Anime · 2 versions", tags: []string{"★", "PT-BR", "EN"}}, rows[0])
	assert.Equal(t, row{number: "02", title: "Heat", desc: "1995 · Movie · StartFlix", tags: []string{"PT-BR"}}, rows[1])
	assert.Equal(t, row{number: "03", title: "Dark", desc: "2017 · TV · Pobreflix", tags: []string{"★", "PT-BR"}}, rows[2])
	assert.Equal(t, row{number: "04", title: "Mushishi", desc: "Goyabu", tags: []string{"PT-BR"}}, rows[3])
	assert.Equal(t, row{number: "05", title: "Naruto", desc: "2002 · Anime · AnimeFire", tags: []string{"PT-BR"}}, rows[4])

	regPin(t, true)
	plainRow := animeGroupItem{group: &groups[0], favs: favs}.row()
	assert.Equal(t, []string{"*", "PT-BR", "EN"}, plainRow.tags)
	noFavs := animeGroupItem{group: &groups[0]}.row()
	assert.Equal(t, []string{"PT-BR", "EN"}, noFavs.tags, "a screen without favorites draws no star")
	assert.Equal(t, row{title: "Unknown title", desc: "Information unavailable"}, animeGroupItem{}.row())
}

func TestRegressionResultsStarFollowsTheToggleAtOnce(t *testing.T) {
	regPin(t, false)
	m, api := regResScene{}.build(t, regSize{100, 30})
	view := func() string { return regRenderRes(m) }
	assert.NotContains(t, view(), "★")
	regPress(m, "ctrl+f")
	assert.Contains(t, view(), "Frieren  ★   PT-BR   EN")
	assert.Contains(t, view(), "★ Frieren")
	assert.Contains(t, view(), "ctrl+f takes it off FAVORITES")
	regPress(m, "ctrl+f")
	assert.NotContains(t, view(), "★")
	assert.Contains(t, view(), "ctrl+f adds it to FAVORITES")
	assert.Len(t, api.toggles, 2)
}

// ---- ctrl+f ---------------------------------------------------------------------------------------

func TestRegressionResultsCtrlF(t *testing.T) {
	regPin(t, false)

	t.Run("toggles the highlighted work with its fields", func(t *testing.T) {
		m, api := regResScene{keys: []string{"ctrl+f", "down", "ctrl+f", "right", "right", "end", "ctrl+f"}}.build(t, regSize{100, 30})
		require.Len(t, api.toggles, 3)
		assert.Equal(t, favorites.Entry{Key: "frieren", Title: "Frieren", Year: "2023", Media: "Anime", Source: "AnimeFire"}, api.toggles[0])
		assert.Equal(t, favorites.Entry{Key: "heat", Title: "Heat", Year: "1995", Media: "Movie", Source: "StartFlix"}, api.toggles[1])
		assert.Equal(t, favorites.Entry{Key: "naruto", Title: "Naruto", Year: "2002", Media: "Anime", Source: "AnimeFire"}, api.toggles[2])
		assert.Equal(t, map[string]bool{"frieren": true, "heat": true, "naruto": true}, m.favKeys)
	})
	t.Run("the chosen version is the one remembered", func(t *testing.T) {
		_, api := regResScene{keys: []string{"right", "ctrl+f"}}.build(t, regSize{100, 30})
		require.Len(t, api.toggles, 1)
		assert.Equal(t, "HiAnime", api.toggles[0].Source)
		assert.Equal(t, "frieren", api.toggles[0].Key, "one favorite per work, not per source")
	})
	t.Run("the key is the cleaned title", func(t *testing.T) {
		_, api := regResScene{keys: []string{"down", "down", "ctrl+f"}}.build(t, regSize{100, 30})
		require.Len(t, api.toggles, 1)
		assert.Equal(t, favorites.Entry{Key: "dark", Title: "Dark", Year: "2017", Media: "TV", Source: "Pobreflix"}, api.toggles[0])
	})
	t.Run("a row with no year and no type", func(t *testing.T) {
		_, api := regResScene{keys: []string{"down", "down", "down", "ctrl+f"}}.build(t, regSize{100, 30})
		require.Len(t, api.toggles, 1)
		assert.Equal(t, favorites.Entry{Key: "mushishi", Title: "Mushishi", Source: "Goyabu"}, api.toggles[0])
	})
	t.Run("it stays on the screen", func(t *testing.T) {
		m, _ := regResScene{}.build(t, regSize{100, 30})
		assert.False(t, regPress(m, "ctrl+f"))
		assert.Nil(t, m.selected)
		assert.NoError(t, m.err)
	})
	t.Run("filter being typed: ctrl+f is ignored", func(t *testing.T) {
		m, api := regResScene{keys: []string{"type:na", "ctrl+f"}}.build(t, regSize{100, 30})
		assert.Empty(t, api.toggles)
		assert.Equal(t, list.Filtering, m.results.FilterState())
		assert.Equal(t, "na", m.results.FilterInput.Value())
	})
	t.Run("filter applied: ctrl+f works on the highlighted match", func(t *testing.T) {
		m, api := regResScene{after: func(_ *testing.T, m *animeResultsModel) { m.results.SetFilterText("hea") }}.build(t, regSize{100, 30})
		require.Equal(t, list.FilterApplied, m.results.FilterState())
		regPress(m, "ctrl+f")
		require.Len(t, api.toggles, 1)
		assert.Equal(t, "heat", api.toggles[0].Key)
	})
	t.Run("no favorites available: nothing happens", func(t *testing.T) {
		m, _ := regResScene{fav: "nil", keys: []string{"ctrl+f"}}.build(t, regSize{100, 30})
		assert.Equal(t, "", m.notice)
		assert.Empty(t, m.favKeys)
	})
	t.Run("an error says so and changes nothing", func(t *testing.T) {
		m, api := regResScene{fav: "fail", keys: []string{"ctrl+f"}}.build(t, regSize{100, 30})
		assert.Equal(t, "Could not update FAVORITES: disk full", m.notice)
		assert.Empty(t, m.favKeys)
		assert.Contains(t, regRenderRes(m), " !  Could not update FAVORITES: disk full")
		assert.NotContains(t, regRenderRes(m), "★")
		// The notice takes rows above the list: the list got shorter.
		before := m.results.Height()
		api.err = nil
		regPress(m, "ctrl+f")
		assert.Equal(t, "", m.notice, "a later success clears the error")
		assert.Greater(t, m.results.Height(), before)
		assert.True(t, m.favKeys["frieren"])
	})
	t.Run("an error text is one clean line", func(t *testing.T) {
		m, api := regResScene{fav: "fail", keys: []string{}}.build(t, regSize{100, 30})
		api.err = errors.New("bad\nthing\x1b[31m  happened")
		regPress(m, "ctrl+f")
		assert.Equal(t, "Could not update FAVORITES: bad thing happened", m.notice)
	})
	t.Run("a success does not clear another notice", func(t *testing.T) {
		NoticeOnNextResults("Nothing there yet.")
		t.Cleanup(func() { NoticeOnNextResults("") })
		m, _ := regResScene{keys: []string{"ctrl+f"}}.build(t, regSize{100, 30})
		assert.Equal(t, "Nothing there yet.", m.notice)
	})
	t.Run("an error replaces another notice", func(t *testing.T) {
		NoticeOnNextResults("Nothing there yet.")
		t.Cleanup(func() { NoticeOnNextResults("") })
		m, _ := regResScene{fav: "fail", keys: []string{"ctrl+f"}}.build(t, regSize{100, 30})
		assert.Equal(t, "Could not update FAVORITES: disk full", m.notice)
		// frozen: the next success then clears the notice, whichever it was.
	})
}

func TestRegressionResultsKeys(t *testing.T) {
	regPin(t, false)
	t.Run("enter opens the chosen version", func(t *testing.T) {
		m, _ := regResScene{keys: []string{"right"}}.build(t, regSize{100, 30})
		assert.True(t, regPress(m, "enter"))
		require.NotNil(t, m.selected)
		assert.Equal(t, "HiAnime", m.selected.Source)
		m, _ = regResScene{}.build(t, regSize{100, 30})
		assert.True(t, regPress(m, "enter"))
		assert.Equal(t, "Animefire.io", m.selected.Source)
		m, _ = regResScene{keys: []string{"left"}}.build(t, regSize{100, 30})
		regPress(m, "enter")
		assert.Equal(t, "HiAnime", m.selected.Source, "left from the first version wraps to the last")
	})
	t.Run("esc goes back, q and ctrl+c quit", func(t *testing.T) {
		m, _ := regResScene{}.build(t, regSize{100, 30})
		assert.True(t, regPress(m, "esc"))
		assert.ErrorIs(t, m.err, ErrSelectionBack)
		m, _ = regResScene{}.build(t, regSize{100, 30})
		assert.True(t, regPress(m, "q"))
		assert.ErrorIs(t, m.err, ErrSelectionCancelled)
		m, _ = regResScene{}.build(t, regSize{100, 30})
		assert.True(t, regPress(m, "ctrl+c"))
		assert.ErrorIs(t, m.err, ErrSelectionCancelled)
	})
	t.Run("the help overlay swallows keys and closes with ? or esc", func(t *testing.T) {
		m, api := regResScene{keys: []string{"?"}}.build(t, regSize{100, 30})
		assert.True(t, m.showHelp)
		assert.False(t, regPress(m, "q", "enter", "ctrl+f", "down", "right"))
		assert.True(t, m.showHelp)
		assert.Empty(t, api.toggles)
		regPress(m, "esc")
		assert.False(t, m.showHelp)
		assert.NoError(t, m.err, "esc closed the help; it did not go back")
		regPress(m, "?", "ctrl+c")
		assert.ErrorIs(t, m.err, ErrSelectionCancelled)
	})
	t.Run("typing filters without a slash; q types while filtering", func(t *testing.T) {
		m, _ := regResScene{keys: []string{"type:n"}}.build(t, regSize{100, 30})
		assert.Equal(t, list.Filtering, m.results.FilterState())
		regPress(m, "q")
		assert.NoError(t, m.err)
		assert.Equal(t, "nq", m.results.FilterInput.Value())
		// The navigation letters do not open the filter: h moves, as in vim.
		m, _ = regResScene{}.build(t, regSize{100, 30})
		regPress(m, "h")
		assert.Equal(t, list.Unfiltered, m.results.FilterState())
	})
	t.Run("enter is held back while the filter is still being computed", func(t *testing.T) {
		m, _ := regResScene{keys: []string{"type:n", "enter"}}.build(t, regSize{100, 30})
		assert.Nil(t, m.selected)
	})
}

func TestRegressionResultsHelpText(t *testing.T) {
	for _, plain := range []bool{false, true} {
		regPin(t, plain)
		m, _ := regResScene{keys: []string{"?"}}.build(t, regSize{140, 40})
		text := regRenderRes(m)
		versionKey := map[bool]string{false: "←→", true: "left/right"}[plain]
		for _, want := range []string{"choose a version", "add to or remove from FAVORITES", "edit the search", "quit"} {
			assert.Contains(t, text, want, regMode(plain))
		}
		assert.Contains(t, text, "THIS SCREEN"+map[bool]string{false: " · ", true: " - "}[plain]+"RESULTS")
		order := []string{"choose a version", "add to or remove from FAVORITES", "edit the search"}
		last := -1
		for _, o := range order {
			at := strings.Index(text, o)
			require.Greater(t, at, last, o)
			last = at
		}
		assert.Contains(t, text, versionKey)
		assert.Contains(t, text, "esc never quits. ctrl+c always does.")
		// Too narrow for the overlay: the list stays.
		m, _ = regResScene{keys: []string{"?"}}.build(t, regSize{20, 30})
		assert.NotContains(t, regRenderRes(m), "THIS SCREEN")
	}
}

func TestRegressionResultsHeaderAndTitles(t *testing.T) {
	regPin(t, false)
	m, _ := regResScene{query: "  fri\x1b[31mer\nen  "}.build(t, regSize{100, 30})
	assert.Equal(t, `"frier en"`, m.shell.Context, "frozen: the query is cleaned to one line and quoted")
	assert.Equal(t, "Search > Results", m.shell.Breadcrumb)
	assert.Equal(t, "GoAnime - Results", m.View().WindowTitle)
	assert.True(t, m.View().AltScreen)
	m, _ = regResScene{query: "   "}.build(t, regSize{100, 30})
	assert.Equal(t, "", m.shell.Context)
	// The status line counts titles and versions.
	assert.Contains(t, regRenderRes(m), "5 titles · 6 versions · closest match first")
}

func TestRegressionResultsConstants(t *testing.T) {
	assert.Equal(t, 2, maxNoticeLines)
	assert.Equal(t, "anime selection cancelled", ErrSelectionCancelled.Error())
	assert.Equal(t, "back from anime selection", ErrSelectionBack.Error())
	assert.Equal(t, "no anime results to select", ErrNoAnimeResults.Error())
}
