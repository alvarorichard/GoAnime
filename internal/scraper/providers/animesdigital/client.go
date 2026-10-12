package animesdigital

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
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
	SourceName  = "AnimesDigital"
	DefaultBase = "https://animesdigital.org"
	maxBodySize = 12 << 20
)

var (
	dubbedSuffixRe = regexp.MustCompile(`(?i)\s*(?:[-–—]\s*)?dublado\s*$`)
	episodeNumRe   = regexp.MustCompile(`(?i)(?:epis[oó]dio|ep\.?)\D*(\d+(?:\.\d+)?)`)
	pageNumRe      = regexp.MustCompile(`/page/(\d+)/?`)
)

// Variants are the title pages available for each audio.
type Variants struct {
	Subtitled string
	Dubbed    string
}

// Client performs AnimesDigital requests and parses its WordPress pages.
type Client struct {
	http       *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewClient builds the production client without doing network I/O.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_ANIMESDIGITAL_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	return &Client{
		http:       util.NewFastClient(),
		baseURL:    base,
		userAgent:  netx.UserAgent,
		maxRetries: 2,
		retryDelay: 350 * time.Millisecond,
	}
}

// NewAnimesDigitalClient keeps the constructor shape used by the other leaf
// providers.
func NewAnimesDigitalClient() *Client { return NewClient() }

// NewClientForTest points a client at an HTTP test server and disables retries.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	return &Client{
		http:       hc,
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  netx.UserAgent,
		maxRetries: 0,
	}
}

var (
	sharedOnce   sync.Once
	sharedClient *Client
)

// Shared returns the process-wide client.
func Shared() *Client {
	sharedOnce.Do(func() { sharedClient = NewClient() })
	return sharedClient
}

// ParseVariants decodes the private fragment used to carry a paired dubbed
// title. A plain URL is classified by its slug.
func ParseVariants(rawURL string) Variants {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return Variants{Subtitled: rawURL}
	}
	encoded, dubbedOnly := "", false
	if values, err := url.ParseQuery(u.Fragment); err == nil {
		encoded = values.Get("goanime-dub")
		dubbedOnly = values.Get("goanime-dublado-only") == "1"
	}
	u.Fragment = ""
	clean := u.String()
	if dubbedOnly {
		return Variants{Dubbed: clean}
	}
	if encoded != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(encoded)
		if err == nil {
			return Variants{Subtitled: clean, Dubbed: string(decoded)}
		}
	}
	if isDubbed(rawURL, "") {
		return Variants{Dubbed: clean}
	}
	return Variants{Subtitled: clean}
}

func markDubbedOnly(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	values := url.Values{}
	values.Set("goanime-dublado-only", "1")
	u.Fragment = values.Encode()
	return u.String()
}

func mergeVariantURLs(subtitled, dubbed string) string {
	if subtitled == "" {
		return dubbed
	}
	if dubbed == "" {
		return subtitled
	}
	u, err := url.Parse(subtitled)
	if err != nil {
		return subtitled
	}
	values := url.Values{}
	values.Set("goanime-dub", base64.RawURLEncoding.EncodeToString([]byte(dubbed)))
	u.Fragment = values.Encode()
	return u.String()
}

type getResult struct {
	doc      *goquery.Document
	finalURL string
}

func (c *Client) getDocument(ctx context.Context, rawURL, layer string) (*getResult, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := c.sleep(ctx, attempt); err != nil {
				return nil, err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		c.decorateRequest(req)
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt, err) {
				continue
			}
			return nil, err
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if c.shouldRetry(attempt, readErr) {
				continue
			}
			return nil, readErr
		}
		if err := netx.CheckHTTPStatus(resp, SourceName+" "+layer); err != nil {
			lastErr = err
			if c.shouldRetry(attempt, err) {
				continue
			}
			return nil, err
		}
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return nil, netx.NewParserError(SourceName, layer, "unreadable HTML", err)
		}
		if err := netx.CheckChallengeDocument(doc, SourceName+" "+layer); err != nil {
			return nil, netx.NewBlockedChallengeError(SourceName, layer, "Cloudflare challenge blocks plain HTTP; open the site in a browser and try again later", err)
		}
		finalURL := rawURL
		if resp.Request != nil && resp.Request.URL != nil {
			finalURL = resp.Request.URL.String()
		}
		return &getResult{doc: doc, finalURL: finalURL}, nil
	}
	return nil, lastErr
}

func (c *Client) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Referer", c.baseURL+"/")
}

