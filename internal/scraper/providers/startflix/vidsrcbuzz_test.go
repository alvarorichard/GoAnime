package startflix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The 2Embed tests run the whole chain — embed page (var Q) → signed play
// call → master → rendition checks → first-segment range probe — against one
// httptest server. The routeVIDSRC transport (vidsrc_test.go) sends even the
// real-world hosts the chain names (relay3.videm.xyz segments) back to the
// same server, so a test never touches the network. Gate and base are one
// env: GOANIME_2EMBED_URL names the fake server, which turns the chain on
// and points it here in a single assignment.

// server2Embed is one entry of Q["ssr"]["servers"].
type server2Embed struct {
	ref  string
	name string
}

// variant2Embed is one rendition: its /_stream id, the height its
// RESOLUTION names, and whether its playlist is the ad-laced kind (2026-10-21:
// whole renditions whose every "segment" sat on TikTok's ad CDN).
type variant2Embed struct {
	id     string
	height int
	ad     bool
}

// play2EmbedCall is one api.php a=play request as the fake recorded it.
type play2EmbedCall struct {
	ref   string
	fresh bool
}

// fake2Embed hosts the whole chain. Tests set the flags they exercise —
// unavailable refs, dying mints, ad-laced variants, a segment relay that
// starts broken — and assert on the recorded request lists afterwards.
type fake2Embed struct {
	mu sync.Mutex

	kind    MediaKind
	id      string // movie id, or the tv TMDB digits
	season  int    // tv only
	episode int
	token   string
	servers []server2Embed

	unavailable  map[string]bool            // ref → play answers {"error":"unavailable"}
	diesUnfresh  map[string]bool            // ref → first (fresh-less) play answers 500
	renditions   map[string][]variant2Embed // master id → renditions in manifest order
	variantMedia map[string]variant2Embed   // variant id → its media playlist
	segFailFirst int                        // the N first segment probes answer 500
	noConfig     bool                       // embed page carries no `var Q`

	playCalls   []play2EmbedCall
	embedPaths  []string
	masterHits  []string
	variantHits []string
	segHits     []string
	otherHits   []string // anything else the chain asked for
}

// addServer registers one Q server: its ref goes on the embed page and its
// master (with these renditions) answers play. Master ids derive from the
// ref, so tests share the ids without collision. The maps lazily initialise
// here because tests build `fake2Embed` literals and set behavior flags
// after the first addServer — every writer runs after this line.
func (f *fake2Embed) addServer(ref, name string, variants ...variant2Embed) {
	if f.unavailable == nil {
		f.unavailable = map[string]bool{}
	}
	if f.diesUnfresh == nil {
		f.diesUnfresh = map[string]bool{}
	}
	if f.renditions == nil {
		f.renditions = map[string][]variant2Embed{}
	}
	if f.variantMedia == nil {
		f.variantMedia = map[string]variant2Embed{}
	}
	f.servers = append(f.servers, server2Embed{ref: ref, name: name})
	masterID := "master-" + ref
	f.renditions[masterID] = variants
	for _, v := range variants {
		f.variantMedia[v.id] = v
	}
}

// setRenditions replaces one master's rendition list, registering every
// variant for the media playlist route the way addServer does. Tests use it
// to re-model an already-registered server's manifest.
func (f *fake2Embed) setRenditions(masterID string, variants ...variant2Embed) {
	if f.variantMedia == nil {
		f.variantMedia = map[string]variant2Embed{}
	}
	if f.renditions == nil {
		f.renditions = map[string][]variant2Embed{}
	}
	f.renditions[masterID] = variants
	for _, v := range variants {
		f.variantMedia[v.id] = v
	}
}

