package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestRememberDownloadRefererSurvivesAGlobalChange(t *testing.T) {
	t.Cleanup(util.ClearGlobalReferer)
	util.SetGlobalReferer("https://a.example/")
	forget := rememberDownloadReferer("https://cdn.example/a.mp4", "https://a.example/")

	// Another episode is resolved while this one still downloads.
	util.SetGlobalReferer("https://b.example/")

	assert.Equal(t, "https://a.example/", downloadReferer("https://cdn.example/a.mp4"), "the pinned download keeps its own referer")
	assert.Equal(t, "https://b.example/", downloadReferer("https://cdn.example/other.mp4"), "an unpinned URL still follows the global one")

	forget()
	assert.Equal(t, "https://b.example/", downloadReferer("https://cdn.example/a.mp4"), "forgetting hands the URL back to the global referer")
}
