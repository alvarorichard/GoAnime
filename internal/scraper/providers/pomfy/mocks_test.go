package pomfy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
)

// Every fake below is a recording of how the production service answers,
// stamped with the day it was captured. The package's tests stand them up
// instead of reaching the network; the meta-tests at the bottom of
// meta_test.go prove the fakes only speak the shapes the real production
// behavior was pinned with, so a fixture that drifts away breaks loudly.

// --- fakeByseCDN: the four byse endpoints with the captcha gate ---

// ByseChallenge is one fake CDN's challenge wiring: which nonce to hand out
// and how hard the PoW is. It is a package-level shape so multiple tests can
// share one definition.
type ByseChallenge struct {
	Nonce string
	Bits  int
}

// fakeByseCDN serves /api/videos/<code>/embed/{details,captcha,captcha/verify,
// playback}. The verify endpoint judges the submitted solution the way the
// server does (hashing OUR nonce), and the playback endpoint is gated by the
// captcha token verify issues. The knobs table lets a scenario break one
// stage — a status/body substitution happens BEFORE the fake's own logic.
type fakeByseCDN struct {
	srv      *httptest.Server
	chal     ByseChallenge
	issued   map[string]bool
	mu       sync.Mutex
	refuseSo bool

	captchaPOSTs  int
	verifyPOSTs   int
	playbackPOSTs int
	detailsPOSTs  int
}

// byseKnobs substitutes one stage's response. Zero values mean "serve the
// recorded behavior".
type byseKnobs struct {
	detailsStatus  int // non-200 makes the details stage fail loudly
	detailsBody    string
	captchaStatus  int
	captchaBody    string
	verifyStatus   int
	verifyBody     string
	playbackStatus int
	playbackBody   string
	// playbackSeal replaces the sealed record with a real one a test built,
	// when the failure lives INSIDE the record rather than around it.
	playbackSeal *playbackSeal
}

func newFakeByseCDN(t *testing.T, nonce string, bits int, seal playbackSeal) *fakeByseCDN {
	return newFakeByseCDNWith(t, nonce, bits, seal, byseKnobs{})
}