// embedPage renders the player page the chain parses: a `var Q = {...}`
// assignment shaped exactly like the live one (2026-10-21), with the
// servers under Q["ssr"]["servers"].
func (f *fake2Embed) embedPage() string {
	if f.noConfig {
		return `<html lang="en"><head><title>VidSrc Player</title><base href="/pl/"></head><body>broken page</body></html>`
	}
	kindWord := "movie"
	if f.kind == KindSeries {
		kindWord = "tv"
	}
	servers := make([]map[string]string, 0, len(f.servers))
	for _, s := range f.servers {
		servers = append(servers, map[string]string{"ref": s.ref, "name": s.name})
	}
	serversJSON, err := json.Marshal(servers)
	if err != nil {
		panic(err)
	}
	middle := ""
	if f.kind == KindSeries {
		middle = `,"s":` + strconv.Itoa(f.season) + `,"e":` + strconv.Itoa(f.episode)
	}
	cfg := `{"type":` + jsonQuote(kindWord) + `,"id":` + jsonQuote(f.id) + middle +
		`,"t":` + jsonQuote(f.token) +
		`,"ssr":{"servers":` + string(serversJSON) + `}}`
	return `<html lang="en"><head><title>VidSrc Player</title><base href="/pl/"></head><body>` +
		`<script>var Q = ` + cfg + `;</script></body></html>`
}

// masterBody renders one master playlist in manifest order.
func (f *fake2Embed) masterBody(masterID string) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	for _, v := range f.renditions[masterID] {
		width := v.height * 16 / 9
		if width == 0 {
			width = 1280
		}
		b.WriteString("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1247000,RESOLUTION=" +
			strconv.Itoa(width) + "x" + strconv.Itoa(v.height) + "\n/_stream?id=" + v.id + "\n")
	}
	return b.String()
}

// variantBody renders one rendition playlist. The clean kind pages its media
// through the live relay origin (rewritten to this server); the ad kind
// mirrors the 2026-10-21 finding: TikTok ad CDN, .image payloads.
func (f *fake2Embed) variantBody(v variant2Embed) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:20\n#EXT-X-PLAYLIST-TYPE:VOD\n")
	for i := 0; i < 2; i++ {
		b.WriteString("#EXTINF:10.010,\n")
		if v.ad {
			fmt.Fprintf(&b, "https://p16-ad-site-sign-sg.tiktokcdn.com/ad-site-i18n-sg/2026%02d~tplv-d5opwmad15-ttam-origin.image?x-expires=1&sig=%d\n", i, i)
		} else {
			fmt.Fprintf(&b, "https://relay3.videm.xyz/seg?u=https%%3A%%2F%%2Fmedia.example%%2Fseg-%d-v1-a1.ts&x=%d\n", i, i)
		}
	}
	return b.String()
}

// answerPlay writes the play verdict for one recorded call.
func (f *fake2Embed) answerPlay(t *testing.T, w http.ResponseWriter, call play2EmbedCall) {
	t.Helper()
	if f.unavailable[call.ref] {
		_, _ = w.Write([]byte(`{"error":"unavailable"}`))
		return
	}
	if !call.fresh && f.diesUnfresh[call.ref] {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream mint died"))
		return
	}
	_, _ = w.Write([]byte(`{"url":"/_stream?id=master-` + call.ref + `","type":"hls"}`))
}

