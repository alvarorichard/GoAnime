package startflix

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// This file implements GoAnime's id-keyed fallback video source: when the
// StartFlix panel cannot serve a title's players (stale records pointing at
// dead file hosts — "Todo Poderoso (2003)" kept five such for months), the
// same title is resolved by IMDb/TMDB id on an embed API, reached headlessly.
//
// The chain, reverse-engineered from vidsrc.io's player JavaScript (an sbx
// wrapper page → landing /embed/{kind}/{id} → player /embed/player/{kind}/{id}):
//
//  1. GET  {base}/vs_src.php?type=movie&id=tt0315327
//     → {"src":"https://<player-host>/embed/movie/tt0315327?vs=<token>"}
//  2. GET the landing page → CFG.playerUrl, the real player page relative to
//     the landing host.
//  3. GET that player page → window.CONFIG names the stream-data API and a
//     single-use apiToken minted at the render, bound to the requesting IP.
//  4. GET streamBase(&season=&episode=)&stream_urls&api_token=…
//     → an envelope whose data.stream_urls is either a plain URL array or a
//     base64 ChaCha20 blob plus data.vs naming a per-window WASM decryptor.
//  5. Run the WASM like the browser does: alloc(len) → write ciphertext →
//     decrypt(ptr,len) → plaintext at memory[ptr+12 … ptr+12+outLen], one
//     master URL per line.
//  6. Each stream host mints a ~4h, IP-bound JWT at origin/generate.php.
//     Stamping ?token=<jwt> on the master URL plays it; the served playlist
//     inlines the token into every variant and segment URI, so mpv needs
//     nothing beyond the master URL (plus the referer already set from
//     Stream.Referer).
//
// Nothing needs a browser or a solver: plain HTTP and one WASM call.

const (
	// DefaultVIDSRCBase is the embed API fronting the fallback source.
	DefaultVIDSRCBase = "https://vidsrc.io"

	// DefaultVIDSRCMirrorBase is the vidsrc.to family's player host. It
	// speaks the same vs_src protocol and keys the same ids, but runs its
	// own frontend uptime — the fallback walks it when the primary front
	// is unreachable, so one shell going down does not take the source
	// with it (live-checked 2026-10-17: it resolves and lands on the same
	// stream hosts).
	DefaultVIDSRCMirrorBase = "https://vsembed.ru"

	// masterBodyLimit caps the master playlist read during the playable
	// probe. It is a few KiB in practice.
	masterBodyLimit = 32 << 10

	// wasmBodyLimit caps every body the chain reads, envelopes included.
	wasmBodyLimit = 16 << 20
)

// vidsrcDisabled marks the refusal when the gate keeps the fallback off.
var vidsrcDisabled = fmt.Errorf("vidsrc fallback disabled")

// vidsrcFallbackAllowed reports whether the id-keyed fallback may run. Under
// `go test` it is off unless GOANIME_VIDSRC_FALLBACK_TEST=1: regression tests
// exercise panel failures and would otherwise pull the real vidsrc.io. The
// dedicated vidsrc tests opt in with that env, same dispatch shape as the
// yt-dlp fallback's gate.
func vidsrcFallbackAllowed() bool {
	if os.Getenv("GOANIME_VIDSRC_FALLBACK_TEST") == "1" {
		return true
	}
	if os.Getenv("GOANIME_VIDSRC_FALLBACK_TEST") == "" {
		return flag.Lookup("test.v") == nil && flag.Lookup("test.run") == nil
	}
	return false
}

var vidsrcHTTPOnce sync.Once
var vidsrcHTTP *http.Client

// vidsrcHTTPDefault builds the chain's real client: one fast
// Chrome-impersonating client reused across every host, since none of the
// hops needs cookie continuity.
var vidsrcHTTPDefault = func() *http.Client {
	vidsrcHTTPOnce.Do(func() { vidsrcHTTP = util.NewFastClient() })
	return vidsrcHTTP
}

// vidsrcHTTPFn returns the HTTP client the chain runs on. A seam so tests can
// route every vidsrc host through one httptest server, exactly the way the
// real chain hops between its own hosts.
var vidsrcHTTPFn = vidsrcHTTPDefault

// vidsrcBasesFn names the embed APIs the fallback walks in order. The env
// override GOANIME_VIDSRC_URL repoints the whole list at one base (mirroring
// GOANIME_STARTFLIX_URL / GOANIME_STARTFLIX_PANEL_URL); the default walks
// the primary and its mirror. A seam, also, so tests can stage two fake
// bases and prove the walk.
var vidsrcBasesFn = func() []string {
	if b := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_VIDSRC_URL")), "/"); b != "" {
		return []string{b}
	}
	return []string{DefaultVIDSRCBase, DefaultVIDSRCMirrorBase}
}

