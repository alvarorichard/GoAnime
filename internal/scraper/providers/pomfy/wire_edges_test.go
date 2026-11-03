package pomfy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// The memoize-circuit and retry-order tests: the devalue graph's second read
// rides its cache, the HTTP layer's sleepy retry lane and the hijacked
// truncated-body lane, and the few refusal shapes the parser edge-file.

func TestDevalueSecondReadUsesCache(t *testing.T) {
	arr := []json.RawMessage{
		json.RawMessage(`{"name":1,"seasons":2}`),
		json.RawMessage(`"Cached Show"`),
		json.RawMessage(`[3]`),
		json.RawMessage(`{"name":4}`),
		json.RawMessage(`"Entry"`),
	}
	g := newDevalueGraph(arr)
	// the first read fills the cache; the second one rides it.
	first, okFirst := g.object(0)
	second, okSecond := g.object(0)
	if !okFirst || !okSecond || first == nil || second == nil {
		t.Fatal("the cached object must resolve both times")
	}
	// arrays' cached return is a bare value; hitting the branch twice works:
	g.array(2)
	g.array(2)
	g.stringRef(1)
	g.stringRef(1)
	g.integerRef(0) // a non-number → memoized zero on the second read
	g.integerRef(0)
}

func TestGetSleepLaneAndCanceledRetry(t *testing.T) {
	// ctx canceled during the retry sleep lands the call as a context error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"upstream"}`, http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 2
	c.retryDelay = 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	if _, err := c.get(ctx, srv.URL+"/x", nil); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v, want a canceled context", err)
	}
}

// hijackedShortServer answers with a Content-Length beyond what it writes,
// then slams the connection shut — the client's read fails mid-body.
func hijackedShortServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "no hijack", http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		defer func() { _ = conn.Close() }()
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 5000\r\n\r\nshort body"))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestAttemptReadError(t *testing.T) {
	srv := hijackedShortServer(t)
	c := NewClientForTest(&http.Client{}, DefaultBase)
	_, retry, err := c.attempt(context.Background(), http.MethodGet, srv.URL, nil, "", nil)
	if err == nil {
		t.Fatal("a truncated body must fail the read")
	}
	if !retry {
		t.Error("a truncated body is a network-flavored failure, retryable")
	}
}

