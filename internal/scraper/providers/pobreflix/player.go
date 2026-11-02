package pobreflix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pobreflix's own player moved behind a Cloudflare Turnstile in October 2026.
// The site lists one "Player Principal" per audio, usually /api/player?id=<tmdb>
// (which may answer 404 "Vídeo indisponível" while the catalog is being moved)
// or a streambetter page such as https://streambetter.lat/filme/<tmdb>.
//
// The page renders a standalone Turnstile widget. GoAnime solves it with the
// optional netx.TurnstileSolver (the cfsolve tool, when installed), posts the
// token to /api/turnstile-verify, and reloads the page with the proof in ?tsv=.
// That reload carries window.__PLAYER_PARAMS__, a query string whose "sources"
// entry names the real media URL. The media endpoint requires the proof-bearing
// page as Referer, which is why Stream.Referer exists.
//
// The requests here use Client.playerTransport (a plain, SSRF-safe transport)
// instead of the shared fast client: the player worker answers the shared
// Chrome TLS impersonation — with identical Chrome headers — by a bare HTTP 500,
// while the same request on a plain Go transport answers 200 (measured
// 2026-11-01). The catalog keeps the shared fingerprint.
//
// Without a solver this file is never reached: the api layer keeps its previous
// panel fallback and explanation.

const (
	// DefaultPlayerBase is where the site's player pages live. The catalog
	// moved from .lol to .lat in October 2026; GOANIME_POBREFLIX_PLAYER_URL
	// repoints it without a release.
	DefaultPlayerBase = "https://streambetter.lat"

	// playerSolveTimeout bounds one cfsolve run. A live solve takes a few
	// seconds; the room is for a slow launch or a retry inside cfsolve.
	playerSolveTimeout = 45 * time.Second

	// playerHTTPTimeout keeps a slow page fetch from inheriting the shared
	// fast client's eight-second API budget.
	playerHTTPTimeout = 20 * time.Second

	turnstileVerifyPath = "/api/turnstile-verify"
)

// Player resolution failures. ErrPlayerSolverMissing is distinct from "this
// page is broken": it means the caller should take its normal no-solver path.
var (
	ErrPlayerSolverMissing = errors.New("pobreflix: the player needs a Turnstile solver, and none is registered")
	ErrPlayerNoTurnstile   = errors.New("pobreflix: the player page carries no Turnstile widget")
	ErrPlayerNoStream      = errors.New("pobreflix: the player page names no playable stream")
)

// PlayerTarget identifies the streambetter page by the ids the catalog carries.
// Kind selects /filme/<tmdb> from /serie/<tmdb>/<season>/<episode>.
type PlayerTarget struct {
	TMDBID  int
	Kind    MediaKind
	Season  int
	Episode int
}

// Stream is one playable file found behind the player gate. Referer is the
// proof-bearing player page the media endpoint requires.
type Stream struct {
	URL     string
	Referer string
	Audio   Audio
}

// playerBaseFromEnv reads the optional player-host override. Tests point it at
// their own server; production normally uses DefaultPlayerBase.
func playerBaseFromEnv() string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_POBREFLIX_PLAYER_URL")), "/")
	if base == "" {
		return DefaultPlayerBase
	}
	return base
}

// playerURL builds the player page for target. It returns "" when the target
// carries no usable id (a series without its slot, say).
func (c *Client) playerURL(target PlayerTarget) string {
	if target.TMDBID <= 0 {
		return ""
	}
	base := strings.TrimRight(c.playerBase, "/")
	if base == "" {
		base = DefaultPlayerBase
	}
	kind := target.Kind
	if kind == "" {
		if target.Season > 0 || target.Episode > 0 {
			kind = KindSeries
		} else {
			kind = KindMovie
		}
	}
	if kind == KindSeries {
		if target.Season <= 0 || target.Episode <= 0 {
			return ""
		}
		return fmt.Sprintf("%s/serie/%d/%d/%d", base, target.TMDBID, target.Season, target.Episode)
	}
	return fmt.Sprintf("%s/filme/%d", base, target.TMDBID)
}

