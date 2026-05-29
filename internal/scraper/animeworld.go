// Package scraper provides web scraping functionality for animeworld.ac
package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

const (
	animeWorldBase       = "https://www.animeworld.ac"
	animeWorldAPIEpisode = "https://www.animeworld.ac/api/episode/info"
	animeWorldSource     = "AnimeWorld"
)

// animeWorldEpisodeURLPattern matches AnimeWorld play paths, e.g. /play/naruto-ita.9XRsD/TMWIn.
// It also matches bare anime pages since they share the same /play/ prefix.
var animeWorldEpisodeURLPattern = regexp.MustCompile(`^/play/[^/?#]+(/[^/?#]+)?$`)

// AnimeWorldClient handles interactions with animeworld.ac.
type AnimeWorldClient struct {
	client        *http.Client
	baseURL       string
	episodeAPIURL string
	userAgent     string
	maxRetries    int
	retryDelay    time.Duration
	scraperType   ScraperType
}

// NewAnimeWorldClient creates a ready-to-use AnimeWorld client.
func NewAnimeWorldClient() *AnimeWorldClient {
	return &AnimeWorldClient{
		client:        util.NewFastClient(),
		baseURL:       animeWorldBase,
		episodeAPIURL: animeWorldAPIEpisode,
		userAgent:     UserAgent,
		maxRetries:    2,
		retryDelay:    300 * time.Millisecond,
		scraperType:   AnimeWorldType,
	}
}

// SearchAnime searches animeworld.ac for the given query and returns matching anime.
func (c *AnimeWorldClient) SearchAnime(query string) ([]*models.Anime, error) {
	query = normalizeQuery(query)
	util.Debug("AnimeWorld search", "query", query)
	return c.searchAnimeHTML(c.buildRequestURL(query))
}

// GetAnimeEpisodes returns all episodes for the anime at animeURL.
func (c *AnimeWorldClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	util.Debug("AnimeWorld episodes", "url", animeURL)
	if err := c.validateEpisodeURL(animeURL); err != nil {
		return nil, err
	}
	return c.searchAnimeEpisodesHTML(animeURL)
}

// GetStreamURL returns the direct video URL for the episode at episodeURL.
// It tries the JSON API first and falls back to HTML scraping.
func (c *AnimeWorldClient) GetStreamURL(episodeURL string) (string, error) {
	util.Debug("AnimeWorld stream URL", "url", episodeURL)
	if err := c.validateEpisodeURL(episodeURL); err != nil {
		return "", err
	}

	source, apiErr := c.searchVideoURLApi(episodeURL)
	if apiErr == nil {
		util.Debug("AnimeWorld stream URL found via API", "url", source)
		return source, nil
	}
	util.Debug("AnimeWorld API strategy failed, trying HTML fallback", "error", apiErr)

	source, htmlErr := c.searchVideoURLHTML(episodeURL)
	if htmlErr == nil {
		util.Debug("AnimeWorld stream URL found via HTML", "url", source)
		return source, nil
	}
	util.Debug("AnimeWorld HTML strategy failed", "error", htmlErr)

	return "", fmt.Errorf("AnimeWorld: all strategies failed for %s: %w",
		episodeURL, errors.Join(apiErr, htmlErr))
}

// searchAnimeHTML fetches searchURL and extracts anime search results.
func (c *AnimeWorldClient) searchAnimeHTML(searchURL string) ([]*models.Anime, error) {
	doc, err := c.fetchDoc(searchURL, "AnimeWorld search")
	if err != nil {
		return nil, err
	}
	return c.extractAnimeSearchResults(doc)
}

// extractAnimeSearchResults parses the .film-list grid into Anime models.
func (c *AnimeWorldClient) extractAnimeSearchResults(doc *goquery.Document) ([]*models.Anime, error) {
	var anime []*models.Anime

	doc.Find(".film-list .item .inner").Each(func(_ int, s *goquery.Selection) {
		nameEl := s.Find("a.name")
		name := strings.TrimSpace(nameEl.Text())
		href, _ := nameEl.Attr("href")
		if name == "" || href == "" {
			return
		}
		imgURL, _ := s.Find("a.poster img").Attr("src")
		anime = append(anime, &models.Anime{
			Name:      name,
			URL:       resolveURL(c.baseURL, href),
			ImageURL:  imgURL,
			Source:    animeWorldSource,
			MediaType: models.MediaTypeAnime,
		})
	})

	util.Debug("AnimeWorld search results", "count", len(anime))
	return anime, nil
}

