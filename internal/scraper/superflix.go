// Package scraper provides web scraping functionality for SuperFlix movies, TV shows, animes and doramas
package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ErrSuperFlixNoServers is returned when /player/bootstrap responds with an
// empty options list. This is a content-availability signal from SuperFlix
// (the upstream JS shows a "not yet released" screen in the same case), not
// a system or scraping error — callers should surface it to the user as
// "this episode has no source on SuperFlix" rather than retrying.
var ErrSuperFlixNoServers = errors.New("superflix: no servers available for this content")

const (
	// SuperFlixBase is the canonical SuperFlix host. Previous hosts
	// (`superflixapi.rest`, `superflixapi.online`, `superflixapi.best`,
	// `superflixapi.fit`) 301-redirect to whichever alias is live; Go's
	// http.Client follows the redirect but downgrades the POST to a GET
	// (dropping the body), which makes /player/bootstrap return HTML 404 and
	// break JSON decoding — so we target the live host directly. `.fit` went
	// dead (NXDOMAIN) 2026-06-18 and rotated to `.cyou`.
	SuperFlixBase = "https://superflixapi.cyou"
	// SuperFlixEmbedHost is the warezcdn embed host listed in the frontend's
	// window.__PLAYER_APIS__. Loading https://warezcdn.lat/{filme|serie}/<tmdb>
	// in a cross-origin iframe funnels through Turnstile to the rotating player
	// host whose getVideo endpoint returns the signed HLS master. Verified live
	// 2026-06-09. superflixapi.cyou is the current API alias if this rotates out.
	SuperFlixEmbedHost = "warezcdn.lat"
	// SuperFlixUserAgent MUST match the UA the CF solver's Firefox presents
	// (see cfBrowserSolver.Solve). Cloudflare binds the cf_clearance cookie to
	// the User-Agent that solved the challenge; if the HTTP client then sends a
	// different UA, CF rejects the clearance and re-challenges in a loop. A
	// Firefox-on-Linux UA is used because the solver drives a real Firefox.
	SuperFlixUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"
)

// Pre-compiled regexes for SuperFlix scraper
var (
	sfCSRFTokenRe   = regexp.MustCompile(`var CSRF_TOKEN\s*=\s*"([^"]+)"`)
	sfPageTokenRe   = regexp.MustCompile(`var PAGE_TOKEN\s*=\s*"([^"]+)"`)
	sfContentIDRe   = regexp.MustCompile(`var INITIAL_CONTENT_ID\s*=\s*(\d+)`)
	sfContentTypeRe = regexp.MustCompile(`var CONTENT_TYPE\s*=\s*"([^"]+)"`)
	sfTitleRe       = regexp.MustCompile(`<title>(?:Player \| )?(.+?)</title>`)
	sfAllEpisodesRe = regexp.MustCompile(`var ALL_EPISODES\s*=\s*(\{.+?\});`)
	// Current rotating frontend injects the full per-season dataset (with
	// air_date, title, epi_num) as `window.allEpisodes = {...};` and renders the
	// anchors client-side from it. The blob carries metadata the anchors don't.
	sfWindowAllEpisodesRe = regexp.MustCompile(`window\.allEpisodes\s*=\s*(\{.+?\});`)
	sfDefaultAudioRe      = regexp.MustCompile(`var defaultAudio\s*=\s*(\[.+?\]);`)
	sfSubtitleRe          = regexp.MustCompile(`var playerjsSubtitle\s*=\s*"(.+?)";`)
	sfSubPartRe           = regexp.MustCompile(`\[(.+?)\](https?://.+)`)
)

// SuperFlixTokens holds the tokens extracted from a SuperFlix player page
type SuperFlixTokens struct {
	CSRF        string
	PageToken   string
	ContentID   string
	ContentType string
	Title       string
}

// SuperFlixServer represents a streaming server option
type SuperFlixServer struct {
	ID   json.RawMessage `json:"ID"`
	Name string          `json:"name"`
}

// SuperFlixSubtitle represents a subtitle track
type SuperFlixSubtitle struct {
	Lang string
	URL  string
}

// SuperFlixStreamResult holds the final stream extraction result
type SuperFlixStreamResult struct {
	StreamURL    string
	Title        string
	Referer      string
	Subtitles    []SuperFlixSubtitle
	DefaultAudio []string
	Thumb        string
}

// SuperFlixEpisode represents a single episode in a season
type SuperFlixEpisode struct {
	EpiNum  json.Number `json:"epi_num"`
	Title   string      `json:"title"`
	AirDate string      `json:"air_date"`
}

