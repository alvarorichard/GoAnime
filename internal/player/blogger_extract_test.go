package player

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

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

func isTransientBloggerExtractError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	transient := []string{
		"no such host",
		"timeout",
		"temporary failure",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"tls handshake timeout",
		"server returned",
		"status 403",
		"status 429",
		"status 500",
		"status 502",
		"status 503",
		"status 521",
		"status 522",
		"status 523",
		"status 524",
		"status 530",
	}
	for _, marker := range transient {
		if strings.Contains(errMsg, marker) {
			return true
		}
	}
	return false
}
