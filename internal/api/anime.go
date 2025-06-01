package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// func SearchAnime(animeName string) (*models.Anime, error) {
// 	start := time.Now()
// 	if util.IsDebug {
// 		log.Printf("[PERF] SearchAnime iniciado para %s", animeName)
// 	}

// 	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", models.AnimeFireURL, url.PathEscape(animeName))

// 	if util.IsDebug {
// 		log.Printf("Searching for anime with URL: %s", currentPageURL)
// 	}

// 	for {
// 		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
// 		if err != nil {
// 			log.Printf("[PERF] SearchAnime falhou para %s em %v", animeName, time.Since(start))
// 			return nil, err
// 		}
// 		if selectedAnime != nil {

// 			aniListInfo, err := FetchAnimeFromAniList(selectedAnime.Name)
// 			if err != nil {
// 				log.Printf("Error fetching additional data from AniList: %v", err)
// 			} else {
// 				selectedAnime.AnilistID = aniListInfo.Data.Media.ID
// 				selectedAnime.MalID = aniListInfo.Data.Media.IDMal
// 				selectedAnime.Details = aniListInfo.Data.Media

// 				if aniListInfo.Data.Media.CoverImage.Large != "" {
// 					selectedAnime.ImageURL = aniListInfo.Data.Media.CoverImage.Large
// 					if util.IsDebug {
// 						log.Printf("Cover image URL retrieved from AniList: %s", selectedAnime.ImageURL)
// 					}
// 				} else {
// 					log.Printf("Cover image URL not found in AniList response for anime: %s", selectedAnime.Name)
// 				}
// 				if util.IsDebug {
// 					log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
// 						aniListInfo.Data.Media.ID, aniListInfo.Data.Media.IDMal,
// 						aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore,
// 						selectedAnime.ImageURL)
// 				}

// 			}
// 			if util.IsDebug {
// 				log.Printf("[PERF] SearchAnime finalizado para %s em %v", animeName, time.Since(start))

// 			}

// 			return selectedAnime, nil
// 		}

// 		if nextPageURL == "" {
// 			if util.IsDebug {
// 				log.Printf("[PERF] SearchAnime não encontrou resultados para %s em %v", animeName, time.Since(start))

// 			}

// 			return nil, errors.New("no anime found with the given name")
// 		}
// 		currentPageURL = models.AnimeFireURL + nextPageURL
// 	}
// }

// func searchAnimeOnPage(pageURL string) (*models.Anime, string, error) {
// 	response, err := getHTTPResponse(pageURL)
// 	if err != nil {
// 		return nil, "", errors.Wrap(err, "failed to perform search request")
// 	}
// 	defer func(Body io.ReadCloser) {
// 		err := Body.Close()
// 		if err != nil {
// 			log.Printf("Error closing response body: %v", err)
// 		}
// 	}(response.Body)

// 	if response.StatusCode != http.StatusOK {
// 		if response.StatusCode == http.StatusForbidden {
// 			return nil, "", errors.New("connection refused: you need to be in Brazil or use a VPN to access the server")
// 		}
// 		return nil, "", errors.Errorf("search failed, server returned: %s", response.Status)
// 	}

// 	doc, err := goquery.NewDocumentFromReader(response.Body)
// 	if err != nil {
// 		return nil, "", errors.Wrap(err, "failed to parse response")
// 	}

// 	animes := ParseAnimes(doc)
// 	if util.IsDebug {
// 		log.Printf("Number of animes found: %d", len(animes))
// 	}

// 	if len(animes) > 0 {
// 		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
// 		if err != nil {
// 			return nil, "", err
// 		}
// 		return selectedAnime, "", nil
// 	}

// 	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
// 	if !exists {
// 		return nil, "", nil
// 	}

// 	return nil, nextPage, nil
// }

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

// // ParseAnimes extracts a list of Anime structs from the search results page.
// func ParseAnimes(doc *goquery.Document) []models.Anime {
// 	var animes []models.Anime

// 	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
// 		urlPath, exists := s.Attr("href")
// 		if !exists {
// 			return
// 		}
// 		url := resolveURL(models.AnimeFireURL, urlPath)

// 		name := strings.TrimSpace(s.Text())

// 		if util.IsDebug {
// 			log.Printf("Parsed Anime - Name: %s, URL: %s", name, url)
// 		}

// 		animes = append(animes, models.Anime{
// 			Name: name,
// 			URL:  url,
// 		})
// 	})

// 	return animes
// }

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

// func FetchAnimeFromAniList(animeName string) (*models.AniListResponse, error) {
// 	cleanedName := CleanTitle(animeName)
// 	if util.IsDebug {
// 		log.Printf("Attempting AniList search with title: %s", cleanedName)

// 	}

// 	query := `
//     query ($search: String) {
//         Media(search: $search, type: ANIME) {
//             id
//             title { romaji english }
//             description
//             genres
//             averageScore
//             episodes
//             status
//             idMal
//             coverImage { large medium }
//         }
//     }`

// 	variables := map[string]interface{}{
// 		"search": cleanedName,
// 	}

// 	requestBody := map[string]interface{}{
// 		"query":     query,
// 		"variables": variables,
// 	}

// 	jsonData, err := json.Marshal(requestBody)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request body: %v", err)
// 	}

