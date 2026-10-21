package startflix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// This file implements the second id-keyed fallback source: 2Embed's own
// player, which fronts on vidsrc.buzz (the first is the vidsrc family in
// vidsrc.go). The two share only the id keys — the engine is wholly
// different, and simpler: plain HTTP, bearer-signed refs, no wasm, no host
// token. Reverse-engineered from the embed page's inline player on
// 2026-10-21:
//
//  1. GET  {base}/embed/{movie|tv}/{id}
//     movie takes an IMDb id (preferred) or TMDB id; tv takes a TMDB id
//     plus an optional /{s}/{e} suffix (a bare /embed/tv/{id} page mints
//     its own default pair, which is what season/episode 0 means). The
//     page carries an inline `var Q = {...}`: a JSON object with a token
//     `t` — minted per render, valid ~4.5h (2026-10-21) — and the
//     candidate servers under Q["ssr"]["servers"], each a map with a
//     signed `ref`.
//  2. GET  {base}/pl/api.php?a=play&ref=…&t=…
//     → {"url":"/_stream?id=…","type":"hls"} — or {"error":"unavailable"}
//     when a server lost its copy. The player page retries once with
//     &fresh=1; the chain mirrors that.
//  3. GET  {base}/_stream?id=… → master playlist. Renditions are
//     /_stream ids of their own (relative in the manifest).
//
// One trap in step 3 motivated the rendition walk below: on 2026-10-21
// several titles' servers answered play with masters whose every rendition
// was an ad playlist — segments pointing at p16-ad-site-sign-sg.tiktokcdn.com
// carrying .image payloads instead of video — and the real segment relay
// (relay*.videm.xyz) answered 500 on its first media segment. A bare master
// probe calls all of that "playable", so the chain reads each rendition,
// refuses any whose segments name an ad host, and range-fetches the first
// segment of the rest before handing the master over. A rendition that
// fails moves on to the next one, then the next signed ref, mirroring the
// player page's own server-hopping.
//
// The chain ships off unless GOANIME_2EMBED_URL names a base — the same env
// both enables it and repoints it. Default-off because the source's content
// state above was recorded on 2026-10-21; standing this on by default would
// spend real playback turns on a chain that answers but does not play. The
// protocol notes above stay so re-auditing it when it recovers is a
// five-minute job, not a new reverse-engineering pass.

const (
	// Default2EmbedBase is the front 2Embed's "Server 2" player runs behind.
	Default2EmbedBase = "https://vidsrc.buzz"

	// Default2EmbedEnv both enables the chain and repoints its base when
	// set: "" (the default) keeps the chain off the air.
	Default2EmbedEnv = "GOANIME_2EMBED_URL"

	// buzzBodyLimit caps every body the chain reads. Embed pages run the
	// size of a normal page (Q included); VOD media playlists go larger —
	// The Godfather's 1080p rendition listed 855 segments and measured
	// 584 KB (2026-10-21) — so the cap is generous, and the drawdown is
	// "the first segments", where every probe's evidence lives.
	buzzBodyLimit = 4 << 20

	// buzzSegmentProbeLimit is the read cap for the first-segment range
	// probe: a slice large enough to prove bytes flow, far below a real
	// segment's size.
	buzzSegmentProbeLimit = 64 << 10

	// buzzRefReportLen is how much of a signed ref debug output quotes, so
	// logs never carry a token wholesale.
	buzzRefReportLen = 12
)

// err2EmbedDisabled mirrors errVIDSRCDisabled for the gate-off path.
var err2EmbedDisabled = fmt.Errorf("2embed fallback disabled")

// err2EmbedUnavailable marks a ref whose server has lost its copy of the
// title (play answers {"error":"unavailable"}): no fresh mint brings a copy
// back, so the fresh=1 retry for such a ref would be a wasted request.
var err2EmbedUnavailable = errors.New("2embed: server copy unavailable")

