// Package scraper provides web scraping functionality for animesdrive.blog
package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

const (
	AnimeDriveBase = "https://animesdrive.blog"
)

// VideoQuality represents video quality options
type VideoQuality int

const (
	QualityMobile VideoQuality = iota
	QualitySD
	QualityHD
	QualityFullHD
	QualityFHD
	QualityUnknown
)

// String returns the display label for the quality
func (q VideoQuality) String() string {
	switch q {
	case QualityMobile:
		return "Mobile / Celular"
	case QualitySD:
		return "SD"
	case QualityHD:
		return "HD"
	case QualityFullHD:
		return "FullHD / HLS"
	case QualityFHD:
		return "FHD"
	default:
		return "Unknown"
	}
}

// Badge returns the short badge for the quality
func (q VideoQuality) Badge() string {
	switch q {
	case QualityMobile:
		return "SD"
	case QualitySD:
		return "SD"
	case QualityHD:
		return "HD"
	case QualityFullHD:
		return "FHD"
	case QualityFHD:
		return "FHD"
	default:
		return ""
	}
}

// ParseVideoQuality parses a label string to VideoQuality
func ParseVideoQuality(label string) VideoQuality {
	lower := strings.ToLower(strings.TrimSpace(label))

	// Mobile/Celular first (most specific)
	if strings.Contains(lower, "mobile") || strings.Contains(lower, "celular") {
		return QualityMobile
	}

	// FullHD or HLS (1080p streaming)
	if strings.Contains(lower, "fullhd") || lower == "hls" {
		return QualityFullHD
	}

	// FHD alone (1080p)
	if lower == "fhd" || (strings.Contains(lower, "fhd") && !strings.Contains(lower, "/")) {
		return QualityFHD
	}

	// SD / HD combined - treat as HD
	if strings.Contains(lower, "sd") && strings.Contains(lower, "hd") {
		return QualityHD
	}

	// SD alone
	if strings.Contains(lower, "sd") && !strings.Contains(lower, "hd") {
		return QualitySD
	}

	// HD alone (720p)
	if strings.Contains(lower, "hd") {
		return QualityHD
	}

	return QualityUnknown
}

// VideoOption represents a server/quality option for an episode
type VideoOption struct {
	Label       string
	Quality     VideoQuality
	ServerName  string
	ServerIndex int
	VideoURL    string
	PostID      string
	Type        string
	Nume        string
}

// ErrBackRequested is returned when user selects the back option
var ErrBackRequested = errors.New("back requested")

// SelectServerWithFuzzyFinder allows the user to select a server/quality option using fuzzy finder
func SelectServerWithFuzzyFinder(options []VideoOption) (*VideoOption, error) {
	if len(options) == 0 {
		return nil, errors.New("no server options available")
	}

	// If only one option, return it directly
	if len(options) == 1 {
		return &options[0], nil
	}

	// Create display list with back option first
	backOption := "← Voltar (seleção de episódio)"
	displayList := make([]string, len(options)+1)
	displayList[0] = backOption
	for i, opt := range options {
		if opt.Label != "" {
			displayList[i+1] = opt.Label
		} else {
			displayList[i+1] = fmt.Sprintf("%s (%s)", opt.Quality.String(), opt.ServerName)
		}
	}

	idx, err := fuzzyfinder.Find(
		displayList,
		func(i int) string {
			return displayList[i]
		},
		fuzzyfinder.WithPromptString("Selecione o servidor: "),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to select server: %w", err)
	}

	if idx < 0 || idx >= len(displayList) {
		return nil, errors.New("invalid server selection")
	}

	// Check if back was selected
	if idx == 0 {
		return nil, ErrBackRequested
	}

	// Adjust index for options (subtract 1 for the back option)
	optionIdx := idx - 1
	return &options[optionIdx], nil
}

// AnimeDriveGenre represents a genre from AnimeDrive
type AnimeDriveGenre struct {
	ID   string
	Name string
	URL  string
}

// AnimeDriveShow represents an anime from AnimeDrive
type AnimeDriveShow struct {
	ID        string
	Title     string
	URL       string
	Thumbnail string
	Rating    string
	Year      string
	IsDubbed  bool
}

// AnimeDriveEpisode represents an episode from AnimeDrive
type AnimeDriveEpisode struct {
	Number    string
	Title     string
	URL       string
	Thumbnail string
	Qualities []string
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

// AnimeDriveClient handles interactions with animesdrive.blog
type AnimeDriveClient struct {
	client     *http.Client
	baseURL    string
	userAgent  string
	maxRetries int
	retryDelay time.Duration
	totalPages int
}

// NewAnimeDriveClient creates a new AnimeDrive client
func NewAnimeDriveClient() *AnimeDriveClient {
	return &AnimeDriveClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    AnimeDriveBase,
		userAgent:  UserAgent,
		maxRetries: 2,
		retryDelay: 350 * time.Millisecond,
		totalPages: 371,
	}
}

