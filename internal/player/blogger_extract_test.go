package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

func TestExtractBloggerGoogleVideoURL(t *testing.T) {
	util.IsDebug = true

	bloggerURL := "https://www.blogger.com/video.g?token=AD6v5dykZRdbBj2paRaH29--_CInECmFwGsPegJF3f5cm_obBJGj1rx32yoCHk9Iv0VZOkkbcZKo3CPzhBwS2OmXSfSDhwCSLWsSnReWdIkgMxMXfQ_IAc98xObMWc1NKMLuB7hkS7w-"

	videoURL, err := extractBloggerGoogleVideoURL(bloggerURL)
	if err != nil {
		t.Fatalf("extractBloggerGoogleVideoURL failed: %v", err)
	}
	t.Logf("Extracted video URL: %s", videoURL)
}