// SuperFlixMedia represents a search result from SuperFlix
type SuperFlixMedia struct {
	Title    string
	Year     string
	Type     string // "Filme", "Série", etc.
	SFType   string // "filme" or "serie"
	TMDBID   string
	IMDBID   string
	ImageURL string // Cover image URL from search results
}

// SuperFlixClient handles interactions with SuperFlix
type SuperFlixClient struct {
	client    *http.Client
	baseURL   string
	userAgent string
	// browserSolver drives the headed browser for episode discovery on the
	// rotating, gated frontend. nil in tests (SetTestConfig) so GetEpisodes
	// falls back to the plain HTTP path against an httptest server.
	browserSolver cfSolver
	maxRetries    int
	retryDelay    time.Duration
	searchCache   sync.Map
}

// NormalizeSuperFlixImageURL converts SuperFlix CloudFront proxy URLs to direct TMDB image URLs.
// Discord's image proxy cannot handle the double-URL format used by SuperFlix:
//
//	https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster.jpg
//
// This extracts the embedded TMDB URL and upgrades to w500 quality:
//
//	https://image.tmdb.org/t/p/w500/poster.jpg
func NormalizeSuperFlixImageURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}
	const tmdbPrefix = "https://image.tmdb.org/t/p/"
	if idx := strings.Index(imageURL, tmdbPrefix); idx > 0 {
		direct := imageURL[idx:]
		// Upgrade thumbnail size for Discord display
		direct = strings.Replace(direct, "/w342/", "/w500/", 1)
		direct = strings.Replace(direct, "/w185/", "/w500/", 1)
		direct = strings.Replace(direct, "/w154/", "/w500/", 1)
		return direct
	}
	return imageURL
}

