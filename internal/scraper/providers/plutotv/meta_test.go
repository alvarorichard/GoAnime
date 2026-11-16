package plutotv

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the tests.
//
// The fixtures are what the tests are fitted to, so a test that still passes
// when a fixture is damaged is not looking at that part of it. The table below
// breaks one thing in the served data at a time and requires the chain's
// outcome to change; a few mutations are controls that must change nothing,
// which shows the table can tell a break from noise.
//
// The code's own mutants — one decision broken in client.go, the tests that
// must fail on it — live in internal/api/plutotv_meta_test.go, where the
// runner shared by the other sources is.

// outcome is everything the chain tells a caller, flattened so two runs can be
// compared.
type outcome struct {
	titles  string
	seasons string
	stream  string
	adFree  bool
	err     string
	// playlist is the text handed to the player, the fake's address taken out
	// so that two runs (two fakes) compare.
	playlist string
}

// realClip is where the real fixture's title lives, as a path on the fake CDN.
const realClip = "/538_Viz_Media/clip/6206bba415c86100136839a7_Death_Note_BR_Renascimento_S1E1/720p/20220214_054014/hls/"

func realContentPath(rendition string) string { return realClip + "0-end/" + rendition + ".m3u8" }

// chain runs search → seasons → stream against a fake serving the fixtures,
// after `mutate` has had its say.
func chain(t *testing.T, mutate func(*fakePluto, map[string]string)) outcome {
	t.Helper()
	f := newFakePluto(t)
	files := map[string]string{
		"search":   fixture(t, "search_2026_11_16.json"),
		"seasons":  fixture(t, "seasons_2026_11_16.json"),
		"master":   fixture(t, "master_2026_11_16.m3u8"),
		"stitched": fixture(t, "stitched_series_2026_11_16.m3u8"),
	}
	if mutate != nil {
		mutate(f, files)
	}

	f.set("/v1/search", files["search"])
	f.serveShow(files["seasons"])
	f.set(f.masterPath(fakeEpisodeID), files["master"])
	for _, rung := range fakeLadder {
		// The fixture is the service's own bytes, media host included: the
		// fake is served at its own address, so the host is pointed at it. The
		// fixture is the lowest rung's; the others are the same playlist under
		// their own rendition's name.
		body := strings.ReplaceAll(files["stitched"], "hls_300", rung.rendition)
		body = strings.ReplaceAll(body, "https://siloh-ns1.plutotv.net", f.url())
		f.set(f.variantPath(fakeEpisodeID, rung.bandwidth), body)
		f.mu.Lock()
		_, mutated := f.routes[realContentPath(rung.rendition)]
		f.mu.Unlock()
		if !mutated {
			f.set(realContentPath(rung.rendition), content(rung.rendition, 3))
		}
	}

	var o outcome
	c := f.client()
	titles, err := c.Search(t.Context(), "naruto")
	if err != nil {
		o.err = "search: " + err.Error()
		return o
	}
	var names []string
	for _, title := range titles {
		names = append(names, string(title.Kind)+":"+title.ID+":"+title.Name)
	}
	o.titles = strings.Join(names, "|")

	show, err := c.Show(t.Context(), SeriesURL(fakeSeriesID))
	if err != nil {
		o.err = "show: " + err.Error()
		return o
	}
	var eps []string
	for _, s := range show.Seasons {
		for _, e := range s.Episodes {
			eps = append(eps, fmt.Sprintf("S%dE%d:%s:%s:%s", s.Number, e.Number, e.ID, e.Name, e.Duration))
		}
	}
	o.seasons = strings.Join(eps, "|")

	stream, err := c.Stream(t.Context(), EpisodeURL(fakeSeriesID, fakeEpisodeID), "worst")
	if err != nil {
		o.err = "stream: " + err.Error()
		return o
	}
	o.stream = strings.TrimPrefix(stream.URL, f.url())
	if i := strings.Index(o.stream, "?"); i >= 0 {
		o.stream = o.stream[:i] // the stitcher's query carries the session of this run
	}
	o.adFree = stream.AdFree
	o.playlist = strings.ReplaceAll(stream.Playlist, f.url(), "")
	return o
}