// buzzFallbackAllowed reports whether the chain may run: only when
// GOANIME_2EMBED_URL names a base. It is not wired to the test-binary gate:
// the vidsrc gate already keeps MovieFallback away from the network under
// `go test`, and the dedicated 2Embed tests set the env themselves.
func buzzFallbackAllowed() bool {
	return strings.TrimSpace(os.Getenv(Default2EmbedEnv)) != ""
}

// buzzBasesFn names the bases the chain walks. The env names exactly one;
// the shape is a list (like vidsrcBasesFn) so a mirror can join the moment
// one is observed in the wild without touching the call sites.
var buzzBasesFn = func() []string {
	b := strings.TrimRight(strings.TrimSpace(os.Getenv(Default2EmbedEnv)), "/")
	if b == "" {
		return nil
	}
	return []string{b}
}

var buzzHTTPOnce sync.Once
var buzzHTTP *http.Client

// buzzHTTPDefault builds the chain's real client: one fast
// Chrome-impersonating client reused across every host, as the vidsrc chain
// does (see vidsrcHTTPDefault) — none of the hops needs cookie continuity.
var buzzHTTPDefault = func() *http.Client {
	buzzHTTPOnce.Do(func() { buzzHTTP = util.NewFastClient() })
	return buzzHTTP
}

// buzzHTTPFn returns the HTTP client the chain runs on. A seam so tests can
// route every host through one httptest server.
var buzzHTTPFn = buzzHTTPDefault

// resolve2Embed runs the chain for one title, walking the env-named bases.
// Errors stay internal: MovieFallback reports the original vidsrc error and
// Debug output records why this chain missed, so user-facing wording still
// explains the panel situation rather than a source the user never chose.
func resolve2Embed(ctx context.Context, kind MediaKind, imdbID string, tmdbID, season, episode int) (*Stream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !buzzFallbackAllowed() {
		return nil, err2EmbedDisabled
	}
	var lastErr error
	for _, base := range buzzBasesFn() {
		stream, err := resolve2EmbedOn(ctx, base, kind, imdbID, tmdbID, season, episode)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lastErr = err
		util.Debug("2Embed fallback base missed", "base", base, "err", err)
	}
	return nil, lastErr
}

// resolve2EmbedOn runs the chain on one base: embed page → per-ref play
// calls → rendition checks. The signed refs are tried in the order the page
// named them, because it ranks them fastest-first.
func resolve2EmbedOn(ctx context.Context, base string, kind MediaKind, imdbID string, tmdbID, season, episode int) (*Stream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pageURL := base + embed2EmbedPath(kind, imdbID, tmdbID, season, episode)
	body, err := buzzGetBody(ctx, pageURL, base+"/")
	if err != nil {
		return nil, err
	}
	cfg := buzzConfig(body)
	if cfg == nil {
		return nil, netx.NewParserError(SourceName, "2embed", "no config object in the embed page", nil)
	}
	token := strField(cfg, "t")
	if token == "" {
		return nil, netx.NewParserError(SourceName, "2embed", "the config carries no signed token", nil)
	}
	refs := buzzServerRefs(cfg)
	if len(refs) == 0 {
		return nil, netx.NewParserError(SourceName, "2embed", "no candidate servers in the config", nil)
	}
	var lastErr error
	for _, ref := range refs {
		stream, err := play2EmbedRef(ctx, base, pageURL, ref, token)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lastErr = err
		util.Debug("2Embed fallback server missed", "server", buzzRefReport(ref), "err", err)
	}
	return nil, fmt.Errorf("%d server(s) listed, none played: %w", len(refs), lastErr)
}

// embed2EmbedPath builds the embed page path for one title. Movies take the
// IMDb id when the record has one and the TMDB id otherwise; series key on
// the TMDB id with the season/episode pair when both are known.
func embed2EmbedPath(kind MediaKind, imdbID string, tmdbID, season, episode int) string {
	if kind == KindMovie {
		id := imdbID
		if id == "" {
			id = strconv.Itoa(tmdbID)
		}
		return "/embed/movie/" + id
	}
	path := "/embed/tv/" + strconv.Itoa(tmdbID)
	if season > 0 && episode > 0 {
		return path + "/" + strconv.Itoa(season) + "/" + strconv.Itoa(episode)
	}
	return path
}