// AlphabetLetters returns the list of letters for A-Z navigation
func (c *AnimeDriveClient) AlphabetLetters() []string {
	return []string{
		"#", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}
}

// PreferredDomains are domains that work well with players
var preferredDomains = []string{
	"tityos.feralhosting.com",
	"feralhosting.com",
	"archive.org",
}

// ProblematicDomains are domains that may have CORS/blocking issues
var problematicDomains = []string{
	"aniplay.online",
	"animeshd.cloud",
	"animes.strp2p.com",
}

func isPreferredDomain(urlStr string) bool {
	lower := strings.ToLower(urlStr)
	for _, d := range preferredDomains {
		if strings.Contains(lower, d) {
			return true
		}
	}
	return false
}

func isProblematicDomain(urlStr string) bool {
	lower := strings.ToLower(urlStr)
	for _, d := range problematicDomains {
		if strings.Contains(lower, d) {
			return true
		}
	}
	return false
}

func (c *AnimeDriveClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", c.baseURL)
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
	// Normalize query: replace hyphens/underscores with spaces for WordPress search
	normalizedQuery := strings.TrimSpace(query)
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "-", " ")
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "_", " ")
	// Collapse multiple spaces
	for strings.Contains(normalizedQuery, "  ") {
		normalizedQuery = strings.ReplaceAll(normalizedQuery, "  ", " ")
	}

	searchURL := fmt.Sprintf("%s/?s=%s", c.baseURL, url.QueryEscape(normalizedQuery))

	util.Debug("AnimeDrive search", "query", query, "normalized", normalizedQuery, "url", searchURL)

	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.decorateRequest(req)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("server returned: %s", resp.Status)
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to parse HTML: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
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

func (c *AnimeDriveClient) extractSearchResults(doc *goquery.Document) []*models.Anime {
	var animes []*models.Anime

	// Search for result cards - multiple selectors to cover theme variations
	selectors := []string{
		"article.item",
		"div.result-item",
		"div.search-page .item",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			// Try to extract title and URL
			titleElement := s.Find("h3 a, h2 a, .title a, a.tip").First()
			imageElement := s.Find("img").First()

			if titleElement.Length() > 0 {
				title := strings.TrimSpace(titleElement.Text())
				urlPath, exists := titleElement.Attr("href")
				if !exists {
					return
				}

				// Only include anime URLs
				if !strings.Contains(urlPath, "/anime/") {
					return
				}

				if title == "" {
					return
				}

				imgURL, _ := imageElement.Attr("src")
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
			}
		})

		if len(animes) > 0 {
			break
		}
	}

	return animes
}

// GetAnimesByPage navigates the anime list by page
func (c *AnimeDriveClient) GetAnimesByPage(page int) ([]AnimeDriveShow, error) {
	util.Debug("AnimeDrive getting animes page", "page", page)

	var pageURL string
	if page == 1 {
		pageURL = fmt.Sprintf("%s/anime/", c.baseURL)
	} else {
		pageURL = fmt.Sprintf("%s/anime/page/%d/", c.baseURL, page)
	}

	req, err := http.NewRequest("GET", pageURL, nil)
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

	var results []AnimeDriveShow

	// Update total pages from pagination
	doc.Find(".pagination a, .wp-pagenavi a").Each(func(i int, s *goquery.Selection) {
		pageNum, err := strconv.Atoi(strings.TrimSpace(s.Text()))
		if err == nil && pageNum > c.totalPages {
			c.totalPages = pageNum
		}
	})

	// Extract animes - multiple selectors to cover theme variations
	selectors := "article.item, .items article, #archive-content article, .content article, .movies-list .ml-item, .animation-2 .item"
	doc.Find(selectors).Each(func(i int, item *goquery.Selection) {
		linkElement := item.Find("a[href*='/anime/']").First()
		titleElement := item.Find("h3, h2, .data h3, .title, .mli-info h2").First()
		imageElement := item.Find("img").First()
		ratingElement := item.Find(".rating, .score, .imdb").First()
		yearElement := item.Find(".year, .date, span.year").First()

		urlPath, exists := linkElement.Attr("href")
		if !exists || urlPath == "" || !strings.Contains(urlPath, "/anime/") {
			return
		}

		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			title, _ = linkElement.Attr("title")
			title = strings.TrimSpace(title)
		}
		if title == "" {
			return
		}

		image, _ := imageElement.Attr("src")
		if image == "" {
			image, _ = imageElement.Attr("data-src")
		}
		if image == "" {
			image, _ = imageElement.Attr("data-lazy-src")
		}

		rating := strings.TrimSpace(ratingElement.Text())
		year := strings.TrimSpace(yearElement.Text())
		isDubbed := strings.Contains(strings.ToLower(title), "dublado")

		results = append(results, AnimeDriveShow{
			ID:        c.extractIDFromURL(urlPath),
			Title:     title,
			URL:       urlPath,
			Thumbnail: image,
			Rating:    rating,
			Year:      year,
			IsDubbed:  isDubbed,
		})
	})

	util.Debug("AnimeDrive found animes on page", "page", page, "count", len(results))
	return results, nil
}

