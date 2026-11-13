package tui

// Rules that the mutation pass over search_home.go and anime_results.go
// showed the screen goldens do not pin by themselves.

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegressionHomeCursorAfterRemovals(t *testing.T) {
	regPin(t, false)
	m := regHome{continues: true, recents: true, favs: 3}.build()
	regPress(m, "down", "end", "x")
	assert.Equal(t, 6, m.cursor, "removing the last row leaves the highlight on the new last row")
	assert.Equal(t, 7, len(m.rows))
	regPress(m, "x")
	assert.Equal(t, 5, m.cursor)
	assert.Empty(t, m.recentAll)
	assert.Equal(t, 6, len(m.rows))

	tab := regHome{favs: 5, scope: 4}.build()
	regPress(tab, "down", "end", "x")
	assert.Equal(t, 3, tab.cursor)
	assert.Equal(t, 4, len(tab.rows))
}

func TestRegressionHomeTabResetsTheHighlight(t *testing.T) {
	regPin(t, false)
	m := regHome{continues: true, recents: true, favs: 3}.build()
	regPress(m, "down", "down", "down", "down", "down")
	require.Equal(t, 4, m.cursor)
	regPress(m, "tab")
	assert.Equal(t, 0, m.cursor)
	assert.True(t, m.focusInput)
	regPress(m, "tab", "tab", "tab", "tab", "down")
	assert.Equal(t, 0, m.scopeIdx)
	assert.Equal(t, 0, m.cursor, "back on ALL the first row is the one highlighted")
}

func TestRegressionHomeQueryLengthIsCountedInLetters(t *testing.T) {
	regPin(t, false)
	m := regHome{query: "é"}.build()
	assert.False(t, regPress(m, "enter"), "one accented letter is one letter")
	m = regHome{query: "日本"}.build()
	assert.True(t, regPress(m, "enter"))
	assert.Equal(t, "日本", m.result.Query)
	m = regHome{query: "日"}.build()
	assert.False(t, regPress(m, "enter"))
}

func TestRegressionHomeAskToClearNeedsAListToClear(t *testing.T) {
	regPin(t, false)
	m := regHome{}.build()
	m.askToClear(homeRowFavorite)
	m.askToClear(homeRowRecent)
	m.askToClear(homeRowContinue)
	assert.Equal(t, clearNothing, m.clearing, "nothing to clear, nothing to ask")
	full := regHome{continues: true, recents: true, favs: 2}.build()
	full.askToClear(homeRowFavorite)
	assert.Equal(t, clearFavoriteRows, full.clearing)
	full.clearing = clearNothing
	full.askToClear(homeRowRecent)
	assert.Equal(t, clearRecentRows, full.clearing)
	full.clearing = clearNothing
	full.askToClear(homeRowContinue)
	assert.Equal(t, clearContinueRows, full.clearing)
}

func TestRegressionHomeOnlyYAndNAnswerTheClearQuestion(t *testing.T) {
	regPin(t, false)
	calls := &regCalls{}
	m := regHome{continues: true, calls: calls}.build()
	regPress(m, "down", "X", "enter", "a", "space", "down", "up", "tab", "r", "e", "x", "X")
	assert.Equal(t, clearContinueRows, m.clearing, "still asking")
	assert.Empty(t, calls.log)
	assert.Equal(t, 0, m.scopeIdx)
	assert.Equal(t, "", m.input.Value())
	regPress(m, "y")
	assert.Equal(t, []string{"clear-continue"}, calls.log)
}

func TestRegressionHomeFilteredCountHeader(t *testing.T) {
	regPin(t, false)
	header := func(query string) string {
		m := regHome{favs: 3, scope: 4, query: query}.build()
		for _, line := range strings.Split(regRender(m), "\n") {
			if strings.HasPrefix(line, "FAVORITES") {
				return strings.TrimSpace(line)
			}
		}
		return ""
	}
	assert.Equal(t, "FAVORITES  3", header(""))
	assert.Equal(t, "FAVORITES  2 of 3", header("t"))
	assert.Equal(t, "FAVORITES  1 of 3", header("heat"))
	assert.Equal(t, "FAVORITES  3", header("a"), "all three match: the plain total")
}

func TestRegressionHomeHostInputIsNotAliasedOrTrusted(t *testing.T) {
	regPin(t, false)
	favs := regFavorites(3)
	opts := SearchHomeOptions{Favorites: favs, SourcesLine: " 6 sources\n on\x1b[31m  ", DeleteFavorite: func(string) error { return nil }}
	m := newSearchHomeModel(opts)
	assert.Equal(t, "6 sources on", m.shell.Context, "the sources line is one clean line")
	regPress(m, "down", "x")
	assert.Equal(t, "frieren", favs[0].Key, "the host's slice is not changed under it")
	assert.Equal(t, "heat", favs[1].Key)
	assert.Equal(t, "pokemon aventuras", favs[2].Key)
	assert.Len(t, m.favoriteAll, 2)
}

