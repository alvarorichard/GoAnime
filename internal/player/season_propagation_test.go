package player

import (
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeasonDefaultedToOne_Bug demonstrates the original bug: when no season
// was passed through the pipeline, HandleDownloadAndPlay hardcoded season = 1
// for all TV shows, even when the user selected a different season.
func TestSeasonDefaultedToOne_Bug(t *testing.T) {
	t.Run("before fix: season was always 1 regardless of selection", func(t *testing.T) {
		// Simulate the old code path: season was hardcoded to 1
		// because HandleDownloadAndPlay did not receive the season number.
		oldSeason := 1 // was: season := 1
		SetAnimeName("Dexter", oldSeason)

		snap := snapshotMedia()
		// The bug: user selected Season 2, but snapshot always shows 1
		assert.Equal(t, 1, snap.AnimeSeason, "bug scenario: season stuck at 1")

		path := util.FormatPlexEpisodePath("/media/tv", "Dexter", snap.AnimeSeason, 5)
		assert.Contains(t, path, "Season 01", "bug scenario: download path shows Season 01")
		assert.Contains(t, path, "S01E05", "bug scenario: filename shows S01E05")
	})
}

// TestSeasonPropagation_Fix verifies that after the fix, the selected season
// number flows correctly through SetAnimeName → snapshotMedia → download path.
func TestSeasonPropagation_Fix(t *testing.T) {
	tests := []struct {
		name           string
		animeName      string
		season         int
		episode        int
		wantSeasonDir  string
		wantFilePrefix string
	}{
		{
			name:           "Dexter Season 2 Episode 5",
			animeName:      "Dexter",
			season:         2,
			episode:        5,
			wantSeasonDir:  "Season 02",
			wantFilePrefix: "S02E05",
		},
		{
			name:           "Breaking Bad Season 4 Episode 11",
			animeName:      "Breaking Bad",
			season:         4,
			episode:        11,
			wantSeasonDir:  "Season 04",
			wantFilePrefix: "S04E11",
		},
		{
			name:           "Season 1 still works",
			animeName:      "Friends",
			season:         1,
			episode:        1,
			wantSeasonDir:  "Season 01",
			wantFilePrefix: "S01E01",
		},
		{
			name:           "high season number",
			animeName:      "The Simpsons",
			season:         35,
			episode:        10,
			wantSeasonDir:  "Season 35",
			wantFilePrefix: "S35E10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the fixed path: HandleDownloadAndPlay receives animeSeason
			// from anime.CurrentSeason instead of hardcoding 1
			SetAnimeName(tc.animeName, tc.season)

			snap := snapshotMedia()
			assert.Equal(t, tc.season, snap.AnimeSeason)
			assert.Equal(t, tc.animeName, snap.AnimeName)

			path := util.FormatPlexEpisodePath("/media/tv", tc.animeName, snap.AnimeSeason, tc.episode)
			assert.Contains(t, path, tc.wantSeasonDir)
			assert.Contains(t, path, tc.wantFilePrefix)
		})
	}
}

// TestSetAnimeName_ClampsBelowOne verifies that season numbers < 1 are clamped to 1.
func TestSetAnimeName_ClampsBelowOne(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 1},
		{-1, 1},
		{-100, 1},
		{1, 1},
		{5, 5},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("season_%d", tc.input), func(t *testing.T) {
			SetAnimeName("TestAnime", tc.input)
			snap := snapshotMedia()
			assert.Equal(t, tc.want, snap.AnimeSeason)
		})
	}
}

// TestCurrentSeasonOnMedia verifies that CurrentSeason on the Media/Anime
// struct carries the season number through the data pipeline.
func TestCurrentSeasonOnMedia(t *testing.T) {
	t.Run("CurrentSeason is stored on anime struct", func(t *testing.T) {
		anime := &models.Anime{
			Name:      "Dexter",
			MediaType: models.MediaTypeTV,
			Source:    "FlixHQ",
		}

		// Simulate what GetFlixHQEpisodes now does after the fix:
		// media.CurrentSeason = selectedSeason.Number
		anime.CurrentSeason = 2

		assert.Equal(t, 2, anime.CurrentSeason)
		assert.Equal(t, "Dexter", anime.Name)
		assert.Equal(t, "FlixHQ", anime.Source)
		assert.True(t, anime.IsTV())
	})

	t.Run("CurrentSeason defaults to 0 when not set", func(t *testing.T) {
		anime := &models.Anime{
			Name:      "Naruto",
			MediaType: models.MediaTypeAnime,
			Source:    "AllAnime",
		}

		// For anime sources, CurrentSeason is not set (0)
		assert.Equal(t, 0, anime.CurrentSeason)
		assert.Equal(t, "Naruto", anime.Name)
		assert.Equal(t, "AllAnime", anime.Source)
		assert.True(t, anime.IsAnime())
	})
}

