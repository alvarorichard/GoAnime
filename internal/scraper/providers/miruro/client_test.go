package miruro

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func diagKind(t *testing.T, err error) netx.DiagnosticKind {
	t.Helper()
	var d *netx.SourceDiagnostic
	require.True(t, errors.As(err, &d), "want a SourceDiagnostic, got %T: %v", err, err)
	return d.Kind
}

// ---------------------------------------------------------------------------
// devalue
// ---------------------------------------------------------------------------

func TestHostIsPinned(t *testing.T) {
	t.Parallel()
	// A rotation must be a deliberate edit, checked against the live site
	// (2026-11-26: miruro.cx, the host ani-cli 5.2.0 reads).
	assert.Equal(t, "https://miruro.cx", miruroBase)
	assert.Equal(t, miruroBase, NewMiruroClient().baseURL)
}

func TestHydrateRoundTripsTheTestEncoder(t *testing.T) {
	t.Parallel()
	// The fixtures are only as good as encodeDevalue; this pins it to the
	// decoder on every value kind the fixtures use.
	values := []any{
		"plain",
		float64(42),
		true,
		nil,
		[]any{},
		map[string]any{},
		[]any{"a", float64(1), nil, map[string]any{"k": []any{false}}},
		map[string]any{"title": map[string]any{"english": nil, "romaji": "R"}, "n": float64(1.5)},
	}
	for _, v := range values {
		got, err := hydrateFlat(encodeDevalue(v), nil)
		require.NoError(t, err, "%v", v)
		assert.Equal(t, v, got)
	}
}

func TestHydrateConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		idx  int
		want any
	}{
		{devalueUndefined, nil},
		{devalueHole, nil},
		{devalueNaN, nil},
		{devaluePosInfinity, nil},
		{devalueNegInfinity, nil},
		{devalueNegZero, float64(0)},
	}
	for _, tt := range tests {
		got, err := hydrateFlat([]any{map[string]any{"v": float64(tt.idx)}}, nil)
		require.NoError(t, err, "constant %d", tt.idx)
		assert.Equal(t, map[string]any{"v": tt.want}, got, "constant %d", tt.idx)
	}

	// A bare constant in place of the array is how devalue writes a whole
	// value that is one.
	got, err := hydrateFlat(float64(devalueUndefined), nil)
	require.NoError(t, err)
	assert.Nil(t, got)

	_, err = hydrateFlat([]any{map[string]any{"v": float64(-9)}}, nil)
	assert.ErrorContains(t, err, "unknown constant -9")
}

func TestHydrateTypedValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		flat []any
		want any
	}{
		{"date", []any{[]any{"Date", "2026-11-26T00:00:00.000Z"}}, "2026-11-26T00:00:00.000Z"},
		{"bigint", []any{[]any{"BigInt", "12345678901234567890"}}, "12345678901234567890"},
		{"url", []any{[]any{"URL", "https://x.example/"}}, "https://x.example/"},
		{"regexp", []any{[]any{"RegExp", "^a+$", "gi"}}, "^a+$"},
		{"regexp without flags", []any{[]any{"RegExp", "x"}}, "x"},
		{"map with a dangling key", []any{[]any{"Map", float64(1), float64(2), float64(1)}, "k", "v"}, map[string]any{"k": "v"}},
		{"null prototype with a dangling key", []any{[]any{"null", "k", float64(1), "dangling"}, "v"}, map[string]any{"k": "v"}},
		{"boxed", []any{[]any{"Object", float64(1)}, "boxed"}, "boxed"},
		{"set", []any{[]any{"Set", float64(1), float64(2)}, "a", "b"}, []any{"a", "b"}},
		{"map", []any{[]any{"Map", float64(1), float64(2)}, float64(7), "seven"}, map[string]any{"7": "seven"}},
		{"null prototype", []any{[]any{"null", "k", float64(1)}, "v"}, map[string]any{"k": "v"}},
		{"unknown tag", []any{[]any{"Uint8Array", "AAEC"}}, nil},
		{"date of the wrong shape", []any{[]any{"Date", float64(3)}}, nil},
		{"boxed of the wrong shape", []any{[]any{"Object"}}, nil},
		{"promise without a resolver", []any{[]any{"Promise", float64(1)}, float64(4)}, nil},
	}
	for _, tt := range tests {
		got, err := hydrateFlat(tt.flat, nil)
		require.NoError(t, err, tt.name)
		assert.Equal(t, tt.want, got, tt.name)
	}
}

func TestHydrateRejectsMalformedData(t *testing.T) {
	t.Parallel()
	deep := []any{}
	for i := 0; i <= maxHydrateDepth+1; i++ {
		deep = append(deep, []any{float64(i + 1)})
	}
	deep = append(deep, "bottom")

	tests := []struct {
		name string
		data any
		want string
	}{
		{"not an array", "x", "not an array"},
		{"empty", []any{}, "empty data array"},
		{"index out of range", []any{map[string]any{"a": float64(5)}}, "out of range"},
		{"index equal to the length", []any{map[string]any{"a": float64(1)}}, "out of range"},
		{"fractional index", []any{map[string]any{"a": 0.5}}, "not an index"},
		{"string index", []any{map[string]any{"a": "1"}}, "not an index"},
		{"self reference", []any{[]any{float64(0)}}, "refers to itself"},
		{"cycle", []any{map[string]any{"a": float64(1)}, map[string]any{"b": float64(0)}}, "refers to itself"},
		{"too deep", deep, "nested too deep"},
		{"null-prototype key", []any{[]any{"null", float64(3), float64(1)}, "v"}, "not a string"},
		{"promise id", []any{[]any{"Promise", float64(1)}, "x"}, "not a number"},
		{"bad child in object", []any{map[string]any{"a": float64(1)}, map[string]any{"b": float64(9)}}, "out of range"},
		{"bad child in set", []any{[]any{"Set", float64(9)}}, "out of range"},
		{"bad map key", []any{[]any{"Map", float64(9), float64(0)}}, "out of range"},
		{"bad map value", []any{[]any{"Map", float64(1), float64(9)}, "k"}, "out of range"},
		{"bad boxed", []any{[]any{"Object", float64(9)}}, "out of range"},
		{"bad null value", []any{[]any{"null", "k", float64(9)}}, "out of range"},
		{"bad promise ref", []any{[]any{"Promise", float64(9)}}, "out of range"},
	}
	for _, tt := range tests {
		resolve := func(int) (any, error) { return nil, nil }
		_, err := hydrateFlat(tt.data, resolve)
		assert.ErrorContains(t, err, tt.want, tt.name)
	}
}

func TestHydrateAcceptsTheDeepestAllowedNesting(t *testing.T) {
	t.Parallel()
	var flat []any
	for i := 0; i < maxHydrateDepth; i++ {
		flat = append(flat, []any{float64(i + 1)})
	}
	flat = append(flat, "bottom")
	got, err := hydrateFlat(flat, nil)
	require.NoError(t, err, "nesting exactly at the limit is allowed")
	for i := 0; i < maxHydrateDepth; i++ {
		got = got.([]any)[0]
	}
	assert.Equal(t, "bottom", got)
}