// 	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(string(jsonData)))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create request: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch data from AniList API: %v", err)
// 	}
// 	defer func(Body io.ReadCloser) {
// 		err := Body.Close()
// 		if err != nil {
// 			log.Printf("Error closing AniList response body: %v", err)
// 		}
// 	}(resp.Body)

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("AniList API request failed with status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result models.AniListResponse
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read AniList API response: %v", err)
// 	}

// 	if err := json.Unmarshal(body, &result); err != nil {
// 		return nil, fmt.Errorf("failed to parse AniList API response: %v", err)
// 	}

// 	if result.Data.Media.ID == 0 {
// 		log.Printf("No results found on AniList for anime: %s", cleanedName)
// 		return nil, fmt.Errorf("no results found on AniList for anime: %s", cleanedName)
// 	}

// 	if util.IsDebug {
// 		log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
// 			result.Data.Media.ID, result.Data.Media.IDMal,
// 			result.Data.Media.Title.Romaji, result.Data.Media.AverageScore,
// 			result.Data.Media.CoverImage.Large)
// 	}

// 	return &result, nil
// }

// // selectAnimeWithGoFuzzyFinder allows the user to select an anime from a list using fuzzy search
// func selectAnimeWithGoFuzzyFinder(animes []models.Anime) (*models.Anime, error) {
// 	if len(animes) == 0 {
// 		return nil, errors.New("no anime provided")
// 	}

// 	sortedAnimes := sortAnimes(animes)
// 	idx, err := fuzzyfinder.Find(
// 		sortedAnimes,
// 		func(i int) string {
// 			return sortedAnimes[i].Name
// 		},
// 	)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
// 	}

// 	if idx < 0 || idx >= len(sortedAnimes) {
// 		return nil, errors.New("invalid index returned by fuzzyfinder")
// 	}

// 	return &sortedAnimes[idx], nil
// }

// // sortAnimes sorts a list of Anime structs alphabetically by name
// func sortAnimes(animeList []models.Anime) []models.Anime {
// 	sort.Slice(animeList, func(i, j int) bool {
// 		return animeList[i].Name < animeList[j].Name
// 	})
// 	return animeList
// }

// // resolveURL resolves relative URLs to absolute URLs based on the base URL.
// func resolveURL(base, ref string) string {
// 	baseURL, err := url.Parse(base)
// 	if err != nil {
// 		return ""
// 	}
// 	refURL, err := url.Parse(ref)
// 	if err != nil {
// 		return ""
// 	}
// 	return baseURL.ResolveReference(refURL).String()
// }

// // CleanTitle removes unwanted words, numbers, and ratings for better API search results
// func CleanTitle(title string) string {
// 	re := regexp.MustCompile(`(?i)(dublado|legendado|todos os episodios)`)
// 	title = re.ReplaceAllString(title, "")

// 	re = regexp.MustCompile(`\s+\d+(\.\d+)?\s+A\d+$`)
// 	title = re.ReplaceAllString(title, "")

// 	title = strings.TrimSpace(title)
// 	return title
// }

// func makeGetRequest(url string, headers map[string]string) (map[string]interface{}, error) {
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create GET request: %w", err)
// 	}

// 	for key, value := range headers {
// 		req.Header.Set(key, value)
// 	}

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to send GET request: %w", err)
// 	}
// 	defer func(Body io.ReadCloser) {
// 		err := Body.Close()
// 		if err != nil {
// 			fmt.Printf("failed to close response body: %v", err)
// 		}
// 	}(resp.Body)

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed with status %d: %s", resp.StatusCode, body)
// 	}

// 	var responseData map[string]interface{}
// 	err = json.Unmarshal(body, &responseData)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	return responseData, nil
// }

// func getHTTPResponse(url string) (*http.Response, error) {
// 	client := &http.Client{}
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

func SearchAnime(animeName string) (*models.Anime, error) {
	start := time.Now()
	if util.IsDebug {
		log.Printf("[PERF] SearchAnime started for %s", animeName)
	}

	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", models.AnimeFireURL, url.PathEscape(animeName))

	for {
		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			log.Printf("[PERF] SearchAnime failed for %s after %v", animeName, time.Since(start))
			return nil, err
		}
		if selectedAnime != nil {
			if err := enrichAnimeData(selectedAnime); err != nil && util.IsDebug {
				log.Printf("Error enriching anime data: %v", err)
			}
			if util.IsDebug {
				log.Printf("[PERF] SearchAnime completed for %s in %v", animeName, time.Since(start))
			}
			return selectedAnime, nil
		}

		if nextPageURL == "" {
			if util.IsDebug {
				log.Printf("[PERF] No results found for %s after %v", animeName, time.Since(start))
			}
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
	} else if util.IsDebug {
		log.Printf("Cover image not found for: %s", anime.Name)
	}

	if util.IsDebug {
		log.Printf("AniList Data: ID:%d, MAL:%d, Title:%s",
			aniListInfo.Data.Media.ID,
			aniListInfo.Data.Media.IDMal,
			aniListInfo.Data.Media.Title.Romaji)
	}
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
	if util.IsDebug {
		log.Printf("Found %d anime(s)", len(animes))
	}

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
			if util.IsDebug {
				log.Printf("Parsed: %s", name)
			}
		}
	})
	return animes
}

func FetchAnimeFromAniList(animeName string) (*models.AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	if util.IsDebug {
		log.Printf("Querying AniList for: %s", cleanedName)
	}

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
		return animes[i].Name
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
	resp, err := httpGetWithUA(url)
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
	re := regexp.MustCompile(`(?i)(dublado|legendado|todos os episodios)|\s+\d+(\.\d+)?\s+A\d+$`)
	return strings.TrimSpace(re.ReplaceAllString(title, ""))
}

func safeClose(closer io.Closer, name string) {
	if err := closer.Close(); err != nil && util.IsDebug {
		log.Printf("Error closing %s: %v", name, err)
	}
}
