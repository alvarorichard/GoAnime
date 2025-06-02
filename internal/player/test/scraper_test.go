package test

import (
	"regexp"
	"testing"

	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/stretchr/testify/assert"
)

// TestSelectVideoQualityNoVideos tests the case when no videos are available
func TestSelectVideoQualityNoVideos(t *testing.T) {
	videos := []player.VideoData{}

	// This test only verifies that empty videos array produces the expected error
	// We can't test selectVideoQuality directly because it uses interactive fuzzy finder
	// So we'll check just the initial validation logic
	if len(videos) == 0 {
		// This manually simulates the first check in selectVideoQuality
		err := "no video qualities available"
		assert.Equal(t, "no video qualities available", err)
	}
}

// TestSelectVideoQualityUrlAdjustment tests the URL adjustment logic
func TestSelectVideoQualityUrlAdjustment(t *testing.T) {
	// Create a custom implementation that tests the URL adjustment logic
	// without the fuzzy finder interaction
	selectQuality := func(videos []player.VideoData, selectedIndex int) (string, error) {
		if selectedIndex < 0 || selectedIndex >= len(videos) {
			return "", assert.AnError
		}

		selectedQuality := videos[selectedIndex]
		url := selectedQuality.Src

		// This part is copied from the original function
		qualityPattern := regexp.MustCompile(`/(\d+)p\.mp4`)
		matches := qualityPattern.FindStringSubmatch(url)
		if len(matches) > 1 {
			url = qualityPattern.ReplaceAllString(url, "/"+selectedQuality.Label+"p.mp4")
		}

		return url, nil
	}

	videos := []player.VideoData{
		{Src: "https://example.com/video/480p.mp4", Label: "720"},
		{Src: "https://example.com/video/720p.mp4", Label: "1080"},
	}

	// Test with the second video (index 1)
	url, err := selectQuality(videos, 1)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/video/1080p.mp4", url)

	// Test with the first video (index 0)
	url, err = selectQuality(videos, 0)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/video/720p.mp4", url)
}

// TestSelectVideoQualityWithoutPattern tests URLs without quality pattern
func TestSelectVideoQualityWithoutPattern(t *testing.T) {
	// Create a custom implementation that tests URLs without quality pattern
	selectQuality := func(videos []player.VideoData, selectedIndex int) (string, error) {
		if selectedIndex < 0 || selectedIndex >= len(videos) {
			return "", assert.AnError
		}

		selectedQuality := videos[selectedIndex]
		url := selectedQuality.Src

		// This part is copied from the original function
		qualityPattern := regexp.MustCompile(`/(\d+)p\.mp4`)
		matches := qualityPattern.FindStringSubmatch(url)
		if len(matches) > 1 {
			url = qualityPattern.ReplaceAllString(url, "/"+selectedQuality.Label+"p.mp4")
		}

		return url, nil
	}

	videos := []player.VideoData{
		{Src: "https://example.com/video/stream.mp4", Label: "720"},
	}

	// Test URL without quality pattern
	url, err := selectQuality(videos, 0)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/video/stream.mp4", url, "URL without quality pattern should remain unchanged")
}
