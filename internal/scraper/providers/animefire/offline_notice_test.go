package animefire

import (
	"net/http"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #229: Wind Breaker is listed with 25 episodes and not one has a file.
// GoAnime knew (the first and middle episodes are both offline) and said so,
// but the warning was printed before the episode list opened and the list
// covered it. The warning now travels to the list as a notice.

// noticeLog collects what a client tells the user about the list it returns.
type noticeLog struct {
	mu   sync.Mutex
	said []string
}

func (n *noticeLog) add(text string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.said = append(n.said, text)
}

func (n *noticeLog) all() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]string(nil), n.said...)
}

func offlineEpisode() func(http.ResponseWriter) {
	return writeJSON(map[string]any{"data": map[string]any{"streams": []map[string]any{
		{"audio": "dublado", "is_offline": true, "url": nil},
		{"audio": "legendado", "is_offline": true, "url": nil},
	}}})
}

func onlineEpisode(t *testing.T) func(http.ResponseWriter) {
	t.Helper()
	cdn := newCDNFixture(t, cdnTrack{audio: "legendado", status: http.StatusOK})
	return writeJSON(map[string]any{"data": map[string]any{"streams": []map[string]any{
		{"audio": "legendado", "is_offline": false, "url": cdn.url("legendado")},
	}}})
}

func listing(ids ...string) []models.Episode {
	eps := make([]models.Episode, len(ids))
	for i, id := range ids {
		eps[i] = models.Episode{Number: id, Num: i + 1, URL: "https://animefire.one/episode/" + id}
	}
	return eps
}

func TestTitleOfflineNoticeCountsTheEpisodes(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "AnimeFire may have no video files for this title: the first and middle episodes are both offline (25 episodes listed). Press ESC and pick the same title from StartFlix or HiAnime.", titleOfflineNotice(25))
	assert.Contains(t, titleOfflineNotice(2), "(2 episodes listed)")
}

func TestWarnIfTitleLooksOffline_TellsTheListWhenBothProbesAreOffline(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/a": offlineEpisode(),
		"/episode/b": offlineEpisode(),
		"/episode/c": offlineEpisode(),
	})
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	c.warnIfTitleLooksOffline(listing("a", "b", "c"))

	require.Len(t, log.all(), 1, "one notice for the list that opens next")
	assert.Equal(t, titleOfflineNotice(3), log.all()[0])
}

func TestWarnIfTitleLooksOffline_StaysQuietWhenSomethingPlays(t *testing.T) {
	t.Parallel()
	t.Run("the first episode plays", func(t *testing.T) {
		t.Parallel()
		srv := newAPIServer(t, map[string]func(http.ResponseWriter){"/episode/a": onlineEpisode(t), "/episode/b": offlineEpisode()})
		var log noticeLog
		c := NewClientForTest(srv.URL)
		c.notice = log.add

		c.warnIfTitleLooksOffline(listing("a", "b"))

		assert.Empty(t, log.all())
		assert.Equal(t, int32(1), srv.hits.Load(), "a title whose first episode plays costs one probe")
	})
	t.Run("only the middle one plays", func(t *testing.T) {
		t.Parallel()
		srv := newAPIServer(t, map[string]func(http.ResponseWriter){"/episode/a": offlineEpisode(), "/episode/b": onlineEpisode(t), "/episode/c": offlineEpisode()})
		var log noticeLog
		c := NewClientForTest(srv.URL)
		c.notice = log.add

		c.warnIfTitleLooksOffline(listing("a", "b", "c"))

		assert.Empty(t, log.all(), "a single gap is normal and must not condemn a title that mostly works")
	})
}

