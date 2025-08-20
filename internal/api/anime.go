package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

// Common HTTP client instance
var httpClient = &http.Client{}

func GetEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {

	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes/%d", animeID, episodeNo)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return fmt.Errorf("error fetching data from Jikan (MyAnimeList) API: %w", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response structure: missing or invalid 'data' field")
	}

	getStringValue := func(field string) string {
		if value, ok := data[field].(string); ok {
			return value
		}
		return ""
	}

	getIntValue := func(field string) int {
		if value, ok := data[field].(float64); ok {
			return int(value)
		}
		return 0
	}

	getBoolValue := func(field string) bool {
		if value, ok := data[field].(bool); ok {
			return value
		}
		return false
	}

	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}
	anime.Episodes[0].Title.Romaji = getStringValue("title_romanji")
	anime.Episodes[0].Title.English = getStringValue("title")
	anime.Episodes[0].Title.Japanese = getStringValue("title_japanese")
	anime.Episodes[0].Aired = getStringValue("aired")
	anime.Episodes[0].Duration = getIntValue("duration")
	anime.Episodes[0].IsFiller = getBoolValue("filler")
	anime.Episodes[0].IsRecap = getBoolValue("recap")
	anime.Episodes[0].Synopsis = getStringValue("synopsis")

	return nil
}

// GetMovieData fetches movie/OVA data for a given anime ID from Jikan API
func GetMovieData(animeID int, anime *models.Anime) error {

	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", animeID)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return fmt.Errorf("error fetching data from Jikan (MyAnimeList) API: %w", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response structure: missing or invalid 'data' field")
	}

	// Helper functions to safely get values
	getStringValue := func(field string) string {
		if value, ok := data[field].(string); ok {
			return value
		}
		return ""
	}

	getIntValue := func(field string) int {
		if value, ok := data[field].(float64); ok {
			return int(value)
		}
		return 0
	}

	getBoolValue := func(field string) bool {
		if value, ok := data[field].(bool); ok {
			return value
		}
		return false
	}

	// Assign values to the Anime struct
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}
	anime.Episodes[0].Title.Romaji = getStringValue("title_romanji")
	anime.Episodes[0].Title.English = getStringValue("title")
	anime.Episodes[0].Title.Japanese = getStringValue("title_japanese")
	anime.Episodes[0].Aired = getStringValue("aired")
	anime.Episodes[0].Duration = getIntValue("duration")
	anime.Episodes[0].IsFiller = getBoolValue("filler")
	anime.Episodes[0].IsRecap = getBoolValue("recap")
	anime.Episodes[0].Synopsis = getStringValue("synopsis")

	return nil
}

// FetchAnimeDetails retrieves additional information for the selected anime
func FetchAnimeDetails(anime *models.Anime) error {
	response, err := http.Get(anime.URL)
	if err != nil {
		return errors.Wrap(err, "failed to get anime details page")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error get details")

		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get anime details page: %s", response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to parse anime details page")
	}

	imageURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
	if !exists || imageURL == "" {
		return errors.New("cover image URL not found")
	}

	return nil
}

func SearchAnime(animeName string) (*models.Anime, error) {
	start := time.Now()
	util.Debugf("[PERF] SearchAnime started for %s", animeName)

	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", models.AnimeFireURL, url.PathEscape(animeName))

	for {
		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			util.Debugf("[PERF] SearchAnime failed for %s after %v", animeName, time.Since(start))
			return nil, err
		}
		if selectedAnime != nil {
			if err := enrichAnimeData(selectedAnime); err != nil {
				util.Errorf("Error enriching anime data: %v", err)
			}
			util.Debugf("[PERF] SearchAnime completed for %s in %v", animeName, time.Since(start))
			return selectedAnime, nil
		}

		if nextPageURL == "" {
			util.Debugf("[PERF] No results found for %s after %v", animeName, time.Since(start))
			return nil, errors.New("no anime found with the given name")
		}
		currentPageURL = models.AnimeFireURL + nextPageURL
	}
}

// Unified function to fetch anime data from Jikan API
func FetchAnimeData(animeID int, episodeNo int, anime *models.Anime) error {
	endpoint := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", animeID)
	if episodeNo > 0 {
		endpoint = fmt.Sprintf("%s/episodes/%d", endpoint, episodeNo)
	}

	data, err := makeGetRequest(endpoint, nil)
	if err != nil {
		return fmt.Errorf("jikan API request failed: %w", err)
	}

	responseData, ok := data["data"].(map[string]interface{})
	if !ok {
		return errors.New("invalid response structure from Jikan API")
	}

	// Ensure anime has at least one episode
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	// Populate episode data from response
	ep := &anime.Episodes[0]
	ep.Title.Romaji = getStringValue(responseData, "title_romanji")
	ep.Title.English = getStringValue(responseData, "title")
	ep.Title.Japanese = getStringValue(responseData, "title_japanese")
	ep.Aired = getStringValue(responseData, "aired")
	ep.Duration = getIntValue(responseData, "duration")
	ep.IsFiller = getBoolValue(responseData, "filler")
	ep.IsRecap = getBoolValue(responseData, "recap")
	ep.Synopsis = getStringValue(responseData, "synopsis")

	return nil
}

// Helper functions for map value extraction
func getStringValue(data map[string]interface{}, field string) string {
	val, _ := data[field].(string)
	return val
}

func getIntValue(data map[string]interface{}, field string) int {
	if val, ok := data[field].(float64); ok {
		return int(val)
	}
	return 0
}

func getBoolValue(data map[string]interface{}, field string) bool {
	val, _ := data[field].(bool)
	return val
}