// GetAnimesByLetter navigates the anime list by letter (A-Z)
func (c *AnimeDriveClient) GetAnimesByLetter(letter string, page int) ([]AnimeDriveShow, error) {
	util.Debug("AnimeDrive getting animes by letter", "letter", letter, "page", page)

	var letterURL string
	letterParam := strings.ToLower(letter)
	if letter == "#" {
		letterParam = "0-9"
	}

	if page == 1 {
		letterURL = fmt.Sprintf("%s/anime/?letter=%s", c.baseURL, letterParam)
	} else {
		letterURL = fmt.Sprintf("%s/anime/page/%d/?letter=%s", c.baseURL, page, letterParam)
	}

	req, err := http.NewRequest("GET", letterURL, nil)
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

	var results []AnimeDriveShow

	doc.Find("article.item, .items article, #archive-content article").Each(func(i int, item *goquery.Selection) {
		linkElement := item.Find("a[href*='/anime/']").First()
		titleElement := item.Find("h3, h2, .data h3, .title").First()
		imageElement := item.Find("img").First()

		urlPath, exists := linkElement.Attr("href")
		if !exists || urlPath == "" || !strings.Contains(urlPath, "/anime/") {
			return
		}

		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			return
		}

		image, _ := imageElement.Attr("src")
		if image == "" {
			image, _ = imageElement.Attr("data-src")
		}

		results = append(results, AnimeDriveShow{
			ID:        c.extractIDFromURL(urlPath),
			Title:     title,
			URL:       urlPath,
			Thumbnail: image,
			IsDubbed:  strings.Contains(strings.ToLower(title), "dublado"),
		})
	})

	util.Debug("AnimeDrive found animes for letter", "letter", letter, "count", len(results))
	return results, nil
}

// GetGenres fetches available genres
func (c *AnimeDriveClient) GetGenres() ([]AnimeDriveGenre, error) {
	util.Debug("AnimeDrive getting genres")

	req, err := http.NewRequest("GET", c.baseURL, nil)
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

	var genres []AnimeDriveGenre
	seenGenres := make(map[string]bool)

	doc.Find("a[href*='/genre/']").Each(func(i int, s *goquery.Selection) {
		urlPath, exists := s.Attr("href")
		if !exists || urlPath == "" {
			return
		}

		name := strings.TrimSpace(s.Text())
		if name == "" {
			return
		}

		if seenGenres[urlPath] {
			return
		}
		seenGenres[urlPath] = true

		genres = append(genres, AnimeDriveGenre{
			ID:   c.extractIDFromURL(urlPath),
			Name: name,
			URL:  urlPath,
		})
	})

	util.Debug("AnimeDrive found genres", "count", len(genres))
	return genres, nil
}

// GetAnimesByGenre fetches animes by genre
func (c *AnimeDriveClient) GetAnimesByGenre(genreURL string, page int) ([]AnimeDriveShow, error) {
	urlStr := genreURL
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = c.baseURL + genreURL
	}

	if page > 1 {
		urlStr = fmt.Sprintf("%s/page/%d/", strings.TrimSuffix(urlStr, "/"), page)
	}

	util.Debug("AnimeDrive getting animes by genre", "url", urlStr)

	req, err := http.NewRequest("GET", urlStr, nil)
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

	var results []AnimeDriveShow

	doc.Find("article.item, .items article").Each(func(i int, item *goquery.Selection) {
		linkElement := item.Find("a[href*='/anime/']").First()
		titleElement := item.Find("h3, h2, .title").First()
		imageElement := item.Find("img").First()

		urlPath, exists := linkElement.Attr("href")
		if !exists || urlPath == "" {
			return
		}

		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			return
		}

		image, _ := imageElement.Attr("src")

		results = append(results, AnimeDriveShow{
			ID:        c.extractIDFromURL(urlPath),
			Title:     title,
			URL:       urlPath,
			Thumbnail: image,
			IsDubbed:  strings.Contains(strings.ToLower(title), "dublado"),
		})
	})

	return results, nil
}

