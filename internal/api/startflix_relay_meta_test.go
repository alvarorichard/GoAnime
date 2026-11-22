package api

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests for the StartFlix playlist relay (hlsrelay.go).
//
// Oppenheimer and Jujutsu Kaisen (2026-11-21) resolved to a host that serves
// its HLS playlist at /pl/<id>: no extension, a MIME type that is not a
// playlist's, and a token every rendition and segment must carry. mpv read it
// as a plain list of files and quit in a third of a second. Each mutant below
// puts one part of that back, or breaks one decision of the relay; the guard
// tests it names must fail on it. Every function of hlsrelay.go must have a
// mutant, and be fully covered by its package's tests.

const (
	sfxPkg       = "./internal/scraper/providers/startflix"
	sfxRelayFile = "internal/scraper/providers/startflix/hlsrelay.go"
)

func sm(name, file, from, to string, tests ...string) mutant {
	return mutant{name: name, edits: []codeEdit{{file, from, to}}, kills: map[string][]string{sfxPkg: tests}}
}

// relayMutants maps each function of hlsrelay.go to its mutants, so the table
// can be held to "every function has one".
var relayMutants = map[string][]mutant{
	"newPlaylistRelay": {
		sm("the relay has no index of what it published", sfxRelayFile,
			"lru: list.New(), byUpstream: map[string]string{}}", "lru: list.New()}",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
	},
	"sharedPlaylistRelay": {
		sm("every stream starts a relay of its own", sfxRelayFile,
			"relayOnce.Do(func() { relayShared = newPlaylistRelay() })", "relayShared = newPlaylistRelay()",
			"Test2EmbedRidesBehindVIDSRCFallback", "TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
	},
	"hasPlaylistName": {
		sm("a .m3u8 address is relayed anyway", sfxRelayFile,
			`return strings.HasSuffix(p, ".m3u8") || strings.HasSuffix(p, ".m3u")`, `return strings.HasSuffix(p, ".m3u")`,
			"TestHasPlaylistName", "TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
		sm("the name is read off the query", sfxRelayFile,
			"p := strings.ToLower(u.Path)", "p := strings.ToLower(u.String())",
			"TestHasPlaylistName"),
		sm("an upper-case extension is no name", sfxRelayFile,
			"p := strings.ToLower(u.Path)", "p := u.Path",
			"TestHasPlaylistName"),
	},
	"nameThePlaylist": {
		sm("the pre-fix player: an unnamed HLS stream is handed over as it is", sfxRelayFile,
			"if s == nil || !s.HLS || hasPlaylistName(s.URL) {\n\t\treturn s\n\t}", "if s != nil || s == nil {\n\t\treturn s\n\t}",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream", "Test2EmbedRidesBehindVIDSRCFallback"),
		sm("a stream that is not HLS is relayed", sfxRelayFile,
			"if s == nil || !s.HLS || hasPlaylistName(s.URL) {", "if s == nil || hasPlaylistName(s.URL) {",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
		sm("the relay reads with the fallback, not the client that read the playlist", sfxRelayFile,
			"\tclient := s.client\n\tif client == nil {", "\tclient := fallback\n\tif client == nil {",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
		sm("the relay reads with no client when the stream names none", sfxRelayFile,
			"\tif client == nil {\n\t\tclient = fallback\n\t}", "\tif client == nil {\n\t\tclient = nil\n\t}",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
		sm("a relay that failed hands over nothing", sfxRelayFile,
			"\t\tutil.Debug(\"StartFlix playlist is played from its own address\", \"host\", s.Host, \"err\", err)\n\t\treturn s", "\t\tutil.Debug(\"StartFlix playlist is played from its own address\", \"host\", s.Host, \"err\", err)\n\t\treturn nil",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
		sm("the relayed stream loses its referer", sfxRelayFile,
			"\tout := *s\n\tout.URL = local", "\tout := Stream{URL: local, Height: s.Height}",
			"TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream"),
	},
	"publish": {
		sm("any address is published", sfxRelayFile,
			`if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {`, `if err != nil || u == nil {`,
			"TestTheRelayRefusesWhatIsNoAddress"),
		sm("a failed listener is reported as success", sfxRelayFile,
			`return "", fmt.Errorf("local playlist relay: %w", err)`, `return "http://127.0.0.1:1/hls/x.m3u8", nil`,
			"TestTheRelayReportsThatItCannotListen"),
		sm("every publication starts a listener of its own", sfxRelayFile,
			`if r.base == "" {`, "if r.base != \"\\x00\" {",
			"TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo"),
		sm("the relay is not known to the download guard", sfxRelayFile,
			"util.RegisterLocalProxyHost(ln.Addr().String())", "_ = ln.Addr().String()",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("a playlist read again gets a new address", sfxRelayFile,
			"\tif id, ok := r.byUpstream[key]; ok {", "\tif id, ok := r.byUpstream[key]; ok && false {",
			"TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo"),
		sm("the referer is not part of what a playlist is known by", sfxRelayFile,
			"\tkey := upstream + \"\\x00\" + referer\n", "\tkey := upstream\n",
			"TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound"),
		sm("the relay remembers one playlist past its bound", sfxRelayFile,
			"for r.lru.Len() > maxRelayedPlaylists {", "for r.lru.Len() > maxRelayedPlaylists+1 {",
			"TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound"),
		sm("the index outlives what it points at", sfxRelayFile,
			"\t\tdelete(r.byUpstream, gone.upstream+\"\\x00\"+gone.referer)\n", "",
			"TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound"),
		sm("the newest playlist is the first forgotten", sfxRelayFile,
			"r.lists[id] = r.lru.PushFront(", "r.lists[id] = r.lru.PushBack(",
			"TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound"),
		sm("the relayed address is not named as a playlist", sfxRelayFile,
			"\treturn r.base + \"/hls/\" + id + \".m3u8\", nil\n}\n\nfunc (r *playlistRelay) lookup", "\treturn r.base + \"/hls/\" + id, nil\n}\n\nfunc (r *playlistRelay) lookup",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
	},
	"lookup": {
		sm("a playlist in use is forgotten like any other", sfxRelayFile,
			"\tr.lru.MoveToFront(el)\n\treturn el.Value.(*relayedPlaylist), true", "\treturn el.Value.(*relayedPlaylist), true",
			"TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound"),
		sm("an unknown id is looked up", sfxRelayFile,
			"\tif !ok {\n\t\treturn nil, false\n\t}\n\tr.lru.MoveToFront(el)", "\tif !ok {\n\t\treturn &relayedPlaylist{upstream: \"http://127.0.0.1:1/\"}, true\n\t}\n\tr.lru.MoveToFront(el)",
			"TestTheRelayServesNothingElse"),
	},
	"ServeHTTP": {
		sm("the relay answers any method", sfxRelayFile,
			"if req.Method != http.MethodGet && req.Method != http.MethodHead {", "if req.Method == \"\\x00\" {",
			"TestTheRelayServesNothingElse"),
		sm("the relay refuses HEAD", sfxRelayFile,
			"if req.Method != http.MethodGet && req.Method != http.MethodHead {", "if req.Method != http.MethodGet {",
			"TestTheRelayServesNothingElse"),
		sm("a path with no name is served", sfxRelayFile,
			"\tif !ok {\n\t\thttp.NotFound(w, req)\n\t\treturn\n\t}\n\tp, found", "\tif !ok && false {\n\t\thttp.NotFound(w, req)\n\t\treturn\n\t}\n\tp, found",
			"TestTheRelayServesNothingElse"),
		sm("a host's failure is handed to the player as an empty playlist", sfxRelayFile,
			"\t\thttp.Error(w, \"playlist unavailable\", http.StatusBadGateway)\n\t\treturn", "\t\tbody = \"#EXTM3U\\n\"",
			"TestTheRelayTellsThePlayerWhenTheHostFails"),
		sm("the playlist is served as plain text", sfxRelayFile,
			`w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")`, `w.Header().Set("Content-Type", "text/plain")`,
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("the host's playlist is served unrewritten", sfxRelayFile,
			"\ttext := r.rewrite(body, p)\n", "\ttext := body\n",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
	},
	"read": {
		sm("the host is asked without its referer", sfxRelayFile,
			"\t\tup.Header.Set(\"Referer\", p.referer)\n", "",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("the host is asked with no User-Agent", sfxRelayFile,
			"\tup.Header.Set(\"User-Agent\", netx.UserAgent)\n", "\tup.Header.Set(\"User-Agent\", \"\")\n",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("the default client has no SSRF guard", sfxRelayFile,
			"client = &http.Client{Transport: netx.SafeScraperTransport(relayTimeout)}", "client = http.DefaultClient",
			"TestTheRelayWithNoClientOfItsOwnKeepsTheSSRFGuard"),
		sm("an error page is relayed as the playlist", sfxRelayFile,
			"if resp.StatusCode != http.StatusOK {\n\t\treturn \"\", fmt.Errorf(\"HTTP %d\", resp.StatusCode)", "if resp.StatusCode >= 600 {\n\t\treturn \"\", fmt.Errorf(\"HTTP %d\", resp.StatusCode)",
			"TestTheRelayTellsThePlayerWhenTheHostFails"),
		sm("a playlist over its limit is read whole", sfxRelayFile,
			"if len(raw) > maxRelayedBytes {", "if len(raw) > maxRelayedBytes+1 {",
			"TestTheRelayTellsThePlayerWhenTheHostFails"),
		sm("a playlist cut short is relayed", sfxRelayFile,
			"\traw, err := io.ReadAll(io.LimitReader(resp.Body, maxRelayedBytes+1))\n\tif err != nil {", "\traw, err := io.ReadAll(io.LimitReader(resp.Body, maxRelayedBytes+1))\n\tif err != nil && raw == nil {",
			"TestTheRelayReportsAPlaylistItCannotRead"),
		sm("what is not a playlist is relayed as one", sfxRelayFile,
			`if !strings.HasPrefix(text, "#EXTM3U") {`, "if text == \"\\x00\" {",
			"TestTheRelayTellsThePlayerWhenTheHostFails"),
	},
	"rewrite": {
		sm("the pre-fix relay: the token is not handed on", sfxRelayFile,
			"if abs.RawQuery == \"\" && base.RawQuery != \"\" && strings.EqualFold(abs.Host, base.Host) {", "if false {",
			"TestTheRelayHandsTheTokenOnToWhatTheHostListsWithoutOne"),
		sm("the token is handed to another host", sfxRelayFile,
			"if abs.RawQuery == \"\" && base.RawQuery != \"\" && strings.EqualFold(abs.Host, base.Host) {", "if abs.RawQuery == \"\" && base.RawQuery != \"\" {",
			"TestTheRelayHandsTheTokenOnToWhatTheHostListsWithoutOne"),
		sm("the token overwrites a URI's own query", sfxRelayFile,
			"if abs.RawQuery == \"\" && base.RawQuery != \"\" && strings.EqualFold(abs.Host, base.Host) {", "if base.RawQuery != \"\" && strings.EqualFold(abs.Host, base.Host) {",
			"TestTheRelayHandsTheTokenOnToWhatTheHostListsWithoutOne"),
		sm("an empty query is handed on as a stray ?", sfxRelayFile,
			"if abs.RawQuery == \"\" && base.RawQuery != \"\" && strings.EqualFold(abs.Host, base.Host) {", "if abs.RawQuery == \"\" && strings.EqualFold(abs.Host, base.Host) {\n\t\t\tabs.ForceQuery = true",
			"TestTheRelayHandsNoQueryOnWhenThePlaylistHadNone"),
		sm("segments stay relative", sfxRelayFile,
			"\t\tdefault:\n\t\t\tlines[i] = resolve(trimmed)", "\t\tdefault:\n\t\t\tlines[i] = trimmed",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("keys stay relative", sfxRelayFile,
			"\t\t\t\treturn `URI=\"` + resolve(attr[len(`URI=\"`):len(attr)-1]) + `\"`", "\t\t\t\treturn attr",
			"TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs"),
		sm("a master's renditions are not relayed", sfxRelayFile,
			"\t\tcase nextIsPlaylist:\n\t\t\tlines[i] = relay(trimmed)", "\t\tcase nextIsPlaylist:\n\t\t\tlines[i] = resolve(trimmed)",
			"TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo", "TestTheRelayHandsTheTokenOnToWhatTheHostListsWithoutOne"),
		sm("every line after a rendition is taken for one", sfxRelayFile,
			"\t\t\tlines[i] = relay(trimmed)\n\t\t\tnextIsPlaylist = false", "\t\t\tlines[i] = relay(trimmed)",
			"TestOnlyTheLineAfterAStreamInfIsARendition"),
		sm("an alternative audio playlist is not relayed", sfxRelayFile,
			"\t\t\t\treturn `URI=\"` + relay(attr[len(`URI=\"`):len(attr)-1]) + `\"`", "\t\t\t\treturn `URI=\"` + resolve(attr[len(`URI=\"`):len(attr)-1]) + `\"`",
			"TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo"),
		sm("a named rendition is relayed too", sfxRelayFile,
			"\t\tif hasPlaylistName(abs) {\n\t\t\treturn abs\n\t\t}", "\t\tif hasPlaylistName(abs) && false {\n\t\t\treturn abs\n\t\t}",
			"TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo"),
		sm("a rendition that cannot be relayed is dropped", sfxRelayFile,
			"\t\tlocal, err := r.publish(abs, p.referer, p.client)\n\t\tif err != nil {\n\t\t\treturn abs", "\t\tlocal, err := r.publish(abs, p.referer, p.client)\n\t\tif err != nil {\n\t\t\treturn \"\"",
			"TestTheRelayLeavesWhatItCannotResolveOrRelay"),
		sm("a line that is no URL is dropped", sfxRelayFile,
			"\t\tif err != nil {\n\t\t\treturn ref\n\t\t}\n\t\tabs := base.ResolveReference(u)", "\t\tif err != nil {\n\t\t\treturn \"\"\n\t\t}\n\t\tabs := base.ResolveReference(u)",
			"TestTheRelayLeavesWhatItCannotResolveOrRelay"),
		sm("CRLF playlists keep their carriage returns", sfxRelayFile,
			`strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")`, `strings.Split(body, "\n")`,
			"TestTheRelayReadsACRLFPlaylist"),
	},
	// The resolvers that know a stream is HLS must say so, or nothing is relayed.
	"resolvers": {
		sm("2Embed does not say its stream is HLS", "internal/scraper/providers/startflix/vidsrcbuzz.go",
			"HLS:     true, // play answered type \"hls\", and the master was read", "HLS:     false, // play answered type \"hls\", and the master was read",
			"Test2EmbedRidesBehindVIDSRCFallback"),
		sm("the fallbacks hand their stream over unrelayed", "internal/scraper/providers/startflix/vidsrc.go",
			"\t\treturn nameThePlaylist(buzzStream, nil), nil\n\t}\n\treturn nil, err\n}\n\n// EpisodeFallback", "\t\treturn buzzStream, nil\n\t}\n\treturn nil, err\n}\n\n// EpisodeFallback",
			"Test2EmbedRidesBehindVIDSRCFallback"),
		sm("the episode fallback hands its stream over unrelayed", "internal/scraper/providers/startflix/vidsrc.go",
			"\tif buzzStream, buzzErr := resolve2Embed(ctx, KindSeries, \"\", tmdbID, season, episode); buzzErr == nil {\n\t\treturn nameThePlaylist(buzzStream, nil), nil", "\tif buzzStream, buzzErr := resolve2Embed(ctx, KindSeries, \"\", tmdbID, season, episode); buzzErr == nil {\n\t\treturn buzzStream, nil",
			"Test2EmbedEpisodeRidesBehindVIDSRCFallback"),
		sm("ResolveStreamFor hands its stream over unrelayed", "internal/scraper/providers/startflix/client.go",
			"\treturn nameThePlaylist(s, c.proxyClient()), nil", "\treturn s, nil",
			"TestResolveStreamForRelaysAnUnnamedHLSStream"),
	},
}

func relayMutantList() []mutant {
	keys := make([]string, 0, len(relayMutants))
	for k := range relayMutants {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []mutant
	for _, k := range keys {
		out = append(out, relayMutants[k]...)
	}
	return out
}

func TestStartFlixRelayGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, relayMutantList())
}

func TestStartFlixRelayFunctionsAreAllMutated(t *testing.T) {
	root := moduleRoot(t)
	for _, fn := range declaredFuncs(t, filepath.Join(root, filepath.FromSlash(sfxRelayFile))) {
		assert.NotEmpty(t, relayMutants[fn], "hlsrelay.go %s has no mutant: break one decision in it and name the test that must fail", fn)
	}
	for _, m := range relayMutantList() {
		for _, e := range m.edits {
			src := readModuleFile(t, root, e.file)
			assert.Equal(t, 1, strings.Count(src, e.from), "mutant %q: its snippet must be in %s exactly once", m.name, e.file)
		}
	}
}

func TestStartFlixRelayFunctionsAreFullyCovered(t *testing.T) {
	skipMutantsWhenShort(t)
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH")
	}
	root := moduleRoot(t)
	profile := filepath.Join(t.TempDir(), "cover.out")
	run := exec.Command(goBin, "test", "-count=1", "-coverprofile="+profile, sfxPkg) // #nosec G204 -- the go toolchain
	run.Dir = root
	out, err := run.CombinedOutput()
	require.NoError(t, err, string(out))
	report := exec.Command(goBin, "tool", "cover", "-func="+profile) // #nosec G204
	report.Dir = root
	table, err := report.CombinedOutput()
	require.NoError(t, err, string(table))
	funcs := declaredFuncs(t, filepath.Join(root, filepath.FromSlash(sfxRelayFile)))
	assert.Empty(t, notFullyCovered(string(table), "hlsrelay.go", funcs))
}

func readModuleFile(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return lf(string(raw))
}