func TestWarnIfTitleLooksOffline_ProbesTheFirstAndTheMiddleEpisode(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var probed []string
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/e1": func(w http.ResponseWriter) {
			mu.Lock()
			probed = append(probed, "e1")
			mu.Unlock()
			offlineEpisode()(w)
		},
		"/episode/e2": func(w http.ResponseWriter) {
			mu.Lock()
			probed = append(probed, "e2")
			mu.Unlock()
			offlineEpisode()(w)
		},
		"/episode/e3": func(w http.ResponseWriter) {
			mu.Lock()
			probed = append(probed, "e3")
			mu.Unlock()
			offlineEpisode()(w)
		},
		"/episode/e4": func(w http.ResponseWriter) {
			mu.Lock()
			probed = append(probed, "e4")
			mu.Unlock()
			offlineEpisode()(w)
		},
	})
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	c.warnIfTitleLooksOffline(listing("e1", "e2", "e3", "e4"))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"e1", "e3"}, probed, "index 0 and len/2 — two probes, not a sweep")
	assert.Len(t, log.all(), 1)
}

func TestWarnIfTitleLooksOffline_NeedsAtLeastTwoEpisodes(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){"/episode/a": offlineEpisode()})
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	c.warnIfTitleLooksOffline(listing("a"))
	c.warnIfTitleLooksOffline(nil)

	assert.Empty(t, log.all())
	assert.Zero(t, srv.hits.Load(), "nothing is probed")
}

func TestWarnIfTitleLooksOffline_IgnoresAnErrorThatIsNotOffline(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/a": writeStatus(http.StatusInternalServerError, "boom"),
		"/episode/b": offlineEpisode(),
	})
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	c.warnIfTitleLooksOffline(listing("a", "b"))

	assert.Empty(t, log.all(), "a failing API says nothing about the files, so the title is not condemned")
}

func TestWarnIfTitleLooksOffline_WorksWithoutANotifier(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){"/episode/a": offlineEpisode(), "/episode/b": offlineEpisode()})
	c := NewClientForTest(srv.URL)
	c.notice = nil

	assert.NotPanics(t, func() { c.warnIfTitleLooksOffline(listing("a", "b")) })
}

// The path the user took: Wind Breaker's 25 listed episodes, none with a file.
func TestGetAnimeEpisodes_TellsTheListWhenTheWholeTitleIsOffline(t *testing.T) {
	t.Parallel()
	episodes := make([]map[string]any, 25)
	routes := map[string]func(http.ResponseWriter){}
	for i := range episodes {
		id := string(rune('a' + i))
		episodes[i] = map[string]any{"id": id, "title": "Wind Breaker", "season": 1, "number": i + 1}
		routes["/episode/"+id] = offlineEpisode()
	}
	routes["/anime/E0CBXaPxqvY"] = writeJSON(map[string]any{"data": map[string]any{"episodes": episodes}})
	srv := newAPIServer(t, routes)
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	eps, err := c.GetAnimeEpisodes("https://animefire.one/anime/E0CBXaPxqvY")

	require.NoError(t, err)
	assert.Len(t, eps, 25, "the episodes are still listed")
	require.Len(t, log.all(), 1, "and the list opens with the reason")
	assert.Equal(t, titleOfflineNotice(25), log.all()[0])
}

func TestGetAnimeEpisodes_SaysNothingAboutATitleThatPlays(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/anime/ok": writeJSON(map[string]any{"data": map[string]any{"episodes": []map[string]any{
			{"id": "a", "title": "x", "season": 1, "number": 1}, {"id": "b", "title": "y", "season": 1, "number": 2},
		}}}),
		"/episode/a": onlineEpisode(t),
	})
	var log noticeLog
	c := NewClientForTest(srv.URL)
	c.notice = log.add

	_, err := c.GetAnimeEpisodes("https://animefire.one/anime/ok")

	require.NoError(t, err)
	assert.Empty(t, log.all())
}

func TestClientsTellTheListThroughTheRealNotifierByDefault(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, NewAnimefireClient().notice, "without it the warning is printed and then covered by the list")
	assert.NotNil(t, NewClientForTest("http://127.0.0.1:1").notice)
}
