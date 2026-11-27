package miruro

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the tests.
//
// A green suite proves nothing on its own: a parser that quietly returns
// empty results, or fixtures that no longer resemble what miruro.cx sends,
// both leave every assertion passing. Two guards close that gap.
//
//   - chainFailures is the contract of the whole chain, search to stream,
//     against the scripted site. The pristine fixture must satisfy it, and
//     each fixture mutant below (one thing broken in what the site sends)
//     must violate it: a mutant that passes means the chain would have
//     shipped that breakage silently.
//   - The helper contracts (candidates, wantedSubtitles, selectVariant,
//     sitePath) are held to their tables, and wrappers that tamper with the
//     real function's output must each break at least one row: a tamper that
//     survives means the table cannot tell right from wrong.

// chainFailures runs search → episodes → stream and lists what is wrong.
func chainFailures(t *testing.T, s *stub) (f []string) {
	t.Helper()
	t.Setenv("GOANIME_MIRURO_AUDIO", "")
	c := NewClientForTest(s.srv.URL)
	ctx := context.Background()
	base := s.srv.URL

	results, err := c.SearchAnime(ctx, "cowboy bebop")
	if err != nil {
		return append(f, "search: "+err.Error())
	}
	if len(results) == 0 {
		return append(f, "search: no results")
	}
	first := results[0]
	if first.Name != "Cowboy Bebop" {
		f = append(f, fmt.Sprintf("search: first result named %q", first.Name))
	}
	if first.URL != base+"/info/"+bebopID {
		f = append(f, "search: first result URL "+first.URL)
	}
	if first.AnilistID != 1 || first.MalID != 1 {
		f = append(f, fmt.Sprintf("search: ids anilist=%d mal=%d", first.AnilistID, first.MalID))
	}
	if first.Year != "1998" {
		f = append(f, "search: year "+first.Year)
	}
	if first.Audio != "Dubbed and subtitled" {
		f = append(f, "search: audio "+first.Audio)
	}
	for _, r := range results {
		if strings.Contains(r.URL, unairedID) {
			f = append(f, "search: an unaired title was listed")
		}
	}

	eps, err := c.GetAnimeEpisodes(ctx, first.URL)
	if err != nil {
		return append(f, "episodes: "+err.Error())
	}
	var numbers []string
	for _, e := range eps {
		numbers = append(numbers, e.Number)
	}
	if !slices.Equal(numbers, []string{"1", "1.5", "2"}) {
		f = append(f, fmt.Sprintf("episodes: numbers %v", numbers))
	}
	if len(eps) == 0 {
		return f
	}
	if want := base + "/watch/" + bebopID + "/" + bebopSlug + "?ep=1"; eps[0].URL != want {
		f = append(f, "episodes: first URL "+eps[0].URL)
	}
	if eps[0].Title.English != "Asteroid Blues" {
		f = append(f, "episodes: first title "+eps[0].Title.English)
	}

	streamURL, meta, err := c.GetEpisodeStreamURL(ctx, eps[0].URL, "best")
	if err != nil {
		return append(f, "stream: "+err.Error())
	}
	if streamURL != base+"/cdn/sun/master.m3u8" {
		f = append(f, "stream: picked "+streamURL)
	}
	if meta["referer"] != base+"/ultra/" || meta["origin"] != base+"/ultra" {
		f = append(f, fmt.Sprintf("stream: headers referer=%q origin=%q", meta["referer"], meta["origin"]))
	}
	if meta["audio_lang"] != "sub" {
		f = append(f, "stream: audio "+meta["audio_lang"])
	}
	if !strings.Contains(meta["subtitles"], base+"/subs/sun-en.jpeg") {
		f = append(f, "stream: subtitles "+meta["subtitles"])
	}

	v, _, err := c.GetEpisodeStreamURL(ctx, eps[0].URL, "720p")
	if err != nil || v != base+"/cdn/sun/index-720.m3u8" {
		f = append(f, fmt.Sprintf("stream 720p: %q %v", v, err))
	}
	return f
}