func TestHydrateSharesARepeatedReference(t *testing.T) {
	t.Parallel()
	// One object referred to twice is built once; both places see it.
	got, err := hydrateFlat([]any{map[string]any{"a": float64(1), "b": float64(1)}, map[string]any{"k": float64(2)}, "v"}, nil)
	require.NoError(t, err)
	m := got.(map[string]any)
	assert.Equal(t, map[string]any{"k": "v"}, m["a"])
	assert.Equal(t, m["a"], m["b"])
}

func TestParseDataResponseReadsTheRealShape(t *testing.T) {
	t.Parallel()
	resp, err := parseDataResponse([]byte(rawWatchBody))
	require.NoError(t, err)
	require.Len(t, resp.Nodes, 2)
	assert.Equal(t, "data", resp.Type)

	page, ok := resp.nodeWith("episodesData")
	require.True(t, ok)
	assert.Equal(t, float64(1), page["episodeNumber"])
	assert.Nil(t, page["backdrop"], "-1 is undefined")
	assert.Equal(t, "2026-11-26T00:00:00.000Z", page["updated"])

	var wp watchPage
	require.NoError(t, decodeInto(page, &wp))
	require.NotNil(t, wp.EpisodesData)
	require.Len(t, wp.EpisodesData.Episodes, 1)
	assert.Equal(t, 1.0, *wp.EpisodesData.Episodes[0].Number)
	assert.Equal(t, "Asteroid Blues", wp.EpisodesData.Episodes[0].Title)

	require.NotNil(t, wp.SourcesData)
	assert.Equal(t, []string{"sun"}, wp.SourcesData.ProviderOrder)
	require.Len(t, wp.SourcesData.Tracks, 1)
	srv := wp.SourcesData.Tracks[0].Providers[0].Servers[0]
	assert.Equal(t, "icarus-2-1", srv.Server)
	assert.Equal(t, "https://ultracloud.example", srv.Headers["Origin"])
	assert.Equal(t, sourceStream{URL: "https://cdn.example/m/abc.m3u8", Format: "hls"}, srv.Streams[0])
}

func TestParseDataResponse(t *testing.T) {
	t.Parallel()
	chunkLine := func(id int, v any) string {
		return mustJSON(map[string]any{"type": "chunk", "id": id, "data": encodeDevalue(v)})
	}
	head := func(root any) string { return strings.TrimSpace(dataBody(root)) }

	t.Run("redirect", func(t *testing.T) {
		resp, err := parseDataResponse([]byte(`{"type":"redirect","location":"/watch/a/b?ep=1"}`))
		require.NoError(t, err)
		assert.Equal(t, dataResponse{Type: "redirect", Location: "/watch/a/b?ep=1"}, resp)
	})
	t.Run("errors", func(t *testing.T) {
		for body, want := range map[string]string{
			"":                                "empty response",
			"   \n  ":                         "empty response",
			"{":                               "malformed data response",
			`{"type":"redirect"}`:             "redirect without a location",
			`{"type":"error","error":"boom"}`: "answered with an error",
			`{"type":"mystery"}`:              "unknown data response type",
			head(map[string]any{}) + "\n{":    "malformed chunk",
			`{"type":"data","nodes":[{"type":"data","data":"x"}]}`: "node 0",
		} {
			_, err := parseDataResponse([]byte(body))
			assert.ErrorContains(t, err, want, "body %q", body)
		}
	})
	t.Run("a skipped node is nil and the next one is still read", func(t *testing.T) {
		resp, err := parseDataResponse([]byte(`{"type":"data","nodes":[{"type":"skip"},null,{"type":"data","data":["page"]}]}`))
		require.NoError(t, err)
		assert.Equal(t, []any{nil, nil, "page"}, resp.Nodes)
	})
	t.Run("promises settle from their chunks, in any order, nested", func(t *testing.T) {
		body := head(map[string]any{"a": promise(1), "b": promise(2), "c": promise(1)}) + "\n" +
			chunkLine(2, map[string]any{"inner": promise(3)}) + "\n" +
			chunkLine(1, "one") + "\n \t\r\n" +
			chunkLine(3, []any{"three"}) + "\n" +
			`{"type":"other","id":9}`
		resp, err := parseDataResponse([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			"a": "one", "c": "one",
			"b": map[string]any{"inner": []any{"three"}},
		}, resp.Nodes[0])
	})
	t.Run("an unsettled or rejected promise is absent", func(t *testing.T) {
		body := head(map[string]any{"a": promise(1), "b": promise(2)}) + "\n" +
			mustJSON(map[string]any{"type": "chunk", "id": 2, "error": "rejected"})
		resp, err := parseDataResponse([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"a": nil, "b": nil}, resp.Nodes[0])
	})
	t.Run("a promise that settles to itself is refused", func(t *testing.T) {
		body := head(map[string]any{"a": promise(1)}) + "\n" + chunkLine(1, map[string]any{"again": promise(1)})
		_, err := parseDataResponse([]byte(body))
		assert.ErrorContains(t, err, "refers to itself")
	})
	t.Run("a broken chunk fails its node", func(t *testing.T) {
		body := head(map[string]any{"a": promise(1)}) + "\n" + `{"type":"chunk","id":1,"data":[{"x":7}]}`
		_, err := parseDataResponse([]byte(body))
		assert.ErrorContains(t, err, "chunk 1")
	})
}

func TestNodeWith(t *testing.T) {
	t.Parallel()
	r := dataResponse{Nodes: []any{nil, "x", map[string]any{"a": 1}, map[string]any{"items": nil}}}
	n, ok := r.nodeWith("items")
	require.True(t, ok, "a key holding null still names the node")
	assert.Equal(t, map[string]any{"items": nil}, n)
	_, ok = r.nodeWith("missing")
	assert.False(t, ok)
}

func TestDecodeInto(t *testing.T) {
	t.Parallel()
	var out struct {
		N float64 `json:"n"`
	}
	require.NoError(t, decodeInto(map[string]any{"n": 2.5}, &out))
	assert.Equal(t, 2.5, out.N)
	assert.Error(t, decodeInto(math.Inf(1), &out), "a value JSON cannot carry is an error")
	assert.Error(t, decodeInto("text", &out), "a value of the wrong shape is an error")
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func TestNewClientForTest(t *testing.T) {
	t.Parallel()
	c := NewClientForTest("http://127.0.0.1:1/")
	assert.Equal(t, "http://127.0.0.1:1", c.baseURL)
	assert.Equal(t, 0, c.maxRetries, "tests fail fast")
	assert.Equal(t, time.Duration(0), c.retryDelay)
	live := NewMiruroClient()
	assert.Equal(t, 2, live.maxRetries)
	assert.Equal(t, 300*time.Millisecond, live.retryDelay)
	assert.Equal(t, 6*time.Second, live.probeTimeout)
}

func TestGetDataSendsNoQueryWhenThereIsNone(t *testing.T) {
	t.Parallel()
	var uri string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri = r.RequestURI
		_, _ = w.Write([]byte(dataBody("x")))
	}))
	defer srv.Close()
	_, err := NewClientForTest(srv.URL).getData(context.Background(), "/search/", nil, "search")
	require.NoError(t, err)
	assert.Equal(t, "/search/__data.json", uri, "no trailing ? and no doubled slash")
}

