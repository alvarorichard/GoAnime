package providers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/offlinecache"
)

const testSource = "animefire"

// fakeChecker is a UnifiedScraper that can also answer IsTitleOffline.
type fakeChecker struct {
	offline map[string]bool
	delay   map[string]time.Duration
	probes  atomic.Int32
}

func (f *fakeChecker) SearchAnime(string, ...any) ([]*models.Anime, error) { return nil, nil }
func (f *fakeChecker) GetAnimeEpisodes(string) ([]models.Episode, error)   { return nil, nil }
func (f *fakeChecker) GetStreamURL(string, ...any) (string, map[string]string, error) {
	return "", nil, nil
}
func (f *fakeChecker) GetType() scraper.ScraperType { return scraper.AnimefireType }
func (f *fakeChecker) IsTitleOffline(u string) bool {
	f.probes.Add(1)
	time.Sleep(f.delay[u])
	return f.offline[u]
}

// noCheckScraper has no IsTitleOffline, so filtering must be a no-op.
type noCheckScraper struct{ fakeChecker }

func (*noCheckScraper) IsTitleOffline() {} // different signature: does not satisfy OfflineChecker

func titles(urls ...string) []*models.Anime {
	out := make([]*models.Anime, len(urls))
	for i, u := range urls {
		out[i] = &models.Anime{Name: u, URL: u}
	}
	return out
}

func urlsOf(as []*models.Anime) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = a.URL
	}
	return out
}

func newTestCache(t *testing.T) *offlinecache.Cache {
	t.Helper()
	return offlinecache.New(t.TempDir()+"/offline.json", time.Hour)
}

func TestDropOfflineTitles_RemovesOnlyOffline(t *testing.T) {
	sc := &fakeChecker{offline: map[string]bool{"dead": true}}
	got := urlsOf(dropOfflineTitles(context.Background(), newTestCache(t), testSource, sc, titles("a", "dead", "b")))
	want := []string{"a", "b"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestDropOfflineTitles_AllOfflineKeepsOriginal(t *testing.T) {
	sc := &fakeChecker{offline: map[string]bool{"x": true, "y": true}}
	got := dropOfflineTitles(context.Background(), newTestCache(t), testSource, sc, titles("x", "y"))
	if len(got) != 2 {
		t.Fatalf("expected original list when everything is offline, got %d", len(got))
	}
}

func TestDropOfflineTitles_NoCheckerIsNoop(t *testing.T) {
	got := dropOfflineTitles(context.Background(), newTestCache(t), testSource, &noCheckScraper{}, titles("a", "b"))
	if len(got) != 2 {
		t.Fatalf("expected no-op, got %d results", len(got))
	}
}

func TestDropOfflineTitles_TimeoutKeepsTitle(t *testing.T) {
	// A check slower than the budget must not drop the title.
	sc := &fakeChecker{
		offline: map[string]bool{"slow": true},
		delay:   map[string]time.Duration{"slow": offlineProbeTimeout + 2*time.Second},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	got := dropOfflineTitles(ctx, newTestCache(t), testSource, sc, titles("slow", "ok"))
	if len(got) != 2 {
		t.Fatalf("slow check should keep the title, got %v", urlsOf(got))
	}
}

func TestDropOfflineTitles_CacheSkipsProbeOnSecondSearch(t *testing.T) {
	cache := newTestCache(t)
	sc := &fakeChecker{offline: map[string]bool{"dead": true}}

	dropOfflineTitles(context.Background(), cache, testSource, sc, titles("a", "dead"))
	if got := sc.probes.Load(); got != 2 {
		t.Fatalf("first search should probe both titles, probed %d", got)
	}

	got := urlsOf(dropOfflineTitles(context.Background(), cache, testSource, sc, titles("a", "dead")))
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("cached title should be dropped, got %v", got)
	}
	if probes := sc.probes.Load(); probes != 3 {
		t.Fatalf("second search should only probe the uncached title (3 total), got %d", probes)
	}
}

func TestDropOfflineTitles_CacheIsPerSource(t *testing.T) {
	cache := newTestCache(t)
	cache.MarkOffline("goyabu", "dead")
	sc := &fakeChecker{}
	got := dropOfflineTitles(context.Background(), cache, testSource, sc, titles("dead", "ok"))
	if len(got) != 2 {
		t.Fatalf("a verdict for another source must not filter this one, got %v", urlsOf(got))
	}
}

func TestDropOfflineTitles_DisabledCacheStillFilters(t *testing.T) {
	cache := offlinecache.New("", 0) // ttl 0: disabled
	sc := &fakeChecker{offline: map[string]bool{"dead": true}}
	got := dropOfflineTitles(context.Background(), cache, testSource, sc, titles("a", "dead"))
	if len(got) != 1 {
		t.Fatalf("filter should work without the cache, got %v", urlsOf(got))
	}
	dropOfflineTitles(context.Background(), cache, testSource, sc, titles("a", "dead"))
	if probes := sc.probes.Load(); probes != 4 {
		t.Fatalf("disabled cache must not skip probes, probed %d", probes)
	}
}