func TestChainContractHoldsOnThePristineSite(t *testing.T) {
	assert.Empty(t, chainFailures(t, newStub(t, nil)))
}

// sourcesOf digs into the fixture's sources; mutants edit through it.
func sourcesOf(fx *fixture) map[string]any { return fx.sources.(map[string]any) }

func tracksOf(fx *fixture) []any { return sourcesOf(fx)["tracks"].([]any) }

func providersOf(fx *fixture, track int) []any {
	return tracksOf(fx)[track].(map[string]any)["providers"].([]any)
}

func sunOf(fx *fixture) map[string]any { return providersOf(fx, 0)[1].(map[string]any) }

func sunServer(fx *fixture) map[string]any {
	return sunOf(fx)["servers"].([]any)[0].(map[string]any)
}

func bebopItem(fx *fixture) map[string]any { return fx.items[0].(map[string]any) }

// fixtureMutants each break one thing the site sends.
var fixtureMutants = map[string]func(base string, fx *fixture){
	"the result list is renamed": func(_ string, fx *fixture) {
		fx.items = nil
	},
	"the English title is gone and the romaji one differs": func(_ string, fx *fixture) {
		t := bebopItem(fx)["title"].(map[string]any)
		t["english"], t["romaji"] = nil, "Kaubōi Bibappu"
	},
	"the AniList id is missing": func(_ string, fx *fixture) {
		delete(bebopItem(fx)["external_ids"].(map[string]any), "anilist")
	},
	"the year moves to another key": func(_ string, fx *fixture) {
		it := bebopItem(fx)
		it["seasonYear"], it["season_year"] = it["season_year"], nil
	},
	"the dub count is null": func(_ string, fx *fixture) {
		bebopItem(fx)["episode_counts"].(map[string]any)["dub"] = nil
	},
	"the unaired title is marked finished": func(_ string, fx *fixture) {
		fx.items[2].(map[string]any)["status"] = "FINISHED"
	},
	"the result id is not an id": func(_ string, fx *fixture) {
		bebopItem(fx)["id"] = "not an id"
	},
	"the episode list is empty": func(_ string, fx *fixture) {
		fx.episodes = []any{}
	},
	"the episode number is a string": func(_ string, fx *fixture) {
		fx.episodes[1].(map[string]any)["number"] = "1"
	},
	"an episode title moves": func(_ string, fx *fixture) {
		e := fx.episodes[1].(map[string]any)
		e["name"], e["title"] = e["title"], nil
	},
	"the provider order is reversed": func(_ string, fx *fixture) {
		fx.order = []any{"knob", "hop", "bee", "sun"}
	},
	"the provider order is gone": func(_ string, fx *fixture) {
		fx.order = nil
	},
	"sun's Origin header is renamed": func(_ string, fx *fixture) {
		h := sunServer(fx)["headers"].(map[string]any)
		h["X-Origin"], h["Origin"] = h["Origin"], nil
	},
	"sun's Referer is wrong": func(base string, fx *fixture) {
		sunServer(fx)["headers"].(map[string]any)["Referer"] = base + "/elsewhere/"
	},
	"the soft-subbed track is renamed": func(_ string, fx *fixture) {
		tracksOf(fx)[0].(map[string]any)["track"] = "softsub"
	},
	"sun's stream is DASH": func(_ string, fx *fixture) {
		sunServer(fx)["streams"].([]any)[0].(map[string]any)["format"] = "dash"
	},
	"sun's stream URL moves": func(_ string, fx *fixture) {
		s := sunServer(fx)["streams"].([]any)[0].(map[string]any)
		s["src"], s["url"] = s["url"], nil
	},
	"sun's only subtitle is Arabic": func(_ string, fx *fixture) {
		sub := sunOf(fx)["subtitles"].([]any)[0].(map[string]any)
		sub["language"], sub["label"] = "ar", "Arabic"
	},
	"sun's subtitle file is relative": func(_ string, fx *fixture) {
		sunOf(fx)["subtitles"].([]any)[0].(map[string]any)["file"] = "/subs/sun-en.jpeg"
	},
	"sun's CDN answers HTML": func(_ string, fx *fixture) {
		c := fx.cdn["sun"]
		c.playlist = "<html>blocked</html>"
		fx.cdn["sun"] = c
	},
	"the master playlist loses its 720p variant": func(_ string, fx *fixture) {
		c := fx.cdn["sun"]
		c.playlist = strings.Replace(masterPlaylist, "1280x720", "1280x719", 1)
		fx.cdn["sun"] = c
	},
	"the sources are null": func(_ string, fx *fixture) {
		fx.sources = nil
	},
}

