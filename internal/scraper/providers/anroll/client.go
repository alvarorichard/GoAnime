// Package anroll implements the plain-HTTP scraper for anroll.io.
package anroll

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

const (
	// SourceName is the display name shown for this source.
	SourceName = "AnimesROLL"
	// DefaultBase is the site root used by the shared client.
	DefaultBase      = "https://anroll.io"
	playerOrigin     = "https://anidrive.click"
	maxBodySize      = 12 << 20
	maxDecodedScript = 2 << 20
	// PlayerUserAgent is sent to the AniDrive player. Its Blogger MP4 links are
	// signed for the User-Agent that asked for the player (Google answers 403 to
	// any other), so it must be the exact one the downloader and mpv use for
	// googlevideo (player.downloadUserAgent; a test keeps them equal).
	PlayerUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var (
	dubbedSuffixRe = regexp.MustCompile(`(?i)\s*(?:[-–—]\s*)?dublado\s*$`)
	decodeCallRe   = regexp.MustCompile(`(?s)\w+\((\[".*"\]),(\[[\d,\s]*\]),"([A-Za-z0-9+/=]+)"\)\s*;?\s*\}\)\(\)\s*;?\s*$`)
	qualityRe      = regexp.MustCompile(`(?i)(\d{3,4})p?`)
)

// Variants are the title pages available for each audio.
type Variants struct {
	Subtitled string
	Dubbed    string
}

// Stream describes the selected player source and the headers it requires.
type Stream struct {
	URL     string
	Referer string
	Origin  string
	Type    string
}

type playerSource struct {
	File  string `json:"file"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// Client performs AnimesROLL and AniDrive requests.
type Client struct {
	http       *http.Client
	playerHTTP *http.Client
	baseURL    string
	playerBase string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
	// probeBlogger: check Blogger MP4 links before handing them out (off in tests: it goes to Google)
	probeBlogger bool
}

// NewClient builds the production client without doing network I/O.
func NewClient() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOANIME_ANROLL_URL")), "/")
	if base == "" {
		base = DefaultBase
	}
	return &Client{
		http:         util.NewFastClient(),
		playerHTTP:   &http.Client{Transport: netx.SafeScraperHTTP1Transport(8 * time.Second), Timeout: 8 * time.Second},
		baseURL:      base,
		playerBase:   playerOrigin,
		userAgent:    netx.UserAgent,
		maxRetries:   2,
		retryDelay:   350 * time.Millisecond,
		probeBlogger: true,
	}
}

// NewAnrollClient keeps the constructor shape used by the other leaf providers.
func NewAnrollClient() *Client { return NewClient() }

// NewClientForTest points both site and player requests at an HTTP test server.
func NewClientForTest(hc *http.Client, baseURL string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		http:       hc,
		playerHTTP: hc,
		baseURL:    baseURL,
		playerBase: baseURL,
		userAgent:  netx.UserAgent,
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

// ParseVariants decodes the private fragment used to carry a paired dubbed title.
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
	if isDubbed(rawURL, "", "") {
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

func (c *Client) get(ctx context.Context, rawURL, layer, referer string, player bool) ([]byte, string, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := c.sleep(ctx, attempt); err != nil {
				return nil, "", err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, "", err
		}
		c.decorateRequest(req, referer, player)
		httpClient := c.http
		if player {
			httpClient = c.playerHTTP
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.shouldRetry(attempt, err) {
				continue
			}
			return nil, "", err
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if c.shouldRetry(attempt, readErr) {
				continue
			}
			return nil, "", readErr
		}
		if player && resp.StatusCode == http.StatusForbidden {
			return nil, "", netx.NewBlockedChallengeError(SourceName, layer, "AniDrive rejected the required iframe headers; plain HTTP playback is unavailable", nil)
		}
		if err := netx.CheckHTTPStatus(resp, SourceName+" "+layer); err != nil {
			lastErr = err
			if c.shouldRetry(attempt, err) {
				continue
			}
			return nil, "", err
		}
		finalURL := rawURL
		if resp.Request != nil && resp.Request.URL != nil {
			finalURL = resp.Request.URL.String()
		}
		return body, finalURL, nil
	}
	return nil, "", lastErr
}

func (c *Client) getDocument(ctx context.Context, rawURL, layer string) (*goquery.Document, string, error) {
	body, finalURL, err := c.get(ctx, rawURL, layer, c.baseURL+"/", false)
	if err != nil {
		return nil, "", err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, "", netx.NewParserError(SourceName, layer, "unreadable HTML", err)
	}
	if err := netx.CheckChallengeDocument(doc, SourceName+" "+layer); err != nil {
		return nil, "", netx.NewBlockedChallengeError(SourceName, layer, "the site blocked plain HTTP; open it in a browser and try again later", err)
	}
	return doc, finalURL, nil
}

func (c *Client) decorateRequest(req *http.Request, referer string, player bool) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", netx.ChromeAcceptLanguage)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if player {
		req.Header.Set("User-Agent", PlayerUserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Sec-Fetch-Dest", "iframe")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		return
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
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

func resolveURL(base, ref string) string {
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
	doc, _, err := c.getDocument(ctx, c.baseURL+"/?s="+url.QueryEscape(q), "search")
	if err != nil {
		return nil, err
	}
	return c.extractSearchResults(doc), nil
}

type searchGroup struct {
	title     string
	subtitled *models.Anime
	dubbed    *models.Anime
}

func (c *Client) extractSearchResults(doc *goquery.Document) []*models.Anime {
	groups := make(map[string]*searchGroup)
	order := make([]string, 0)
	doc.Find("article.anime-card").Each(func(_ int, card *goquery.Selection) {
		link := card.Find(`a[href*="/anime/"]`).First()
		href, ok := link.Attr("href")
		if !ok || strings.TrimSpace(href) == "" {
			return
		}
		href = resolveURL(c.baseURL, href)
		img := card.Find("img").First()
		title, _ := img.Attr("alt")
		if title = strings.TrimSpace(title); title == "" {
			title = strings.TrimSpace(card.Find(".card-title").First().Text())
		}
		if title == "" {
			return
		}
		audio := strings.TrimSpace(card.Find(".card-sub").First().Text())
		mediaType := models.MediaTypeAnime
		kind := strings.ToLower(strings.TrimSpace(card.Find(".card-type").First().Text()))
		if kind == "filme" || kind == "movie" {
			mediaType = models.MediaTypeMovie
		}
		baseTitle := strings.TrimSpace(dubbedSuffixRe.ReplaceAllString(title, ""))
		// The media type is part of the key: a series and a movie sharing a
		// title are different works and must not overwrite each other.
		key := string(mediaType) + "|" + strings.ToLower(strings.Join(strings.Fields(baseTitle), " "))
		group := groups[key]
		if group == nil {
			group = &searchGroup{title: baseTitle}
			groups[key] = group
			order = append(order, key)
		}
		imageURL, _ := img.Attr("src")
		anime := &models.Anime{
			Name:      title,
			URL:       href,
			ImageURL:  resolveURL(c.baseURL, imageURL),
			Year:      strings.TrimSpace(card.Find(".card-year").First().Text()),
			MediaType: mediaType,
		}
		if isDubbed(href, title, audio) {
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

func isDubbed(rawURL, title, audio string) bool {
	return strings.EqualFold(strings.TrimSpace(audio), "Dublado") ||
		dubbedSuffixRe.MatchString(strings.TrimSpace(title)) ||
		strings.Contains(strings.ToLower(rawURL), "-dublado")
}

// GetAnimeEpisodes returns all episodes in ascending order.
func (c *Client) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return c.GetAnimeEpisodesContext(context.Background(), animeURL)
}

// GetAnimeEpisodesContext is the cancellable episode implementation.
func (c *Client) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	doc, _, err := c.getDocument(ctx, animeURL, "episodes")
	if err != nil {
		return nil, err
	}
	episodes := make([]models.Episode, 0)
	seen := make(map[string]bool)
	doc.Find("a.ep-text-item").Each(func(_ int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if !ok || strings.TrimSpace(href) == "" {
			return
		}
		href = resolveURL(c.baseURL, href)
		if seen[href] {
			return
		}
		raw, _ := item.Attr("data-ep")
		num, parseErr := strconv.Atoi(strings.TrimSpace(raw))
		if parseErr != nil || num < 0 {
			return
		}
		seen[href] = true
		episodes = append(episodes, models.Episode{Number: fmt.Sprintf("Episódio %d", num), Num: num, URL: href})
	})
	sort.Slice(episodes, func(i, j int) bool {
		if episodes[i].Num == episodes[j].Num {
			return episodes[i].URL < episodes[j].URL
		}
		return episodes[i].Num < episodes[j].Num
	})
	if len(episodes) == 0 {
		return nil, netx.NewParserError(SourceName, "episodes", "no episode links found", nil)
	}
	return episodes, nil
}

// GetEpisodeStream resolves the AniDrive player and selects its best source.
func (c *Client) GetEpisodeStream(episodeURL string) (Stream, error) {
	return c.GetEpisodeStreamContext(context.Background(), episodeURL)
}

// GetEpisodeStreamContext is the cancellable stream implementation.
func (c *Client) GetEpisodeStreamContext(ctx context.Context, episodeURL string) (Stream, error) {
	doc, _, err := c.getDocument(ctx, episodeURL, "episode")
	if err != nil {
		return Stream{}, err
	}
	playerURL := ""
	doc.Find("iframe").EachWithBreak(func(_ int, frame *goquery.Selection) bool {
		src, _ := frame.Attr("src")
		if isAniDrivePlayer(src) {
			playerURL = src
			return false
		}
		return true
	})
	if playerURL == "" {
		doc.Find(`script[type="application/ld+json"]`).EachWithBreak(func(_ int, script *goquery.Selection) bool {
			var value any
			if json.Unmarshal([]byte(script.Text()), &value) == nil {
				playerURL = findEmbedURL(value)
			}
			return playerURL == ""
		})
	}
	if playerURL == "" {
		return Stream{}, netx.NewParserError(SourceName, "episode", "no AniDrive player iframe found", nil)
	}
	playerURL = c.rewritePlayerURL(playerURL)
	body, _, err := c.get(ctx, playerURL, "player", episodeURL, true)
	if err != nil {
		return Stream{}, err
	}
	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return Stream{}, netx.NewParserError(SourceName, "player", "unreadable player HTML", err)
	}
	if err := netx.CheckChallengeDocument(doc, SourceName+" player"); err != nil {
		return Stream{}, netx.NewBlockedChallengeError(SourceName, "player", "AniDrive blocked the iframe request; plain HTTP playback is unavailable", err)
	}
	sources, err := extractPlayerSources(doc)
	if err != nil {
		return Stream{}, err
	}
	best, err := selectBestSource(sources)
	if err != nil {
		return Stream{}, err
	}
	streamURL, err := netx.ValidateStreamURL(best.File, SourceName)
	if err != nil {
		return Stream{}, err
	}
	stream := Stream{URL: streamURL, Type: best.Type}
	if looksLikeHLS(best) {
		stream.Referer = playerOrigin + "/"
		stream.Origin = playerOrigin
	} else if err := c.checkBloggerLink(ctx, streamURL); err != nil {
		return Stream{}, err
	}
	return stream, nil
}

// checkBloggerLink probes a Blogger MP4 once. Google sometimes refuses these
// links with 403 for a while (seen 2026-10-16: every episode of a title refused,
// then the same links worked minutes later). Without this probe the user only
// sees "download URL expired" once per episode, with no hint that waiting or
// another source is the way out.
func (c *Client) checkBloggerLink(ctx context.Context, streamURL string) error {
	if !c.probeBlogger || !c.bloggerRefuses(ctx, streamURL) {
		return nil
	}
	return &netx.SourceDiagnostic{
		Source:     SourceName,
		Layer:      "stream",
		Kind:       netx.DiagnosticDownloadExpired,
		StatusCode: http.StatusForbidden,
		Message: "Google is refusing this title's Blogger video right now (HTTP 403); it is usually temporary — " +
			"try again in a few minutes, or use another source",
	}
}

// bloggerRefuses reports whether Google answers a Blogger MP4 link with 403.
// Anything else (not a Blogger link, network trouble) is not proof of a
// refused link, so the player is left to try.
func (c *Client) bloggerRefuses(ctx context.Context, streamURL string) bool {
	u, err := url.Parse(streamURL)
	if err != nil || !strings.HasSuffix(strings.ToLower(u.Hostname()), ".googlevideo.com") || u.Query().Get("source") != "blogger" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, http.NoBody)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", PlayerUserAgent)
	req.Header.Set("Range", "bytes=0-0")
	resp, err := c.playerHTTP.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusForbidden
}

func (c *Client) rewritePlayerURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || c.playerBase == playerOrigin {
		return rawURL
	}
	b, err := url.Parse(c.playerBase)
	if err != nil {
		return rawURL
	}
	u.Scheme, u.Host = b.Scheme, b.Host
	return u.String()
}

// isAniDrivePlayer reports whether rawURL is an AniDrive player token page:
// the host must be anidrive.click itself, so a URL that only mentions it in
// its path or query is not fetched.
func isAniDrivePlayer(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), "anidrive.click") && strings.HasPrefix(u.Path, "/token/")
}

func findEmbedURL(value any) string {
	switch v := value.(type) {
	case map[string]any:
		if raw, ok := v["embedUrl"].(string); ok && isAniDrivePlayer(raw) {
			return raw
		}
		for _, child := range v {
			if found := findEmbedURL(child); found != "" {
				return found
			}
		}
	case []any:
		for _, child := range v {
			if found := findEmbedURL(child); found != "" {
				return found
			}
		}
	}
	return ""
}

func extractPlayerSources(doc *goquery.Document) ([]playerSource, error) {
	var sources []playerSource
	var decodeErr error
	doc.Find("script:not([src])").EachWithBreak(func(_ int, script *goquery.Selection) bool {
		decoded, err := decodeScriptLayers(script.Text())
		if err != nil {
			decodeErr = err
			return true
		}
		raw, ok := jsonValueAfterKey(decoded, "sources")
		if !ok {
			return true
		}
		if err := json.Unmarshal([]byte(raw), &sources); err != nil {
			decodeErr = err
			return true
		}
		return len(sources) == 0
	})
	if len(sources) == 0 {
		return nil, netx.NewParserError(SourceName, "player", "no sources found in the decoded player config", decodeErr)
	}
	return sources, nil
}

func decodeScriptLayers(script string) (string, error) {
	decoded := strings.TrimSpace(script)
	for range 8 {
		match := decodeCallRe.FindStringSubmatch(decoded)
		if len(match) == 0 {
			return decoded, nil
		}
		var parts []string
		var order []int
		if err := json.Unmarshal([]byte(match[1]), &parts); err != nil {
			return "", fmt.Errorf("decode player parts: %w", err)
		}
		if err := json.Unmarshal([]byte(match[2]), &order); err != nil {
			return "", fmt.Errorf("decode player order: %w", err)
		}
		key, err := base64.StdEncoding.DecodeString(match[3])
		if err != nil || len(key) == 0 {
			return "", errors.New("decode player: invalid XOR key")
		}
		var joined strings.Builder
		for _, index := range order {
			if index < 0 || index >= len(parts) {
				return "", errors.New("decode player: part index out of range")
			}
			joined.WriteString(parts[index])
			if joined.Len() > maxDecodedScript*2 {
				return "", errors.New("decode player: encoded script is too large")
			}
		}
		data, err := base64.StdEncoding.DecodeString(joined.String())
		if err != nil {
			return "", fmt.Errorf("decode player payload: %w", err)
		}
		if len(data) > maxDecodedScript {
			return "", errors.New("decode player: decoded script is too large")
		}
		for i := range data {
			data[i] ^= key[i%len(key)]
		}
		decoded = string(data)
	}
	return "", errors.New("decode player: too many obfuscation layers")
}

func jsonValueAfterKey(script, key string) (string, bool) {
	needle := `"` + key + `"`
	start := strings.Index(script, needle)
	if start < 0 {
		return "", false
	}
	start += len(needle)
	colon := strings.IndexByte(script[start:], ':')
	if colon < 0 {
		return "", false
	}
	start += colon + 1
	for start < len(script) && (script[start] == ' ' || script[start] == '\t' || script[start] == '\r' || script[start] == '\n') {
		start++
	}
	if start >= len(script) || (script[start] != '[' && script[start] != '{') {
		return "", false
	}
	open := script[start]
	closingDelimiter := byte(']')
	if open == '{' {
		closingDelimiter = '}'
	}
	depth, inString, escaped := 0, false, false
	for i := start; i < len(script); i++ {
		ch := script[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case ch == '\\':
				escaped = true
			case ch == '"':
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		switch ch {
		case open:
			depth++
		case closingDelimiter:
			depth--
			if depth == 0 {
				return script[start : i+1], true
			}
		}
	}
	return "", false
}

func looksLikeHLS(source playerSource) bool {
	return strings.Contains(strings.ToLower(source.File), ".m3u8") ||
		strings.EqualFold(source.Type, "hls") || strings.EqualFold(source.Type, "application/x-mpegurl")
}

func selectBestSource(sources []playerSource) (playerSource, error) {
	valid := make([]playerSource, 0, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.File) != "" {
			valid = append(valid, source)
		}
	}
	if len(valid) == 0 {
		return playerSource{}, netx.NewParserError(SourceName, "player", "player source list is empty", nil)
	}
	sort.SliceStable(valid, func(i, j int) bool {
		return sourceQuality(valid[i]) > sourceQuality(valid[j])
	})
	return valid[0], nil
}

func sourceQuality(source playerSource) int {
	if looksLikeHLS(source) {
		return 10000
	}
	match := qualityRe.FindStringSubmatch(source.Label)
	if len(match) < 2 {
		return 0
	}
	quality, _ := strconv.Atoi(match[1])
	return quality
}