func TestGetDataFollowsTheSiteRedirect(t *testing.T) {
	t.Parallel()
	s := newStub(t, nil)
	c := NewClientForTest(s.srv.URL)
	resp, err := c.getData(context.Background(), "/watch/"+bebopID+"/x", map[string][]string{"ep": {"1"}}, "episodes")
	require.NoError(t, err)
	assert.Equal(t, "/watch/"+bebopID+"/"+bebopSlug, resp.Path)
	assert.Equal(t, []string{
		"/watch/" + bebopID + "/x/__data.json",
		"/watch/" + bebopID + "/" + bebopSlug + "/__data.json",
	}, s.paths())
}

func TestGetDataRefusesRedirectsItShouldNotFollow(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"absolute":    `{"type":"redirect","location":"https://evil.example/watch/a/b"}`,
		"scheme-less": `{"type":"redirect","location":"//evil.example/watch/a/b"}`,
		"relative":    `{"type":"redirect","location":"watch/a/b"}`,
		"unparsable":  `{"type":"redirect","location":"/watch/%zz"}`,
	}
	for name, body := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }))
		_, err := NewClientForTest(srv.URL).getData(context.Background(), "/x", nil, "episodes")
		srv.Close()
		require.Error(t, err, name)
		assert.Equal(t, netx.DiagnosticParserBroken, diagKind(t, err), name)
	}

	var hits atomic.Int32
	loop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"type":"redirect","location":"/again"}`))
	}))
	defer loop.Close()
	_, err := NewClientForTest(loop.URL).getData(context.Background(), "/x", nil, "episodes")
	assert.ErrorContains(t, err, "too many redirects")
	assert.Equal(t, int32(maxRedirects+1), hits.Load(), "the limit is exact")
}

func TestGetDataReportsUnreadableData(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"type":"mystery"}`)) }))
	defer srv.Close()
	_, err := NewClientForTest(srv.URL).getData(context.Background(), "/x", nil, "search")
	assert.Equal(t, netx.DiagnosticParserBroken, diagKind(t, err))
}

func TestSitePath(t *testing.T) {
	t.Parallel()
	u, err := sitePath("/watch/a/b?ep=2")
	require.NoError(t, err)
	assert.Equal(t, "/watch/a/b", u.Path)
	assert.Equal(t, "2", u.Query().Get("ep"))
	for _, bad := range []string{"https://x.example/a", "//x.example/a", "a/b", "", "http:/a"} {
		_, err := sitePath(bad)
		assert.ErrorIs(t, err, errOffSite, bad)
	}
	_, err = sitePath("%zz")
	assert.Error(t, err)
}

func TestGetBodyClassifiesFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		ctype  string
		body   string
		want   netx.DiagnosticKind
	}{
		{"cloudflare interstitial", 403, "text/html; charset=UTF-8", "<html><title>Just a moment...</title></html>", netx.DiagnosticBlockedChallenge},
		{"cloudflare challenge script", 503, "text/html", `<script src="/cdn-cgi/challenge-platform/h/b/cf-chl-x"></script>`, netx.DiagnosticBlockedChallenge},
		{"cloudflare block page", 403, "text/html", "<title>Attention Required! | Cloudflare</title>", netx.DiagnosticBlockedChallenge},
		{"plain 404", 404, "application/json", `{"message":"Not found"}`, netx.DiagnosticParserBroken},
		{"origin down", 502, "text/plain", "bad gateway", netx.DiagnosticSourceUnavailable},
		{"html where data was expected", 200, "text/html", "<html>home</html>", netx.DiagnosticBlockedChallenge},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", tt.ctype)
			w.WriteHeader(tt.status)
			_, _ = w.Write([]byte(tt.body))
		}))
		_, err := NewClientForTest(srv.URL).getBody(context.Background(), srv.URL+"/x", "search")
		srv.Close()
		require.Error(t, err, tt.name)
		assert.Equal(t, tt.want, diagKind(t, err), tt.name)
	}
}

func TestIsChallenge(t *testing.T) {
	t.Parallel()
	assert.True(t, isChallenge("text/html", []byte("<TITLE>Just a moment...</TITLE>")), "case-insensitive")
	assert.False(t, isChallenge("application/json", []byte("<title>Just a moment</title>")), "only HTML is a challenge page")
	assert.False(t, isChallenge("text/html", []byte("<title>Miruro</title>")))
}

func TestGetBodyRetriesServerErrorsAndStops(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError) // 500 itself is retried
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c := NewClientForTest(srv.URL)
	c.maxRetries = 2
	c.retryDelay = 20 * time.Millisecond
	start := time.Now()
	body, err := c.getBody(context.Background(), srv.URL+"/x", "search")
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))
	assert.Equal(t, int32(3), hits.Load())
	assert.GreaterOrEqual(t, time.Since(start), 40*time.Millisecond, "each retry waits the delay")
	c.retryDelay = 0

	hits.Store(0)
	c.maxRetries = 1
	_, err = c.getBody(context.Background(), srv.URL+"/x", "search")
	require.Error(t, err)
	assert.Equal(t, int32(2), hits.Load(), "one retry means two attempts")

	// A 4xx is the answer, not a hiccup.
	var four atomic.Int32
	srv4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		four.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv4.Close()
	c4 := NewClientForTest(srv4.URL)
	c4.maxRetries = 3
	_, err = c4.getBody(context.Background(), srv4.URL+"/x", "search")
	require.Error(t, err)
	assert.Equal(t, int32(1), four.Load())
}

func TestGetBodyRetriesTransportErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	dead := srv.URL
	srv.Close()
	c := NewClientForTest(dead)
	c.maxRetries = 1
	c.retryDelay = 30 * time.Millisecond
	start := time.Now()
	_, err := c.getBody(context.Background(), dead+"/x", "search")
	require.Error(t, err)
	assert.Equal(t, netx.DiagnosticParserBroken, diagKind(t, err))
	assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond, "a failed connection is retried after the delay")
	assert.Less(t, time.Since(start), 5*time.Second)

	_, err = c.getBody(context.Background(), "://bad", "search")
	assert.ErrorContains(t, err, "bad request URL")
}

func TestGetBodyRetriesATruncatedBody(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte(`{"type":`))
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = conn.Close()
		}
	}))
	defer srv.Close()
	c := NewClientForTest(srv.URL)
	c.maxRetries = 1
	c.retryDelay = 30 * time.Millisecond
	start := time.Now()
	_, err := c.getBody(context.Background(), srv.URL+"/x", "search")
	require.Error(t, err)
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF, "the cause of a cut-off read is kept")
	assert.ErrorContains(t, err, "failed to read response")
	assert.Equal(t, int32(2), hits.Load())
	assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond)
}

func TestDecorateRequest(t *testing.T) {
	t.Parallel()
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()
	_, err := NewClientForTest(srv.URL).getBody(context.Background(), srv.URL+"/x", "search")
	require.NoError(t, err)
	assert.Equal(t, netx.UserAgent, got.Get("User-Agent"), "the site answers 403 to a client without a browser User-Agent")
	assert.Equal(t, netx.EnglishAcceptLanguage, got.Get("Accept-Language"))
	assert.Equal(t, srv.URL+"/", got.Get("Referer"))
	assert.Contains(t, got.Get("Accept"), "application/json")
}

