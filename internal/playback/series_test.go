package playback

import (
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests in this file mutate package-level vars used as DI seams.
// Do not add t.Parallel() — these tests must run sequentially.

func restoreSeriesState() {
	seriesSelectEpisodeWithFuzzyFinder = player.SelectEpisodeWithFuzzyFinder
	seriesExtractEpisodeNumber = player.ExtractEpisodeNumber
	seriesSelectEpisodeWithFuzzy = SelectEpisodeWithFuzzy
	seriesFindEpisodeByNumber = FindEpisodeByNumber
	seriesIsAllAnimeSource = seriesIsAllAnimeSourceDefault
	seriesHandleAllAnimeNavigation = handleAllAnimeNavigation
	seriesCheckIsSeries = seriesCheckIsSeriesDefault
	seriesCheckIsSeriesEnhanced = seriesCheckIsSeriesEnhancedDefault
}

var (
	seriesIsAllAnimeSourceDefault      = seriesIsAllAnimeSource
	seriesCheckIsSeriesDefault         = seriesCheckIsSeries
	seriesCheckIsSeriesEnhancedDefault = seriesCheckIsSeriesEnhanced
)

func TestSelectInitialEpisodeParsesSelection(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesSelectEpisodeWithFuzzyFinder = func([]models.Episode) (string, string, error) {
		return "https://example.com/ep12", "Episode 12", nil
	}
	seriesExtractEpisodeNumber = func(value string) string {
		assert.Equal(t, "Episode 12", value)
		return "12"
	}

	url, number, episodeNum, err := SelectInitialEpisode([]models.Episode{{Number: "Episode 12"}})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/ep12", url)
	assert.Equal(t, "Episode 12", number)
	assert.Equal(t, 12, episodeNum)
}

func TestSelectInitialEpisodePropagatesBackRequest(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesSelectEpisodeWithFuzzyFinder = func([]models.Episode) (string, string, error) {
		return "", "", player.ErrBackRequested
	}

	_, _, episodeNum, err := SelectInitialEpisode(nil)
	require.ErrorIs(t, err, player.ErrBackRequested)
	assert.Equal(t, -1, episodeNum)
}

func TestHandleUserNavigationUsesExpectedTargets(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesSelectEpisodeWithFuzzy = func([]models.Episode) (string, string, int, error) {
		return "manual-url", "Episode 7", 7, nil
	}

	var requested []int
	seriesFindEpisodeByNumber = func(_ []models.Episode, num int) (string, string, int, error) {
		requested = append(requested, num)
		return fmt.Sprintf("url-%d", num), fmt.Sprintf("Episode %d", num), num, nil
	}

	url, numStr, epNum := handleUserNavigation("e", nil, 5, 20)
	assert.Equal(t, "manual-url", url)
	assert.Equal(t, "Episode 7", numStr)
	assert.Equal(t, 7, epNum)

	url, numStr, epNum = handleUserNavigation("p", nil, 1, 20)
	assert.Equal(t, "url-1", url)
	assert.Equal(t, "Episode 1", numStr)
	assert.Equal(t, 1, epNum)

	url, numStr, epNum = handleUserNavigation("n", nil, 20, 20)
	assert.Equal(t, "url-20", url)
	assert.Equal(t, "Episode 20", numStr)
	assert.Equal(t, 20, epNum)

	assert.Equal(t, []int{1, 20}, requested)
}

func TestHandleUserNavigationEnhancedUsesAllAnimeBranch(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesIsAllAnimeSource = func(anime *models.Anime) bool {
		return anime.Source == "AllAnime"
	}
	seriesHandleAllAnimeNavigation = func(input string, episodes []models.Episode, currentNum, totalEpisodes int, anime *models.Anime) (string, string, int) {
		assert.Equal(t, "n", input)
		assert.Equal(t, 5, currentNum)
		assert.Equal(t, 20, totalEpisodes)
		assert.Equal(t, "AllAnime", anime.Source)
		return "allanime-url", "Episode 6", 6
	}

	url, numStr, epNum := handleUserNavigationEnhanced("n", nil, 5, 20, &models.Anime{Source: "AllAnime"})
	assert.Equal(t, "allanime-url", url)
	assert.Equal(t, "Episode 6", numStr)
	assert.Equal(t, 6, epNum)
}

func TestHandleUserNavigationEnhancedFallsBackToRegularNavigation(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesIsAllAnimeSource = func(*models.Anime) bool { return false }
	seriesFindEpisodeByNumber = func(_ []models.Episode, num int) (string, string, int, error) {
		return fmt.Sprintf("url-%d", num), fmt.Sprintf("Episode %d", num), num, nil
	}

	url, numStr, epNum := handleUserNavigationEnhanced("n", nil, 2, 10, &models.Anime{Source: "Animefire.io"})
	assert.Equal(t, "url-3", url)
	assert.Equal(t, "Episode 3", numStr)
	assert.Equal(t, 3, epNum)
}

func TestCheckIfSeriesReturnsSafeFallbackOnError(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesCheckIsSeries = func(string) (bool, int, error) {
		return true, 24, fmt.Errorf("upstream failed")
	}

	series, totalEpisodes := CheckIfSeries("https://example.com/anime")
	assert.False(t, series)
	assert.Equal(t, 1, totalEpisodes)
}

func TestCheckIfSeriesEnhancedReturnsSafeFallbackOnError(t *testing.T) {
	restoreSeriesState()
	t.Cleanup(restoreSeriesState)

	seriesCheckIsSeriesEnhanced = func(*models.Anime) (bool, int, error) {
		return true, 24, fmt.Errorf("upstream failed")
	}

	series, totalEpisodes := CheckIfSeriesEnhanced(&models.Anime{Name: "Naruto"})
	assert.False(t, series)
	assert.Equal(t, 1, totalEpisodes)
}