// Enrich anime data from AniList
func enrichAnimeData(anime *models.Anime) error {
	aniListInfo, err := FetchAnimeFromAniList(anime.Name)
	if err != nil {
		return fmt.Errorf("AniList enrichment failed: %w", err)
	}

	anime.AnilistID = aniListInfo.Data.Media.ID
	anime.MalID = aniListInfo.Data.Media.IDMal
	anime.Details = aniListInfo.Data.Media

	if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
		anime.ImageURL = cover
	} else {
		util.Debugf("Cover image not found for: %s", anime.Name)
	}

	util.Debugf("AniList Data: ID:%d, MAL:%d, Title:%s",
		aniListInfo.Data.Media.ID,
		aniListInfo.Data.Media.IDMal,
		aniListInfo.Data.Media.Title.Romaji)
	return nil
}

func searchAnimeOnPage(pageURL string) (*models.Anime, string, error) {
	resp, err := httpGetWithUA(pageURL)
	if err != nil {
		return nil, "", errors.Wrap(err, "HTTP request failed")
	}
	defer safeClose(resp.Body, "search page response body")

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return nil, "", errors.New("access restricted: VPN required")
		}
		return nil, "", fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "HTML parsing failed")
	}

	animes := ParseAnimes(doc)
	util.Debugf("Found %d anime(s)", len(animes))

	if len(animes) > 0 {
		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
		return selectedAnime, "", err
	}

	if nextPage, exists := doc.Find(".pagination .next a").Attr("href"); exists {
		return nil, nextPage, nil
	}
	return nil, "", nil
}

func ParseAnimes(doc *goquery.Document) []models.Anime {
	var animes []models.Anime

	doc.Find(".row.ml-1.mr-1 a").Each(func(_ int, s *goquery.Selection) {
		if urlPath, exists := s.Attr("href"); exists {
			name := strings.TrimSpace(s.Text())
			animes = append(animes, models.Anime{
				Name: name,
				URL:  resolveURL(models.AnimeFireURL, urlPath),
			})
			util.Debugf("Parsed: %s", name)
		}
	})
	return animes
}

func FetchAnimeFromAniList(animeName string) (*models.AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	util.Debugf("Querying AniList for: %s", cleanedName)

	query := `query ($search: String) {
        Media(search: $search, type: ANIME) {
            id
            title { romaji english }
            idMal
            coverImage { large }
        }
    }`

	jsonData, err := json.Marshal(map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"search": cleanedName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("JSON marshal failed: %w", err)
	}

	resp, err := httpPost("https://graphql.anilist.co", jsonData)
	if err != nil {
		return nil, fmt.Errorf("AniList request failed: %w", err)
	}
	defer safeClose(resp.Body, "AniList response body")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList returned: %s", resp.Status)
	}

	var result models.AniListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	if result.Data.Media.ID == 0 {
		return nil, errors.New("no matching anime found on AniList")
	}
	return &result, nil
}

func selectAnimeWithGoFuzzyFinder(animes []models.Anime) (*models.Anime, error) {
	if len(animes) == 0 {
		return nil, errors.New("no anime available for selection")
	}

	sort.Slice(animes, func(i, j int) bool {
		return animes[i].Name < animes[j].Name
	})

	idx, err := fuzzyfinder.Find(animes, func(i int) string {
		name := animes[i].Name
		name = strings.ReplaceAll(name, "[AllAnime]", "[English]")
		name = strings.ReplaceAll(name, "[AnimeFire]", "[Portuguese]")
		return name
	})
	if err != nil {
		return nil, fmt.Errorf("fuzzy selection failed: %w", err)
	}

	if idx < 0 || idx >= len(animes) {
		return nil, errors.New("invalid selection index")
	}
	return &animes[idx], nil
}

// HTTP helper functions
func httpGetWithUA(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	return httpClient.Do(req)
}

func httpPost(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func makeGetRequest(url string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	// Set default User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	// Set additional headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}
	defer safeClose(resp.Body, "API response body")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}
	return responseData, nil
}

// Utility functions
func resolveURL(base, ref string) string {
	baseURL, _ := url.Parse(base)
	refURL, _ := url.Parse(ref)
	return baseURL.ResolveReference(refURL).String()
}

func CleanTitle(title string) string {
	// Remove common AnimeFire suffixes and patterns
	re := regexp.MustCompile(`(?i)(dublado|legendado|todos os episodios)|\s+\d+(\.\d+)?\s+A\d+$|\s+\d+(\.\d+)?$`)
	cleaned := strings.TrimSpace(re.ReplaceAllString(title, ""))

	// Remove anything in parentheses if it contains "dublado", "legendado", etc.
	re2 := regexp.MustCompile(`(?i)\s*\([^)]*(?:dublado|legendado|dub|sub)[^)]*\)`)
	cleaned = strings.TrimSpace(re2.ReplaceAllString(cleaned, ""))

	// Remove source tags like 🔥[AnimeFire] or 🌐[AllAnime]
	re3 := regexp.MustCompile(`(?i)[🔥🌐]?\[(?:animefire|allanime)\]\s*`)
	cleaned = strings.TrimSpace(re3.ReplaceAllString(cleaned, ""))

	// Remove AllAnime specific patterns like "(171 episodes)" or "(1 episodes)"
	re4 := regexp.MustCompile(`\s*\(\d+\s+episodes?\)`)
	cleaned = strings.TrimSpace(re4.ReplaceAllString(cleaned, ""))

	// Remove special titles and common additions
	re5 := regexp.MustCompile(`(?i):\s*(Jump Festa \d+|The All Magic Knights|Sword of the Wizard King|Mahou Tei no Ken).*$`)
	cleaned = strings.TrimSpace(re5.ReplaceAllString(cleaned, ""))

	// Additional cleanup for colons and extra spaces
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

func safeClose(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		util.Debugf("Error closing %s: %v", name, err)
	}
}
