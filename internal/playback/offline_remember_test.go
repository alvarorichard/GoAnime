package playback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/offlinecache"
)

func TestRememberRemovedTitle(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	removed := fmt.Errorf("x: %w", player.ErrSourceVideoRemoved)

	rememberRemovedTitle(&models.Anime{URL: "https://animefire/dead", Source: "Animefire.io"}, removed)
	rememberRemovedTitle(&models.Anime{URL: "https://animefire/net", Source: "Animefire.io"}, errors.New("timeout"))
	rememberRemovedTitle(&models.Anime{URL: "https://goyabu/dead", Source: "Goyabu"}, removed)

	c := offlinecache.Shared()
	if !isHidden(c, "AnimeFire", "https://animefire/dead") {
		t.Error("removed AnimeFire title not recorded")
	}
	if isHidden(c, "AnimeFire", "https://animefire/net") {
		t.Error("a non-removal error must not hide the title")
	}
	if isHidden(c, "AnimeFire", "https://goyabu/dead") || isHidden(c, "Goyabu", "https://goyabu/dead") {
		t.Error("only AnimeFire titles are recorded")
	}
}

func isHidden(c *offlinecache.Cache, source, url string) bool {
	return c.Set(source, []string{url})[url]
}

// The two ends of the feature together: a title playback records as removed is
// the same title the next AnimeFire search leaves out.
func TestRememberedRemovedTitleIsHiddenFromTheNextSearch(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv(offlinecache.EnvHide, "")
	removed := fmt.Errorf("x: %w", player.ErrSourceVideoRemoved)

	rememberRemovedTitle(&models.Anime{URL: "https://animefire/dead", Source: "Animefire.io"}, removed)
	rememberRemovedTitle(&models.Anime{URL: "https://animefire/net", Source: "Animefire.io"}, errors.New("timeout"))

	results := []*models.Anime{
		{URL: "https://animefire/ok", Name: "ok"},
		{URL: "https://animefire/dead", Name: "dead"},
		{URL: "https://animefire/net", Name: "net"},
	}
	got := providers.HideOfflineTitles(offlinecache.Shared(), "AnimeFire", results)
	if len(got) != 2 || got[0].URL != "https://animefire/ok" || got[1].URL != "https://animefire/net" {
		t.Fatalf("got %v, want only the removed title hidden", got)
	}
}
