package miruro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Fixtures are written as plain Go values and encoded to devalue by
// encodeDevalue below, which follows the format rather than calling the
// decoder: a change to the decoder has to be matched here on purpose instead
// of cancelling itself out. rawWatchBody is the counterpart, written by hand
// in the exact shape miruro.cx sent on 2026-11-26.

// undefinedT marks a devalue undefined in a fixture.
type undefinedT struct{}

var undefined = undefinedT{}

// promise marks a ["Promise", id] in a fixture; a chunk line settles it.
type promise int

// encodeDevalue flattens a value the way devalue does: every value once, in
// a flat array, referred to by index, the root at 0.
func encodeDevalue(root any) []any {
	var flat []any
	var add func(v any) int
	add = func(v any) int {
		if _, ok := v.(undefinedT); ok {
			return devalueUndefined
		}
		idx := len(flat)
		flat = append(flat, nil)
		switch x := v.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			m := make(map[string]any, len(x))
			for _, k := range keys {
				m[k] = float64(add(x[k]))
			}
			flat[idx] = m
		case []any:
			a := make([]any, 0, len(x))
			for _, c := range x {
				a = append(a, float64(add(c)))
			}
			flat[idx] = a
		case promise:
			flat[idx] = []any{"Promise", float64(add(float64(x)))}
		case int:
			flat[idx] = float64(x)
		default: // string, float64, bool, nil
			flat[idx] = x
		}
		return idx
	}
	add(root)
	return flat
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// layoutNode stands in for the root layout's data, which every page carries
// first and this scraper never reads.
var layoutNode = map[string]any{
	"branding":  map[string]any{"mark": "/icon.png"},
	"providers": []any{map[string]any{"name": "sun", "sub": false}},
}

// dataBody is a non-streamed __data.json answer.
func dataBody(nodes ...any) string {
	wire := make([]any, 0, len(nodes))
	for _, n := range nodes {
		wire = append(wire, map[string]any{"type": "data", "data": encodeDevalue(n)})
	}
	return mustJSON(map[string]any{"type": "data", "nodes": wire}) + "\n"
}