// ---- results --------------------------------------------------------------------------------------

func TestRegressionResultsListWidth(t *testing.T) {
	for _, tt := range []struct {
		plain bool
		w     int
		want  int
	}{
		{false, 140, 84 - resultsGutter},
		{false, 100, 60 - resultsGutter},
		{false, 99, 99},
		{false, 80, 80},
		{true, 140, 140},
		{true, 100, 100},
	} {
		regPin(t, tt.plain)
		m, _ := regResScene{}.build(t, regSize{tt.w, 40})
		assert.Equal(t, tt.want, m.results.Width(), "%s width %d", regMode(tt.plain), tt.w)
	}
}

func TestRegressionResultsVersionCycle(t *testing.T) {
	regPin(t, false)
	m, api := regResScene{keys: []string{"right", "right", "right"}}.build(t, regSize{100, 30})
	assert.Equal(t, 1, m.versionSel[0], "three steps over two versions: the second")
	regPress(m, "ctrl+f")
	assert.Equal(t, "HiAnime", api.toggles[0].Source)
	m, _ = regResScene{keys: []string{"left", "left", "left"}}.build(t, regSize{100, 30})
	assert.Equal(t, 1, m.versionSel[0])
	// A stale or wild selection reads as the first version rather than failing.
	m.versionSel[0] = 2
	assert.Equal(t, "Animefire.io", m.chosenVersion(0).Source)
	m.versionSel[0] = -1
	assert.Equal(t, "Animefire.io", m.chosenVersion(0).Source)
	assert.Nil(t, m.chosenVersion(-1))
	assert.Nil(t, m.chosenVersion(99))
}

func TestRegressionResultsEnterWaitsForTheFilter(t *testing.T) {
	regPin(t, false)
	m, _ := regResScene{keys: []string{"type:n"}}.build(t, regSize{100, 30})
	require.Equal(t, list.Filtering, m.results.FilterState())
	require.True(t, m.filterPending)
	regPress(m, "enter")
	assert.Equal(t, list.Filtering, m.results.FilterState(), "enter is ignored until the matches arrive")
	assert.Nil(t, m.selected)
	_, _ = m.Update(list.FilterMatchesMsg(nil))
	assert.False(t, m.filterPending)
}

func TestRegressionResultsNoticeCutOver(t *testing.T) {
	regPin(t, false)
	theme := CurrentTheme()
	notice := "abcdefghij klmnopqrst"
	wide := ansi.Strip(renderNotice(&theme, notice, 16)) // room 12: wrapped under the block
	assert.Equal(t, " !  abcdefghij\n    klmnopqrst\n", wide)
	narrow := ansi.Strip(renderNotice(&theme, notice, 15)) // room 11: just the text, cut
	assert.Equal(t, "abcdefghij klm…\n", narrow)
	assert.Equal(t, "", renderNotice(&theme, "", 80))
	regPin(t, true)
	theme = CurrentTheme()
	assert.Equal(t, "[!] abcdefghij\n    klmnopqrst\n", ansi.Strip(renderNotice(&theme, notice, 16)))
}

func TestRegressionResultsLabelsAndFilterValue(t *testing.T) {
	regPin(t, false)
	for in, want := range map[models.MediaType]string{
		models.MediaTypeAnime: "Anime", models.MediaTypeMovie: "Movie", models.MediaTypeTV: "TV",
		"": "", "Special": "Special", "a\nb\x1b[31m": "a b",
	} {
		assert.Equal(t, want, mediaTypeLabel(in), "%q", in)
	}
	groups := groupAnimeResults(regAnimes())
	assert.Equal(t, "Frieren 2023 Anime AnimeFire PT-BR PT-BR Legendado HiAnime EN English Sub", animeGroupItem{group: &groups[0]}.FilterValue())
	assert.Equal(t, "Mushishi Goyabu PT-BR PT-BR", animeGroupItem{group: &groups[3]}.FilterValue(), "no year and no type leave out")
	assert.Equal(t, "Naruto 2002 Anime AnimeFire PT-BR PT-BR 1080p", animeGroupItem{group: &groups[4]}.FilterValue())
	assert.Equal(t, "", animeGroupItem{}.FilterValue())

	bare := groupAnimeResults([]*models.Anime{{Name: "Nameless"}})
	require.Len(t, bare, 1)
	assert.Equal(t, "Information unavailable", bare[0].description(), "no year, no type and no source")
	assert.Equal(t, "Information unavailable", animeGroupItem{}.Description())
	assert.Equal(t, "Unknown title", animeGroupItem{}.Title())
}
