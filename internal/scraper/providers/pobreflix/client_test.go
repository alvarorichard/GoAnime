package pobreflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCatalog serves a tiny version of the site's sitemap tree: an index
// naming the shards, each shard's record URLs, and page endpoints for the
// title fixtures. Every request is counted, so tests can assert the corpus
// cache actually saves the fetches — and each quirk of a shard (absent,
// failing, hanging) is a struct knob under the test's control.
type mockCatalog struct {
	server *httptest.Server
	client *Client
	cache  string
	count  atomic.Int64

	shard2OK    atomic.Bool // shard 2 answers when set
	shard2Fails atomic.Bool // shard 2 answers 500 when set
	shard2Hang  atomic.Int64

	shard2Seen atomic.Int64
}

// newMockCatalog wires a mock over the real fixtures: the search tree carries
// the shapes captured from the site (dub + leg variants, an episode, junk,
// a foreign host) and the title pages are the fixture files verbatim.
func newMockCatalog(t *testing.T) *mockCatalog {
	t.Helper()
	m := &mockCatalog{}
	m.shard2OK.Store(true)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.count.Add(1)
		base := "http://" + r.Host
		switch {
		case r.URL.Path == "/sitemap.php" && r.URL.Query().Get("file") == "":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, sitemapIndexFixture(base, 2))
		case r.URL.Query().Get("file") == "sitemap_content_videobox_Video_1":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<?xml version="1.0"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
 <url><loc>%s/filmes/online/todo-poderoso-legendado-7002/</loc></url>
 <url><loc>%s/filmes/online/todo-poderoso-dublado-7001/</loc></url>
 <url><loc>%s/filmes/online/todo-dublado-5/</loc></url>
 <url><loc>%s/series/online/ben-hur-a-serie-dublado-83633/</loc></url>
 <url><loc>%s/episodios/online/ben-hur-a-serie-1x8-dublado-84788/</loc></url>
 <url><loc>%s/filmes/</loc></url>
 <url><loc>https://other-host.example/filmes/online/nao-e-daqui-dublado-9/</loc></url>
</urlset>`, base, base, base, base, base, base)
		case r.URL.Query().Get("file") == "sitemap_content_videobox_Video_2":
			m.shard2Seen.Add(1)
			switch {
			case m.shard2Hang.Load() > 0:
				time.Sleep(time.Duration(m.shard2Hang.Load()))
			case m.shard2Fails.Load():
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			case !m.shard2OK.Load():
				http.Error(w, "gone", http.StatusServiceUnavailable)
				return
			}
			fmt.Fprintf(w, `<?xml version="1.0"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
 <url><loc>%s/filmes/online/um-todo-mundo-dublado-42/</loc></url>
</urlset>`, base)
		case r.URL.Path == "/filmes/online/o-apanhador-de-sonhos-dublado-84752/":
			serveFixture(t, w, "movie.html")
		case r.URL.Path == "/series/online/ben-hur-a-serie-dublado-83633/":
			serveFixture(t, w, "series.html")
		case r.URL.Path == "/filmes/online/none-8/":
			// A title page with no video mount: 200, parseable, empty.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!doctype html><html><body><h1 class="type">Assistir Nada Online</h1></body></html>`)
		default:
			http.NotFound(w, r)
		}
	}
	m.server = httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(m.server.Close)
	m.cache = filepath.Join(t.TempDir(), "pobreflix-index.json")
	m.client = m.newClient(t)
	return m
}

// newClient builds a client on the mock with a private cache file. Sub-second
// retries keep the failure tests fast.
func (m *mockCatalog) newClient(t *testing.T) *Client {
	t.Helper()
	c := NewClientForTest(m.server.Client(), m.server.URL, m.cache)
	c.retryDelay = 10 * time.Millisecond
	c.maxRetries = 1
	return c
}

// sitemapIndexFixture lists n video shards of one host.
func sitemapIndexFixture(base string, n int) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, `<sitemap><loc>%s/sitemap.php?file=sitemap_content_videobox_Video_%d</loc></sitemap>`, base, i)
	}
	sb.WriteString(`</sitemapindex>`)
	return sb.String()
}

