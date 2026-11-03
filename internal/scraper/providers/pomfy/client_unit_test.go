package pomfy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The 0% rows after the flows landed: constructors, the URL builders, the
// small shapers, and the for-test helpers' own round trip.

func TestNewClientEnvBase(t *testing.T) {
	t.Setenv("GOANIME_POMFY_URL", "https://mirror.example/")
	c := NewClient()
	assert.Equal(t, "https://mirror.example", c.BaseURL(), "env wins and the trailing slash drops")
	assert.Equal(t, apiCatalog+"/api/tmdb-season", c.catalogBase+"/api/tmdb-season", "the season host stays the partner one")

	t.Setenv("GOANIME_POMFY_URL", "   ")
	bare := NewClient()
	assert.Equal(t, DefaultBase, bare.BaseURL(), "a blank env falls to the default watch site")
}

func TestSharedReturnsTheSameClient(t *testing.T) {
	first, second := Shared(), Shared()
	require.NotNil(t, first, "the shared client bootstraps itself")
	assert.Same(t, first, second, "shared must not rebuild")
}

func TestSeriesPlayURLShape(t *testing.T) {
	got := SeriesPlayURL(1396, 2, 1)
	assert.Equal(t, apiCatalog+"/serie/1396/2/1", got)
}

func TestTMDBConfiguredSeesBothSources(t *testing.T) {
	// The public pool makes the search always configured, key or no key:
	assert.True(t, TMDBConfigured(), "the public pool keeps the search available")
}

func TestSearchSeamBothBranches(t *testing.T) {
	c := &Client{}
	assert.NotNil(t, c.searchSeam(), "a nil seam must fall to the shared one")

	fake := func(query string) ([]Media, error) { return []Media{{Title: query}}, nil }
	c.SetSearchSeam(fake)
	got, err := c.searchSeam()("clube")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "clube", got[0].Title)
}

// Search's guards: the empty query answers nothing; the catalog error and
// the search error ride through.
func TestSearchEdges(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)

	got, err := c.Search(t.Context(), "   ")
	require.NoError(t, err)
	assert.Nil(t, got, "an all-space query answers empty")

	// A dead catalog poisons the search before TMDB is asked.
	site.mu.Lock()
	site.breakCatalog = true
	site.mu.Unlock()
	if _, err := c.Search(t.Context(), "x"); err == nil {
		t.Fatal("a broken catalog must fail the search")
	}
}

func TestToAnimeModelTable(t *testing.T) {
	movie := Media{Title: "Clube da Luta", TMDBID: 550, URL: "https://pomfy.test/assistir/550",
		Poster: "https://image.tmdb.org/t/p/w300/a.jpg", Year: "1999", Rating: 8.4,
		Overview: "desc", Genres: []string{"Drama"}, Kind: KindMovie}.ToAnimeModel()
	movieAssertions(t, movie)

	series := Media{Title: "Breaking Bad", TMDBID: 1396, Kind: KindSeries}.ToAnimeModel()
	assert.Equal(t, models.MediaTypeTV, series.MediaType, "series sinks to TV")
	assert.Equal(t, SourceName, series.Source, "the canonical Source string rides")
	assert.Equal(t, 0.0, series.Rating, "no rating stays zero")

	blank := Media{}.ToAnimeModel()
	assert.Equal(t, models.MediaTypeTV, blank.MediaType, "an empty kind defaults to the TV shape")
	assert.Nil(t, nil, blank.Genres)
}

func movieAssertions(t *testing.T, anime *models.Anime) {
	t.Helper()
	require.NotNil(t, anime)
	assert.Equal(t, models.MediaTypeMovie, anime.MediaType, "movie kind")
	assert.Equal(t, 550, anime.TMDBID, "id rides")
	assert.Equal(t, "Clube da Luta", anime.Name, "title rides")
	assert.Equal(t, SourceName, anime.Source, "source rides")
	assert.Equal(t, "1999", anime.Year, "year rides")
	assert.Equal(t, 8.4, anime.Rating, "rating rides")
	assert.Equal(t, "desc", anime.Overview, "overview rides")
	require.Len(t, anime.Genres, 1, "genres ride")
}

