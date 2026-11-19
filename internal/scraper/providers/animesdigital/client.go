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
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
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
	// playerFileRe finds the playlist the embedded player is set up with:
	// jwplayer's `file: "https:\/\/host\/path\/index.m3u8"`.
	playerFileRe = regexp.MustCompile(`\bfile\s*:\s*"([^"]+)"`)
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
	// blogHosts names hosts trusted as Blogger's beyond *.blogspot.com, so a
	// test can serve player 2's post from a local server.
	blogHosts map[string]bool
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
	c := &Client{
		http:       hc,
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  netx.UserAgent,
		maxRetries: 0,
	}
	if u, err := url.Parse(c.baseURL); err == nil && u.Hostname() != "" {
		c.blogHosts = map[string]bool{strings.ToLower(u.Hostname()): true}
	}
	return c
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

// GetEpisodeStreamURL resolves an episode to what plays: player 1's HLS
// playlist when its host answers for it, else player 2's Blogger video.
func (c *Client) GetEpisodeStreamURL(episodeURL string) (string, error) {
	return c.GetEpisodeStreamURLContext(context.Background(), episodeURL)
}

// GetEpisodeStreamURLContext is the cancellable stream implementation.
func (c *Client) GetEpisodeStreamURLContext(ctx context.Context, episodeURL string) (string, error) {
	page, err := c.getDocument(ctx, episodeURL, "stream")
	if err != nil {
		return "", err
	}
	var playerOne, alternate, cdnPlayer string
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
		switch {
		case strings.EqualFold(u.Hostname(), "api.anivideo.net") && strings.Contains(u.Path, "videohls.php"):
			if direct := strings.TrimSpace(u.Query().Get("d")); direct != "" {
				playerOne = direct
			}
		case isCDNPlayer(u):
			if cdnPlayer == "" {
				cdnPlayer = src
			}
		case alternate == "":
			alternate = src
		}
	})
	if playerOne == "" && alternate == "" && cdnPlayer == "" {
		return "", netx.NewParserError(SourceName, "stream", "no supported player iframe found", nil)
	}

	// Since 2026-11-18 player 1 is a page on the video host, /cdn/v1.<token>,
	// that sets its player up with the playlist (issue #237). The old
	// videohls.php?d=<playlist> address is kept for episodes that still use it.
	var cdnErr error
	if playerOne == "" && cdnPlayer != "" {
		if playerOne, cdnErr = c.cdnPlayerFile(ctx, cdnPlayer, episodeURL); cdnErr != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			util.Debug("AnimesDigital player 1's page gave no playlist", "error", cdnErr)
		}
	}

	// Player 1 is a playlist on the site's video host, and that host has no
	// file for some episodes it still lists: it answers 520 for the playlist
	// (issue #236). Handing that URL to the player made it close at once, with
	// nothing said. So the playlist is asked for first, and when it does not
	// answer, player 2's copy of the episode is used instead.
	var playerOneErr error
	if playerOne != "" {
		stream, err := netx.ValidateStreamURL(playerOne, SourceName)
		if err != nil {
			return "", err
		}
		if playerOneErr = c.playlistAnswers(ctx, stream); playerOneErr == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		util.Debug("AnimesDigital player 1 does not answer", "error", playerOneErr)
	}
	if alternate != "" {
		stream, err := c.playerTwoStream(ctx, alternate)
		if err == nil {
			return stream, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		util.Debug("AnimesDigital player 2 is unusable", "error", err)
		if playerOne == "" {
			return "", netx.NewParserError(SourceName, "stream", "player 1 is unavailable and player 2 could not be read", err)
		}
	}
	if playerOne == "" && cdnErr != nil {
		return "", fmt.Errorf("%s: the video host does not answer for this episode (%v): %w", SourceName, cdnErr, netx.ErrMediaOffline)
	}
	// The site lists the episode and neither player has a file for it.
	return "", fmt.Errorf("%s: the video host does not answer for this episode (%v): %w", SourceName, playerOneErr, netx.ErrMediaOffline)
}

// isCDNPlayer reports player 1's current address: a page on the video host
// under /cdn/.
func isCDNPlayer(u *url.URL) bool {
	return strings.EqualFold(u.Hostname(), "api.anivideo.net") && strings.HasPrefix(u.Path, "/cdn/")
}

// cdnPlayerFile opens player 1's page the way the site's own page does (as a
// frame of the episode) and answers the playlist it sets its player up with.
// The host refuses a request that does not name the episode as its Referer
// (403), and an episode it has no file for.
func (c *Client) cdnPlayerFile(ctx context.Context, playerURL, episodeURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playerURL, http.NoBody)
	if err != nil {
		return "", err
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", episodeURL)
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("player 1's page: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return "", err
	}
	for _, m := range playerFileRe.FindAllSubmatch(body, -1) {
		file := strings.TrimSpace(strings.ReplaceAll(string(m[1]), `\/`, "/"))
		if strings.HasPrefix(file, "https://") && isPlaylistPath(file) {
			return file, nil
		}
	}
	return "", errors.New("player 1's page names no playlist")
}

// streamProbeBudget bounds the question "does player 1 serve something the
// player can play": its playlist and the first piece of video it names.
const streamProbeBudget = 12 * time.Second

// playlistProbeLimit is how much of a playlist is read to find its first piece.
const playlistProbeLimit = 256 << 10

// PlayerReferer is the page that embeds player 1. The video host guards its
// segments by Referer and, for some titles, answers 403 to any other: the
// generic HLS referer the player falls back on let the playlist load and then
// failed on every segment, so mpv sat on a black window until it gave up
// (issue #229: Wind Breaker, and Takagi-san 2 episodes 9 and 10).
const PlayerReferer = "https://api.anivideo.net/"