// MovieFallback resolves a movie by its IMDb id (preferred) or TMDB id on the
// fallback source. Either alone is accepted because a panel record can carry
// only one of the two.
func MovieFallback(ctx context.Context, imdbID string, tmdbID int) (*Stream, error) {
	if !vidsrcFallbackAllowed() {
		return nil, vidsrcDisabled
	}
	if imdbID == "" && tmdbID == 0 {
		return nil, fmt.Errorf("no IMDb or TMDB id to resolve on the fallback source")
	}
	id := imdbID
	if id == "" {
		id = strconv.Itoa(tmdbID)
	}
	return resolveVIDSRC(ctx, KindMovie, id, 0, 0)
}

// EpisodeFallback resolves one series episode by its TMDB id on the fallback
// source. season/episode 0 mean "whatever the source names as Pilot".
func EpisodeFallback(ctx context.Context, tmdbID int, season, episode int) (*Stream, error) {
	if !vidsrcFallbackAllowed() {
		return nil, vidsrcDisabled
	}
	if tmdbID == 0 {
		return nil, fmt.Errorf("no TMDB id to resolve an episode on the fallback source")
	}
	return resolveVIDSRC(ctx, KindSeries, strconv.Itoa(tmdbID), season, episode)
}

// resolveVIDSRC runs the whole chain for one title, walking the fallback
// bases until one answers. Errors stay internal: the caller reports the
// original panel error to the user and Debug output records why each base
// missed, so user-facing wording still explains the panel situation rather
// than a source they never chose.
func resolveVIDSRC(ctx context.Context, kind MediaKind, id string, season, episode int) (*Stream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var lastErr error
	for _, base := range vidsrcBasesFn() {
		stream, err := resolveVIDSRCOn(ctx, base, kind, id, season, episode)
		if err == nil {
			return stream, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		util.Debug("VIDSRC fallback base missed", "base", base, "err", err)
	}
	return nil, lastErr
}

// resolveVIDSRCOn runs the chain on one embed API base.
func resolveVIDSRCOn(ctx context.Context, base string, kind MediaKind, id string, season, episode int) (*Stream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// (1) Which player host carries this title today.
	var res struct {
		Src string `json:"src"`
	}
	vsURL := fmt.Sprintf("%s/vs_src.php?type=%s&id=%s", base, kind, url.QueryEscape(id))
	if err := vidsrcGetJSON(ctx, vsURL, "", &res); err != nil {
		return nil, err
	}
	landing := strings.TrimSpace(res.Src)
	if landing == "" {
		return nil, netx.NewParserError(SourceName, "vidsrc-src", "no player page named for "+string(kind)+" "+id, nil)
	}

	// (2) Landing page: CFG.playerUrl points at the real player page.
	landingBody, err := vidsrcGetBody(ctx, landing, base+"/")
	if err != nil {
		return nil, err
	}
	playerRef := windowJSONField(landingBody, "window.CFG", "playerUrl")
	if playerRef == "" {
		return nil, netx.NewParserError(SourceName, "vidsrc-landing", "no playerUrl in the landing page", nil)
	}
	playerURL := absolutize(playerRef, landing)

	// (3) Player page: the stream-data API and the single-use token.
	playerBody, err := vidsrcGetBody(ctx, playerURL, landing)
	if err != nil {
		return nil, err
	}
	config := windowJSON(playerBody, "window.CONFIG")
	if config == nil {
		return nil, netx.NewParserError(SourceName, "vidsrc-player", "no CONFIG in the player page", nil)
	}
	streamBase := strField(config, "streamBase")
	api := strField(config, "api")
	dataBase := streamBase
	if dataBase == "" {
		dataBase = api
	}
	token := strField(config, "apiToken")
	if dataBase == "" {
		return nil, netx.NewParserError(SourceName, "vidsrc-player", "no stream API in the player page", nil)
	}

	// (4) Series carry season/episode, and when the caller named none, the
	// source's own defaults (its eps map names data.season and data.episode).
	series := kind == KindSeries || streamBase != ""
	if series && (season <= 0 || episode <= 0) {
		var eps struct {
			Data struct {
				Season  int    `json:"season"`
				Episode string `json:"episode"`
			} `json:"data"`
		}
		if err := vidsrcGetJSON(ctx, dataBase, playerURL, &eps); err == nil {
			if season <= 0 {
				season = eps.Data.Season
			}
			if episode <= 0 && eps.Data.Episode != "" {
				episode, _ = strconv.Atoi(eps.Data.Episode)
			}
		}
	}

	urls, _, err := fetchVIDSRCEnvelope(ctx, playerURL, dataBase, series, season, episode, token)
	if err != nil {
		return nil, err
	}

	// (5) Token + probe each master on the host it came from. The first
	// playable one wins; the list usually carries two or three rotating
	// copies of the same encode.
	referer := masterReferer(playerURL)
	for _, raw := range urls {
		master, host, err := resolveVIDSRCMaster(ctx, raw, referer)
		if err != nil {
			util.Debug("VIDSRC fallback master missed", "host", host, "err", err)
			continue
		}
		return master, nil
	}
	return nil, fmt.Errorf("vidsrc fallback: %d URL(s) listed, none playable", len(urls))
}

// fetchVIDSRCEnvelope calls the stream-data endpoint, decrypting the
// ChaCha20 envelope through the WASM when the URLs arrive protected. Returns
// the master URLs and the data URL used (for caller diagnostics).
func fetchVIDSRCEnvelope(ctx context.Context, playerURL, dataBase string, series bool, season, episode int, token string) ([]string, string, error) {
	// The single-use token rides the query; season/episode ride alongside.
	withSeason := series && (season > 0 || episode > 0)

	fetch := func(tok string) ([]byte, string, error) {
		u, err := url.Parse(dataBase)
		if err != nil {
			return nil, dataBase, err
		}
		q := u.Query()
		if withSeason {
			if season > 0 {
				q.Set("season", strconv.Itoa(season))
			}
			if episode > 0 {
				q.Set("episode", strconv.Itoa(episode))
			}
		}
		q.Set("stream_urls", "")
		q.Set("api_token", tok)
		u.RawQuery = q.Encode()
		dataURL := u.String()
		// A 403 rides a JSON envelope ({"status_code":"403"...}); the caller
		// reads it to run the token refresh.
		body, err := vidsrcGetBody(ctx, dataURL, playerURL, http.StatusForbidden)
		return body, dataURL, err
	}

	raw, dataURL, err := fetch(token)
	if err != nil {
		return nil, dataURL, err
	}

	var env struct {
		StatusCode any `json:"status_code"`
		Data       struct {
			StreamURLs jsonAny `json:"stream_urls"`
		} `json:"data"`
		VS struct {
			W       int    `json:"w"`
			WasmURL string `json:"wasm_url"`
			Wasm    string `json:"wasm"`
		} `json:"vs"`
	}
	if err := jsonx.Unmarshal(vidsrcTrimedBOM(raw), &env); err != nil {
		return nil, dataURL, netx.NewParserError(SourceName, "vidsrc-data", "unreadable stream envelope", err)
	}
	if code := statusOf(env.StatusCode); code != 0 && code != http.StatusOK {
		// The token is single-use: a stale one answers "invalid api token"
		// with 403. The player recovers by re-rendering the page for a fresh
		// token; mirror that once.
		if code == http.StatusForbidden && token != "" {
			if fresh := vidsrcRefreshToken(ctx, playerURL); fresh != "" && fresh != token {
				return fetchVIDSRCEnvelope(ctx, playerURL, dataBase, series, season, episode, fresh)
			}
		}
		return nil, dataURL, fmt.Errorf("vidsrc fallback: stream API answered %d for this title", code)
	}

	// Protection off: URLs as an array.
	if plain := env.Data.StreamURLs.asArray(); plain != nil {
		return cleanStreams(plain), dataURL, nil
	}

	// Protection on: one encrypted string plus the per-window WASM; absent
	// those, treat the string as the source's plain newline list.
	enc := env.Data.StreamURLs.asString()
	switch {
	case enc == "":
		return nil, dataURL, fmt.Errorf("vidsrc fallback: envelope carried no stream_urls")
	case env.VS.WasmURL != "" || env.VS.Wasm != "":
		wasmBytes, err := vidsrcWASM(ctx, env.VS.WasmURL, env.VS.Wasm, dataBase)
		if err != nil {
			return nil, dataURL, err
		}
		plain, err := decryptVIDSRC(wasmBytes, enc)
		if err != nil {
			return nil, dataURL, err
		}
		return plain, dataURL, nil
	default:
		return cleanStreams(strings.Split(enc, "\n")), dataURL, nil
	}
}

// statusOf reads status_code, which arrives as number or string ("200").
func statusOf(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(n))
		return i
	default:
		return 0
	}
}