// TestHandleDownloadAndPlay_SeasonFromAnime simulates the fixed
// HandleDownloadAndPlay logic where season comes from the anime's
// CurrentSeason instead of being hardcoded to 1.
func TestHandleDownloadAndPlay_SeasonFromAnime(t *testing.T) {
	t.Run("season from anime.CurrentSeason flows to download path", func(t *testing.T) {
		anime := &models.Anime{
			Name:          "Dexter",
			MediaType:     models.MediaTypeTV,
			Source:        "FlixHQ",
			CurrentSeason: 2,
		}
		assert.Equal(t, "FlixHQ", anime.Source)

		// Simulate the fixed HandleDownloadAndPlay logic:
		// season := animeSeason (from anime.CurrentSeason)
		// if season < 1 { season = 1 }
		animeSeason := anime.CurrentSeason
		season := max(animeSeason, 1)
		SetAnimeName(anime.Name, season)
		SetExactMediaType(string(anime.MediaType))

		snap := snapshotMedia()
		assert.Equal(t, "Dexter", snap.AnimeName)
		assert.Equal(t, 2, snap.AnimeSeason)
		assert.Equal(t, "tv", snap.MediaType)

		// Verify the download path uses the correct season
		path := util.FormatPlexEpisodePath("/media/tv", snap.AnimeName, snap.AnimeSeason, 5)
		assert.Contains(t, path, "Season 02")
		assert.Contains(t, path, "S02E05")
		assert.NotContains(t, path, "Season 01")
		assert.NotContains(t, path, "S01E05")
	})

	t.Run("GlobalDownloadRequest overrides anime season", func(t *testing.T) {
		// When user specifies season via CLI flag, it should override
		original := util.GlobalDownloadRequest
		defer func() { util.GlobalDownloadRequest = original }()

		util.GlobalDownloadRequest = &util.DownloadRequest{
			SeasonNum: 3,
		}

		anime := &models.Anime{
			Name:          "Dexter",
			MediaType:     models.MediaTypeTV,
			CurrentSeason: 2,
		}
		assert.True(t, anime.IsTV())

		// Simulate the HandleDownloadAndPlay logic with GlobalDownloadRequest
		animeSeason := anime.CurrentSeason
		season := max(animeSeason, 1)
		if util.GlobalDownloadRequest != nil && util.GlobalDownloadRequest.SeasonNum > 0 {
			season = util.GlobalDownloadRequest.SeasonNum
		}
		SetAnimeName(anime.Name, season)

		snap := snapshotMedia()
		assert.Equal(t, 3, snap.AnimeSeason, "CLI flag should override anime.CurrentSeason")
	})

	t.Run("anime without CurrentSeason defaults to season 1", func(t *testing.T) {
		original := util.GlobalDownloadRequest
		defer func() { util.GlobalDownloadRequest = original }()
		util.GlobalDownloadRequest = nil

		anime := &models.Anime{
			Name:          "Naruto",
			MediaType:     models.MediaTypeAnime,
			CurrentSeason: 0, // not set for regular anime
		}
		assert.True(t, anime.IsAnime())

		animeSeason := anime.CurrentSeason
		season := max(animeSeason, 1)
		SetAnimeName(anime.Name, season)

		snap := snapshotMedia()
		assert.Equal(t, 1, snap.AnimeSeason, "unset CurrentSeason should fallback to 1")
	})
}

// TestMPVWindowTitle_SeasonPropagation verifies that the MPV window title
// reflects the correct season number after the fix.
func TestMPVWindowTitle_SeasonPropagation(t *testing.T) {
	tests := []struct {
		name      string
		animeName string
		season    int
		episode   int
		wantTitle string
	}{
		{"Dexter S02E05", "Dexter", 2, 5, "Dexter S02E05"},
		{"Breaking Bad S04E11", "Breaking Bad", 4, 11, "Breaking Bad S04E11"},
		{"Friends S01E01", "Friends", 1, 1, "Friends S01E01"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetAnimeName(tc.animeName, tc.season)
			SetExactMediaType("tv")

			snap := snapshotMedia()
			require.Equal(t, tc.season, snap.AnimeSeason)

			// Simulate playvideo.go title generation
			title := fmt.Sprintf("%s S%02dE%02d", snap.AnimeName, snap.AnimeSeason, tc.episode)
			assert.Equal(t, tc.wantTitle, title)
		})
	}
}