// GetAnimeDetails fetches anime details including episode list
func (c *AnimeDriveClient) GetAnimeDetails(animeURL string) (*AnimeDriveDetails, error) {
	util.Debug("AnimeDrive getting details", "url", animeURL)

	urlStr := animeURL
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = c.baseURL + animeURL
	}

	req, err := http.NewRequest("GET", urlStr, nil)
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

	// Extract title
	titleElement := doc.Find("h1.entry-title, h1, .sheader .data h1").First()
	title := strings.TrimSpace(titleElement.Text())
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
	episodeSelectors := "#seasons .episodios li a, .episodios li a, ul.episodios a, .se-a a, #episodes a, .episodelist a"
	doc.Find(episodeSelectors).Each(func(i int, s *goquery.Selection) {
		epURL, exists := s.Attr("href")
		if !exists || !strings.Contains(epURL, "episodio") {
			return
		}

		epTitle := strings.TrimSpace(s.Text())

		// Extract episode number
		epNumber := "0"
		epMatch := regexp.MustCompile(`(?i)episodio[s]?[-_]?(\d+)`).FindStringSubmatch(epURL)
		if len(epMatch) > 1 {
			epNumber = epMatch[1]
		} else {
			numMatch := regexp.MustCompile(`(\d+)`).FindStringSubmatch(epTitle)
			if len(numMatch) > 1 {
				epNumber = numMatch[1]
			}
		}

		episodes = append(episodes, AnimeDriveEpisode{
			Number: epNumber,
			Title:  epTitle,
			URL:    epURL,
		})
	})

	// Sort episodes by number
	sort.Slice(episodes, func(i, j int) bool {
		numA, _ := strconv.Atoi(episodes[i].Number)
		numB, _ := strconv.Atoi(episodes[j].Number)
		return numA < numB
	})

	util.Debug("AnimeDrive found episodes", "count", len(episodes))

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
	details, err := c.GetAnimeDetails(animeURL)
	if err != nil {
		return nil, err
	}

	var episodes []models.Episode
	for _, ep := range details.Episodes {
		num, _ := strconv.Atoi(ep.Number)
		episodes = append(episodes, models.Episode{
			Number: ep.Number,
			Num:    num,
			URL:    ep.URL,
			Title: models.TitleDetails{
				Romaji: ep.Title,
			},
		})
	}

	return episodes, nil
}

// GetVideoOptions extracts all quality/server options from an episode
func (c *AnimeDriveClient) GetVideoOptions(episodeURL string) ([]VideoOption, error) {
	util.Debug("AnimeDrive getting video options", "url", episodeURL)

	urlStr := episodeURL
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = c.baseURL + episodeURL
	}

	req, err := http.NewRequest("GET", urlStr, nil)
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

	var options []VideoOption

	// Search for all player/server options
	playerSelectors := ".dooplay_player_option, [class*='player_option'], .source-box li, .player_nav li, .server-item"
	serverIndex := 0

	doc.Find(playerSelectors).Each(func(i int, s *goquery.Selection) {
		dataPost, hasPost := s.Attr("data-post")
		dataType := "tv"
		if dt, exists := s.Attr("data-type"); exists {
			dataType = dt
		}
		dataNume := fmt.Sprintf("%d", serverIndex+1)
		if dn, exists := s.Attr("data-nume"); exists {
			dataNume = dn
		}

		// Extract server label
		serverLabel := strings.TrimSpace(s.Find(".title, .server, span").First().Text())
		if serverLabel == "" {
			serverLabel = strings.TrimSpace(s.Text())
		}

		quality := ParseVideoQuality(serverLabel)

		if hasPost {
			label := serverLabel
			if label == "" {
				label = fmt.Sprintf("Server %d", serverIndex+1)
			}

			options = append(options, VideoOption{
				Label:       label,
				Quality:     quality,
				ServerName:  fmt.Sprintf("Server %d", serverIndex+1),
				ServerIndex: serverIndex,
				PostID:      dataPost,
				Type:        dataType,
				Nume:        dataNume,
			})

			util.Debug("AnimeDrive found option", "label", serverLabel, "post", dataPost, "nume", dataNume)
		}

		serverIndex++
	})

	// If no structured options found, try other elements
	if len(options) == 0 {
		doc.Find("[data-post][data-nume]").Each(func(i int, s *goquery.Selection) {
			dataPost, _ := s.Attr("data-post")
			dataType := "tv"
			if dt, exists := s.Attr("data-type"); exists {
				dataType = dt
			}
			dataNume := fmt.Sprintf("%d", i+1)
			if dn, exists := s.Attr("data-nume"); exists {
				dataNume = dn
			}
			label := strings.TrimSpace(s.Text())

			if dataPost != "" {
				if label == "" {
					label = fmt.Sprintf("Server %d", i+1)
				}

				options = append(options, VideoOption{
					Label:       label,
					Quality:     ParseVideoQuality(label),
					ServerName:  fmt.Sprintf("Server %d", i+1),
					ServerIndex: i,
					PostID:      dataPost,
					Type:        dataType,
					Nume:        dataNume,
				})
			}
		})
	}

	util.Debug("AnimeDrive total video options", "count", len(options))
	return options, nil
}

