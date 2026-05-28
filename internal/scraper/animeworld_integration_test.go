package scraper_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
)

// Integration Tests following superflix_integration_test.go as an example
// Real website used

func TestIntegration_RealAnimeWorld_SearchAndVerify(t *testing.T) {
	if v := strings.ToLower(os.Getenv("GITHUB_ACTIONS")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if testing.Short() {
		t.Skip("Skipping integration test that hits real AnimeWorld website")
	}

	searches := []string{"Naruto", "Bleach", "One Piece"}
	client := scraper.NewAnimeWorldClient()

	for _, query := range searches {
		t.Run(query, func(t *testing.T) {
			anime, err := client.SearchAnime(query)
			assert.NoError(t, err)
			assert.NotEmpty(t, anime)

			for _, a := range anime {
				assert.NotEmpty(t, a.Name)
				assert.NotEmpty(t, a.URL)
				assert.NotEmpty(t, a.ImageURL)
			}
		})
	}
}

func TestIntegration_RealAnimeWorld_SearchEpisodes(t *testing.T) {
	if v := strings.ToLower(os.Getenv("GITHUB_ACTIONS")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if testing.Short() {
		t.Skip("Skipping integration test that hits real AnimeWorld website")
	}

	// finished airing anime
	table := []struct {
		animeName     string
		animeURL      string
		totalEpisodes int
		// some episodes are were double aired on TV
		// For these three animes, I manually checked their number
		doubleEpisodeNumber int
	}{
		{
			animeName:           "Naruto Shippuden",
			animeURL:            "https://www.animeworld.ac/play/naruto-shippuden.v3U8a/ZXFbr",
			totalEpisodes:       500,
			doubleEpisodeNumber: 12,
		},
		{
			animeName:           "Bleach",
			animeURL:            "https://www.animeworld.ac/play/bleach.1xU4f/d7tDd",
			totalEpisodes:       366,
			doubleEpisodeNumber: 6,
		},
		{
			animeName:           "Dragonball Z",
			animeURL:            "https://www.animeworld.ac/play/dragon-ball-z-ita.NND05/O7sga",
			totalEpisodes:       291,
			doubleEpisodeNumber: 0,
		},
		//{
		//	animeName:     "One Piece",
		//	animeURL:      "https://www.animeworld.ac/play/one-piece-subita.qzG-LE/HPKmX1",
		//	totalEpisodes: getInfinite(),
		//},
	}

	client := scraper.NewAnimeWorldClient()
	for _, tt := range table {
		t.Run(tt.animeName, func(t *testing.T) {
			eps, err := client.GetAnimeEpisodes(tt.animeURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.totalEpisodes-tt.doubleEpisodeNumber, len(eps))

			for _, ep := range eps {
				assert.NotEmpty(t, ep.URL)

				// check if double episode is correctly extracted
				// Episodio 12-13  => Num = 12
				epStr := ep.Number
				epStr = strings.Split(epStr, " ")[1]
				epStrLeft := strings.Split(epStr, "-")[0]
				epNum, err := strconv.Atoi(epStrLeft)
				assert.NoError(t, err)
				assert.Equal(t, ep.Num, epNum)
			}
		})
	}

}

func TestIntegration_RealAnimeWorld_GetStreamURL(t *testing.T) {
	if v := strings.ToLower(os.Getenv("GITHUB_ACTIONS")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream call")
	}
	if testing.Short() {
		t.Skip("Skipping integration test that hits real AnimeWorld website")
	}

	client := scraper.NewAnimeWorldClient()
	table := []struct {
		animeName  string
		episodeURL string
	}{
		{animeName: "Naruto Shippuden EP 24", episodeURL: "https://www.animeworld.ac/play/naruto-shippuden.v3U8a/KtvO4"},
		{animeName: "One Piece EP 405", episodeURL: "https://www.animeworld.ac/play/one-piece-subita.qzG-LE/mJk6qG"},
		{animeName: "Frieren EP 7", episodeURL: "https://www.animeworld.ac/play/sousou-no-frieren.kbcPv/wBsiqn"},
	}

	for _, tt := range table {
		t.Run(tt.animeName, func(t *testing.T) {
			stream, err := client.GetStreamURL(tt.episodeURL)
			assert.NoError(t, err)
			assert.NotEmpty(t, stream)
			assert.True(t, strings.HasSuffix(stream, ".mp4"))
		})
	}
}
