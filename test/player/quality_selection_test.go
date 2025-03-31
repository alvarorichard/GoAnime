package player_test

import (
	"regexp"
	"testing"

	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/stretchr/testify/assert"
)

// TestVideoQualityPatternMatching tests the regex pattern used in selectVideoQuality
func TestVideoQualityPatternMatching(t *testing.T) {
	// Test cases for URL quality pattern matching
	testCases := []struct {
		name          string
		url           string
		shouldMatch   bool
		expectedGroup string
	}{
		{
			name:          "Standard 480p quality",
			url:           "https://example.com/video/480p.mp4",
			shouldMatch:   true,
			expectedGroup: "480",
		},
		{
			name:          "Standard 720p quality",
			url:           "https://example.com/video/720p.mp4",
			shouldMatch:   true,
			expectedGroup: "720",
		},
		{
			name:          "Standard 1080p quality",
			url:           "https://example.com/video/1080p.mp4",
			shouldMatch:   true,
			expectedGroup: "1080",
		},
		{
			name:        "No quality in URL",
			url:         "https://example.com/video/stream.mp4",
			shouldMatch: false,
		},
		{
			name:        "Non-numeric quality",
			url:         "https://example.com/video/HD.mp4",
			shouldMatch: false,
		},
	}

	// Compile the same regex pattern used in selectVideoQuality
	qualityPattern := regexp.MustCompile(`/(\d+)p\.mp4`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := qualityPattern.FindStringSubmatch(tc.url)

			if tc.shouldMatch {
				assert.Greater(t, len(matches), 1, "URL should match the quality pattern")
				if len(matches) > 1 {
					assert.Equal(t, tc.expectedGroup, matches[1], "Extracted quality should match expected")
				}
			} else {
				assert.LessOrEqual(t, len(matches), 1, "URL should not match the quality pattern")
			}
		})
	}
}

// TestVideoQualityReplacement tests the URL replacement logic used in selectVideoQuality
func TestVideoQualityReplacement(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		quality     string
		expectedUrl string
	}{
		{
			name:        "Replace 480p with 720p",
			url:         "https://example.com/video/480p.mp4",
			quality:     "720",
			expectedUrl: "https://example.com/video/720p.mp4",
		},
		{
			name:        "Replace 720p with 1080p",
			url:         "https://example.com/video/720p.mp4",
			quality:     "1080",
			expectedUrl: "https://example.com/video/1080p.mp4",
		},
		{
			name:        "Replace 1080p with 480p",
			url:         "https://example.com/video/1080p.mp4",
			quality:     "480",
			expectedUrl: "https://example.com/video/480p.mp4",
		},
	}

	qualityPattern := regexp.MustCompile(`/(\d+)p\.mp4`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is the same replacement logic used in selectVideoQuality
			result := qualityPattern.ReplaceAllString(tc.url, "/"+tc.quality+"p.mp4")
			assert.Equal(t, tc.expectedUrl, result)
		})
	}
}

// TestCreateVideoDataStructure tests creating VideoData structures
func TestCreateVideoDataStructure(t *testing.T) {
	// Test that we can create VideoData objects correctly
	videoData := player.VideoData{
		Src:   "https://example.com/video/720p.mp4",
		Label: "720",
	}

	assert.Equal(t, "https://example.com/video/720p.mp4", videoData.Src)
	assert.Equal(t, "720", videoData.Label)
}
