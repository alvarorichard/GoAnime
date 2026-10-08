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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// DefaultBase is the catalog host. The site also answers on mirrors
	// (startflix.live and startflix.vip are linked from every page), so
	// GOANIME_STARTFLIX_URL can repoint it without waiting for a release.
	DefaultBase = "https://www.startflix.biz"

	// userAgent describes a current Chrome, matching the TLS fingerprint
	// util.NewFastClient presents. Neither the site, the panel nor the Byse CDN
	// checked it on 2026-10-07; it is set so the request looks like one browser
	// rather than a Chrome handshake carrying some other product's name.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"

	maxBodyBytes = 5 << 20
)

// Client talks to StartFlix, its video panel and the player hosts.
type Client struct {
	http       *http.Client
	stream     *http.Client // upstream client of the local stream proxy; no overall timeout
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration

	panels  sync.Map // title page URL -> Panel
	seasons sync.Map // panel URL -> []Season
}

// NewClient builds a client. It performs no network I/O.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_STARTFLIX_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	return &Client{
		http:       util.NewFastClient(),
		stream:     newStreamClient(),
		baseURL:    base,
		userAgent:  userAgent,
		maxRetries: 2,
		retryDelay: 400 * time.Millisecond,
	}
}

// NewClientWithHTTP builds a client on a caller-supplied HTTP client, for
// callers that route requests through their own transport (the metadata
// enricher does, so its tests can serve these pages from a mock).
func NewClientWithHTTP(hc *http.Client) *Client {
	c := NewClient()
	c.http = hc
	return c
}

// NewClientForTest points a client at a test server.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	return &Client{
		http:       hc,
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  userAgent,
		maxRetries: 0,
		retryDelay: 0,
	}
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client, so the panel and season caches are
// shared between the search adapter and the playback path.
func Shared() *Client {
	sharedOnce.Do(func() { sharedClient = NewClient() })
	return sharedClient
}

// BaseURL is the catalog host this client searches.
func (c *Client) BaseURL() string { return c.baseURL }

type getOptions struct {
	referer string
	accept  string
	ajax    bool
	layer   string
	// allow lists non-2xx statuses whose body the caller wants to read itself
	// (the panel's "not found" page, a Byse error record).
	allow []int
}

func (c *Client) get(ctx context.Context, rawURL string, opt getOptions) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			}
		}
		body, retry, err := c.getOnce(ctx, rawURL, opt)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry || ctx.Err() != nil {
			break
		}
		util.Debug("StartFlix request failed, retrying", "url", rawURL, "attempt", attempt+1, "err", err)
	}
	return nil, lastErr
}

func (c *Client) getOnce(ctx context.Context, rawURL string, opt getOptions) (body []byte, retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	if opt.accept != "" {
		req.Header.Set("Accept", opt.accept)
	} else {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	}
	if opt.referer != "" {
		req.Header.Set("Referer", opt.referer)
	}
	if opt.ajax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ctx.Err() == nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, true, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 || slices.Contains(opt.allow, resp.StatusCode) {
		return body, false, nil
	}
	layer := opt.layer
	if layer == "" {
		layer = "http"
	}
	return nil, resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests,
		netx.NewHTTPStatusError(SourceName, layer, resp.StatusCode)
}

func (c *Client) getDocument(ctx context.Context, rawURL string, opt getOptions) (*goquery.Document, []byte, error) {
	body, err := c.get(ctx, rawURL, opt)
	if err != nil {
		return nil, nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, nil, netx.NewParserError(SourceName, opt.layer, "unreadable HTML", err)
	}
	return doc, body, nil
}

// Search queries the catalog. Hyphens and underscores become spaces: CLI
// arguments arrive slugged ("the-boys") and WordPress matches words.
func (c *Client) Search(ctx context.Context, query string) ([]Media, error) {
	q := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(query)), " ")
	if q == "" {
		return nil, nil
	}
	doc, _, err := c.getDocument(ctx, c.baseURL+"/?s="+url.QueryEscape(q), getOptions{
		referer: c.baseURL + "/",
		layer:   "search",
	})
	if err != nil {
		return nil, err
	}
	results := parseSearchResults(doc)
	util.Debug("StartFlix search", "query", q, "results", len(results))
	return results, nil
}

// Panel finds the video panel a title page embeds. Cached per title page.
func (c *Client) Panel(ctx context.Context, pageURL string) (Panel, error) {
	if p, ok := c.panels.Load(pageURL); ok {
		return p.(Panel), nil
	}
	doc, _, err := c.getDocument(ctx, pageURL, getOptions{referer: c.baseURL + "/", layer: "title"})
	if err != nil {
		return Panel{}, err
	}
	panel, ok := parsePanel(doc)
	if !ok {
		return Panel{}, ErrNoPanel
	}
	c.panels.Store(pageURL, panel)
	util.Debug("StartFlix panel", "page", pageURL, "panel", panel.URL)
	return panel, nil
}

// Seasons lists a series panel's seasons. Cached per panel.
func (c *Client) Seasons(ctx context.Context, panel Panel) ([]Season, error) {
	if panel.Kind != KindSeries {
		return nil, fmt.Errorf("startflix: %s is not a series panel", panel.URL)
	}
	if s, ok := c.seasons.Load(panel.URL); ok {
		return s.([]Season), nil
	}
	doc, body, err := c.getDocument(ctx, panel.URL, getOptions{
		referer: c.baseURL + "/",
		layer:   "panel",
		allow:   []int{http.StatusNotFound},
	})
	if err != nil {
		return nil, err
	}
	if isPanelNotFound(body) {
		return nil, ErrNotOnPanel
	}
	seasons := parseSeasons(doc)
	if len(seasons) == 0 {
		return nil, netx.NewParserError(SourceName, "panel", "no seasons in the series panel", nil)
	}
	c.seasons.Store(panel.URL, seasons)
	return seasons, nil
}