// start runs the fake server for the test's lifetime.
func (f *fake2Embed) start(t *testing.T) *httptest.Server {
	t.Helper()
	f.mu.Lock()
	if f.token == "" {
		f.token = "testtoken-buzz"
	}
	f.mu.Unlock()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case strings.HasPrefix(r.URL.Path, "/embed/movie/"):
			f.embedPaths = append(f.embedPaths, r.URL.Path)
			if f.id != "" {
				assert.Equal(t, "/embed/movie/"+f.id, r.URL.Path, "the embed page names the movie id")
			}
			_, _ = w.Write([]byte(f.embedPage()))
		case strings.HasPrefix(r.URL.Path, "/embed/tv/"):
			f.embedPaths = append(f.embedPaths, r.URL.Path)
			if f.season > 0 {
				want := fmt.Sprintf("/embed/tv/%s/%d/%d", f.id, f.season, f.episode)
				assert.Equal(t, want, r.URL.Path, "the tv embed page names the season and episode")
			}
			_, _ = w.Write([]byte(f.embedPage()))
		case strings.HasPrefix(r.URL.Path, "/pl/api.php"):
			q := r.URL.Query()
			assert.Equal(t, "play", q.Get("a"), "the play endpoint serves a=play")
			assert.Equal(t, f.token, q.Get("t"), "the play call carries the signed token")
			call := play2EmbedCall{ref: q.Get("ref"), fresh: q.Get("fresh") != ""}
			f.playCalls = append(f.playCalls, call)
			f.answerPlay(t, w, call)
		case strings.HasPrefix(r.URL.Path, "/_stream"):
			id := r.URL.Query().Get("id")
			if _, ok := f.renditions[id]; ok {
				f.masterHits = append(f.masterHits, id)
				_, _ = w.Write([]byte(f.masterBody(id)))
				return
			}
			if v, ok := f.variantMedia[id]; ok {
				f.variantHits = append(f.variantHits, id)
				_, _ = w.Write([]byte(f.variantBody(v)))
				return
			}
			http.NotFound(w, r)
		case strings.HasPrefix(r.URL.Path, "/seg"):
			f.segHits = append(f.segHits, r.URL.String())
			if f.segFailFirst > 0 {
				f.segFailFirst--
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("relay crash"))
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte{0x47, 0x40, 0x01, 0x10, 0x47, 0x40, 0x02, 0x12})
		default:
			f.otherHits = append(f.otherHits, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// buzzHTTPClientForTest points the chain at the fake server for the test's
// lifetime and turns the gate on by naming the fake in the enabling env.
func buzzHTTPClientForTest(t *testing.T, srv *httptest.Server) *url.URL {
	t.Helper()
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	t.Setenv(Default2EmbedEnv, srv.URL)
	prev := buzzHTTPFn
	buzzHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { buzzHTTPFn = prev })
	return target
}

// oneCleanServer is the baseline fixture: a single server whose rendition is
// a 720p playlist of relay segments.
func oneCleanServer(ref string) func(f *fake2Embed) {
	return func(f *fake2Embed) {
		f.addServer(ref, "Server VNY", variant2Embed{id: "var-" + ref, height: 720})
	}
}

func Test2EmbedMovieResolves(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	srv := f.start(t)
	target := buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	// The validated rendition is what the player gets, not the master.
	assert.Equal(t, srv.URL+"/_stream?id=var-ref-vny", stream.URL)
	assert.Equal(t, srv.URL+"/embed/movie/tt0315327", stream.Referer,
		"the referer is the embed page URL, as the page's meta referrer implies")
	assert.Equal(t, target.Host, stream.Host, "the bare host, as the vidsrc chain reports it")
	assert.Equal(t, 720, stream.Height)
	assert.Equal(t, AudioSubtitled, stream.Audio, "the fallback serves the original audio")

	require.Len(t, f.embedPaths, 1)
	assert.Equal(t, []play2EmbedCall{{ref: "ref-vny"}}, f.playCalls,
		"one play call, no fresh retry when the first mint plays")
	assert.Equal(t, []string{"master-ref-vny"}, f.masterHits)
	assert.Equal(t, []string{"var-ref-vny"}, f.variantHits,
		"the rendition playlist is read before the stream is handed over")
	require.Len(t, f.segHits, 1, "the first segment is range-probed")
	assert.Empty(t, f.otherHits)
}

func Test2EmbedTVResolvesWithSeasonEpisode(t *testing.T) {
	f := &fake2Embed{kind: KindSeries, id: "147826", season: 2, episode: 7}
	oneCleanServer("ref-xpm")(f)
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindSeries, "", 147826, 2, 7)
	require.NoError(t, err)

	assert.Equal(t, 720, stream.Height)
	require.Len(t, f.embedPaths, 1)
	assert.Equal(t, "/embed/tv/147826/2/7", f.embedPaths[0],
		"the tv embed path carries the season/episode pair")
	assert.Equal(t, []play2EmbedCall{{ref: "ref-xpm"}}, f.playCalls)
}