func TestSleepHonoursCancellation(t *testing.T) {
	t.Parallel()
	c := NewMiruroClient()
	c.retryDelay = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { c.sleep(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sleep ignored a cancelled context")
	}

	c.retryDelay = 10 * time.Millisecond
	start := time.Now()
	c.sleep(context.Background())
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond, "an uncancelled sleep waits the delay out")
	c.retryDelay = 0
	c.sleep(context.Background())
}

// ---------------------------------------------------------------------------
// identity
// ---------------------------------------------------------------------------

func TestAnimeID(t *testing.T) {
	t.Parallel()
	good := map[string]string{
		bebopID:                             bebopID,
		"https://miruro.cx/info/" + bebopID: bebopID,
		"https://miruro.cx/info/" + bebopID + "/bebop":     bebopID,
		"https://miruro.cx/watch/" + bebopID + "/x?ep=3":   bebopID,
		"  https://mirror.example/watch/" + bebopID + "/ ": bebopID,
	}
	for in, want := range good {
		got, err := AnimeID(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, bad := range []string{"", "   ", "https://miruro.cx/search?q=x", "short", "https://miruro.cx/info/has space here"} {
		_, err := AnimeID(bad)
		assert.Error(t, err, bad)
	}
}

func TestParseEpisodeURL(t *testing.T) {
	t.Parallel()
	got, err := parseEpisodeURL("https://miruro.cx/watch/" + bebopID + "/cowboy-bebop?ep=12.5")
	require.NoError(t, err)
	assert.Equal(t, episodeRef{ID: bebopID, Slug: "cowboy-bebop", Number: "12.5"}, got)

	got, err = parseEpisodeURL("https://miruro.cx/watch/" + bebopID + "?foo=1&ep=3")
	require.NoError(t, err)
	assert.Equal(t, episodeRef{ID: bebopID, Number: "3"}, got)

	for _, bad := range []string{
		"",
		"https://miruro.cx/watch/" + bebopID + "/cowboy-bebop",
		"https://miruro.cx/watch/" + bebopID + "/cowboy-bebop?ep=",
		"https://miruro.cx/watch/" + bebopID + "/cowboy-bebop?ep=1x",
		"https://miruro.cx/search?ep=1",
	} {
		_, err := parseEpisodeURL(bad)
		assert.Error(t, err, bad)
	}
}

func TestURLBuilders(t *testing.T) {
	t.Parallel()
	c := NewClientForTest("https://m.example")
	assert.Equal(t, "https://m.example/info/"+bebopID, c.animeURL(bebopID))
	assert.Equal(t, "https://m.example/watch/"+bebopID+"/cowboy-bebop?ep=2", c.episodeURL(bebopID, "cowboy-bebop", "2"))
	assert.Equal(t, "/watch/"+bebopID+"/x", watchPath(bebopID, ""), "the site wants some slug and redirects to the right one")
	assert.Equal(t, "cowboy-bebop", slugFromPath("/watch/"+bebopID+"/cowboy-bebop"))
	assert.Equal(t, "", slugFromPath("/watch/"+bebopID))
	assert.Equal(t, "", slugFromPath("/search"))
}

func TestPreferredAudio(t *testing.T) {
	tests := []struct{ env, want string }{
		{"", trackHardSub}, {"sub", trackHardSub}, {"jpn", trackHardSub}, {"klingon", trackHardSub},
		{"dub", trackDub}, {" DUB ", trackDub}, {"dubbed", trackDub}, {"eng", trackDub},
	}
	for _, tt := range tests {
		t.Setenv("GOANIME_MIRURO_AUDIO", tt.env)
		assert.Equal(t, tt.want, preferredAudio(), "GOANIME_MIRURO_AUDIO=%q", tt.env)
	}
}

func TestNormalizeQuality(t *testing.T) {
	t.Parallel()
	tests := map[string]int{"": 0, "best": 0, " Auto ": 0, "hd": 0, "720p": 720, "1080": 1080, "4k 2160p": 2160}
	for in, want := range tests {
		assert.Equal(t, want, normalizeQuality(in), in)
	}
}

func TestResolveRef(t *testing.T) {
	t.Parallel()
	base := "https://cdn.example/a/b/master.m3u8?t=1"
	assert.Equal(t, "https://cdn.example/a/b/index-720.m3u8", resolveRef(base, "index-720.m3u8"))
	assert.Equal(t, "https://cdn.example/root.m3u8", resolveRef(base, "/root.m3u8"))
	assert.Equal(t, "http://other.example/x.m3u8", resolveRef(base, "http://other.example/x.m3u8"))
	assert.Equal(t, "rel.m3u8", resolveRef("::bad", "rel.m3u8"))
	assert.Equal(t, "%zz", resolveRef(base, "%zz"))
}

// ---------------------------------------------------------------------------
// search
// ---------------------------------------------------------------------------

func TestSearchAnime(t *testing.T) {
	t.Parallel()
	s := newStub(t, nil)
	results, err := NewClientForTest(s.srv.URL).SearchAnime(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, results, 2, "the unaired title is dropped")

	bebop := results[0]
	assert.Equal(t, &models.Anime{
		Name:      "Cowboy Bebop",
		URL:       s.srv.URL + "/info/" + bebopID,
		ImageURL:  "https://img.example/" + bebopID + ".jpg",
		Source:    "Miruro",
		MediaType: models.MediaTypeAnime,
		Year:      "1998",
		Audio:     "Dubbed and subtitled",
		AnilistID: 1,
		MalID:     1,
		Details: models.AniListDetails{
			Title: models.Title{English: "Cowboy Bebop", Romaji: "Cowboy Bebop"},
		},
	}, bebop)

	movie := results[1]
	assert.Equal(t, "Kaubōi Bibappu: Tengoku no Tobira", movie.Name, "no English title: the romaji one")
	assert.Equal(t, "Subtitled", movie.Audio, "a null dub count is no dub")
	assert.Equal(t, 5, movie.AnilistID)
	assert.Equal(t, []string{"/search/__data.json"}, s.paths())
}

func TestSearchAnimeFailures(t *testing.T) {
	t.Parallel()
	_, err := NewClientForTest("http://unused.example").SearchAnime(context.Background(), "  ")
	assert.ErrorContains(t, err, "empty query")

	s := newStub(t, nil)
	s.failSearch = http.StatusServiceUnavailable
	_, err = NewClientForTest(s.srv.URL).SearchAnime(context.Background(), "x")
	assert.Equal(t, netx.DiagnosticSourceUnavailable, diagKind(t, err))

	layout := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(dataBody(layoutNode, map[string]any{"mode": "anime"})))
	}))
	defer layout.Close()
	_, err = NewClientForTest(layout.URL).SearchAnime(context.Background(), "x")
	assert.ErrorContains(t, err, "no result list")

	wrong := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(dataBody(map[string]any{"items": "not a list"})))
	}))
	defer wrong.Close()
	_, err = NewClientForTest(wrong.URL).SearchAnime(context.Background(), "x")
	assert.ErrorContains(t, err, "unreadable result list")

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(dataBody(map[string]any{"items": []any{}})))
	}))
	defer empty.Close()
	got, err := NewClientForTest(empty.URL).SearchAnime(context.Background(), "x")
	require.NoError(t, err, "no match is an answer, not a failure")
	assert.Empty(t, got)
}