// Players lists the player buttons on an episode endpoint or a movie panel.
func (c *Client) Players(ctx context.Context, playersURL string) ([]Player, error) {
	u, err := url.Parse(playersURL)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("startflix: invalid players URL %q", playersURL)
	}
	origin := u.Scheme + "://" + u.Host
	opt := getOptions{layer: "players", allow: []int{http.StatusNotFound}}
	if strings.HasPrefix(u.Path, "/episodio/") {
		// The panel fetches episodes with jQuery's $.get from its own page.
		opt.referer = origin + "/"
		opt.ajax = true
	} else {
		opt.referer = c.baseURL + "/"
	}
	doc, body, err := c.getDocument(ctx, playersURL, opt)
	if err != nil {
		return nil, err
	}
	if isPanelNotFound(body) {
		return nil, ErrNotOnPanel
	}
	players := parsePlayers(doc)
	for i := range players {
		players[i].referer = origin + "/"
	}
	return players, nil
}

// Stream resolves the first playable server listed at playersURL.
func (c *Client) Stream(ctx context.Context, playersURL string) (*Stream, error) {
	players, err := c.Players(ctx, playersURL)
	if err != nil {
		return nil, err
	}
	return c.ResolveStream(ctx, players)
}

// NoStreamError reports why no player yielded a stream: hosts this package has
// no resolver for, and supported hosts that failed. It matches
// ErrNoSupportedServer under errors.Is.
type NoStreamError struct {
	Unsupported []string
	Failures    []error
}

func (e *NoStreamError) Error() string {
	var b strings.Builder
	b.WriteString(ErrNoSupportedServer.Error())
	if len(e.Unsupported) > 0 {
		b.WriteString(" (unsupported: " + strings.Join(e.Unsupported, ", ") + ")")
	}
	for _, f := range e.Failures {
		b.WriteString("; " + f.Error())
	}
	return b.String()
}

func (e *NoStreamError) Is(target error) bool { return target == ErrNoSupportedServer }

func (e *NoStreamError) Unwrap() []error { return e.Failures }

// playerRank orders resolution attempts: Byse first (a single JSON call), then
// Abyss (one page plus a probe, then served through the local proxy), then
// direct files (the panel's own video element, often dead), then the rest.
func playerRank(p Player) int {
	u, err := url.Parse(p.URL)
	switch {
	case err != nil:
		return 4
	case isByseHost(u.Host):
		return 0
	case isAbyssHost(u.Host):
		return 1
	case p.IsDirectFile():
		return 2
	default:
		return 3
	}
}

// ResolveStream tries each player in turn and returns the first stream.
func (c *Client) ResolveStream(ctx context.Context, players []Player) (*Stream, error) {
	if len(players) == 0 {
		return nil, ErrNoPlayers
	}
	ordered := slices.Clone(players)
	slices.SortStableFunc(ordered, func(a, b Player) int { return playerRank(a) - playerRank(b) })

	noStream := &NoStreamError{}
	for _, p := range ordered {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		u, err := url.Parse(p.URL)
		if err != nil || u.Host == "" {
			continue
		}
		var s *Stream
		switch {
		case isByseHost(u.Host):
			s, err = c.resolveByse(ctx, u)
		case isAbyssHost(u.Host):
			s, err = c.resolveAbyss(ctx, u, p)
		case p.IsDirectFile():
			s, err = c.resolveDirect(ctx, p, u)
		default:
			if !slices.Contains(noStream.Unsupported, u.Host) {
				noStream.Unsupported = append(noStream.Unsupported, u.Host)
			}
			continue
		}
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			util.Debug("StartFlix player failed", "host", u.Host, "label", p.Label, "err", err)
			noStream.Failures = append(noStream.Failures, fmt.Errorf("%s: %w", u.Host, err))
			continue
		}
		if p.Subtitles != "" {
			s.Subtitles = append(s.Subtitles, Subtitle{URL: p.Subtitles, Language: "pt-br", Label: "Português"})
		}
		util.Debug("StartFlix stream resolved", "host", s.Host, "label", p.Label)
		return s, nil
	}
	return nil, noStream
}

// resolveDirect checks that a direct-file player still serves media. The panel
// hands these straight to its <video> element, so a dead one (apiblogger.click
// answered 302 with an empty Location on 2026-10-07) would otherwise reach mpv.
func (c *Client) resolveDirect(ctx context.Context, p Player, u *url.URL) (*Stream, error) {
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Range", "bytes=0-1")
	if p.referer != "" {
		req.Header.Set("Referer", p.referer)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, netx.NewHTTPStatusError(SourceName, "direct", resp.StatusCode)
	}
	if ct := strings.ToLower(resp.Header.Get("Content-Type")); strings.HasPrefix(ct, "text/html") {
		return nil, errors.New("served a web page instead of media")
	}
	// The listed URL, not where it redirected: players and the downloader follow
	// redirects themselves, and a redirect target is often a signed URL that
	// expires long before a binge reaches it.
	return &Stream{URL: u.String(), Referer: p.referer, Host: u.Host}, nil
}
