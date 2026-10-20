package animedrive

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	AnimeDriveBase = "https://animesdrive.cloud"
)

// Pre-compiled regexes
var (
	animeDriveEpRe    = regexp.MustCompile(`(?i)episodios?[-_]?(\d+)`)
	animeDriveDigitRe = regexp.MustCompile(`(\d+)`)
)

// AnimeDriveEpisode represents an episode from AnimeDrive
type AnimeDriveEpisode struct {
	Number string
	Title  string
	URL    string
}

// AnimeDriveDetails represents anime details including episodes
type AnimeDriveDetails struct {
	ID        string
	Title     string
	URL       string
	Thumbnail string
	Synopsis  string
	Episodes  []AnimeDriveEpisode
}

// AnimeDriveClient handles interactions with animesdrive.cloud
type AnimeDriveClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
}

// NewAnimeDriveClient creates a new AnimeDrive client
func NewAnimeDriveClient() *AnimeDriveClient {
	return &AnimeDriveClient{
		client:     util.NewFastClient(),
		baseURL:    AnimeDriveBase,
		userAgent:  netx.UserAgent,
		maxRetries: 2,
		retryDelay: 100 * time.Millisecond,
	}
}

// NewClientForTest returns a client pointed at a test server with retries only for tests
func NewClientForTest(serverURL string) *AnimeDriveClient {
	c := NewAnimeDriveClient()
	c.baseURL = serverURL
	c.maxRetries = 0
	c.retryDelay = 0
	return c
}

// SetHTTPClientForTest swaps the underlying HTTP client. Only for tests that
// need to mock transports (e.g. live-like blogger 200 HTML) while driving the
// real client through adapter/provider cascades.
func (c *AnimeDriveClient) SetHTTPClientForTest(client *http.Client) {
	c.client = client
}

// isAvailableStreamURL rejects stale direct links before handing them to mpv.
// AnimeDrive sometimes leaves dead ccdn.xyz URLs in the episode HTML.
func (c *AnimeDriveClient) isAvailableStreamURL(streamURL string) bool {
	req, err := http.NewRequest(http.MethodGet, streamURL, http.NoBody)
	if err != nil {
		return false
	}

	req.Header.Set("Range", "bytes=0-0")
	c.decorateRequest(req)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (c *AnimeDriveClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", netx.AcceptLanguage)
	req.Header.Set("Referer", strings.TrimRight(c.baseURL, "/")+"/")
}

func (c *AnimeDriveClient) shouldRetry(attempt int) bool {
	return attempt < c.maxRetries
}

func (c *AnimeDriveClient) sleep() {
	if c.retryDelay > 0 {
		time.Sleep(c.retryDelay)
	}
}

func (c *AnimeDriveClient) extractIDFromURL(urlStr string) string {
	cleanURL := strings.TrimSuffix(urlStr, "/")
	parts := strings.Split(cleanURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// SearchAnime searches for anime on AnimeDrive
func (c *AnimeDriveClient) SearchAnime(query string) ([]*models.Anime, error) {
	return c.SearchAnimeContext(context.Background(), query)
}

// SearchAnimeContext makes search requests and retries cancellable.
func (c *AnimeDriveClient) SearchAnimeContext(ctx context.Context, query string) ([]*models.Anime, error) {
	// Normalize query: replace hyphens/underscores with spaces for WordPress search
	normalizedQuery := strings.TrimSpace(query)
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "-", " ")
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "_", " ")
	// Collapse multiple spaces
	for strings.Contains(normalizedQuery, "  ") {
		normalizedQuery = strings.ReplaceAll(normalizedQuery, "  ", " ")
	}

	searchURL := c.baseURL + "/?s=" + url.QueryEscape(normalizedQuery)

	util.Debug("AnimeDrive search", "query", query, "normalized", normalizedQuery, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := range attempts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, "GET", searchURL, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req) // #nosec G704
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			if c.shouldRetry(attempt) {
				if err := c.sleepContext(ctx); err != nil {
					return nil, err
				}
				continue
			}
			return nil, lastErr
		}

		if err := netx.CheckHTTPStatus(resp, "AnimeDrive search"); err != nil {
			lastErr = err
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				if err := c.sleepContext(ctx); err != nil {
					return nil, err
				}
				continue
			}
			return nil, lastErr
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to parse HTML: %w", err)
			if c.shouldRetry(attempt) {
				if err := c.sleepContext(ctx); err != nil {
					return nil, err
				}
				continue
			}
			return nil, lastErr
		}

		if err := netx.CheckChallengeDocument(doc, "AnimeDrive search"); err != nil {
			lastErr = err
			if c.shouldRetry(attempt) {
				if err := c.sleepContext(ctx); err != nil {
					return nil, err
				}
				continue
			}
			return nil, lastErr
		}

		animes := c.extractSearchResults(doc)
		util.Debug("AnimeDrive search results", "count", len(animes))
		return animes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve results from AnimeDrive")
}