// vidsrcRefreshToken re-renders the player page for a fresh single-use token.
func vidsrcRefreshToken(ctx context.Context, playerURL string) string {
	body, err := vidsrcGetBody(ctx, playerURL, "")
	if err != nil {
		return ""
	}
	config := windowJSON(body, "window.CONFIG")
	if config == nil {
		return ""
	}
	return strField(config, "apiToken")
}

// vidsrcWASM materialises the decryptor from the inline base64 blob or URL.
func vidsrcWASM(ctx context.Context, wasmURL, wasmB64, dataBase string) ([]byte, error) {
	switch {
	case wasmB64 != "":
		raw, err := base64.StdEncoding.DecodeString(wasmB64)
		if err != nil {
			return nil, fmt.Errorf("vidsrc fallback: inline decryptor unreadable: %w", err)
		}
		return raw, nil
	case wasmURL == "":
		return nil, fmt.Errorf("vidsrc fallback: envelope carried no decryptor")
	default:
		raw, err := vidsrcGetBody(ctx, absolutize(wasmURL, dataBase), "")
		if err != nil {
			return nil, fmt.Errorf("vidsrc fallback: decryptor download failed: %w", err)
		}
		return raw, nil
	}
}

// resolveVIDSRCMaster stamps the host token onto one master URL and verifies
// it plays, so the user never lands on a dead playlist.
func resolveVIDSRCMaster(ctx context.Context, raw, referer string) (*Stream, string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, raw, fmt.Errorf("malformed stream URL")
	}
	origin := u.Scheme + "://" + u.Host
	tok, err := vidsrcHostToken(ctx, origin, referer)
	if err != nil {
		return nil, u.Host, err
	}
	master := stampToken(u.String(), tok)
	served, height := probeVIDSRCMaster(ctx, master, referer)
	if height <= 0 || len(served) == 0 {
		return nil, u.Host, fmt.Errorf("master did not load")
	}
	return &Stream{
		URL:     master,
		Referer: referer,
		Host:    u.Host,
		Height:  height,
	}, u.Host, nil
}