// ResolveVideoURLWithType resolves the video URL for a specific option
// Returns the URL and type (mp4, hls, or iframe)
func (c *AnimeDriveClient) ResolveVideoURLWithType(option VideoOption) (string, string, error) {
	if option.VideoURL != "" {
		return option.VideoURL, "mp4", nil
	}

	if option.PostID == "" || option.Nume == "" {
		return "", "", errors.New("missing post ID or nume")
	}

	util.Debug("AnimeDrive resolving video", "label", option.Label, "nume", option.Nume)

	apiURL := fmt.Sprintf("%s/wp-json/dooplayer/v2/%s/%s/%s",
		c.baseURL, option.PostID, option.Type, option.Nume)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("server returned: %s", resp.Status)
	}

	var apiData struct {
		EmbedURL string `json:"embed_url"`
		Type     string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if apiData.EmbedURL == "" {
		return "", "", errors.New("no embed URL in response")
	}

	util.Debug("AnimeDrive got embed URL", "type", apiData.Type, "url", apiData.EmbedURL)

	// If it's MP4 type, extract the source
	if apiData.Type == "mp4" {
		sourceMatch := regexp.MustCompile(`source=([^&]+)`).FindStringSubmatch(apiData.EmbedURL)
		if len(sourceMatch) > 1 {
			decodedSource, err := url.QueryUnescape(sourceMatch[1])
			if err == nil {
				util.Debug("AnimeDrive extracted MP4", "url", decodedSource)
				return decodedSource, "mp4", nil
			}
		}
	}

	// If it's an HLS link
	if strings.Contains(apiData.EmbedURL, ".m3u8") {
		return apiData.EmbedURL, "hls", nil
	}

	// Return iframe for other types
	videoType := apiData.Type
	if videoType == "" {
		videoType = "iframe"
	}
	return apiData.EmbedURL, videoType, nil
}

// ResolveVideoURL resolves the video URL for a specific option (compatibility)
func (c *AnimeDriveClient) ResolveVideoURL(option VideoOption) (string, error) {
	urlStr, _, err := c.ResolveVideoURLWithType(option)
	return urlStr, err
}

// GetVideoURLWithSelection allows user to select a server/quality before getting the video URL
func (c *AnimeDriveClient) GetVideoURLWithSelection(episodeURL string) (string, error) {
	util.Debug("AnimeDrive getting video URL with selection", "url", episodeURL)

	// First get all options
	options, err := c.GetVideoOptions(episodeURL)
	if err != nil || len(options) == 0 {
		util.Debug("AnimeDrive no options found, using fallback")
		return c.getVideoURLFallback(episodeURL)
	}

	// Resolve video URLs for all options to show what's actually available
	var resolvedOptions []VideoOption
	for _, opt := range options {
		videoURL, videoType, err := c.ResolveVideoURLWithType(opt)
		if err == nil && videoURL != "" {
			opt.VideoURL = videoURL
			// Update label to include type info
			if opt.Label == "" {
				opt.Label = opt.Quality.String()
			}
			if videoType == "hls" {
				opt.Label = opt.Label + " (HLS)"
			}
			resolvedOptions = append(resolvedOptions, opt)
		}
	}

	if len(resolvedOptions) == 0 {
		util.Debug("AnimeDrive no resolved options, using fallback")
		return c.getVideoURLFallback(episodeURL)
	}

	// Let user select a server
	selected, err := SelectServerWithFuzzyFinder(resolvedOptions)
	if err != nil {
		// If user selected back, return the error to propagate it
		if errors.Is(err, ErrBackRequested) {
			return "", ErrBackRequested
		}
		// If selection fails (e.g., user cancelled), try auto-selection
		util.Debug("AnimeDrive server selection failed, using auto-select", "error", err)
		return c.GetVideoURL(episodeURL)
	}

	if selected.VideoURL != "" {
		util.Debug("AnimeDrive user selected server", "label", selected.Label, "url", selected.VideoURL)
		return selected.VideoURL, nil
	}

	return c.getVideoURLFallback(episodeURL)
}