func (c *AnimeDriveClient) sleepContext(ctx context.Context) error {
	if c.retryDelay <= 0 {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.retryDelay):
		return nil
	}
}

func (c *AnimeDriveClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	var animes []*models.Anime

	// Search for result cards - multiple selectors to cover theme variations
	// Priority 1: div.result-item article (DooPlay search results)
	// Priority 2: article.w_item_a (widget/featured items)
	// Priority 3: legacy selectors
	selectors := []string{
		"div.result-item article",
		"article.w_item_a",
		"article.item",
		"div.search-page .item",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			var title, urlPath, imgURL string

			// Try DooPlay search result structure: div.title > a
			titleLink := s.Find("div.title a, div.details .title a").First()
			if titleLink.Length() > 0 {
				title = strings.TrimSpace(titleLink.Text())
				urlPath, _ = titleLink.Attr("href")
			}

			// Fallback: direct link + h3 structure (w_item_a / widget)
			if urlPath == "" {
				linkElement := s.Find("a").First()
				urlPath, _ = linkElement.Attr("href")
				titleEl := s.Find("h3, h2, .data h3").First()
				title = strings.TrimSpace(titleEl.Text())
				if title == "" {
					title, _ = linkElement.Attr("title")
					title = strings.TrimSpace(title)
				}
			}

			if urlPath == "" || !strings.Contains(urlPath, "/anime/") {
				return
			}
			if title == "" {
				return
			}

			imageElement := s.Find("img").First()
			imgURL, _ = imageElement.Attr("src")
			if imgURL == "" {
				imgURL, _ = imageElement.Attr("data-src")
			}
			if imgURL == "" {
				imgURL, _ = imageElement.Attr("data-lazy-src")
			}

			animes = append(animes, &models.Anime{
				Name:     title,
				URL:      c.resolveURL(urlPath),
				ImageURL: imgURL,
			})
		})

		if len(animes) > 0 {
			break
		}
	}

	return animes
}

// GetAnimeDetails fetches anime details including episode list
func (c *AnimeDriveClient) GetAnimeDetails(animeURL string) (*AnimeDriveDetails, error) {
	util.Debug("AnimeDrive getting details", "url", animeURL)

	urlStr := c.resolveURL(animeURL)

	req, err := http.NewRequest("GET", urlStr, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AnimeDrive details"); err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, netx.NewParserError("AnimeDrive", "details", "failed to parse HTML", err)
	}

	if err := netx.CheckChallengeDocument(doc, "AnimeDrive details"); err != nil {
		return nil, err
	}

	// Extract title - prioritize specific DooPlay selectors over generic h1
	titleElement := doc.Find(".sheader .data h1, h1.entry-title").First()
	title := strings.TrimSpace(titleElement.Text())
	if title == "" {
		title = doc.Find("h1").First().Text()
		title = strings.TrimSpace(title)
	}
	if title == "" {
		title = "Unknown"
	}

	// Extract poster image
	posterElement := doc.Find(".poster img, .sheader .poster img, img.wp-post-image").First()
	poster, _ := posterElement.Attr("src")
	if poster == "" {
		poster, _ = posterElement.Attr("data-src")
	}

	// Extract synopsis
	synopsisElement := doc.Find(".wp-content p, .description p, #info .wp-content").First()
	synopsis := strings.TrimSpace(synopsisElement.Text())

	// Extract episodes
	var episodes []AnimeDriveEpisode
	seenEpisodeURLs := make(map[string]struct{})
	appendEpisode := func(epURL, epTitle, epNumber string) {
		epURL = strings.TrimSpace(epURL)
		if epURL == "" || !strings.Contains(strings.ToLower(epURL), "episodio") {
			return
		}
		if _, seen := seenEpisodeURLs[epURL]; seen {
			return
		}
		seenEpisodeURLs[epURL] = struct{}{}

		epTitle = strings.TrimSpace(epTitle)
		epNumber = strings.TrimSpace(epNumber)
		if epNumber == "" || epNumber == "0" {
			if match := animeDriveEpRe.FindStringSubmatch(epURL); len(match) > 1 {
				epNumber = match[1]
			} else if match := animeDriveDigitRe.FindStringSubmatch(epTitle); len(match) > 1 {
				epNumber = match[1]
			} else {
				epNumber = "0"
			}
		}
		episodes = append(episodes, AnimeDriveEpisode{Number: epNumber, Title: epTitle, URL: epURL})
	}

	doc.Find(".episode-info").Each(func(_ int, info *goquery.Selection) {
		titleLink := info.Find(".episode-title a[href]").First()
		if titleLink.Length() == 0 {
			titleLink = info.Find("h3 a[href]").First()
		}
		miniNumber := info.Parent().Find(".episode-mini-number").First()
		epNumber := strings.TrimSpace(miniNumber.Text())
		if epNumber == "" {
			if match := animeDriveDigitRe.FindStringSubmatch(miniNumber.AttrOr("aria-label", "")); len(match) > 1 {
				epNumber = match[1]
			}
		}
		appendEpisode(titleLink.AttrOr("href", ""), titleLink.Text(), epNumber)
	})

	if len(episodes) == 0 {
		episodeSelectors := ".episode-card, #seasons .episodios li a, .episodios li a, ul.episodios a, .se-a a, #episodes a, .episodelist a, [data-episode-number], [data-episode-title]"
		doc.Find(episodeSelectors).Each(func(_ int, item *goquery.Selection) {
			link := item
			if !item.Is("a") {
				link = item.Find("h3.episode-title a[href], a[href]").First()
			}
			if link.Length() == 0 {
				return
			}
			epTitle := strings.TrimSpace(link.Text())
			if epTitle == "" {
				epTitle = item.AttrOr("data-episode-title", "")
			}
			appendEpisode(link.AttrOr("href", ""), epTitle, item.AttrOr("data-episode-number", ""))
		})
	}

	// Sort episodes by number
	sort.Slice(episodes, func(i, j int) bool {
		numA, _ := strconv.Atoi(episodes[i].Number)
		numB, _ := strconv.Atoi(episodes[j].Number)
		return numA < numB
	})

	return &AnimeDriveDetails{
		ID:        c.extractIDFromURL(urlStr),
		Title:     title,
		URL:       urlStr,
		Thumbnail: poster,
		Synopsis:  synopsis,
		Episodes:  episodes,
	}, nil
}