// vidsrcHostToken fetches the per-host JWT (origin/generate.php), accepting
// both the raw token and a JSON wrapper.
// vidsrcRatePause is how long the host-token fetch waits out a 429 before
// its single retry. A var so tests do not sleep real seconds.
var vidsrcRatePause = 5 * time.Second

// vidsrcHostTokenOnce performs one token fetch attempt (see vidsrcHostToken).
func vidsrcHostTokenOnce(ctx context.Context, origin, referer string) ([]byte, error) {
	return vidsrcGetBody(ctx, origin+"/generate.php", referer)
}

func vidsrcHostToken(ctx context.Context, origin, referer string) (string, error) {
	raw, err := vidsrcHostTokenOnce(ctx, origin, referer)
	if err != nil {
		// The token endpoint rate-limits bursts: playing twice in quick
		// succession rates out on the second run, and the IP-bound limit
		// clears in seconds. One short, ctx-bounded retry turns that from
		// "the title will not resolve" into a brief pause once.
		var diag *netx.SourceDiagnostic
		if errors.Is(err, context.Canceled) || ctx.Err() != nil ||
			!errors.As(err, &diag) || diag.Kind != netx.DiagnosticBlockedChallenge || diag.StatusCode != http.StatusTooManyRequests {
			return "", fmt.Errorf("host token fetch failed: %w", err)
		}
		util.Debug("VIDSRC token endpoint rate-limited, pausing once", "origin", origin)
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("host token fetch failed: %w", ctx.Err())
		case <-time.After(vidsrcRatePause):
		}
		raw, err = vidsrcHostTokenOnce(ctx, origin, referer)
		if err != nil {
			return "", fmt.Errorf("host token fetch still refused after a rate-limit pause: %w", err)
		}
	}
	tok := strings.TrimSpace(string(raw))
	if strings.HasPrefix(tok, "{") {
		var wrapped struct {
			Token string `json:"token"`
			Data  string `json:"data"`
		}
		if jsonx.Unmarshal(raw, &wrapped) == nil {
			switch {
			case wrapped.Token != "":
				tok = wrapped.Token
			case wrapped.Data != "":
				tok = wrapped.Data
			}
		}
	}
	if tok == "" {
		return "", fmt.Errorf("host answered without a token")
	}
	return tok, nil
}

