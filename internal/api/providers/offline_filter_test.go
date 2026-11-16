package providers

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/offlinecache"
)

func animes(urls ...string) []*models.Anime {
	out := make([]*models.Anime, len(urls))
	for i, u := range urls {
		out[i] = &models.Anime{URL: u, Name: u}
	}
	return out
}

func TestHideOfflineTitles(t *testing.T) {
	c := offlinecache.New("", time.Hour)
	_ = c.MarkOffline("AnimeFire", "dead")

	got := hideOfflineTitles(c, "AnimeFire", animes("a", "dead", "b"))
	if len(got) != 2 || got[0].URL != "a" || got[1].URL != "b" {
		t.Fatalf("got %v, want a and b", got)
	}
}

func TestHideOfflineTitles_NothingRecordedIsNoOp(t *testing.T) {
	in := animes("a", "b")
	if got := hideOfflineTitles(offlinecache.New("", time.Hour), "AnimeFire", in); len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
}

func TestHideOfflineTitles_AllRecordedKeepsList(t *testing.T) {
	c := offlinecache.New("", time.Hour)
	_ = c.MarkOffline("AnimeFire", "x")
	_ = c.MarkOffline("AnimeFire", "y")
	if got := hideOfflineTitles(c, "AnimeFire", animes("x", "y")); len(got) != 2 {
		t.Fatalf("got %d results, want the full list back", len(got))
	}
}

func TestHideOfflineTitles_EnvSwitchesOff(t *testing.T) {
	t.Setenv(offlinecache.EnvHide, "0")
	c := offlinecache.New("", time.Hour)
	_ = c.MarkOffline("AnimeFire", "dead")
	if got := hideOfflineTitles(c, "AnimeFire", animes("a", "dead")); len(got) != 2 {
		t.Fatalf("got %d results, want both listed", len(got))
	}
}