// NewSuperFlixClient creates a new SuperFlix client.
//
// The HTTP client is wrapped with cfFallbackTransport: on a 403/503/429
// response carrying Cloudflare-challenge markers, the request is replayed
// through a real, headed Firefox (driven via Playwright) to obtain a
// cf_clearance cookie, which is then attached to the retried request and
// every subsequent request for the same host via the cookie jar.
func NewSuperFlixClient() *SuperFlixClient {
	jar, _ := newCookieJar()
	base := safeScraperTransport(30 * time.Second)
	transport := &cfFallbackTransport{
		base:   base,
		solver: defaultCFSolver,
		jar:    jar,
		// Solve budget. The CF gate may need a manual Turnstile checkbox click
		// in the real Chrome window, so allow ~3min. After the first solve the
		// persistent Chrome profile usually clears it automatically in seconds.
		timeout: 180 * time.Second,
	}
	return &SuperFlixClient{
		client: &http.Client{
			// Wall-clock cap on the ENTIRE Do, including a CF browser solve.
			// Must exceed the solve budget above. Fast (non-challenged)
			// requests still return immediately — this only raises the ceiling.
			Timeout:   210 * time.Second,
			Transport: transport,
			Jar:       jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		baseURL:       SuperFlixBase,
		userAgent:     SuperFlixUserAgent,
		browserSolver: defaultCFSolver,
		maxRetries:    2,
		retryDelay:    200 * time.Millisecond,
	}
}

func (c *SuperFlixClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
}

// SetTestConfig overrides the base URL and HTTP client for testing.
// This should only be used in test code.
func (c *SuperFlixClient) SetTestConfig(baseURL string, httpClient *http.Client) {
	c.baseURL = baseURL
	c.client = httpClient
	c.browserSolver = nil // force the plain-HTTP episode path against httptest
	c.maxRetries = 0
	c.retryDelay = 0
}

// SearchMedia searches SuperFlix for movies/series/animes
func (c *SuperFlixClient) SearchMedia(query string) ([]*SuperFlixMedia, error) {
	return c.SearchMediaWithContext(context.Background(), query)
}

// SearchMediaWithContext searches with context support
func (c *SuperFlixClient) SearchMediaWithContext(ctx context.Context, query string) ([]*SuperFlixMedia, error) {
	// CLI args arrive hyphenated like "the-boys" (TreatingAnimeName joins
	// words with dashes), but SuperFlix's search engine treats the dash as
	// a literal character and returns "Nenhum resultado encontrado".
	// Restore spaces so titles like "The Boys" actually match.
	normalized := strings.TrimSpace(query)
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	cacheKey := strings.ToLower(normalized)
	if cached, ok := c.searchCache.Load(cacheKey); ok {
		return cached.([]*SuperFlixMedia), nil
	}

	searchURL := fmt.Sprintf("%s/pesquisar?s=%s", c.baseURL, url.QueryEscape(normalized))
	util.Debug("SuperFlix search", "query", query, "normalized", normalized, "url", searchURL)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	results := c.parseCards(doc)
	c.searchCache.Store(cacheKey, results)
	return results, nil
}

// parseCards extracts media cards from SuperFlix HTML
func (c *SuperFlixClient) parseCards(doc *goquery.Document) []*SuperFlixMedia {
	var results []*SuperFlixMedia
	seen := make(map[string]bool)

	doc.Find("div.group\\/card").Each(func(i int, card *goquery.Selection) {
		var title, imageURL string
		if img := card.Find("img"); img.Length() > 0 {
			title, _ = img.Attr("alt")
			// Extract cover image URL from src, data-src, or srcset
			if src, ok := img.Attr("src"); ok && src != "" && !strings.HasPrefix(src, "data:") {
				imageURL = src
			}
			if imageURL == "" {
				if dataSrc, ok := img.Attr("data-src"); ok && dataSrc != "" {
					imageURL = dataSrc
				}
			}
			if imageURL == "" {
				if srcset, ok := img.Attr("srcset"); ok && srcset != "" {
					// Take the first URL from srcset (format: "url size, url size, ...")
					parts := strings.Fields(strings.Split(srcset, ",")[0])
					if len(parts) > 0 {
						imageURL = parts[0]
					}
				}
			}
		}
		if title == "" {
			if h3 := card.Find("h3"); h3.Length() > 0 {
				title = strings.TrimSpace(h3.Text())
			}
		}
		if title == "" {
			return
		}

		var tmdbID, imdbID, linkURL string

		card.Find("button").Each(func(_ int, btn *goquery.Selection) {
			msg, _ := btn.Attr("data-msg")
			copyVal, _ := btn.Attr("data-copy")
			switch {
			case strings.Contains(msg, "TMDB"):
				tmdbID = copyVal
			case strings.Contains(msg, "IMDB"):
				imdbID = copyVal
			case strings.Contains(msg, "Link"):
				linkURL = copyVal
			}
		})

		// Extract type and year from metadata.
		// The div.mt-3 contains child spans: one for the year (e.g. "2017")
		// and one for the type (e.g. "Anime", "Filme", "Série").
		// Older HTML used "|" separators; current HTML uses separate <span>s.
		var tipo, year string
		card.Find("div.mt-3 span").Each(func(_ int, span *goquery.Selection) {
			text := strings.TrimSpace(span.Text())
			if text == "" {
				return
			}
			// A 4-digit number starting with 1 or 2 is a year.
			if len(text) == 4 && (text[0] == '1' || text[0] == '2') {
				if _, err := strconv.Atoi(text); err == nil {
					year = text
					return
				}
			}
			tipo = text
		})
		// Fallback: try legacy "|"-separated format inside div.mt-3.
		if tipo == "" && year == "" {
			metaText := strings.TrimSpace(card.Find("div.mt-3").Text())
			metaParts := splitAndTrim(metaText, "|")
			if len(metaParts) > 0 {
				tipo = metaParts[len(metaParts)-1]
			}
			if len(metaParts) > 1 {
				year = metaParts[1]
			}
		}

		sfType := "serie"
		if strings.Contains(linkURL, "/filme/") {
			sfType = "filme"
		}

		key := tmdbID
		if key == "" {
			key = title
		}
		if seen[key] {
			return
		}
		seen[key] = true

		if tipo == "" {
			if sfType == "filme" {
				tipo = "Filme"
			} else {
				tipo = "Série"
			}
		}

		results = append(results, &SuperFlixMedia{
			Title:    title,
			Year:     year,
			Type:     tipo,
			SFType:   sfType,
			TMDBID:   tmdbID,
			IMDBID:   imdbID,
			ImageURL: NormalizeSuperFlixImageURL(imageURL),
		})
	})

	return results
}

// GetPlayerPage loads the player page for a given content
func (c *SuperFlixClient) GetPlayerPage(ctx context.Context, mediaType, mediaID, season, episode string) (string, error) {
	path := fmt.Sprintf("/%s/%s", mediaType, mediaID)
	if season != "" {
		path += "/" + season
	}
	if episode != "" {
		path += "/" + episode
	}

	pageURL := c.baseURL + path
	util.Debug("SuperFlix player page", "url", pageURL)

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// ExtractTokens extracts CSRF_TOKEN, PAGE_TOKEN, etc. from player HTML
func (c *SuperFlixClient) ExtractTokens(html string) *SuperFlixTokens {
	tokens := &SuperFlixTokens{}
	if m := sfCSRFTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.CSRF = m[1]
	}
	if m := sfPageTokenRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.PageToken = m[1]
	}
	if m := sfContentIDRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentID = m[1]
	}
	if m := sfContentTypeRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.ContentType = m[1]
	}
	if m := sfTitleRe.FindStringSubmatch(html); len(m) > 1 {
		tokens.Title = m[1]
	}
	return tokens
}