func TestTheChainOverTheFixturesIsWhatTheOtherTestsAssume(t *testing.T) {
	o := chain(t, nil)
	require.Empty(t, o.err)
	assert.True(t, o.adFree)
	assert.Equal(t, realContentPath("hls_300"), o.stream, "the lowest rung, the title alone")
	eps := strings.Split(o.seasons, "|")
	require.Len(t, eps, 3)
	assert.Equal(t, "S1E1:"+fakeEpisodeID+":Renascimento:23m15s", eps[0])
	assert.Regexp(t, `^S1E2:[0-9a-f]{24}:Confronto:23m30s$`, eps[1])
	assert.Regexp(t, `^S1E3:`, eps[2])
	assert.Equal(t, 14, strings.Count(o.titles, "|")+1)
}

// tester is what replaceOnce needs of a test, so a test of the helper can hand
// it a recorder.
type tester interface {
	require.TestingT
	Helper()
}

// replaceOnce edits one snippet of a fixture, and fails the mutation itself if
// the snippet is not there exactly once: a mutation that changes nothing would
// "survive" for the wrong reason.
func replaceOnce(t tester, files map[string]string, which, from, to string) {
	t.Helper()
	require.Equal(t, 1, strings.Count(files[which], from), "the %s fixture no longer holds the snippet this mutation breaks:\n%s", which, from)
	require.NotEqual(t, from, to, "a mutation that replaces a snippet with itself changes nothing")
	files[which] = strings.Replace(files[which], from, to, 1)
}

type fixtureMutation struct {
	name string
	// mustChange is whether the chain's outcome has to differ from the
	// baseline. A control (false) shows an irrelevant edit goes unnoticed.
	mustChange bool
	layer      string
	mutate     func(t *testing.T, f *fakePluto, files map[string]string)
}