func TestToResults(t *testing.T) {
	t.Parallel()
	c := NewClientForTest("https://m.example")
	year := 2001
	zero := 0
	sub := 3
	items := []searchItem{
		{ID: bebopID, Status: "FINISHED", SeasonYear: &year},
		{ID: bebopID}, // duplicate
		{ID: "bad id"},
		{ID: ""},
		{ID: "Nameless_session-0123456789abcd", Status: "FINISHED"},
		{ID: "Unaired_session-0123456789abcde", Status: "not_yet_released"},
		{ID: "Zero_season_session-0123456789a", SeasonYear: &zero},
	}
	items[0].Title.Native = "  ネイティブ  "
	items[0].EpisodeCounts.Sub = &sub
	items[0].EpisodeCounts.Dub = &zero
	items[0].ExternalIDs = map[string][]string{"anilist": {"x", " 42 "}, "mal": {"0"}}
	items[1].Title.English = "Duplicate"
	items[2].Title.English = "Bad"
	items[3].Title.English = "Empty"
	items[5].Title.English = "Later"
	items[6].Title.Romaji = "Zero"

	got := c.toResults(items)
	require.Len(t, got, 2)
	assert.Equal(t, "ネイティブ", got[0].Name, "the native title is the last resort")
	assert.Equal(t, "2001", got[0].Year)
	assert.Equal(t, "Subtitled", got[0].Audio, "a zero dub count is no dub")
	assert.Equal(t, 42, got[0].AnilistID, "the first id that parses")
	assert.Equal(t, 0, got[0].MalID, "0 is no id")
	assert.Equal(t, "Zero", got[1].Name)
	assert.Equal(t, "", got[1].Year, "a zero year is no year")
	assert.Equal(t, "", got[1].Audio)
}

func TestSmallHelpers(t *testing.T) {
	t.Parallel()
	one, zero := 1, 0
	assert.True(t, positive(&one))
	assert.False(t, positive(&zero))
	assert.False(t, positive(nil))
	assert.Equal(t, 7, firstID([]string{"", "-3", "0", "7", "9"}), "0 is no id: keep looking")
	assert.Equal(t, 0, firstID(nil))
	assert.Equal(t, "b", firstNonEmpty(" ", "b", "c"))
	assert.Equal(t, "", firstNonEmpty())
}

// ---------------------------------------------------------------------------
// episodes
// ---------------------------------------------------------------------------

func TestGetAnimeEpisodes(t *testing.T) {
	t.Parallel()
	s := newStub(t, nil)
	eps, err := NewClientForTest(s.srv.URL).GetAnimeEpisodes(context.Background(), s.srv.URL+"/info/"+bebopID)
	require.NoError(t, err)
	require.Len(t, eps, 3)

	assert.Equal(t, models.Episode{
		Number:   "1",
		Num:      1,
		URL:      s.srv.URL + "/watch/" + bebopID + "/" + bebopSlug + "?ep=1",
		Title:    models.TitleDetails{English: "Asteroid Blues"},
		Aired:    "1998-04-03",
		Synopsis: "Synopsis of Asteroid Blues",
	}, eps[0], "episode URLs carry the canonical slug from the redirect")
	assert.Equal(t, "1.5", eps[1].Number, "a recap keeps its fraction")
	assert.Equal(t, 1, eps[1].Num)
	assert.True(t, eps[1].IsFiller)
	assert.Equal(t, "2", eps[2].Number, "sorted by number, not by the site's order")

	assert.Equal(t, []string{
		"/watch/" + bebopID + "/x/__data.json",
		"/watch/" + bebopID + "/" + bebopSlug + "/__data.json",
	}, s.paths())
}

func TestGetAnimeEpisodesUsesAKnownSlug(t *testing.T) {
	t.Parallel()
	s := newStub(t, nil)
	_, err := NewClientForTest(s.srv.URL).GetAnimeEpisodes(context.Background(), s.srv.URL+"/watch/"+bebopID+"/"+bebopSlug+"?ep=4")
	require.NoError(t, err)
	assert.Equal(t, []string{"/watch/" + bebopID + "/" + bebopSlug + "/__data.json"}, s.paths(), "no redirect when the slug is already right")
}

func TestGetAnimeEpisodesOfAnUnairedTitle(t *testing.T) {
	t.Parallel()
	s := newStub(t, nil)
	_, err := NewClientForTest(s.srv.URL).GetAnimeEpisodes(context.Background(), unairedID)
	require.ErrorIs(t, err, netx.ErrNoEpisodes)
	var d *netx.SourceDiagnostic
	assert.False(t, errors.As(err, &d), "an unaired title must not read as a layout change")
}

func TestGetAnimeEpisodesFailures(t *testing.T) {
	t.Parallel()
	_, err := NewClientForTest("http://unused.example").GetAnimeEpisodes(context.Background(), "https://miruro.cx/search")
	assert.Error(t, err)

	s := newStub(t, nil)
	s.failWatch = http.StatusInternalServerError
	_, err = NewClientForTest(s.srv.URL).GetAnimeEpisodes(context.Background(), bebopID)
	assert.Equal(t, netx.DiagnosticSourceUnavailable, diagKind(t, err))

	partial := newStub(t, func(_ string, fx *fixture) {
		fx.episodes = []any{map[string]any{"title": "No number"}, episode(3, "Three", false)}
	})
	eps, err := NewClientForTest(partial.srv.URL).GetAnimeEpisodes(context.Background(), bebopID)
	require.NoError(t, err)
	require.Len(t, eps, 1, "an entry without a number is skipped, not the rest of the list")
	assert.Equal(t, "3", eps[0].Number)

	numberless := newStub(t, func(_ string, fx *fixture) {
		fx.episodes = []any{map[string]any{"title": "No number"}, map[string]any{"number": -1}}
	})
	_, err = NewClientForTest(numberless.srv.URL).GetAnimeEpisodes(context.Background(), bebopID)
	assert.ErrorContains(t, err, "none carried a number")

	duplicate := newStub(t, func(_ string, fx *fixture) {
		fx.episodes = []any{episode(1, "A", false), episode(1, "B", false), episode(0, "Zero", false)}
	})
	eps, err = NewClientForTest(duplicate.srv.URL).GetAnimeEpisodes(context.Background(), bebopID)
	require.NoError(t, err)
	require.Len(t, eps, 2, "a repeated number is listed once")
	assert.Equal(t, "0", eps[0].Number, "an episode 0 is an episode")
	assert.Equal(t, "A", eps[1].Title.English, "the first of a repeated number wins")

	for name, page := range map[string]any{
		"no episode list key": map[string]any{"media": map[string]any{}},
		"null episode list":   map[string]any{"episodesData": nil},
		"unreadable list":     map[string]any{"episodesData": "x"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(dataBody(layoutNode, page)))
		}))
		_, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), bebopID)
		srv.Close()
		require.Error(t, err, name)
		assert.Equal(t, netx.DiagnosticParserBroken, diagKind(t, err), name)
	}
}

// ---------------------------------------------------------------------------
// stream
// ---------------------------------------------------------------------------

func episodeURL(s *stub, ep string) string {
	return s.srv.URL + "/watch/" + bebopID + "/" + bebopSlug + "?ep=" + ep
}