// GetAnimeEpisodes converts AnimeDrive episodes to models.Episode format
func (c *AnimeDriveClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("AnimeDrive episodes", "url", animeURL)

	details, err := c.GetAnimeDetails(animeURL)
	if err != nil {
		return nil, err
	}

	var episodes []models.Episode
	for _, ep := range details.Episodes {
		num, _ := strconv.Atoi(ep.Number)
		util.Debug("AnimeDrive episode candidate",
			"number", ep.Number,
			"url", ep.URL,
		)

		episodes = append(episodes, models.Episode{
			Number: ep.Number,
			Num:    num,
			URL:    ep.URL,
			Title: models.TitleDetails{
				Romaji: ep.Title,
			},
		})
	}

	util.Debug("AnimeDrive found episodes", "count", len(episodes))

	return episodes, nil
}

// resolveURL resolves relative URLs to absolute URLs
func (c *AnimeDriveClient) resolveURL(ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	if strings.HasPrefix(ref, "/") {
		return c.baseURL + ref
	}
	return c.baseURL + "/" + ref
}

// GetStreamURL resolves the episode page to the first playable server URL.
// The site marks its preferred server with is-active; quality only reorders
// that list (best prefers FHD > HD > SD, worst reverses it) and never filters
// by audio, so Legendado/Dublado pages keep the site's order.
func (c *AnimeDriveClient) GetStreamURL(episodeURL string) (videoURL string, metadata map[string]string, err error) {
	return c.GetStreamURLWithQuality(episodeURL, "best")
}