// streamedBody is a __data.json answer whose promises later lines settle,
// written in descending id order as the site does (it sends whichever
// settles first, not the order the page refers to them).
func streamedBody(nodes []any, chunks map[int]any) string {
	var b strings.Builder
	b.WriteString(strings.TrimSuffix(dataBody(nodes...), "\n"))
	ids := make([]int, 0, len(chunks))
	for id := range chunks {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	for _, id := range ids {
		b.WriteString("\n")
		b.WriteString(mustJSON(map[string]any{"type": "chunk", "id": id, "data": encodeDevalue(chunks[id])}))
	}
	return b.String() + "\n"
}

const (
	bebopID   = "C0wb0yBebop_session-0123456789ab"
	bebopSlug = "cowboy-bebop"
	unairedID = "N0tYetAired_session-0123456789ab"
)

// fixture is one scripted miruro.cx plus its stream CDNs.
type fixture struct {
	items    []any          // the search page's result list
	episodes []any          // the watch page's episode list
	sources  any            // the watch page's sources; nil sends null
	order    []any          // providerOrder
	cdn      map[string]cdn // /cdn/<name>/... by name
}

// cdn is one stream host: what it serves and the headers it insists on.
type cdn struct {
	playlist string
	referer  string // required Referer, "" for none
	origin   string // required Origin, "" for none
	refuse   bool   // answer 403 whatever is sent
}

const masterPlaylist = "#EXTM3U\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360\nindex-360.m3u8\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720\nindex-720.m3u8\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080\nhttps://elsewhere.example/1080.m3u8\n"

const mediaPlaylist = "#EXTM3U\n#EXT-X-TARGETDURATION:10\n#EXTINF:10,\nseg0.ts\n#EXT-X-ENDLIST\n"

func resultItem(id, english, romaji, status string, year, sub, dub any, anilist, mal string) map[string]any {
	ext := map[string]any{"kitsu": []any{"1"}}
	if anilist != "" {
		ext["anilist"] = []any{anilist}
	}
	if mal != "" {
		ext["mal"] = []any{mal}
	}
	var eng any = english
	if english == "" {
		eng = nil
	}
	return map[string]any{
		"id":             id,
		"title":          map[string]any{"english": eng, "romaji": romaji, "native": "カウボーイビバップ"},
		"format":         "TV",
		"status":         status,
		"season_year":    year,
		"episode_counts": map[string]any{"raw": 26, "sub": sub, "dub": dub},
		"external_ids":   ext,
		"cover_url":      "https://img.example/" + id + ".jpg",
		"description":    "A bounty hunter crew.",
	}
}

func episode(n any, title string, filler bool) map[string]any {
	return map[string]any{
		"id": fmt.Sprint(n), "number": n, "title": title, "airDate": "1998-04-03",
		"description": "Synopsis of " + title, "filler": filler, "image": "https://img.example/e.jpg",
	}
}

func subtitle(lang, label, file string, def bool) map[string]any {
	return map[string]any{"language": lang, "label": label, "file": file, "format": "vtt", "default": def}
}

func server(name string, headers map[string]any, url string) map[string]any {
	var h any
	if headers != nil {
		h = headers
	}
	return map[string]any{
		"server": name, "headers": h, "embed": nil,
		"streams": []any{map[string]any{"url": url, "format": "hls", "quality": nil}},
	}
}

func embedOnly(name, url string) map[string]any {
	return map[string]any{"server": name, "headers": nil, "embed": map[string]any{"url": url}, "streams": []any{}}
}

func provider(name string, subs []any, servers ...any) map[string]any {
	if subs == nil {
		subs = []any{}
	}
	return map[string]any{
		"provider": name, "servers": servers, "subtitles": subs,
		"skipTimes": nil, "downloads": nil, "thumbnails": nil,
	}
}

func track(name string, providers ...any) map[string]any {
	return map[string]any{"track": name, "providers": providers}
}

// pristineFixture is the unmutated site, shared by client_test.go and
// meta_test.go. base is the stub's own URL, which the stream URLs point at.
//
// It mirrors what the site served on 2026-11-26 for a typical title: the
// soft-subbed track listed first with "sun" (Referer and Origin required) and
// "bee" (Referer only, Arabic flagged as its default subtitle), a burned-in
// "sub" track on "knob" (no headers), a dubbed track on "bee", and an
// embed-only server that must be skipped.
func pristineFixture(base string) *fixture {
	return &fixture{
		items: []any{
			resultItem(bebopID, "Cowboy Bebop", "Cowboy Bebop", "FINISHED", 1998, 26, 26, "1", "1"),
			resultItem("R0majiOnly_session-0123456789ab", "", "Kaubōi Bibappu: Tengoku no Tobira", "FINISHED", 2001, 1, nil, "5", "5"),
			resultItem(unairedID, "Cowboy Bebop 2", "Cowboy Bebop 2", "NOT_YET_RELEASED", nil, nil, nil, "999", ""),
		},
		episodes: []any{
			episode(2, "Stray Dog Strut", false),
			episode(1, "Asteroid Blues", false),
			episode(1.5, "Recap", true),
		},
		order: []any{"sun", "bee", "hop", "knob"},
		sources: map[string]any{
			"episodeNumber": 1,
			"tracks": []any{
				track(trackSoftSub,
					provider("bee",
						[]any{
							subtitle("ar", "Arabic", base+"/subs/ar.vtt", true),
							subtitle("en", "English", base+"/subs/bee-en.vtt", false),
							subtitle("pt-BR", "Portuguese (Brazil)", base+"/subs/pt.vtt", false),
						},
						server("HD-1", map[string]any{"Referer": base + "/mega/"}, base+"/cdn/bee/master.m3u8"),
						embedOnly("Vidstream-2", base+"/embed/bee"),
					),
					provider("sun",
						[]any{subtitle("en", "English (Full Subtitles / ASS)", base+"/subs/sun-en.jpeg", false)},
						server("icarus-3-1", map[string]any{"Origin": base + "/ultra", "Referer": base + "/ultra/"}, base+"/cdn/sun/master.m3u8"),
					),
				),
				track(trackHardSub,
					provider("knob", nil, server("Vidplay", nil, base+"/cdn/knob/index.m3u8")),
				),
				track(trackDub,
					provider("bee", []any{subtitle("en", "English", base+"/subs/dub-en.vtt", true)},
						server("HD-1", map[string]any{"Referer": base + "/mega/"}, base+"/cdn/beedub/master.m3u8"),
					),
				),
			},
		},
		cdn: map[string]cdn{
			"sun":    {playlist: masterPlaylist, referer: base + "/ultra/", origin: base + "/ultra"},
			"bee":    {playlist: masterPlaylist, referer: base + "/mega/"},
			"beedub": {playlist: masterPlaylist, referer: base + "/mega/"},
			"knob":   {playlist: mediaPlaylist},
		},
	}
}

// stub serves one fixture and records the paths it was asked for.
type stub struct {
	srv *httptest.Server
	fx  *fixture

	mu   sync.Mutex
	hits []string
	// Fault injection: when non-zero, the matching page answers with it.
	failSearch, failWatch int
}

// newStub serves the pristine fixture after edit has had its say.
func newStub(t *testing.T, edit func(base string, fx *fixture)) *stub {
	t.Helper()
	s := &stub{}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	s.fx = pristineFixture(s.srv.URL)
	if edit != nil {
		edit(s.srv.URL, s.fx)
	}
	return s
}

func (s *stub) paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.hits...)
}