// ExtractEpisodes extracts episodes from a SuperFlix player/serie page.
//
// Two formats are supported:
//   - Legacy player page: a `var ALL_EPISODES = {...}` JS object (air-date
//     filtered, since it can contain unreleased placeholders).
//   - Current rotating frontend (superflix.bond / primeflix.mom /
//     lospobreflix.site / …): episodes rendered as
//     `<a data-season data-episode data-episode-id href="/episodio/...">`.
//     These are already release-filtered by the site, so no air-date pass.
func (c *SuperFlixClient) ExtractEpisodes(html string) (map[string][]SuperFlixEpisode, error) {
	if m := sfAllEpisodesRe.FindStringSubmatch(html); len(m) >= 2 {
		var result map[string][]SuperFlixEpisode
		if err := json.Unmarshal([]byte(m[1]), &result); err != nil {
			return nil, fmt.Errorf("failed to parse ALL_EPISODES: %w", err)
		}
		return filterEpisodesByAirDate(result, time.Now()), nil
	}

	if blob := parseWindowAllEpisodes(html); len(blob) > 0 {
		return blob, nil
	}

	if fe := parseFrontendEpisodes(html); len(fe) > 0 {
		return fe, nil
	}
	return nil, nil
}

// parseWindowAllEpisodes reads the `window.allEpisodes = {...};` blob the current
// rotating frontend injects. Unlike the rendered anchors (which expose only
// season/episode numbers), the blob carries every season at once with full
// metadata — title, epi_num and air_date. Episodes are air-date filtered the
// same way as the legacy ALL_EPISODES blob, since the dataset can include
// unreleased placeholders.
func parseWindowAllEpisodes(html string) map[string][]SuperFlixEpisode {
	m := sfWindowAllEpisodesRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil
	}
	var result map[string][]SuperFlixEpisode
	if err := json.Unmarshal([]byte(m[1]), &result); err != nil {
		util.Debug("SuperFlix: failed to parse window.allEpisodes", "err", err)
		return nil
	}
	return filterEpisodesByAirDate(result, time.Now())
}

// parseFrontendEpisodes reads episodes from the rotating SuperFlix frontend
// serie page. Each `<a data-episode-id>` anchor carries the season and episode
// numbers we need to build the player URL later. Only the currently-loaded
// season's episodes are present on a given page (other seasons live at
// /serie/<slug>/<n>); GetEpisodes fetches those separately and merges.
func parseFrontendEpisodes(html string) map[string][]SuperFlixEpisode {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	out := make(map[string][]SuperFlixEpisode)
	doc.Find("a[data-episode-id]").Each(func(_ int, a *goquery.Selection) {
		season, _ := a.Attr("data-season")
		epnum, _ := a.Attr("data-episode")
		if season == "" || epnum == "" {
			return
		}
		out[season] = append(out[season], SuperFlixEpisode{
			EpiNum: json.Number(epnum),
			Title:  "Episódio " + epnum,
		})
	})
	return out
}

// parseFrontendSeasons returns the distinct season numbers linked on a frontend
// serie page (the season dropdown / "/serie/<slug>/<n>" links).
func parseFrontendSeasons(html string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var seasons []string
	re := regexp.MustCompile(`/serie/[a-z0-9-]+/(\d+)$`)
	doc.Find(`a[href]`).Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		href = strings.SplitN(href, "?", 2)[0]
		href = strings.SplitN(href, "#", 2)[0]
		if mm := re.FindStringSubmatch(href); len(mm) > 1 {
			if !seen[mm[1]] {
				seen[mm[1]] = true
				seasons = append(seasons, mm[1])
			}
		}
	})
	sort.Slice(seasons, func(i, j int) bool {
		ai, _ := strconv.Atoi(seasons[i])
		aj, _ := strconv.Atoi(seasons[j])
		return ai < aj
	})
	return seasons
}

// filterEpisodesByAirDate drops episodes with empty/"null" air_date and
// episodes whose air_date is strictly after the current UTC day.
//
// Comparison is done at day granularity in UTC so the result does not drift
// across midnight: a previous version used `t.After(now.Add(24*time.Hour))`,
// which kept tomorrow's episodes any time `now`'s UTC time-of-day was past
// 00:00 — flaky for any caller running with `now` in a timezone west of UTC.
func filterEpisodesByAirDate(result map[string][]SuperFlixEpisode, now time.Time) map[string][]SuperFlixEpisode {
	utcNow := now.UTC()
	today := time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day(), 0, 0, 0, 0, time.UTC)
	for season, episodes := range result {
		var validEpisodes []SuperFlixEpisode
		for _, ep := range episodes {
			if ep.AirDate == "" || ep.AirDate == "null" {
				continue
			}
			if t, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
				if t.After(today) {
					continue
				}
			}
			validEpisodes = append(validEpisodes, ep)
		}
		result[season] = validEpisodes
	}
	return result
}