// GetVideoURL extracts the direct MP4 link from an episode (best quality available)
// Prioritizes reliable servers (tityos) over problematic ones (aniplay)
func (c *AnimeDriveClient) GetVideoURL(episodeURL string) (string, error) {
	util.Debug("AnimeDrive getting video URL", "url", episodeURL)

	// First get all options
	options, err := c.GetVideoOptions(episodeURL)
	if err != nil || len(options) == 0 {
		return c.getVideoURLFallback(episodeURL)
	}

	// Collect all available MP4 links
	type mp4Link struct {
		url         string
		option      VideoOption
		preferred   bool
		problematic bool
	}

	var allMP4Links []mp4Link

	for _, option := range options {
		videoURL, videoType, err := c.ResolveVideoURLWithType(option)
		if err == nil && videoType == "mp4" && videoURL != "" {
			allMP4Links = append(allMP4Links, mp4Link{
				url:         videoURL,
				option:      option,
				preferred:   isPreferredDomain(videoURL),
				problematic: isProblematicDomain(videoURL),
			})
			util.Debug("AnimeDrive found MP4", "label", option.Label, "url", videoURL, "preferred", isPreferredDomain(videoURL))
		}
	}

	// Sort: preferred domains first, problematic ones last
	sort.Slice(allMP4Links, func(i, j int) bool {
		// Preferred comes first
		if allMP4Links[i].preferred && !allMP4Links[j].preferred {
			return true
		}
		if !allMP4Links[i].preferred && allMP4Links[j].preferred {
			return false
		}
		// Problematic goes last
		if allMP4Links[i].problematic && !allMP4Links[j].problematic {
			return false
		}
		if !allMP4Links[i].problematic && allMP4Links[j].problematic {
			return true
		}
		return false
	})

	// Return the first link from sorted list
	if len(allMP4Links) > 0 {
		best := allMP4Links[0]
		util.Debug("AnimeDrive selected best", "url", best.url, "preferred", best.preferred)
		return best.url, nil
	}

	// Fallback: try iframes
	util.Debug("AnimeDrive no MP4 found, trying iframes...")
	for _, option := range options {
		videoURL, videoType, err := c.ResolveVideoURLWithType(option)
		if err == nil && videoType == "iframe" {
			util.Debug("AnimeDrive trying iframe", "url", videoURL)
			extractedURL, err := c.extractFromIframe(videoURL)
			if err == nil && extractedURL != "" {
				util.Debug("AnimeDrive extracted from iframe", "url", extractedURL)
				return extractedURL, nil
			}
		}
	}

	return c.getVideoURLFallback(episodeURL)
}

