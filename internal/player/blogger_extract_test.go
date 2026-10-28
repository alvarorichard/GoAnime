//go:build live

package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// TestExtractBloggerGoogleVideoURL checks Google's current session and RPC
// response in the daily source-health workflow; regular CI uses local fixtures.
func TestExtractBloggerGoogleVideoURL(t *testing.T) {
	util.IsDebug = true

	bloggerURL := "https://www.blogger.com/video.g?token=AD6v5dykZRdbBj2paRaH29--_CInECmFwGsPegJF3f5cm_obBJGj1rx32yoCHk9Iv0VZOkkbcZKo3CPzhBwS2OmXSfSDhwCSLWsSnReWdIkgMxMXfQ_IAc98xObMWc1NKMLuB7hkS7w-"

	videoURL, err := extractBloggerGoogleVideoURL(bloggerURL)
	if err != nil {
		if isTransientBloggerExtractError(err) {
			t.Skipf("Blogger source unavailable in test environment: %v", err)
		}
		t.Fatalf("extractBloggerGoogleVideoURL failed: %v", err)
	}
	t.Logf("Extracted video URL: %s", videoURL)
}