// buzzServerRefs walks the config for the server list. The live shape is
// Q["ssr"]["servers"], each entry a map with a string "ref"; the same key at
// the top level and then a deterministic (key-sorted) depth-first scan
// follow, so a shifted shape degrades to "no candidate servers", not to a
// pick that changes run to run.
func buzzServerRefs(cfg map[string]any) []string {
	if ssr, ok := cfg["ssr"].(map[string]any); ok {
		if out := buzzRefsOf(ssr["servers"]); out != nil {
			return out
		}
	}
	if out := buzzRefsOf(cfg["servers"]); out != nil {
		return out
	}
	return buzzFirstServerList(cfg)
}

// buzzRefsOf reads a server list: every map with a non-empty string "ref".
func buzzRefsOf(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if ref := strField(m, "ref"); ref != "" {
			out = append(out, ref)
		}
	}
	return out
}

// buzzFirstServerList depth-first searches a decoded config for the first
// list of maps with "ref" entries, scanning map keys in sorted order so the
// result is stable when a config carries look-alike lists.
func buzzFirstServerList(v any) []string {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if out := buzzFirstServerList(t[k]); out != nil {
				return out
			}
		}
	case []any:
		if out := buzzRefsOf(t); len(out) > 0 {
			return out
		}
		for _, item := range t {
			if out := buzzFirstServerList(item); out != nil {
				return out
			}
		}
	}
	return nil
}

// play2EmbedRef resolves one signed ref to a playable stream. One &fresh=1
// retry mirrors the player page's own recovery for a mint that died
// (checked 2026-10-21: the page re-calls play with fresh=1 when a server's
// first answer comes back unusable).
func play2EmbedRef(ctx context.Context, base, pageURL, ref, token string) (*Stream, error) {
	lastErr := fmt.Errorf("play not attempted")
	for _, fresh := range []bool{false, true} {
		stream, err := play2EmbedRefOnce(ctx, base, pageURL, ref, token, fresh)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lastErr = err
		if errors.Is(err, err2EmbedUnavailable) {
			// The server has no copy regardless of the mint: any further
			// retry would burn the same answer back.
			break
		}
	}
	return nil, lastErr
}

func play2EmbedRefOnce(ctx context.Context, base, pageURL, ref, token string, fresh bool) (*Stream, error) {
	var out struct {
		URL   string `json:"url"`
		Type  string `json:"type"`
		Error string `json:"error"`
	}
	if err := buzzGetJSON(ctx, buzzPlayURL(base, ref, token, fresh), pageURL, &out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		return nil, fmt.Errorf("%w: %s", err2EmbedUnavailable, out.Error)
	}
	if out.URL == "" {
		return nil, fmt.Errorf("play answered without a stream")
	}
	if !strings.EqualFold(out.Type, "hls") {
		return nil, fmt.Errorf("play answered type %q, not hls", out.Type)
	}
	master := absolutize(out.URL, base)
	mu, err := url.Parse(master)
	if err != nil || mu.Host == "" {
		return nil, fmt.Errorf("play answered a malformed stream URL")
	}
	streamURL, height, err := validate2EmbedMaster(ctx, master, pageURL)
	if err != nil {
		return nil, err
	}
	return &Stream{
		// The validated rendition, not the master: the master ranks its
		// renditions itself, and an ad-laced one may sit first (checked
		// 2026-10-21, The Godfather — both renditions ad playlists in the
		// worst case). Landing the player on the rendition that survived
		// the checks skips that trap entirely.
		URL: streamURL,
		// The embed page is the document hls.js runs inside, and its meta
		// referrer policy (no-referrer-when-downgrade) sends the full page
		// URL as the referer — so that is what the player carries too.
		Referer: pageURL,
		Host:    mu.Host,
		Height:  height,
		Audio:   AudioSubtitled,
	}, nil
}