// ResolvePlayer clears the Turnstile on serverURL's page and returns the media
// URL behind it, preferring the source whose label matches want.
//
// It performs no network I/O when no solver is registered: callers rely on
// that to keep the old no-solver behavior and its request counts.
func (c *Client) ResolvePlayer(ctx context.Context, serverURL string, target PlayerTarget, want Audio) (Stream, error) {
	if !netx.TurnstileSolverAvailable() {
		return Stream{}, ErrPlayerSolverMissing
	}
	if target.TMDBID <= 0 {
		return Stream{}, fmt.Errorf("pobreflix: player target has no TMDB id")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return Stream{}, err
	}
	// The resolver runs on its own client: it keeps the cookie jar the widget
	// was fetched with, and it must not inherit the shared fast client's
	// eight-second API budget.
	hc := &http.Client{
		Jar:       jar,
		Timeout:   playerHTTPTimeout,
		Transport: c.playerTransport, // production; nil in tests
	}

	listed := strings.TrimSpace(serverURL)
	pageURL := listed
	if pageURL == "" {
		pageURL = c.playerURL(target)
	}
	if pageURL == "" {
		return Stream{}, fmt.Errorf("pobreflix: no player page for TMDB %d", target.TMDBID)
	}

	body, finalURL, status, err := c.fetchPlayerPage(ctx, hc, pageURL, "")
	if err != nil && !isCatalogPlayerURL(listed) {
		return Stream{}, fmt.Errorf("pobreflix: player page: %w", err)
	}
	if isCatalogPlayerURL(listed) && (err != nil || playerNeedsFallback(body, status)) {
		fallback := c.playerURL(target)
		if fallback != "" && fallback != pageURL {
			util.Debug("Pobreflix player endpoint did not answer; using the player host directly",
				"listed", redactURL(pageURL), "fallback", redactURL(fallback))
			body, finalURL, status, err = c.fetchPlayerPage(ctx, hc, fallback, "")
		}
	}
	if err != nil {
		return Stream{}, fmt.Errorf("pobreflix: player page: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return Stream{}, fmt.Errorf("pobreflix: player page answered HTTP %d", status)
	}
	if params, ok := parsePlayerParams(body); ok {
		return c.streamFromParams(params, finalURL, want)
	}

	siteKey, action, cdata, ok := parseTurnstile(body)
	if !ok {
		return Stream{}, ErrPlayerNoTurnstile
	}
	util.Debug("Pobreflix: answering the player's Turnstile", "player", hostOfURL(finalURL), "audio", want)

	token, solverOK, err := netx.SolveTurnstileFor(ctx, netx.TurnstileRequest{
		PageURL: finalURL,
		SiteKey: siteKey,
		Action:  action,
		CData:   cdata,
	}, playerSolveTimeout)
	if !solverOK {
		return Stream{}, ErrPlayerSolverMissing
	}
	if err != nil {
		return Stream{}, fmt.Errorf("pobreflix: Turnstile solve failed: %w", err)
	}

	proof, err := c.verifyTurnstile(ctx, hc, finalURL, token)
	if err != nil {
		return Stream{}, err
	}
	solvedURL := setQueryParam(finalURL, "tsv", proof)
	body, finalURL, status, err = c.fetchPlayerPage(ctx, hc, solvedURL, finalURL)
	if err != nil {
		return Stream{}, fmt.Errorf("pobreflix: solved player page: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return Stream{}, fmt.Errorf("pobreflix: solved player page answered HTTP %d", status)
	}
	params, ok := parsePlayerParams(body)
	if !ok {
		return Stream{}, fmt.Errorf("%w after the Turnstile was answered", ErrPlayerNoStream)
	}
	return c.streamFromParams(params, finalURL, want)
}

// fetchPlayerPage reads one page with the resolver's own cookie jar, keeping
// the final URL a redirect landed on: the Turnstile proof and the media Referer
// belong to that address.
func (c *Client) fetchPlayerPage(ctx context.Context, hc *http.Client, rawURL, referer string) (body []byte, finalURL string, status int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, rawURL, 0, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, rawURL, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, rawURL, resp.StatusCode, err
	}
	finalURL = rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return body, finalURL, resp.StatusCode, nil
}

// verifyTurnstile exchanges the widget token for the proof the page expects in
// ?tsv=. It uses the same client (and so the same cookie jar) that loaded the
// widget.
func (c *Client) verifyTurnstile(ctx context.Context, hc *http.Client, pageURL, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", errors.New("pobreflix: the Turnstile solver returned an empty token")
	}
	u, err := url.Parse(pageURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("pobreflix: cannot verify a Turnstile for %q", pageURL)
	}
	verifyURL := (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: turnstileVerifyPath}).String()
	payload, err := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: token})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", u.Scheme+"://"+u.Host)
	req.Header.Set("Referer", pageURL)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("pobreflix: Turnstile verification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("pobreflix: Turnstile verification: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("pobreflix: Turnstile verification answered HTTP %d", resp.StatusCode)
	}
	var answer struct {
		Success bool   `json:"success"`
		Proof   string `json:"proof"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &answer); err != nil {
		return "", netx.NewParserError(SourceName, "player", "unreadable Turnstile answer", err)
	}
	proof := strings.TrimSpace(answer.Proof)
	if !answer.Success || proof == "" {
		reason := strings.TrimSpace(answer.Error)
		if reason == "" {
			reason = strings.TrimSpace(answer.Message)
		}
		if reason == "" {
			reason = "no proof was returned"
		}
		return "", fmt.Errorf("pobreflix: the player refused the Turnstile answer: %s", reason)
	}
	return proof, nil
}

// playerSource is one entry of __PLAYER_PARAMS__'s "sources" JSON. The live
// page (2026-11-01) carries kind/source id/audio label/url/vttUrl.
type playerSource struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// playerParamsRe captures the JS string assigned to window.__PLAYER_PARAMS__.
var playerParamsRe = regexp.MustCompile(`(?s)window\.__PLAYER_PARAMS__\s*=\s*("(?:[^"\\]|\\.)*")`)