// searchAnimeEpisodesHTML fetches the anime page and extracts its episode list.
func (c *AnimeWorldClient) searchAnimeEpisodesHTML(animeURL string) ([]models.Episode, error) {
	doc, err := c.fetchDoc(animeURL, "AnimeWorld episodes")
	if err != nil {
		return nil, err
	}
	return c.extractEpisodesSearchResults(doc)
}

// rawEpisode holds intermediate episode data before conversion to models.Episode.
type rawEpisode struct {
	numberStr string
	number    int
	url       string
}

// extractEpisodesSearchResults parses episode links from the anime page.
//
// AnimeWorld occasionally airs two episodes back-to-back; in that case
// data-episode-num contains a range like "45-46". The Num field is set to
// the first episode in the range; the full string is preserved in Number.
func (c *AnimeWorldClient) extractEpisodesSearchResults(doc *goquery.Document) ([]models.Episode, error) {
	var raw []rawEpisode

	doc.Find("ul.episodes li.episode a").Each(func(_ int, s *goquery.Selection) {
		numStr, exists := s.Attr("data-episode-num")
		if !exists {
			return
		}
		num, err := strconv.Atoi(numStr)
		if err != nil {
			util.Debug("AnimeWorld skipping episode with non-integer num", "data-episode-num", numStr)
			return
		}
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		raw = append(raw, rawEpisode{
			numberStr: fmt.Sprintf("%d", num),
			number:    num,
			url:       href,
		})
	})

	episodes := c.normalizeEpisodes(raw)
	util.Debug("AnimeWorld episodes parsed", "count", len(episodes))
	return episodes, nil
}

// normalizeEpisodes converts raw episode data to models.Episode, resolving
// double-episode ranges (e.g. "45-46") to their first episode number.
func (c *AnimeWorldClient) normalizeEpisodes(rawEps []rawEpisode) []models.Episode {
	episodes := make([]models.Episode, 0, len(rawEps))
	for _, raw := range rawEps {
		firstNum := raw.number
		if parts := strings.SplitN(raw.numberStr, "-", 2); len(parts) > 0 {
			if n, err := strconv.Atoi(parts[0]); err == nil {
				firstNum = n
			}
		}
		episodes = append(episodes, models.Episode{
			Number: "Episodio " + raw.numberStr,
			Num:    firstNum,
			URL:    resolveURL(c.baseURL, raw.url),
		})
	}
	return episodes
}

// animeWorldAPIResponse is the JSON shape returned by /api/episode/info.
type animeWorldAPIResponse struct {
	Grabber string `json:"grabber"` // direct video URL
	Name    string `json:"name"`
	Target  string `json:"target"`
}

// searchVideoURLApi resolves the video URL for episodeURL via the AnimeWorld JSON API.
func (c *AnimeWorldClient) searchVideoURLApi(episodeURL string) (string, error) {
	episodeID, err := c.extractEpisodeIDFromURL(episodeURL)
	if err != nil {
		return "", err
	}

	util.Debug("AnimeWorld API request", "episodeID", episodeID)

	reqURL := fmt.Sprintf("%s?id=%s", c.episodeAPIURL, episodeID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create API request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkHTTPStatus(resp, "AnimeWorld episode API"); err != nil {
		return "", err
	}

	var body animeWorldAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode API response: %w", err)
	}
	if body.Grabber == "" {
		util.Debug("AnimeWorld API returned empty grabber", "episodeID", episodeID)
		return "", errors.New("API returned no video URL")
	}

	util.Debug("AnimeWorld API response", "grabber", body.Grabber, "name", body.Name)
	return body.Grabber, nil
}