func newFakeByseCDNWith(t *testing.T, nonce string, bits int, seal playbackSeal, knobs byseKnobs) *fakeByseCDN {
	t.Helper()
	fake := &fakeByseCDN{
		issued: make(map[string]bool),
		chal:   ByseChallenge{Nonce: nonce, Bits: bits},
	}
	mux := http.NewServeMux()
	base := "/api/videos/{code}/embed"

	write := func(w http.ResponseWriter, status int, body string) {
		if status == 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
			return
		}
		if status >= 300 {
			http.Error(w, body, status)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}

	mux.HandleFunc("POST "+base+"/details", func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.detailsPOSTs++
		fake.mu.Unlock()
		code := r.PathValue("code")
		if knobs.detailsStatus != 0 || knobs.detailsBody != "" {
			status := knobs.detailsStatus
			if status == 0 {
				status = http.StatusOK
			}
			write(w, status, knobs.detailsBody)
			return
		}
		_, _ = fmt.Fprintf(w, `{"code":%q,"title":"Fake Title","embed_frame_url":"https://frame.test/f/%s"}`, code, code)
	})
	mux.HandleFunc("POST "+base+"/captcha", func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.captchaPOSTs++
		fake.mu.Unlock()
		if knobs.captchaStatus != 0 || knobs.captchaBody != "" {
			status := knobs.captchaStatus
			if status == 0 {
				status = http.StatusOK
			}
			write(w, status, knobs.captchaBody)
			return
		}
		if r.Header.Get("X-Embed-Origin") == "" || r.Header.Get("X-Embed-Parent") == "" {
			http.Error(w, `{"error":"no embed identity"}`, http.StatusForbidden)
			return
		}
		_, _ = fmt.Fprintf(w,
			`{"pow_nonce":%q,"pow_difficulty":%d,"pow_token":"tok-1","expires_in":1800,"algorithm":"sha256-leading-zero-bits"}`,
			fake.chal.Nonce, fake.chal.Bits)
	})
	mux.HandleFunc("POST "+base+"/captcha/verify", func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.verifyPOSTs++
		fake.mu.Unlock()
		if knobs.verifyStatus != 0 || knobs.verifyBody != "" {
			status := knobs.verifyStatus
			if status == 0 {
				status = http.StatusOK
			}
			write(w, status, knobs.verifyBody)
			return
		}
		var body struct {
			PowToken    string      `json:"pow_token"`
			Solution    string      `json:"solution"`
			Fingerprint fingerprint `json:"fingerprint"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, `{"error":"bad body"}`, http.StatusBadRequest)
			return
		}
		if body.PowToken != "tok-1" {
			http.Error(w, `{"error":"unknown token"}`, http.StatusForbidden)
			return
		}
		if fake.refuseSo || !checkNonce(fake.chal.Nonce+":", body.Solution, fake.chal.Bits) {
			http.Error(w, `{"error":"rejected"}`, http.StatusForbidden)
			return
		}
		token := "captcha-ok-" + body.Fingerprint.Token
		fake.mu.Lock()
		fake.issued[token] = true
		fake.mu.Unlock()
		_, _ = fmt.Fprintf(w, `{"status":"ok","token":%q}`, token)
	})
	mux.HandleFunc("POST "+base+"/playback", func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.playbackPOSTs++
		fake.mu.Unlock()
		if knobs.playbackSeal != nil {
			fake.mu.Lock()
			ok := fake.issued[r.Header.Get("X-Captcha-Token")]
			fake.mu.Unlock()
			if !ok {
				http.Error(w, `{"error":"no captcha token"}`, http.StatusForbidden)
				return
			}
			w.Write([]byte(`{"playback":`))
			_, _ = w.Write(mustJSON(t, *knobs.playbackSeal))
			_, _ = w.Write([]byte(`}`))
			return
		}
		if knobs.playbackStatus != 0 || knobs.playbackBody != "" {
			status := knobs.playbackStatus
			if status == 0 {
				status = http.StatusOK
			}
			write(w, status, knobs.playbackBody)
			return
		}
		fake.mu.Lock()
		ok := fake.issued[r.Header.Get("X-Captcha-Token")]
		fake.mu.Unlock()
		if !ok {
			http.Error(w, `{"error":"no captcha token"}`, http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"playback":%s}`, mustJSON(t, seal))
	})
	fake.srv = httptest.NewServer(mux)
	t.Cleanup(fake.srv.Close)
	return fake
}

// --- the world: catalog site + byse CDN over one routed client ---

// hostRoutedTransport maps each authority (host:port) onto its own test
// server, so the recorded catalog fixture can keep naming cdn.test while
// every resolution stays inside the process.
type hostRoutedTransport struct {
	targets map[string]*url.URL
}

func (r hostRoutedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, ok := r.targets[req.URL.Host]
	if !ok {
		return nil, fmt.Errorf("pomfy test world: no route for host %q", req.URL.Host)
	}
	out := req.Clone(req.Context())
	out.URL.Scheme, out.URL.Host, out.Host = target.Scheme, target.Host, target.Host
	return http.DefaultTransport.RoundTrip(out)
}

// fakeWorld pairs a recorded catalog site with a fake byse CDN and the
// routed client that reads them.
type fakeWorld struct {
	site *fakeCatalogSite
	cdn  *fakeByseCDN
	hc   *http.Client
}