// stampToken appends ?token=<jwt> (or &token= when a query exists).
func stampToken(rawURL, tok string) string {
	if tok == "" {
		return rawURL
	}
	sep := "?"
	if strings.Contains(rawURL, "?") {
		sep = "&"
	}
	return rawURL + sep + "token=" + url.QueryEscape(tok)
}

// probeVIDSRCMaster checks the master loads and reads its best resolution.
// Returns (nil, 0) when it does not; the caller tries the next URL.
func probeVIDSRCMaster(ctx context.Context, masterURL, referer string) ([]byte, int) {
	proctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(proctx, http.MethodGet, masterURL, http.NoBody)
	if err != nil {
		return nil, 0
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := vidsrcHTTPFn().Do(req)
	if err != nil {
		return nil, 0
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, 0
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, masterBodyLimit))
	if err != nil {
		return nil, 0
	}
	if !bytes.Contains(body, []byte("#EXTM3U")) {
		return nil, 0
	}
	return body, resolutionHeight(body)
}

// resolutionHeight reads the tallest RESOLUTION=WxH in a master playlist.
func resolutionHeight(body []byte) int {
	best := 0
	for _, m := range resolutionRe.FindAllSubmatch(body, -1) {
		if h, _ := strconv.Atoi(string(m[2])); h > best {
			best = h
		}
	}
	return best
}

var resolutionRe = regexp.MustCompile(`RESOLUTION=(\d+)x(\d+)`)

// masterReferer prefers the player-page origin: what the real player would
// send (hls.js runs inside the player page). The panel uses referrerpolicy
// "origin" for its cross-host pieces, so origin-level is the safe form.
func masterReferer(playerURL string) string {
	if o := originOf(playerURL); o != "" {
		return o + "/"
	}
	return ""
}

// vidsrcGetBody GETs one URL with a browser UA. JSON Accept is set along the
// chain; HTML pages ignore it. allow lists extra statuses whose body the
// caller still wants to read — the stream endpoint answers token problems
// with a 403 that carries a JSON envelope.
func vidsrcGetBody(ctx context.Context, rawURL, referer string, allow ...int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := vidsrcHTTPFn().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, wasmBodyLimit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 && !slices.Contains(allow, resp.StatusCode) {
		return nil, netx.NewHTTPStatusError(SourceName, "vidsrc", resp.StatusCode)
	}
	return body, nil
}

// vidsrcGetJSON is vidsrcGetBody into a JSON struct.
func vidsrcGetJSON(ctx context.Context, rawURL, referer string, into any) error {
	body, err := vidsrcGetBody(ctx, rawURL, referer)
	if err != nil {
		return err
	}
	if err := jsonx.Unmarshal(vidsrcTrimedBOM(body), into); err != nil {
		return netx.NewParserError(SourceName, "vidsrc", "unreadable JSON", err)
	}
	return nil
}

// vidsrcTrimedBOM strips a UTF-8 BOM should one ever show up.
func vidsrcTrimedBOM(b []byte) []byte {
	return []byte(strings.TrimPrefix(strings.TrimPrefix(string(b), "\ufeff"), "\xff\xfe"))
}

// windowJSON extracts the `<script>window.CFG/CONFIG = {...}</script>` object
// from a page body. Both objects are flat, so a brace scan is enough.
func windowJSON(body []byte, windowVar string) map[string]any {
	start := strings.Index(string(body), windowVar)
	if start < 0 {
		return nil
	}
	brace := strings.IndexByte(string(body[start:]), '{') + start
	if brace < start {
		return nil
	}
	end := strings.IndexByte(string(body[brace:]), '}') + brace
	s := string(body[brace : end+1])
	cfg := map[string]any{}
	if jsonx.Unmarshal([]byte(s), &cfg) != nil {
		return nil
	}
	return cfg
}

// windowJSONField is windowJSON + one string field.
func windowJSONField(body []byte, windowVar, field string) string {
	cfg := windowJSON(body, windowVar)
	if cfg == nil {
		return ""
	}
	return strField(cfg, field)
}

// strField reads one field from a decoded JSON object.
func strField(m map[string]any, key string) string {
	switch v := m[key].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

// absolutize resolves a playerUrl/wasm_url against the page that named it.
func absolutize(ref, pageURL string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return ref
	}
	return base.ResolveReference(u).String()
}

// originOf is scheme://host[:port] of a URL, or "" when it cannot be parsed.
func originOf(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// cleanStreams trims and drops empties from a URL list.
func cleanStreams(urls []string) []string {
	out := make([]string, 0, len(urls))
	for _, s := range urls {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
