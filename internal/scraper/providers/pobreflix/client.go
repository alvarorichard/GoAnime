package pobreflix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	// DefaultBase is the catalog host. The site moves between mirror domains
	// (the old .sale host is dead; .gay is live as of 2026-10-24), so
	// GOANIME_POBREFLIX_URL can repoint this without waiting for a release.
	DefaultBase = "https://pobreflixtv.gay"

	// userAgent describes a current Chrome, matching the TLS fingerprint
	// util.NewFastClient presents. Neither the site nor its player checked it
	// on 2026-10-24.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"

	maxBodyBytes = 16 << 20 // current sitemap-movies.xml is ~2.6 MB; headroom for growth
)

// Client talks to Pobreflix: the sitemap-backed search corpus and the title
// pages. It owns the corpus so the index is built once and shared.
type Client struct {
	http       *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration

	// search is the sitemap-backed index, built lazily when the site's own
	// search API misses and cached on disk across runs.
	search corpus
	// apiEnabled lets production ask the site's /api catalog first; tests
	// pin the sitemap corpus so their request counts stay deterministic.
	apiEnabled bool
	titles     sync.Map // title page URL -> Title
	episodes   sync.Map // episode page URL -> EpisodeData
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
		apiEnabled: true,
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

// sitemapShardRe names a video shard inside the old sitemap index. The site
// generated the file names; the numeric suffix orders them.
var sitemapShardRe = regexp.MustCompile(`file=((?:[a-z0-9_]+_)?videobox_Video_\d+)`)

// currentSitemapRe names a current content sitemap. The site publishes one
// movie and one series file (and, on busier weeks, numbered continuations).
var currentSitemapRe = regexp.MustCompile(`(?i)/sitemap-(movies|series)(?:-(\d+))?\.xml$`)

// sitemapShards lists the content sitemaps the index publishes. The current
// site replaced the old sitemap.php?file=… shards with two large
// sitemap-movies.xml / sitemap-series.xml files; both layouts are supported so
// a mirror still running the old theme keeps working.
func (c *Client) sitemapShards(ctx context.Context) ([]shard, error) {
	body, err := c.get(ctx, c.baseURL+"/sitemap.php", getOptions{
		accept: "text/xml,application/xml,text/html,*/*;q=0.8",
	})
	if err == nil {
		seen := map[string]bool{}
		var out []shard
		for _, m := range locRe.FindAllStringSubmatch(string(body), -1) {
			name := sitemapShardRe.FindStringSubmatch(m[1])
			if len(name) < 2 || seen[name[1]] {
				continue
			}
			seen[name[1]] = true
			out = append(out, shard{name: name[1], url: c.baseURL + "/sitemap.php?file=" + name[1]})
		}
		if len(out) > 0 {
			sortShards(out)
			return out, nil
		}
	}

	body, err = c.get(ctx, c.baseURL+"/sitemap.xml", getOptions{
		accept: "text/xml,application/xml,text/html,*/*;q=0.8",
	})
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []shard
	for _, m := range locRe.FindAllStringSubmatch(string(body), -1) {
		loc := strings.TrimSpace(m[1])
		name := currentSitemapRe.FindStringSubmatch(pathOfURL(loc))
		if len(name) < 2 {
			continue
		}
		order := 1
		if name[1] == "series" {
			order = 2
		}
		if len(name) > 2 && name[2] != "" {
			if n, err := strconv.Atoi(name[2]); err == nil {
				order = order*1000 + n
			}
		}
		key := name[1]
		if len(name) > 2 && name[2] != "" {
			key += "-" + name[2]
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, shard{name: key, url: loc, order: order})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("pobreflix: sitemap lists no movie/series files")
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
	// The site's own catalog API answers in one request and carries the
	// record's real name, which the sitemap corpus can only approximate from
	// the slug. Production asks it first; tests pin the corpus instead so
	// their request counts stay deterministic.
	if c.apiEnabled {
		if results, err := c.apiSearch(ctx, q); err != nil {
			util.Debug("Pobreflix catalog API failed, falling back to the sitemap corpus", "err", err)
		} else if len(results) > 0 {
			return results, nil
		}
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

// catalogResponse is the /api?action=catalog envelope.
type catalogResponse struct {
	Type  string         `json:"type"`
	Total int            `json:"total"`
	Items []embeddedItem `json:"items"`
}

// apiSearch asks the site's own catalog for movies and series. It is the
// authoritative name source; a failure of either half does not discard the
// other.
func (c *Client) apiSearch(ctx context.Context, query string) ([]Media, error) {
	type outcome struct {
		kind  MediaKind
		items []embeddedItem
		err   error
	}
	types := []struct {
		param string
		kind  MediaKind
	}{
		{"movie", KindMovie},
		{"series", KindSeries},
	}
	ch := make(chan outcome, len(types))
	for _, t := range types {
		go func(param string, kind MediaKind) {
			items, err := c.catalog(ctx, param, query)
			ch <- outcome{kind: kind, items: items, err: err}
		}(t.param, t.kind)
	}

	var (
		out     []Media
		lastErr error
		seen    = map[string]bool{}
	)
	for range types {
		res := <-ch
		if res.err != nil {
			lastErr = res.err
			continue
		}
		for _, item := range res.items {
			slug := strings.TrimSpace(item.Slug)
			if slug == "" {
				continue
			}
			path := "/filme/"
			if res.kind == KindSeries {
				path = "/serie/"
			}
			rawURL := c.baseURL + path + url.PathEscape(slug) + ".html"
			if seen[rawURL] {
				continue
			}
			seen[rawURL] = true
			title := strings.Join(strings.Fields(item.Name), " ")
			if title == "" {
				title = slugToTitle(displaySlug(slug))
			}
			out = append(out, Media{
				Title:   title,
				URL:     rawURL,
				Kind:    res.kind,
				Audio:   audioFromEmbedded(item.Audio),
				VideoID: item.InternalID,
			})
		}
	}
	if len(out) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, nil
	}
	sort.SliceStable(out, func(i, j int) bool {
		si := models.RelevanceScore(query, out[i].Title, "")
		sj := models.RelevanceScore(query, out[j].Title, "")
		if si != sj {
			return si > sj
		}
		if out[i].Audio != out[j].Audio {
			return out[i].Audio == AudioDubbed
		}
		return out[i].VideoID > out[j].VideoID
	})
	if len(out) > maxSearchResults {
		out = out[:maxSearchResults]
	}
	return out, nil
}

// catalog reads one /api?action=catalog page.
func (c *Client) catalog(ctx context.Context, kind, query string) ([]embeddedItem, error) {
	params := url.Values{}
	params.Set("action", "catalog")
	params.Set("type", kind)
	params.Set("limit", strconv.Itoa(maxSearchResults))
	params.Set("sort", "recent")
	params.Set("search", query)
	body, err := c.get(ctx, c.baseURL+"/api?"+params.Encode(), getOptions{
		accept:  "application/json",
		referer: c.baseURL + "/",
	})
	if err != nil {
		return nil, err
	}
	var resp catalogResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, netx.NewParserError(SourceName, "search", "unreadable catalog JSON", err)
	}
	return resp.Items, nil
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
	t.Servers = c.absolutizeServers(t.Servers)
	c.titles.Store(pageURL, t)
	util.Debug("Pobreflix title", "page", pageURL, "kind", t.Kind, "video", t.VideoID, "tmdb", t.TMDBID,
		"direct", t.HasDirect(), "servers", len(t.Servers), "seasons", len(t.Seasons))
	return t, nil
}

// Episode reads a current-site episode page: the series it belongs to, its
// slot and the playable files the page embeds.
func (c *Client) Episode(ctx context.Context, pageURL string) (EpisodeData, error) {
	if e, ok := c.episodes.Load(pageURL); ok {
		return e.(EpisodeData), nil
	}
	doc, err := c.getDocument(ctx, pageURL, getOptions{referer: c.baseURL + "/"})
	if err != nil {
		return EpisodeData{}, err
	}
	raw, ok := embeddedEpisodeData(doc)
	if !ok {
		return EpisodeData{}, ErrNoEpisode
	}
	series := titleFromEmbeddedItem(raw.Series, KindSeries, pageURL)
	series.Audio = audioFromEmbedded(raw.Episode.Audio)
	if series.Audio == AudioUnknown {
		series.Audio = audioFromEmbedded(raw.Series.Audio)
	}
	ep := EpisodeData{
		Series:  series,
		Season:  raw.Episode.Season,
		Number:  raw.Episode.Episode,
		Name:    strings.Join(strings.Fields(raw.Episode.Title), " "),
		Servers: c.absolutizeServers(serversFromEmbedded(raw.Servers)),
	}
	c.episodes.Store(pageURL, ep)
	util.Debug("Pobreflix episode", "page", pageURL, "season", ep.Season, "episode", ep.Number, "servers", len(ep.Servers))
	return ep, nil
}

// EpisodeURL builds the current site's episode-page URL from a series title.
func (c *Client) EpisodeURL(t Title, season, episode int) string {
	path := t.EpisodePath(season, episode)
	if path == "" {
		return ""
	}
	return c.baseURL + path
}

// absolutizeServers resolves the site's relative download endpoints against
// the catalog host. Movie files are already absolute R2 URLs.
func (c *Client) absolutizeServers(servers []Server) []Server {
	for i := range servers {
		servers[i].URL = c.absoluteURL(servers[i].URL)
		servers[i].Download = c.absoluteURL(servers[i].Download)
	}
	return servers
}

// absoluteURL resolves one URL against the catalog host.
func (c *Client) absoluteURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(raw, "://") {
		return raw
	}
	if strings.HasPrefix(raw, "//") {
		if u, err := url.Parse(c.baseURL); err == nil {
			return u.Scheme + ":" + raw
		}
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return c.baseURL + raw
}

// ResolveDownload follows an episode's download endpoint to its final media
// URL. The endpoint answers 302; when it does not, the endpoint itself is
// returned (mpv and the downloader follow redirects on their own).
func (c *Client) ResolveDownload(ctx context.Context, rawURL string) string {
	rawURL = c.absoluteURL(rawURL)
	if rawURL == "" {
		return ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, http.NoBody)
	if err != nil {
		return rawURL
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	req.Header.Set("Referer", c.baseURL+"/")
	resp, err := c.http.Do(req)
	if err != nil {
		return rawURL
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if loc := strings.TrimSpace(resp.Header.Get("Location")); loc != "" {
			return c.absoluteURL(loc)
		}
	}
	if resp.Request != nil && resp.Request.URL != nil {
		if final := resp.Request.URL.String(); final != "" && final != rawURL {
			return final
		}
	}
	return rawURL
}
