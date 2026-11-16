package playback

import (
	"errors"
	"fmt"
	"testing"

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
	if !c.IsOffline("AnimeFire", "https://animefire/dead") {
		t.Error("removed AnimeFire title not recorded")
	}
	if c.IsOffline("AnimeFire", "https://animefire/net") {
		t.Error("a non-removal error must not hide the title")
	}
	if c.IsOffline("AnimeFire", "https://goyabu/dead") || c.IsOffline("Goyabu", "https://goyabu/dead") {
		t.Error("only AnimeFire titles are recorded")
	}
}