// buzzPlayURL builds one signed play call. The player page resolves
// api.php against <base href="/pl/">, so the endpoint lives at the
// origin's /pl/api.php regardless of the embed path.
func buzzPlayURL(base, ref, token string, fresh bool) string {
	q := url.Values{"a": {"play"}, "ref": {ref}, "t": {token}}
	if fresh {
		q.Set("fresh", "1")
	}
	return strings.TrimRight(base, "/") + "/pl/api.php?" + q.Encode()
}

// rendition2Embed is one master entry: the rendition playlist URL and the
// height its RESOLUTION names (0 when the manifest does not say).
type rendition2Embed struct {
	url    string
	height int
}

// validate2EmbedMaster walks a master the browser would: reads it, checks
// each listed rendition in turn, and returns the first one that survives —
// its URL and height — so the caller never hands the player a dead or
// ad-flooded playlist. A master carrying no rendition (a media playlist
// straight away) is validated on its own segments.
func validate2EmbedMaster(ctx context.Context, masterURL, referer string) (string, int, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	body, err := buzzGetBody(probeCtx, masterURL, referer)
	if err != nil {
		return "", 0, err
	}
	if !bytes.Contains(body, []byte("#EXTM3U")) {
		return "", 0, fmt.Errorf("the master carries no playlist")
	}
	renditions := rendition2EmbedURLs(body, masterURL)
	if len(renditions) == 0 {
		height := resolutionHeight(body)
		if err := probe2EmbedSegments(probeCtx, body, masterURL, referer); err != nil {
			return "", 0, err
		}
		return masterURL, height, nil
	}
	var lastErr error
	for _, r := range renditions {
		playlist, err := buzzGetBody(probeCtx, r.url, referer)
		if err != nil {
			lastErr = err
			continue
		}
		if adLaced2Embed(playlist) {
			lastErr = fmt.Errorf("rendition %s is ad-laced", r.url)
			continue
		}
		if err := probe2EmbedSegments(probeCtx, playlist, r.url, referer); err != nil {
			lastErr = err
			continue
		}
		return r.url, r.height, nil
	}
	return "", 0, fmt.Errorf("all %d rendition(s) unusable: %w", len(renditions), lastErr)
}

// rendition2EmbedURLs pairs every #EXT-X-STREAM-INF header with the URL
// line that follows it, resolved against the master. Manifests alternate
// header and URL on consecutive lines (The Godfather 2026-10-21:
// RESOLUTION lines then /_stream?id=… paths).
func rendition2EmbedURLs(body []byte, masterURL string) []rendition2Embed {
	lines := strings.Split(string(body), "\n")
	var out []rendition2Embed
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			continue
		}
		height := 0
		if m := resolutionRe.FindStringSubmatch(line); m != nil {
			height, _ = strconv.Atoi(m[2])
		}
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])
			if next == "" {
				continue
			}
			if strings.HasPrefix(next, "#") {
				break // header with no URL line: not a real rendition
			}
			out = append(out, rendition2Embed{url: absolutize(next, masterURL), height: height})
			i = j
			break
		}
	}
	return out
}

// buzzAdHostFragments name the CDNs observed answering segment-sized ads
// instead of video (2026-10-21: TikTok's ad CDN, p16-ad-site-sign-sg.tiktokcdn.com,
// served .image payloads as "segments" across whole renditions). A segment
// naming one of them is an ad, not media.
var buzzAdHostFragments = []string{"tiktokcdn", "ad-site-sign"}

// adLaced2Embed reports whether any video segment in a playlist points at
// an ad host. Clean playlists never do (2026-10-21: every real rendition
// paged its media through the vidsrc.buzz relay origin), and the adapter
// cannot skip embedded ads — it skips the rendition instead.
func adLaced2Embed(body []byte) bool {
	for _, seg := range playlist2EmbedLines(body) {
		if is2EmbedAdSegment(seg) {
			return true
		}
	}
	return false
}