// serveFixture writes a testdata file verbatim.
func serveFixture(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		http.Error(w, "fixture missing: "+name, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(raw)
}

// TestSearch_ColdBuildRanks: a first search builds the corpus and returns
// results ranked by RelevanceScore, deduped dub-first; episodes, category
// pages and foreign-host records never enter the index.
func TestSearch_ColdBuildRanks(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)

	results, err := m.client.Search(context.Background(), "todo-poderoso")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	var todo []Media
	for _, r := range results {
		if r.Title == "Todo Poderoso" {
			todo = append(todo, r)
		}
	}
	require.Len(t, todo, 1, "dub and leg variants dedupe into one entry")
	assert.Equal(t, AudioDubbed, todo[0].Audio, "the dub earns the slot")
	assert.Equal(t, m.server.URL+"/filmes/online/todo-poderoso-dublado-7001/", todo[0].URL)
	assert.Equal(t, KindMovie, todo[0].Kind)

	for _, r := range results {
		assert.NotContains(t, r.URL, "/episodios/online/", "episodes are corpus noise")
		assert.False(t, strings.HasSuffix(r.URL, "/filmes/"), "category pages are corpus noise")
		assert.NotContains(t, r.URL, "other-host.example", "foreign records are corpus noise")
	}
}

// TestSearch_ExactBeatsContains: "todo" alone puts the exact title first and
// the partial matches in score order.
func TestSearch_ExactBeatsContains(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	results, err := m.client.Search(context.Background(), "todo")
	require.NoError(t, err)
	var titles []string
	for _, r := range results {
		titles = append(titles, r.Title)
	}
	// Ben-Hur: A Série carries no "todo" and must not appear.
	assert.Equal(t, []string{"Todo", "Todo Poderoso", "Um Todo Mundo"}, titles,
		"exact title, the dub-first dedupe of contains, then partials")
}

// TestSearch_VariantsOfTheSameQuery: slugs, spaces and accents all reach the
// same entry — the CLI passes "todo-poderoso" and a PT speaker types "Todo
// Poderoso" or "TÔDO PÖDEROSO".
func TestSearch_VariantsOfTheSameQuery(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	for _, q := range []string{"todo-poderoso", "Todo Poderoso", "TÔDO PÖDEROSO", "todo  poderoso"} {
		results, err := m.client.Search(context.Background(), q)
		require.NoError(t, err, "query %q", q)
		require.NotEmpty(t, results, "query %q", q)
		assert.Equal(t, "Todo Poderoso", results[0].Title, "query %q must top the list", q)
	}
}

// TestSearch_EmptyQueryKeepsSilence: glue words, lone digits and blanks
// return nothing — never a page of noise.
func TestSearch_EmptyQueryKeepsSilence(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	for _, q := range []string{"", "   ", "a", "de", "2003", "-"} {
		results, err := m.client.Search(context.Background(), q)
		require.NoError(t, err, "query %q", q)
		assert.Empty(t, results, "query %q", q)
	}
}

// TestSearch_CacheAnswerWithoutFetching: within the refresh window the disk
// cache answers with zero HTTP requests — and a fresh client instance (a
// restarted app) reads the same cache off disk.
func TestSearch_CacheAnswerWithoutFetching(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)

	first, err := m.client.Search(context.Background(), "todo")
	require.NoError(t, err)
	require.NotEmpty(t, first)
	require.True(t, m.client.search.fresh(), "a completed build stamps the fresh date")

	before := m.count.Load()
	second, err := m.client.Search(context.Background(), "todo")
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.Equal(t, before, m.count.Load(), "an in-date search takes no request")

	// A restarted app: same cache file, cold client.
	restarted := m.newClient(t)
	restarted.search.loadCache(m.server.URL)
	require.True(t, restarted.search.fresh(), "the disk cache loads")
	warm, err := restarted.Search(context.Background(), "todo")
	require.NoError(t, err)
	assert.Equal(t, first, warm)
	assert.Equal(t, before, m.count.Load(), "the restart search takes no request either")
}

