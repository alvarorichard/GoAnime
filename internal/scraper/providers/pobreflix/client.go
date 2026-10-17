package pobreflix

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// DefaultBase is the catalog host. The site runs mirror domains
	// (pobreflixtv.sale and friends), so GOANIME_POBREFLIX_URL can repoint
	// this without waiting for a release.
	DefaultBase = "https://www.pobreflixtv.sale"

	// userAgent describes a current Chrome, matching the TLS fingerprint
	// util.NewFastClient presents. Neither the site nor its player checked it
	// on 2026-10-16.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"

	maxBodyBytes = 8 << 20 // a sitemap shard reached ~180 KB; headroom for CMS noise
)

// Client talks to Pobreflix: the sitemap-backed search corpus and the title
// pages. It owns the corpus so the index is built once and shared.
type Client struct {
	http       *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration

	// search is the sitemap-backed index, built lazily on the first Search
	// and cached on disk across runs.
	search corpus
	titles sync.Map // title page URL -> Title
}

// NewClient builds a client. It loads the corpus cache from disk (cheap — it
// is the search's warm path) but performs no network I/O.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_POBREFLIX_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	c := &Client{
		http:       util.NewFastClient(),
		baseURL:    base,
		userAgent:  userAgent,
		maxRetries: 2,
		retryDelay: 400 * time.Millisecond,
	}
	c.search.cachePath, _ = corpusCachePath()
	c.search.loadCache(c.baseURL)
	return c
}

// NewClientForTest points a client at a test server; the corpus cache lives
// in cachePath (tests pass a t.TempDir() file) so suite runs never fight over
// the user's cache.
func NewClientForTest(hc *http.Client, baseURL, cachePath string) *Client {
	return &Client{
		http:      hc,
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
		search:    corpus{cachePath: cachePath},
	}
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client, so the search corpus is built once
// and the title cache is shared between the search adapter and the playback
// path.
func Shared() *Client {
	sharedOnce.Do(func() { sharedClient = NewClient() })
	return sharedClient
}

// BaseURL is the catalog host this client searches.
func (c *Client) BaseURL() string { return c.baseURL }

type getOptions struct {
	referer string
	accept  string
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
		util.Debug("Pobreflix request failed, retrying", "url", rawURL, "attempt", attempt+1, "err", err)
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
	return nil, resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests,
		netx.NewHTTPStatusError(SourceName, "http", resp.StatusCode)
}

func (c *Client) getDocument(ctx context.Context, rawURL string, opt getOptions) (*goquery.Document, error) {
	body, err := c.get(ctx, rawURL, opt)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, netx.NewParserError(SourceName, "title", "unreadable HTML", err)
	}
	return doc, nil
}

// sitemapShardRe names a video shard inside the sitemap index. The site
// generates the file names; the numeric suffix orders them.
var sitemapShardRe = regexp.MustCompile(`file=((?:[a-z0-9_]+_)?videobox_Video_\d+)`)

// sitemapShards lists the video shards the index publishes.
func (c *Client) sitemapShards(ctx context.Context) ([]shard, error) {
	body, err := c.get(ctx, c.baseURL+"/sitemap.php", getOptions{
		accept: "text/xml,application/xml,text/html,*/*;q=0.8",
	})
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []shard
	for _, m := range locRe.FindAllStringSubmatch(string(body), -1) {
		name := sitemapShardRe.FindStringSubmatch(m[1])
		if name == nil || seen[name[1]] {
			continue
		}
		seen[name[1]] = true
		out = append(out, shard{name: name[1], url: c.baseURL + "/sitemap.php?file=" + name[1]})
	}
	sortShards(out)
	return out, nil
}

// sortShards orders by the numeric suffix, deterministically.
func sortShards(shards []shard) {
	sort.Slice(shards, func(i, j int) bool { return shards[i].number() < shards[j].number() })
}

// Search ranks the catalog against the query. Hyphens and underscores become
// spaces: CLI arguments arrive slugged ("todo-poderoso").
//
// The first search builds the corpus: the sitemap index, then every shard,
// bounded-parallel. The corpus is cached on disk; within the refresh window a
// search costs nothing, after it only the shards missing from the index are
// re-fetched. A build that runs out of deadline keeps what it fetched — the
// partial index is searchable already, so the error never defeats a result.
func (c *Client) Search(ctx context.Context, query string) ([]Media, error) {
	q := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(query)), " ")
	if q == "" {
		return nil, nil
	}
	if c.search.fresh() {
		return c.search.rank(q), nil
	}
	// Build or resume. A finished build (or an old cache carrying entries)
	// is searched this round; a cold cache that learned nothing is an error
	// the registry reports to the TRY screen.
	err := c.search.refresh(ctx, c)
	results := c.search.rank(q)
	if len(results) > 0 || err == nil {
		c.search.saveCache()
	}
	if len(results) == 0 && err != nil && !c.search.hasEntries() {
		return nil, err
	}
	return results, nil
}

// Title reads a title page. Cached per page: the search opens one page for
// the TMDB id and the episode-name fills may open it again.
func (c *Client) Title(ctx context.Context, pageURL string) (Title, error) {
	if t, ok := c.titles.Load(pageURL); ok {
		return t.(Title), nil
	}
	doc, err := c.getDocument(ctx, pageURL, getOptions{referer: c.baseURL + "/"})
	if err != nil {
		return Title{}, err
	}
	t, ok := parseTitle(doc, pageURL)
	if !ok {
		return Title{}, ErrNoTitle
	}
	c.titles.Store(pageURL, t)
	util.Debug("Pobreflix title", "page", pageURL, "kind", t.Kind, "video", t.VideoID, "tmdb", t.TMDBID, "seasons", len(t.Seasons))
	return t, nil
}