// is2EmbedAdSegment matches the two ad signals seen in the wild: an ad-CDN
// fragment in the host, and the .image payload suffix the ad playlists used
// for their fake "video" segments.
func is2EmbedAdSegment(raw string) bool {
	host := ""
	if u, err := url.Parse(raw); err == nil {
		host = strings.ToLower(u.Host)
	}
	for _, frag := range buzzAdHostFragments {
		if host != "" && strings.Contains(host, frag) {
			return true
		}
	}
	return strings.Contains(raw, ".image?")
}

// playlist2EmbedLines lists the URL lines of an m3u8 — the ones that carry
// none of the comment tags. Comment tags and blank lines are not segments.
func playlist2EmbedLines(body []byte) []string {
	var out []string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// probe2EmbedSegments range-fetches a playlist's first segment. The 2026-10-21
// relay answered 500 on every segment while its playlists still paged 200,
// so stopping at the playlist would hand the player a stream that dies on
// its first chunk. The probe reads a small slice — enough to prove bytes
// flow, far below a real segment — and accepts both 200 and 206 (the CDN
// honours the Range or not, either way the segment serves).
func probe2EmbedSegments(ctx context.Context, body []byte, playlistURL, referer string) error {
	segments := playlist2EmbedLines(body)
	if len(segments) == 0 {
		return fmt.Errorf("the playlist lists no segments")
	}
	seg := absolutize(segments[0], playlistURL)
	segCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(segCtx, http.MethodGet, seg, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Range", "bytes=0-65535")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := buzzHTTPFn().Do(req)
	if err != nil {
		return fmt.Errorf("segment probe failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("segment probe answered %s", resp.Status)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, buzzSegmentProbeLimit))
	return nil
}

// buzzRefReport quotes a signed ref for debug output: a stable prefix, so
// logs identify a server without printing a whole token.
func buzzRefReport(ref string) string {
	if len(ref) > buzzRefReportLen {
		return ref[:buzzRefReportLen]
	}
	return ref
}

var buzzConfigMarkers = [][]byte{[]byte("var Q = "), []byte("var Q=")}

// buzzConfig extracts the `var Q = {...}` object from an embed page: a
// string-aware brace scan — the object carries quoted HTML and escaped
// quotes, so neither could a naive first-`}` cut be trusted — then JSON.
func buzzConfig(body []byte) map[string]any {
	start := -1
	for _, marker := range buzzConfigMarkers {
		if i := bytes.Index(body, marker); i >= 0 {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	brace := bytes.IndexByte(body[start:], '{') + start
	if brace < start {
		return nil
	}
	depth, inString, esc := 0, false, false
	for i := brace; i < len(body); i++ {
		c := body[i]
		if inString {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				cfg := map[string]any{}
				if jsonx.Unmarshal(body[brace:i+1], &cfg) == nil && len(cfg) > 0 {
					return cfg
				}
				return nil
			}
		}
	}
	return nil
}

// buzzGetBody GETs one URL with the browser UA a JSON Accept along the
// chain; HTML pages ignore the Accept. Bodies cap at buzzBodyLimit.
func buzzGetBody(ctx context.Context, rawURL, referer string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := buzzHTTPFn().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, buzzBodyLimit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, netx.NewHTTPStatusError(SourceName, "2embed", resp.StatusCode)
	}
	return body, nil
}

// buzzGetJSON is buzzGetBody into a JSON struct.
func buzzGetJSON(ctx context.Context, rawURL, referer string, into any) error {
	body, err := buzzGetBody(ctx, rawURL, referer)
	if err != nil {
		return err
	}
	if err := jsonx.Unmarshal(vidsrcTrimedBOM(body), into); err != nil {
		return netx.NewParserError(SourceName, "2embed", "unreadable JSON", err)
	}
	return nil
}