func subtitleURLs(t *testing.T, raw string) []string {
	t.Helper()
	var tracks []struct {
		URL      string `json:"url"`
		Language string `json:"language"`
		Label    string `json:"label"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &tracks))
	out := make([]string, 0, len(tracks))
	for _, tr := range tracks {
		out = append(out, tr.URL)
	}
	return out
}

func TestGetEpisodeStreamURLPicksTheFirstSoftSubbedProviderTheSiteRanks(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	s := newStub(t, nil)
	got, meta, err := NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	require.NoError(t, err)
	assert.Equal(t, s.srv.URL+"/cdn/sun/master.m3u8", got, "sun is listed second but ranked first")
	assert.Equal(t, map[string]string{
		"source":     "miruro",
		"audio_lang": "sub",
		"referer":    s.srv.URL + "/ultra/",
		"origin":     s.srv.URL + "/ultra",
		"subtitles":  `[{"url":"` + s.srv.URL + `/subs/sun-en.jpeg","language":"English","label":"English (Full Subtitles / ASS)"}]`,
	}, meta)
}

func TestGetEpisodeStreamURLFallsBackWhenAMirrorRefuses(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	s := newStub(t, func(_ string, fx *fixture) {
		c := fx.cdn["sun"]
		c.refuse = true
		fx.cdn["sun"] = c
	})
	got, meta, err := NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	require.NoError(t, err)
	assert.Equal(t, s.srv.URL+"/cdn/bee/master.m3u8", got)
	assert.Equal(t, s.srv.URL+"/mega/", meta["referer"])
	assert.NotContains(t, meta, "origin", "bee's servers name no Origin")
	assert.Equal(t, []string{s.srv.URL + "/subs/bee-en.vtt", s.srv.URL + "/subs/pt.vtt"}, subtitleURLs(t, meta["subtitles"]),
		"English then Portuguese; Arabic, the site's default, is not offered")
}

func TestGetEpisodeStreamURLSendsTheHeadersEachCDNWants(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	// Drop the Origin from sun's server: the CDN refuses, and the next
	// provider plays. This is the one rule a parser test alone cannot see.
	s := newStub(t, func(base string, fx *fixture) {
		tracks := fx.sources.(map[string]any)["tracks"].([]any)
		sun := tracks[0].(map[string]any)["providers"].([]any)[1].(map[string]any)
		srv := sun["servers"].([]any)[0].(map[string]any)
		srv["headers"] = map[string]any{"Referer": base + "/ultra/"}
	})
	got, _, err := NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	require.NoError(t, err)
	assert.Equal(t, s.srv.URL+"/cdn/bee/master.m3u8", got)
}

func TestGetEpisodeStreamURLHonoursTheAudioPreference(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "dub")
	s := newStub(t, nil)
	got, meta, err := NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	require.NoError(t, err)
	assert.Equal(t, s.srv.URL+"/cdn/beedub/master.m3u8", got)
	assert.Equal(t, "dub", meta["audio_lang"])
	assert.NotContains(t, meta, "subtitles", "a dubbed episode puts no subtitle picker in front of the viewer")

	// The episode has no dub: the subtitled tracks are the fallback.
	noDub := newStub(t, func(_ string, fx *fixture) {
		c := fx.cdn["beedub"]
		c.refuse = true
		fx.cdn["beedub"] = c
	})
	got, meta, err = NewClientForTest(noDub.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(noDub, "1"), "best")
	require.NoError(t, err)
	assert.Equal(t, noDub.srv.URL+"/cdn/sun/master.m3u8", got)
	assert.Equal(t, "sub", meta["audio_lang"], "the metadata says what was picked, not what was asked")
}

func TestGetEpisodeStreamURLSelectsTheRequestedQuality(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	s := newStub(t, nil)
	c := NewClientForTest(s.srv.URL)
	tests := map[string]string{
		"720p":  s.srv.URL + "/cdn/sun/index-720.m3u8",
		"1080":  "https://elsewhere.example/1080.m3u8",
		"480p":  s.srv.URL + "/cdn/sun/master.m3u8",
		"best":  s.srv.URL + "/cdn/sun/master.m3u8",
		"":      s.srv.URL + "/cdn/sun/master.m3u8",
		"360":   s.srv.URL + "/cdn/sun/index-360.m3u8",
		"bogus": s.srv.URL + "/cdn/sun/master.m3u8",
	}
	for quality, want := range tests {
		got, _, err := c.GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), quality)
		require.NoError(t, err, quality)
		assert.Equal(t, want, got, "quality %q", quality)
	}
}

func TestGetEpisodeStreamURLFailures(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	_, _, err := NewClientForTest("http://unused.example").GetEpisodeStreamURL(context.Background(), "https://miruro.cx/watch/"+bebopID+"/x", "best")
	assert.ErrorContains(t, err, "no ?ep=")

	s := newStub(t, nil)
	s.failWatch = http.StatusNotFound
	_, _, err = NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	assert.Equal(t, netx.DiagnosticParserBroken, diagKind(t, err))

	offline := newStub(t, func(_ string, fx *fixture) { fx.sources = nil })
	_, _, err = NewClientForTest(offline.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(offline, "1"), "best")
	assert.ErrorIs(t, err, netx.ErrMediaOffline, "null sources: the site has nothing to play")

	noTracks := newStub(t, func(_ string, fx *fixture) { fx.sources = map[string]any{"tracks": []any{}} })
	_, _, err = NewClientForTest(noTracks.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(noTracks, "1"), "best")
	assert.ErrorIs(t, err, netx.ErrMediaOffline)

	embeds := newStub(t, func(base string, fx *fixture) {
		fx.sources = map[string]any{"tracks": []any{track(trackDub, provider("bun", nil, embedOnly("HD-1", base+"/e")))}}
	})
	_, _, err = NewClientForTest(embeds.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(embeds, "1"), "best")
	assert.ErrorContains(t, err, "only embed players")

	dead := newStub(t, func(_ string, fx *fixture) {
		for name, c := range fx.cdn {
			c.refuse = true
			fx.cdn[name] = c
		}
	})
	_, _, err = NewClientForTest(dead.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(dead, "1"), "best")
	assert.ErrorContains(t, err, "none of the 4 streams tried")
}

func TestGetEpisodeStreamURLStopsAfterMaxCandidates(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	var probes atomic.Int32
	s := newStub(t, func(base string, fx *fixture) {
		var servers []any
		for i := 0; i < maxCandidates+3; i++ {
			servers = append(servers, server("s", nil, base+"/cdn/dead/"+string(rune('a'+i))+".m3u8"))
		}
		fx.sources = map[string]any{"tracks": []any{track(trackHardSub, provider("knob", nil, servers...))}}
		fx.cdn["dead"] = cdn{refuse: true}
	})
	inner := s.srv.Config.Handler
	s.srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/cdn/") {
			probes.Add(1)
		}
		inner.ServeHTTP(w, r)
	})
	_, _, err := NewClientForTest(s.srv.URL).GetEpisodeStreamURL(context.Background(), episodeURL(s, "1"), "best")
	assert.ErrorContains(t, err, "none of the 8 streams tried")
	assert.Equal(t, int32(maxCandidates), probes.Load())
}

func TestGetEpisodeStreamURLHonoursCancellation(t *testing.T) {
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	s := newStub(t, nil)
	c := NewClientForTest(s.srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := c.GetEpisodeStreamURL(ctx, episodeURL(s, "1"), "best")
	assert.ErrorIs(t, err, context.Canceled)

	// Cancelled between the page and the first probe.
	ctx2, cancel2 := context.WithCancel(context.Background())
	inner := s.srv.Config.Handler
	s.srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner.ServeHTTP(w, r)
		if strings.HasSuffix(r.URL.Path, dataSuffix) {
			cancel2()
		}
	})
	_, _, err = c.GetEpisodeStreamURL(ctx2, episodeURL(s, "1"), "best")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAudioOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "dub", audioOf(trackDub))
	assert.Equal(t, "sub", audioOf(trackSoftSub))
	assert.Equal(t, "sub", audioOf(trackHardSub))
}

func TestTrackPlan(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []planStep{
		{track: trackSoftSub, needSubs: true}, {track: trackHardSub}, {track: trackSoftSub}, {track: trackDub},
	}, trackPlan(trackHardSub))
	assert.Equal(t, []planStep{
		{track: trackDub}, {track: trackSoftSub, needSubs: true}, {track: trackHardSub}, {track: trackSoftSub},
	}, trackPlan(trackDub))
}

// candidateCases is the table candidates is held to, shared with the
// metatest in meta_test.go.
func candidateCases() []struct {
	name   string
	data   *sourcesData
	prefer string
	want   []string // provider/server per candidate, in order
} {
	en := []sourceSubtitle{{Language: "en", Label: "English", File: "https://s.example/en.vtt"}}
	hls := func(u string) []sourceStream { return []sourceStream{{URL: u, Format: "hls"}} }
	return []struct {
		name   string
		data   *sourcesData
		prefer string
		want   []string
	}{
		{
			name: "providers by the site's order, unknown ones last, servers as listed",
			data: &sourcesData{
				ProviderOrder: []string{"sun", "bee"},
				Tracks: []sourceTrack{{Track: trackHardSub, Providers: []sourceProvider{
					{Provider: "mystery", Servers: []sourceServer{{Server: "m1", Streams: hls("https://c/m1")}}},
					{Provider: "bee", Servers: []sourceServer{{Server: "b1", Streams: hls("https://c/b1")}, {Server: "b2", Streams: hls("https://c/b2")}}},
					{Provider: "sun", Servers: []sourceServer{{Server: "s1", Streams: hls("https://c/s1")}}},
				}}},
			},
			prefer: trackHardSub,
			want:   []string{"sun/s1", "bee/b1", "bee/b2", "mystery/m1"},
		},
		{
			name: "soft subs with English first, burned in next, soft subs without, dub last",
			data: &sourcesData{Tracks: []sourceTrack{
				{Track: trackDub, Providers: []sourceProvider{{Provider: "d", Servers: []sourceServer{{Server: "d1", Streams: hls("https://c/d1")}}}}},
				{Track: trackSoftSub, Providers: []sourceProvider{
					{Provider: "bare", Servers: []sourceServer{{Server: "x1", Streams: hls("https://c/x1")}}},
					{Provider: "subbed", Subtitles: en, Servers: []sourceServer{{Server: "y1", Streams: hls("https://c/y1")}}},
				}},
				{Track: trackHardSub, Providers: []sourceProvider{{Provider: "h", Servers: []sourceServer{{Server: "h1", Streams: hls("https://c/h1")}}}}},
			}},
			prefer: trackHardSub,
			want:   []string{"subbed/y1", "h/h1", "bare/x1", "d/d1"},
		},
		{
			name: "dub first when preferred",
			data: &sourcesData{Tracks: []sourceTrack{
				{Track: trackHardSub, Providers: []sourceProvider{{Provider: "h", Servers: []sourceServer{{Server: "h1", Streams: hls("https://c/h1")}}}}},
				{Track: trackDub, Providers: []sourceProvider{{Provider: "d", Servers: []sourceServer{{Server: "d1", Streams: hls("https://c/d1")}}}}},
			}},
			prefer: trackDub,
			want:   []string{"d/d1", "h/h1"},
		},
		{
			name: "embeds, non-HTTP URLs, unknown formats and repeats are skipped",
			data: &sourcesData{Tracks: []sourceTrack{{Track: trackHardSub, Providers: []sourceProvider{{Provider: "p", Servers: []sourceServer{
				{Server: "embed"},
				{Server: "ftp", Streams: []sourceStream{{URL: "ftp://c/x", Format: "hls"}}},
				{Server: "blank", Streams: []sourceStream{{URL: "", Format: "hls"}}},
				{Server: "dash", Streams: []sourceStream{{URL: "https://c/x.mpd", Format: "dash"}}},
				{Server: "mp4", Streams: []sourceStream{{URL: " https://c/x.mp4 ", Format: " MP4 "}}},
				{Server: "plain-http", Streams: []sourceStream{{URL: "http://c/y.m3u8", Format: "hls"}}},
				{Server: "repeat", Streams: hls("https://c/x.mp4")},
				{Server: "several", Streams: []sourceStream{
					{URL: "ftp://c/z", Format: "hls"},
					{URL: "https://c/z.mpd", Format: "dash"},
					{URL: "https://c/x.mp4", Format: "mp4"},
					{URL: "https://c/z.m3u8", Format: "hls"},
				}},
			}}}}}},
			prefer: trackHardSub,
			want:   []string{"p/mp4", "p/plain-http", "p/several"},
		},
		{
			name: "a track this source does not know is ignored",
			data: &sourcesData{Tracks: []sourceTrack{
				{Track: "raw", Providers: []sourceProvider{{Provider: "r", Servers: []sourceServer{{Server: "r1", Streams: hls("https://c/r1")}}}}},
			}},
			prefer: trackHardSub,
			want:   nil,
		},
	}
}

func TestCandidates(t *testing.T) {
	t.Parallel()
	for _, tt := range candidateCases() {
		var got []string
		for _, c := range candidates(tt.data, tt.prefer) {
			got = append(got, c.provider+"/"+c.server)
		}
		assert.Equal(t, tt.want, got, tt.name)
	}
}

func TestCandidatesCarryHeadersAndSubtitles(t *testing.T) {
	t.Parallel()
	data := &sourcesData{Tracks: []sourceTrack{
		{Track: trackSoftSub, Providers: []sourceProvider{{
			Provider:  "sun",
			Subtitles: []sourceSubtitle{{Language: "en", File: "https://s/en.vtt"}, {Language: "fr", File: "https://s/fr.vtt"}},
			Servers:   []sourceServer{{Server: "a", Headers: map[string]string{"referer": " https://r/ ", "ORIGIN": "https://o"}, Streams: []sourceStream{{URL: "https://c/a", Format: "hls"}}}},
		}}},
		{Track: trackHardSub, Providers: []sourceProvider{{
			Provider:  "knob",
			Subtitles: []sourceSubtitle{{Language: "en", File: "https://s/en2.vtt"}},
			Servers:   []sourceServer{{Server: "b", Streams: []sourceStream{{URL: "https://c/b", Format: "hls"}}}},
		}}},
	}}
	got := candidates(data, trackHardSub)
	require.Len(t, got, 2)
	assert.Equal(t, "https://r/", got[0].referer, "header names are matched without case")
	assert.Equal(t, "https://o", got[0].origin)
	assert.Equal(t, []sourceSubtitle{{Language: "en", File: "https://s/en.vtt"}}, got[0].subtitles)
	assert.Empty(t, got[1].subtitles, "burned-in subtitles need no file")
	assert.Equal(t, "", got[1].referer)
}

func TestOrderedProvidersDoesNotReorderItsInput(t *testing.T) {
	t.Parallel()
	in := []sourceProvider{{Provider: "b"}, {Provider: "a"}}
	out := orderedProviders(in, map[string]int{"a": 0, "b": 1})
	assert.Equal(t, "a", out[0].Provider)
	assert.Equal(t, "b", in[0].Provider, "the page data is left as it was")
}

func TestProbe(t *testing.T) {
	t.Parallel()
	var gotRange, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRange, gotUA = r.Header.Get("Range"), r.Header.Get("User-Agent")
		switch r.URL.Path {
		case "/ok.m3u8":
			_, _ = w.Write([]byte("  \n#EXTM3U\n#EXT-X-ENDLIST\n"))
		case "/html.m3u8":
			_, _ = w.Write([]byte("<html>no</html>"))
		case "/clip.mp4":
			w.WriteHeader(http.StatusPartialContent)
		case "/gone.mp4":
			w.WriteHeader(http.StatusGone)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := NewClientForTest(srv.URL)

	body, ok := c.probe(context.Background(), candidate{url: srv.URL + "/ok.m3u8", format: "hls"})
	assert.True(t, ok)
	assert.Contains(t, string(body), "#EXT-X-ENDLIST")
	assert.Equal(t, "", gotRange, "a playlist is read whole")
	assert.Equal(t, netx.UserAgent, gotUA)

	_, ok = c.probe(context.Background(), candidate{url: srv.URL + "/html.m3u8", format: "hls"})
	assert.False(t, ok, "a 200 that is not a playlist will not play")

	body, ok = c.probe(context.Background(), candidate{url: srv.URL + "/clip.mp4", format: "mp4"})
	assert.True(t, ok)
	assert.Nil(t, body)
	assert.Equal(t, "bytes=0-0", gotRange, "a file is checked with one byte")

	_, ok = c.probe(context.Background(), candidate{url: srv.URL + "/gone.mp4", format: "mp4"})
	assert.False(t, ok)
	_, ok = c.probe(context.Background(), candidate{url: srv.URL + "/missing.m3u8", format: "hls"})
	assert.False(t, ok)
	_, ok = c.probe(context.Background(), candidate{url: "://bad", format: "hls"})
	assert.False(t, ok)

	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := dead.URL
	dead.Close()
	_, ok = c.probe(context.Background(), candidate{url: deadURL + "/x.m3u8", format: "hls"})
	assert.False(t, ok)
}

func TestProbeGivesUpOnASlowMirror(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	c := NewClientForTest(srv.URL)
	c.probeTimeout = 50 * time.Millisecond
	start := time.Now()
	_, ok := c.probe(context.Background(), candidate{url: srv.URL + "/slow.m3u8", format: "hls"})
	assert.False(t, ok)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestSelectVariant(t *testing.T) {
	t.Parallel()
	base := "https://cdn.example/a/master.m3u8"
	got, ok := selectVariant([]byte(masterPlaylist), base, "720p")
	assert.True(t, ok)
	assert.Equal(t, "https://cdn.example/a/index-720.m3u8", got)

	for _, q := range []string{"best", "", "480"} {
		_, ok := selectVariant([]byte(masterPlaylist), base, q)
		assert.False(t, ok, q)
	}
	_, ok = selectVariant([]byte(mediaPlaylist), base, "720")
	assert.False(t, ok, "a media playlist has no variants")
	_, ok = selectVariant([]byte("#EXTM3U\n#EXT-X-STREAM-INF:RESOLUTION=1280x720\n#comment\n"), base, "720")
	assert.False(t, ok, "a variant line followed by a tag is no variant")
	got, ok = selectVariant([]byte("#EXTM3U\n#EXT-X-STREAM-INF:RESOLUTION=1280x720\n#comment\n"+
		"#EXT-X-STREAM-INF:RESOLUTION=1280x720\nsecond.m3u8\n"), base, "720")
	assert.True(t, ok, "a broken entry does not stop the search")
	assert.Equal(t, "https://cdn.example/a/second.m3u8", got)
	_, ok = selectVariant([]byte("#EXTM3U\n#EXT-X-STREAM-INF:RESOLUTION=1280x720"), base, "720")
	assert.False(t, ok, "a variant line at the end has no URL")
	_, ok = selectVariant([]byte("#EXTM3U\n#EXT-X-STREAM-INF:RESOLUTION=1280x720\n\n"), base, "720")
	assert.False(t, ok, "a variant line followed by a blank has no URL")
	_, ok = selectVariant([]byte("#EXT-X-STREAM-INF:RESOLUTION=1x99999999999999999999\nx.m3u8\n"), base, "720")
	assert.False(t, ok, "an unreadable height is no match")
}

func TestSubtitleLanguage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    sourceSubtitle
		want string
	}{
		{sourceSubtitle{Language: "en"}, "English"},
		{sourceSubtitle{Language: " EN-US "}, "English"},
		{sourceSubtitle{Language: "eng"}, "English"},
		{sourceSubtitle{Language: "x", Label: "English (2)"}, "English"},
		{sourceSubtitle{Language: "pt"}, "Portuguese"},
		{sourceSubtitle{Language: "pt-BR"}, "Portuguese"},
		{sourceSubtitle{Language: "por"}, "Portuguese"},
		{sourceSubtitle{Label: "Português (Brasil)"}, "Portuguese"},
		{sourceSubtitle{Language: "ar", Label: "Arabic"}, ""},
		{sourceSubtitle{Language: "ch", Label: "Chinese (Traditional)"}, ""},
		{sourceSubtitle{Language: "enx"}, ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, subtitleLanguage(tt.s), "%+v", tt.s)
	}
}

func TestWantedSubtitles(t *testing.T) {
	t.Parallel()
	subs := []sourceSubtitle{
		{Language: "ar", File: "https://s/ar", Default: true},
		{Language: "pt", File: "https://s/pt"},
		{Language: "en", File: "relative.vtt"},
		{Language: "en", File: "https://s/en2"},
		{Language: "en", File: "https://s/en1", Default: true},
		{Language: "en", File: " "},
	}
	var files []string
	for _, s := range wantedSubtitles(subs) {
		files = append(files, s.File)
	}
	assert.Equal(t, []string{"https://s/en1", "https://s/en2", "https://s/pt"}, files,
		"English before Portuguese, a language's default first, nothing that cannot be fetched")
	assert.Empty(t, wantedSubtitles(nil))
}

func TestEncodeSubtitles(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", encodeSubtitles(nil))
	got := encodeSubtitles([]sourceSubtitle{
		{Language: "en", Label: " English (CR) ", File: " https://s/en.vtt "},
		{Language: "pt-BR", File: "https://s/pt.vtt"},
	})
	assert.JSONEq(t, `[
		{"url":"https://s/en.vtt","language":"English","label":"English (CR)"},
		{"url":"https://s/pt.vtt","language":"Portuguese","label":"Portuguese"}
	]`, got, "the language is the plain name; a missing label falls back to it")
}