func Test2EmbedRotatesThroughUnavailableServers(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	f.addServer("ref-dead", "Server XPM1", variant2Embed{id: "var-dead", height: 1080})
	f.addServer("ref-working", "Server VNY", variant2Embed{id: "var-working", height: 720})
	f.unavailable["ref-dead"] = true
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, srv.URL+"/_stream?id=var-working", stream.URL)
	assert.Equal(t, []play2EmbedCall{
		{ref: "ref-dead"},
		{ref: "ref-working"},
	}, f.playCalls, "the unavailable ref costs exactly one call, no fresh retry")
}

func Test2EmbedFreshRetryRecoversDyingMint(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	f.diesUnfresh["ref-vny"] = true
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, srv.URL+"/_stream?id=var-ref-vny", stream.URL)
	assert.Equal(t, []play2EmbedCall{
		{ref: "ref-vny"},
		{ref: "ref-vny", fresh: true},
	}, f.playCalls, "one fresh=1 retry mirrors the player page's own recovery")
}

func Test2EmbedSkipsAdLacedRenditionForTheCleanOne(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	// The live shape that motivated the walk: a master whose first (tallest)
	// rendition is the ad playlist and whose second is the real one
	// (2026-10-21, The Godfather).
	f.setRenditions("master-ref-vny",
		variant2Embed{id: "var-ad-1080", height: 1080, ad: true},
		variant2Embed{id: "var-clean-720", height: 720})
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, srv.URL+"/_stream?id=var-clean-720", stream.URL,
		"the master's tallest rendition was an ad playlist; the clean one plays")
	assert.Equal(t, 720, stream.Height, "the height comes from the validated rendition")
	assert.Equal(t, []string{"var-ad-1080", "var-clean-720"}, f.variantHits,
		"the ad playlist was read and refused, then the clean one validated")
	require.Len(t, f.segHits, 1, "no segment of the ad playlist is ever fetched, the clean one's is")
}

func Test2EmbedSkipsServerWithOnlyAdLacedRenditions(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	f.addServer("ref-ad", "Server VNR",
		variant2Embed{id: "var-ad-a", height: 1080, ad: true},
		variant2Embed{id: "var-ad-b", height: 720, ad: true})
	f.addServer("ref-clean", "Server VNY", variant2Embed{id: "var-clean", height: 720})
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, srv.URL+"/_stream?id=var-clean", stream.URL)
	assert.Equal(t, []play2EmbedCall{
		{ref: "ref-ad"},
		{ref: "ref-ad", fresh: true},
		{ref: "ref-clean"},
	}, f.playCalls, "the ad-only server is walked, retried, then the next server plays")
	require.Len(t, f.segHits, 1, "the ad playlist costs no segment probe, the clean one's only")
}

func Test2EmbedSkipsDeadSegmentRelayToNextServer(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	f.addServer("ref-relaydead", "Server VNY", variant2Embed{id: "var-deadrelay", height: 720})
	f.addServer("ref-ok", "Server XPM1", variant2Embed{id: "var-ok", height: 1080})
	// The 2026-10-21 state in miniature: the first server's relay answers
	// 500 on its segment (in both of its attempts), the second serves.
	f.segFailFirst = 2
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.NoError(t, err)

	assert.Equal(t, srv.URL+"/_stream?id=var-ok", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	assert.Equal(t, []play2EmbedCall{
		{ref: "ref-relaydead"},
		{ref: "ref-relaydead", fresh: true},
		{ref: "ref-ok"},
	}, f.playCalls)
	require.Len(t, f.segHits, 3, "two failed probes on the dead relay, one serving one")
}

func Test2EmbedHonestErrorWhenConfigMissing(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	f.noConfig = true
	srv := f.start(t)
	buzzHTTPClientForTest(t, srv)

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	require.Error(t, err)
	assert.Nil(t, stream)
	assert.ErrorContains(t, err, "no config object")
	assert.Empty(t, f.playCalls, "a page without the config object is never played from")
}

func Test2EmbedGateOffByDefault(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	srv := f.start(t)
	// The env is the gate: unset, the chain must not even reach the fake.
	// Point the seam at it anyway, so any accidental request shows up.
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	t.Setenv(Default2EmbedEnv, "")
	prev := buzzHTTPFn
	buzzHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { buzzHTTPFn = prev })

	stream, err := resolve2Embed(context.Background(), KindMovie, "tt0315327", 310, 0, 0)
	assert.ErrorIs(t, err, err2EmbedDisabled)
	assert.Nil(t, stream)
	assert.Empty(t, f.playCalls)
	assert.Empty(t, f.embedPaths, "the gate keeps the chain off the network entirely")
}