func TestEveryFixtureMutantBreaksTheChain(t *testing.T) {
	for name, mutate := range fixtureMutants {
		t.Run(name, func(t *testing.T) {
			failures := chainFailures(t, newStub(t, mutate))
			assert.NotEmpty(t, failures, "mutant %q survived: the chain did not notice", name)
		})
	}
}

// ---- helper contracts and the tampers they must catch ----------------------

type candidatesFn func(*sourcesData, string) []candidate

func candidateFailures(fn candidatesFn) (f []string) {
	for _, tt := range candidateCases() {
		var got []string
		for _, c := range fn(tt.data, tt.prefer) {
			got = append(got, c.provider+"/"+c.server)
		}
		if !slices.Equal(got, tt.want) {
			f = append(f, fmt.Sprintf("%s: got %v, want %v", tt.name, got, tt.want))
		}
	}
	return f
}

var candidateTampers = map[string]func(candidatesFn) candidatesFn{
	"reversed": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, p string) []candidate { out := fn(d, p); slices.Reverse(out); return out }
	},
	"first dropped": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, p string) []candidate {
			out := fn(d, p)
			if len(out) > 0 {
				out = out[1:]
			}
			return out
		}
	},
	"preference ignored": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, _ string) []candidate { return fn(d, trackHardSub) }
	},
	"preference inverted": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, p string) []candidate {
			if p == trackDub {
				return fn(d, trackHardSub)
			}
			return fn(d, trackDub)
		}
	},
	"provider order ignored": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, p string) []candidate {
			return fn(&sourcesData{Tracks: d.Tracks}, p)
		}
	},
	"a repeat kept": func(fn candidatesFn) candidatesFn {
		return func(d *sourcesData, p string) []candidate {
			out := fn(d, p)
			if len(out) > 0 {
				out = append(out, out[len(out)-1])
			}
			return out
		}
	},
}

func TestCandidateContract(t *testing.T) {
	t.Parallel()
	require.Empty(t, candidateFailures(candidates), "the real function must satisfy its own table")
	for name, tamper := range candidateTampers {
		assert.NotEmpty(t, candidateFailures(tamper(candidates)), "tamper %q survived", name)
	}
}

type subtitlesFn func([]sourceSubtitle) []sourceSubtitle

var subtitleCase = []sourceSubtitle{
	{Language: "ar", File: "https://s/ar", Default: true},
	{Language: "pt-BR", File: "https://s/pt"},
	{Language: "en", File: "https://s/en2"},
	{Language: "en", File: "https://s/en1", Default: true},
	{Language: "en", File: "relative.vtt"},
}

func subtitleFailures(fn subtitlesFn) (f []string) {
	var files []string
	for _, s := range fn(subtitleCase) {
		files = append(files, s.File)
	}
	if want := []string{"https://s/en1", "https://s/en2", "https://s/pt"}; !slices.Equal(files, want) {
		f = append(f, fmt.Sprintf("got %v, want %v", files, want))
	}
	if got := fn(nil); len(got) != 0 {
		f = append(f, "nothing in, something out")
	}
	return f
}