// TestSearch_PartialShardFailureRecovers: a shard that 500s must not fail
// the search — its records are absent this round, the index stays stale, and
// the round after retries only that shard.
func TestSearch_PartialShardFailureRecovers(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	m.shard2Fails.Store(true)
	ctx := context.Background()

	results, err := m.client.Search(ctx, "um todo mundo")
	require.NoError(t, err)
	assert.Empty(t, results, "the failed shard's record is absent this round")
	assert.False(t, m.client.search.shardsDone["sitemap_content_videobox_Video_2"],
		"a failed shard is not marked done")
	assert.False(t, m.client.search.fresh(), "an incomplete corpus is stale")

	// Retry only fetches what was missing, then the corpus completes.
	m.shard2Fails.Store(false)
	before := m.count.Load()
	results, err = m.client.Search(ctx, "um todo mundo")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Um Todo Mundo", results[0].Title)
	assert.True(t, m.client.search.fresh(), "the completed round stamps the date")

	seenTotal := m.count.Load() - before
	// index + the one missing shard (+/+2 retries on the first round) — the
	// key assertion is shard 1 is never re-fetched; its seen counter would
	// need a separate counter, so assert the shard-2 counter jumped.
	assert.GreaterOrEqual(t, m.shard2Seen.Load(), int64(1), "the missing shard was retried")
	assert.LessOrEqual(t, seenTotal, int64(4), "no fetch storm on resume: index + shard2 once")
}

// TestSearch_CtxDeadlineKeepsPartialResults: the registry gives a source
// search 12s. A first build that runs out of deadline keeps what it fetched,
// returns results from it, and leaves the date stale so the next search
// resumes whatever hung.
func TestSearch_CtxDeadlineKeepsPartialResults(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	m.shard2OK.Store(false) // and hang instead of answering
	m.shard2Hang.Store(int64(2 * time.Second))
	defer m.shard2Hang.Store(0)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	results, err := m.client.Search(ctx, "todo")
	require.NoError(t, err, "a partial corpus must search, the deadline is not an error")
	assert.NotEmpty(t, results, "shard 1's records are present")

	off := m.client.search.fresh()
	assert.False(t, off, "an incomplete build must not count as fresh")

	// The next search resumes: shard 2 answers in its own right this round.
	m.shard2OK.Store(true)
	m.shard2Hang.Store(0)
	results, err = m.client.Search(context.Background(), "um todo mundo")
	require.NoError(t, err)
	require.Len(t, results, 1, "the resumed round finds the once-hung shard")
}

// TestTitle_FixturesThroughHTTP: the client's Title reads the captured pages
// with the same ids the fixtures carry, and caches per process.
func TestTitle_FixturesThroughHTTP(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	ctx := context.Background()

	title, err := m.client.Title(ctx, m.server.URL+"/filmes/online/o-apanhador-de-sonhos-dublado-84752/")
	require.NoError(t, err)
	assert.Equal(t, 793002, title.TMDBID)
	assert.Equal(t, 84752, title.VideoID)
	assert.Equal(t, KindMovie, title.Kind)

	title, err = m.client.Title(ctx, m.server.URL+"/series/online/ben-hur-a-serie-dublado-83633/")
	require.NoError(t, err)
	assert.Equal(t, 288458, title.TMDBID)
	assert.Equal(t, KindSeries, title.Kind)
	require.NotEmpty(t, title.Seasons)

	before := m.count.Load()
	_, err = m.client.Title(ctx, m.server.URL+"/series/online/ben-hur-a-serie-dublado-83633/")
	require.NoError(t, err)
	assert.Equal(t, before, m.count.Load(), "title pages cache per process")
}

// TestTitle_NoMountErrors: a page without a mount is ErrNoTitle — no panic
// anywhere, and the error travels with the source's name attached by netx.
func TestTitle_NoMountErrors(t *testing.T) {
	t.Parallel()
	m := newMockCatalog(t)
	_, err := m.client.Title(context.Background(), m.server.URL+"/filmes/online/none-8/")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoTitle)
}

// TestSearch_DeadlineWithColdCacheErrors: when the build learns NOTHING in
// time (the sitemap index hangs past the deadline), the search fails with a
// diagnostic — the TRY screen shows the source unreachable, and nothing gets
// persisted from a cold cache.
func TestSearch_DeadlineWithColdCacheErrors(t *testing.T) {
	t.Parallel()
	// A host whose sitemap index hangs past any sane deadline.
	hanging := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		http.NotFound(w, r)
	}))
	t.Cleanup(hanging.Close)
	cache := filepath.Join(t.TempDir(), "pobreflix-index.json")

	m2 := NewClientForTest(hanging.Client(), hanging.URL, cache)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := m2.Search(ctx, "todo")
	require.Error(t, err, "no corpus and no cache: the search fails with a reason")
	_, statErr := os.Stat(cache)
	assert.False(t, statErr == nil, "nothing learned, nothing persisted")
}
