package topcine

import (
	"bytes"
	"context"
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
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

const (
	// DefaultBase is the catalog host. The number in the name is a mirror
	// generation (topcine3), so GOANIME_TOPCINE_URL can repoint it without
	// waiting for a release.
	DefaultBase = "https://topcine3.site"

	// userAgent describes a current Chrome, matching the TLS fingerprint
	// util.NewFastClient presents. Neither the site nor its player checked it
	// on 2026-10-09.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"

	maxBodyBytes = 5 << 20
)

// Client talks to TopCine and the player its title pages embed.
type Client struct {
	http       *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration

	titles sync.Map // title page URL -> Title
}

// NewClient builds a client. It performs no network I/O.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_TOPCINE_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	return &Client{
		http:       util.NewFastClient(),
		baseURL:    base,
		userAgent:  userAgent,
		maxRetries: 2,
		retryDelay: 400 * time.Millisecond,
	}
}

// NewClientForTest points a client at a test server.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	return &Client{
		http:      hc,
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
	}
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client, so the title cache is shared between
// the search adapter and the playback path.
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
		util.Debug("TopCine request failed, retrying", "url", rawURL, "attempt", attempt+1, "err", err)
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
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, false, nil
	}
	layer := opt.layer
	if layer == "" {
		layer = "http"
	}
	return nil, resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests,
		netx.NewHTTPStatusError(SourceName, layer, resp.StatusCode)
}

func (c *Client) getDocument(ctx context.Context, rawURL string, opt getOptions) (*goquery.Document, error) {
	body, err := c.get(ctx, rawURL, opt)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, netx.NewParserError(SourceName, opt.layer, "unreadable HTML", err)
	}
	return doc, nil
}

// Search queries the catalog's first result page (20 titles). Hyphens and
// underscores become spaces: CLI arguments arrive slugged ("the-boys").
//
// The form posts to /buscar.php, which only redirects to /buscar — and over
// plain http — so the final URL is asked for directly.
func (c *Client) Search(ctx context.Context, query string) ([]Media, error) {
	q := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(query)), " ")
	if q == "" {
		return nil, nil
	}
	base, err := url.Parse(c.baseURL + "/")
	if err != nil {
		return nil, fmt.Errorf("topcine: invalid base URL %q: %w", c.baseURL, err)
	}
	doc, err := c.getDocument(ctx, c.baseURL+"/buscar?q="+url.QueryEscape(q), getOptions{
		referer: c.baseURL + "/",
		layer:   "search",
	})
	if err != nil {
		return nil, err
	}
	results := parseSearchResults(doc, base)
	util.Debug("TopCine search", "query", q, "results", len(results))
	return results, nil
}

// Title reads a title page. Cached per page.
func (c *Client) Title(ctx context.Context, pageURL string) (Title, error) {
	if t, ok := c.titles.Load(pageURL); ok {
		return t.(Title), nil
	}
	doc, err := c.getDocument(ctx, pageURL, getOptions{referer: c.baseURL + "/", layer: "title"})
	if err != nil {
		return Title{}, err
	}
	t, ok := parseTitle(doc)
	if !ok {
		return Title{}, ErrNoPlayer
	}
	c.titles.Store(pageURL, t)
	util.Debug("TopCine title", "page", pageURL, "kind", t.Kind, "tmdb", t.TMDBID, "seasons", len(t.Seasons))
	return t, nil
}

type languagesRecord struct {
	OK           bool    `json:"sucesso"`
	HasDubbed    bool    `json:"tem_dublado"`
	HasSubtitled bool    `json:"tem_legendado"`
	DubbedURL    *string `json:"url_dublado"`
	SubtitledURL *string `json:"url_legendado"`
}

// Languages asks a player which audio tracks it has — the request its "Abrir
// Player" button makes. It needs no referer.
func (c *Client) Languages(ctx context.Context, player string) (Languages, error) {
	u, ok := playerURL(player)
	if !ok {
		return Languages{}, fmt.Errorf("topcine: invalid player URL %q", player)
	}
	q := u.Query()
	q.Set("acao", "carregar_idiomas")
	u.RawQuery = q.Encode()
	body, err := c.get(ctx, u.String(), getOptions{
		referer: player,
		accept:  "application/json, text/javascript, */*; q=0.01",
		ajax:    true,
		layer:   "player",
	})
	if err != nil {
		return Languages{}, err
	}
	var rec languagesRecord
	if err := jsonx.Unmarshal(body, &rec); err != nil {
		return Languages{}, netx.NewParserError(SourceName, "player", "unreadable language record", err)
	}
	if !rec.OK {
		return Languages{}, netx.NewParserError(SourceName, "player", "player refused the language request", nil)
	}
	out := Languages{Dubbed: rec.HasDubbed, Subtitled: rec.HasSubtitled}
	for _, raw := range []*string{rec.DubbedURL, rec.SubtitledURL} {
		if raw == nil {
			continue
		}
		if h, ok := playerURL(*raw); ok && !slices.Contains(out.Hosts, h.Host) {
			out.Hosts = append(out.Hosts, h.Host)
		}
	}
	return out, nil
}