func newFakeWorld(t *testing.T) *fakeWorld {
	t.Helper()
	world := &fakeWorld{
		site: newFakeCatalogSite(t),
		cdn:  newFakeByseCDN(t, "ref", 6, sealFixture(t, recordJSON(t))),
	}
	target, perr := url.Parse(world.site.srv.URL)
	if perr != nil {
		t.Fatal(perr)
	}
	cdnURL, perr := url.Parse(world.cdn.srv.URL)
	if perr != nil {
		t.Fatal(perr)
	}
	world.hc = &http.Client{Transport: hostRoutedTransport{targets: map[string]*url.URL{
		"pomfy.my":       target,
		"cdn.test":       cdnURL,
		cdnUHost(cdnURL): cdnURL,
	}}}
	return world
}

// cdnUHost extracts the authority (host:port) of the fake CDN's URL, so the
// direct-by-URL resolves that skip the catalog route on their own host.
func cdnUHost(u *url.URL) string {
	return u.Host
}

func (w *fakeWorld) client() *Client {
	return NewClientForTest(w.hc, "https://pomfy.my")
}

// recordingTransport concedes nothing: it snapshots request headers, so a
// meta-test can check the headers the resolver really sent.
type recordingTransport struct {
	inner http.RoundTripper
	mu    sync.Mutex
	seen  []http.Header
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.seen = append(r.seen, req.Header.Clone())
	r.mu.Unlock()
	return r.inner.RoundTrip(req)
}

// --- fake catalog + watch + title payloads (the pomfy.my site) ---

// recordedCatalog, recordedWatchSources and recordedMovieTitle are the
// recorded site shapes, byte-for-byte enough for both unit tests and the
// meta-tests that pin the mocks.
const recordedCatalog = `{"filmes":[550,403,603,604,624860],"series":[1396,1399]}`

// Recorded tmdb-season for one season: epNums lists the TMDB season, sbEps
// the subset really on the CDN, tmdbEps the names.
const recordedSeason = `{"epNums":[1,2,3],"sbEps":[1,3],"tmdbEps":{
	"1":{"name":"Seven Thirty-Seven","still":"https://image.tmdb.org/t/p/w300/6Uo1z56uKnX60JXvYYFWACHR31u.jpg"},
	"2":{"name":"Grilled","still":"https://image.tmdb.org/t/p/w300/th3SNe0gNgRguv8VveLd4f2JcaH.jpg"},
	"3":{"name":"Bit by a Dead Bee","still":"https://image.tmdb.org/t/p/w300/abgJTOWYPdZGxAWhsg6lmtM8qcU.jpg"}}}`

// Recorded watch-sources for a movie (kind/label/embedUrl/token), with a
// flyfile entry that must be dropped by the client-side filter.
const recordedWatchSources = `{"mode":"direct","sources":[
	{"kind":"byse","label":"Servidor 1","description":"Mais rápido e com boa qualidade","embedUrl":"https://cdn.test/e/gkcq8gbiwh4r","token":"gkcq8gbiwh4r.aabbcc"},
	{"kind":"flyfile","label":"Servidor 2","description":"Lento, mas com ótima qualidade","embedUrl":"https://fly.test/x","token":"z"}]}`

// the movie title fixture mirrors assistir/<id>/__data.json (nodes[1].data is
// devalue: object keys point into the array by index).
var movieTitleFixture = compact(`{"nodes":[
  {"data":[{"topbar":0},null,[]]},
  {"data":[
	{"id":1,"tipo":2,"title":3,"year":4,"overview":5,"image":6,"ogTitle":7,"availableSeasons":8,"episodeNumbersMap":9,"backdropPath":10,"watch":11},
	"550","filme","Clube da Luta","1999",
	"Um homem deprimido que sofre de insônia conhece um estranho vendedor de sabonetes chamado Tyler Durden.",
	"/c6OLXfKAk5BKeR6broC8pYiCquX.jpg","Clube da Luta (1999) — Pomfy",[],{},null,
	{"mode":12,"sources":13},"direct",[14],
	{"kind":15,"label":16,"description":17,"embedUrl":18,"token":19},
	"byse","Servidor 1","Mais rápido e com boa qualidade",
	"https://pomfy-cdn.shop/e/gkcq8gbiwh4r","gkcq8gbiwh4r.deadbeef"
  ]}
]}`)

