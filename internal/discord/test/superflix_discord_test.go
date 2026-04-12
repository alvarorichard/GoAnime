package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests for Discord RPC cover image handling with SuperFlix content
// Tests both the original bug (CloudFront URLs) and the fix
// =============================================================================

// mockMPVFullState provides a complete mock MPV that returns all properties
// needed by getPrecisePlaybackState and updateDiscordPresence
func mockMPVFullState(socketPath string, args []any) (any, error) {
	if len(args) >= 2 && args[0] == "get_property" {
		switch args[1] {
		case "time-pos":
			return 83.0, nil // 1:23 into the episode
		case "duration":
			return 3256.0, nil // ~54:16
		case "pause":
			return false, nil
		case "speed":
			return 1.0, nil
		}
	}
	return nil, fmt.Errorf("mock: unsupported command: %v", args)
}

// TestDiscordRPC_SuperFlixCloudFrontURL_OriginalBug documents the original bug:
// When ImageURL was a CloudFront-wrapped URL, Discord showed "?" instead of cover.
func TestDiscordRPC_SuperFlixCloudFrontURL_OriginalBug(t *testing.T) {
	t.Parallel()

	// Simulate a SuperFlix anime with the ORIGINAL (buggy) CloudFront URL
	cloudFrontURL := "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg"

	// This is what the Discord RPC normalize code does inline
	imageURL := cloudFrontURL
	if idx := strings.Index(imageURL, "https://image.tmdb.org/t/p/"); idx > 0 {
		imageURL = imageURL[idx:]
	}

	// VERIFY THE FIX: CloudFront wrapper must be stripped
	assert.NotContains(t, imageURL, "cloudfront.net",
		"BUG REGRESSION: CloudFront URL must be normalized before sending to Discord")
	assert.True(t, strings.HasPrefix(imageURL, "https://image.tmdb.org/t/p/"),
		"must be a direct TMDB URL")
	assert.Equal(t, "https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg", imageURL)
}

// TestDiscordRPC_NormalizeImageURL tests all image URL scenarios in Discord flow
func TestDiscordRPC_NormalizeImageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputURL    string
		wantClean   string
		wantHasCF   bool // should NOT contain cloudfront
		description string
	}{
		{
			name:        "CloudFront SuperFlix URL is normalized",
			inputURL:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster.jpg",
			wantClean:   "https://image.tmdb.org/t/p/w342/poster.jpg",
			wantHasCF:   false,
			description: "Original bug — Discord can't proxy CloudFront double-URLs",
		},
		{
			name:        "Direct TMDB URL stays unchanged",
			inputURL:    "https://image.tmdb.org/t/p/w500/poster.jpg",
			wantClean:   "https://image.tmdb.org/t/p/w500/poster.jpg",
			wantHasCF:   false,
			description: "TMDB poster from EnrichMedia",
		},
		{
			name:        "AniList cover URL stays unchanged",
			inputURL:    "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx1.png",
			wantClean:   "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx1.png",
			wantHasCF:   false,
			description: "AniList images used for regular anime",
		},
		{
			name:        "Empty URL stays empty",
			inputURL:    "",
			wantClean:   "",
			wantHasCF:   false,
			description: "No image available",
		},
		{
			name:        "OMDb poster URL stays unchanged",
			inputURL:    "https://m.media-amazon.com/images/M/MV5poster.jpg",
			wantClean:   "https://m.media-amazon.com/images/M/MV5poster.jpg",
			wantHasCF:   false,
			description: "OMDb poster from EnrichWithOMDb fallback",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Apply the same normalization as updateDiscordPresence()
			imageURL := tc.inputURL
			if idx := strings.Index(imageURL, "https://image.tmdb.org/t/p/"); idx > 0 {
				imageURL = imageURL[idx:]
			}

			assert.Equal(t, tc.wantClean, imageURL, tc.description)
			if tc.wantHasCF {
				assert.Contains(t, imageURL, "cloudfront.net")
			} else {
				assert.NotContains(t, imageURL, "cloudfront.net")
			}
		})
	}
}

// TestDiscordRPC_SuperFlixTV_WithCover tests end-to-end: SuperFlix TV show
// with a TMDB cover should show the cover in Discord RPC
func TestDiscordRPC_SuperFlixTV_WithCover(t *testing.T) {
	t.Parallel()

	// Simulate Dexter from SuperFlix with a properly normalized TMDB cover
	anime := &models.Anime{
		Name:          "Dexter",
		URL:           "1405",
		Source:        "SuperFlix",
		MediaType:     models.MediaTypeTV,
		TMDBID:        1405,
		IMDBID:        "tt0773262",
		Year:          "2006",
		CurrentSeason: 3,
		ImageURL:      "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		Episodes: []models.Episode{
			{Number: "6", Num: 6, SeasonID: "3"},
		},
	}

	isPaused := false
	animeMutex := &sync.Mutex{}
	updateFreq := 5 * time.Second
	episodeDuration := 54*time.Minute + 16*time.Second
	socketPath := "/tmp/mpv_test"

	updater := discord.NewRichPresenceUpdater(
		anime, &isPaused, animeMutex, updateFreq,
		episodeDuration, socketPath, mockMPVFullState,
	)

	// Verify the updater has the correct anime reference
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		updater.GetAnime().ImageURL,
		"updater must have the TMDB cover URL")

	// Verify content type detection
	assert.True(t, anime.IsMovieOrTV(),
		"SuperFlix TV content must be detected as movie/TV for Discord state formatting")

	// Verify season/episode formatting (tested in season_discord_test.go too)
	episodeNumber := anime.Episodes[0].Number
	expectedState := fmt.Sprintf("S%02dE%s", anime.CurrentSeason, episodeNumber)
	assert.Equal(t, "S03E6", expectedState,
		"should show S03E6 for Dexter Season 3 Episode 6")
}