// GetStreamURLWithQuality is GetStreamURL with explicit quality ordering:
// "worst" prefers the lowest-ranked server, anything else the highest.
func (c *AnimeDriveClient) GetStreamURLWithQuality(episodeURL, quality string) (videoURL string, metadata map[string]string, err error) {
	util.Debug("AnimeDriveClient stream URL extraction", "episodeURL", episodeURL)

	urlStr := c.resolveURL(episodeURL)

	req, err := http.NewRequest("GET", urlStr, http.NoBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req) // #nosec G704
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := netx.CheckHTTPStatus(resp, "AnimeDrive stream"); err != nil {
		return "", nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil, netx.NewParserError("AnimeDrive", "stream", "failed to parse HTML", err)
	}

	if err := netx.CheckChallengeDocument(doc, "AnimeDrive stream"); err != nil {
		return "", nil, err
	}

	type track struct {
		src   string
		label string
	}
	var tracks []track

	names := make(map[string]string)
	doc.Find("button.animeq-player__server[data-animeq-switch]").Each(func(_ int, button *goquery.Selection) {
		id := button.AttrOr("data-animeq-switch", "")
		names[id] = button.AttrOr("data-source-name", "")
	})

	sources := doc.Find(".animeq-player__source[data-animeq-source]")
	var unsupportedEmbeds []string
	if sources.Length() > 0 {
		// Each server is a .animeq-player__source block numbered by
		// data-animeq-source; the button with the same data-animeq-switch
		// number names it. Blocks mix <video> and <iframe>: only direct
		// videos and Blogger embeds play. Third-party iframes (animeshd,
		// strp2p, rpmplay) are recorded for diagnostics so an embed-only
		// episode fails with the host list instead of a bare parse error.
		sources.Each(func(_ int, source *goquery.Selection) {
			id := source.AttrOr("data-animeq-source", "")
			label := names[id]
			streamURL := source.Find("video source[src]").First().AttrOr("src", "")
			if streamURL == "" {
				streamURL = source.Find("video[src]").First().AttrOr("src", "")
			}
			if streamURL == "" {
				// The player resolves Blogger embeds; animeshd/strp2p/rpmplay
				// iframes are not playable, so only Blogger is collected.
				streamURL = source.Find(`iframe[src*="blogger.com/video"]`).First().AttrOr("src", "")
			}
			// goquery already decodes entities such as &#038; (&) in the
			// ccdn HLS URLs, but unescape explicitly so a raw-extracted URL
			// can never reach the probe with entities intact.
			streamURL = html.UnescapeString(strings.TrimSpace(streamURL))
			if streamURL != "" {
				tracks = append(tracks, track{src: streamURL, label: label})
				return
			}
			if embedSrc := strings.TrimSpace(source.Find("iframe[src]").First().AttrOr("src", "")); embedSrc != "" {
				if host := animeDriveEmbedHost(embedSrc); host != "" {
					unsupportedEmbeds = append(unsupportedEmbeds, host)
				}
			}
		})
	} else {
		// Older pages place all video sources in one video element.
		doc.Find("video.animeq-player__video source[src]").Each(func(i int, source *goquery.Selection) {
			src := html.UnescapeString(strings.TrimSpace(source.AttrOr("src", "")))
			if src == "" {
				return
			}
			label := doc.Find("button.animeq-player__server[data-source-name]").Eq(i).AttrOr("data-source-name", "Legendado")
			tracks = append(tracks, track{src: src, label: label})
		})
	}

	if len(tracks) == 0 {
		if len(unsupportedEmbeds) > 0 {
			hosts := dedupStrings(unsupportedEmbeds)
			return "", nil, netx.NewParserError("AnimeDrive", "stream",
				fmt.Sprintf("no video tracks found in animeq-player (found %d unsupported embeds: %s)", len(unsupportedEmbeds), strings.Join(hosts, ", ")), nil)
		}
		return "", nil, netx.NewParserError("AnimeDrive", "stream", "no video tracks found in animeq-player", nil)
	}

	util.Debug("AnimeDrive found tracks", "count", len(tracks))
	for _, t := range tracks {
		util.Debug("AnimeDrive track", "label", t.label, "src", t.src)
	}

	preferLow := strings.EqualFold(strings.TrimSpace(quality), "worst")
	sort.SliceStable(tracks, func(i, j int) bool {
		left, right := animeDriveQualityRank(tracks[i].label), animeDriveQualityRank(tracks[j].label)
		if preferLow {
			return left < right
		}
		return left > right
	})

	var selected *track
	for i := range tracks {
		validatedURL, validateErr := netx.ValidateStreamURL(tracks[i].src, "AnimeDrive")
		if validateErr != nil || !c.isAvailableStreamURL(validatedURL) {
			continue
		}
		tracks[i].src = validatedURL
		selected = &tracks[i]
		break
	}
	if selected == nil {
		labels := make([]string, 0, len(tracks))
		for _, t := range tracks {
			if t.label != "" {
				labels = append(labels, t.label)
			}
		}
		if len(labels) > 0 {
			return "", nil, fmt.Errorf("none of the %d servers on this episode responded (labels: %s)", len(tracks), strings.Join(dedupStrings(labels), ", "))
		}
		return "", nil, fmt.Errorf("none of the %d servers on this episode responded", len(tracks))
	}

	metadata = map[string]string{
		"source":     "animedrive",
		"audio_lang": selected.label,
		"referer":    strings.TrimRight(c.baseURL, "/") + "/",
	}

	return selected.src, metadata, nil
}

// animeDriveEmbedHost extracts the host of an unsupported embed iframe so
// embed-only failures can name the hosts instead of failing opaquely.
func animeDriveEmbedHost(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return ""
	}
	return strings.ToLower(parsed.Host)
}

// dedupStrings preserves order while removing duplicates.
func dedupStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func animeDriveQualityRank(label string) int {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "4k"), strings.Contains(lower, "2160"), strings.Contains(lower, "fhd"), strings.Contains(lower, "fullhd"), strings.Contains(lower, "full-hd"), strings.Contains(lower, "1080"):
		return 4
	case strings.Contains(lower, "hd"), strings.Contains(lower, "720"):
		return 3
	case strings.Contains(lower, "sd"), strings.Contains(lower, "480"), strings.Contains(lower, "360"):
		return 2
	default:
		return 1
	}
}
