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
	animeWorldApiEpisode = "https://www.animeworld.ac/api/episode/info"
	animeWorldSource     = "AnimeWorld"
)

var animeWorldEpisodeURLPattern = regexp.MustCompile(`^/play/[^/?#]+(/[^/?#]+)?$`)

// TODO CHECK ALSO THE MANAGER THERE ARE ADAOTER ETC... I DONT THINK THIS SHOULD IMPLEMENT THE SCRAPER INTERFACE, BUT MAYBE AN ADAPTER MUST BE USED!
// ALSO FOR SCRAPERTYPE I DUNNOOOO. GOYABUCLIENT DOESNT IMPLEMENT INTERFACE, BUT ITS IN A ADAPTOR
// QUEL MAP PUZZA, NON CI STA DOCUMENTAZIONE CHE DICE CHE COSA E'
type AnimeWorldClient struct {
	client        *http.Client
	baseURL       string
	episodeApiURL string
	userAgent     string
	maxRetries    int
	retryDelay    time.Duration
	scraperType   ScraperType
}

func NewAnimeWorldClient() *AnimeWorldClient {
	return &AnimeWorldClient{
		client:        util.NewFastClient(),
		baseURL:       animeWorldBase,
		episodeApiURL: animeWorldApiEpisode,
		userAgent:     UserAgent,
		maxRetries:    2,
		retryDelay:    300 * time.Millisecond,
		scraperType:   AnimeWorldType,
	}
}

// SearchAnime searches for anime on animeworld.ac
func (c *AnimeWorldClient) SearchAnime(query string) ([]*models.Anime, error) {
	// query "attacco dei giganti"
	// url => https://www.animeworld.ac/search?keyword=attacco+dei+giganti
	query = normalizeQuery(query)

	util.Debug("AnimeWorld search", "query", query)

	searchURL := c.buildRequestURL(query)
	return c.searchAnimeHTML(searchURL)
}

// GetAnimeEpisodes scrapes the episodes of an anime given its URL
func (c *AnimeWorldClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	if err := c.validateEpisodeURL(animeURL); err != nil {
		return nil, err
	}
	return c.searchAnimeEpisodesHTML(animeURL)
}

// GetStreamURL fetches the video source URL of an episode
func (c *AnimeWorldClient) GetStreamURL(episodeURL string) (string, error) {
	if err := c.validateEpisodeURL(episodeURL); err != nil {
		return "", err
	}
	source, apiErr := c.searchVideoURLApi(episodeURL)
	if apiErr == nil {
		return source, nil
	}
	// fallback to HTML
	source, htmlErr := c.searchVideoURLHTML(episodeURL)
	if htmlErr == nil {
		return source, nil
	}
	return "", fmt.Errorf("AnimeWorld search failed for episode URL: %s: %w",
		episodeURL,
		errors.Join(apiErr, htmlErr))
}

func (c *AnimeWorldClient) searchAnimeHTML(searchURL string) ([]*models.Anime, error) {
	doc, err := c.fetchDoc(searchURL, "AnimeWorld Anime search HTML")
	if err != nil {
		return nil, err
	}
	return c.extractAnimeSearchResults(doc)
}

func (c *AnimeWorldClient) extractAnimeSearchResults(doc *goquery.Document) ([]*models.Anime, error) {
	var anime []*models.Anime

	doc.Find(".film-list .item .inner").Each(func(_ int, s *goquery.Selection) {
		nameEl := s.Find("a.name")
		name := strings.TrimSpace(nameEl.Text())
		href, _ := nameEl.Attr("href")
		if name == "" || href == "" {
			return
		}

		// unhandled error: an empty imgURL is self-explanatory
		imgURL, _ := s.Find("a.poster img").Attr("src")

		anime = append(anime, &models.Anime{
			Name:      name,
			URL:       resolveURL(c.baseURL, href),
			ImageURL:  imgURL,
			Source:    animeWorldSource,
			MediaType: models.MediaTypeAnime,
		})
	})
	return anime, nil
}

func (c *AnimeWorldClient) searchAnimeEpisodesHTML(animeURL string) ([]models.Episode, error) {
	doc, err := c.fetchDoc(animeURL, "AnimeWorld Episodes search HTML")
	if err != nil {
		return nil, err
	}
	return c.extractEpisodesSearchResults(doc)
}

// TODO NUM EPISODE CHECK IF BREAKS THE APPLICATION

// animeWorldRawEpisodes represents the raw episode data scraped from World Anime
type animeWorldRawEpisodes struct {
	episodeNumberStr string
	episodeNumber    int
	episodeURL       string
}

// extractEpisodesSearchResults ignores episodes that for some resons are not found
//
// handles also the case of a "double episode". It's not common, but sometimes
// a video source contains two episodes one after others. Probably because it was aired in this way in Italy
// A double episode can be for example "45-46"
// I don't know how to handle this case for the Num field in models.Episode.
func (c *AnimeWorldClient) extractEpisodesSearchResults(doc *goquery.Document) ([]models.Episode, error) {
	var episodes []animeWorldRawEpisodes
	doc.Find("ul.episodes li.episode a").Each(func(i int, s *goquery.Selection) {
		numberStr, exists := s.Attr("data-episode-num")
		if !exists {
			return
		}
		num, err := strconv.Atoi(numberStr)
		if err != nil {
			return
		}
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		episodes = append(episodes, animeWorldRawEpisodes{
			episodeNumberStr: fmt.Sprintf("%d", num),
			episodeNumber:    num,
			episodeURL:       href,
		})
	})
	return c.adjustEpisodes(episodes), nil
}