var subtitleTampers = map[string]func(subtitlesFn) subtitlesFn{
	"everything kept": func(subtitlesFn) subtitlesFn {
		return func(in []sourceSubtitle) []sourceSubtitle { return in }
	},
	"order reversed": func(fn subtitlesFn) subtitlesFn {
		return func(in []sourceSubtitle) []sourceSubtitle { out := fn(in); slices.Reverse(out); return out }
	},
	"Portuguese dropped": func(fn subtitlesFn) subtitlesFn {
		return func(in []sourceSubtitle) []sourceSubtitle {
			return slices.DeleteFunc(fn(in), func(s sourceSubtitle) bool { return subtitleLanguage(s) == "Portuguese" })
		}
	},
	"site order trusted": func(fn subtitlesFn) subtitlesFn {
		return func(in []sourceSubtitle) []sourceSubtitle {
			keep := fn(in)
			var out []sourceSubtitle
			for _, s := range in {
				if slices.Contains(keep, s) {
					out = append(out, s)
				}
			}
			return out
		}
	},
}

func TestSubtitleContract(t *testing.T) {
	t.Parallel()
	require.Empty(t, subtitleFailures(wantedSubtitles))
	for name, tamper := range subtitleTampers {
		assert.NotEmpty(t, subtitleFailures(tamper(wantedSubtitles)), "tamper %q survived", name)
	}
}

type variantFn func([]byte, string, string) (string, bool)

func variantFailures(fn variantFn) (f []string) {
	base := "https://cdn.example/a/master.m3u8"
	rows := []struct {
		quality string
		want    string
		ok      bool
	}{
		{"720p", "https://cdn.example/a/index-720.m3u8", true},
		{"360", "https://cdn.example/a/index-360.m3u8", true},
		{"1080", "https://elsewhere.example/1080.m3u8", true},
		{"best", "", false},
		{"480", "", false},
	}
	for _, r := range rows {
		got, ok := fn([]byte(masterPlaylist), base, r.quality)
		if got != r.want || ok != r.ok {
			f = append(f, fmt.Sprintf("%s: got %q %v", r.quality, got, ok))
		}
	}
	return f
}

var variantTampers = map[string]func(variantFn) variantFn{
	"always the master": func(variantFn) variantFn {
		return func([]byte, string, string) (string, bool) { return "", false }
	},
	"never resolved": func(fn variantFn) variantFn {
		return func(p []byte, u, q string) (string, bool) {
			got, ok := fn(p, u, q)
			return strings.TrimPrefix(got, "https://cdn.example/a/"), ok
		}
	},
	"first variant whatever was asked": func(fn variantFn) variantFn {
		return func(p []byte, u, _ string) (string, bool) { return fn(p, u, "360") }
	},
}

func TestVariantContract(t *testing.T) {
	t.Parallel()
	require.Empty(t, variantFailures(selectVariant))
	for name, tamper := range variantTampers {
		assert.NotEmpty(t, variantFailures(tamper(selectVariant)), "tamper %q survived", name)
	}
}

type sitePathFn func(string) bool // reports acceptance

func sitePathFailures(fn sitePathFn) (f []string) {
	for loc, want := range map[string]bool{
		"/watch/a/b?ep=1":         true,
		"/info/a":                 true,
		"https://evil.example/a":  false,
		"//evil.example/a":        false,
		"watch/a":                 false,
		"javascript:alert(1)":     false,
		"https://miruro.cx/watch": false,
	} {
		if fn(loc) != want {
			f = append(f, fmt.Sprintf("%s: accepted=%v", loc, !want))
		}
	}
	return f
}

func realSitePath(loc string) bool { _, err := sitePath(loc); return err == nil }

func TestSitePathContract(t *testing.T) {
	t.Parallel()
	require.Empty(t, sitePathFailures(realSitePath))
	tampers := map[string]sitePathFn{
		"accepts everything":        func(string) bool { return true },
		"accepts any slash prefix":  func(l string) bool { return strings.HasPrefix(l, "/") },
		"accepts absolute URLs too": func(l string) bool { return realSitePath(l) || strings.HasPrefix(l, "https://") },
		"refuses queries":           func(l string) bool { return realSitePath(l) && !strings.Contains(l, "?") },
	}
	for name, tamper := range tampers {
		assert.NotEmpty(t, sitePathFailures(tamper), "tamper %q survived", name)
	}
}