// compact squeezes the multiline fixture literals into one line.
func compact(s string) string {
	joined := strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(joined)
}

// fakeCatalogSite serves catalogo-ids, watch-sources, and the two title
// payload flavors off the SAME host (in production assistir/__data.json and
// serie/__data.json live on pomfy.my). knobs let a scenario break one
// endpoint without rewriting the fixture.
type fakeCatalogSite struct {
	srv          *httptest.Server
	moviesWCalls int
	titleHits    []string // path suffixes requested
	catalogCalls int
	mu           sync.Mutex
	breakCatalog bool
	breakWatch   bool
	breakTitles  bool
}

func newFakeCatalogSite(t *testing.T) *fakeCatalogSite {
	t.Helper()
	site := &fakeCatalogSite{}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/catalogo-ids", func(w http.ResponseWriter, _ *http.Request) {
		site.mu.Lock()
		site.catalogCalls++
		broken := site.breakCatalog
		site.mu.Unlock()
		if broken {
			http.Error(w, `{"error":"catalog down"}`, http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(recordedCatalog))
	})
	mux.HandleFunc("GET /api/watch-sources", func(w http.ResponseWriter, r *http.Request) {
		site.mu.Lock()
		site.moviesWCalls++
		broken := site.breakWatch
		site.mu.Unlock()
		if broken {
			http.Error(w, `{"error":"site down"}`, http.StatusBadGateway)
			return
		}
		q := r.URL.Query()
		if q.Get("id") == "" || q.Get("tipo") == "" {
			http.Error(w, `{"error":"missing id/tipo"}`, http.StatusBadRequest)
			return
		}
		if q.Get("tipo") == "serie" && (q.Get("temporada") == "" || q.Get("episodio") == "") {
			http.Error(w, `{"error":"missing temporada/episodio"}`, http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(recordedWatchSources))
	})
	mux.HandleFunc("GET /assistir/", func(w http.ResponseWriter, r *http.Request) {
		site.mu.Lock()
		site.titleHits = append(site.titleHits, r.URL.Path)
		broken := site.breakTitles
		site.mu.Unlock()
		if broken {
			http.Error(w, `{"error":"gone"}`, http.StatusNotFound)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/__data.json") {
			_, _ = w.Write([]byte(movieTitleFixture))
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /serie/", func(w http.ResponseWriter, r *http.Request) {
		site.mu.Lock()
		site.titleHits = append(site.titleHits, r.URL.Path)
		broken := site.breakTitles
		site.mu.Unlock()
		if broken {
			http.Error(w, `{"error":"gone"}`, http.StatusNotFound)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/__data.json") {
			_, _ = w.Write([]byte(showPayloadJSON()))
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /api/tmdb-season", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") == "" || r.URL.Query().Get("s") == "" {
			http.Error(w, `{"error":"missing id/s"}`, http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(recordedSeason))
	})
	site.srv = httptest.NewServer(mux)
	t.Cleanup(site.srv.Close)
	return site
}

func (s *fakeCatalogSite) newClient(t *testing.T) *Client {
	t.Helper()
	return NewClientForTest(s.srv.Client(), s.srv.URL)
}

// --- fake TMDB with the pool contract ---

// fakeTMDB is a TMDB-lookalike that speaks /search/multi the real server
// does: keyed (unauthorized key → 401), paged, and localized. It RECORDS
// which keys and pages each search used, for the rotation and meta-tests.
type fakeTMDB struct {
	srv         *httptest.Server
	mu          sync.Mutex
	validQ      string
	results     []map[string]any
	usedKeys    map[string][]int // key → pages used
	quota       map[string]bool  // keys that answer 429
	userKey     string           // a key outside the pool the fake must accept
	forceStatus int              // when set, every search answers with it
}

func newFakeTMDB(t *testing.T, query string, results []map[string]any) *fakeTMDB {
	t.Helper()
	fake := &fakeTMDB{
		validQ:   query,
		results:  results,
		usedKeys: make(map[string][]int),
		quota:    make(map[string]bool),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /3/search/multi", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("api_key")
		page := r.URL.Query().Get("page")
		q := r.URL.Query().Get("query")
		if key == "" || q == "" || page == "" {
			http.Error(w, `{"status_message":"missing parameters"}`, http.StatusBadRequest)
			return
		}
		if fake.forceStatus != 0 {
			http.Error(w, `{"forced":"yes"}`, fake.forceStatus)
			return
		}
		// Only the recorded pool keys (and a user key this server knows) are
		// welcome: anything else is a 401, which is what a real quota-dead
		// or typo key meets.
		known := key == fake.userKey
		for _, k := range publicTMDBKeys {
			if k == key {
				known = true
			}
		}
		if !known {
			http.Error(w, `{"success":false,"status_code":7}`, http.StatusUnauthorized)
			return
		}
		// Record the attempt before anything else: the rotation tests read
		// every key that was tried, dead or alive.
		fake.mu.Lock()
		fake.usedKeys[key] = append(fake.usedKeys[key], atoi(page))
		fake.mu.Unlock()
		if fake.quota[key] {
			http.Error(w, `{"success":false,"status_code":29}`, http.StatusTooManyRequests)
			return
		}
		if q != fake.validQ {
			_, _ = w.Write([]byte(`{"page":1,"results":[]}`))
			return
		}
		if atoi(page) > 1 {
			// every search ends with a short page: page 2 is empty here.
			_, _ = w.Write([]byte(`{"page":2,"results":[]}`))
			return
		}
		_, _ = fmt.Fprintf(w, `{"page":1,"results":%s}`, mustJSON(t, fake.results))
	})
	fake.srv = httptest.NewServer(mux)
	t.Cleanup(fake.srv.Close)
	return fake
}

// --- the seal helpers (production framing, sixteen-byte parts) ---

// sealFixture seals plaintext through the production helper itself, so the
// fixture answers EXACTLY what a live byse CDN would — key choice, GCM
// framing and the version-9 parts.
func sealFixture(t *testing.T, plaintext []byte) playbackSeal {
	t.Helper()
	return sealForTestFn(plaintext)
}

func recordJSON(t *testing.T) []byte {
	t.Helper()
	return mustJSON(t, sealedRecord{
		Sources: []sealedSource{
			{URL: "https://edge.test/hls2/low/master.m3u8", Label: "480p", MimeType: "application/vnd.apple.mpegurl", Height: 480},
			{URL: "https://edge.test/hls2/high/master.m3u8", Label: "1080p", MimeType: "application/vnd.apple.mpegurl", Height: 1080},
		},
		Tracks: []byseTrack{{URL: "https://edge.test/pt-br.vtt", Label: "Português", Language: "pt-br", Kind: "subtitles"}},
	})
}

// --- shared small helpers the test files use ---

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// withFakePool swaps the text-search pool for keys and restores it on test
// end. Serial: the pool tests are not parallel.
func withFakePool(t *testing.T, keys []string) func() {
	t.Helper()
	prior := publicTMDBKeys
	publicTMDBKeys = keys
	return func() { publicTMDBKeys = prior }
}

// withUserClient points newTMDBClientFn at a TMDB client for the given fake
// and key; the cleanup restores the real factory, so leaked tests never
// carry a dead fake into the next one.
func withUserClient(t *testing.T, fake *fakeTMDB, key string) func() {
	t.Helper()
	client := movie.NewTMDBClientForTest(fake.srv.Client(), fake.srv.URL+"/3", key)
	newTMDBClientFn = func() *movie.TMDBClient { return client }
	return func() { newTMDBClientFn = userTMDBClientFn }
}