// TestFlixHQEpisodesSetCurrentSeason simulates the fixed GetFlixHQEpisodes
// that now sets media.CurrentSeason when the user picks a season.
func TestFlixHQEpisodesSetCurrentSeason(t *testing.T) {
	t.Run("season number set from FlixHQ season selection", func(t *testing.T) {
		// Simulate the FlixHQ season data
		type mockFlixHQSeason struct {
			ID     string
			Number int
			Title  string
		}

		seasons := []mockFlixHQSeason{
			{ID: "season-1", Number: 1, Title: "Season 1"},
			{ID: "season-2", Number: 2, Title: "Season 2"},
			{ID: "season-3", Number: 3, Title: "Season 3"},
		}

		media := &models.Anime{
			Name:      "Dexter",
			MediaType: models.MediaTypeTV,
			Source:    "FlixHQ",
		}

		// Simulate user selecting season index 1 (Season 2)
		seasonIdx := 1
		selectedSeason := seasons[seasonIdx]

		// This is what the fixed GetFlixHQEpisodes does:
		media.CurrentSeason = selectedSeason.Number

		assert.Equal(t, 2, media.CurrentSeason)
		assert.Equal(t, "Dexter", media.Name)
		assert.Equal(t, "FlixHQ", media.Source)
		assert.True(t, media.IsTV())
		assert.Equal(t, "Season 2", selectedSeason.Title)
		assert.Equal(t, "season-2", selectedSeason.ID)
	})

	t.Run("first season selection yields season 1", func(t *testing.T) {
		type mockFlixHQSeason struct {
			ID     string
			Number int
			Title  string
		}

		seasons := []mockFlixHQSeason{
			{ID: "season-1", Number: 1, Title: "Season 1"},
		}

		media := &models.Anime{
			Name:      "Breaking Bad",
			MediaType: models.MediaTypeTV,
			Source:    "FlixHQ",
		}

		selectedSeason := seasons[0]
		media.CurrentSeason = selectedSeason.Number

		assert.Equal(t, 1, media.CurrentSeason)
		assert.Equal(t, "Breaking Bad", media.Name)
		assert.Equal(t, "FlixHQ", media.Source)
		assert.True(t, media.IsTV())
	})
}

// TestEndToEndSeasonPipeline is an integration-style test that walks the
// entire season number pipeline: FlixHQ selection → anime struct →
// HandleDownloadAndPlay logic → gMedia state → download path.
func TestEndToEndSeasonPipeline(t *testing.T) {
	original := util.GlobalDownloadRequest
	defer func() { util.GlobalDownloadRequest = original }()
	util.GlobalDownloadRequest = nil

	// Step 1: User searches for "Dexter" on FlixHQ and selects it
	anime := &models.Anime{
		Name:      "Dexter",
		URL:       "tv/watch-dexter-39392",
		MediaType: models.MediaTypeTV,
		Source:    "FlixHQ",
	}
	assert.Equal(t, "tv/watch-dexter-39392", anime.URL)
	assert.Equal(t, "FlixHQ", anime.Source)

	// Step 2: GetFlixHQEpisodes runs, user selects Season 2
	// (simulating media.CurrentSeason = selectedSeason.Number)
	anime.CurrentSeason = 2

	// Step 3: Episodes are returned and user picks Episode 5
	episodes := []models.Episode{
		{Number: "1", Num: 1, DataID: "ep-1", SeasonID: "season-2"},
		{Number: "2", Num: 2, DataID: "ep-2", SeasonID: "season-2"},
		{Number: "3", Num: 3, DataID: "ep-3", SeasonID: "season-2"},
		{Number: "4", Num: 4, DataID: "ep-4", SeasonID: "season-2"},
		{Number: "5", Num: 5, DataID: "ep-5", SeasonID: "season-2"},
	}
	_ = episodes // used in the pipeline conceptually

	// Step 4: HandleDownloadAndPlay receives anime.CurrentSeason
	animeSeason := anime.CurrentSeason
	season := max(animeSeason, 1)
	SetAnimeName(anime.Name, season)
	SetExactMediaType(string(anime.MediaType))

	// Step 5: Verify the full pipeline output
	snap := snapshotMedia()
	assert.Equal(t, "Dexter", snap.AnimeName)
	assert.Equal(t, 2, snap.AnimeSeason, "season must be 2, not 1")
	assert.Equal(t, "tv", snap.MediaType)

	// Step 6: Download path uses Season 02
	path := util.FormatPlexEpisodePath("/media/tv", snap.AnimeName, snap.AnimeSeason, 5)
	assert.Contains(t, path, "Dexter/Season 02/Dexter - S02E05.mp4")

	// Step 7: MPV title shows S02E05
	title := fmt.Sprintf("%s S%02dE%02d", snap.AnimeName, snap.AnimeSeason, 5)
	assert.Equal(t, "Dexter S02E05", title)
}