// StreamReferer is the Referer the player must send for a stream this client
// returned: PlayerReferer for player 1's playlist, none for player 2's Blogger
// video, which is resolved by its own address.
func StreamReferer(stream string) string {
	if stream == "" || strings.HasPrefix(stream, bloggerVideoPrefix) {
		return ""
	}
	return PlayerReferer
}

// playlistAnswers reports whether target is something the player can start: nil
// when its playlist answers and the first piece of video the playlist names
// does too, and why not otherwise (the status it answered with, or what it sent
// instead of a playlist).
//
// The playlist alone is not enough. The video host guards the pieces apart from
// the playlist, and for some titles answered the playlist and then 403 for
// every piece: the player opened a black window, waited 45 seconds and closed
// (issue #229). Asking for the first piece with the Referer the player will
// send finds that now, while player 2 can still be offered instead.
func (c *Client) playlistAnswers(ctx context.Context, target string) error {
	ctx, cancel := context.WithTimeout(ctx, streamProbeBudget)
	defer cancel()
	piece, err := c.firstPiece(ctx, target, 0)
	if err != nil {
		return err
	}
	return c.pieceAnswers(ctx, piece)
}

// firstPiece reads the playlist at target and returns the address of the first
// piece of video it names, following one level of variant playlists.
func (c *Client) firstPiece(ctx context.Context, target string, depth int) (string, error) {
	resp, err := c.probe(ctx, target, "application/vnd.apple.mpegurl, application/x-mpegURL, */*", "")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, playlistProbeLimit))
	text := strings.TrimLeft(string(body), "\ufeff \t\r\n")
	if !strings.HasPrefix(text, "#EXTM3U") {
		return "", errors.New("not a playlist")
	}
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		next := c.resolveURL(target, line)
		if depth == 0 && isPlaylistPath(next) {
			return c.firstPiece(ctx, next, depth+1)
		}
		return next, nil
	}
	return "", errors.New("the playlist lists no video")
}

func isPlaylistPath(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && strings.HasSuffix(strings.ToLower(u.Path), ".m3u8")
}

// pieceAnswers asks for the first byte of one piece of video.
func (c *Client) pieceAnswers(ctx context.Context, piece string) error {
	resp, err := c.probe(ctx, piece, "*/*", "bytes=0-0")
	if err != nil {
		return fmt.Errorf("first piece: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("first piece: HTTP %d", resp.StatusCode)
	}
	return nil
}

// probe asks for target the way the player will: with the player's Referer.
func (c *Client) probe(ctx context.Context, target, accept, byteRange string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return nil, err
	}
	c.decorateRequest(req)
	req.Header.Set("Referer", PlayerReferer)
	req.Header.Set("Accept", accept)
	if byteRange != "" {
		req.Header.Set("Range", byteRange)
	}
	return c.http.Do(req)
}

// bloggerVideoPrefix is how an uploaded Blogger video is embedded; the player
// resolves such an address to its file.
const bloggerVideoPrefix = "https://www.blogger.com/video.g?token="

// playerTwoStream reads player 2: an address on the site that stands for the
// n-th video (from zero) uploaded to a Blogger post. The post's URL is its
// first segment in base64, ROT13-scrambled on older episodes, and the path is
// either /<post>/<n>/bg.mp4 or, on older episodes, /<post>/0/<n>/bg.mp4 (the
// first number is always 0). It answers that video's Blogger address.
func (c *Client) playerTwoStream(ctx context.Context, playerURL string) (string, error) {
	u, err := url.Parse(playerURL)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if (len(parts) != 3 && len(parts) != 4) || parts[len(parts)-1] != "bg.mp4" {
		return "", errors.New("player 2 has an address this client does not know")
	}
	index, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil || index < 0 {
		return "", errors.New("player 2 names no video")
	}
	raw, ok := decodePlayerTwoPost(parts[0])
	if !ok {
		return "", errors.New("player 2 names no post")
	}
	post, err := url.Parse(raw)
	if err != nil || post.Scheme != "https" || post.User != nil || !c.isBlogHost(post.Hostname()) {
		return "", errors.New("player 2 points outside Blogger")
	}
	page, err := c.getDocument(ctx, post.String(), "stream")
	if err != nil {
		return "", err
	}
	var videos []string
	page.doc.Find("iframe").Each(func(_ int, s *goquery.Selection) {
		if src := strings.TrimSpace(s.AttrOr("src", "")); strings.HasPrefix(src, bloggerVideoPrefix) && len(src) > len(bloggerVideoPrefix) {
			videos = append(videos, src)
		}
	})
	if index >= len(videos) {
		return "", fmt.Errorf("player 2's post holds %d videos, not the %d this episode needs", len(videos), index+1)
	}
	return videos[index], nil
}

// decodePlayerTwoPost turns player 2's first path segment into the Blogger
// post's address: base64 of the URL, or of its ROT13 scramble.
func decodePlayerTwoPost(segment string) (string, bool) {
	for _, candidate := range []string{segment, rot13(segment)} {
		raw, err := base64.StdEncoding.DecodeString(candidate)
		if err != nil {
			raw, err = base64.RawURLEncoding.DecodeString(strings.TrimRight(candidate, "="))
		}
		if text := strings.TrimSpace(string(raw)); err == nil && strings.HasPrefix(text, "https://") {
			return text, true
		}
	}
	return "", false
}

func rot13(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		}
		return r
	}, s)
}

// isBlogHost reports a host player 2's post may live on: a Blogger blog, or a
// host a test named.
func (c *Client) isBlogHost(host string) bool {
	host = strings.ToLower(host)
	return strings.HasSuffix(host, ".blogspot.com") || c.blogHosts[host]
}