// Bootstrap calls /player/bootstrap to get server list
func (c *SuperFlixClient) Bootstrap(ctx context.Context, tokens *SuperFlixTokens) ([]SuperFlixServer, error) {
	bootstrapURL := c.baseURL + "/player/bootstrap"

	form := url.Values{
		"contentid":  {tokens.ContentID},
		"type":       {tokens.ContentType},
		"_token":     {tokens.CSRF},
		"page_token": {tokens.PageToken},
		"pageToken":  {tokens.PageToken},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("bootstrap", resp, body); err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Options []SuperFlixServer `json:"options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode bootstrap response: %w", err)
	}

	return result.Data.Options, nil
}

// GetSourceURL calls /player/source to get the redirect URL for a video
func (c *SuperFlixClient) GetSourceURL(ctx context.Context, videoID string, tokens *SuperFlixTokens) (string, error) {
	sourceURL := c.baseURL + "/player/source"

	form := url.Values{
		"video_id":   {videoID},
		"page_token": {tokens.PageToken},
		"host":       {""},
		"site":       {""},
		"_token":     {tokens.CSRF},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sourceURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("X-Page-Token", tokens.PageToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", c.baseURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("source", resp, body); err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			VideoURL string `json:"video_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode source response: %w", err)
	}

	if result.Data.VideoURL == "" {
		return "", fmt.Errorf("no video URL in source response")
	}

	return result.Data.VideoURL, nil
}

// ResolveRedirect follows the SuperFlix redirect to get the external player URL
func (c *SuperFlixClient) ResolveRedirect(ctx context.Context, redirectURL string) (baseURL, videoHash, playerHTML string, err error) {
	// Use the client's transport if available, otherwise fall back to safe transport
	transport := c.client.Transport
	if transport == nil {
		transport = safeScraperTransport(30 * time.Second)
	}

	// Use a client that does NOT follow redirects automatically
	noRedirectClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", redirectURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", c.baseURL+"/")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	location := redirectURL
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location = resp.Header.Get("Location")
		if location == "" {
			return "", "", "", fmt.Errorf("redirect with no Location header")
		}
	}

	// Follow to the final page
	req2, err := http.NewRequestWithContext(ctx, "GET", location, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create follow request: %w", err)
	}
	c.decorateRequest(req2)
	req2.Header.Set("Referer", c.baseURL+"/")

	followClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	resp2, err := followClient.Do(req2)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to follow redirect: %w", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp2.Body, 5*1024*1024))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read player page: %w", err)
	}

	finalURL := resp2.Request.URL.String()

	if strings.Contains(finalURL, "/video/") {
		parts := strings.SplitN(finalURL, "/video/", 2)
		baseURL = parts[0]
		videoHash = strings.SplitN(parts[1], "?", 2)[0]
		videoHash = strings.SplitN(videoHash, "#", 2)[0]
	} else {
		idx := strings.LastIndex(finalURL, "/")
		if idx > 0 {
			baseURL = finalURL[:idx]
			videoHash = strings.SplitN(finalURL[idx+1:], "?", 2)[0]
		}
	}

	return baseURL, videoHash, string(body), nil
}

// ExtractPlayerExtras extracts defaultAudio and subtitles from the external player HTML
func (c *SuperFlixClient) ExtractPlayerExtras(html string) (defaultAudio []string, subtitles []SuperFlixSubtitle) {
	if m := sfDefaultAudioRe.FindStringSubmatch(html); len(m) > 1 {
		_ = json.Unmarshal([]byte(m[1]), &defaultAudio)
	}

	if m := sfSubtitleRe.FindStringSubmatch(html); len(m) > 1 {
		for part := range strings.SplitSeq(m[1], ",") {
			sm := sfSubPartRe.FindStringSubmatch(part)
			if len(sm) > 2 {
				subtitles = append(subtitles, SuperFlixSubtitle{
					Lang: sm[1],
					URL:  sm[2],
				})
			}
		}
	}
	return
}