func (c *Client) shouldRetry(attempt int, err error) bool {
	if attempt >= c.maxRetries || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var diagnostic *netx.SourceDiagnostic
	if errors.As(err, &diagnostic) {
		return diagnostic.Kind == netx.DiagnosticSourceUnavailable
	}
	return true
}

func (c *Client) sleep(ctx context.Context, attempt int) error {
	if c.retryDelay <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.retryDelay * time.Duration(attempt)):
		return nil
	}
}

func (c *Client) resolveURL(base, ref string) string {
	u, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return ref
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	return b.ResolveReference(u).String()
}

// SearchAnime searches using the WordPress search page.
func (c *Client) SearchAnime(query string) ([]*models.Anime, error) {
	return c.SearchAnimeContext(context.Background(), query)
}

// SearchAnimeContext is the cancellable search implementation.
func (c *Client) SearchAnimeContext(ctx context.Context, query string) ([]*models.Anime, error) {
	q := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(query)), " ")
	if q == "" {
		return nil, nil
	}
	page, err := c.getDocument(ctx, c.baseURL+"/?s="+url.QueryEscape(q), "search")
	if err != nil {
		return nil, err
	}
	return c.extractSearchResults(page.doc), nil
}

type searchGroup struct {
	title     string
	subtitled *models.Anime
	dubbed    *models.Anime
}

func (c *Client) extractSearchResults(doc *goquery.Document) []*models.Anime {
	groups := make(map[string]*searchGroup)
	order := make([]string, 0)
	seenURLs := make(map[string]bool)

	doc.Find(`a[href*="/anime/"]`).Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}
		href = c.resolveURL(c.baseURL, href)
		if seenURLs[href] {
			return
		}
		title := strings.TrimSpace(s.Find(".title_anime").First().Text())
		if title == "" {
			title, _ = s.Attr("title")
			title = strings.TrimSpace(strings.TrimPrefix(title, "Assistir "))
			if cut, _, ok := strings.Cut(title, " Online"); ok {
				title = strings.TrimSpace(cut)
			}
		}
		if title == "" {
			return
		}
		seenURLs[href] = true
		imageURL := ""
		img := s.Find("img").First()
		for _, attr := range []string{"src", "data-src", "data-lazy-src"} {
			if value, exists := img.Attr(attr); exists && strings.TrimSpace(value) != "" {
				imageURL = c.resolveURL(c.baseURL, value)
				break
			}
		}
		baseTitle := strings.TrimSpace(dubbedSuffixRe.ReplaceAllString(title, ""))
		key := strings.ToLower(strings.Join(strings.Fields(baseTitle), " "))
		group := groups[key]
		if group == nil {
			group = &searchGroup{title: baseTitle}
			groups[key] = group
			order = append(order, key)
		}
		anime := &models.Anime{Name: title, URL: href, ImageURL: imageURL, MediaType: models.MediaTypeAnime}
		if isDubbed(href, title) {
			group.dubbed = anime
		} else {
			group.subtitled = anime
		}
	})

	results := make([]*models.Anime, 0, len(order))
	for _, key := range order {
		group := groups[key]
		switch {
		case group.subtitled != nil && group.dubbed != nil:
			anime := *group.subtitled
			anime.Name = group.title
			anime.URL = mergeVariantURLs(group.subtitled.URL, group.dubbed.URL)
			if anime.ImageURL == "" {
				anime.ImageURL = group.dubbed.ImageURL
			}
			results = append(results, &anime)
		case group.subtitled != nil:
			results = append(results, group.subtitled)
		case group.dubbed != nil:
			anime := *group.dubbed
			anime.URL = markDubbedOnly(anime.URL)
			results = append(results, &anime)
		}
	}
	return results
}

func isDubbed(rawURL, title string) bool {
	return dubbedSuffixRe.MatchString(strings.TrimSpace(title)) ||
		strings.Contains(strings.ToLower(rawURL), "-dublado")
}

type numberedEpisode struct {
	label string
	value float64
	url   string
}

func (c *Client) parseEpisodePage(doc *goquery.Document) ([]numberedEpisode, string, int) {
	episodes := make([]numberedEpisode, 0)
	seen := make(map[string]bool)
	doc.Find(`a[href*="/video/"]`).Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || href == "" {
			return
		}
		href = c.resolveURL(c.baseURL, href)
		if seen[href] {
			return
		}
		match := episodeNumRe.FindStringSubmatch(strings.Join(strings.Fields(s.Text()), " "))
		if len(match) < 2 {
			return
		}
		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return
		}
		seen[href] = true
		episodes = append(episodes, numberedEpisode{label: match[1], value: value, url: href})
	})

	root, maxPage := "", 1
	doc.Find(`a[href*="/page/"]`).Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href = c.resolveURL(c.baseURL, href)
		match := pageNumRe.FindStringSubmatch(href)
		if len(match) < 2 {
			return
		}
		page, err := strconv.Atoi(match[1])
		if err != nil || page < 1 || page > 200 {
			return
		}
		if page > maxPage {
			maxPage = page
		}
		if root == "" {
			root = strings.TrimRight(pageNumRe.ReplaceAllString(href, "/"), "/")
		}
	})
	return episodes, root, maxPage
}