// getVideoURLFallback is a fallback method to extract video URL
func (c *AnimeDriveClient) getVideoURLFallback(episodeURL string) (string, error) {
	util.Debug("AnimeDrive using fallback method", "url", episodeURL)

	urlStr := episodeURL
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = c.baseURL + episodeURL
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Try to find player data and use dooplayer API
	playerOption := doc.Find(".dooplay_player_option, [class*='player_option']").First()
	if playerOption.Length() > 0 {
		dataPost, hasPost := playerOption.Attr("data-post")
		dataType := "tv"
		if dt, exists := playerOption.Attr("data-type"); exists {
			dataType = dt
		}
		dataNume := "1"
		if dn, exists := playerOption.Attr("data-nume"); exists {
			dataNume = dn
		}

		if hasPost {
			util.Debug("AnimeDrive found player data", "post", dataPost, "type", dataType, "nume", dataNume)

			apiURL := fmt.Sprintf("%s/wp-json/dooplayer/v2/%s/%s/%s",
				c.baseURL, dataPost, dataType, dataNume)

			apiReq, err := http.NewRequest("GET", apiURL, nil)
			if err == nil {
				c.decorateRequest(apiReq)
				apiReq.Header.Set("Referer", urlStr)

				apiResp, err := c.client.Do(apiReq)
				if err == nil {
					defer func() { _ = apiResp.Body.Close() }()

					if apiResp.StatusCode == http.StatusOK {
						var apiData struct {
							EmbedURL string `json:"embed_url"`
						}

						if err := json.NewDecoder(apiResp.Body).Decode(&apiData); err == nil && apiData.EmbedURL != "" {
							util.Debug("AnimeDrive got embed URL", "url", apiData.EmbedURL)

							// Extract source from embed URL
							sourceMatch := regexp.MustCompile(`source=([^&]+)`).FindStringSubmatch(apiData.EmbedURL)
							if len(sourceMatch) > 1 {
								decodedSource, err := url.QueryUnescape(sourceMatch[1])
								if err == nil {
									util.Debug("AnimeDrive extracted MP4", "url", decodedSource)
									return decodedSource, nil
								}
							}

							return apiData.EmbedURL, nil
						}
					}
				}
			}
		}
	}

	// Try all player options
	doc.Find("[data-post][data-nume], .dooplay_player_option").Each(func(i int, s *goquery.Selection) {
		dataPost, hasPost := s.Attr("data-post")
		dataType := "tv"
		if dt, exists := s.Attr("data-type"); exists {
			dataType = dt
		}
		dataNume := "1"
		if dn, exists := s.Attr("data-nume"); exists {
			dataNume = dn
		}

		if hasPost {
			apiURL := fmt.Sprintf("%s/wp-json/dooplayer/v2/%s/%s/%s",
				c.baseURL, dataPost, dataType, dataNume)

			apiReq, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				return
			}
			c.decorateRequest(apiReq)
			apiReq.Header.Set("Referer", urlStr)

			apiResp, err := c.client.Do(apiReq)
			if err != nil {
				return
			}
			defer func() { _ = apiResp.Body.Close() }()

			if apiResp.StatusCode == http.StatusOK {
				var apiData struct {
					EmbedURL string `json:"embed_url"`
				}

				if err := json.NewDecoder(apiResp.Body).Decode(&apiData); err == nil && apiData.EmbedURL != "" {
					sourceMatch := regexp.MustCompile(`source=([^&]+)`).FindStringSubmatch(apiData.EmbedURL)
					if len(sourceMatch) > 1 {
						decodedSource, _ := url.QueryUnescape(sourceMatch[1])
						if decodedSource != "" {
							util.Debug("AnimeDrive extracted MP4 from option", "url", decodedSource)
							// Can't return from here, but could store the result
						}
					}
				}
			}
		}
	})

	// Get HTML content for regex searches
	html, _ := doc.Html()

	// Fallback: search for tityos URL directly
	tityosMatch := regexp.MustCompile(`https?://tityos\.feralhosting\.com/[^\s<>"]+\.mp4`).FindString(html)
	if tityosMatch != "" {
		util.Debug("AnimeDrive found tityos URL", "url", tityosMatch)
		return tityosMatch, nil
	}

	// Method 2: search for any direct MP4 link
	mp4Match := regexp.MustCompile(`https?://[^\s<>"]+\.mp4`).FindString(html)
	if mp4Match != "" {
		util.Debug("AnimeDrive found MP4 URL", "url", mp4Match)
		return mp4Match, nil
	}

	// Method 3: search in iframes
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		iframeSrc, _ := s.Attr("src")
		if iframeSrc == "" {
			iframeSrc, _ = s.Attr("data-src")
		}

		if iframeSrc != "" {
			util.Debug("AnimeDrive found iframe", "src", iframeSrc)
			// Could try to extract from iframe here
		}
	})

	// Method 4: search in scripts for player configurations
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()

		// Search for file: "url" or source: "url" common in players
		fileMatch := regexp.MustCompile(`(?i)(?:file|source|src|url)\s*[=:]\s*["']([^"']+\.mp4)`).FindStringSubmatch(scriptContent)
		if len(fileMatch) > 1 {
			util.Debug("AnimeDrive found in script", "url", fileMatch[1])
			// Could return this
		}

		// Search for HLS/M3U8 links
		hlsMatch := regexp.MustCompile(`["']([^"']+\.m3u8[^"']*)["']`).FindStringSubmatch(scriptContent)
		if len(hlsMatch) > 1 {
			util.Debug("AnimeDrive found HLS", "url", hlsMatch[1])
			// Could return this
		}
	})

	// Method 5: search for data attributes
	doc.Find("[data-video], [data-src], [data-url]").Each(func(i int, s *goquery.Selection) {
		dataVideo, _ := s.Attr("data-video")
		if dataVideo == "" {
			dataVideo, _ = s.Attr("data-src")
		}
		if dataVideo == "" {
			dataVideo, _ = s.Attr("data-url")
		}

		if strings.Contains(dataVideo, ".mp4") || strings.Contains(dataVideo, ".m3u8") {
			util.Debug("AnimeDrive found data attribute", "url", dataVideo)
			// Could return this
		}
	})

	return "", errors.New("no video URL found")
}