// parsePlayerParams reads the query string the player page embeds. A page that
// already carries it is past the gate (a reused proof or a cookie), so no solve
// is needed.
func parsePlayerParams(body []byte) (url.Values, bool) {
	m := playerParamsRe.FindSubmatch(body)
	if len(m) < 2 {
		return nil, false
	}
	raw, err := strconv.Unquote(string(m[1]))
	if err != nil {
		return nil, false
	}
	params, err := url.ParseQuery(raw)
	if err != nil {
		return nil, false
	}
	if params.Get("sources") == "" && params.Get("url") == "" {
		return nil, false
	}
	return params, true
}

// streamFromParams picks the source matching want and resolves its URL against
// the proof-bearing page.
func (c *Client) streamFromParams(params url.Values, pageURL string, want Audio) (Stream, error) {
	var sources []playerSource
	_ = json.Unmarshal([]byte(params.Get("sources")), &sources)
	source, ok := pickPlayerSource(sources, want)
	if !ok {
		raw := strings.TrimSpace(params.Get("url"))
		if raw == "" {
			return Stream{}, ErrPlayerNoStream
		}
		source = playerSource{Kind: "stream", URL: raw}
	}
	resolved := resolveAgainst(pageURL, strings.TrimSpace(source.URL))
	if resolved == "" {
		return Stream{}, ErrPlayerNoStream
	}
	audio := playerAudio(source.Label)
	if audio == AudioUnknown && want != AudioUnknown {
		// A page fetched for the wanted audio whose source has no label is
		// taken as that audio, like a server entry with no audio.
		audio = want
	}
	return Stream{URL: resolved, Referer: pageURL, Audio: audio}, nil
}

// pickPlayerSource prefers a playable source labelled with want, then the first
// playable source (a single-audio page), and refuses embeds.
func pickPlayerSource(sources []playerSource, want Audio) (playerSource, bool) {
	var first *playerSource
	for i := range sources {
		s := sources[i]
		if !playablePlayerSource(s) {
			continue
		}
		if first == nil {
			copied := s
			first = &copied
		}
		if want != AudioUnknown && playerAudio(s.Label) == want {
			return s, true
		}
	}
	if first != nil {
		return *first, true
	}
	return playerSource{}, false
}