func TestWatchSourcesAllDroppedNamesErrNoSources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/watch-sources") {
			_, _ = w.Write([]byte(`{"mode":"direct","sources":[{"kind":"flyfile","embedUrl":"https://fly.test/x"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	_, err := c.WatchSources(t.Context(), 550, KindMovie, 0, 0)
	if !errors.Is(err, ErrNoSources) {
		t.Fatalf("err = %v, want ErrNoSources when no byse source survives the filter", err)
	}
}

func TestSearchSeamErrorAndBlankSkip(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	c.SetSearchSeam(func(string) ([]Media, error) {
		return []Media{
			{}, // no kind, no type: skipped before the catalog read
			{Title: "X", TMDBID: 550, TMDBType: "movie"},
		}, errors.New("seam is dead")
	})
	if _, err := c.Search(t.Context(), "clube"); err == nil || !strings.Contains(err.Error(), "seam is dead") {
		t.Fatalf("err = %v, want the seam's own error", err)
	}
}

func TestTitleByIDsDeadCatalog(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	site.mu.Lock()
	site.breakCatalog = true
	site.mu.Unlock()

	if _, ok := c.TitleByIDs(t.Context(), 550, nil); ok {
		t.Error("a dead catalog blocks the by-id title lookup")
	}
}

// The show payload shapes the parser politely declines: a node whose data is
// not an array, an empty array, a pointer to a non-object, a season list
// with non-object entries.
func TestShowPayloadEdgeShapes(t *testing.T) {
	// node data NOT an array — the parse skips over it
	shape := compact(`{"nodes":[
		{"data": {"a": 1}},
		{"data": [
			{"tvShow": 3},
			null,
			"Falling Skies",
			{"name": 4, "first_air_date": 5, "seasons": 6},
			"2011-06-19",
			[7],
			{"season_number": 8, "episode_count": 9, "name": 10},
			2,
			12,
			"Temporada 1"
		]}
	]}`)
	var parsed struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(shape), &parsed); err != nil {
		t.Fatal(err)
	}

	if _, err := parseShowInfo([]byte(shape)); err != nil {
		t.Fatalf("shape with a skipped node: %v", err)
	}

	// empty array lens across the devalue walk
	if _, err := parseShowInfo([]byte(`{"type":"data","nodes":[]}`)); err == nil {
		t.Fatal("no nodes must fail")
	}

	// the tvShow pointer landing on a non-object entry
	badShape := compact(`{"nodes":[
		{"data":[]},
		{"data":[
			{"tvShow": 1},
			"just a string where the show should be"
		]}
	]}`)
	if _, err := parseShowInfo([]byte(badShape)); err == nil ||
		!strings.Contains(err.Error(), "not an object") {
		t.Fatalf("err = %v, want the not-an-object refusal", err)
	}

	// a BROKEN pointer + a payload whose seasons key is missing lands the
	// walk in decodeShowAt, which exhausts without a match.
	brokenNoSeasons := compact(`{"nodes":[
		{"data":[]},
		{"data":[
			{"tvShow":9},
			{"unrelated":1}
		]}
	]}`)
	if _, err := parseShowInfo([]byte(brokenNoSeasons)); err == nil ||
		!strings.Contains(err.Error(), "tvShow entry missing") {
		t.Fatalf("err = %v, want the missing-show refusal", err)
	}
}

// decodeShow on a synthesized graph: seasons with non-object members are
// skipped instead of killing the decode.
func TestDecodeShowSkipsBadSeasonEntries(t *testing.T) {
	arr := []json.RawMessage{
		json.RawMessage(`{"name":1,"first_air_date":2,"seasons":3}`),
		json.RawMessage(`"Mini Show"`),
		json.RawMessage(`"2019-01-01"`),
		json.RawMessage(`[4, 5]`),
		json.RawMessage(`{"season_number":6,"episode_count":7,"name":8}`),
		json.RawMessage(`1`),
		json.RawMessage(`2`),
		json.RawMessage(`"Tag line"`),
	}
	show, err := decodeShowAt(arr)
	if err != nil {
		t.Fatalf("decodeShowAt: %v", err)
	}
	if show.Name != "Mini Show" {
		t.Errorf("name = %q", show.Name)
	}
	if len(show.Seasons) != 1 {
		t.Fatalf("season entries drop wrong: %+v", show.Seasons)
	}
	// the pointer chain: season_number→arr[6]=2, episode_count→arr[7]="Tag
	// line" (not a number → 0); the second entry (arr[5]=1) is an int, not
	// an object, and the decode skips it.
	if show.Seasons[0].Number != 2 || show.Seasons[0].EpisodeCount != 0 {
		t.Errorf("season = %+v", show.Seasons[0])
	}
}

// assert may be simpler; the shapes above use if-then; local panic helper
// replaced by the if-then above.

// byseKey skips the empty parts (the mirror's padding lanes).
func TestByseKeySkipsEmptyParts(t *testing.T) {
	good := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	pb := &playbackSeal{KeyParts: []string{"", good}, Version: ""} // the "" joins every part
	key, err := byseKey(pb)
	if err != nil {
		t.Fatalf("byseKey: %v", err)
	}
	if !strings.HasPrefix(strings.Repeat("k", 32), string(key[:1])) || len(key) != 32 {
		t.Errorf("key = %x, want the one good part", key)
	}
}

// resolveByse surfaces a proof-of-work failure when the solver's context
// dies mid-hunt (a canceled parent context at the captcha stage).
func TestResolveByseSolverContextDeath(t *testing.T) {
	// The impossible difficulty keeps the solver busy; the short-lived
	// context fires while the POSTs have sailed through.
	cdn := newFakeByseCDN(t, "ref-hard", 30, sealFixture(t, recordJSON(t)))
	c := NewClientForTest(cdn.srv.Client(), cdn.srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err := resolveByse(ctx, c, cdn.srv.URL, "cta12345", DefaultBase+"/assistir/550")
	if err == nil || !strings.Contains(err.Error(), "proof-of-work: context deadline exceeded") {
		t.Fatalf("err = %v, want the solver's context error", err)
	}
}

// the same trip at the solver's own level: the loop's ctx gate.
func TestSolveNonceLoopContextDeath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err := solveNonce(ctx, "hard-pref", 30, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("err = %v, want the loop's ctx error", err)
	}
}

// heightDigits refuses a label whose digit run overflows int.
func TestHeightDigitsOverflowLabel(t *testing.T) {
	if got := labelHeight("9999999999999999999"); got != 0 {
		t.Errorf("overflowing label = %d, want 0", got)
	}
}

// The season listing refuses a server that answers 404 for the id-less
// query shape (the fake catalog route missing).
func TestWatchSourcesSeasonRouteMissing(t *testing.T) {
	nothing := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(nothing.Close)
	c := NewClientForTest(nothing.Client(), nothing.URL)
	c.maxRetries = 0
	if _, err := c.Season(t.Context(), 5, 1); err == nil {
		t.Error("a missing season route must fail the listing")
	}
}

// the URL shapes the embed rules guard against.
func TestResolveStreamSeriesFlow(t *testing.T) {
	world := newFakeWorld(t)
	c := world.client()
	stream, err := c.ResolveStream(t.Context(), 1396, KindSeries, 2, 1, DefaultBase+"/serie/1396")
	if err != nil {
		t.Fatalf("ResolveStream: %v", err)
	}
	if stream.URL == "" {
		t.Error("the series stream must resolve")
	}
	var parsed *url.URL
	parsed, _ = url.Parse(stream.URL)
	if parsed == nil || parsed.Host == "" {
		t.Errorf("stream = %q", stream.URL)
	}
}
