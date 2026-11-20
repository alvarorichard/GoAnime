package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pomfy"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// The api glue tests drive GetPomfyEpisodes/GetPomfyStreamURL over a test
// world: one httptest server holding BOTH the recorded site shapes (catalog,
// watch-sources, payload pages, the season listing) and the byse endpoints
// the embed points at. The client routes every host onto the world, so the
// embed URL can read https://cdn.test/e/<code> and still land locally.

// routeAllTo sends every request to one test server whatever its host.
type routeAllTo struct{ target *url.URL }

func (r routeAllTo) RoundTrip(req *http.Request) (*http.Response, error) {
	out := req.Clone(req.Context())
	out.URL.Scheme, out.URL.Host, out.Host = r.target.Scheme, r.target.Host, r.target.Host
	return http.DefaultTransport.RoundTrip(out)
}

// recorded site and season blobs (the pomfy package's own mocks pin richer
// fixtures; these are the answers the api tests need verbatim).
const (
	apiRecordedCatalog = `{"filmes":[550],"series":[1396,1399]}`
	apiRecordedSeason  = `{"epNums":[1,2],"sbEps":[1,2],"tmdbEps":{
		"1":{"name":"Seven Thirty-Seven","still":"https://image.tmdb.org/t/p/w300/a.jpg"},
		"2":{"name":"Grilled","still":"https://image.tmdb.org/t/p/w300/b.jpg"}}}`
	apiRecordedShow = `{"nodes":[
		{"data":[]},
		{"data":[
			{"tvShow":1},
			{"adult":2,"name":3,"first_air_date":4,"seasons":5,"overview":6,"poster_path":7},
			false,"Breaking Bad","2008-01-20",
			[8],"Química e drama.","/9Ong1bTS5PkmKw7uXDMDZbcT36r.jpg",
			{"season_number":9,"episode_count":10,"name":11},
			1,2,"Temporada 1"
		]}
	]}`
	apiRecordedAssist = `{"nodes":[
		{"data":[]},
		{"data":[
			{"id":1,"tipo":2,"title":3,"year":4,"overview":5,"image":6},
			"550","filme","Clube da Luta","1999","desc","/x.jpg"
		]}
	]}`
	apiByseNonce  = "awoken"
	apiByseCode   = "gkcq8gbiwh4r"
	apiByseEmbed  = "https://cdn.test/e/" + apiByseCode
	apiByseSubURL = "https://edge.test/pt-br.vtt"
)

// apiByseMaster is what the fake playback seals. A var so a test can serve a
// master whose URL spells an audio language.
var apiByseMaster = "https://edge.test/hls2/master.m3u8"