// adjustEpisodes synchronizes the episode name number with the episode name if there are double episodes
// Otherwise just maps to []models.Episode. See Tests for more info
func (c *AnimeWorldClient) adjustEpisodes(rawEps []animeWorldRawEpisodes) []models.Episode {
	var episodes []models.Episode
	for _, raw := range rawEps {
		// Parse the actual episode number from the string (handles "3", "3-4", etc.)
		firstNum := raw.episodeNumber // fallback if parsing fails
		if parts := strings.SplitN(raw.episodeNumberStr, "-", 2); len(parts) > 0 {
			if n, err := strconv.Atoi(parts[0]); err == nil {
				firstNum = n
			}
		}
		episodes = append(episodes, models.Episode{
			Number: "Episodio " + raw.episodeNumberStr,
			Num:    firstNum,
			URL:    resolveURL(c.baseURL, raw.episodeURL),
		})
	}
	return episodes
}

type animeWorldAPIResponse struct {
	Grabber string `json:"grabber"` // source URL
	Name    string `json:"name"`
	Target  string `json:"target"`
}

// searchVideoURLApi fetches the videoURL source of the episode from the AnimeWorld api.
// This is the preferred way.
func (c *AnimeWorldClient) searchVideoURLApi(episodeURL string) (string, error) {
	// extract the episode ID
	episodeID, err := c.extractEpisodeIDFromURL(episodeURL)
	if err != nil {
		return "", err
	}
	reqURL := fmt.Sprintf("%s?id=%s", c.episodeApiURL, episodeID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AnimeWorld api call: %w", err)
	}
	defer resp.Body.Close()
	var rBody animeWorldAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&rBody); err != nil {
		return "", fmt.Errorf("AnimeWorld animeWorldAPIResponse unmarshalling: %w", err)
	}
	if rBody.Grabber == "" {
		return "", fmt.Errorf("AnimeWorld no episode url found from the api")
	}
	return rBody.Grabber, nil
}

func (c *AnimeWorldClient) extractEpisodeIDFromURL(episodeURL string) (string, error) {
	// Example episode URL: https://www.animeworld.ac/play/naruto-shippuden-ita.9XRsD/TMWIn
	// ID: TMWIn
	lastDotIndex := strings.LastIndex(episodeURL, "/")
	if lastDotIndex == -1 {
		return "", errors.New("invalid URL format: episode ID not found")
	}
	episodeID := episodeURL[lastDotIndex+1:]
	if len(episodeID) == 0 {
		return "", errors.New("invalid URL format: episode ID is empty")
	}
	return episodeID, nil
}

// searchVideoURLHTML fetches the video source URL of the episode.
// AnimeWorld can offer more than one URL. The first scraped URL is returned
func (c *AnimeWorldClient) searchVideoURLHTML(episodeURL string) (string, error) {
	doc, err := c.fetchDoc(episodeURL, "AnimeWorld episode video source fetch HTML")
	if err != nil {
		return "", fmt.Errorf("AnimeWorld failed to parse HTML: %s", episodeURL)
	}

	get := func(selector string) (string, bool) {
		sel := doc.Find(selector)
		if sel.Length() == 0 {
			return "", false
		}

		href, ok := sel.Attr("href")
		if !ok {
			return "", false
		}

		vURL, err := c.normalizeAnimeWorldVideoURL(href)
		if err != nil {
			return "", false
		}

		return vURL, true
	}

	// preference order
	if url, ok := get("#alternativeDownloadLink"); ok {
		return url, nil
	}

	if url, ok := get("#downloadLink"); ok {
		return url, nil
	}

	return "", fmt.Errorf("AnimeWorld no video source found for episode: %s", episodeURL)
}

// fetchDoc makes an HTTP call to the given url and parses the HTML to a goquery.Document struct.
// It requires a httpSource string that will be used by checkHTTPStatus
func (c *AnimeWorldClient) fetchDoc(url, httpSource string) (*goquery.Document, error) {
	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
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
			continue
		}

		if err := checkHTTPStatus(resp, httpSource); err != nil {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("bad animeWorldAPIResponse (attempt %d): %w", attempt+1, err)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}

		return doc, nil
	}

	return nil, lastErr
}

func (c *AnimeWorldClient) buildRequestURL(query string) string {
	query = strings.ReplaceAll(query, " ", "+")
	return fmt.Sprintf("%s/search?keyword=%s", c.baseURL, query)
}

func (c *AnimeWorldClient) decorateRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Language", "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	// referer? try if you get 403
}

// validateEpisodeURL works also on anime URL because AnimeWorld
// the anime URL and the URL of the first episode of the first season
// is the same
func (c *AnimeWorldClient) validateEpisodeURL(episodeURL string) error {
	u, err := url.Parse(episodeURL)
	if err != nil {
		return fmt.Errorf("AnimeWorld invalid URL: %s", episodeURL)
	}
	if !animeWorldEpisodeURLPattern.MatchString(u.Path) {
		return fmt.Errorf("AnimeWorld invalid URL path: %s", u.String())
	}
	return nil
}

// normalizeAnimeWorldVideoURL converts wrapper URLs (download-file.php?id=...)
// into direct .mp4 URLs by extracting the id parameter.
// Returns the original URL if already a direct .mp4 link.
func (c *AnimeWorldClient) normalizeAnimeWorldVideoURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	// Case 1: download-file.php?id=...
	if strings.Contains(u.Path, "download-file.php") {
		id := u.Query().Get("id")
		if id == "" {
			return "", fmt.Errorf("missing id in url: %s", raw)
		}

		// build proper URL using url.URL
		normalized := &url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
			Path:   fmt.Sprintf("/%s", strings.TrimPrefix(id, "/")),
		}

		return normalized.String(), nil
	}

	// Case 2: already direct mp4
	if strings.HasSuffix(u.Path, ".mp4") {
		return u.String(), nil
	}

	return "", fmt.Errorf("unsupported video url: %s", raw)
}
