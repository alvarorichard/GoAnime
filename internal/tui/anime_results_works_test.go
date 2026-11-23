package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Grouping is by work, and a work is a title plus, for a movie or a series,
// its release year and kind. The bug these guard: TopCine and StartFlix list
// iCarly (2007) and iCarly (2021) under one name, the results screen folded
// them into one row, and the 2021 series — the one that plays — could not be
// reached from the row that said iCarly.

func tcResult(name, url, source, year string, media models.MediaType) *models.Anime {
	return &models.Anime{Name: "[PT-BR] " + name, URL: url, Source: source, Year: year, MediaType: media}
}

// TestGroupSameTitleRemakesStayApart is the iCarly search as the logs show it.
func TestGroupSameTitleRemakesStayApart(t *testing.T) {
	t.Parallel()
	reunion := tcResult("iCarly: A Reunião", "tc/filme/icarly-a-reuniao", "TopCine", "2021", models.MediaTypeMovie)
	old := tcResult("iCarly", "tc/serie/icarly-2", "TopCine", "2007", models.MediaTypeTV)
	revival := tcResult("iCarly", "tc/serie/icarly", "TopCine", "2021", models.MediaTypeTV)
	sfx2007 := tcResult("iCarly", "sf/icarly-2007", "StartFlix", "2007", models.MediaTypeTV)
	sfx2021 := tcResult("iCarly", "sf/icarly-2021", "StartFlix", "2021", models.MediaTypeTV)

	groups := groupAnimeResults([]*models.Anime{reunion, old, revival, sfx2007, sfx2021})

	require.Len(t, groups, 3, "the reunion movie, the 2007 series and the 2021 series are three works")
	assert.Equal(t, "iCarly: A Reunião", groups[0].title)
	assert.Equal(t, "2007", groups[1].year)
	require.Len(t, groups[1].versions, 2, "each year gathers its own sources")
	assert.Same(t, old, groups[1].versions[0].anime)
	assert.Same(t, sfx2007, groups[1].versions[1].anime)
	assert.Equal(t, "2021", groups[2].year)
	require.Len(t, groups[2].versions, 2)
	assert.Same(t, revival, groups[2].versions[0].anime, "the 2021 row opens the 2021 series")
	assert.Same(t, sfx2021, groups[2].versions[1].anime)
}

// TestGroupSameWorkRules pins each rule on its own, so a regression names the
// rule it broke.
func TestGroupSameWorkRules(t *testing.T) {
	t.Parallel()
	type res struct {
		name, year string
		media      models.MediaType
	}
	cases := []struct {
		name  string
		a, b  res
		works int
	}{
		{"same title, same year, same kind merge", res{"Dexter", "2006", models.MediaTypeTV}, res{"Dexter", "2006", models.MediaTypeTV}, 1},
		{"same title, different years split", res{"iCarly", "2007", models.MediaTypeTV}, res{"iCarly", "2021", models.MediaTypeTV}, 2},
		{"a movie and its series stay apart", res{"Dexter", "2006", models.MediaTypeTV}, res{"Dexter", "2006", models.MediaTypeMovie}, 2},
		{"movie remakes split by year", res{"Fantasia", "1940", models.MediaTypeMovie}, res{"Fantasia", "2000", models.MediaTypeMovie}, 2},
		{"an unknown year joins a known one", res{"Dexter", "2006", models.MediaTypeTV}, res{"Dexter", "", models.MediaTypeTV}, 1},
		{"a known year joins an unknown one", res{"Dexter", "", models.MediaTypeTV}, res{"Dexter", "2006", models.MediaTypeTV}, 1},
		{"different titles never merge", res{"Dexter", "2006", models.MediaTypeTV}, res{"Dextrose", "2006", models.MediaTypeTV}, 2},
		{"tags and case do not make a new work", res{"Dexter", "2006", models.MediaTypeTV}, res{"DEXTER", "2006", models.MediaTypeTV}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := &models.Anime{Name: "[PT-BR] " + tc.a.name, Year: tc.a.year, MediaType: tc.a.media, Source: "TopCine"}
			b := &models.Anime{Name: "[TV] [PT-BR] " + tc.b.name, Year: tc.b.year, MediaType: tc.b.media, Source: "StartFlix"}
			assert.Len(t, groupAnimeResults([]*models.Anime{a, b}), tc.works)
			assert.Len(t, groupAnimeResults([]*models.Anime{b, a}), tc.works, "the order of the results never decides it")
		})
	}
}

// TestGroupAnimeYearsNeverSplit keeps what already worked: anime sources
// report the year of a season, which differs between sources for one work.
func TestGroupAnimeYearsNeverSplit(t *testing.T) {
	t.Parallel()
	ptbr := &models.Anime{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime}
	en := &models.Anime{Name: "[English] Frieren", Source: "HiAnime", Year: "2024", MediaType: models.MediaTypeAnime}
	none := &models.Anime{Name: "[PT-BR] Frieren", Source: "Goyabu", MediaType: models.MediaTypeAnime}
	groups := groupAnimeResults([]*models.Anime{ptbr, en, none})
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].versions, 3)
	assert.Equal(t, "2023", groups[0].year, "the first year the group saw stays")
}

// TestGroupYearlessCopyJoinsFirstMatch: a copy that names no year fits every
// group of its title, and goes to the closest match, the first.
func TestGroupYearlessCopyJoinsFirstMatch(t *testing.T) {
	t.Parallel()
	old := tcResult("iCarly", "a", "TopCine", "2007", models.MediaTypeTV)
	revival := tcResult("iCarly", "b", "TopCine", "2021", models.MediaTypeTV)
	bare := &models.Anime{Name: "[PT-BR] iCarly", URL: "c", Source: "Pomfy", MediaType: models.MediaTypeAnime}
	groups := groupAnimeResults([]*models.Anime{old, revival, bare})
	require.Len(t, groups, 2)
	require.Len(t, groups[0].versions, 2)
	assert.Same(t, bare, groups[0].versions[1].anime)
	assert.Len(t, groups[1].versions, 1)
}

// TestResultsScreenPicksTheRemake drives the screen: the row that says 2021
// opens the 2021 series.
func TestResultsScreenPicksTheRemake(t *testing.T) {
	t.Parallel()
	old := tcResult("iCarly", "tc/serie/icarly-2", "TopCine", "2007", models.MediaTypeTV)
	revival := tcResult("iCarly", "tc/serie/icarly", "TopCine", "2021", models.MediaTypeTV)
	m := newAnimeResultsModel([]*models.Anime{old, revival})
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	assert.Same(t, revival, m.selected)
	assert.Equal(t, "tc/serie/icarly", m.selected.URL)
}