func TestWatchSourcesErrorShapes(t *testing.T) {
	// the site sends an error object
	errSite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/watch-sources") {
			_, _ = w.Write([]byte(`{"error":"the site is not sure"}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(errSite.Close)
	c := NewClientForTest(errSite.Client(), errSite.URL)
	_, err := c.WatchSources(t.Context(), 550, KindMovie, 0, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSources), "a site error forces ErrNoSources, got %v", err)

	// unparseable body
	junkSite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/watch-sources") {
			_, _ = w.Write([]byte(`{{`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(junkSite.Close)
	cj := NewClientForTest(junkSite.Client(), junkSite.URL)
	_, err = cj.WatchSources(t.Context(), 550, KindMovie, 0, 0)
	require.Error(t, err, "unparseable must fail loudly")
}

func TestCatalogDataGarbageAndCtxErr(t *testing.T) {
	junk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	t.Cleanup(junk.Close)
	c := NewClientForTest(junk.Client(), junk.URL)
	_, err := c.catalogData(t.Context())
	require.Error(t, err, "unparseable catalog must fail")

	// expired context: the getter returns what ctx.Err says
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := c.catalogData(ctx); err == nil {
		t.Error("a dead context must fail the catalog")
	}
}

func TestHasTitleCatalogFail(t *testing.T) {
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"down"}`, http.StatusInternalServerError)
	}))
	t.Cleanup(broken.Close)
	c := NewClientForTest(broken.Client(), broken.URL)
	if _, err := c.HasTitle(t.Context(), 550, KindMovie); err == nil {
		t.Error("a 5xx catalog must fail the membership check")
	}
}