// Test2EmbedRidesBehindVIDSRCFallback wires the whole exported surface: the
// vidsrc family misses (its endpoint is outside this fake), then the 2Embed
// chain plays the title. It also pins the caller contract — a 2Embed success
// replaces the vidsrc miss.
func Test2EmbedRidesBehindVIDSRCFallback(t *testing.T) {
	f := &fake2Embed{kind: KindMovie, id: "tt0315327"}
	oneCleanServer("ref-vny")(f)
	srv := f.start(t)

	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	t.Setenv(Default2EmbedEnv, srv.URL)
	prevBuzz := buzzHTTPFn
	buzzHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { buzzHTTPFn = prevBuzz })
	prevVIDSRC := vidsrcHTTPFn
	vidsrcHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { vidsrcHTTPFn = prevVIDSRC })

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err, "the vidsrc miss hands over to 2Embed, which plays")
	assert.Equal(t, srv.URL+"/_stream?id=var-ref-vny", stream.URL)
	assert.Equal(t, []play2EmbedCall{{ref: "ref-vny"}}, f.playCalls)
	assert.Contains(t, f.otherHits, "/vs_src.php",
		"the vidsrc chain actually ran and missed against this fake first")
}

// Test2EmbedEpisodeRidesBehindVIDSRCFallback is EpisodeFallback's wiring
// twin of the movie test above.
func Test2EmbedEpisodeRidesBehindVIDSRCFallback(t *testing.T) {
	f := &fake2Embed{kind: KindSeries, id: "147826", season: 2, episode: 7}
	oneCleanServer("ref-xpm")(f)
	srv := f.start(t)

	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	t.Setenv(Default2EmbedEnv, srv.URL)
	prevBuzz := buzzHTTPFn
	buzzHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { buzzHTTPFn = prevBuzz })
	prevVIDSRC := vidsrcHTTPFn
	vidsrcHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { vidsrcHTTPFn = prevVIDSRC })

	stream, err := EpisodeFallback(context.Background(), 147826, 2, 7)
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/_stream?id=var-ref-xpm", stream.URL)
	require.Len(t, f.embedPaths, 1)
	assert.Equal(t, "/embed/tv/147826/2/7", f.embedPaths[0])
}

// Test2EmbedLiveShawshank is the opt-in integration check: the chain must
// resolve The Shawshank Redemption (IMDb tt0111161) against the live
// videosrc.buzz player. Status note for whoever runs this: on 2026-10-21 the
// source's renditions were ad-laced and its segment relay answered 500, so
// the chain reports "all N rendition(s) unusable" — the test starts passing
// the day the recordings play again, with no code change.
func Test2EmbedLiveShawshank(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") != "1" {
		t.Skip("live test; run with GOANIME_LIVE=1")
	}
	t.Setenv(Default2EmbedEnv, Default2EmbedBase)
	prev := buzzHTTPFn
	buzzHTTPFn = buzzHTTPDefault
	t.Cleanup(func() { buzzHTTPFn = prev })

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	stream, err := resolve2Embed(ctx, KindMovie, "tt0111161", 278, 0, 0)
	require.NoError(t, err, "the 2Embed chain must resolve movie/tt0111161 again")

	assert.NotEmpty(t, stream.Height, "the validated rendition names a real resolution")
	assert.Equal(t, "vidsrc.buzz", stream.Host)
	assert.Equal(t, Default2EmbedBase+"/embed/movie/tt0111161", stream.Referer)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stream.URL, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", stream.Referer)
	resp, err := buzzHTTPFn().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "stream status: %d", resp.StatusCode)
	assert.True(t, strings.Contains(string(body), "#EXTM3U"), "the stream URL serves a playlist")
}