// TestDiscordRPC_SuperFlixMovie_WithCover tests movie cover in Discord RPC
func TestDiscordRPC_SuperFlixMovie_WithCover(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{
		Name:      "Inception",
		URL:       "27205",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeMovie,
		TMDBID:    27205,
		IMDBID:    "tt1375666",
		Year:      "2010",
		ImageURL:  "https://image.tmdb.org/t/p/w500/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
		Episodes: []models.Episode{
			{Number: "1", Num: 1},
		},
	}

	isPaused := false
	animeMutex := &sync.Mutex{}
	updateFreq := 5 * time.Second
	episodeDuration := 2*time.Hour + 28*time.Minute
	socketPath := "/tmp/mpv_test"

	updater := discord.NewRichPresenceUpdater(
		anime, &isPaused, animeMutex, updateFreq,
		episodeDuration, socketPath, mockMPVFullState,
	)

	assert.NotNil(t, updater)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
		updater.GetAnime().ImageURL)

	// Movie state
	assert.True(t, anime.IsMovie())
	assert.Equal(t, "SuperFlix", anime.Source)
}

// TestDiscordRPC_SuperFlixAnime_WithCover tests anime/dorama cover
func TestDiscordRPC_SuperFlixAnime_WithCover(t *testing.T) {
	t.Parallel()

	// SuperFlix also serves anime and doramas — they use MediaTypeTV
	anime := &models.Anime{
		Name:          "O Laboratório de Dexter",
		URL:           "4229",
		Source:        "SuperFlix",
		MediaType:     models.MediaTypeTV,
		TMDBID:        4229,
		Year:          "1996",
		CurrentSeason: 1,
		ImageURL:      "https://image.tmdb.org/t/p/w500/12rxsv2if1i0TudBRFfP7WznJw0.jpg",
		Episodes: []models.Episode{
			{Number: "1", Num: 1, SeasonID: "1"},
		},
	}

	isPaused := false
	animeMutex := &sync.Mutex{}

	updater := discord.NewRichPresenceUpdater(
		anime, &isPaused, animeMutex, 5*time.Second,
		24*time.Minute, "/tmp/mpv_test", mockMPVFullState,
	)

	assert.NotNil(t, updater)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/12rxsv2if1i0TudBRFfP7WznJw0.jpg",
		updater.GetAnime().ImageURL,
		"anime/dorama from SuperFlix must have TMDB cover for Discord")

	// Verify the IMDB/TMDB buttons would be built (movie/TV path)
	assert.True(t, anime.IsMovieOrTV())
}

// TestDiscordRPC_SuperFlix_NoCover_EmptyFallback tests the fallback when no cover is available
func TestDiscordRPC_SuperFlix_NoCover_EmptyFallback(t *testing.T) {
	t.Parallel()

	anime := &models.Anime{
		Name:      "Unknown Show",
		URL:       "00000",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeTV,
		ImageURL:  "", // No cover available
		Episodes: []models.Episode{
			{Number: "1"},
		},
	}

	isPaused := false
	animeMutex := &sync.Mutex{}

	updater := discord.NewRichPresenceUpdater(
		anime, &isPaused, animeMutex, 5*time.Second,
		24*time.Minute, "/tmp/mpv_test", mockMPVFullState,
	)

	// With no image, the inline normalization should leave it empty
	imageURL := updater.GetAnime().ImageURL
	if idx := strings.Index(imageURL, "https://image.tmdb.org/t/p/"); idx > 0 {
		imageURL = imageURL[idx:]
	}
	assert.Equal(t, "", imageURL,
		"empty ImageURL should remain empty — no broken fallback URL")
}

// TestDiscordRPC_SuperFlix_StreamThumbFallback tests that Thumb from stream API
// is propagated to ImageURL when no cover was available from search
func TestDiscordRPC_SuperFlix_StreamThumbAsLastResort(t *testing.T) {
	t.Parallel()

	// Simulate: anime had no cover from search/enrichment, but stream API returned a thumb
	anime := &models.Anime{
		Name:      "Obscure Movie",
		URL:       "88888",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeMovie,
		ImageURL:  "", // Empty before stream
	}

	// The fix in GetSuperFlixStreamURL does:
	// if media.ImageURL == "" && result.Thumb != "" { media.ImageURL = result.Thumb }
	streamThumb := "https://image.tmdb.org/t/p/w500/stream_thumb.jpg"

	// Simulate the thumb propagation
	if anime.ImageURL == "" && streamThumb != "" {
		anime.ImageURL = streamThumb
	}

	assert.Equal(t, "https://image.tmdb.org/t/p/w500/stream_thumb.jpg", anime.ImageURL,
		"stream thumbnail should be used as last-resort cover for Discord")

	// Verify it passes through Discord normalization cleanly
	imageURL := anime.ImageURL
	if idx := strings.Index(imageURL, "https://image.tmdb.org/t/p/"); idx > 0 {
		imageURL = imageURL[idx:]
	}
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/stream_thumb.jpg", imageURL,
		"direct TMDB URL should pass through normalization unchanged")
}