// ShowInfo: the show's unleaky paths. The site errors, the payload runs the
// fallback scan after a broken pointer shape, and no-show shapes fail
// loudly.
func TestShowInfoFailPaths(t *testing.T) {
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"gone"}`, http.StatusNotFound)
	}))
	t.Cleanup(broken.Close)
	c := NewClientForTest(broken.Client(), broken.URL)
	if _, err := c.ShowInfo(t.Context(), 1396); err == nil {
		t.Error("a missing show payload must error")
	}
}

// The recorded array decodes via the walking scan itself (the fallback's
// body), and a broken head pointer routes through it.
func TestDecodeShowFallbackScan(t *testing.T) {
	// recorded fixture, straight from the production shapes
	parsed := struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}{}
	require.NoError(t, json.Unmarshal([]byte(showPayloadJSON()), &parsed))
	var arr []json.RawMessage
	require.NoError(t, json.Unmarshal(parsed.Nodes[1].Data, &arr))

	show, err := decodeShowAt(arr)
	require.NoError(t, err)
	assert.Equal(t, "Breaking Bad", show.Name)

	// Now the head pointer goes out of bounds: the node walk must fall back
	// to the scan above and still read the show.
	mangled := make([]json.RawMessage, len(arr))
	copy(mangled, arr)
	mangled[0] = json.RawMessage(`{"tvShow":99999}`)
	show, err = decodeShowNode(mangled)
	require.NoError(t, err)
	assert.Equal(t, "Breaking Bad", show.Name)
}

func TestSeasonListingFailures(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	c.maxRetries = 0

	// missing id/s → the 400 answer
	chaos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"missing"}`, http.StatusBadRequest)
	}))
	t.Cleanup(chaos.Close)
	bad := NewClientForTest(chaos.Client(), chaos.URL)
	if _, err := bad.Season(t.Context(), 1396, 2); err == nil {
		t.Error("a 400 season listing must error")
	}

	if _, err := bad.Season(t.Context(), 404, 2); err == nil {
		t.Error("an unroutable server must error")
	}

	// a season reply with no episodes must be ErrNoEpisodeSource
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/tmdb-season") {
			_, _ = w.Write([]byte(`{"epNums":[],"sbEps":[],"tmdbEps":{}}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(empty.Close)
	ll := NewClientForTest(empty.Client(), empty.URL)
	_, err := ll.Season(t.Context(), 1396, 9)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoEpisodeSource), "an empty season is the product's emptiness, got %v", err)
}

// TitleByIDs: the broken payload shapes fail; a nil catalog makes the
// function read its own.
func TestTitleByIDFailureBranches(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	site.mu.Lock()
	site.breakTitles = true
	site.mu.Unlock()

	ctx := t.Context()
	if _, ok := c.TitleByIDs(ctx, 550, nil); ok {
		t.Error("a broken movie payload must not return a title")
	}
	if _, ok := c.TitleByIDs(ctx, 1396, nil); ok {
		t.Error("a broken series payload must not return a title")
	}
}

func TestTMDBSearchUserKeyPaths(t *testing.T) {
	fake := newFakeTMDB(t, "clube da luta", fakeMedia(3))
	fake.userKey = "a-user-key0000"

	setTMDBBase(t, fake.srv.URL+"/3")
	resetSearchCache(t)

	// 200 through the fake user client
	withUserClient(t, fake, "a-user-key0000")
	mov, err := tmdbSearch("clube da luta")
	require.NoError(t, err)
	require.Len(t, mov, 3)

	// 429 falls to the pool, which the fake also carries
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "clube da luta", fakeMedia(2))
	tmdb.userKey = "quota"
	tmdb.quota["quota"] = true
	setTMDBBase(t, tmdb.srv.URL+"/3")
	withUserClient(t, tmdb, "quota")
	mov, err = tmdbSearch("clube da luta")
	require.NoError(t, err)
	require.Len(t, mov, 2)

	// a 400 must NOT go to the pool
	resetSearchCache(t)
	off := newFakeTMDB(t, "clube da luta", fakeMedia(1))
	off.userKey = "bad"
	off.forceStatus = http.StatusBadRequest
	setTMDBBase(t, off.srv.URL+"/3")
	withUserClient(t, off, "bad")
	_, err = tmdbSearch("clube da luta")
	require.Error(t, err)
}

// The pool itself: exhaustion and emptiness.
func TestTMDBSearchPoolEdges(t *testing.T) {
	// every pool key over quota → the last answer travels
	resetSearchCache(t)
	tmdb := newFakeTMDB(t, "clube da luta", fakeMedia(1))
	setTMDBBase(t, tmdb.srv.URL+"/3")
	restorePool := withFakePool(t, []string{PublicKeyForTesting(0)})
	tmdb.quota[PublicKeyForTesting(0)] = true
	_, err := tmdbSearch("clube da luta")
	require.Error(t, err)
	restorePool()

	// empty pool → ErrNoTMDBAPIKey
	resetSearchCache(t)
	restorePool = withFakePool(t, nil)
	_, err = tmdbSearch("clube da luta")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoTMDBAPIKey), "an empty pool is exactly the no-key state, got %v", err)
	restorePool()
}

// the for-test helpers round trip in-package.
func TestForTestHelpersRoundTrip(t *testing.T) {
	sel := SealForTest([]byte(`{"sources":[{"url":"https://edge.test/hx/master.m3u8","label":"1080p","mime_type":"application/vnd.apple.mpegurl","height":1080}],"tracks":[]}`))

	var wrapped struct {
		Playback playbackSeal `json:"playback"`
	}
	require.NoError(t, json.Unmarshal(sel, &wrapped))
	opened, err := decryptBysePlayback(&wrapped.Playback)
	require.NoError(t, err)
	require.NotEmpty(t, opened.Sources, "the sealed playback must open back up")

	sol, err := SolveForTest("feprefix", 5)
	require.NoError(t, err)
	assert.True(t, CheckForTest("feprefix:", sol, 5), "the solver's own answer must pass its judge")
	assert.False(t, CheckForTest("feprefix:", "0", 5))
}