// GetAnimeEpisodes retrieves every paginated episode and returns ascending
// order, even though the site lists newest first.
func (c *Client) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return c.GetAnimeEpisodesContext(context.Background(), animeURL)
}

// GetAnimeEpisodesContext is the cancellable episode implementation.
func (c *Client) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	variants := ParseVariants(animeURL)
	target := variants.Subtitled
	if target == "" {
		target = variants.Dubbed
	}
	first, err := c.getDocument(ctx, target, "episodes")
	if err != nil {
		return nil, err
	}
	all, root, maxPage := c.parseEpisodePage(first.doc)
	if root == "" {
		root = strings.TrimRight(first.finalURL, "/")
	}

	if maxPage > 1 {
		type pageResult struct {
			episodes []numberedEpisode
			err      error
		}
		jobs := make(chan int)
		results := make(chan pageResult, maxPage-1)
		workers := min(6, maxPage-1)
		var wg sync.WaitGroup
		for range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for page := range jobs {
					doc, err := c.getDocument(ctx, fmt.Sprintf("%s/page/%d/", root, page), "episodes")
					if err != nil {
						results <- pageResult{err: err}
						continue
					}
					episodes, _, _ := c.parseEpisodePage(doc.doc)
					results <- pageResult{episodes: episodes}
				}
			}()
		}
		go func() {
			defer close(jobs)
			for page := 2; page <= maxPage; page++ {
				select {
				case jobs <- page:
				case <-ctx.Done():
					return
				}
			}
		}()
		go func() {
			wg.Wait()
			close(results)
		}()
		var firstErr error
		for result := range results {
			if result.err != nil && firstErr == nil {
				firstErr = result.err
			}
			all = append(all, result.episodes...)
		}
		if firstErr != nil {
			return nil, firstErr
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	byURL := make(map[string]numberedEpisode, len(all))
	for _, episode := range all {
		byURL[episode.url] = episode
	}
	all = all[:0]
	for _, episode := range byURL {
		all = append(all, episode)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].value == all[j].value {
			return all[i].url < all[j].url
		}
		return all[i].value < all[j].value
	})
	if len(all) == 0 {
		return nil, netx.NewParserError(SourceName, "episodes", "no episode links found", nil)
	}
	episodes := make([]models.Episode, 0, len(all))
	for _, episode := range all {
		episodes = append(episodes, models.Episode{
			Number: "Episódio " + episode.label,
			Num:    int(episode.value),
			URL:    episode.url,
		})
	}
	return episodes, nil
}

// GetEpisodeStreamURL extracts the HLS URL stored in player 1's d parameter.
func (c *Client) GetEpisodeStreamURL(episodeURL string) (string, error) {
	return c.GetEpisodeStreamURLContext(context.Background(), episodeURL)
}

// GetEpisodeStreamURLContext is the cancellable stream implementation.
func (c *Client) GetEpisodeStreamURLContext(ctx context.Context, episodeURL string) (string, error) {
	page, err := c.getDocument(ctx, episodeURL, "stream")
	if err != nil {
		return "", err
	}
	var playerOne, alternate string
	page.doc.Find("iframe.metaframe.rptss, iframe.rptss.metaframe").Each(func(_ int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if !ok || strings.TrimSpace(src) == "" {
			return
		}
		src = c.resolveURL(c.baseURL, src)
		u, err := url.Parse(src)
		if err != nil {
			return
		}
		if strings.EqualFold(u.Hostname(), "api.anivideo.net") && strings.Contains(u.Path, "videohls.php") {
			if direct := strings.TrimSpace(u.Query().Get("d")); direct != "" {
				playerOne = direct
			}
		} else if alternate == "" {
			alternate = src
		}
	})
	if playerOne != "" {
		return netx.ValidateStreamURL(playerOne, SourceName)
	}
	if alternate != "" {
		return "", netx.NewParserError(SourceName, "stream", "player 1 is unavailable; player 2 currently redirects to an unusable api.anivideo.net endpoint", nil)
	}
	return "", netx.NewParserError(SourceName, "stream", "no supported player iframe found", nil)
}