// playablePlayerSource rejects iframe/embed entries: mpv cannot play their
// HTML, and that is the exact failure this path exists to avoid.
func playablePlayerSource(s playerSource) bool {
	raw := strings.TrimSpace(s.URL)
	if raw == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(s.Kind)) {
	case "iframe", "embed", "html":
		return false
	}
	if strings.Contains(raw, "/api/proxy?") || strings.Contains(raw, "/api/extract?") {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch strings.ToLower(path.Ext(u.Path)) {
	case ".mp4", ".m4v", ".mkv", ".webm", ".mov", ".m3u8", ".mpd":
		return true
	}
	return strings.EqualFold(strings.TrimSpace(s.Kind), "stream")
}

// playerAudio maps a source label such as "Dublado" or "Legendado (HD)" to the
// package's audio labels.
func playerAudio(label string) Audio {
	lower := strings.ToLower(strings.TrimSpace(label))
	switch {
	case strings.Contains(lower, "dubl"):
		return AudioDubbed
	case strings.Contains(lower, "legend"):
		return AudioSubtitled
	}
	return audioFromEmbedded(label)
}

// parseTurnstile finds the widget's sitekey (and optional action/cdata). The
// live page is server-rendered; the script-rendered form is accepted too so a
// player that starts calling turnstile.render() does not silently lose the
// solver.
func parseTurnstile(body []byte) (siteKey, action, cdata string, ok bool) {
	if doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body)); err == nil {
		widget := doc.Find(".cf-turnstile").First()
		if widget.Length() > 0 {
			siteKey = strings.TrimSpace(widget.AttrOr("data-sitekey", ""))
			action = strings.TrimSpace(widget.AttrOr("data-action", ""))
			cdata = strings.TrimSpace(widget.AttrOr("data-cdata", ""))
		}
	}
	if siteKey == "" {
		if m := turnstileSiteKeyRe.FindSubmatch(body); len(m) > 1 {
			siteKey = string(m[1])
		}
	}
	if action == "" {
		if m := turnstileActionRe.FindSubmatch(body); len(m) > 1 {
			action = string(m[1])
		}
	}
	if cdata == "" {
		if m := turnstileCDataRe.FindSubmatch(body); len(m) > 1 {
			cdata = string(m[1])
		}
	}
	return siteKey, action, cdata, strings.TrimSpace(siteKey) != ""
}

var (
	turnstileSiteKeyRe = regexp.MustCompile(`(?i)(?:data-sitekey\s*=\s*["']|["']?sitekey["']?\s*[:=]\s*["'])([^"']+)["']`)
	turnstileActionRe  = regexp.MustCompile(`(?i)(?:data-action\s*=\s*["']|["']?action["']?\s*[:=]\s*["'])([^"']+)["']`)
	turnstileCDataRe   = regexp.MustCompile(`(?i)(?:data-cdata\s*=\s*["']|["']?cdata["']?\s*[:=]\s*["'])([^"']+)["']`)
)

// playerNeedsFallback says a response is neither the gate nor a solved page, so
// a listed /api/player endpoint can be replaced with the player host.
func playerNeedsFallback(body []byte, status int) bool {
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		if _, ok := parsePlayerParams(body); ok {
			return false
		}
		if _, _, _, ok := parseTurnstile(body); ok {
			return false
		}
	}
	return true
}

// isCatalogPlayerURL recognizes the catalog's own player endpoint. It is the
// one URL whose failure justifies rebuilding the page from the TMDB id.
func isCatalogPlayerURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	p := u.Path
	return p == "/api/player" || strings.HasPrefix(p, "/api/player/")
}

// setQueryParam returns rawURL with one query parameter replaced.
func setQueryParam(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

// resolveAgainst resolves a player-relative URL against its page.
func resolveAgainst(base, raw string) string {
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	r, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return b.ResolveReference(r).String()
}

// redactURL keeps a player URL out of debug logs without its proof/token query.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.RawQuery = ""
	return u.String()
}