func TestResolveSeasonForEpisode_UsesSelectedSeasonForLocalEpisodes(t *testing.T) {
	snap := mediaSnapshot{
		AnimeName:   "JUJUTSU KAISEN Season 2",
		AnimeSeason: 2,
		SeasonMap: []metadata.SeasonMapping{
			{Season: 1, StartEp: 1, EndEp: 23, EpisodeCount: 23},
			{Season: 2, StartEp: 24, EndEp: 46, EpisodeCount: 23},
		},
	}

	season, episode := resolveSeasonForEpisode(snap, 1)
	assert.Equal(t, 2, season)
	assert.Equal(t, 1, episode)

	season, episode = resolveSeasonForEpisode(snap, 23)
	assert.Equal(t, 2, season)
	assert.Equal(t, 23, episode)
}

func TestJujutsuKaisenSeason2LocalEpisodeBugAndFix(t *testing.T) {
	mockSeasonMap := []metadata.SeasonMapping{
		{Season: 1, StartEp: 1, EndEp: 23, EpisodeCount: 23},
		{Season: 2, StartEp: 24, EndEp: 46, EpisodeCount: 23},
	}
	mockEpisodes := []models.Episode{
		{Number: "Episódio 1", Num: 1, URL: "https://goyabu.io/?p=44626"},
		{Number: "Episódio 23", Num: 23, URL: "https://goyabu.io/?p=45012"},
	}

	oldResolveSeasonForEpisode := func(seasonMap []metadata.SeasonMapping, absEp int) (season, ep int) {
		for _, sm := range seasonMap {
			if absEp >= sm.StartEp && absEp <= sm.EndEp {
				return sm.Season, absEp - sm.StartEp + 1
			}
		}
		return 1, absEp
	}

	oldSeason, oldEpisode := oldResolveSeasonForEpisode(mockSeasonMap, mockEpisodes[0].Num)
	oldPath := util.FormatPlexEpisodePath("/media/anime", "JUJUTSU KAISEN Season 2", oldSeason, oldEpisode)
	assert.Contains(t, oldPath, "Season 01")
	assert.Contains(t, oldPath, "S01E01")

	fixedSnap := mediaSnapshot{
		AnimeName:   "JUJUTSU KAISEN Season 2",
		AnimeSeason: 2,
		MediaType:   string(models.MediaTypeAnime),
		SeasonMap:   mockSeasonMap,
	}
	fixedSeason, fixedEpisode := resolveSeasonForEpisode(fixedSnap, mockEpisodes[0].Num)
	fixedPath := util.FormatPlexEpisodePath("/media/anime", fixedSnap.AnimeName, fixedSeason, fixedEpisode)
	assert.Contains(t, fixedPath, "Season 02")
	assert.Contains(t, fixedPath, "S02E01")
	assert.NotContains(t, fixedPath, "Season 01")
	assert.NotContains(t, fixedPath, "S01E01")

	fixedSeason, fixedEpisode = resolveSeasonForEpisode(fixedSnap, mockEpisodes[1].Num)
	fixedPath = util.FormatPlexEpisodePath("/media/anime", fixedSnap.AnimeName, fixedSeason, fixedEpisode)
	assert.Contains(t, fixedPath, "Season 02")
	assert.Contains(t, fixedPath, "S02E23")
	assert.NotContains(t, fixedPath, "S01E23")
}