func (s *stub) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.hits = append(s.hits, r.URL.Path)
	s.mu.Unlock()

	switch {
	case r.URL.Path == "/search/__data.json":
		if s.failSearch != 0 {
			w.WriteHeader(s.failSearch)
			return
		}
		if r.URL.Query().Get("q") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, dataBody(layoutNode, map[string]any{"mode": "anime", "items": s.fx.items, "hasMore": true}))

	case strings.HasPrefix(r.URL.Path, "/watch/") && strings.HasSuffix(r.URL.Path, dataSuffix):
		if s.failWatch != 0 {
			w.WriteHeader(s.failWatch)
			return
		}
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/watch/"), dataSuffix), "/")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		id, slug := parts[0], parts[1]
		canonical := bebopSlug
		if id == unairedID {
			canonical = "cowboy-bebop-2"
		}
		ep := r.URL.Query().Get("ep")
		if slug != canonical {
			_, _ = fmt.Fprint(w, mustJSON(map[string]any{
				"type": "redirect", "location": "/watch/" + id + "/" + canonical + "?ep=" + ep,
			}))
			return
		}
		w.Header().Set("Content-Type", "text/sveltekit-data")
		episodes, sources := s.fx.episodes, s.fx.sources
		if id == unairedID {
			episodes, sources = []any{}, nil
		}
		var src any
		if sources != nil {
			m := map[string]any{"providerOrder": s.fx.order}
			for k, v := range sources.(map[string]any) {
				m[k] = v
			}
			src = m
		}
		page := map[string]any{
			"media":          map[string]any{"id": 1, "title": map[string]any{"romaji": "Cowboy Bebop"}},
			"relatedData":    promise(2),
			"episodesData":   promise(4),
			"sourcesData":    promise(5),
			"anilistId":      1,
			"tvdbBackground": undefined,
		}
		_, _ = fmt.Fprint(w, streamedBody([]any{layoutNode, page}, map[int]any{
			2: []any{},
			4: map[string]any{"episodes": episodes, "currentEpisode": nil},
			5: src,
		}))

	case strings.HasPrefix(r.URL.Path, "/cdn/"):
		name := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/cdn/"), "/", 2)[0]
		c, ok := s.fx.cdn[name]
		switch {
		case !ok:
			w.WriteHeader(http.StatusNotFound)
		case c.refuse,
			c.referer != "" && r.Header.Get("Referer") != c.referer,
			c.origin != "" && r.Header.Get("Origin") != c.origin:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, "<html><head><title>403 Forbidden</title></head></html>")
		default:
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = fmt.Fprint(w, c.playlist)
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// rawWatchBody is a watch page response written by hand in the shape the
// site sent on 2026-11-26, trimmed to the fields read here: two nodes, the
// page's promises referring to an index whose value is the chunk id, chunks
// out of order, a hoisted shared object, an undefined and a Date.
const rawWatchBody = `{"type":"data","nodes":[{"type":"data","data":[{"branding":1},{"mark":2},"/icon.png"]},{"type":"data","data":[{"media":1,"episodesData":4,"sourcesData":6,"episodeNumber":3,"backdrop":-1,"updated":8},{"id":2},154587,1,["Promise",5],2,["Promise",7],5,["Date","2026-11-26T00:00:00.000Z"]]}]}
{"type":"chunk","id":5,"data":[{"episodeNumber":1,"tracks":2,"providerOrder":20},1,[3],{"track":4,"providers":5},"sub",[6],{"provider":7,"servers":8,"subtitles":9},"sun",[10],[],{"server":11,"headers":12,"embed":13,"streams":14},"icarus-2-1",{"Origin":15,"Referer":16},null,[17],"https://ultracloud.example","https://ultracloud.example/",{"url":18,"format":19,"quality":13},"https://cdn.example/m/abc.m3u8","hls",[21],"sun"]}
{"type":"chunk","id":2,"data":[{"episodes":1,"currentEpisode":2},[2],{"id":3,"number":4,"title":5,"filler":6},"1",1,"Asteroid Blues",false]}
`

func init() {
	// A typo in the hand-written JSON above should fail every test loudly
	// rather than one test obscurely.
	for i, line := range strings.Split(strings.TrimSpace(rawWatchBody), "\n") {
		var v any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			panic(fmt.Sprintf("rawWatchBody line %d is not JSON: %v", i, err))
		}
	}
}