// GetVideoAPI calls the external player's API to get the actual stream URL
func (c *SuperFlixClient) GetVideoAPI(ctx context.Context, playerBaseURL, videoHash, referer string) (streamURL, thumbURL string, err error) {
	apiURL := fmt.Sprintf("%s/player/index.php?data=%s&do=getVideo", playerBaseURL, videoHash)

	form := url.Values{
		"hash": {videoHash},
		"r":    {c.baseURL + "/"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if err := ensureJSONResponse("video API", resp, body); err != nil {
		return "", "", err
	}

	var result struct {
		SecuredLink string `json:"securedLink"`
		VideoSource string `json:"videoSource"`
		VideoImage  string `json:"videoImage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to decode video API response: %w", err)
	}

	switch {
	case result.SecuredLink != "":
		streamURL = result.SecuredLink
	case result.VideoSource != "":
		streamURL = result.VideoSource
	default:
		return "", "", fmt.Errorf("no stream URL in video API response")
	}

	return streamURL, result.VideoImage, nil
}

// embedStreamSolver is the subset of the browser solver used to extract a live
// stream by driving the warezcdn embed through its Turnstile gate and capturing
// the player's getVideo response. The transport's cfSolver interface stays
// minimal (Solve only); only the real *cfBrowserSolver implements this, so the
// type assertion in GetStreamURL is false for test fakes / a nil solver.
type embedStreamSolver interface {
	SniffEmbedStream(ctx context.Context, embedURL string, timeout time.Duration) (*CFStreamResult, error)
}

// GetStreamURL resolves the playable stream for SuperFlix content.
//
// The legacy player-page→tokens→bootstrap→source pipeline is dead: the current
// site serves a Turnstile-gated embed with no inline tokens. So in production
// (browser solver present) we drive the embed through the gate and sniff the
// player's getVideo response for the signed HLS master. The legacy pipeline
// below is retained only for the httptest-backed unit tests (which null out the
// browser solver via SetTestConfig).
func (c *SuperFlixClient) GetStreamURL(ctx context.Context, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	if solver, ok := c.browserSolver.(embedStreamSolver); ok {
		return c.getStreamViaBrowser(ctx, solver, mediaType, mediaID, season, episode)
	}

	html, err := c.GetPlayerPage(ctx, mediaType, mediaID, season, episode)
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}

	tokens := c.ExtractTokens(html)
	if tokens.CSRF == "" || tokens.PageToken == "" {
		return nil, fmt.Errorf("failed to extract tokens from player page")
	}

	servers, err := c.Bootstrap(ctx, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap: %w", err)
	}
	if len(servers) == 0 {
		// Empty bootstrap on a fully-loaded player page means SuperFlix has
		// no provider for this specific content (typically placeholder
		// episodes whose `air_date` is null in ALL_EPISODES). The upstream
		// site renders a "not yet released" screen in the same case.
		// Annotate the error with the player URL and contentid so triage
		// doesn't confuse this with a network or scraping failure.
		playerPath := fmt.Sprintf("/%s/%s", mediaType, mediaID)
		if season != "" {
			playerPath += "/" + season
		}
		if episode != "" {
			playerPath += "/" + episode
		}
		return nil, fmt.Errorf("%w (url=%s%s, contentid=%s) — try another episode or source",
			ErrSuperFlixNoServers, c.baseURL, playerPath, tokens.ContentID)
	}

	// Pick first non-fallback server
	var videoIDStr string
	for _, s := range servers {
		var raw string
		if err := json.Unmarshal(s.ID, &raw); err == nil {
			if !strings.HasPrefix(raw, "fallback") {
				videoIDStr = raw
				break
			}
		}
		// Try as number
		var num json.Number
		if err := json.Unmarshal(s.ID, &num); err == nil {
			videoIDStr = num.String()
			break
		}
	}
	if videoIDStr == "" {
		// Fallback: use first server
		var raw string
		if err := json.Unmarshal(servers[0].ID, &raw); err == nil {
			videoIDStr = raw
		} else {
			var num json.Number
			if err := json.Unmarshal(servers[0].ID, &num); err == nil {
				videoIDStr = num.String()
			} else {
				return nil, fmt.Errorf("failed to parse server ID")
			}
		}
	}

	redirectURL, err := c.GetSourceURL(ctx, videoIDStr, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get source URL: %w", err)
	}

	playerBaseURL, videoHash, playerHTML, err := c.ResolveRedirect(ctx, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve redirect: %w", err)
	}

	referer := fmt.Sprintf("%s/video/%s", playerBaseURL, videoHash)
	streamURL, thumbURL, err := c.GetVideoAPI(ctx, playerBaseURL, videoHash, referer)
	if err != nil {
		return nil, fmt.Errorf("failed to get video from API: %w", err)
	}

	result := &SuperFlixStreamResult{
		StreamURL: streamURL,
		Title:     tokens.Title,
		Referer:   playerBaseURL + "/",
		Thumb:     NormalizeSuperFlixImageURL(thumbURL),
	}

	defaultAudio, subtitles := c.ExtractPlayerExtras(playerHTML)
	result.DefaultAudio = defaultAudio
	result.Subtitles = subtitles

	return result, nil
}

// getStreamViaBrowser resolves the stream, preferring a browser-free path.
//
// The only browser-gated step is mapping tmdb→(playerHost, videoHash) through
// warezcdn's Turnstile gate. The player host's getVideo endpoint that turns that
// pair into a fresh signed HLS link is NOT gated, so once the pair is cached we
// replay over plain HTTP with no browser. The headed browser therefore runs only
// on the FIRST play of a title — or when the cached host rotates out and the
// HTTP getVideo fails, which transparently falls back to a re-solve.
func (c *SuperFlixClient) getStreamViaBrowser(ctx context.Context, solver embedStreamSolver, mediaType, mediaID, season, episode string) (*SuperFlixStreamResult, error) {
	key := streamCacheKey(mediaType, mediaID, season, episode)

	// 1. Cached (host, hash) → pure-HTTP getVideo, no browser.
	if ent, ok := defaultStreamCache.get(key); ok {
		referer := ent.Host + "/video/" + ent.Hash
		streamURL, thumb, err := c.GetVideoAPI(ctx, ent.Host, ent.Hash, referer)
		if err == nil && streamURL != "" {
			util.Debug("SuperFlix stream from cache (no browser)", "key", key, "host", ent.Host)
			return &SuperFlixStreamResult{
				StreamURL: streamURL,
				Referer:   ent.Host + "/",
				Thumb:     NormalizeSuperFlixImageURL(thumb),
			}, nil
		}
		util.Debug("SuperFlix cached stream stale, re-solving", "key", key, "err", err)
	}

	// 2. Cache miss / stale → drive the headed browser through the gate once,
	//    capture the stream + (host, hash), and cache the pair for next time.
	var embedURL string
	if mediaType == "serie" {
		s, e := season, episode
		if s == "" {
			s = "1"
		}
		if e == "" {
			e = "1"
		}
		embedURL = fmt.Sprintf("https://%s/serie/%s/%s/%s", SuperFlixEmbedHost, mediaID, s, e)
	} else {
		embedURL = fmt.Sprintf("https://%s/filme/%s", SuperFlixEmbedHost, mediaID)
	}

	res, err := solver.SniffEmbedStream(ctx, embedURL, 0)
	if err != nil {
		return nil, fmt.Errorf("superflix embed stream sniff failed (%s): %w", embedURL, err)
	}

	defaultStreamCache.put(key, streamCacheEntry{Host: res.PlayerHost, Hash: res.VideoHash})

	referer := res.Referer
	if referer == "" {
		referer = "https://" + SuperFlixEmbedHost + "/"
	}
	return &SuperFlixStreamResult{
		StreamURL: res.StreamURL,
		Referer:   referer,
	}, nil
}

// GetEpisodes returns the seasons and episodes for a series.
//
// Legacy player pages embed every season in one ALL_EPISODES blob. The current
// rotating frontend renders only the loaded season's episodes per page and
// links the others at /serie/<slug>/<n>, so when we detect that format we fetch
// each remaining season (via the gateway serie/<tmdb>/<n>, which redirects to
// the right frontend season) and merge. Per-season fetches reuse the already
// cleared CF profile, so they don't re-trigger the challenge.
func (c *SuperFlixClient) GetEpisodes(ctx context.Context, tmdbID string) (map[string][]SuperFlixEpisode, error) {
	// Production path: drive the browser solver directly. It returns the final
	// (rotating) frontend URL, which we need to resolve the per-season links
	// onto the right domain. The transport/HTTP path can't expose that.
	if c.browserSolver != nil {
		return c.getEpisodesViaBrowser(ctx, tmdbID)
	}

	// Test path (SetTestConfig): plain HTTP against an httptest server.
	html, err := c.GetPlayerPage(ctx, "serie", tmdbID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to load player page: %w", err)
	}
	return c.ExtractEpisodes(html)
}

// getEpisodesViaBrowser solves the serie page, parses the loaded season, then
// solves each remaining season's frontend URL and merges. Per-season solves
// reuse the warm CF profile, so they don't re-trigger the challenge.
func (c *SuperFlixClient) getEpisodesViaBrowser(ctx context.Context, tmdbID string) (map[string][]SuperFlixEpisode, error) {
	base := strings.TrimSuffix(c.baseURL, "/")
	res, err := c.browserSolver.Solve(ctx, base+"/serie/"+tmdbID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load serie page: %w", err)
	}

	episodes := make(map[string][]SuperFlixEpisode)

	// Legacy ALL_EPISODES (rare now) carries every season in one shot.
	if legacy, lErr := c.ExtractEpisodes(res.HTML); lErr == nil && sfAllEpisodesRe.MatchString(res.HTML) {
		return legacy, nil
	}

	// Current frontend injects every season with air_date in window.allEpisodes,
	// so a single solve covers all seasons — no per-season fetch needed.
	if blob := parseWindowAllEpisodes(res.HTML); len(blob) > 0 {
		return blob, nil
	}

	for s, eps := range parseFrontendEpisodes(res.HTML) {
		episodes[s] = eps
	}

	// Resolve the other seasons' URLs against the solved frontend domain and
	// fetch each that we don't already have.
	for season, seasonURL := range resolveFrontendSeasonURLs(res.HTML, res.FinalURL) {
		if _, ok := episodes[season]; ok {
			continue
		}
		sres, sErr := c.browserSolver.Solve(ctx, seasonURL, 0)
		if sErr != nil {
			util.Debug("SuperFlix: failed to load season page", "season", season, "url", seasonURL, "err", sErr)
			continue
		}
		for s, eps := range parseFrontendEpisodes(sres.HTML) {
			if _, ok := episodes[s]; !ok {
				episodes[s] = eps
			}
		}
	}

	if len(episodes) == 0 {
		return nil, nil
	}
	return episodes, nil
}

// resolveFrontendSeasonURLs maps season number -> absolute URL for every
// /serie/<slug>/<n> link on a frontend serie page, resolved against the page's
// final (post-redirect) URL so they hit the correct rotating domain.
func resolveFrontendSeasonURLs(html, finalURL string) map[string]string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	var basePtr *url.URL
	if finalURL != "" {
		basePtr, _ = url.Parse(finalURL)
	}
	re := regexp.MustCompile(`/serie/[a-z0-9-]+/(\d+)$`)
	out := make(map[string]string)
	doc.Find(`a[href]`).Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		clean := strings.SplitN(href, "?", 2)[0]
		clean = strings.SplitN(clean, "#", 2)[0]
		m := re.FindStringSubmatch(clean)
		if m == nil {
			return
		}
		season := m[1]
		if _, ok := out[season]; ok {
			return
		}
		abs := clean
		if basePtr != nil {
			if ref, perr := url.Parse(clean); perr == nil {
				abs = basePtr.ResolveReference(ref).String()
			}
		}
		out[season] = abs
	})
	return out
}

// ToAnimeModel converts SuperFlixMedia to models.Anime for compatibility
func (m *SuperFlixMedia) ToAnimeModel() *models.Anime {
	anime := &models.Anime{
		Name:     m.Title,
		URL:      m.TMDBID, // Store TMDB ID as URL identifier
		Source:   "SuperFlix",
		Year:     m.Year,
		ImageURL: m.ImageURL,
	}

	lowerType := strings.ToLower(m.Type)
	switch {
	case m.SFType == "filme":
		anime.MediaType = models.MediaTypeMovie
	case lowerType == "anime" || lowerType == "dorama":
		anime.MediaType = models.MediaTypeAnime
	default:
		anime.MediaType = models.MediaTypeTV
	}

	if m.IMDBID != "" {
		anime.IMDBID = m.IMDBID
	}

	// Parse TMDB ID for direct API lookups during enrichment
	if m.TMDBID != "" {
		if id, err := strconv.Atoi(m.TMDBID); err == nil {
			anime.TMDBID = id
		}
	}

	util.Debug("SuperFlix ToAnimeModel", "title", m.Title, "tmdbID", m.TMDBID, "imageURL", anime.ImageURL)

	return anime
}

// ensureJSONResponse fails fast when a SuperFlix API endpoint replies with an
// HTML body or a non-2xx status. Without this, callers get the unhelpful
// `invalid character '<' looking for beginning of value` JSON error — which
// hides real causes like the host having moved (the .rest → .online 301 that
// silently downgrades POST → GET) or a Cloudflare/captcha interstitial.
//
// Trust the body, not the Content-Type header. Some upstream players (e.g.
// firevideoplayer.com behind llanfairpwllgwyngy.com) serve real JSON with
// `Content-Type: text/html`, so a header-only check would reject valid
// responses.
func ensureJSONResponse(label string, resp *http.Response, body []byte) error {
	trimmed := strings.TrimLeft(string(body), " \t\r\n\ufeff")
	looksHTML := len(trimmed) > 0 && trimmed[0] == '<'

	if looksHTML {
		finalURL := ""
		if resp.Request != nil && resp.Request.URL != nil {
			finalURL = resp.Request.URL.String()
		}
		return fmt.Errorf("%s endpoint returned HTML (status %d, url=%q) — provider may have moved or is blocking the request", label, resp.StatusCode, finalURL)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s endpoint returned status %d", label, resp.StatusCode)
	}
	return nil
}

// Helper: split string by separator and trim each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