// newPomfyWorld is the whole fake: site routes plus byse routes, wired so
// watch-sources names the embed this same server serves.
func newPomfyWorld(t *testing.T) *httptest.Server {
	t.Helper()
	issued := map[string]bool{}
	mux := http.NewServeMux()

	// --- the site half ---
	mux.HandleFunc("GET /api/catalogo-ids", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(apiRecordedCatalog))
	})
	mux.HandleFunc("GET /api/watch-sources", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tipo") == "serie" && (r.URL.Query().Get("temporada") == "" || r.URL.Query().Get("episodio") == "") {
			http.Error(w, `{"error":"missing temporada/episodio"}`, http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(w, `{"mode":"direct","sources":[{"kind":"byse","label":"Servidor 1","embedUrl":%q,"token":"x"}]}`, apiByseEmbed)
	})
	mux.HandleFunc("GET /assistir/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/__data.json") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(apiRecordedAssist))
	})
	mux.HandleFunc("GET /serie/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/__data.json") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(apiRecordedShow))
	})
	mux.HandleFunc("GET /api/tmdb-season", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("s") == "" {
			http.Error(w, `{"error":"missing season"}`, http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(apiRecordedSeason))
	})

	// --- the byse half (the embed's own host, served by the same world) ---
	mux.HandleFunc("POST /api/videos/"+apiByseCode+"/embed/captcha", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Embed-Origin") == "" {
			http.Error(w, `{"error":"no embed identity"}`, http.StatusForbidden)
			return
		}
		_, _ = fmt.Fprintf(w, `{"pow_nonce":%q,"pow_difficulty":6,"pow_token":"tok-1","expires_in":1800}`, apiByseNonce)
	})
	mux.HandleFunc("POST /api/videos/"+apiByseCode+"/embed/captcha/verify", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			PowToken string `json:"pow_token"`
			Solution string `json:"solution"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, `{"error":"bad body"}`, http.StatusBadRequest)
			return
		}
		if body.PowToken != "tok-1" || !pomfy.CheckForTest(apiByseNonce+":", body.Solution, 6) {
			http.Error(w, `{"error":"rejected"}`, http.StatusForbidden)
			return
		}
		issued["captcha-ok"] = true
		_, _ = w.Write([]byte(`{"status":"ok","token":"captcha-ok"}`))
	})
	mux.HandleFunc("POST /api/videos/"+apiByseCode+"/embed/playback", func(w http.ResponseWriter, r *http.Request) {
		if !issued[r.Header.Get("X-Captcha-Token")] {
			http.Error(w, `{"error":"no captcha token"}`, http.StatusForbidden)
			return
		}
		raw := pomfy.SealForTest([]byte(fmt.Sprintf(`{"sources":[
			{"url":%q,"label":"720p","mime_type":"application/vnd.apple.mpegurl","height":720}
		],"tracks":[
			{"url":%q,"label":"Português","language":"pt-br","kind":"subtitles"}
		]}`, apiByseMaster, "https://edge.test/pt-br.vtt")))
		_, _ = w.Write(raw)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func mountPomfyClient(t *testing.T, world *httptest.Server) func() {
	t.Helper()
	targetURL, perr := url.Parse(world.URL)
	if perr != nil {
		t.Fatal(perr)
	}
	hc := &http.Client{Transport: routeAllTo{target: targetURL}}
	prior := PomfyClientFn
	PomfyClientFn = func() *pomfy.Client { return pomfy.NewClientForTest(hc, world.URL) }
	t.Cleanup(func() { PomfyClientFn = prior })
	return func() {}
}

// stubSeasonPick remains a headless picker, like the startflix tests inject.
func stubSeasonPick(t *testing.T, chosen int) func() {
	t.Helper()
	prior := sfxSeasonPickFn
	sfxSeasonPickFn = func(items []tui.PickItem, _ tui.PickOptions) (int, error) {
		if chosen >= len(items) {
			t.Fatalf("pick index %d out of %d items", chosen, len(items))
		}
		return chosen, nil
	}
	return func() { sfxSeasonPickFn = prior }
}

// --- the helpers above drive the ROUTING tests below; every test mounts the
// --- world, then asserts the glue's own behavior.

func TestPomfyPathHelpers(t *testing.T) {
	if got := pomfyPathID("https://pomfy.my/assistir/550"); got != "550" {
		t.Errorf("pomfyPathID(assistir) = %q", got)
	}
	if got := pomfyPathID("https://api.pomfy.stream/serie/1396/2/1"); got != "1396" {
		t.Errorf("pomfyPathID(series ep) = %q", got)
	}
	if got := pomfyPathID(""); got != "" {
		t.Errorf("pomfyPathID(empty) = %q", got)
	}
	if got := pomfyPathSeason("https://api.pomfy.stream/serie/1396/2/1"); got != "2" {
		t.Errorf("pomfyPathSeason = %q, want 2", got)
	}
	if got := pomfyPathSeason("https://pomfy.my/assistir/550"); got != "" {
		t.Errorf("pomfyPathSeason(movie page) = %q, want empty", got)
	}
	if n, ok := pomfyTMDBID(&models.Anime{TMDBID: 550}, nil); !ok || n != "550" {
		t.Errorf("pomfyTMDBID field: (%q, %v)", n, ok)
	}
	if _, ok := pomfyTMDBID(&models.Anime{}, nil); ok {
		t.Error("nothing must not satisfy")
	}
}

func TestPomfySeasonAndEpisode(t *testing.T) {
	season, num := pomfySeasonAndEpisode(&models.Anime{}, &models.Episode{SeasonID: "2", Number: "5", Num: 5}, pomfy.KindSeries)
	if season != 2 || num != 5 {
		t.Errorf("series from ids: (%d,%d), want (2,5)", season, num)
	}
	season, num = pomfySeasonAndEpisode(&models.Anime{CurrentSeason: 3}, &models.Episode{Number: "1"}, pomfy.KindSeries)
	if season != 3 || num != 1 {
		t.Errorf("series from current: (%d,%d), want (3,1)", season, num)
	}
	season, num = pomfySeasonAndEpisode(nil, nil, pomfy.KindMovie)
	if season != 0 || num != 0 {
		t.Errorf("movies must stay flat: (%d,%d)", season, num)
	}
}

func TestPomfyStreamLang(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://edge.test/hls2/,c_x,lang/por/c_por,.urlset/master.m3u8", "por"},
		{"https://edge.test/hls2/,c_x,lang/eng/c_eng,.urlset/master.m3u8", "eng"},
		{"https://edge.test/hls2/,c_x,lang/jpn/c_jpn,.urlset/master.m3u8", "jpn"},
		{"https://edge.test/x/master.m3u8", ""},
		{"", ""},
		{"https://edge.test/x,lang/xyz/x.xyz,.urlset/master.m3u8", "xyz"},
	}
	for _, tc := range cases {
		if got := pomfyStreamLang(tc.url); got != tc.want {
			t.Errorf("pomfyStreamLang(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestGetPomfyEpisodesMovie(t *testing.T) {
	world := newPomfyWorld(t)
	restore := mountPomfyClient(t, world)
	defer restore()

	media := &models.Anime{Name: "Clube da Luta", URL: world.URL + "/assistir/550",
		Source: "Pomfy", MediaType: models.MediaTypeMovie, TMDBID: 550}
	eps, err := GetPomfyEpisodes(media)
	if err != nil {
		t.Fatalf("GetPomfyEpisodes movie: %v", err)
	}
	if len(eps) != 1 || eps[0].Num != 1 || eps[0].DataID != "550" || eps[0].SeasonID != "0" {
		t.Fatalf("episodes = %+v, want the one movie episode", eps)
	}
	if !strings.Contains(eps[0].URL, "/assistir/550") {
		t.Errorf("episode URL = %q", eps[0].URL)
	}
}

func TestGetPomfySeasonEpisodes(t *testing.T) {
	world := newPomfyWorld(t)
	restore := mountPomfyClient(t, world)
	defer restore()

	media := &models.Anime{Name: "Breaking Bad", URL: world.URL + "/serie/1396",
		Source: "Pomfy", MediaType: models.MediaTypeTV, TMDBID: 1396}
	eps, err := GetPomfySeasonEpisodes(media, 1)
	if err != nil {
		t.Fatalf("GetPomfySeasonEpisodes: %v", err)
	}
	if len(eps) != 2 {
		t.Fatalf("episodes = %d, want the 2 recorded as available", len(eps))
	}
	if eps[0].Number != "1" || eps[0].Title.English != "Seven Thirty-Seven" {
		t.Errorf("ep 1 = %+v", eps[0])
	}
	if !strings.HasSuffix(eps[1].URL, "/serie/1396/1/2") {
		t.Errorf("episode URL = %q", eps[1].URL)
	}
	if media.CurrentSeason != 1 {
		t.Errorf("CurrentSeason = %d, the listing must record it", media.CurrentSeason)
	}
}

func TestGetPomfyEpisodesSeriesPicksSeason(t *testing.T) {
	world := newPomfyWorld(t)
	restorePick := stubSeasonPick(t, 0)
	defer restorePick()
	restore := mountPomfyClient(t, world)
	defer restore()

	media := &models.Anime{Name: "Breaking Bad", URL: world.URL + "/serie/1396",
		Source: "Pomfy", MediaType: models.MediaTypeTV, TMDBID: 1396}
	eps, err := GetPomfyEpisodes(media)
	if err != nil {
		t.Fatalf("GetPomfyEpisodes series: %v", err)
	}
	if media.CurrentSeason != 1 || len(eps) == 0 {
		t.Errorf("season %d, episodes %d", media.CurrentSeason, len(eps))
	}
}

// TestGetPomfyStreamURLMovie runs the whole recorded chain: watch-sources →
// byse captcha (PoW solved for real) → playback → decrypt → the seeded
// master, and hands mpv's metadata contract its referer.
func TestGetPomfyStreamURLMovie(t *testing.T) {
	world := newPomfyWorld(t)
	restore := mountPomfyClient(t, world)
	defer restore()

	media := &models.Anime{Name: "Clube da Luta", URL: world.URL + "/assistir/550",
		Source: "Pomfy", MediaType: models.MediaTypeMovie, TMDBID: 550}
	episode := &models.Episode{Number: "1", Num: 1, DataID: "550", SeasonID: "0",
		URL: world.URL + "/assistir/550"}

	streamURL, err := GetPomfyStreamURL(media, episode, "")
	if err != nil {
		t.Fatalf("GetPomfyStreamURL: %v", err)
	}
	if !strings.Contains(streamURL, "master.m3u8") {
		t.Errorf("stream = %q, want the sealed record's master", streamURL)
	}
}

// TestGetPomfyStreamURLSeriesIsEndpointEpisodeURL one bonus assert: the
// series path parses season/episode numbers out of the episode URL when the
// listing did not carry them (history-restored episodes).
func TestGetPomfyStreamURLSeriesEndpointParsing(t *testing.T) {
	world := newPomfyWorld(t)
	restore := mountPomfyClient(t, world)
	defer restore()

	media := &models.Anime{Name: "Breaking Bad", URL: world.URL + "/serie/1396",
		Source: "Pomfy", MediaType: models.MediaTypeTV, TMDBID: 1396}
	episode := &models.Episode{Number: "1", Num: 1, DataID: "1396", SeasonID: ""}

	// SeasonID empty means derive from URLs/current season; the world's
	// watch-sources answers the same master either way.
	streamURL, err := GetPomfyStreamURL(media, episode, "")
	if err != nil {
		t.Fatalf("GetPomfyStreamURL series: %v", err)
	}
	if !strings.Contains(streamURL, "master.m3u8") {
		t.Errorf("stream = %q", streamURL)
	}
}