// extractFromIframe tries to extract video URL from an iframe
func (c *AnimeDriveClient) extractFromIframe(iframeURL string) (string, error) {
	urlStr := iframeURL
	if strings.HasPrefix(urlStr, "//") {
		urlStr = "https:" + urlStr
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	html, _ := doc.Html()

	// Search for MP4
	mp4Match := regexp.MustCompile(`https?://[^\s<>"]+\.mp4`).FindString(html)
	if mp4Match != "" {
		return mp4Match, nil
	}

	// Search for M3U8
	m3u8Match := regexp.MustCompile(`https?://[^\s<>"]+\.m3u8`).FindString(html)
	if m3u8Match != "" {
		return m3u8Match, nil
	}

	return "", errors.New("no video URL found in iframe")
}

// GetLatestReleases fetches recent/latest anime releases
func (c *AnimeDriveClient) GetLatestReleases() ([]AnimeDriveShow, error) {
	util.Debug("AnimeDrive getting latest releases")

	req, err := http.NewRequest("GET", c.baseURL, nil)
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

	var results []AnimeDriveShow

	selectors := ".items article, .animation-2 .item, #archive-content article, .content article.item"
	doc.Find(selectors).Each(func(i int, item *goquery.Selection) {
		linkElement := item.Find("a").First()
		titleElement := item.Find("h3, .data h3, .title").First()
		imageElement := item.Find("img").First()
		ratingElement := item.Find(".rating, .score").First()
		yearElement := item.Find(".year, span.year").First()

		if linkElement.Length() == 0 {
			return
		}

		urlPath, exists := linkElement.Attr("href")
		if !exists {
			return
		}

		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			title, _ = linkElement.Attr("title")
			title = strings.TrimSpace(title)
		}

		if !strings.Contains(urlPath, "/anime/") || title == "" {
			return
		}

		image, _ := imageElement.Attr("src")
		if image == "" {
			image, _ = imageElement.Attr("data-src")
		}

		rating := strings.TrimSpace(ratingElement.Text())
		year := strings.TrimSpace(yearElement.Text())
		isDubbed := strings.Contains(strings.ToLower(title), "dublado")

		results = append(results, AnimeDriveShow{
			ID:        c.extractIDFromURL(urlPath),
			Title:     title,
			URL:       urlPath,
			Thumbnail: image,
			Rating:    rating,
			Year:      year,
			IsDubbed:  isDubbed,
		})
	})

	util.Debug("AnimeDrive found latest releases", "count", len(results))
	return results, nil
}

// GetFilms fetches available films/movies
func (c *AnimeDriveClient) GetFilms(page int) ([]AnimeDriveShow, error) {
	util.Debug("AnimeDrive getting films", "page", page)

	var pageURL string
	if page == 1 {
		pageURL = fmt.Sprintf("%s/filme/", c.baseURL)
	} else {
		pageURL = fmt.Sprintf("%s/filme/page/%d/", c.baseURL, page)
	}

	req, err := http.NewRequest("GET", pageURL, nil)
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

	var results []AnimeDriveShow

	doc.Find("article.item, .items article").Each(func(i int, item *goquery.Selection) {
		linkElement := item.Find("a[href*='/filme/']").First()
		titleElement := item.Find("h3, h2, .title").First()
		imageElement := item.Find("img").First()
		ratingElement := item.Find(".rating, .score").First()

		urlPath, exists := linkElement.Attr("href")
		if !exists || urlPath == "" {
			return
		}

		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			return
		}

		image, _ := imageElement.Attr("src")
		rating := strings.TrimSpace(ratingElement.Text())

		results = append(results, AnimeDriveShow{
			ID:        c.extractIDFromURL(urlPath),
			Title:     title,
			URL:       urlPath,
			Thumbnail: image,
			Rating:    rating,
			IsDubbed:  strings.Contains(strings.ToLower(title), "dublado"),
		})
	})

	util.Debug("AnimeDrive found films", "count", len(results))
	return results, nil
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

// GetStreamURL gets the streaming URL for a specific episode (auto-selects best server)
func (c *AnimeDriveClient) GetStreamURL(episodeURL string) (string, map[string]string, error) {
	videoURL, err := c.GetVideoURL(episodeURL)
	if err != nil {
		return "", nil, err
	}

	metadata := map[string]string{
		"source": "animedrive",
	}

	return videoURL, metadata, nil
}

// GetStreamURLWithSelection gets the streaming URL with user server selection
func (c *AnimeDriveClient) GetStreamURLWithSelection(episodeURL string) (string, map[string]string, error) {
	videoURL, err := c.GetVideoURLWithSelection(episodeURL)
	if err != nil {
		return "", nil, err
	}

	metadata := map[string]string{
		"source": "animedrive",
	}

	return videoURL, metadata, nil
}

// TotalPages returns the total number of pages available
func (c *AnimeDriveClient) TotalPages() int {
	return c.totalPages
}