var fixtureMutations = []fixtureMutation{
	// ---- search
	{"a series becomes a channel", true, "search", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "search", `"id": "69dd4f068642675fcc3d899c",`, `"id": "69dd4f068642675fcc3d899c", "x": 1,`)
		files["search"] = strings.Replace(files["search"], `"type": "series"`, `"type": "channel"`, 1)
	}},
	{"a title loses its name", true, "search", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "search", `"name": "Naruto: Exame Chuunin"`, `"name": ""`)
	}},
	{"a title's id is cut short", true, "search", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "search", `"id": "69dd4f068642675fcc3d899c"`, `"id": "69dd4f06864267"`)
	}},
	{"the answer has no data", true, "search", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["search"] = `{"suggestions": []}`
	}},
	{"a channel is renamed (control)", false, "search", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "search", `"name": "Pluto TV Anime Ação"`, `"name": "Pluto TV Anime Action"`)
	}},
	{"a title's language changes (control)", false, "search", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["search"] = strings.Replace(files["search"], `"language": "PT"`, `"language": "EN"`, 1)
	}},

	// ---- seasons
	{"the series id is another's", true, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"_id": "6206bba315c8610013683974"`, `"_id": "6206bba315c8610013683975"`)
	}},
	{"an episode loses its duration", true, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"originalContentDuration": 1395000`, `"originalContentDuration": 0`)
	}},
	{"the episode list is empty", true, "seasons", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["seasons"] = `{"_id": "6206bba315c8610013683974", "name": "Death Note", "seasons": [{"number": 1, "episodes": []}]}`
	}},
	{"an episode is renamed", true, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"name": "Renascimento"`, `"name": "Rebirth"`)
	}},
	{"an episode's number moves", true, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"number": 2,`, `"number": 7,`)
	}},
	{"an episode's stitched path changes (control)", false, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"/stitch/hls/episode/6206bbaa15c8610013683a28/master.m3u8"`, `"/stitch/hls/episode/elsewhere/master.m3u8"`)
	}},
	{"an episode's slug changes (control)", false, "seasons", func(t *testing.T, _ *fakePluto, files map[string]string) {
		replaceOnce(t, files, "seasons", `"slug": "renascimento-2007-1-1"`, `"slug": "renascimento-2008-1-1"`)
	}},

	// ---- master
	{"the master lists no rung", true, "master", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["master"] = "#EXTM3U\n#EXT-X-VERSION:5\n"
	}},
	{"the lowest rung is gone", true, "master", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		lines := strings.Split(files["master"], "\n")
		var kept []string
		for i := 0; i < len(lines); i++ {
			if strings.Contains(lines[i], "BANDWIDTH=640930") {
				i++ // its URI line
				continue
			}
			kept = append(kept, lines[i])
		}
		files["master"] = strings.Join(kept, "\n")
	}},
	{"the audio rendition is gone (control)", false, "master", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		var kept []string
		for _, l := range strings.Split(files["master"], "\n") {
			if !strings.HasPrefix(l, "#EXT-X-MEDIA:") {
				kept = append(kept, l)
			}
		}
		files["master"] = strings.Join(kept, "\n")
	}},

	// ---- the stitched playlist
	{"the title is cut off before its end", true, "stitched", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.ReplaceAll(files["stitched"], "/hls/1295297-end/", "/hls/1295297-1295300/")
	}},
	{"the title no longer starts at its start", true, "stitched", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.ReplaceAll(files["stitched"], "/hls/0-498332/", "/hls/10-498332/")
	}},
	{"the title's segments are keyed by a key system", true, "stitched", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.Replace(files["stitched"], "METHOD=AES-128", "METHOD=SAMPLE-AES", 1)
	}},
	{"every ad is gone (control)", false, "stitched", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		var kept []string
		for _, l := range strings.Split(files["stitched"], "\n") {
			if l != "" && !strings.HasPrefix(l, "#") && strings.Contains(l, "/mp4/") {
				if n := len(kept); n > 0 && strings.HasPrefix(kept[n-1], "#EXTINF") {
					kept = kept[:n-1]
				}
				continue
			}
			kept = append(kept, l)
		}
		files["stitched"] = strings.Join(kept, "\n")
	}},

	// ---- the title's own playlist
	{"the title's own playlist is forbidden", true, "own playlist", func(_ *testing.T, f *fakePluto, _ map[string]string) {
		f.setRoute(realContentPath("hls_300"), http.StatusForbidden, "x")
	}},
	{"the title's own playlist is a page", true, "own playlist", func(_ *testing.T, f *fakePluto, _ map[string]string) {
		f.set(realContentPath("hls_300"), "<html>blocked</html>")
	}},
}

func TestFixtureMutationsAreNoticed(t *testing.T) {
	baseline := chain(t, nil)
	require.Empty(t, baseline.err, "the baseline is the real chain over the real fixtures")

	for _, m := range fixtureMutations {
		t.Run(m.name, func(t *testing.T) {
			got := chain(t, func(f *fakePluto, files map[string]string) { m.mutate(t, f, files) })
			if m.mustChange {
				assert.NotEqual(t, baseline, got, "the chain did not notice: the tests over this part of the fixtures are vacuous")
			} else {
				assert.Equal(t, baseline, got, "a control changed the outcome: the table cannot tell a break from noise")
			}
		})
	}
}

// The table is only worth its name if it reaches every layer and keeps its
// controls.
func TestFixtureMutationsCoverEveryLayerAndKeepTheirControls(t *testing.T) {
	names := map[string]bool{}
	perLayer := map[string]int{}
	controls := map[string]int{}
	for _, m := range fixtureMutations {
		assert.False(t, names[m.name], "duplicate mutation %q", m.name)
		names[m.name] = true
		assert.NotNil(t, m.mutate, m.name)
		perLayer[m.layer]++
		if !m.mustChange {
			controls[m.layer]++
		}
		assert.Equal(t, !m.mustChange, strings.HasSuffix(m.name, "(control)"), "%q: a mutation that must change nothing says so, and only it", m.name)
	}
	for _, layer := range []string{"search", "seasons", "master", "stitched", "own playlist"} {
		assert.GreaterOrEqual(t, perLayer[layer]-controls[layer], 2, "the %s layer has fewer than two real mutations", layer)
	}
	for _, layer := range []string{"search", "seasons", "master", "stitched"} {
		assert.GreaterOrEqual(t, controls[layer], 1, "the %s layer has no control", layer)
	}
	assert.Len(t, perLayer, 5, "a mutation names a layer this check does not know")
}

// recorder is a tester that records a failure instead of failing the run.
type recorder struct {
	failed   bool
	messages []string
}

func (r *recorder) Helper() {}
func (r *recorder) Errorf(format string, args ...interface{}) {
	r.messages = append(r.messages, fmt.Sprintf(format, args...))
}
func (r *recorder) FailNow() { r.failed = true; panic(r) }

// try runs fn and reports whether it failed the recorder.
func (r *recorder) try(fn func()) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = r.failed
		}
	}()
	fn()
	return false
}

// replaceOnce is what keeps a stale mutation from passing for the wrong
// reason; this holds it to that.
func TestReplaceOnceFailsWhenTheMutationWouldChangeNothing(t *testing.T) {
	t.Run("a snippet that occurs once is replaced", func(t *testing.T) {
		files := map[string]string{"x": "aa bb aa"}
		r := &recorder{}
		assert.False(t, r.try(func() { replaceOnce(r, files, "x", "bb", "cc") }))
		assert.Equal(t, "aa cc aa", files["x"])
	})
	t.Run("a snippet that is not there fails", func(t *testing.T) {
		files := map[string]string{"x": "aa bb aa"}
		r := &recorder{}
		assert.True(t, r.try(func() { replaceOnce(r, files, "x", "zz", "cc") }))
		assert.Equal(t, "aa bb aa", files["x"], "nothing was edited")
		require.NotEmpty(t, r.messages)
		assert.Contains(t, r.messages[0], "no longer holds the snippet")
	})
	t.Run("a snippet that occurs twice fails", func(t *testing.T) {
		files := map[string]string{"x": "aa bb aa"}
		r := &recorder{}
		assert.True(t, r.try(func() { replaceOnce(r, files, "x", "aa", "cc") }))
		assert.Equal(t, "aa bb aa", files["x"])
	})
	t.Run("a replacement by itself fails", func(t *testing.T) {
		files := map[string]string{"x": "aa bb aa"}
		r := &recorder{}
		assert.True(t, r.try(func() { replaceOnce(r, files, "x", "bb", "bb") }))
	})
}

// The recorder is itself a piece of test machinery: it must tell a pass from a
// failure, or the test above passes whatever replaceOnce does.
func TestTheRecorderTellsAPassFromAFailure(t *testing.T) {
	r := &recorder{}
	assert.False(t, r.try(func() {}), "a body that does nothing did not fail")
	assert.True(t, r.try(func() { r.Errorf("x"); r.FailNow() }), "a body that fails the recorder did")
	assert.Equal(t, []string{"x"}, r.messages)
	assert.False(t, (&recorder{}).try(func() { panic("a plain panic") }), "a panic that is not a failure of the recorder is not reported as one")
}

// ---- the playlist assembled when the CDN refuses the title's own ------------

// refuseOwnPlaylists makes the fake CDN answer as the refusing one does: 403
// for the title's own playlist at every rung, so the title is read out of the
// stitched playlist.
func refuseOwnPlaylists(f *fakePluto) {
	for _, rung := range fakeLadder {
		f.setRoute(realContentPath(rung.rendition), http.StatusForbidden, "")
	}
}

// assembledMutations break (or, for the controls, only disturb) the real
// stitched playlist while the title's own is refused. What must change is the
// playlist handed to the player; what must not is anything spliced around the
// title, which the assembled playlist never carried.
var assembledMutations = []fixtureMutation{
	{"a segment of the title is dropped", true, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		lines := strings.Split(files["stitched"], "\n")
		for i, l := range lines {
			if strings.Contains(l, "/hls/498333-941317/") && strings.HasSuffix(strings.TrimSpace(l), ".ts") {
				lines = append(lines[:i-2], lines[i+1:]...) // its key, its duration, itself
				break
			}
		}
		files["stitched"] = strings.Join(lines, "\n")
	}},
	{"a segment of the title changes its duration", true, "assembled", func(t *testing.T, _ *fakePluto, files map[string]string) {
		lines := strings.Split(files["stitched"], "\n")
		for i, l := range lines {
			if strings.Contains(l, "/hls/0-498332/") && strings.HasSuffix(strings.TrimSpace(l), ".ts") {
				require.True(t, strings.HasPrefix(lines[i-1], "#EXTINF:5,"), lines[i-1])
				lines[i-1] = "#EXTINF:4,"
				break
			}
		}
		files["stitched"] = strings.Join(lines, "\n")
	}},
	{"an ad changes its duration (control)", false, "assembled", func(t *testing.T, _ *fakePluto, files map[string]string) {
		// 5.133 is the length of an ad's segment; the title's own are 5 and 5.1.
		replaceOnce(t, files, "stitched", "#EXTINF:5.133,", "#EXTINF:4.133,")
	}},
	{"a key file of the title is another", true, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.Replace(files["stitched"], "hls_300_keyfile_0.key", "hls_300_keyfile_9.key", 1)
	}},
	{"a segment's IV is another", true, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.Replace(files["stitched"], "IV=0x00000000000000000000000000000001", "IV=0x000000000000000000000000000000ff", 1)
	}},
	{"a key leaves its IV to the sequence number", true, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		i := strings.Index(files["stitched"], ",IV=0x")
		files["stitched"] = files["stitched"][:i] + files["stitched"][i+len(",IV=0x")+32:]
	}},
	{"two ranges of the title change places", true, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		s := strings.ReplaceAll(files["stitched"], "/hls/498333-941317/", "/hls/@@/")
		s = strings.ReplaceAll(s, "/hls/941318-1295296/", "/hls/498333-941317/")
		files["stitched"] = strings.ReplaceAll(s, "/hls/@@/", "/hls/941318-1295296/")
	}},
	{"the ads are other ads (control)", false, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.ReplaceAll(files["stitched"], "_ad%2Fcreative%2F", "_ad%2Fcreative%2Fzz")
	}},
	{"every ad is gone (control)", false, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		var kept []string
		for _, l := range strings.Split(files["stitched"], "\n") {
			if l != "" && !strings.HasPrefix(l, "#") && strings.Contains(l, "/mp4/") {
				if n := len(kept); n > 0 && strings.HasPrefix(kept[n-1], "#EXTINF") {
					kept = kept[:n-1]
				}
				continue
			}
			kept = append(kept, l)
		}
		files["stitched"] = strings.Join(kept, "\n")
	}},
	{"the wall clock of the stitching moves (control)", false, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.ReplaceAll(files["stitched"], "#EXT-X-PROGRAM-DATE-TIME:2026-11-16", "#EXT-X-PROGRAM-DATE-TIME:2027-01-01")
	}},
	{"the closing bumper is another clip (control)", false, "assembled", func(_ *testing.T, _ *fakePluto, files map[string]string) {
		files["stitched"] = strings.ReplaceAll(files["stitched"], "ad_bumper_animation_orbit_10", "ad_bumper_animation_orbit_11")
	}},
}

func TestTheAssembledPlaylistOverTheRealFixtureIsTheTitleAlone(t *testing.T) {
	o := chain(t, func(f *fakePluto, _ map[string]string) { refuseOwnPlaylists(f) })
	require.Empty(t, o.err)
	assert.True(t, o.adFree)
	assert.Equal(t, "/v2/stitch/hls/episode/"+fakeEpisodeID+"/master.m3u8", o.stream, "the URL that answers")
	tags, media := playlistLines(o.playlist)
	require.NotEmpty(t, media)
	for _, line := range media {
		assert.True(t, strings.HasPrefix(line, realClip), "not the title's: %s", line)
	}
	assert.Equal(t, 3, countPrefix(tags, "#EXT-X-DISCONTINUITY"), "the real title is four ranges")
	assert.Equal(t, len(media), countPrefix(tags, "#EXT-X-KEY:METHOD=AES-128"))
	assert.NotEqual(t, chain(t, nil).playlist, o.playlist, "it is not the title's own playlist: that one was refused")
}

func TestAssembledPlaylistMutationsAreNoticed(t *testing.T) {
	baseline := chain(t, func(f *fakePluto, _ map[string]string) { refuseOwnPlaylists(f) })
	require.Empty(t, baseline.err)
	require.True(t, baseline.adFree, "the baseline is the title read out of the real stitched playlist")

	for _, m := range assembledMutations {
		t.Run(m.name, func(t *testing.T) {
			got := chain(t, func(f *fakePluto, files map[string]string) {
				refuseOwnPlaylists(f)
				m.mutate(t, f, files)
			})
			if m.mustChange {
				assert.NotEqual(t, baseline.playlist, got.playlist, "the playlist handed to the player did not notice")
			} else {
				assert.Equal(t, baseline, got, "a control changed the outcome: what is spliced around the title leaked into it")
			}
		})
	}
}

func TestAssembledPlaylistMutationsKeepTheirControls(t *testing.T) {
	names := map[string]bool{}
	breaking, controls := 0, 0
	for _, m := range assembledMutations {
		assert.False(t, names[m.name], "duplicate mutation %q", m.name)
		names[m.name] = true
		assert.Equal(t, "assembled", m.layer)
		assert.Equal(t, !m.mustChange, strings.HasSuffix(m.name, "(control)"), "%q", m.name)
		if m.mustChange {
			breaking++
		} else {
			controls++
		}
	}
	assert.GreaterOrEqual(t, breaking, 5, "too few real mutations to mean anything")
	assert.GreaterOrEqual(t, controls, 3, "too few controls to tell a break from noise")
}

// The playlist is part of what the chain compares: a mutation that only
// changed it would otherwise pass for "nothing changed".
func TestTheOutcomeCarriesThePlaylist(t *testing.T) {
	a := chain(t, nil)
	require.NotEmpty(t, a.playlist)
	assert.NotContains(t, a.playlist, "127.0.0.1", "the fake's own address is taken out, so runs compare")
	b := a
	b.playlist += "#x"
	assert.NotEqual(t, a, b)
}