// extractEpisodeIDFromURL returns the last path segment of episodeURL, which
// AnimeWorld uses as the episode ID (e.g. "TMWIn" from /play/naruto.9XRsD/TMWIn).
func (c *AnimeWorldClient) extractEpisodeIDFromURL(episodeURL string) (string, error) {
	idx := strings.LastIndex(episodeURL, "/")
	if idx == -1 || idx == len(episodeURL)-1 {
		return "", fmt.Errorf("cannot extract episode ID from URL: %s", episodeURL)
	}
	return episodeURL[idx+1:], nil
}

// searchVideoURLHTML scrapes the download links from the episode page as a
// fallback when the API is unavailable.
func (c *AnimeWorldClient) searchVideoURLHTML(episodeURL string) (string, error) {
	doc, err := c.fetchDoc(episodeURL, "AnimeWorld episode page")
	if err != nil {
		return "", err
	}

	for _, selector := range []string{"#alternativeDownloadLink", "#downloadLink"} {
		sel := doc.Find(selector)
		if sel.Length() == 0 {
			util.Debug("AnimeWorld HTML selector miss", "selector", selector)
			continue
		}
		href, ok := sel.Attr("href")
		if !ok {
			util.Debug("AnimeWorld HTML selector found but no href", "selector", selector)
			continue
		}
		videoURL, err := c.normalizeVideoURL(href)
		if err != nil {
			util.Debug("AnimeWorld video URL normalization failed", "selector", selector, "href", href, "error", err)
			continue
		}
		util.Debug("AnimeWorld HTML selector hit", "selector", selector, "url", videoURL)
		return videoURL, nil
	}

	return "", fmt.Errorf("no video source found for episode: %s", episodeURL)
}

// fetchDoc performs a GET request to url with retries and returns the parsed HTML document.
func (c *AnimeWorldClient) fetchDoc(url, httpSource string) (*goquery.Document, error) {
	var lastErr error

	for attempt := range c.maxRetries + 1 {
		if attempt > 0 {
			util.Debug("AnimeWorld retrying request", "source", httpSource, "attempt", attempt+1, "url", url)
			time.Sleep(c.retryDelay)
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.decorateRequest(req)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d): %w", attempt+1, err)
			util.Debug("AnimeWorld request error", "source", httpSource, "attempt", attempt+1, "error", err)
			continue
		}

		if err := checkHTTPStatus(resp, httpSource); err != nil {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("bad response (attempt %d): %w", attempt+1, err)
			util.Debug("AnimeWorld bad HTTP status", "source", httpSource, "status", resp.StatusCode, "attempt", attempt+1)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("parse HTML: %w", err)
		}

		util.Debug("AnimeWorld page fetched", "source", httpSource, "status", resp.StatusCode, "url", url)
		return doc, nil
	}

	return nil, lastErr
}

// buildRequestURL constructs the search endpoint URL for the given query.
func (c *AnimeWorldClient) buildRequestURL(query string) string {
	return fmt.Sprintf("%s/search?keyword=%s", c.baseURL, url.QueryEscape(query))
}

// decorateRequest adds standard browser-like headers to req.
func (c *AnimeWorldClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Language", "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
}

// validateEpisodeURL returns an error if episodeURL does not match the expected
// AnimeWorld /play/ path pattern. Anime page URLs are also accepted because
// AnimeWorld uses the same /play/ prefix for both.
func (c *AnimeWorldClient) validateEpisodeURL(episodeURL string) error {
	u, err := url.Parse(episodeURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", episodeURL, err)
	}
	if !animeWorldEpisodeURLPattern.MatchString(u.Path) {
		return fmt.Errorf("unexpected URL path %q", u.String())
	}
	return nil
}

// normalizeVideoURL converts download-file.php?id=<path> wrapper URLs into
// direct .mp4 URLs. Already-direct .mp4 URLs are returned unchanged.
func (c *AnimeWorldClient) normalizeVideoURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid video URL: %w", err)
	}

	if strings.Contains(u.Path, "download-file.php") {
		id := u.Query().Get("id")
		if id == "" {
			return "", fmt.Errorf("missing id parameter in URL: %s", raw)
		}
		return (&url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
			Path:   "/" + strings.TrimPrefix(id, "/"),
		}).String(), nil
	}

	if strings.HasSuffix(u.Path, ".mp4") {
		return u.String(), nil
	}

	return "", fmt.Errorf("unsupported video URL: %s", raw)
}
